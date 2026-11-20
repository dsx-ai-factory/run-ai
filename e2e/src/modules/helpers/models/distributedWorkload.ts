/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseWorkload, crds, CustomResourceDefinition, objectGenerator } from '..';

class MpiJobInfo {
    slotsPerWorker?: number;
}

class PyTorchJobInfo {
    minReplicas?: number;

    maxReplicas?: number;
}

export class DistributedWorkload extends BaseWorkload {
    workers: number;

    jobType: string;

    mpiJob?: MpiJobInfo;

    pyTorchJob?: PyTorchJobInfo;

    constructor(image: string, namespace: string, name?: string) {
        super(name || objectGenerator.generateNameByPrefix('distr-workload'), namespace, image);
    }

    getCrd(): CustomResourceDefinition {
        return crds.runaiDistributedWorkloadCrd;
    }

    getFlatFields(): string[] {
        const result = super.getFlatFields();
        result.push('jobType');
        return result;
    }
}

export const DistributedJobType = {
    MpiJobType: 'MPIJob',
    PyTorchJobType: 'PyTorchJob',
    TensorJobType: 'TFJob',
    XGBoostJobType: 'XGBoostJob'
};
