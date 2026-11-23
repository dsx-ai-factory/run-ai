/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BaseWorkload,
    crds,
    CustomResourceDefinition,
    flowRootState,
    objectGenerator,
    Utils,
    versionUtils
} from '..';
import { config } from '../index';
import { InferenceWorkloadCreationOptions } from '../utils/WorkloadUtils';
import { InferencePhase } from '../utils/WaitUtils';
import { kubeCli } from '../../cli';

export class InferenceWorkload extends BaseWorkload {
    minScale: number;

    maxScale: number;

    initialScale: number;

    activationScale: number;

    scaleToZeroRetentionSeconds: number;

    scaleDownDelaySeconds?: number;

    initializationTimeoutSeconds?: number;

    metric: string;

    metricName?: string;

    target: number;

    targetUtilizationPercentage?: string;

    concurrencyHardLimit?: number;

    mps: boolean;

    servingPortAuthorizationType: string;

    servingPortAuthorizedUsers: string;

    servingPortAuthorizedGroups: string;

    constructor(namespace: string, name?: string, options?: InferenceWorkloadCreationOptions) {
        // using this image as default is important, as it handles SIGTERM properly thus
        // not causing long termination times as might happen with a more naive implementation
        const image = options?.image || config.quickStartImage.inferenceHelloWorld;

        super(name || objectGenerator.generateNameByPrefix('inference-workload'), namespace, image);

        this.gpu = String(options?.gpus || 0);
        this.command = options?.command;
        this.minScale = options?.minScale ?? 1;
        this.maxScale = options?.maxScale ?? 1;
        this.initialScale = options?.initialScale;
        this.activationScale = options?.activationScale;
        this.scaleToZeroRetentionSeconds = options?.scaleToZeroRetentionSeconds;
        this.scaleDownDelaySeconds = options?.scaleDownDelaySeconds;
        this.initializationTimeoutSeconds = options?.initializationTimeoutSeconds;
        this.metric = options?.metric;
        this.metricName = options?.metricName;
        this.target = options?.target;
        this.targetUtilizationPercentage = options?.targetUtilizationPercentage;
        this.concurrencyHardLimit = options?.concurrencyHardLimit;
        this.mps = options?.mps;
        this.largeShm = options?.largeShm;
        this.pvcs = options?.pvcs;
        this.nfs = options?.nfs;
        this.volumes = options?.volumes;
        this.gitSync = options?.gitSync;
        this.configMapVolumes = options?.configMapVolumes;
        this.readinessProbe = options?.readinessProbe;

        this.ports = {
            port1: {
                container: 8080,
                external: Utils.randomInteger(50000, 60000),
                protocol: 'http'
            }
        };

        if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.17')) {
            this.ports['port1'].serviceType = 'ServingPort';
        }

        this.servingPortAuthorizationType = options?.servingPortAuthorizationType;
        this.servingPortAuthorizedUsers = options?.servingPortAuthorizedUsers;
        this.servingPortAuthorizedGroups = options?.servingPortAuthorizedGroups;
        this.workingDir = options?.workingDir;
    }

    getCrd(): CustomResourceDefinition {
        return crds.runaiInferenceWorkloadCrd;
    }

    async validatePhase(expectedPhase: InferencePhase): Promise<boolean> {
        const workload = await kubeCli.getInferenceWorkload(this.name, this.testingSpecificFields.namespace);

        if (!workload || workload['status'] === undefined || workload['status']['phase'] === undefined) {
            return false;
        }

        return workload['status']['phase'] === expectedPhase;
    }
}
