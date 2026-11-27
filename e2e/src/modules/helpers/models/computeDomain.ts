/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KubernetesObject } from '@kubernetes/client-node';

export interface ComputeDomain extends KubernetesObject {
    spec: {
        channel: {
            allocationMode?: 'All' | 'Single';
            resourceClaimTemplate: {
                name: string;
            };
        };
        numNodes: number;
    };
    status?: {
        nodes?: Array<{
            cliqueID: string;
            index?: number;
            ipAddress: string;
            name: string;
            status?: 'Ready' | 'NotReady';
        }>;
        status: 'Ready' | 'NotReady';
    };
}
