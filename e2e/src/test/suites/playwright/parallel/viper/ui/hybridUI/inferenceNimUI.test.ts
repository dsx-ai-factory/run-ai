/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { InferencePage } from '../../../../../../../modules/ui/runaiV2/pages/InferencePage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    credentialUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    versionUtils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { expect } from 'chai';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test } from '@playwright/test';
import {
    ECredentialsSource,
    ECredentialsSourceType,
    EModelSourceForNim,
    IServingPortAccessOptions,
    IServingPortOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { Inference } from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { InferenceFlatUIUtil } from '../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import {
    DockerRegistry,
    DockerRegistryCreationRequest,
    EnvironmentVariable,
    GenericSecret,
    GenericSecretCreationRequest,
    KeyValuePair,
    NgcApiKey,
    NgcApiKeyCreationRequest,
    Scope
} from '../../../../../../../modules/api/apis/assetsApi';
import { HybridForm } from '../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { NimElementValidationUtils } from '../../../../../../../modules/ui/utils/inference/NimElementValidationUtils';
import { ElementValidationUtils } from '../../../../../../../modules/ui/utils/ElementValidationUtils';

test.describe(
    `Inferences NIM UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper} ${tags.runOn.nightly}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'inference-nim-ui';
        let project: Project;
        let traceName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let inference: InferencePage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let hybridForm: HybridForm;
        let nimElementValidationUtils: NimElementValidationUtils;
        let elementValidationUtils: ElementValidationUtils;

        const currentUser = flowRootState.users.system_administrator;

        let inferenceName: string;
        let inferenceWorkload: Inference;
        const nimModelName = 'meta/llama-3.1-405b-instruct';
        const apiKey = 'someProfile';
        const profileOrHash = 'someProfile';
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

        function isPlaintextAccessTokenEnvironmentVariable(name: string): boolean {
            return name === 'HF_TOKEN' || name === 'NGC_API_KEY';
        }

        function redactPlaintextAccessTokenValues<T extends { name?: string; value?: string }>(
            environmentVariables: Array<T>
        ): Array<T> {
            return environmentVariables.map((environmentVariable) => {
                if (!isPlaintextAccessTokenEnvironmentVariable(environmentVariable.name ?? '')) {
                    return environmentVariable;
                }
                return { ...environmentVariable, value: undefined };
            });
        }

        type ExpectedInferenceEnvVar =
            | { name: string; value: string }
            | { name: string; secret: { name: string; key: string } };

        function matchesExpectedInferenceEnv(
            a: EnvironmentVariable,
            exp: ExpectedInferenceEnvVar
        ): boolean {
            if (a.name !== exp.name) {
                return false;
            }
            if ('secret' in exp && exp.secret) {
                return (
                    a.secret?.name === exp.secret.name && a.secret?.key === exp.secret.key
                );
            }
            if (isPlaintextAccessTokenEnvironmentVariable(exp.name)) {
                return a.value === undefined;
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
            const label = 'inferenceWorkload.spec.environmentVariables';

            expect(actual, `${label}: expected an array, got undefined`).to.not.be.undefined;
            expect(actual, `${label}: expected an array, got null`).to.not.be.null;
            expect(
                actual!.length,
                `${label}: length mismatch — expected ${expected.length}, got ${actual!.length}. ` +
                    `expected=${JSON.stringify(redactPlaintextAccessTokenValues(expected))} actual=${JSON.stringify(actual)}`
            ).to.equal(expected.length);

            for (const exp of expected) {
                const found = actual!.some((a) => matchesExpectedInferenceEnv(a, exp));
                const detail =
                    'secret' in exp && exp.secret
                        ? `expected secret name=${JSON.stringify(exp.secret.name)} key=${JSON.stringify(exp.secret.key)}`
                        : isPlaintextAccessTokenEnvironmentVariable(exp.name)
                          ? 'expected value to be omitted'
                          : `expected value ${JSON.stringify((exp as { name: string; value: string }).value)}`;
                expect(
                    found,
                    `${label}: missing or wrong entry for ${exp.name} (${detail}). ` +
                        `actual=${JSON.stringify(actual)}`
                ).to.be.true;
            }
        }

        async function createNgcApiKeySecret(
            secretName: string,
            projectId: string,
            ngcApiKey: string,
            isPrivate: boolean = false
        ): Promise<NgcApiKey> {
            try {
                const ngcApiKeySecret: NgcApiKeyCreationRequest = {
                    meta: {
                        name: secretName,
                        description: null,
                        scope: isPrivate ? Scope.Tenant : Scope.Project,
                        projectId: isPrivate ? undefined : parseInt(projectId),
                        clusterId: null,
                        departmentId: null,
                        _private: isPrivate
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

        async function createDockerRegistry(): Promise<DockerRegistry> {
            const credentialName = objectGenerator.generateNameByPrefix('dr');
            try {
                const dockerRegistryPayload: DockerRegistryCreationRequest = {
                    meta: {
                        name: credentialName,
                        description: 'Test personal docker registry for API workload submission',
                        scope: Scope.Tenant,
                        _private: true
                    },
                    spec: {
                        existingSecretName: null,
                        user: 'dr-user',
                        password: 'dr-password',
                        url: 'docker.io'
                    }
                };

                const dockerRegistryResponse =
                    await credentialUtils.createDockerRegistryAsset(dockerRegistryPayload);
                await credentialUtils.waitForDockerRegistryAssetToBeReady(dockerRegistryResponse);
                return dockerRegistryResponse;
            } catch (error) {
                logger.error(
                    `failed to create docker registry ${credentialName}. Error: ${JSON.stringify(error)}`
                );
                throw error;
            }
        }

        function expectNimWorkload(
            inferenceWorkload: Inference,
            inferenceName: string,
            servingPort: IServingPortOptions,
            environmentVariables: Array<EnvironmentVariable>,
            checkNfs: boolean = true
        ): void {
            expect(inferenceWorkload).to.not.be.undefined;
            logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

            expect(inferenceWorkload.name).to.equal(inferenceName);
            expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
            expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
            expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
            expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;
            expect(inferenceWorkload.spec.environmentVariables).to.deep.equal(
                redactPlaintextAccessTokenValues(environmentVariables)
            );
            expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
            expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal('authorizedUsersOrGroups');
            expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
            expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;

            if (checkNfs) {
                expect(inferenceWorkload.spec.storage.nfs).to.not.be.undefined;
                const nfs = inferenceWorkload.spec.storage.nfs[0];
                expect(nfs).to.not.be.undefined;
                expect(nfs.path).to.equal('/test/sample');
                expect(nfs.readOnly).to.be.false;
                expect(nfs.mountPath).to.equal('/home/local/data');
                expect(nfs.server).to.equal('test-server');
            }

            const nimLabel = inferenceWorkload.spec.labels.find(
                (label) => label.name === 'run.ai/nim-workload'
            );
            expect(nimLabel).to.not.be.undefined;
            expect(nimLabel.value).to.be.equal('true');
        }

        async function initBrowser(): Promise<BrowserEngine> {
            // Initialize browser once for all tests with video recording AND tracing support
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
            elementValidationUtils = new ElementValidationUtils(browserEngine.currentPage);

            const currentUser = flowRootState.users.system_administrator;
            await auth.signIn(currentUser.email, currentUser.password);

            return browserEngine;
        }

        test.beforeAll('before all create inference workload based on nim UI', async function () {
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

            browserEngine = await initBrowser();

            logger.info('Browser initialized and user signed in');
        });

        test.afterAll('after all create inference workload based on nim UI', async ({}, testInfo) => {
            // Close browser once after all tests and save video
            if (browserEngine && browserEngine.browser) {
                try {
                    await browserEngine.stopAndArchive(testInfo, suiteName);
                    logger.info('Browser closed and video saved');
                } catch (error) {
                    logger.error(`Error closing browser: ${error}`);
                }
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

        test.beforeEach('before each create inference workload based on nim UI', async function () {
            if (!browserEngine || !browserEngine.browser) {
                browserEngine = await initBrowser();
            }

            inferenceName = objectGenerator.generateNameByPrefix('inference');

            // Generate trace name from test title
            traceName = objectGenerator.generateNameByPrefix(suiteName);

            // Start tracing for this test (context and page stay alive)
            await browserEngine.startTracing(traceName);

            // Initialize page objects for this test
            inference = new InferencePage(browserEngine.currentPage);
            nimElementValidationUtils = new NimElementValidationUtils(browserEngine.currentPage);

            // Navigate to workloads page for each test
            await workload.navigateToWorkloads();
        });

        test.afterEach('after each create inference workload based on nim UI', async ({}, testInfo) => {
            logger.debug(`waiting for workload ${inferenceName} to be deleted`);

            if (inferenceWorkload) {
                // delete inference
                await workloadsUtilsV2.deleteWorkload(
                    inferenceWorkload.workloadId,
                    WorkloadType.INFERENCE,
                    flowRootState.users.system_administrator.accessToken
                );
            }
            inferenceWorkload = undefined;

            logger.debug(`inference ${inferenceName} deleted!`);

            // Stop and archive tracing for this test (browser, context, and page remain alive)
            await browserEngine.stopAndArchiveTracing(testInfo, traceName);
        });

        supportedFrom(SupportedVersions.V2_19)(function () {
            test('should create nim workload with model profile automatically and access model from api key', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with model profile manually and access model from api key', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        profileOrHash
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
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
                        value: undefined,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with model profile automatically and access model from storage only', async function () {
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        fromStorageOnly: true
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;
                expect(inferenceWorkload.spec.storage.nfs).to.deep.equal([
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
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with model profile manually and access model from storage only', async function () {
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        profileOrHash,
                        fromStorageOnly: true
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;
                expect(inferenceWorkload.spec.storage.nfs).to.deep.equal([
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
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with access model from api key and storage', async function () {
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;
                expect(inferenceWorkload.spec.storage.nfs).to.deep.equal([
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
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with access model from shared credentials', async function () {
                const shouldSelectSourceType = versionUtils.isVersionSupported(
                    flowRootState.cluster.version,
                    SupportedVersions.V2_23
                );
                const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                    secretName,
                    parseInt(project.id),
                    [{ key: 'NGC_API_KEY', value: 'someToken' }]
                );

                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        credentials: {
                            source: ECredentialsSource.Shared,
                            sourceType: shouldSelectSourceType
                                ? ECredentialsSourceType.GenericSecret
                                : undefined,
                            secretName: secretName
                        }
                    }
                });

                await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test('should create nim workload with access model from my credentials generic secret', async function () {
                    const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                    const isPrivate = true;
                    const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                        secretName,
                        parseInt(project.id),
                        [{ key: 'NGC_API_KEY', value: 'someToken' }],
                        isPrivate
                    );

                    await inferenceFlatUIUtil.createNimFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        modelSection: {
                            modelName: nimModelName,
                            credentials: {
                                source: ECredentialsSource.Private,
                                sourceType: ECredentialsSourceType.GenericSecret,
                                secretName: secretName
                            }
                        }
                    });

                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // NIM specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                    expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                    expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: undefined,
                            userCredential: {
                                name: genericSecret.meta.name,
                                key: 'NGC_API_KEY'
                            },
                            configMap: undefined,
                            podFieldRef: undefined,
                            secret: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);
                    expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                    // other data
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                    logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                });

                test('should create nim workload with access model from shared credentials generic secret and docker registry', async function () {
                    const dockerRegistry = await createDockerRegistry();
                    const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                    const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                        secretName,
                        parseInt(project.id),
                        [{ key: 'NGC_API_KEY', value: 'someToken' }]
                    );

                    await inferenceFlatUIUtil.createNimFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        modelSection: {
                            modelName: nimModelName,
                            credentials: {
                                source: ECredentialsSource.Shared,
                                sourceType: ECredentialsSourceType.GenericSecret,
                                secretName: secretName
                            },
                            nimPullImageCredential: {
                                dockerRegistry: dockerRegistry.meta.name
                            }
                        }
                    });

                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                    await credentialUtils.deleteDockerRegistryAsset(dockerRegistry.meta.id);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // NIM specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                    expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                    expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
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
                    expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                    expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                        dockerRegistry.meta.name
                    );

                    logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                });

                test('should create nim workload with access model from my credentials generic secret and docker registry', async function () {
                    const dockerRegistry = await createDockerRegistry();
                    const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                    const isPrivate = true;
                    const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                        secretName,
                        parseInt(project.id),
                        [{ key: 'NGC_API_KEY', value: 'someToken' }],
                        isPrivate
                    );

                    await inferenceFlatUIUtil.createNimFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        modelSection: {
                            modelName: nimModelName,
                            credentials: {
                                source: ECredentialsSource.Private,
                                sourceType: ECredentialsSourceType.GenericSecret,
                                secretName: secretName
                            },
                            nimPullImageCredential: {
                                dockerRegistry: dockerRegistry.meta.name
                            }
                        }
                    });

                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                    await credentialUtils.deleteDockerRegistryAsset(dockerRegistry.meta.id);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // NIM specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                    expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                    expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        ...baseNimEnvironmentVariables,
                        {
                            name: 'NGC_API_KEY',
                            value: undefined,
                            userCredential: {
                                name: genericSecret.meta.name,
                                key: 'NGC_API_KEY'
                            },
                            configMap: undefined,
                            podFieldRef: undefined,
                            secret: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);
                    expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;
                    expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                    expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                        dockerRegistry.meta.name
                    );
                    logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                });
            });

            test('should create nim workload with model profile automatically and access model from api key and autoscaling', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
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

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(2);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(4);
                expect(inferenceWorkload.spec.autoscaling.metric).to.be.equal('throughput');
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.be.equal(97);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with model profile automatically and access model from api key and serving port access to specific users', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
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

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.equal(
                    'authorizedUsersOrGroups'
                );
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.be.deep.equal([
                    currentUser.email
                ]);
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with custom compute resource', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '3'
                        }
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.be.equal(3);
                expect(inferenceWorkload.spec.compute.cpuMemoryRequest).to.be.equal('0M');
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with compute resource from asset', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    },
                    computeSection: {
                        assetName: 'cpu-only'
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // NIM specific data
                expect(inferenceWorkload.spec.image).to.be.equal(`nvcr.io/nim/${nimModelName}`);
                expect(inferenceWorkload.spec.security.runAsUid).to.be.undefined;
                expect(inferenceWorkload.spec.security.runAsGid).to.be.undefined;
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    ...baseNimEnvironmentVariables,
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
                expect(inferenceWorkload.spec.compute.largeShmRequest).to.be.true;

                // other data
                expect(inferenceWorkload.spec.compute.cpuCoreRequest).to.be.equal(0.1);
                expect(inferenceWorkload.spec.compute.cpuMemoryRequest).to.be.equal('100M');
                expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should test top row when inference workload created from nim', async function () {
                const inferenceName2 = objectGenerator.generateNameByPrefix('nim-inference-2');

                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    }
                });

                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName2,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    }
                });

                await waitUtils.waitForElementToExist(workload, `td:has-text("${inferenceName2}")`);

                const workloadNames: string[] = await workload.getWorkloadNames();
                expect(workloadNames[0].trim()).to.equal(inferenceName2);
                logger.debug('The last created NIM inference is the first!');
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
            });

            test('should create nim workload with default serving port details', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;

                const defaultServingPort = 8000;
                const defaultServingPortProtocol = 'http';
                expect(inferenceWorkload.spec.servingPort.container).to.equal(defaultServingPort);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal(defaultServingPortProtocol);
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should create nim workload with custom serving port protocol and port', async function () {
                const servingPort = {
                    port: 4000,
                    protocol: 'gRPC'
                };

                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey,
                        servingPort
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expect(inferenceWorkload).to.not.be.undefined;

                expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceWorkload.spec.servingPort.protocol).to.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.undefined;

                logger.debug(`NIM inference: ${inferenceName} created successfully!`);
            });

            test('should clone nim workload correctly', async function () {
                const copiedInferenceName = objectGenerator.generateShortNameByPrefix('copied-inference');
                const port: number = 4000;
                const servingPort: IServingPortOptions = {
                    port,
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
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
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

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

                expectNimWorkload(inferenceWorkload, inferenceName, servingPort, environmentVariables);

                await workload.toggleWorkload(inferenceName);
                await workload.clickOnCopyAndEditWorkload();

                await inference.fillInferenceName(copiedInferenceName);

                await nimElementValidationUtils.expectEnvironmentVariables([
                    ...updateNimServerPort(baseNimEnvironmentVariables, port),
                    {
                        name: 'NGC_API_KEY',
                        value: '', // should be empty
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

                // fill the NGC_API_KEY value
                const environmentVariableBoxes = inference.locator('.environment-variable-box');
                const ngcApiKeyIndex = environmentVariables.findIndex((env) => env.name === 'NGC_API_KEY');
                const ngcApiKeyBox = environmentVariableBoxes.nth(ngcApiKeyIndex);
                const ngcApiKeyValueInput = ngcApiKeyBox.locator(
                    hybridForm.hybridEnvironmentSection.envVarValue
                );
                await ngcApiKeyValueInput.fill(apiKey);
                await inference.clickOnCreateInferenceBtn();

                const copiedInferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                    copiedInferenceName,
                    project.id.toString()
                );

                expectNimWorkload(
                    copiedInferenceWorkload,
                    copiedInferenceName,
                    servingPort,
                    environmentVariables
                );

                await workloadUtils.deleteMultipleWorkloads([copiedInferenceWorkload.workloadId]);
                logger.debug(`NIM inference: ${copiedInferenceName} copied successfully!`);
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
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
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

                await inference.clickOnCreateInferenceBtn();

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expectNimWorkload(inferenceWorkload, inferenceName, servingPort, environmentVariables);
                logger.debug(`NIM inference: ${inferenceName} copied successfully!`);
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
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
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
                await inferenceFlatUIUtil.fillDataSourceData({ nfs: dataSourceName });
                await inference.clickOnCreateInferenceBtn();

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                // we want to make sure that all the data is still there
                expectNimWorkload(inferenceWorkload, inferenceName, servingPort, environmentVariables);
                logger.debug(`NIM inference: ${inferenceName} copied successfully!`);
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
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
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
                await inferenceFlatUIUtil.fillEnvironmentData({ assetName: 'gpt2' });
                await inferenceFlatUIUtil.fillComputeData({ assetName: 'two-gpus' });
                await inferenceFlatUIUtil.fillDataSourceData({ nfs: dataSourceName });

                // go back to the nim page
                await inference.locator('[aid="model-section"] [aid="back-btn"]').click();

                // submit the inference
                await inference.clickOnCreateInferenceBtn();

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );

                expectNimWorkload(inferenceWorkload, inferenceName, servingPort, environmentVariables, false);
                logger.debug(`NIM inference: ${inferenceName} copied successfully!`);
            });

            test('should check that the nim label is added to the workload', async function () {
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        apiKey
                    }
                });

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
                test.describe('nim advanced setup request timeout', function () {
                    test('should continue to advanced setup when create nim workload with custom request timeout values', async function () {
                        const servingPort = {
                            port: 4000
                        };
                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelName: nimModelName,
                                apiKey,
                                servingPort
                            },
                            advancedSetup: true
                        });

                        await inferenceFlatUIUtil.fillGeneralSection({
                            initializationTimeout: {
                                minutes: 5,
                                seconds: 7
                            },
                            requestTimeout: {
                                minutes: 2,
                                seconds: 30
                            }
                        });

                        // submit the inference
                        await inference.clickOnCreateInferenceBtn();

                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                        expect(
                            inferenceWorkload.spec.servingConfiguration?.initializationTimeoutSeconds
                        ).to.equal(307); // 5 minutes * 60 (seconds) + 7 seconds
                        expect(inferenceWorkload.spec.servingConfiguration?.requestTimeoutSeconds).to.equal(
                            150
                        ); // 2 minutes * 60 (seconds) + 30 seconds

                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });
                });
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test.describe('nim ngc api key tests', function () {
                    test('should create nim workload with docker registry when manually token is selected', async function () {
                        const dockerRegistry = await createDockerRegistry();

                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelName: nimModelName,
                                apiKey,
                                nimPullImageCredential: {
                                    dockerRegistry: dockerRegistry.meta.name
                                }
                            }
                        });

                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                        expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                        expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                            dockerRegistry.meta.name
                        );

                        await credentialUtils.deleteDockerRegistryAsset(dockerRegistry.meta.id);
                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });

                    test('should create nim workload with docker registry when manually token is selected and advanced setup is enabled', async function () {
                        const dockerRegistry = await createDockerRegistry();

                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelName: nimModelName,
                                apiKey,
                                nimPullImageCredential: {
                                    dockerRegistry: dockerRegistry.meta.name
                                }
                            },
                            advancedSetup: true
                        });

                        await hybridForm.openImagePullPolicySection();
                        // Wait for the credential list to load asynchronously before asserting
                        await hybridForm.page
                            .locator(hybridForm.hybridEnvironmentSection.imagePullSecretsSelect)
                            .filter({ hasText: dockerRegistry.meta.name })
                            .waitFor({ state: 'visible', timeout: 10000 });
                        await elementValidationUtils.expectElementValueToEqual(
                            hybridForm.hybridEnvironmentSection.imagePullSecretsSelect,
                            dockerRegistry.meta.name
                        );

                        // submit the inference
                        await inference.clickOnCreateInferenceBtn();

                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                        expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                        expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                            dockerRegistry.meta.name
                        );

                        await credentialUtils.deleteDockerRegistryAsset(dockerRegistry.meta.id);
                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });

                    test('should create nim workload with ngc api key source type when shared secret is selected', async function () {
                        const secretName = objectGenerator.generateNameByPrefix('ngc-api-key');
                        const ngcApiKey: NgcApiKey = await createNgcApiKeySecret(
                            secretName,
                            project.id,
                            'someToken'
                        );

                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelName: nimModelName,
                                credentials: {
                                    source: ECredentialsSource.Shared,
                                    sourceType: ECredentialsSourceType.NgcApiKey, // default
                                    secretName: secretName
                                }
                            }
                        });

                        await credentialUtils.deleteNgcApiKeyAsset(ngcApiKey.meta.id);

                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                        expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
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

                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });

                    test('should create nim workload with ngc api key source type when my credentials source is selected', async function () {
                        const isPrivate = true;
                        const secretName = objectGenerator.generateNameByPrefix('ngc-api-key');
                        const ngcApiKey: NgcApiKey = await createNgcApiKeySecret(
                            secretName,
                            project.id,
                            'someToken',
                            isPrivate
                        );

                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelName: nimModelName,
                                credentials: {
                                    source: ECredentialsSource.Private,
                                    sourceType: ECredentialsSourceType.NgcApiKey, // default
                                    secretName: secretName
                                }
                            }
                        });

                        await credentialUtils.deleteNgcApiKeyAsset(ngcApiKey.meta.id);

                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                        expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                            ...baseNimEnvironmentVariables,
                            {
                                name: 'NGC_API_KEY',
                                value: undefined,
                                userCredential: {
                                    name: ngcApiKey.meta.name,
                                    key: 'NGC_API_KEY'
                                },
                                configMap: undefined,
                                podFieldRef: undefined,
                                secret: undefined,
                                description: undefined,
                                exclude: undefined
                            }
                        ]);

                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });

                    test('should create nim workload with ngc api key source type when shared credentials source is selected and continue to advanced setup', async function () {
                        const secretName = objectGenerator.generateNameByPrefix('ngc-api-key');
                        const ngcApiKey: NgcApiKey = await createNgcApiKeySecret(
                            secretName,
                            project.id,
                            'someToken'
                        );

                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelName: nimModelName,
                                credentials: {
                                    source: ECredentialsSource.Shared,
                                    sourceType: ECredentialsSourceType.NgcApiKey, // default
                                    secretName: secretName
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
                        await inference.clickOnCreateInferenceBtn();

                        await credentialUtils.deleteNgcApiKeyAsset(ngcApiKey.meta.id);

                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                        expect(inferenceWorkload.spec.environmentVariables).to.deep.equal(
                            expectedEnvironmentVariables
                        );

                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });
                });

                test.describe('create NIM workload with models from Hugging Face', function () {
                    const huggingFaceModelNameWithoutToken = 'microsoft/DialoGPT-small';
                    const huggingFaceModelNameWithToken = 'meta-llama/Llama-2-13b-hf';
                    const hfToken = process.env.HUGGING_FACE_TOKEN;
                    const nimImageForHuggingFaceModel = 'nvcr.io/nim/nvidia/llm-nim:1.12';
                    const secretName = objectGenerator.generateNameByPrefix('ngc-api-key');
                    const genericSecretName = objectGenerator.generateNameByPrefix('gs-asset');

                    let dockerRegistry: DockerRegistry;
                    let ngcApiKey: NgcApiKey;
                    let genericSecret: GenericSecret;
                    test.beforeAll(async function () {
                        const [dockerRegistryResponse, ngcApiKeyResponse, genericSecretResponse] = await Promise.all([
                            createDockerRegistry(),
                            createNgcApiKeySecret(
                                secretName,
                                project.id,
                                'someToken',
                                true
                            ),
                            credentialUtils.createGenericSecretByApi(
                                genericSecretName,
                                parseInt(project.id),
                                [{ key: 'HF_TOKEN', value: hfToken }]
                            )
                        ]);

                        dockerRegistry = dockerRegistryResponse;
                        ngcApiKey = ngcApiKeyResponse;
                        genericSecret = genericSecretResponse;
                    });
    
                    test.afterAll(async function () {
                        await Promise.all([
                            credentialUtils.deleteDockerRegistryAsset(dockerRegistry.meta.id),
                            credentialUtils.deleteNgcApiKeyAsset(ngcApiKey.meta.id),
                            credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id)
                        ]);
                    });
    
                    test('should create NIM workload with models from Hugging Face without token', async function () {
                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelSource: EModelSourceForNim.HuggingFace,
                                modelName: huggingFaceModelNameWithoutToken,
                                nimPullImageCredential: {
                                    dockerRegistry: dockerRegistry.meta.name
                                }
                            }
                        });
    
                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
    
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
    
                        expect(inferenceWorkload.spec.image).to.be.equal(nimImageForHuggingFaceModel);
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
                            },
                        ];
                        expectInferenceSpecEnvironmentVariables(
                            inferenceWorkload.spec.environmentVariables,
                            huggingFaceWithoutTokenEnvVars
                        );
    
                        expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                        expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                            dockerRegistry.meta.name
                        );
    
                        expect(inferenceWorkload.spec.labels).to.have.length(1);
                        expect(inferenceWorkload.spec.labels[0].name).to.equal('run.ai/nim-workload');
                        expect(inferenceWorkload.spec.labels[0].value).to.equal('true');
                        expect(inferenceWorkload.spec.servingPort.container).to.equal(8000);
                        expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
    
                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });
    
                    test('should create NIM workload with gated model from Hugging Face', async function () {
                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelSource: EModelSourceForNim.HuggingFace,
                                modelName: huggingFaceModelNameWithToken,
                                hfToken,
                                nimPullImageCredential: {
                                    dockerRegistry: dockerRegistry.meta.name
                                }
                            }
                        });
    
                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
    
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
    
                        expect(inferenceWorkload.spec.image).to.be.equal(nimImageForHuggingFaceModel);
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
                                value: hfToken
                            }
                        ];
                        expectInferenceSpecEnvironmentVariables(
                            inferenceWorkload.spec.environmentVariables,
                            huggingFaceWithTokenEnvVars
                        );
    
                        expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                        expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                            dockerRegistry.meta.name
                        );
    
                        expect(inferenceWorkload.spec.labels).to.have.length(1);
                        expect(inferenceWorkload.spec.labels[0].name).to.equal('run.ai/nim-workload');
                        expect(inferenceWorkload.spec.labels[0].value).to.equal('true');
                        expect(inferenceWorkload.spec.servingPort.container).to.equal(8000);
                        expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
    
                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });
    
                    test('should create NIM workload with not a gated model from Hugging Face with custom serving port', async function () {                    
                        const servingPort = {
                            port: 3000,
                        };
                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelSource: EModelSourceForNim.HuggingFace,
                                modelName: huggingFaceModelNameWithoutToken,
                                nimPullImageCredential: {
                                    dockerRegistry: dockerRegistry.meta.name
                                },
                                servingPort
                            }
                        });
    
                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
    
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
    
                        expect(inferenceWorkload.spec.image).to.be.equal(nimImageForHuggingFaceModel);
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
                            },
                        ];
                        expectInferenceSpecEnvironmentVariables(
                            inferenceWorkload.spec.environmentVariables,
                            huggingFaceWithoutTokenEnvVars
                        );
    
                        expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                        expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                            dockerRegistry.meta.name
                        );
    
                        expect(inferenceWorkload.spec.labels).to.have.length(1);
                        expect(inferenceWorkload.spec.labels[0].name).to.equal('run.ai/nim-workload');
                        expect(inferenceWorkload.spec.labels[0].value).to.equal('true');
                        expect(inferenceWorkload.spec.servingPort.container).to.equal(servingPort.port);
                        expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
    
                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });
    
                    test('should create NIM workload with gated model from Hugging Face with shared credentials', async function () {
                        const shouldSelectSourceType = versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_23
                        );
        
                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelSource: EModelSourceForNim.HuggingFace,
                                modelName: huggingFaceModelNameWithToken,
                                credentials: {
                                    // hugging-face-access-section
                                    source: ECredentialsSource.Shared,
                                    secretName: genericSecretName
                                },
                                nimPullImageCredential: {
                                    // nim-image-pull-credential-selector
                                    dockerRegistry: dockerRegistry.meta.name
                                }
                            }
                        });
    
                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
    
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
    
                        expect(inferenceWorkload.spec.image).to.be.equal(nimImageForHuggingFaceModel);
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
                            inferenceWorkload.spec.environmentVariables,
                            huggingFaceWithTokenEnvVars
                        );
    
                        expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                        expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.equal(
                            dockerRegistry.meta.name
                        );
    
                        expect(inferenceWorkload.spec.labels).to.have.length(1);
                        expect(inferenceWorkload.spec.labels[0].name).to.equal('run.ai/nim-workload');
                        expect(inferenceWorkload.spec.labels[0].value).to.equal('true');
                        expect(inferenceWorkload.spec.servingPort.container).to.equal(8000);
                        expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
    
                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });

                    test('should create NIM workload with models from Hugging Face with NGC API key', async function () {
                        await inferenceFlatUIUtil.createNimFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            modelSection: {
                                modelSource: EModelSourceForNim.HuggingFace,
                                modelName: huggingFaceModelNameWithoutToken,
                                nimPullImageCredential: {
                                    ngcApiKey: ngcApiKey.meta.name
                                }
                            }
                        });
    
                        inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );
    
                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
    
                        expect(inferenceWorkload.spec.image).to.be.equal(nimImageForHuggingFaceModel);
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
                            },
                        ];
                        expectInferenceSpecEnvironmentVariables(
                            inferenceWorkload.spec.environmentVariables,
                            huggingFaceWithoutTokenEnvVars
                        );
    
                        expect(inferenceWorkload.spec.imagePullSecrets).to.have.length(1);
                        expect(inferenceWorkload.spec.imagePullSecrets[0].name).to.include(
                            ngcApiKey.meta.name
                        );
    
                        expect(inferenceWorkload.spec.labels).to.have.length(1);
                        expect(inferenceWorkload.spec.labels[0].name).to.equal('run.ai/nim-workload');
                        expect(inferenceWorkload.spec.labels[0].value).to.equal('true');
                        expect(inferenceWorkload.spec.servingPort.container).to.equal(8000);
                        expect(inferenceWorkload.spec.servingPort.protocol).to.equal('http');
    
                        logger.debug(`NIM inference: ${inferenceName} created successfully!`);
                    });
                });
            });

        });
    }
);
