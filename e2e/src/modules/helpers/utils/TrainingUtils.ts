/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../api';
import { Training, TrainingCreationRequestV2, TrainingV2 } from '../../api/apis/assetsApi';
import { Training as TrainingModel } from '../../api/apis/workloadsApi/generated/model/training';
import { flowRootState } from '../config/config';
import {
    computeResourceUtils,
    environmentUtils,
    logger,
    objectGenerator,
    policyUtils,
    trainingUtils,
    waitUtils,
    workloadsCrudUtils
} from '../index';
import { Project } from '../../api/apis/projectsDepartmentsApi';
import { MatchExpressionOperator } from '../../api/apis/workloadsApi/generated/model/matchExpressionOperator';
import http from 'http';
import { AssetKind } from '../../api/apis/assetsApi';
import { WorkspaceSpecSpecAllOfCompute } from '../../api/apis/workloadsApi/generated/api';

export class TrainingUtils {
    async createTraining(
        trainingWorkload: TrainingCreationRequestV2,
        accessToken?: string
    ): Promise<TrainingV2> {
        assetsApi.trainingApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        try {
            const response = await assetsApi.trainingApi.createTrainingV2(trainingWorkload);
            logger.info(`created training ${trainingWorkload.meta.name}`);
            return response.body;
        } catch (e) {
            logger.error(`failed to create training ${trainingWorkload.meta.name} \n${e}`);
            throw e;
        }
    }

    async createTrainingWorkloads(amount: number, project: Project): Promise<TrainingV2[]> {
        async function getTrainingWorkloadPayload(name: string): Promise<TrainingCreationRequestV2> {
            const envName = (await environmentUtils.presetNames()).jupyter;
            return {
                meta: {
                    projectId: project.id.toString(),
                    name: name,
                    clusterId: project.clusterUuid
                },
                assets: {
                    environment: await environmentUtils.getId(envName),
                    compute: await computeResourceUtils.getId('cpu-only')
                },
                specificRunParams: {
                    command: 'sleep',
                    args: 'infinity'
                }
            };
        }

        const trainingWorkloads = [];
        for (let i = 0; i < amount; i++) {
            const name = objectGenerator.generateShortNameByPrefix(`training-${i}`);
            const workloadPayload = await getTrainingWorkloadPayload(name);
            const trainingWorkload = await trainingUtils.createTraining(workloadPayload);
            trainingWorkloads.push(trainingWorkload);
        }
        return trainingWorkloads;
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
        const trainingCreationRequest: TrainingCreationRequestV2 = {
            meta: {
                projectId: projectId,
                name: workloadName,
                clusterId: flowRootState.cluster.uuid
            },
            assets: {
                environment: await environmentUtils.getId(environment),
                compute: await computeResourceUtils.getId(compute),
                datasources: [
                    {
                        id: dataSourceId,
                        kind: AssetKind.ConfigMap
                    }
                ]
            }
        };

        const workload = await this.createTraining(trainingCreationRequest);
        return workload.meta.workloadId;
    }
}
