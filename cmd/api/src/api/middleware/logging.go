// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/packages/go/headers"
)

// PanicHandler is a middleware func that sets up a defer-recovery trap to capture any unhandled panics that bubble
// up the request handler stack.
func PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovery := recover(); recovery != nil {
				slog.ErrorContext(request.Context(), fmt.Sprintf("[panic recovery] %s - [stack trace] %s", recovery, debug.Stack()))
			}
		}()

		next.ServeHTTP(response, request)
	})
}

// responseRecorder is a type that implements the io.ReadCloser interface. This allows the logging handler to
// track certain elements of the request without having to be informed directly by consumers of the request.
type recordedReader struct {
	bytesRead int64
	delegate  io.ReadCloser
}

func (s *recordedReader) Read(p []byte) (int, error) {
	read, err := s.delegate.Read(p)
	s.bytesRead += int64(read)

	return read, err
}

func (s *recordedReader) Close() error {
	return s.delegate.Close()
}

// responseRecorder is a type that implements the http.ResponseWriter interface. This allows the logging handler to
// track certain elements of the response without having to be informed directly by consumers of the response writer.
type responseRecorder struct {
	statusCode   int
	bytesWritten int64
	delegate     http.ResponseWriter
}

func (s *responseRecorder) Header() http.Header {
	return s.delegate.Header()
}

func (s *responseRecorder) Write(buffer []byte) (int, error) {
	if s.statusCode == 0 {
		s.statusCode = http.StatusOK
	}

	written, err := s.delegate.Write(buffer)
	s.bytesWritten += int64(written)

	return written, err
}

func (s *responseRecorder) WriteHeader(statusCode int) {
	s.statusCode = statusCode
	s.delegate.WriteHeader(statusCode)
}

func getSignedRequestDate(request *http.Request) (string, bool) {
	requestDateHeader := request.Header.Get(headers.RequestDate.String())
	return requestDateHeader, requestDateHeader != ""
}

func setSignedRequestFields(request *http.Request, logAttrs *[]slog.Attr) {
	// Log the token ID and request date if the request contains either header
	if requestDateHeader, hasHeader := getSignedRequestDate(request); hasHeader {
		*logAttrs = append(*logAttrs, slog.String("signed_request_date", requestDateHeader))
	}

	if authScheme, schemeParameter, err := parseAuthorizationHeader(request); err == nil {
		switch authScheme {
		case api.AuthorizationSchemeBHESignature:
			if _, err := uuid.FromString(schemeParameter); err == nil {
				*logAttrs = append(*logAttrs, slog.String("token_id", schemeParameter))
			}
		}
	}
}

// LoggingMiddleware is a middleware func that outputs a log for each request-response lifecycle. It includes timestamped
// information organized into fields suitable for searching or parsing.
func LoggingMiddleware(idResolver auth.IdentityResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			var (
				logAttrs       []slog.Attr
				requestContext = ctx.FromRequest(request)
				deadline       time.Time

				loggedResponse = &responseRecorder{
					delegate: response,
				}

				loggedRequestBody = &recordedReader{
					bytesRead: 0,
					delegate:  request.Body,
				}
			)

			// assign a deadline, but only if a valid timeout has been supplied via the prefer header
			timeout, err := RequestWaitDuration(request)
			if err != nil {
				slog.ErrorContext(request.Context(), fmt.Sprintf("Error parsing prefer header for timeout: %v", err))
			} else if timeout > 0 {
				deadline = time.Now().Add(timeout * time.Second)
			}

			// Wrap the request body so that we can tell how much was read
			request.Body = loggedRequestBody

			// Defer the log statement and then serve the request
			defer func() {
				slog.LogAttrs(request.Context(), slog.LevelInfo, fmt.Sprintf("%s %s", request.Method, request.URL.RequestURI()), logAttrs...)

				if !deadline.IsZero() && time.Now().After(deadline) {
					slog.WarnContext(
						request.Context(),
						fmt.Sprintf("%s %s took longer than the configured timeout of %0.f seconds", request.Method, request.URL.RequestURI(), timeout.Seconds()),
					)
				}
			}()

			next.ServeHTTP(loggedResponse, request)

			// Log the token ID and request date if the request contains either header
			setSignedRequestFields(request, &logAttrs)

			// Add the fields that we care about before exiting
			logAttrs = append(logAttrs,
				slog.String("proto", request.Proto),
				slog.String("referer", request.Referer()),
				slog.String("user_agent", request.UserAgent()),
				slog.Int64("request_bytes", loggedRequestBody.bytesRead),
				slog.Int64("response_bytes", loggedResponse.bytesWritten),
				slog.Int("status", loggedResponse.statusCode),
				slog.Duration("elapsed", time.Since(requestContext.StartTime.UTC())),
			)
		})
	}
}
