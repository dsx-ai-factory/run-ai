/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    computeResourceUtils,
    credentialUtils,
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    runaiHealthChecker,
    tags,
    templateUtils,
    User,
    versionUtils,
    workloadsUtilsV2
} from '../../../../../../../../modules/helpers';
import { Project } from '../../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { test, TestInfo } from '@playwright/test';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { WorkloadTemplatePage } from '../../../../../../../../modules/ui/runaiV2/pages/WorkloadTemplatePage';
import { IInferenceTemplateMeta } from '../../../../../../../../modules/helpers/models/inferenceTemplateUI.models';
import {
    ECredentialsSource,
    ECredentialsSourceType,
    EModelSourceForNim,
    IAutoScalingOptions,
    IComputeOptions,
    IDataSourceOptions,
    IEnvironmentOptions,
    IInferenceType,
    IPVCOptions,
    IServingPortAccessOptions,
    IServingPortOptions
} from '../../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadType';
import { expect } from 'chai';
import { InferenceTemplateUIUtil } from '../../../../../../../../modules/ui/utils/inference/inferenceTemplateUI.util';
import { InferenceTemplate } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceTemplate';
import {
    EnvironmentVariable,
    GenericSecret,
    GenericSecretCreationRequest,
    KeyValuePair,
    NgcApiKey,
    NgcApiKeyCreationRequest,
    PVCAsset,
    PVCAssetSpec,
    PVCCreationRequest,
    PvcVolumeMode,
    Scope,
    ScopeType,
    ServingPortAccessAuthorizationTypeEnum,
    ServingPortProtocol
} from '../../../../../../../../modules/api/apis/assetsApi';
import { NimElementValidationUtils } from '../../../../../../../../modules/ui/utils/inference/NimElementValidationUtils';
import { InferenceSpecSpec } from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(
    `Inferences Template Nim UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'inference-nim-template';
        let project: Project;
        let traceName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloadTemplate: WorkloadTemplatePage;
        let inferenceTemplateUIUtil: InferenceTemplateUIUtil;
        let templateName: string;
        let templateMeta: IInferenceTemplateMeta;

        let nimElementValidationUtils: NimElementValidationUtils;

        const nimModelName = 'nvidia/llm-nim';
        const apiKey = 'myApiKey';
        const profileOrHash = 'myProfile';
        const NIM_SERVER_PORT = 'NIM_SERVER_PORT';
        const baseNimEnvironmentVariables = [
            {
                name: 'OUTLINES_CACHE_DIR',
                value: '/tmp/outlines',
                secret: undefined,
                configMap: undefined,
                podFieldRef: undefined,
                userCredential: undefined,
                description: undefined,
                exclude: undefined
            },
            {
                name: NIM_SERVER_PORT,
                value: '8000',
                secret: undefined,
                configMap: undefined,
                podFieldRef: undefined,
                userCredential: undefined,
                description: undefined,
                exclude: undefined
            },
            {
                name: 'NIM_JSONL_LOGGING',
                value: '1',
                secret: undefined,
                configMap: undefined,
                podFieldRef: undefined,
                userCredential: undefined,
                description: undefined,
                exclude: undefined
            },
            {
                name: 'NIM_LOG_LEVEL',
                value: 'INFO',
                secret: undefined,
                configMap: undefined,
                podFieldRef: undefined,
                userCredential: undefined,
                description: undefined,
                exclude: undefined
            }
        ];

        const currentUser: User = flowRootState.users.system_administrator;

        async function initBrowser(): Promise<BrowserEngine> {
            // Initialize browser once for all tests with video recording AND tracing support
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloadTemplate = new WorkloadTemplatePage(browserEngine.currentPage);
            inferenceTemplateUIUtil = new InferenceTemplateUIUtil(browserEngine.currentPage);

            await auth.signIn(currentUser.email, currentUser.password);

            return browserEngine;
        }

        test.beforeAll('before all create inference template based on Nim UI', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('project-e2e-nim-inf-template');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });

            logger.info(`created project ${projectName}`);

            browserEngine = await initBrowser();

            logger.info('Browser initialized and user signed in');
        });

        test.afterAll('after all create inference template based on Nim UI', async ({}, testInfo) => {
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

            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
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

        test.beforeEach('before each create inference template based on Nim UI', async function () {
            if (!browserEngine || !browserEngine.browser) {
                browserEngine = await initBrowser();
            }

            templateName = objectGenerator.generateNameByPrefix('inference-template');

            // Generate trace name from test title
            traceName = objectGenerator.generateNameByPrefix(suiteName);

            // Start tracing for this test (context and page stay alive)
            await browserEngine.startTracing(traceName);

            // Initialize page objects for this test
            nimElementValidationUtils = new NimElementValidationUtils(browserEngine.currentPage);

            // Navigate to templates page for each test
            await workloadTemplate.navigateToWorkloadTemplates();

            templateMeta = {
                projectName: project.name,
                inferenceType: IInferenceType.Nim,
                templateName: templateName
            };
        });

        test.afterEach('after each create inference template based on Nim UI', async ({}, testInfo) => {
            logger.debug(`waiting for template ${templateName} to be deleted`);

            if (!templateName) {
                logger.warn('template name is not set, skipping cleanup');
                return;
            }

            const inferenceTemplate: InferenceTemplate | undefined = await templateUtils.getTemplateV2ByName(
                WorkloadType.Inference,
                templateName
            );

            if (!inferenceTemplate) {
                logger.warn(`template ${templateName} not found, skipping cleanup`);
                return;
            }
            await templateUtils.deleteWorkloadTemplateV2(inferenceTemplate.meta.id, WorkloadType.Inference);

            logger.debug(`inference template ${templateName} deleted!`);

            // Stop and archive tracing for this test (browser, context, and page remain alive)
            await browserEngine.stopAndArchiveTracing(testInfo, traceName);
        });

        function updateNimServerPort(
            envVars: Array<EnvironmentVariable>,
            port: number
        ): Array<EnvironmentVariable> {
            return envVars.map((env) => {
                if (env.name === NIM_SERVER_PORT) {
                    return { ...env, value: port.toString() };
                }
                return env;
            });
        }

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

        function expectNimTemplate(
            inferenceTemplate: InferenceTemplate,
            templateName: string,
            servingPort: IServingPortOptions,
            environmentVariables: Array<EnvironmentVariable>,
            checkNfs: boolean = true
        ): void {
            expect(inferenceTemplate).to.not.be.undefined;
            logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

            expect(inferenceTemplate.meta.name).to.equal(templateName);
            expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
            expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
            expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
            expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;
            expect(inferenceTemplate.spec.environmentVariables).to.deep.equal(environmentVariables);
            expect(inferenceTemplate.spec.servingPort.container).to.equal(servingPort.port);
            expect(inferenceTemplate.spec.servingPort.authorizationType).to.equal('authorizedUsersOrGroups');
            expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
            expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

            if (checkNfs) {
                expect(inferenceTemplate.spec.storage.nfs).to.not.be.undefined;
                const nfs = inferenceTemplate.spec.storage.nfs[0];
                expect(nfs).to.not.be.undefined;
                expect(nfs.path).to.equal('/test/sample');
                expect(nfs.readOnly).to.be.false;
                expect(nfs.mountPath).to.equal('/home/local/data');
                expect(nfs.server).to.equal('test-server');
            }

            const nimLabel = inferenceTemplate.spec.labels.find(
                (label) => label.name === 'run.ai/nim-workload'
            );
            expect(nimLabel).to.not.be.undefined;
            expect(nimLabel.value).to.be.equal('true');
        }

        type ExpectedInferenceEnvVar =
            | { name: string; value: string }
            | { name: string; secret: { name: string; key: string } };

        function matchesExpectedInferenceEnv(a: EnvironmentVariable, exp: ExpectedInferenceEnvVar): boolean {
            if (a.name !== exp.name) {
                return false;
            }
            if ('secret' in exp && exp.secret) {
                return a.secret?.name === exp.secret.name && a.secret?.key === exp.secret.key;
            }
            return a.value === (exp as { name: string; value: string }).value;
        }
        /**
         * Asserts workload env vars contain every expected entry and that counts match.
         * Each item is either a plain { name, value } or { name, secret: { name, key } } for secret-backed vars.
         * Use when the API returns full EnvironmentVariable objects (extra fields) but tests assert minimal shapes.
         */
        function expectInferenceSpecEnvironmentVariables(
            actual: Array<EnvironmentVariable> | null | undefined,
            expected: Array<ExpectedInferenceEnvVar>
        ): void {
            const label = 'inferenceTemplate.spec.environmentVariables';

            expect(actual, `${label}: expected an array, got undefined`).to.not.be.undefined;
            expect(actual, `${label}: expected an array, got null`).to.not.be.null;
            expect(
                actual!.length,
                `${label}: length mismatch — expected ${expected.length}, got ${actual!.length}. ` +
                    `expected=${JSON.stringify(expected)} actual=${JSON.stringify(actual)}`
            ).to.equal(expected.length);

            for (const exp of expected) {
                const found = actual!.some((a) => matchesExpectedInferenceEnv(a, exp));
                const detail =
                    'secret' in exp && exp.secret
                        ? `expected secret name=${JSON.stringify(exp.secret.name)} key=${JSON.stringify(exp.secret.key)}`
                        : `expected value ${JSON.stringify((exp as { name: string; value: string }).value)}`;
                expect(
                    found,
                    `${label}: missing or wrong entry for ${exp.name} (${detail}). ` +
                        `actual=${JSON.stringify(actual)}`
                ).to.be.true;
            }
        }

        async function createTemplate(
            metaConfig: {
                workloadType: WorkloadType;
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

        async function createPvcAsset(
            name: string,
            projectId: string,
            spec: PVCAssetSpec
        ): Promise<PVCAsset> {
            const pvc: PVCCreationRequest = {
                meta: {
                    name: name,
                    scope: Scope.Project,
                    projectId: parseInt(projectId)
                },
                spec
            };

            const pvcAsset = await dataSourcesUtils.createPvcAsset(pvc);
            logger.info(`created pvc asset ${pvcAsset.meta.name}`);

            await dataSourcesUtils.waitForPvcAssetToBeReady(
                pvcAsset.meta.name,
                pvcAsset.meta.id,
                parseInt(projectId),
                'Inference'
            );
            logger.info(`created pvc asset ${pvcAsset.meta.name} is ready`);
            return pvcAsset;
        }

        supportedFrom(SupportedVersions.V2_23)(function () {
            test('should create new empty nim inference template with default values only', async function () {
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal(
                    baseNimEnvironmentVariables
                );
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal('http');
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.imagePullPolicy).to.be.equal('IfNotPresent');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new empty nim inference template with default values only and with nim label', async function () {
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal(
                    baseNimEnvironmentVariables
                );
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;

                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.imagePullPolicy).to.be.equal('IfNotPresent');
                expect(inferenceTemplate.spec.labels.length).to.be.equal(1);
                expect(inferenceTemplate.spec.labels[0].name).to.be.equal('run.ai/nim-workload');
                expect(inferenceTemplate.spec.labels[0].value).to.be.equal('true');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with model profile automatically and access model from api key', async function () {
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with model profile manually and access model from api key', async function () {
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        profileOrHash
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NIM_MODEL_PROFILE',
                        value: profileOrHash,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with model profile automatically and access model from storage only', async function () {
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        fromStorageOnly: true
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NIM_CACHE_PATH',
                        value: '/test/sample',
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;
                expect(inferenceTemplate.spec.storage.nfs).to.deep.equal([
                    {
                        path: '/test/sample',
                        readOnly: false,
                        server: 'test-server',
                        mountPath: '/home/local/data',
                        name: undefined,
                        exclude: undefined
                    }
                ]);
                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with model profile manually and access model from storage only', async function () {
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        profileOrHash,
                        fromStorageOnly: true
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NIM_MODEL_PROFILE',
                        value: profileOrHash,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'NIM_CACHE_PATH',
                        value: '/test/sample',
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;
                expect(inferenceTemplate.spec.storage.nfs).to.deep.equal([
                    {
                        path: '/test/sample',
                        readOnly: false,
                        server: 'test-server',
                        mountPath: '/home/local/data',
                        name: undefined,
                        exclude: undefined
                    }
                ]);
                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with access model from api key and storage', async function () {
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'NIM_CACHE_PATH',
                        value: '/test/sample',
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;
                expect(inferenceTemplate.spec.storage.nfs).to.deep.equal([
                    {
                        path: '/test/sample',
                        readOnly: false,
                        server: 'test-server',
                        mountPath: '/home/local/data',
                        name: undefined,
                        exclude: undefined
                    }
                ]);
                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with access model from credentials', async function () {
                const shouldSelectSourceType = versionUtils.isVersionSupported(
                    flowRootState.cluster.version,
                    SupportedVersions.V2_23
                );

                const secretName = objectGenerator.generateNameByPrefix('generic-secret');
                const genericSecret: GenericSecret = await createGenericSecret(
                    secretName,
                    Scope.Project,
                    project.id,
                    undefined,
                    [{ key: 'NGC_API_KEY', value: 'someToken' }]
                );

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        credentials: {
                            source: ECredentialsSource.Shared,
                            sourceType: shouldSelectSourceType
                                ? ECredentialsSourceType.GenericSecret
                                : undefined,
                            secretName: secretName,
                            sourceDisabled: true
                        }
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
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
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with model profile automatically and access model from api key and autoscaling', async function () {
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    },
                    computeSection: {
                        autoScaling: {
                            minReplicas: 2,
                            maxReplicas: 4,
                            thresholdValue: 97
                        }
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(2);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(4);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.equal('throughput');
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.equal(97);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with model profile automatically and access model from api key and serving port access to specific users', async function () {
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        servingPort: {
                            access: {
                                option: IServingPortAccessOptions.Users
                            }
                        }
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // NIM specific data
                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.security.runAsUid).to.be.equal(1000);
                expect(inferenceTemplate.spec.security.runAsGid).to.be.equal(1000);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceTemplate.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.equal(
                    'authorizedUsersOrGroups'
                );
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.deep.equal([
                    currentUser.email
                ]);
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with default serving port details', async function () {
                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                const defaultServingPort = 8000;
                const defaultServingPortProtocol = 'http';
                expect(inferenceTemplate.spec.servingPort.container).to.equal(defaultServingPort);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(defaultServingPortProtocol);
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create nim template with custom serving port protocol and port', async function () {
                const servingPort = {
                    port: 4000,
                    protocol: 'gRPC'
                };

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        servingPort
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                expect(inferenceTemplate.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new nim inference from scratch with compute', async function () {
                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 2,
                    maxReplicas: 4,
                    thresholdValue: 50
                };

                const computeSection: IComputeOptions = {
                    gpu: {
                        devicesInput: '3'
                    },
                    autoScaling
                };

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    },
                    computeSection
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(autoScaling.minReplicas);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceTemplate.spec.autoscaling.metric).to.equal('throughput');
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.equal(
                    autoScaling.thresholdValue
                );
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new nim inference template with compute with custom auto-scaling period', async function () {
                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 0,
                    maxReplicas: 4,
                    thresholdValue: 50,
                    autoScalingPeriod: '1m'
                };

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '3'
                        },
                        autoScaling
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // compute section
                expect(inferenceTemplate.spec.compute.cpuCoreRequest).to.equal(0);
                expect(inferenceTemplate.spec.compute.cpuMemoryRequest).to.equal('0M');
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.equal(3);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(0);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceTemplate.spec.autoscaling.metric).to.equal('throughput');
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.equal(
                    autoScaling.thresholdValue
                );
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(60); // 1m

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new nim inference template using an existing compute asset', async function () {
                const computeResourceName = objectGenerator.generateShortNameByPrefix('cr-two-gpus', 5);
                const computeCreationRequest = {
                    meta: {
                        name: computeResourceName,
                        scope: Scope.Cluster,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 0.2,
                        cpuMemoryRequest: '200M',
                        gpuDevicesRequest: 2
                    }
                };

                const computeAsset = await computeResourceUtils.create(computeCreationRequest);
                logger.debug(`computeAsset created: ${JSON.stringify(computeAsset)}`);

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 2,
                    maxReplicas: 2
                };

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    },
                    computeSection: {
                        assetName: computeResourceName,
                        autoScaling
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // compute section
                expect(inferenceTemplate.spec.compute.cpuCoreRequest).to.equal(0.2);
                expect(inferenceTemplate.spec.compute.cpuMemoryRequest).to.equal('200M');
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.equal(2);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(autoScaling.minReplicas);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;
                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                // cleanup
                await computeResourceUtils.deleteByName(computeResourceName);
                await computeResourceUtils.waitForComputeResourceToBeDeleted(computeResourceName);
                logger.debug(`computeResource: ${computeResourceName} deleted successfully!`);
                logger.debug('test completed successfully!');
            });

            test('should create new nim template with new nfs from drawer', async function ({}, testInfo: TestInfo) {
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                const dataSourceOptions: IDataSourceOptions = {
                    nfs: dataSourceName
                };

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    },
                    dataSourceOptions
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);
                expect(inferenceTemplate.spec.storage.nfs).to.not.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new nim template with new pvc from drawer', async function () {
                const pvcOptions: IPVCOptions = {
                    name: objectGenerator.generateNameByPrefix('pvc-test'),
                    storageClass: 'none',
                    accessMode: 'Read-write by one node',
                    claimSize: '1',
                    claimUnit: 'GB',
                    containerPath: '/home',
                    volumeMode: 'Filesystem'
                };
                const dataSourceOptions: IDataSourceOptions = {
                    pvc: pvcOptions
                };

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    },
                    dataSourceOptions
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);
                expect(inferenceTemplate.spec.storage.pvc[0]).to.be.exist;
                expect(inferenceTemplate.spec.storage.pvc[0].path).to.be.equal(pvcOptions.containerPath);
                expect(inferenceTemplate.spec.storage.pvc[0].existingPvc).to.be.true;
                expect(inferenceTemplate.spec.storage.pvc[0].claimName).to.be.contains(pvcOptions.name);
                expect(inferenceTemplate.spec.storage.pvc[0].readOnly).to.be.false;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new nim template with compute resource asset and data sources', async function () {
                const computeName = objectGenerator.generateNameByPrefix('custom-compute');
                const servingPort: IServingPortOptions = {
                    protocol: 'HTTP',
                    port: 8000
                };
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName
                    },
                    computeSection: {
                        newComputeResource: {
                            assetName: computeName,
                            projectName: project.name,
                            gpu: {
                                devicesInput: 3
                            },
                            isProjectScope: false
                        }
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                expect(inferenceTemplate.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(servingPort.port);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal(
                    servingPort.protocol.toLowerCase()
                );

                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.gpuPortionRequest).to.be.equal(0.03);
                expect(inferenceTemplate.spec.compute.gpuRequestType).to.be.equal('portion');
                expect(inferenceTemplate.spec.storage.nfs).to.deep.equal([
                    {
                        path: '/test/sample',
                        readOnly: false,
                        server: 'test-server',
                        mountPath: '/home/local/data',
                        name: undefined,
                        exclude: undefined
                    }
                ]);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                // cleanup
                const computeId = await computeResourceUtils.getId(computeName);
                await computeResourceUtils.delete(computeId);
                await computeResourceUtils.waitForComputeResourceToBeDeleted(computeName);
            });

            test('should continue to advanced setup when create nim workload', async function () {
                const port: number = 4000;
                const servingPort: IServingPortOptions = {
                    port,
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        servingPort
                    },
                    computeSection: {
                        autoScaling: {
                            minReplicas: 0,
                            maxReplicas: 4,
                            thresholdValue: 97,
                            autoScalingPeriod: '1m'
                        }
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    },
                    advancedSetup: true
                });

                // check that the environment variables are filled correctly
                const environmentVariables = [
                    ...updateNimServerPort(baseNimEnvironmentVariables, port),
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'NIM_CACHE_PATH',
                        value: '/test/sample',
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ];
                await nimElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                expectNimTemplate(inferenceTemplate, templateName, servingPort, environmentVariables);
                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should continue to advanced setup and create new data source', async function () {
                const port: number = 4000;
                const servingPort: IServingPortOptions = {
                    port,
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        servingPort
                    },
                    computeSection: {
                        autoScaling: {
                            minReplicas: 0,
                            maxReplicas: 4,
                            thresholdValue: 97,
                            autoScalingPeriod: '1m'
                        }
                    },
                    advancedSetup: true
                });

                // we want to make sure that all the data is still there
                const environmentVariables = [
                    ...updateNimServerPort(baseNimEnvironmentVariables, port),
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ];
                await nimElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                // go to a different page
                await inferenceTemplateUIUtil.fillDataSourceData({ nfs: dataSourceName });
                await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // we want to make sure that all the data is still there
                expectNimTemplate(inferenceTemplate, templateName, servingPort, environmentVariables);
                logger.debug(`NIM inference: ${templateName} created successfully!`);
            });

            test('should support modifying advanced setup fields, navigating back to NIM configuration, and completing submission with the origin data', async function () {
                const port: number = 4000;
                const servingPort: IServingPortOptions = {
                    port,
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');

                await inferenceTemplateUIUtil.createNimTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        servingPort
                    },
                    computeSection: {
                        autoScaling: {
                            minReplicas: 0,
                            maxReplicas: 4,
                            thresholdValue: 97,
                            autoScalingPeriod: '1m'
                        }
                    },
                    advancedSetup: true
                });

                // check that the environment variables are filled correctly
                const environmentVariables = [
                    ...updateNimServerPort(baseNimEnvironmentVariables, port),
                    {
                        name: 'NGC_API_KEY',
                        value: apiKey,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ];

                await nimElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                // change the environment and compute data
                await inferenceTemplateUIUtil.fillEnvironmentData({ assetName: 'gpt2' });
                await inferenceTemplateUIUtil.fillComputeData({ assetName: 'two-gpus' });
                await inferenceTemplateUIUtil.fillDataSourceData({ nfs: dataSourceName });

                // go back to the nim page
                await workloadTemplate.locator('[aid="model-section"] [aid="back-btn"]').click();

                // submit the inference
                await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expectNimTemplate(inferenceTemplate, templateName, servingPort, environmentVariables, false);
                logger.debug(`NIM inference: ${templateName} created successfully!`);
            });

            test('should copy an existing inference nim template', async function () {
                const copiedTemplateName = objectGenerator.generateNameByPrefix('inf-template-copy');
                const pvcName = objectGenerator.generateNameByPrefix('inf-pvc-storage');
                const pvcSpec = {
                    path: '/home',
                    claimName: objectGenerator.generateNameByPrefix('inf-pvc-storage-claim'),
                    existingPvc: false,
                    readOnly: false,
                    ephemeral: false,
                    claimInfo: {
                        size: '100M',
                        accessModes: {
                            readWriteOnce: true
                        },
                        volumeMode: PvcVolumeMode.Filesystem
                    }
                };
                const pvc = await createPvcAsset(pvcName, project.id, pvcSpec);
                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };

                const metaConfigOriginal = {
                    workloadType: WorkloadType.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName
                };

                const servingPort = {
                    protocol: ServingPortProtocol.Grpc,
                    container: 3131,
                    authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                    authorizedGroups: ['group1', 'group2']
                };

                const autoscaling = {
                    minReplicas: 0,
                    maxReplicas: 2,
                    metric: 'concurrency',
                    metricThreshold: 98,
                    scaleToZeroRetentionSeconds: 300
                };

                const templateSpec: InferenceSpecSpec = {
                    image: 'nvcr.io/nim/openai/gpt-oss-20b',
                    environmentVariables: [
                        {
                            name: 'OUTLINES_CACHE_DIR',
                            value: '/tmp/outlines'
                        },
                        {
                            name: NIM_SERVER_PORT,
                            value: '8000'
                        },
                        {
                            name: 'NIM_JSONL_LOGGING',
                            value: '1'
                        },
                        {
                            name: 'NIM_LOG_LEVEL',
                            value: 'INFO'
                        },
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        },
                        {
                            name: 'NIM_CACHE_PATH',
                            value: pvc.spec.path
                        }
                    ],
                    servingPort,
                    compute: {
                        gpuDevicesRequest: 3,
                        largeShmRequest: true
                    },
                    security: {
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    autoscaling,
                    storage: {
                        pvc: [pvcStorage]
                    }
                };

                const originTemplate = await createTemplate(metaConfigOriginal, templateSpec);
                expect(originTemplate).to.not.be.undefined;
                logger.debug(`origin template: ${JSON.stringify(originTemplate)}`);

                await workloadTemplate.page.reload();
                await workloadTemplate.toggleTemplate(metaConfigOriginal.templateName);
                await workloadTemplate.clickOnCopyTemplate();

                await workloadTemplate.fillTemplateName(copiedTemplateName);
                await workloadTemplate.clickOnCreateInferenceTemplateBtn();

                await workloadTemplate.page.waitForTimeout(300);
                const inferenceTemplateCopied: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, copiedTemplateName);

                expect(inferenceTemplateCopied).to.not.be.undefined;
                logger.debug(`copied template: ${JSON.stringify(inferenceTemplateCopied)}`);

                expect(inferenceTemplateCopied.meta.name).to.equal(copiedTemplateName);

                // model section
                expect(inferenceTemplateCopied.spec.image).to.equal(templateSpec.image);
                expect(inferenceTemplateCopied.spec.environmentVariables).to.deep.equals(
                    templateSpec.environmentVariables.map((env) => ({
                        name: env.name,
                        value: env.value,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }))
                );

                // serving port
                expect(inferenceTemplateCopied.spec.servingPort.container).to.equal(servingPort.container);
                expect(inferenceTemplateCopied.spec.servingPort.protocol).to.equal(servingPort.protocol);
                expect(inferenceTemplateCopied.spec.servingPort.authorizationType).to.equal(
                    servingPort.authorizationType
                );
                expect(inferenceTemplateCopied.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplateCopied.spec.servingPort.authorizedGroups).to.deep.equals(
                    servingPort.authorizedGroups
                );

                // compute section
                expect(inferenceTemplateCopied.spec.compute.gpuDevicesRequest).to.equal(3);
                expect(inferenceTemplateCopied.spec.compute.largeShmRequest).to.equal(true);

                // security section
                expect(inferenceTemplateCopied.spec.security.runAsGid).to.equal(1000);
                expect(inferenceTemplateCopied.spec.security.runAsUid).to.equal(1000);

                // auto scaling
                expect(inferenceTemplateCopied.spec.autoscaling.minReplicas).to.equal(
                    autoscaling.minReplicas
                );
                expect(inferenceTemplateCopied.spec.autoscaling.maxReplicas).to.equal(
                    autoscaling.maxReplicas
                );
                expect(inferenceTemplateCopied.spec.autoscaling.metric).to.equal(autoscaling.metric);
                expect(inferenceTemplateCopied.spec.autoscaling.metricThreshold).to.equal(
                    autoscaling.metricThreshold
                );
                expect(inferenceTemplateCopied.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(
                    autoscaling.scaleToZeroRetentionSeconds
                );

                // storage section
                expect(inferenceTemplateCopied.spec.storage.pvc[0]).to.be.exist;
                expect(inferenceTemplateCopied.spec.storage.pvc[0].path).to.equal(pvcStorage.path);
                expect(inferenceTemplateCopied.spec.storage.pvc[0].existingPvc).to.equal(
                    pvcStorage.existingPvc
                );
                expect(inferenceTemplateCopied.spec.storage.pvc[0].claimName).to.equal(pvcStorage.claimName);
                expect(inferenceTemplateCopied.spec.storage.pvc[0].readOnly).to.equal(pvcStorage.readOnly);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                await templateUtils.deleteWorkloadTemplateV2(
                    inferenceTemplateCopied.meta.id,
                    WorkloadType.Inference
                );
            });

            test('should edit an existing inference nim template', async function () {
                const pvcName = objectGenerator.generateNameByPrefix('inf-pvc-storage');
                const pvcSpec = {
                    path: '/home',
                    claimName: objectGenerator.generateNameByPrefix('inf-pvc-storage-claim'),
                    existingPvc: false,
                    readOnly: false,
                    ephemeral: false,
                    claimInfo: {
                        size: '100M',
                        accessModes: {
                            readWriteOnce: true
                        },
                        volumeMode: PvcVolumeMode.Filesystem
                    }
                };
                const pvc = await createPvcAsset(pvcName, project.id, pvcSpec);
                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };

                const metaConfigOriginal = {
                    workloadType: WorkloadType.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName
                };

                const templateSpec: InferenceSpecSpec = {
                    image: 'nvcr.io/nim/openai/gpt-oss-20b',
                    environmentVariables: [
                        {
                            name: 'OUTLINES_CACHE_DIR',
                            value: '/tmp/outlines'
                        },
                        {
                            name: NIM_SERVER_PORT,
                            value: '8000'
                        },
                        {
                            name: 'NIM_JSONL_LOGGING',
                            value: '1'
                        },
                        {
                            name: 'NIM_LOG_LEVEL',
                            value: 'INFO'
                        },
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        },
                        {
                            name: 'NIM_CACHE_PATH',
                            value: pvc.spec.path
                        }
                    ],
                    servingPort: {
                        protocol: ServingPortProtocol.Grpc,
                        container: 3131,
                        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                        authorizedGroups: ['group1', 'group2']
                    },
                    compute: {
                        gpuDevicesRequest: 3,
                        largeShmRequest: true
                    },
                    security: {
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    autoscaling: {
                        minReplicas: 0,
                        maxReplicas: 2,
                        metric: 'concurrency',
                        metricThreshold: 98,
                        scaleToZeroRetentionSeconds: 300
                    },
                    storage: {
                        pvc: [pvcStorage]
                    }
                };

                const expectedTemplateSpec: InferenceSpecSpec = {
                    image: `nvcr.io/nim/openai/${nimModelName}`,
                    environmentVariables: [
                        {
                            name: 'OUTLINES_CACHE_DIR',
                            value: '/tmp/outlines'
                        },
                        {
                            name: NIM_SERVER_PORT,
                            value: '8000'
                        },
                        {
                            name: 'NIM_JSONL_LOGGING',
                            value: '1'
                        },
                        {
                            name: 'NIM_LOG_LEVEL',
                            value: 'INFO'
                        },
                        {
                            name: 'NGC_API_KEY',
                            value: apiKey
                        },
                        {
                            name: 'NIM_CACHE_PATH',
                            value: pvc.spec.path
                        }
                    ],
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8080,
                        authorizationType: undefined, //when it is public, it will be set to undefined
                        authorizedUsers: undefined,
                        authorizedGroups: undefined
                    },
                    compute: {
                        cpuCoreRequest: 0,
                        cpuMemoryRequest: '0M',
                        gpuDevicesRequest: 0,
                        largeShmRequest: true
                    },
                    security: {
                        runAsGid: 1000,
                        runAsUid: 1000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 2,
                        metric: 'concurrency',
                        metricThreshold: 54
                    },
                    storage: {
                        pvc: [pvcStorage]
                    }
                };

                function expectTemplate(
                    inferenceTemplate: InferenceTemplate,
                    expectedData: InferenceSpecSpec
                ): void {
                    // model section
                    expect(inferenceTemplate.spec.image).to.equal(expectedData.image);

                    // serving port
                    expect(inferenceTemplate.spec.servingPort.container).to.equal(
                        expectedData.servingPort.container
                    );
                    expect(inferenceTemplate.spec.servingPort.protocol).to.equal(
                        expectedData.servingPort.protocol
                    );
                    expect(inferenceTemplate.spec.servingPort.authorizationType).to.equal(
                        expectedData.servingPort.authorizationType
                    );
                    expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                    expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.deep.equals(
                        expectedData.servingPort.authorizedGroups
                    );

                    expect(inferenceTemplate.spec.security.runAsGid).to.equal(1000);
                    expect(inferenceTemplate.spec.security.runAsUid).to.equal(1000);

                    // compute section
                    expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.equal(
                        expectedData.compute.gpuDevicesRequest
                    );
                    expect(inferenceTemplate.spec.compute.largeShmRequest).to.equal(true);

                    // auto scaling
                    expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(
                        expectedData.autoscaling.minReplicas
                    );
                    expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(
                        expectedData.autoscaling.maxReplicas
                    );
                    expect(inferenceTemplate.spec.autoscaling.metric).to.equal(
                        expectedData.autoscaling.metric
                    );
                    expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.equal(
                        expectedData.autoscaling.metricThreshold
                    );
                    expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(
                        expectedData.autoscaling.scaleToZeroRetentionSeconds
                    );

                    // storage section
                    expect(inferenceTemplate.spec.storage.pvc[0]).to.be.exist;
                    expect(inferenceTemplate.spec.storage.pvc[0].path).to.equal(
                        expectedData.storage.pvc[0].path
                    );
                    expect(inferenceTemplate.spec.storage.pvc[0].existingPvc).to.equal(
                        expectedData.storage.pvc[0].existingPvc
                    );
                    expect(inferenceTemplate.spec.storage.pvc[0].claimName).to.equal(
                        expectedData.storage.pvc[0].claimName
                    );
                    expect(inferenceTemplate.spec.storage.pvc[0].readOnly).to.equal(
                        expectedData.storage.pvc[0].readOnly
                    );
                }

                const originTemplate = await createTemplate(metaConfigOriginal, templateSpec);
                expect(originTemplate).to.not.be.undefined;
                logger.debug(`origin template: ${JSON.stringify(originTemplate)}`);
                logger.debug(`origin template spec: ${JSON.stringify(templateSpec)}`);

                expectTemplate(originTemplate, templateSpec);

                await workloadTemplate.page.reload();
                await workloadTemplate.toggleTemplate(metaConfigOriginal.templateName);
                await workloadTemplate.clickOnEditTemplate();

                const computeData: IComputeOptions = {
                    gpu: {
                        devicesInput: '0'
                    },
                    cpu: {
                        compute: {
                            request: '0.1',
                            limit: '0.2'
                        }
                    },
                    autoScaling: {
                        minReplicas: expectedTemplateSpec.autoscaling.minReplicas,
                        maxReplicas: expectedTemplateSpec.autoscaling.maxReplicas,
                        thresholdValue: expectedTemplateSpec.autoscaling.metricThreshold
                    }
                };

                const envData: IEnvironmentOptions = {
                    image: expectedTemplateSpec.image,
                    servingPort: {
                        protocol: expectedTemplateSpec.servingPort.protocol.toString(),
                        port: expectedTemplateSpec.servingPort.container,
                        access: {
                            option: IServingPortAccessOptions.Public
                        }
                    }
                };
                await inferenceTemplateUIUtil.fillEnvironmentData(envData);
                await inferenceTemplateUIUtil.fillComputeData(computeData);
                await workloadTemplate.clickOnCreateInferenceTemplateBtn();

                await workloadTemplate.page.waitForTimeout(300);
                const editedTemplate: InferenceTemplate | undefined = await templateUtils.getTemplateV2ByName(
                    WorkloadType.Inference,
                    templateName
                );

                expect(editedTemplate).to.not.be.undefined;
                logger.debug(`edited template: ${JSON.stringify(editedTemplate)}`);
                logger.debug(`edited template spec: ${JSON.stringify(expectedTemplateSpec)}`);

                expectTemplate(editedTemplate, expectedTemplateSpec);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test.describe('nim ngc api key tests', function () {
                    test('should create nim template with ngc api key source type', async function () {
                        const secretName = objectGenerator.generateNameByPrefix('ngc-api-key');
                        const ngcApiKey: NgcApiKey = await createNgcApiKeySecret(
                            secretName,
                            project.id,
                            'someToken'
                        );

                        await inferenceTemplateUIUtil.createNimTemplate({
                            templateMeta,
                            modelSection: {
                                modelName: nimModelName,
                                credentials: {
                                    source: ECredentialsSource.Shared,
                                    sourceType: ECredentialsSourceType.NgcApiKey,
                                    secretName: secretName,
                                    sourceDisabled: true
                                }
                            }
                        });

                        const inferenceTemplate: InferenceTemplate | undefined =
                            await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                        await credentialUtils.deleteNgcApiKeyAsset(ngcApiKey.meta.id);
                        expect(inferenceTemplate).to.not.be.undefined;
                        logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                        // NIM specific data
                        expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                            ...baseNimEnvironmentVariables,
                            {
                                name: 'NGC_API_KEY',
                                value: undefined,
                                secret: {
                                    name: ngcApiKey.clusterInfo.resources[0].name,
                                    key: 'NGC_API_KEY'
                                },
                                configMap: undefined,
                                podFieldRef: undefined,
                                userCredential: undefined,
                                description: undefined,
                                exclude: undefined
                            }
                        ]);

                        logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
                    });

                    test('should create nim template with ngc api key source type and continue to advanced setup', async function () {
                        const secretName = objectGenerator.generateNameByPrefix('ngc-api-key');
                        const ngcApiKey: NgcApiKey = await createNgcApiKeySecret(
                            secretName,
                            project.id,
                            'someToken'
                        );

                        await inferenceTemplateUIUtil.createNimTemplate({
                            templateMeta,
                            modelSection: {
                                modelName: nimModelName,
                                credentials: {
                                    source: ECredentialsSource.Shared,
                                    sourceType: ECredentialsSourceType.NgcApiKey,
                                    secretName: secretName,
                                    sourceDisabled: true
                                }
                            },
                            advancedSetup: true
                        });

                        const expectedEnvironmentVariables = [
                            ...baseNimEnvironmentVariables,
                            {
                                name: 'NGC_API_KEY',
                                value: undefined,
                                secret: {
                                    name: ngcApiKey.clusterInfo.resources[0].name,
                                    key: 'NGC_API_KEY'
                                },
                                configMap: undefined,
                                podFieldRef: undefined,
                                userCredential: undefined,
                                description: undefined,
                                exclude: undefined
                            }
                        ];
                        await nimElementValidationUtils.expectEnvironmentVariablesWithSharedCredentials(
                            expectedEnvironmentVariables
                        );

                        // submit the inference
                        await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                        await workloadTemplate.page.waitForTimeout(300);
                        const inferenceTemplate: InferenceTemplate | undefined =
                            await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                        await credentialUtils.deleteNgcApiKeyAsset(ngcApiKey.meta.id);
                        expect(inferenceTemplate).to.not.be.undefined;
                        logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                        // NIM specific data
                        expect(inferenceTemplate.spec.environmentVariables).to.deep.equal(
                            expectedEnvironmentVariables
                        );

                        logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
                    });

                    test('should create nim template with generic secret source type and continue to advanced setup', async function () {
                        const secretName = objectGenerator.generateNameByPrefix('generic-secret');
                        const genericSecret: GenericSecret = await createGenericSecret(
                            secretName,
                            Scope.Project,
                            project.id,
                            undefined,
                            [{ key: 'NGC_API_KEY', value: 'someToken' }]
                        );

                        await inferenceTemplateUIUtil.createNimTemplate({
                            templateMeta,
                            modelSection: {
                                modelName: nimModelName,
                                credentials: {
                                    source: ECredentialsSource.Shared,
                                    sourceType: ECredentialsSourceType.GenericSecret,
                                    secretName: secretName,
                                    sourceDisabled: true
                                }
                            },
                            advancedSetup: true
                        });

                        const expectedEnvironmentVariables = [
                            ...baseNimEnvironmentVariables,
                            {
                                name: 'NGC_API_KEY',
                                value: undefined,
                                secret: {
                                    name: genericSecret.clusterInfo.resources[0].name,
                                    key: 'NGC_API_KEY'
                                },
                                configMap: undefined,
                                podFieldRef: undefined,
                                userCredential: undefined,
                                description: undefined,
                                exclude: undefined
                            }
                        ];
                        await nimElementValidationUtils.expectEnvironmentVariablesWithSharedCredentials(
                            expectedEnvironmentVariables
                        );

                        // submit the inference
                        await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                        await workloadTemplate.page.waitForTimeout(300);
                        const inferenceTemplate: InferenceTemplate | undefined =
                            await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                        await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                        expect(inferenceTemplate).to.not.be.undefined;
                        logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                        // NIM specific data
                        expect(inferenceTemplate.spec.environmentVariables).to.deep.equal(
                            expectedEnvironmentVariables
                        );

                        logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
                    });
                });
            });

            test.describe('create NIM template with models from Hugging Face', function () {
                const huggingFaceModelNameWithoutToken = 'microsoft/DialoGPT-small';
                const huggingFaceModelNameWithToken = 'meta-llama/Llama-2-13b-hf';
                const hfToken = process.env.HUGGING_FACE_TOKEN;
                const nimImageForHuggingFaceModel = 'nvcr.io/nim/nvidia/llm-nim:1.12';

                test('should create NIM template with not a gated model from Hugging Face', async function () {
                    await inferenceTemplateUIUtil.createNimTemplate({
                        templateMeta,
                        modelSection: {
                            modelSource: EModelSourceForNim.HuggingFace,
                            modelName: huggingFaceModelNameWithoutToken
                        }
                    });

                    const inferenceTemplate: InferenceTemplate | undefined =
                        await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                    expect(inferenceTemplate).to.not.be.undefined;
                    logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                    expect(inferenceTemplate.spec.image).to.be.equal(nimImageForHuggingFaceModel);
                    const huggingFaceWithoutTokenEnvVars = [
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
                    expectInferenceSpecEnvironmentVariables(
                        inferenceTemplate.spec.environmentVariables,
                        huggingFaceWithoutTokenEnvVars
                    );

                    expect(inferenceTemplate.spec.labels).to.have.length(1);
                    expect(inferenceTemplate.spec.labels[0].name).to.equal('run.ai/nim-workload');
                    expect(inferenceTemplate.spec.labels[0].value).to.equal('true');
                    expect(inferenceTemplate.spec.servingPort.container).to.equal(8000);
                    expect(inferenceTemplate.spec.servingPort.protocol).to.equal('http');

                    logger.debug(
                        `inferenceTemplate: ${templateName} with model from Hugging Face created successfully!`
                    );
                });

                test('should create NIM template with gated model from Hugging Face', async function () {
                    await inferenceTemplateUIUtil.createNimTemplate({
                        templateMeta,
                        modelSection: {
                            modelSource: EModelSourceForNim.HuggingFace,
                            modelName: huggingFaceModelNameWithToken
                        }
                    });

                    const inferenceTemplate: InferenceTemplate | undefined =
                        await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                    expect(inferenceTemplate).to.not.be.undefined;
                    logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                    expect(inferenceTemplate.spec.image).to.be.equal(nimImageForHuggingFaceModel);
                    const huggingFaceWithoutTokenEnvVars = [
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
                    expectInferenceSpecEnvironmentVariables(
                        inferenceTemplate.spec.environmentVariables,
                        huggingFaceWithoutTokenEnvVars
                    );

                    expect(inferenceTemplate.spec.labels).to.have.length(1);
                    expect(inferenceTemplate.spec.labels[0].name).to.equal('run.ai/nim-workload');
                    expect(inferenceTemplate.spec.labels[0].value).to.equal('true');
                    expect(inferenceTemplate.spec.servingPort.container).to.equal(8000);
                    expect(inferenceTemplate.spec.servingPort.protocol).to.equal('http');

                    logger.debug(
                        `inferenceTemplate: ${templateName} with model from Hugging Face created successfully!`
                    );
                });

                test('should create NIM template with not a gated model from Hugging Face with custom serving port', async function () {
                    const servingPort = {
                        port: 3000
                    };

                    await inferenceTemplateUIUtil.createNimTemplate({
                        templateMeta,
                        modelSection: {
                            modelSource: EModelSourceForNim.HuggingFace,
                            modelName: huggingFaceModelNameWithoutToken,
                            servingPort
                        }
                    });

                    const inferenceTemplate: InferenceTemplate | undefined =
                        await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                    expect(inferenceTemplate).to.not.be.undefined;
                    logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                    expect(inferenceTemplate.spec.image).to.be.equal(nimImageForHuggingFaceModel);
                    const huggingFaceWithoutTokenEnvVars = [
                        {
                            name: 'NIM_SERVER_PORT',
                            value: servingPort.port.toString()
                        },
                        {
                            name: 'NIM_HTTP_API_PORT',
                            value: servingPort.port.toString()
                        },
                        {
                            name: 'NIM_MODEL_NAME',
                            value: `hf://${huggingFaceModelNameWithoutToken}`
                        }
                    ];
                    expectInferenceSpecEnvironmentVariables(
                        inferenceTemplate.spec.environmentVariables,
                        huggingFaceWithoutTokenEnvVars
                    );

                    expect(inferenceTemplate.spec.labels).to.have.length(1);
                    expect(inferenceTemplate.spec.labels[0].name).to.equal('run.ai/nim-workload');
                    expect(inferenceTemplate.spec.labels[0].value).to.equal('true');
                    expect(inferenceTemplate.spec.servingPort.container).to.equal(servingPort.port);
                    expect(inferenceTemplate.spec.servingPort.protocol).to.equal('http');

                    logger.debug(
                        `inferenceTemplate: ${templateName} with model from Hugging Face created successfully!`
                    );
                });

                test('should create NIM template with gated model from Hugging Face with shared credentials', async function () {
                    const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                    const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                        secretName,
                        parseInt(project.id),
                        [{ key: 'HF_TOKEN', value: hfToken }]
                    );

                    await inferenceTemplateUIUtil.createNimTemplate({
                        templateMeta,
                        modelSection: {
                            modelSource: EModelSourceForNim.HuggingFace,
                            modelName: huggingFaceModelNameWithToken,
                            credentials: {
                                source: ECredentialsSource.Shared,
                                secretName: secretName,
                                sourceDisabled: true
                            }
                        }
                    });

                    const inferenceTemplate: InferenceTemplate | undefined =
                        await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                    expect(inferenceTemplate).to.not.be.undefined;
                    logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                    expect(inferenceTemplate.spec.image).to.be.equal(nimImageForHuggingFaceModel);
                    const huggingFaceWithTokenEnvVars = [
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
                            secret: {
                                name: genericSecret.clusterInfo.resources[0].name,
                                key: 'HF_TOKEN'
                            }
                        }
                    ];
                    expectInferenceSpecEnvironmentVariables(
                        inferenceTemplate.spec.environmentVariables,
                        huggingFaceWithTokenEnvVars
                    );

                    expect(inferenceTemplate.spec.labels).to.have.length(1);
                    expect(inferenceTemplate.spec.labels[0].name).to.equal('run.ai/nim-workload');
                    expect(inferenceTemplate.spec.labels[0].value).to.equal('true');
                    expect(inferenceTemplate.spec.servingPort.container).to.equal(8000);
                    expect(inferenceTemplate.spec.servingPort.protocol).to.equal('http');

                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);

                    logger.debug(
                        `inferenceTemplate: ${templateName} with model from Hugging Face created successfully!`
                    );
                });
            });
        });
    }
);
