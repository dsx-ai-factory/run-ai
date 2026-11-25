/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../api';
import { Training } from '../../api/apis/assetsApi';
import { TrainingCreationRequest, WorkspaceSpecSpecAllOfCompute } from '../../api/apis/workloadsApi/generated/api';
import { Training as TrainingModel } from '../../api/apis/workloadsApi/generated/model/training';
import { flowRootState } from '../config/config';
import {
    computeResourceUtils,
    environmentUtils,
    logger,
    objectGenerator,
    policyUtils,
    waitUtils,
    workloadsCrudUtils
} from '../index';
import { Project } from '../../api/apis/projectsDepartmentsApi';
import { MatchExpressionOperator } from '../../api/apis/workloadsApi/generated/model/matchExpressionOperator';
import http from 'http';

export class TrainingUtils {
    async createTraining(
        trainingWorkload: TrainingCreationRequest,
        accessToken?: string
    ): Promise<TrainingModel> {
        try {
            const response = await workloadsCrudUtils.submitTraining(trainingWorkload, accessToken);
            logger.info(`created training ${trainingWorkload.name}`);
            return response.body;
        } catch (e) {
            logger.error(`failed to create training ${trainingWorkload.name} \n${e}`);
            throw e;
        }
    }

    async createTrainingWorkloads(
        countOrNames: number | string[],
        project: Project,
        configuration?: Pick<TrainingCreationRequest, 'assets' | 'spec'>,
        accessToken?: string
    ): Promise<TrainingModel[]> {
        const names = Array.isArray(countOrNames)
            ? countOrNames
            : Array.from({ length: countOrNames }, (_, i) =>
                  objectGenerator.generateShortNameByPrefix(`training-${i}`)
              );
        if (names.length === 0) {
            return [];
        }

        const workloadConfiguration = configuration ?? {
            assets: {
                environment: await environmentUtils.getId((await environmentUtils.presetNames()).jupyter),
                compute: await computeResourceUtils.getId('cpu-only')
            },
            spec: { command: 'sleep', args: 'infinity' }
        };

        const workloads: TrainingModel[] = [];
        for (const name of names) {
            const workload = await this.createTraining(
                {
                    name,
                    projectId: project.id.toString(),
                    clusterId: project.clusterUuid,
                    ...workloadConfiguration
                },
                accessToken
            );
            workloads.push(workload);
        }
        return workloads;
    }

    async list(): Promise<Array<Training>> {
        assetsApi.trainingApi.accessToken = flowRootState.users.editor.accessToken;
        return (await assetsApi.trainingApi.listTrainings(flowRootState.cluster.uuid)).body.entries;
    }

    async waitForTrainingToBeDeleted(name: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const trainingsNames = (await this.list()).map((training) => training.meta.name);
                return !trainingsNames.includes(name);
            },
            'waiting for training to be deleted',
            true,
            15
        );
    }

    async createTrainingV2(
        projectId: string,
        workloadName: string,
        trainingSpec: {
            image: string;
            command: string;
            compute?: WorkspaceSpecSpecAllOfCompute;
            backoffLimit?: number;
        },
        excludeNodeNames?: string[],
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: TrainingModel }> {
        if (excludeNodeNames) {
            trainingSpec['nodeAffinityRequired'] = policyUtils.getNodeSelectorTerm(
                MatchExpressionOperator.NotIn,
                excludeNodeNames
            ).nodeAffinityRequired;
        }

        if (accessToken) {
            workloadsCrudUtils.accessToken = accessToken;
        } else {
            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
        }

        return workloadsCrudUtils.createTraining(
            workloadName,
            false,
            flowRootState.cluster.uuid,
            projectId,
            trainingSpec
        );
    }

    async createArbitraryTrainingWithConfigmap(
        workloadName: string,
        projectId: string,
        dataSourceId: string,
        environment = 'tensorboard',
        compute = 'cpu-only'
    ): Promise<string> {
        const trainingCreationRequest: TrainingCreationRequest = {
            projectId: projectId,
            name: workloadName,
            clusterId: flowRootState.cluster.uuid,
            assets: {
                environment: await environmentUtils.getId(environment),
                compute: await computeResourceUtils.getId(compute),
                datasources: [dataSourceId]
            }
        };

        const workload = await this.createTraining(trainingCreationRequest);
        return workload.workloadId;
    }
}
