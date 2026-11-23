/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    assetsApiV2Utils,
    computeResourceUtils,
    crds,
    environmentUtils,
    flowRootState,
    logger,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import {
    PodAffinityType,
    TolerationEffect,
    TolerationOperator,
    WorkspaceCreationRequestV2
} from '../../../../../../modules/api/apis/assetsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { expect } from 'chai';
import { description, tag } from 'allure-js-commons';
import { test } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

test.describe(`Workload with tolerations and topology, tags: ${tags.teams.viper}`, () => {
    let project: Project;

    async function getWorkloadCreationPayload(workloadName: string): Promise<WorkspaceCreationRequestV2> {
        return {
            meta: {
                name: workloadName,
                projectId: project.id.toString(),
                clusterId: flowRootState.cluster.uuid
            },
            assets: {
                environment: await environmentUtils.getId('tensorboard'),
                compute: await computeResourceUtils.getId('cpu-only')
            },
            specificRunParams: {
                tolerations: [
                    {
                        key: 'key-1',
                        operator: TolerationOperator.Exists,
                        effect: TolerationEffect.NoSchedule
                    },
                    {
                        key: 'key-2',
                        operator: TolerationOperator.Equal,
                        value: 'value-2',
                        effect: TolerationEffect.NoExecute
                    }
                ],
                podAffinity: {
                    type: PodAffinityType.Preferred,
                    key: 'topology-1'
                }
            }
        };
    }

    test.beforeAll('creating project', async function () {
        project = await projectResourcesUtils.createProjectAndNamespace('project');
        workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
    });
    test.afterAll('deleting project', async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
        } catch (e) {
            logger.error('Failed to cleanup project', e);
        }
    });

    function addInfo(): void {
        tag(`${tags.teams.viper}`);
        description(
            'Flow:\n' +
                '1. Submit a workload with tolerations and topology.\n' +
                '2. Check:\n' +
                '* workload yaml contains the correct data\n'
        );
    }

    test('Create a workload with tolerations and topology', async () => {
        addInfo();
        const workloadName = 'workload-with-tolerations-and-topology';

        await test.step('Create a workload with tolerations and topology', async function () {
            const workloadCreationPayload = await getWorkloadCreationPayload(workloadName);
            await assetsApiV2Utils.createWorkspace(workloadCreationPayload);
        });

        await test.step('Check workload yaml', async () => {
            if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)) {
                await waitUtils.waitForBatchJobToExist(workloadName, project.namespace);
                const job = await kubeCli.getBatchJob(workloadName, project.namespace);
                const podSpec = job.spec?.template.spec;
                const preferredAffinity =
                    podSpec?.affinity?.podAffinity?.preferredDuringSchedulingIgnoredDuringExecution?.[0];

                expect(preferredAffinity?.podAffinityTerm.topologyKey).to.equal('topology-1');
                expect(podSpec?.tolerations).to.deep.include.members([
                    {
                        effect: 'NoSchedule',
                        key: 'key-1',
                        operator: 'Exists'
                    },
                    {
                        effect: 'NoExecute',
                        key: 'key-2',
                        operator: 'Equal',
                        value: 'value-2'
                    }
                ]);
            } else {
                await waitUtils.waitAndGetCrd(
                    crds.runaiInteractiveWorkloadCrd,
                    workloadName,
                    project.namespace
                );
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const workspaceCr: any = await kubeCli.getInteractiveWorkload(
                    workloadName,
                    project.namespace
                );

                expect(workspaceCr.spec.podAffinityTopology.value).to.equal('topology-1');
                expect(workspaceCr.spec.podAffinitySchedulingRule.value).to.equal(
                    'preferredDuringSchedulingIgnoredDuringExecution'
                );
                expect(workspaceCr.spec.tolerations.items).to.deep.equal({
                    toleration: {
                        value: {
                            effect: 'NoSchedule',
                            operator: 'Exists',
                            tolerationKey: 'key-1'
                        }
                    },
                    'toleration-1': {
                        value: {
                            effect: 'NoExecute',
                            operator: 'Equal',
                            tolerationKey: 'key-2',
                            tolerationValue: 'value-2'
                        }
                    }
                });
            }
        });
    });
});
