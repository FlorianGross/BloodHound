// Copyright 2024 Specter Ops, Inc.
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

package model

import (
	"time"

	"github.com/lib/pq"
)

type AnalysisRequestType string

const (
	AnalysisRequestAnalysis AnalysisRequestType = "analysis"
	AnalysisRequestDeletion AnalysisRequestType = "deletion"
)

type AnalysisRequest struct {
	RequestedBy string              `json:"requested_by"`
	RequestType AnalysisRequestType `json:"request_type"`
	RequestedAt time.Time           `json:"requested_at"`

	DeleteAllGraph        bool           `json:"delete_all_graph"`                       // Deletes all nodes and edges in the graph
	DeleteSourcelessGraph bool           `json:"delete_sourceless_graph"`                // Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourceKinds     pq.StringArray `gorm:"type:text[];column:delete_source_kinds"` // Deletes all nodes and edges per kind provided.
}
