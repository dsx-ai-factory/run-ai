/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseWorkload } from './baseWorkload';
import { InteractiveWorkload } from './interactiveWorkload';
import { TrainingWorkload } from './trainingWorkload';
import { config, InferenceWorkload } from '..';

export enum WorkloadType {
    Training = 1,
    Interactive,
    Inference
}

export type WorkloadSpecUnsafe = {
    type: WorkloadType;

    // Any other field is optional
    [key: string]: unknown;
};

export class WorkloadFactoryUnsafe {
    private static readonly requiredFields = ['namespace'];

    static Create(workloadSpec: WorkloadSpecUnsafe): BaseWorkload {
        WorkloadFactoryUnsafe.validateWorkloadSpec(workloadSpec);
        workloadSpec = WorkloadFactoryUnsafe.setDefaults(workloadSpec);

        let workload: BaseWorkload;

        const image = workloadSpec.image as string;
        const namespace = workloadSpec.namespace as string;
        const name = workloadSpec.name as string;

        switch (workloadSpec.type) {
            case WorkloadType.Training:
                workload = new TrainingWorkload(image, namespace, name);
                break;
            case WorkloadType.Interactive:
                workload = new InteractiveWorkload(image, namespace, name);
                break;
            case WorkloadType.Inference:
                workload = new InferenceWorkload(namespace, name);
                break;
            default:
                throw new Error('Unsupported workload type');
        }

        for (const field in workloadSpec) {
            if (field === 'type') {
                continue;
            }

            workload[field] = workloadSpec[field];
        }

        return workload;
    }

    private static setDefaults(workload: WorkloadSpecUnsafe): WorkloadSpecUnsafe {
        WorkloadFactoryUnsafe.setWorkloadTypeDefaults(workload);

        return workload;
    }

    private static setWorkloadTypeDefaults(workload: WorkloadSpecUnsafe): void {
        switch (workload.type) {
            case WorkloadType.Training:
                workload.image = workload.image || config.quickStartImage.training;
                break;
            case WorkloadType.Interactive:
                workload.image = workload.image || config.quickStartImage.interactive;
                break;
            case WorkloadType.Inference:
                workload.image = workload.image || config.quickStartImage.inference;
                break;
            default:
                throw new Error('Unsupported workload type');
        }
    }

    private static validateWorkloadSpec(workload: WorkloadSpecUnsafe): void {
        for (const field of WorkloadFactoryUnsafe.requiredFields) {
            if (!workload[field]) {
                throw new Error(`Workload missing required field ${field}`);
            }
        }
    }
}
