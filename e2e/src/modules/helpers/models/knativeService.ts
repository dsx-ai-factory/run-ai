/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KubernetesObject } from '@kubernetes/client-node';

export interface KnativeService extends KubernetesObject {
    status: {
        address: {
            url: string;
        };
        url: string;
    };
}

export interface KnativeRevision extends KubernetesObject {
    status: {
        actualReplicas: number;
    };
}
