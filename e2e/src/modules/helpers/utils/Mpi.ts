/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { kubeCli } from '../../cli';

export async function isMpiInstalled(): Promise<boolean> {
    const existingDeployment = await kubeCli.AppsV1Api.listNamespacedDeployment({
        namespace: 'mpi-operator'
    });
    for (const deploymentObj of existingDeployment.items) {
        if (deploymentObj.spec.replicas >= 1) {
            return true;
        }
    }
    return false;
}
