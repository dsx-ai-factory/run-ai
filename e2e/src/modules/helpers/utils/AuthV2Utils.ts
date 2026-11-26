/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    settings,
    waitUtils
} from '..';
import { identityManagerApi, projectsDepartmentsApi } from '../../api';
import { User } from '../../api/apis/identityManagerApi/generated/model/user';
import axios, { AxiosResponse } from 'axios';
import { saasTenantName } from '../config/config';

export class AuthV2Utils {
    /**
     * Waits until a new user's permissions are live before tests use their tokens.
     *
     * Freshly-created users can get transient 403s until their access rules propagate
     * through the authorization layer. Probe a representative permission
     * (Editor's `department:read` via GET departments) and wait until it resolves.
     * Best-effort: on timeout, warn and proceed — never fail the run.
     */
    async waitForPermissionsPropagation(timeoutSeconds = 60, delaySeconds = 2): Promise<void> {
        const editor = flowRootState.users?.editor;
        if (!editor?.accessToken) {
            logger.warn(
                'waitForPermissionsPropagation: Editor user is not initialized, skipping propagation gate'
            );
            return;
        }
        projectsDepartmentsApi.departmentsApi.accessToken = editor.accessToken;
        const clusterUuid = flowRootState.cluster.uuid;

        try {
            await waitUtils.waitForCondition(
                async () => {
                    try {
                        await projectsDepartmentsApi.departmentsApi.getDepartments(clusterUuid);
                        return true;
                    } catch (error) {
                        const statusCode = error?.statusCode ?? error?.response?.statusCode;
                        if (statusCode !== 403) {
                            logger.warn(
                                `waitForPermissionsPropagation: unexpected error (status ${statusCode}) while probing permissions: ${error}`
                            );
                        }
                        return false;
                    }
                },
                'new user permissions to propagate (Editor department:read)',
                true,
                timeoutSeconds,
                delaySeconds
            );
            logger.debug('Editor permissions propagated');
        } catch {
            logger.warn(
                `waitForPermissionsPropagation: Editor permissions did not propagate within ${timeoutSeconds}s; proceeding anyway`
            );
        }
    }

    getRoleNameInK8s(roleName: string): string {
        return roleName.replace(/[^a-zA-Z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '').toLowerCase();
    }

    async getLoginScreenURL(
        accessToken: string,
        authenticationFlow?: string,
        redirectUrl?: string
    ): Promise<number> {
        identityManagerApi.loginApi.accessToken = accessToken;
        const getLoginScreenURL = await identityManagerApi.loginApi.getLoginScreenURL(
            accessToken,
            authenticationFlow,
            redirectUrl
        );
        return getLoginScreenURL.response.statusCode;
    }

    async getUsers(accessToken: string): Promise<User[]> {
        identityManagerApi.usersApi.accessToken = accessToken;
        return (await identityManagerApi.usersApi.getUsers()).body;
    }

    async sendWorkloadAuthAuthRequest(
        clientId: string,
        redirectUri?: string,
        responseType?: string
    ): Promise<AxiosResponse> {
        let authUrl = `${config.appUrl}/auth/realms/runai/protocol/openid-connect/auth`;
        if (!settings.isSelfHosted) {
            authUrl = `${config.appUrl}/auth/realms/${saasTenantName}/protocol/openid-connect/auth`;
        }
        const params: { client_id: string; redirect_uri?: string; response_type?: string } = {
            client_id: clientId
        };

        if (redirectUri) {
            params.redirect_uri = redirectUri;
        }
        if (responseType) {
            params.response_type = responseType;
        }
        try {
            return await axios.get(authUrl, { params }); // Return the response if successful
        } catch (error) {
            if (error.response) {
                return error.response; // Return the response with the error status and details
            }
            throw new Error('Request failed without a response');
        }
    }

    async connectToJupyterWithAppCredentials(
        jupyterUrl: string,
        accessToken: string
    ): Promise<{
        baseUrl: string;
        client: {
            get: (endpoint: string) => Promise<AxiosResponse>;
            post: (endpoint: string, data?: any) => Promise<AxiosResponse>;
            put: (endpoint: string, data?: any) => Promise<AxiosResponse>;
            delete: (endpoint: string) => Promise<AxiosResponse>;
        };
        apiEndpoints: {
            contents: string;
            kernels: string;
            sessions: string;
            terminals: string;
            status: string;
        };
    }> {
        logger.info(`Connecting to Jupyter at: ${jupyterUrl}`);
        logger.info('Using application access token for authentication');
        const headers = {
            Authorization: `Bearer ${accessToken}`,
            'Content-Type': 'application/json',
            Accept: 'application/json'
        };
        const client = {
            get: async (endpoint: string): Promise<AxiosResponse> => {
                const url = `${jupyterUrl}${endpoint}`;
                logger.debug(`GET ${url}`);
                try {
                    return await axios.get(url, { headers });
                } catch (error) {
                    logger.error(`Failed to GET ${url}: ${error.message}`);
                    throw error;
                }
            },

            post: async (endpoint: string, data?: any): Promise<AxiosResponse> => {
                const url = `${jupyterUrl}${endpoint}`;
                logger.debug(`POST ${url}`);
                try {
                    return await axios.post(url, data, { headers });
                } catch (error) {
                    logger.error(`Failed to POST ${url}: ${error.message}`);
                    throw error;
                }
            },

            put: async (endpoint: string, data?: any): Promise<AxiosResponse> => {
                const url = `${jupyterUrl}${endpoint}`;
                logger.debug(`PUT ${url}`);
                try {
                    return await axios.put(url, data, { headers });
                } catch (error) {
                    logger.error(`Failed to PUT ${url}: ${error.message}`);
                    throw error;
                }
            },

            delete: async (endpoint: string): Promise<AxiosResponse> => {
                const url = `${jupyterUrl}${endpoint}`;
                logger.debug(`DELETE ${url}`);
                try {
                    return await axios.delete(url, { headers });
                } catch (error) {
                    logger.error(`Failed to DELETE ${url}: ${error.message}`);
                    throw error;
                }
            }
        };

        const apiEndpoints = {
            contents: '/api/contents', // File/notebook operations
            kernels: '/api/kernels', // Kernel management
            sessions: '/api/sessions', // Session management
            terminals: '/api/terminals', // Terminal access
            status: '/api/status' // Jupyter server status
        };

        logger.info('Jupyter client created successfully');
        logger.info('Available API endpoints:');
        Object.entries(apiEndpoints).forEach(([name, endpoint]) => {
            logger.info(`${name}: ${jupyterUrl}${endpoint}`);
        });

        let connectionTestPassed = false;
        const maxRetries = 60;
        const retryDelay = 2000;

        for (let attempt = 1; attempt <= maxRetries; attempt++) {
            try {
                logger.info(` Connection test attempt ${attempt}/${maxRetries}`);
                const statusResponse = await client.get(apiEndpoints.status);
                logger.info(` Connection test successful - Jupyter status: ${statusResponse.status}`);
                connectionTestPassed = true;
                break;
            } catch (error) {
                const status = error.response?.status;
                const transient = !status || status >= 500 || status === 401;
                logger.warn(` Connection test failed on attempt ${attempt}: ${error.message} (status=${status})`);

                if (status === 401) {
                    logger.info(' Got 401 — auth proxy may still be initializing, will retry...');
                }

                if (!transient) {
                    logger.error(` Non-retryable status ${status}, aborting`);
                    throw error;
                }
                if (attempt < maxRetries) {
                    logger.info(` Waiting ${retryDelay / 1000} seconds before retry...`);
                    await new Promise((resolve) => setTimeout(resolve, retryDelay));
                } else {
                    logger.error(' All connection attempts failed. Jupyter may still be starting up.');
                }
            }
        }

        if (!connectionTestPassed) {
            logger.warn(
                ' Connection test failed after all retries. Client created but connection not verified.'
            );
        }

        return {
            baseUrl: jupyterUrl,
            client,
            apiEndpoints
        };
    }
}
