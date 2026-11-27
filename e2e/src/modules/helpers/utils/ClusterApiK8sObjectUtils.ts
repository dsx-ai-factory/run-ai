/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import axios, { AxiosResponse } from 'axios';
import { config, flowRootState } from '../config/config';

export interface ObjectWrapper {
    object: Record<string, any>;
    retrievedAt: string;
}

const REQUEST_TIMEOUT_MS = 30_000;

export class ClusterApiK8sObjectUtils {
    private baseUrl(): string {
        return `${config.clusterUrl}/cluster-api/api/v2`;
    }

    // cluster-api auth calls Keycloak /userinfo, which accepts OAuth access
    // tokens but rejects OIDC id tokens — so .idToken does not work here.
    private headers(): Record<string, string> {
        return {
            Authorization: `Bearer ${flowRootState.users.system_administrator.accessToken}`,
            Accept: 'application/json'
        };
    }

    private get(path: string): Promise<AxiosResponse<ObjectWrapper>> {
        return axios.get<ObjectWrapper>(`${this.baseUrl()}/${path}`, {
            headers: this.headers(),
            timeout: REQUEST_TIMEOUT_MS,
            validateStatus: () => true
        });
    }

    getWorkloadObject(workloadId: string): Promise<AxiosResponse<ObjectWrapper>> {
        return this.get(`workloads/${workloadId}/object`);
    }

    getPodObject(workloadId: string, podId: string): Promise<AxiosResponse<ObjectWrapper>> {
        return this.get(`workloads/${workloadId}/pods/${podId}/object`);
    }

    getElementObject(
        workloadId: string,
        elementId: string
    ): Promise<AxiosResponse<ObjectWrapper>> {
        return this.get(`workloads/${workloadId}/elements/${elementId}/object`);
    }
}
