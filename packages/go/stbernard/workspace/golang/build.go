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

package golang

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/git"
)

// BuildMainPackages builds all main packages for the given module
func BuildMainPackages(workRoot string, modPath string, env environment.Environment) error {
	var (
		err      error
		version  semver.Version
		buildDir = filepath.Join(workRoot, "dist") + string(filepath.Separator)
	)

	if version, err = git.ParseLatestVersionFromTags(workRoot, env); err != nil {
		slog.Warn(fmt.Sprintf("Failed to parse version from git tags, falling back to environment variable: %v", err))
		parsedVersion, err := semver.NewVersion(env[environment.VersionVarName])
		if err != nil {
			return fmt.Errorf("error parsing version from environment variable: %w", err)
		}
		version = *parsedVersion
	}

	slog.Info(fmt.Sprintf("Building for version %s", version.Original()))

	if err := buildModuleMainPackages(buildDir, modPath, version, env); err != nil {
		return fmt.Errorf("build main packages: %w", err)
	}

	return nil
}

func buildModuleMainPackages(buildDir string, modPath string, version semver.Version, env environment.Environment) error {
	var (
		wg   sync.WaitGroup
		errs []error
		mu   sync.Mutex

		command             = "go"
		majorString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/cmd/api/src/version.majorVersion=%d'", version.Major())
		minorString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/cmd/api/src/version.minorVersion=%d'", version.Minor())
		patchString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/cmd/api/src/version.patchVersion=%d'", version.Patch())
		prereleaseString    = fmt.Sprintf("-X 'github.com/specterops/bloodhound/cmd/api/src/version.prereleaseVersion=%s'", version.Prerelease())
		ldflagArgComponents = []string{majorString, minorString, patchString}
	)

	if version.Prerelease() != "" {
		ldflagArgComponents = append(ldflagArgComponents, prereleaseString)
	}

	args := []string{"build", "-ldflags", strings.Join(ldflagArgComponents, " "), "-o", buildDir}

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("list module packages: %w", err)
	} else {
		for _, pkg := range packages {
			if pkg.Name == "main" && !strings.Contains(pkg.Dir, "plugin") {
				wg.Add(1)
				go func(p GoPackage) {
					defer wg.Done()

					if _, err := cmdrunner.Run(command, args, p.Dir, env); err != nil {
						mu.Lock()
						errs = append(errs, fmt.Errorf("go build for package %s: %w", p.Import, err))
						mu.Unlock()
					}

					slog.Info(fmt.Sprintf("Built package %s", p.Import))
				}(pkg)
			}
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}
