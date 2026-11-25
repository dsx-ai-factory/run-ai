/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    runaiHealthChecker,
    tags,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { PUBLIC_VLLM_IMAGE } from '../../../../../../modules/helpers/config/const';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { expect } from 'chai';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test } from '@playwright/test';
import {
    IInferenceType,
    IServingPortAccessOptions
} from '../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { Inference } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { User } from '../../../../../../modules/helpers/config/types';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    EImagePullPolicyOptions,
    IAutoScalingOptions,
    IDataSourceOptions,
    IServingPortOptions
} from '../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { InferenceFlatUIUtil } from '../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(
    `Inferences CI UI, tags: ${tags.tool.ui} ${tags.runOn.parallelCI} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        let project: Project;
        let videoName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let inferenceName: string;
        let inferenceWorkload: Inference;

        const currentUser: User = flowRootState.users.system_administrator;

        test.beforeAll('before all create inference workload CI UI', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('p-inf');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });

            logger.info(`created project ${projectName}`);
        });

        test.afterAll('after all create inference workload CI UI', async function () {
            if (!(await runaiHealthChecker.isSelfHostedCluster())) {
                // This condition should be removed when fixing problem with saas envs with cluster version.
                return;
            }

            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                return;
            }

            try {
                await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id);
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                logger.info(
                    `deleted workloads and project ${project.name} and namespace ${project.status.namespace}`
                );
            } catch (error) {
                logger.error(
                    `failed to delete workloads and project ${project.name} and namespace ${project.status.namespace}. Error: ${JSON.stringify(error)}`
                );
            }
        });

        test.beforeEach('before each create inference workload CI UI', async function () {
            await browserEngine.start();
            inferenceName = objectGenerator.generateNameByPrefix('inference');
            auth = new AuthPage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
            await auth.signIn(currentUser.email, currentUser.password);
            await workload.navigateToWorkloads();
        });

        test.afterEach('after each create inference workload CI UI', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'inferences-' + videoName);

            logger.debug(`waiting for workload ${inferenceName} to be deleted`);

            // delete inference
            if (inferenceWorkload) {
                await workloadsUtilsV2.deleteWorkload(
                    inferenceWorkload.workloadId,
                    WorkloadType.INFERENCE,
                    flowRootState.users.system_administrator.accessToken
                );
            }
            inferenceWorkload = undefined;

            logger.debug(`inference ${inferenceName} deleted!`);
        });

        supportedFrom(SupportedVersions.V2_19)(function () {
            test('should create new flat inference including all fields', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'inference any-workload submission path does not support exposedUrls yet'
                );
                videoName = 'should-create-new-flat-inference-including-all-fields';

                const servingPort: IServingPortOptions = {
                    protocol: 'HTTP',
                    port: 4000,
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 0,
                    maxReplicas: 4,
                    thresholdValue: 50,
                    autoScalingPeriod: '1m'
                };

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                const dataSourceOptions: IDataSourceOptions = {
                    nfs: dataSourceName
                };

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        image: PUBLIC_VLLM_IMAGE,
                        imagePullPolicy: EImagePullPolicyOptions.Never,
                        tool: 'Jupyter',
                        runtimeSettings: {
                            cmd: 'start-notebook.sh',
                            args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                        },
                        servingPort
                    },
                    computeSection: {
                        assetName: 'cpu-only',
                        autoScaling
                    },
                    dataSourceOptions,
                    generalSectionOptions: {
                        annotations: [
                            { name: 'ann1', value: 'ann-val1' },
                            { name: 'ann2', value: 'ann-val2' }
                        ],
                        labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                console.log(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // environment section
                expect(inferenceWorkload.spec.image).to.equal(PUBLIC_VLLM_IMAGE);
                expect(inferenceWorkload.spec.imagePullPolicy).to.equal('Never');
                expect(inferenceWorkload.spec.command).to.equal('start-notebook.sh');
                expect(inferenceWorkload.spec.args).to.equal(
                    '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                );

                expect(inferenceWorkload.spec.exposedUrls).to.deep.equal([
                    {
                        container: 8888,
                        toolType: 'jupyter-notebook',
                        toolName: 'Jupyter',
                        authorizationType: undefined,
                        authorizedGroups: undefined,
                        authorizedUsers: undefined,
                        url: undefined,
                        exclude: undefined,
                        name: undefined
                    }
                ]);

                // serving port
                expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal(
                    'authorizedUsersOrGroups'
                );
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;

                // compute
                expect(inferenceWorkload.spec.compute.cpuCoreRequest).to.equal(0.1);
                expect(inferenceWorkload.spec.compute.cpuMemoryRequest).to.equal('100M');

                // auto scaling
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.equal(0);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceWorkload.spec.autoscaling.metric).to.equal('throughput');
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.equal(
                    autoScaling.thresholdValue
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(60); // 1m

                // storage
                expect(inferenceWorkload.spec.storage.nfs).to.deep.equal([
                    {
                        path: '/test/sample',
                        readOnly: false,
                        server: 'test-server',
                        mountPath: '/home/local/data',
                        exclude: undefined,
                        name: undefined
                    }
                ]);

                // general section
                expect(
                    inferenceWorkload.spec.annotations.find(
                        (annotation) => annotation.name === 'ann1' && annotation.value === 'ann-val1'
                    )
                ).to.not.be.undefined;
                expect(
                    inferenceWorkload.spec.annotations.find(
                        (annotation) => annotation.name === 'ann2' && annotation.value === 'ann-val2'
                    )
                ).to.not.be.undefined;
                expect(
                    inferenceWorkload.spec.labels.find(
                        (label) => label.name === 'lbl1' && label.value === 'lbl-val1'
                    )
                ).to.not.be.undefined;

                logger.debug(`inference ${inferenceName} created successfully!`);
            });
        });
    }
);
