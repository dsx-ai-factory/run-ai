/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KubernetesObject } from '@kubernetes/client-node';

export interface LeaderWorkerSet extends KubernetesObject {
    status: {
        replicas: number;
        readyReplicas: number;
    };
}
