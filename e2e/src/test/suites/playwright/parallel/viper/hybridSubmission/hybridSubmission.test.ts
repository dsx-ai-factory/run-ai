/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    objectGenerator,
    logger,
    tags,
    projectResourcesUtils,
    operations,
    dataSourcesUtils,
    credentialUtils,
    unleashClient,
    environmentUtils
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import { expect } from 'chai';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { Scope, WorkloadType } from '../../../../../../modules/api/apis/assetsApi';
import { V1ObjectMeta } from '@kubernetes/client-node';
import { SecretType } from '../../../../../../modules/api/apis/k8sObjectsTrackerApi/generated/api';
import { EFeatureToggleKeys } from '../../../../../../modules/helpers/utils/Unleash/featureToggle';

test.describe(
    `Hybrid submission UI, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.viper}`,
    { tag: [tags.services.policy, tags.services.workloads, tags.services.frontend, tags.runOn.parallelCI] },
    function () {
        let project: Project;
        let videoName: string;

        let hybridForm: HybridForm;
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workload: WorkloadPage;
        let workloadName: string;

        const createClusterConfigmap = async (): Promise<V1ObjectMeta> => {
            try {
                const cm = await dataSourcesUtils.createConfigMapResource('cm', project.namespace);
                logger.info(`creating configmap in namespace: ${project.namespace}`);
                return cm;
            } catch (e) {
                logger.error(`failed to create configmap in namespace: ${project.namespace}`);
                throw e;
            }
        };
        const createClusterSecret = async (): Promise<V1ObjectMeta> => {
            try {
                const cm = await credentialUtils.createExistingSecret(
                    SecretType.AccessKey,
                    false,
                    undefined,
                    project.namespace
                );
                logger.info(`creating secret in namespace: ${project.namespace}`);
                return cm;
            } catch (e) {
                logger.error(`failed to create secret in namespace: ${project.namespace}`);
                throw e;
            }
        };

        test.beforeAll('creating project and namespace', async function () {
            browserEngine = new BrowserEngine();
            project = await projectResourcesUtils.createProjectAndNamespace('h-p');
            logger.info(`created project ${project.name} ,namespace: "${project.namespace}"`);
        });

        test.afterAll('deleting project and namespace', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${project.name} an namespace ${project.namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${project.name} an namespace ${project.namespace}`);
                throw error;
            }
        });

        test.beforeEach('start browser', async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.l2_researcher.email,
                flowRootState.users.l2_researcher.password
            );
            workloadName = objectGenerator.generateNameByPrefix('hybrid-w');
            await workload.navigateToWorkloads();
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, 'hybrid-w' + videoName);
            } catch (e) {
                logger.error(e);
            }
        });

        async function fillWorkloadCreationFirstPage(): Promise<void> {
            await workload.openWorkspaceForm();

            await workload.selectClusterFromDropDown(flowRootState.cluster.name);
            await workload.workloadFormSelectProject(project.name);
            await workload.workloadFormSelectTemplate();
            await hybridForm.fillWorkspaceName(workloadName);
            await hybridForm.waitForWorkloadNameCount(workloadName);
            await hybridForm.clickOnContinue(WorkloadType.Workspace);
        }

        test.describe('Positive casual scenarios', function () {
            test('create from scratch', async function () {
                test.setTimeout(210 * 1000);

                videoName = 'hybrid-from-scratch';
                await fillWorkloadCreationFirstPage();
                await hybridForm.fillImageInput('jupyter/scipy-notebook');

                await hybridForm.clickOnSubSectionByLabel(hybridForm.imagePullSection);
                await hybridForm.clickOnRadioOptionByLabel('Always pull the image from the registry');

                await hybridForm.clickOnSubSectionByLabel('Tools');
                await hybridForm.addToolTypeByLabel('Jupyter');

                await hybridForm.clickOnSubSectionByLabel('Runtime settings');
                await hybridForm.clickOnCommandAndArguments();
                await hybridForm.fillCmdInput('start-notebook.sh');
                await hybridForm.fillArgsInput(
                    '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                );

                await hybridForm.clickOnSubSectionByLabel('Security');

                await hybridForm.clickOnSubmit();
                await workload.waitForWorkloadUIStatus(workloadName);
                await workload.deleteWorkload(workloadName);
            });
            test('create from scratch with Jupyter LoadBalancer connection', async function () {
                test.setTimeout(210 * 1000);

                async function validateLoadBalancerService(
                    workloadName: string,
                    namespace: string
                ): Promise<void> {
                    logger.info(`Validating LoadBalancer service for workload: ${workloadName}`);
                    try {
                        if (
                            await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)
                        ) {
                            const job = await kubeCli.getBatchJob(workloadName, namespace);
                            expect(job).to.not.be.undefined;
                        } else {
                            const workloadCrd = await kubeCli.getInteractiveWorkload(
                                workloadName,
                                namespace
                            );
                            expect(workloadCrd).to.not.be.undefined;
                        }

                        const servicesResponse = await kubeCli.CoreV1Api.listNamespacedService({ namespace });
                        const services = servicesResponse.items;

                        const workloadServices = services.filter((svc) => {
                            const svcName = svc.metadata.name;
                            const workloadPrefix = workloadName.split('-')[0];
                            return (
                                (svcName.includes(workloadPrefix) || svcName.includes(workloadName)) &&
                                svc.spec.type === 'LoadBalancer'
                            );
                        });

                        expect(workloadServices).to.have.length.greaterThan(
                            0,
                            `No LoadBalancer service found for workload ${workloadName}`
                        );

                        const loadBalancerService = workloadServices[0];
                        const serviceName = loadBalancerService.metadata.name;

                        expect(loadBalancerService.spec.type).to.equal('LoadBalancer');

                        expect(loadBalancerService.spec.ports).to.have.length.greaterThan(0);
                        const port = loadBalancerService.spec.ports[0];
                        expect(port.port).to.be.a('number');
                        expect(port.targetPort).to.be.a('number');
                    } catch (error) {
                        logger.error(`Failed to validate LoadBalancer service: ${error.message}`);
                        throw error;
                    }
                }

                videoName = 'hybrid-jupyter-loadbalancer';
                await fillWorkloadCreationFirstPage();
                await hybridForm.fillImageInput('jupyter/scipy-notebook');

                await hybridForm.clickOnSubSectionByLabel(hybridForm.imagePullSection);
                await hybridForm.clickOnRadioOptionByLabel('Always pull the image from the registry');

                await hybridForm.clickOnSubSectionByLabel('Tools');
                await hybridForm.addToolTypeByLabel('Jupyter');

                await hybridForm.selectOptionByValue(
                    hybridForm.hybridEnvironmentSection.toolConnectionTypeSelect,
                    'Load Balancer'
                );

                await hybridForm.clickOnSubSectionByLabel('Runtime settings');
                await hybridForm.clickOnCommandAndArguments();
                await hybridForm.fillCmdInput('start-notebook.sh');
                await hybridForm.fillArgsInput(
                    '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                );

                await hybridForm.clickOnSubSectionByLabel('Security');

                await hybridForm.clickOnSubmit();
                await workload.waitForWorkloadUIStatus(workloadName);

                await validateLoadBalancerService(workloadName, project.namespace);

                await workload.deleteWorkload(workloadName);
            });

            test('create from scratch using asset templates', async function () {
                test.setTimeout(210 * 1000);

                videoName = 'hybrid-asset-templates';

                await fillWorkloadCreationFirstPage();

                await hybridForm.openEnvironmentsDrawer();
                await hybridForm.clickOnDrawerItemByLabel((await environmentUtils.presetNames()).jupyter);

                await hybridForm.clearRuntimeSettings();

                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                await hybridForm.clickOnSubmit();
                await workload.waitForWorkloadUIStatus(workloadName);
                await workload.deleteWorkload(workloadName);
            });

            test('create from scratch using asset templates and create another workload using copy & edit', async function () {
                test.setTimeout(210 * 1000);

                videoName = 'hybrid-copy-and-edit';

                // create a workload
                await fillWorkloadCreationFirstPage();

                await hybridForm.openEnvironmentsDrawer();
                await hybridForm.clickOnDrawerItemByLabel((await environmentUtils.presetNames()).jupyter);

                await hybridForm.clearRuntimeSettings();

                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                await hybridForm.clickOnSubmit();
                await workload.waitForWorkloadUIStatus(workloadName);

                // copy and edit the workload recently created
                const copiedWorkloadName = objectGenerator.generateShortNameByPrefix('copy-workspace');
                await workload.toggleWorkload(workloadName);
                await workload.clickOnCopyAndEditWorkload();
                await hybridForm.fillWorkspaceName(copiedWorkloadName);
                await hybridForm.clickOnSubmit();
                await workload.waitForWorkloadUIStatus(copiedWorkloadName);

                await workload.deleteWorkload(workloadName);
                await workload.deleteWorkload(copiedWorkloadName);
            });

            test('submit a hybrid workload with environment variable using configmap', async function () {
                videoName = 'hybrid-env-var-with-configmap';
                const cmInCluster = await createClusterConfigmap();

                // create a workload
                await fillWorkloadCreationFirstPage();

                await hybridForm.fillImageInput('alpine');
                await hybridForm.clickOnSubSectionByLabel('Runtime settings');
                await hybridForm.clickOnCommandAndArguments();
                await hybridForm.fillCmdInput('sleep 1000');
                await hybridForm.clickOnAddEnvVarBtn();
                await hybridForm.fillEnvVarName('some-name');
                await hybridForm.selectEnvVarSource('ConfigMap');
                await hybridForm.selectEnvVarConfigMap(cmInCluster.name);
                await hybridForm.fillEnvVarConfigMapKeyInput('key1');

                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                await hybridForm.clickOnSubmit();
                await workload.waitForWorkloadUIStatus(workloadName);

                await workload.deleteWorkload(workloadName);
            });

            test('submit a hybrid workload with environment variable using secret', async function ({}, testInfo) {
                if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'WLR-backed workspaces do not yet expose the native Initializing status semantics'
                    );
                }

                videoName = 'hybrid-env-var-with-secret';
                const secretInCluster = await createClusterSecret();

                // create a workload
                await fillWorkloadCreationFirstPage();

                await hybridForm.fillImageInput('alpine');
                await hybridForm.clickOnSubSectionByLabel('Runtime settings');
                await hybridForm.clickOnCommandAndArguments();
                await hybridForm.fillCmdInput('sleep 1000');
                await hybridForm.clickOnAddEnvVarBtn();
                await hybridForm.fillEnvVarName('some-name');

                await hybridForm.selectEnvVarSource('Secret');
                await hybridForm.selectEnvVarSecret(secretInCluster.name);
                await hybridForm.selectEnvVarSecretKey('AccessKeyId');

                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                await hybridForm.clickOnSubmit();

                await workload.waitForWorkloadUIStatus(workloadName, 'Initializing');
                await workload.deleteWorkload(workloadName);
            });

            test('submit a hybrid workload with configmap using subpath', async function () {
                videoName = 'hybrid-configmap-subpath';

                const cmInCluster = await createClusterConfigmap();

                const cmAsset = await dataSourcesUtils.createConfigMapAsset({
                    meta: {
                        name: cmInCluster.name,
                        scope: Scope.Project,
                        projectId: project.id
                    },
                    spec: {
                        configMap: cmInCluster.name,
                        mountPath: '/var/path'
                    }
                });

                // create a workload
                await fillWorkloadCreationFirstPage();

                await hybridForm.fillImageInput('alpine');
                await hybridForm.clickOnSubSectionByLabel('Runtime settings');
                await hybridForm.clickOnCommandAndArguments();
                await hybridForm.fillCmdInput('sleep 1000');

                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                await hybridForm.openDataSourcesSection();
                await hybridForm.openDataSourcesDrawer();
                await hybridForm.clickOnDrawerItemByLabel(cmAsset.meta.name, false);
                await hybridForm.clickOnLoadFromDrawer();

                await hybridForm.fillConfigMapSubPathInput('some-subpath');

                await hybridForm.clickOnSubmit();

                await workload.waitForWorkloadUIStatus(workloadName);
                await workload.deleteWorkload(workloadName);
            });
        });
    }
);
