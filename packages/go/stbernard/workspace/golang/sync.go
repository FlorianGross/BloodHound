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

package golang

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// TidyModules runs go mod tidy for all module paths passed
func TidyModules(modPath string, env environment.Environment) error {
	var (
		command = "go"
		args    = []string{"mod", "tidy"}
	)

	if _, err := cmdrunner.Run(command, args, modPath, env); err != nil {
		return fmt.Errorf("go mod tidy in %s: %w", modPath, err)
	}

	return nil
}

// SyncWorkspace runs go work sync in the given directory with a given set of environment
// variables
func SyncWorkspace(cwd string, env environment.Environment) error {
	var (
		command = "go"
		args    = []string{"work", "sync"}
	)

	// Skip this if go.work doesn't exist
	if _, err := os.Stat(filepath.Join(cwd, "go.work")); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if _, err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return fmt.Errorf("go work sync: %w", err)
	} else {
		return nil
	}
}
