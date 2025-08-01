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

package ad_test

import (
	"testing"

	ad2 "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
)

func TestSelectGPOContainerCandidateFilter(t *testing.T) {
	var (
		computer = graph.NewNode(0, graph.NewProperties(), ad.Computer)
		group    = graph.NewNode(1, graph.NewProperties().Set(common.SystemTags.String(), ad.AdminTierZero), ad.Group)
		user     = graph.NewNode(2, graph.NewProperties().Set(common.SystemTags.String(), ad.AdminTierZero), ad.User)
	)

	assert.False(t, ad2.SelectGPOContainerCandidateFilter(computer))
	assert.False(t, ad2.SelectGPOTierZeroCandidateFilter(group))
	assert.True(t, ad2.SelectGPOTierZeroCandidateFilter(user))
}
