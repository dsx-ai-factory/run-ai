/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    credentialUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    runaiHealthChecker,
    tags,
    templateUtils,
    User,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../../modules/helpers';
import { Project } from '../../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadPage } from '../../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { expect } from 'chai';
import { WorkloadType } from '../../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test } from '@playwright/test';
import {
    Inference,
    InferenceSpecSpec,
    InferenceTemplate
} from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { InferenceFlatUIUtil } from '../../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import {
    GenericSecret,
    GenericSecretCreationRequest,
    KeyValuePair,
    NgcApiKey,
    NgcApiKeyCreationRequest,
    Scope,
    ServingPortAccessAuthorizationTypeEnum,
    ServingPortProtocol,
    UidGidSource
} from '../../../../../../../../modules/api/apis/assetsApi';
import {
    ScopeType,
    WorkloadType as WorkloadTypeApi
} from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { NimElementValidationUtils } from '../../../../../../../../modules/ui/utils/inference/NimElementValidationUtils';

test.describe(
    `Inference NIM based on template UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'inference-nim-from-template';
        let project: Project;
        let traceName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let inferenceName: string;
        let inferenceWorkload: Inference;
        let template: InferenceTemplate;

        const nimModelName = 'nvidia/llm-nim';
        const image = `nvcr.io/nim/${nimModelName}`;
        const apiKey = 'myApiKey';
        const profileOrHash = 'myProfile';
        const baseNimEnvironmentVariables = [
            {
                name: 'OUTLINES_CACHE_DIR',
                value: '/tmp/outlines'
            },
            {
                name: 'NIM_SERVER_PORT',
                value: '8000'
            },
            {
                name: 'NIM_JSONL_LOGGING',
                value: '1'
            },
            {
                name: 'NIM_LOG_LEVEL',
                value: 'INFO'
            }
        ];

        const currentUser: User = flowRootState.users.system_administrator;

        async function createTemplate(
            metaConfig: {
                workloadType: WorkloadTypeApi;
                scopeType: ScopeType;
                scopeId: string;
                templateName: string;
                description?: string;
            },
            templateSpec: InferenceSpecSpec
        ): Promise<InferenceTemplate> {
            try {
                const hasNimLabel = templateSpec.labels?.some(
                    (label) => label.name === 'run.ai/nim-workload'
                );

                if (!hasNimLabel) {
                    templateSpec.labels = [
                        ...(templateSpec.labels || []),
                        { name: 'run.ai/nim-workload', value: 'true' }
                    ];
                }

                const response = await templateUtils.createWorkloadTemplateV2(metaConfig, templateSpec);
                return response.body as InferenceTemplate;
            } catch (error) {
                logger.error(
                    `failed to create template ${metaConfig.templateName}. Error: ${JSON.stringify(error)}`
                );
                throw error;
            }
        }

        async function createNgcApiKeySecret(
            secretName: string,
            projectId: string,
            ngcApiKey: string
        ): Promise<NgcApiKey> {
            try {
                const ngcApiKeySecret: NgcApiKeyCreationRequest = {
                    meta: {
                        name: secretName,
                        description: null,
                        scope: Scope.Project,
                        projectId: parseInt(projectId),
                        clusterId: undefined,
                        departmentId: null
                    },
                    spec: { ngcApiKey }
                };

                const ngcApiKeyResponse = await credentialUtils.createNgcApiKeyAsset(ngcApiKeySecret);
                await credentialUtils.waitForNgcApiKeyAssetToBeReady(ngcApiKeyResponse);
                return ngcApiKeyResponse;
            } catch (error) {
                logger.error(
                    `failed to create ngc api key secret ${secretName}. Error: ${JSON.stringify(error)}`
                );
                throw error;
            }
        }

        async function initBrowser(): Promise<BrowserEngine> {
            // Initialize browser once for all tests with video recording AND tracing support
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
            await auth.signIn(currentUser.email, currentUser.password);

            return browserEngine;
        }

        test.beforeAll('before all create inference based on nim template UI', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('project-e2e-for-nim');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });

            logger.info(`created project ${projectName}`);

            browserEngine = await initBrowser();

            logger.info('Browser initialized and user signed in');
        });

        test.afterAll('after all create inference based on nim template UI', async ({}, testInfo) => {
            // Close browser once after all tests and save video
            if (browserEngine && browserEngine.browser) {
                try {
                    await browserEngine.stopAndArchive(testInfo, suiteName);
                    logger.info('Browser closed and video saved');
                } catch (error) {
                    logger.error(`Error closing browser: ${error}`);
                }
            }

            if (!(await runaiHealthChecker.isSelfHostedCluster())) {
                // This condition should be removed when fixing problem with saas envs with cluster version.
                return;
            }

            try {
                await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id, 120);
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

        test.beforeEach('before each create inference based on nim template UI', async function () {
            if (!browserEngine || !browserEngine.browser) {
                browserEngine = await initBrowser();
            }

            inferenceName = objectGenerator.generateNameByPrefix('inference');

            // Generate trace name from test title
            traceName = objectGenerator.generateNameByPrefix(suiteName);

            // Start tracing for this test (context and page stay alive)
            await browserEngine.startTracing(traceName);

            // Navigate to workloads page for each test
            await workload.navigateToWorkloads();
        });

        test.afterEach('after each create inference based on nim template UI', async ({}, testInfo) => {
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

            if (template) {
                await templateUtils.deleteWorkloadTemplateV2(template.meta.id, WorkloadTypeApi.Inference);
            }
            template = undefined;

            logger.debug(`inference ${inferenceName} deleted!`);

            // Stop and archive tracing for this test (browser, context, and page remain alive)
            await browserEngine.stopAndArchiveTracing(testInfo, traceName);
        });

        async function createGenericSecret(
            secretName: string,
            scope: Scope,
            projectId: string | undefined,
            clusterId: string | undefined,
            keyValuePairs: Array<KeyValuePair>
        ): Promise<GenericSecret> {
            const genericSecret: GenericSecretCreationRequest = {
                meta: {
                    name: secretName,
                    description: null,
                    scope: scope,
                    projectId: parseInt(projectId),
                    clusterId: clusterId,
                    departmentId: null
                },
                spec: { existingSecretName: null, keyValuePairs }
            };

            const genericSecretResponse = await credentialUtils.createGenericSecretAsset(genericSecret);
            await credentialUtils.waitForGenericSecretAssetToBeReady(genericSecretResponse);
            return genericSecretResponse;
        }

        supportedFrom(SupportedVersions.V2_23)(function () {
            test('should check that the environment variables does not display the value for api key', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // check that the environment variables hide the api key
                const nimElementValidationUtils = new NimElementValidationUtils(browserEngine.currentPage);
                await nimElementValidationUtils.expectEnvironmentVariables([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NGC_API_KEY',
                        value: ''
                    }
                ]);

                await inferenceFlatUIUtil.fillEnvironmentData({
                    runtimeSettings: {
                        environmentVariablesValuesByName: new Map([['NGC_API_KEY', apiKey]])
                    }
                });

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                logger.info('test inference template passed!');
            });

            test('should create new nim inference using an existing nim template with profile automatically and access model from api key', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName,
                    environmentSection: {
                        runtimeSettings: {
                            environmentVariablesValuesByName: new Map([['NGC_API_KEY', apiKey]])
                        }
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.image).to.be.equals(templateSpec.image);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(
                    templateSpec.servingPort.container
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(
                    templateSpec.servingPort.protocol
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.equals(
                    templateSpec.servingPort.authorizationType
                );
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.deep.equals(
                    templateSpec.servingPort.authorizedGroups
                );
                expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.be.equals(
                    templateSpec.compute.gpuDevicesRequest
                );
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equals(
                    templateSpec.autoscaling.minReplicas
                );
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equals(
                    templateSpec.autoscaling.maxReplicas
                );
                expect(inferenceWorkload.spec.autoscaling.metric).to.be.equals(
                    templateSpec.autoscaling.metric
                );
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.be.equals(
                    templateSpec.autoscaling.metricThreshold
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.equals(
                    templateSpec.autoscaling.scaleToZeroRetentionSeconds
                );
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables.map((envVar) => ({
                        ...envVar,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    })),
                    {
                        name: 'NGC_API_KEY',
                        value: undefined,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                logger.info('test inference template passed!');
            });

            test('should create new nim inference using an existing nim template with profile manually and access model from api key', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        },
                        {
                            name: 'NIM_MODEL_PROFILE',
                            value: profileOrHash
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName,
                    environmentSection: {
                        runtimeSettings: {
                            environmentVariablesValuesByName: new Map([['NGC_API_KEY', apiKey]])
                        }
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.image).to.be.equals(templateSpec.image);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(
                    templateSpec.servingPort.container
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(
                    templateSpec.servingPort.protocol
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.equals(
                    templateSpec.servingPort.authorizationType
                );
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.deep.equals(
                    templateSpec.servingPort.authorizedGroups
                );
                expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.be.equals(
                    templateSpec.compute.gpuDevicesRequest
                );
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equals(
                    templateSpec.autoscaling.minReplicas
                );
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equals(
                    templateSpec.autoscaling.maxReplicas
                );
                expect(inferenceWorkload.spec.autoscaling.metric).to.be.equals(
                    templateSpec.autoscaling.metric
                );
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.be.equals(
                    templateSpec.autoscaling.metricThreshold
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.equals(
                    templateSpec.autoscaling.scaleToZeroRetentionSeconds
                );

                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables.map((envVar) => ({
                        ...envVar,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    })),
                    {
                        name: 'NGC_API_KEY',
                        value: undefined,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'NIM_MODEL_PROFILE',
                        value: profileOrHash,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                logger.info('test inference template passed!');
            });

            test('should create new nim inference using an existing nim template with model profile automatically and access model from pvc storage only', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const pvcStorage = {
                    claimName: 'my-pvc-project-we5ks',
                    existingPvc: true,
                    path: '/test/sample',
                    readOnly: false
                };

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NIM_CACHE_PATH',
                            value: pvcStorage.path
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    storage: {
                        pvc: [pvcStorage]
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.image).to.be.equals(templateSpec.image);
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables.map((envVar) => ({
                        ...envVar,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    })),
                    {
                        name: 'NIM_CACHE_PATH',
                        value: pvcStorage.path,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceWorkload.spec.storage.pvc[0]).to.be.exist;
                expect(inferenceWorkload.spec.storage.pvc[0].claimName).to.be.equal(pvcStorage.claimName);
                expect(inferenceWorkload.spec.storage.pvc[0].existingPvc).to.be.true;
                expect(inferenceWorkload.spec.storage.pvc[0].path).to.be.equal(pvcStorage.path);
                expect(inferenceWorkload.spec.storage.pvc[0].readOnly).to.be.false;
                logger.info('test inference template passed!');
            });

            test('should create new nim inference using an existing nim template with model profile automatically and access model from nfs storage only', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const nfsStorage = {
                    mountPath: '/home',
                    path: '/test/sample',
                    readOnly: false,
                    server: 'myserver'
                };
                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NIM_CACHE_PATH',
                            value: nfsStorage.path
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    storage: {
                        nfs: [nfsStorage]
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.image).to.be.equals(templateSpec.image);
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables.map((envVar) => ({
                        ...envVar,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    })),
                    {
                        name: 'NIM_CACHE_PATH',
                        value: nfsStorage.path,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceWorkload.spec.storage.nfs[0]).to.be.exist;
                expect(inferenceWorkload.spec.storage.nfs[0].mountPath).to.be.equal(nfsStorage.mountPath);
                expect(inferenceWorkload.spec.storage.nfs[0].path).to.be.equal(nfsStorage.path);
                expect(inferenceWorkload.spec.storage.nfs[0].readOnly).to.be.false;
                expect(inferenceWorkload.spec.storage.nfs[0].server).to.be.equal(nfsStorage.server);
                logger.info('test inference template passed!');
            });

            test('should create nim inference using an existing nim template with access model from credentials', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const secretName = objectGenerator.generateNameByPrefix('generic-secret');
                const genericSecret: GenericSecret = await createGenericSecret(
                    secretName,
                    Scope.Project,
                    project.id,
                    undefined,
                    [{ key: 'NGC_API_KEY', value: 'someToken' }]
                );

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            secret: {
                                name: `genericsecret-${secretName}`,
                                key: 'NGC_API_KEY'
                            }
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.image).to.be.equals(templateSpec.image);
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables.map((envVar) => ({
                        ...envVar,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    })),
                    {
                        name: 'NGC_API_KEY',
                        value: undefined,
                        secret: {
                            name: `genericsecret-${genericSecret.meta.name}`,
                            key: 'NGC_API_KEY'
                        },
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                logger.info('test inference template passed!');
            });

            test('should test the inference serving port using an existing nim template', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
                };

                const servingPort = {
                    protocol: ServingPortProtocol.Grpc,
                    container: 3131,
                    authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                    authorizedGroups: ['group1', 'group2']
                };

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort,
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName,
                    environmentSection: {
                        runtimeSettings: {
                            environmentVariablesValuesByName: new Map([['NGC_API_KEY', apiKey]])
                        }
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.container);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal(servingPort.protocol);
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.equal(
                    servingPort.authorizationType
                );
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.deep.equal(
                    servingPort.authorizedGroups
                );

                logger.info('test inference template environment section passed!');
            });

            test('should test the inference compute section elements using an existing template', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
                };

                const autoscaling = {
                    minReplicas: 0,
                    maxReplicas: 5,
                    metric: 'concurrency',
                    metricThreshold: 50,
                    scaleToZeroRetentionSeconds: 300
                };

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        }
                    ],
                    servingPort: {
                        container: 3131
                    },
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    compute: {
                        gpuDevicesRequest: 2,
                        cpuCoreRequest: 1,
                        cpuCoreLimit: 2,
                        largeShmRequest: true
                    },
                    autoscaling
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName,
                    environmentSection: {
                        runtimeSettings: {
                            environmentVariablesValuesByName: new Map([['NGC_API_KEY', apiKey]])
                        }
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // compute section
                expect(inferenceWorkload.spec.compute.cpuCoreRequest).to.equal(1);
                expect(inferenceWorkload.spec.compute.cpuCoreLimit).to.equal(2);
                expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.equal(2);

                // auto scaling
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.equal(autoscaling.minReplicas);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.equal(autoscaling.maxReplicas);
                expect(inferenceWorkload.spec.autoscaling.metric).to.equal('concurrency');
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.equal(
                    autoscaling.metricThreshold
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(
                    autoscaling.scaleToZeroRetentionSeconds
                );

                logger.info('test inference template compute section passed!');
            });

            test('should check that the nim label is added to the template', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image,
                    environmentVariables: [
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage,
                        capabilities: [],
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    compute: {
                        largeShmRequest: true
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                await inferenceFlatUIUtil.fillEnvironmentData({
                    runtimeSettings: {
                        environmentVariablesValuesByName: new Map([['NGC_API_KEY', apiKey]])
                    }
                });

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.labels.length).to.be.equal(1);
                expect(inferenceWorkload.spec.labels[0].name).to.be.equal('run.ai/nim-workload');
                expect(inferenceWorkload.spec.labels[0].value).to.be.equal('true');

                logger.info('test inference template passed!');
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test('should create nim inference using an existing nim template with access model from ngc api key credentials', async function () {
                    const metaConfig = {
                        workloadType: WorkloadTypeApi.Inference,
                        scopeType: ScopeType.Project,
                        scopeId: project.id.toString(),
                        templateName: objectGenerator.generateNameByPrefix('nim-template')
                    };

                    const secretName = objectGenerator.generateNameByPrefix('ngc-api-key');
                    const ngcApiKey: NgcApiKey = await createNgcApiKeySecret(
                        secretName,
                        project.id,
                        'someToken'
                    );

                    const templateSpec: InferenceSpecSpec = {
                        image,
                        environmentVariables: [
                            ...baseNimEnvironmentVariables,
                            {
                                name: 'NGC_API_KEY',
                                secret: {
                                    name: ngcApiKey.clusterInfo.resources[0].name,
                                    key: 'NGC_API_KEY'
                                }
                            }
                        ],
                        security: {
                            uidGidSource: UidGidSource.FromTheImage,
                            capabilities: [],
                            runAsGid: 1000,
                            runAsUid: 1000
                        },
                        servingPort: {
                            protocol: ServingPortProtocol.Http,
                            container: 8000
                        },
                        autoscaling: {
                            minReplicas: 1,
                            maxReplicas: 1
                        },
                        compute: {
                            largeShmRequest: true
                        }
                    };

                    template = await createTemplate(metaConfig, templateSpec);
                    expect(template).to.not.be.undefined;
                    logger.debug(`template: ${JSON.stringify(template)}`);

                    await inferenceFlatUIUtil.createFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        templateName: metaConfig.templateName
                    });

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    expect(inferenceWorkload.spec.image).to.be.equals(templateSpec.image);
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        ...baseNimEnvironmentVariables.map((envVar) => ({
                            ...envVar,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        })),
                        {
                            name: 'NGC_API_KEY',
                            value: undefined,
                            secret: {
                                name: ngcApiKey.clusterInfo.resources[0].name,
                                key: 'NGC_API_KEY'
                            },
                            userCredential: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);
                    await credentialUtils.deleteNgcApiKeyAsset(ngcApiKey.meta.id);
                    logger.info('test inference template passed!');
                });
            });

            test.describe('create NIM workload with models from Hugging Face', function () {
                const huggingFaceModelNameWithoutToken = 'microsoft/DialoGPT-small';
                const huggingFaceModelNameWithToken = 'meta-llama/Llama-2-13b-hf';
                const hfToken = process.env.HUGGING_FACE_TOKEN;
                const nimImageForHuggingFaceModel = 'nvcr.io/nim/nvidia/llm-nim:1.12';

                test('should create new nim inference using an existing nim template with not a gated model from Hugging Face', async function () {
                    const metaConfig = {
                        workloadType: WorkloadTypeApi.Inference,
                        scopeType: ScopeType.Project,
                        scopeId: project.id.toString(),
                        templateName: objectGenerator.generateNameByPrefix('nim-template')
                    };
                    const environmentVariables = [
                        {
                            name: 'NIM_SERVER_PORT',
                            value: '8000'
                        },
                        {
                            name: 'NIM_HTTP_API_PORT',
                            value: '8000'
                        },
                        {
                            name: 'NIM_MODEL_NAME',
                            value: `hf://${huggingFaceModelNameWithoutToken}`
                        }
                    ];

                    const templateSpec: InferenceSpecSpec = {
                        image: nimImageForHuggingFaceModel,
                        environmentVariables,
                        security: {
                            uidGidSource: UidGidSource.FromTheImage,
                            capabilities: [],
                            runAsGid: 1000,
                            runAsUid: 1000
                        },
                        servingPort: {
                            protocol: ServingPortProtocol.Http,
                            container: 8000
                        },
                        autoscaling: {
                            minReplicas: 1,
                            maxReplicas: 1
                        },
                        compute: {
                            largeShmRequest: true
                        }
                    };

                    template = await createTemplate(metaConfig, templateSpec);
                    expect(template).to.not.be.undefined;
                    logger.debug(`template: ${JSON.stringify(template)}`);

                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        templateName: metaConfig.templateName
                    });

                    // check that the environment variables hide the api key
                    const nimElementValidationUtils = new NimElementValidationUtils(
                        browserEngine.currentPage
                    );
                    await nimElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                    // create inference
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    logger.info('test inference template passed!');
                });

                test('should create new nim inference using an existing nim template with gated model from Hugging Face', async function () {
                    const metaConfig = {
                        workloadType: WorkloadTypeApi.Inference,
                        scopeType: ScopeType.Project,
                        scopeId: project.id.toString(),
                        templateName: objectGenerator.generateNameByPrefix('nim-template')
                    };
                    const environmentVariables = [
                        {
                            name: 'NIM_SERVER_PORT',
                            value: '8000'
                        },
                        {
                            name: 'NIM_HTTP_API_PORT',
                            value: '8000'
                        },
                        {
                            name: 'NIM_MODEL_NAME',
                            value: `hf://${huggingFaceModelNameWithToken}`
                        }
                    ];

                    const templateSpec: InferenceSpecSpec = {
                        image: nimImageForHuggingFaceModel,
                        environmentVariables: [...environmentVariables, { name: 'HF_TOKEN', value: hfToken }],
                        security: {
                            uidGidSource: UidGidSource.FromTheImage,
                            capabilities: [],
                            runAsGid: 1000,
                            runAsUid: 1000
                        },
                        servingPort: {
                            protocol: ServingPortProtocol.Http,
                            container: 8000
                        },
                        autoscaling: {
                            minReplicas: 1,
                            maxReplicas: 1
                        },
                        compute: {
                            largeShmRequest: true
                        }
                    };

                    template = await createTemplate(metaConfig, templateSpec);
                    expect(template).to.not.be.undefined;
                    logger.debug(`template: ${JSON.stringify(template)}`);

                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        templateName: metaConfig.templateName
                    });

                    // check that the environment variables hide the api key
                    const nimElementValidationUtils = new NimElementValidationUtils(
                        browserEngine.currentPage
                    );
                    await nimElementValidationUtils.expectEnvironmentVariables([
                        ...environmentVariables,
                        { name: 'HF_TOKEN', value: '' }
                    ]);

                    await inferenceFlatUIUtil.fillEnvironmentData({
                        runtimeSettings: {
                            environmentVariablesValuesByName: new Map([['HF_TOKEN', hfToken]])
                        }
                    });

                    // create inference
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                    logger.info('test inference template passed!');
                });

                test('should create new nim inference using an existing nim template with gated model from Hugging Face with shared credentials', async function () {
                    const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                    const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                        secretName,
                        parseInt(project.id),
                        [{ key: 'HF_TOKEN', value: hfToken }]
                    );

                    const metaConfig = {
                        workloadType: WorkloadTypeApi.Inference,
                        scopeType: ScopeType.Project,
                        scopeId: project.id.toString(),
                        templateName: objectGenerator.generateNameByPrefix('nim-template')
                    };
                    const environmentVariables = [
                        {
                            name: 'NIM_SERVER_PORT',
                            value: '8000'
                        },
                        {
                            name: 'NIM_HTTP_API_PORT',
                            value: '8000'
                        },
                        {
                            name: 'NIM_MODEL_NAME',
                            value: `hf://${huggingFaceModelNameWithToken}`
                        },
                        {
                            name: 'HF_TOKEN',
                            secret: { name: genericSecret.clusterInfo.resources[0].name, key: 'HF_TOKEN' }
                        }
                    ];

                    const templateSpec: InferenceSpecSpec = {
                        image: nimImageForHuggingFaceModel,
                        environmentVariables,
                        security: {
                            uidGidSource: UidGidSource.FromTheImage,
                            capabilities: [],
                            runAsGid: 1000,
                            runAsUid: 1000
                        },
                        servingPort: {
                            protocol: ServingPortProtocol.Http,
                            container: 8000
                        },
                        autoscaling: {
                            minReplicas: 1,
                            maxReplicas: 1
                        },
                        compute: {
                            largeShmRequest: true
                        }
                    };

                    template = await createTemplate(metaConfig, templateSpec);
                    expect(template).to.not.be.undefined;
                    logger.debug(`template: ${JSON.stringify(template)}`);

                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        templateName: metaConfig.templateName
                    });

                    // check that the environment variables hide the api key
                    const nimElementValidationUtils = new NimElementValidationUtils(
                        browserEngine.currentPage
                    );
                    await nimElementValidationUtils.expectEnvironmentVariablesWithSharedCredentials(environmentVariables);

                    // create inference
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                    logger.info('test inference template passed!');
                });
            });
        });
    }
);
