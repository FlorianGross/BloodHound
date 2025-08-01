// Copyright 2025 Specter Ops, Inc.
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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, waitFor } from '../../test-utils';
import { apiClient } from '../../utils';
import DetailsRoot from './DetailsRoot';

const handlers = [
    rest.get('/api/v2/features', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'tier_management_engine',
                        enabled: true,
                    },
                ],
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const assetGroupSpy = vi.spyOn(apiClient, 'getAssetGroupTags');

describe('DetailsRoot', () => {
    it('calls getAssetGroupTags for topTagId', () => {
        render(<DetailsRoot />);
        waitFor(() => {
            expect(assetGroupSpy).toHaveBeenCalled();
        });
    });
});
