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

// organize-imports-ignore
import React from 'react';
import { createTheme } from '@mui/material/styles';
import { CssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material';
import { render, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { BrowserRouter } from 'react-router-dom';
import { NotificationsProvider } from './providers';
import { darkPalette } from './constants';
import { SnackbarProvider } from 'notistack';

/**
 * @description SetUpQueryClient takes in stateMaps in the form of an array of objects where each object has a "key" key and a "data" key
 *
 * @param  stateMaps These maps are looped over for hydrating the queryClient with the state that is required for the test(s) the queryClient is being used for
 *
 */
export const SetUpQueryClient = (stateMaps) => {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
                refetchOnMount: false,
                refetchOnWindowFocus: false,
                staleTime: Infinity,
            },
        },
    });

    stateMaps.forEach(({ key, data }) => {
        queryClient.setQueryData(key, data);
    });

    return queryClient;
};

const theme = createTheme(darkPalette);
const defaultTheme = {
    ...theme,
    palette: {
        ...theme.palette,
        neutral: { ...darkPalette.neutral },
        color: { ...darkPalette.color },
        tertiary: { ...darkPalette.tertiary },
    },
};

const createDefaultQueryClient = () => {
    return new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
            },
        },
    });
};

const createProviders = ({ queryClient, route, theme, children }) => {
    window.history.pushState({}, 'Initialize', route);
    return (
        <QueryClientProvider client={queryClient}>
            <StyledEngineProvider injectFirst>
                <ThemeProvider theme={theme}>
                    <NotificationsProvider>
                        <CssBaseline />
                        <BrowserRouter>
                            <SnackbarProvider>{children}</SnackbarProvider>
                        </BrowserRouter>
                    </NotificationsProvider>
                </ThemeProvider>
            </StyledEngineProvider>
        </QueryClientProvider>
    );
};

const customRender = (
    ui,
    { theme = defaultTheme, route = '/', queryClient = createDefaultQueryClient(), ...renderOptions } = {}
) => {
    const AllTheProviders = ({ children }) => createProviders({ queryClient, route, theme, children });
    return render(ui, { wrapper: AllTheProviders, ...renderOptions });
};

const customRenderHook = (
    hook,
    { queryClient = createDefaultQueryClient(), theme = defaultTheme, route = '/', ...renderOptions } = {}
) => {
    const AllTheProviders = ({ children }) => createProviders({ queryClient, route, theme, children });
    return renderHook(hook, { wrapper: AllTheProviders, ...renderOptions });
};

// re-export everything
export * from '@testing-library/react';
// override render and renderHook methods
export { customRender as render, customRenderHook as renderHook };

export const longWait = (cb) => {
    waitFor(cb, { timeout: 10000 });
};
