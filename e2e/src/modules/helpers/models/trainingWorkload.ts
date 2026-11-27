/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseWorkload, crds, CustomResourceDefinition, objectGenerator } from '..';

export class TrainingWorkload extends BaseWorkload {
    constructor(image: string, namespace: string, name?: string) {
        super(name || objectGenerator.generateNameByPrefix('training-workload'), namespace, image);
    }

    getCrd(): CustomResourceDefinition {
        return crds.runaiTrainingWorkloadCrd;
    }
}
