/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { InferencePage } from '../../../../../../../modules/ui/runaiV2/pages/InferencePage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
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
import { test, TestInfo } from '@playwright/test';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';
import {
    EVolumeAccessMode,
    IInferenceType,
    IServingPortAccessOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { Inference } from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    ECredentialsSource,
    IDataSourceOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { InferenceFlatUIUtil } from '../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import {
    EnvironmentVariable,
    GenericSecret,
    PVCAsset,
    PVCAssetSpec,
    PVCCreationRequest,
    PvcVolumeMode,
    Scope
} from '../../../../../../../modules/api/apis/assetsApi';
import { HybridForm } from '../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { HuggingFaceElementValidationUtils } from '../../../../../../../modules/ui/utils/inference/HuggingFaceElementValidationUtils';

test.describe(
    `Inferences Flat UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper} ${tags.runOn.nightly}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'inference-hf-ui';
        let project: Project;
        let traceName: string;
        let namespace: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let inference: InferencePage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let hybridForm: HybridForm;
        let huggingFaceElementValidationUtils: HuggingFaceElementValidationUtils;
        let inferenceName: string;

        const currentUser = flowRootState.users.system_administrator;

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

        async function initBrowser(): Promise<BrowserEngine> {
            // Initialize browser once for all tests with video recording AND tracing support
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);

            const currentUser = flowRootState.users.system_administrator;
            await auth.signIn(currentUser.email, currentUser.password);

            return browserEngine;
        }

        test.beforeAll(
            'before all create inference workload based on hugging face and model catalog UI',
            async function () {
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
                namespace = project.status.namespace;

                logger.info(`created project ${projectName}`);

                browserEngine = await initBrowser();

                logger.info('Browser initialized and user signed in');
            }
        );

        test.afterAll(
            'after all create inference workload based on hugging face and model catalog UI',
            async ({}, testInfo) => {
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

                if (!project) {
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
            }
        );

        test.beforeEach(
            'before each create inference workload based on hugging face and model catalog UI',
            async function () {
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
                huggingFaceElementValidationUtils = new HuggingFaceElementValidationUtils(
                    browserEngine.currentPage
                );
            }
        );

        test.afterEach(
            'after each create inference workload based on hugging face and model catalog UI',
            async function ({}, testInfo) {
                // Stop and archive tracing for this test (browser, context, and page remain alive)
                await browserEngine.stopAndArchiveTracing(testInfo, traceName);
            }
        );

        supportedFrom(SupportedVersions.V2_19)(function () {
            test.describe('Create inference workload based on hugging face UI', function () {
                // Skip the entire HF inference UI suite on kind/CI preview envs.
                // These envs are too slow for the 30s "wait for workload to be in
                // control plane" wait that every HF-model test depends on.

                test.skip(
                    (process.env.CLUSTER_NAME_PREFIX ?? '').startsWith('ci-'),
                    'inferenceHuggingFaceUI tests are too slow on kind/CI preview envs (workload-to-control-plane > 30s)'
                );

                const huggingFaceModelNameWithoutToken = 'microsoft/DialoGPT-small';
                const huggingFaceModelNameWithToken = 'meta-llama/Llama-2-13b-hf';
                const vllmImage = 'vllm/vllm-openai:latest';
                const tgiImage = 'ghcr.io/huggingface/text-generation-inference:latest';
                const hfToken = process.env.HUGGING_FACE_TOKEN;
                const huggingFaceDefaultPort = 8000;
                const huggingFaceDefaultProtocol = 'http';
                const hfHomeDefaultValue = '/tmp/huggingface';

                function redactPlaintextAccessTokenValues(
                    environmentVariables: Array<EnvironmentVariable>
                ): Array<EnvironmentVariable> {
                    return environmentVariables.map((environmentVariable) => {
                        if (
                            environmentVariable.name !== 'HF_TOKEN' &&
                            environmentVariable.name !== 'NGC_API_KEY'
                        ) {
                            return environmentVariable;
                        }
                        return { ...environmentVariable, value: undefined };
                    });
                }

                function expectHuggingFaceWorkload(
                    inferenceWorkload: Inference,
                    inferenceName: string,
                    image: string,
                    args: string,
                    environmentVariables: Array<EnvironmentVariable>,
                    checkNfs: boolean = true
                ): void {
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    expect(inferenceWorkload.name).to.equal(inferenceName);
                    expect(inferenceWorkload.spec.image).to.be.equal(image);
                    expect(inferenceWorkload.spec.args).to.be.equal(args);
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal(
                        redactPlaintextAccessTokenValues(environmentVariables)
                    );
                    expect(inferenceWorkload.spec.servingPort.container).to.equal(huggingFaceDefaultPort);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
                    expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal(
                        'authorizedUsersOrGroups'
                    );
                    expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([
                        currentUser.email
                    ]);
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
                }

                test.beforeEach(
                    'before each create inference workload based on hugging face UI',
                    async function () {
                        await workload.navigateToWorkloads();
                    }
                );

                test.afterEach('after each create inference workload based on hugging face UI', async () => {
                    logger.debug(`waiting for workload ${inferenceName} to be deleted`);

                    // delete inference — wrap in try/catch so cleanup is tolerant of
                    // skipped tests (skipPlaywrightTest skips the body, but afterEach
                    // still runs; the workload may not exist if the test never created it)
                    try {
                        await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                            inferenceName,
                            project.id,
                            WorkloadType.INFERENCE,
                            flowRootState.users.system_administrator.accessToken
                        );
                    } catch (error) {
                        logger.warn(`Failed to delete inference ${inferenceName}: ${error?.message ?? error}`);
                    }

                    logger.debug(`inference ${inferenceName} deleted!`);
                });

                test('should create hugging face workload without token from new workload page', async function () {
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithoutToken
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithoutToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should default to credentials and warn before providing a token', async function () {
                    try {
                        await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                            {
                                clusterName: flowRootState.cluster.name,
                                projectName: project.name,
                                workloadName: inferenceName
                            },
                            IInferenceType.VLLM
                        );

                        await hybridForm.selectFromDropDownWithInput(
                            hybridForm.hybridHuggingFaceModelSection.modelNameSelectInput,
                            huggingFaceModelNameWithToken
                        );

                        const {
                            credentials,
                            hfToken: hfTokenInput,
                            plaintextTokenWarning,
                            provideToken,
                            selectCredential
                        } = hybridForm.hybridHuggingFaceModelSection;

                        await browserEngine.currentPage
                            .locator(`.q-radio[aria-checked="true"] ${selectCredential}`)
                            .waitFor({ state: 'visible' });
                        await browserEngine.currentPage
                            .locator(credentials.sourceSelect)
                            .waitFor({ state: 'visible' });
                        expect(await browserEngine.currentPage.locator(credentials.sourceSelect).isVisible()).to
                            .be.true;
                        expect(await browserEngine.currentPage.locator(hfTokenInput).isVisible()).to.be.false;
                        expect(await browserEngine.currentPage.locator(plaintextTokenWarning).isVisible()).to.be
                            .false;

                        await hybridForm.selectHuggingFaceTokenManually();

                        await browserEngine.currentPage
                            .locator(`.q-radio[aria-checked="true"] ${provideToken}`)
                            .waitFor({ state: 'visible' });
                        await browserEngine.currentPage.locator(hfTokenInput).waitFor({ state: 'visible' });
                        expect(await browserEngine.currentPage.locator(hfTokenInput).isVisible()).to.be.true;

                        const warning = browserEngine.currentPage.locator(plaintextTokenWarning);
                        await warning.waitFor({ state: 'visible' });
                        expect(await warning.isVisible()).to.be.true;
                        expect(await warning.innerText()).to.contain(
                            'The token will be visible in the workload spec and via the API.'
                        );
                        expect(await warning.innerText()).to.contain(
                            'For secured storage insert the Hugging Face token as a new credential.'
                        );
                    } finally {
                        await hybridForm.closeWorkloadForm();
                    }
                });

                test('should create hugging face workload with token from new workload page', async function () {
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithToken,
                            hfToken
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: undefined,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should create hugging face workload model with credentials from new workload page', async function () {
                    const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                    const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                        secretName,
                        parseInt(project.id),
                        [{ key: 'HF_TOKEN', value: hfToken }]
                    );
                    await browserEngine.currentPage.waitForTimeout(5000);

                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithToken,
                            credentials: {
                                source: ECredentialsSource.Shared,
                                secretName
                            }
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            secret: {
                                name: genericSecret.clusterInfo.resources[0].name,
                                key: 'HF_TOKEN'
                            },
                            value: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                supportedFrom(SupportedVersions.V2_23)(function () {
                    test('should create hugging face workload model with my credentials from new workload page', async function () {
                        const secretName = objectGenerator.generateNameByPrefix('gs-asset');
                        const isPrivate = true;
                        const genericSecret: GenericSecret = await credentialUtils.createGenericSecretByApi(
                            secretName,
                            parseInt(project.id),
                            [{ key: 'HF_TOKEN', value: hfToken }],
                            isPrivate
                        );
                        await browserEngine.currentPage.waitForTimeout(5000);

                        await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                            clusterName: flowRootState.cluster.name,
                            projectName: project.name,
                            workloadName: inferenceName,
                            inferenceType: IInferenceType.VLLM,
                            modelSection: {
                                modelName: huggingFaceModelNameWithToken,
                                credentials: {
                                    source: ECredentialsSource.Private,
                                    secretName
                                }
                            }
                        });

                        const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                            inferenceName,
                            project.id.toString()
                        );

                        expect(inferenceWorkload).to.not.be.undefined;
                        logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                        // HF specific data
                        expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                        expect(inferenceWorkload.spec.args).to.be.equal(
                            `--model ${huggingFaceModelNameWithToken}`
                        );
                        expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                            {
                                name: 'HF_HOME',
                                value: hfHomeDefaultValue,
                                secret: undefined,
                                configMap: undefined,
                                podFieldRef: undefined,
                                userCredential: undefined,
                                description: undefined,
                                exclude: undefined
                            },
                            {
                                name: 'HF_TOKEN',
                                userCredential: {
                                    name: genericSecret.meta.name,
                                    key: 'HF_TOKEN'
                                },
                                value: undefined,
                                configMap: undefined,
                                podFieldRef: undefined,
                                secret: undefined,
                                description: undefined,
                                exclude: undefined
                            },
                        ]);

                        expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                        expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                        expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                        expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                        expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                        await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);

                        logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                    });
                });

                test('should create hugging face workload model with TGI image', async function () {
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.TGI,
                        modelSection: {
                            modelName: huggingFaceModelNameWithoutToken,
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(tgiImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model-id ${huggingFaceModelNameWithoutToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal(
                        huggingFaceDefaultProtocol
                    );
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should create hugging face workload serving port access to specific users', async function () {
                    const servingPort = {
                        access: {
                            option: IServingPortAccessOptions.Users
                        }
                    };

                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithoutToken,
                            servingPort,
                            shouldOpenServingPortSection: true
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithoutToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.servingPort.container).to.equal(huggingFaceDefaultPort);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
                    expect(inferenceWorkload.spec.servingPort.authorizationType).to.equal(
                        'authorizedUsersOrGroups'
                    );
                    expect(inferenceWorkload.spec.servingPort.authorizedUsers).to.deep.equal([
                        currentUser.email
                    ]);

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should create hugging face workload with custom compute resource', async function () {
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithoutToken
                        },
                        computeSection: {
                            gpu: {
                                devicesInput: '3'
                            }
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithoutToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal(
                        huggingFaceDefaultProtocol
                    );
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.be.equal(3);
                    expect(inferenceWorkload.spec.compute.cpuMemoryRequest).to.be.equal('0M');

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should create hugging face workload with compute resource from asset', async function () {
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithoutToken
                        },
                        computeSection: {
                            assetName: 'cpu-only'
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithoutToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal(
                        huggingFaceDefaultProtocol
                    );
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.compute.cpuCoreRequest).to.be.equal(0.1);
                    expect(inferenceWorkload.spec.compute.cpuMemoryRequest).to.be.equal('100M');

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should create hugging face workload with new compute resource asset', async function () {
                    const computeName = objectGenerator.generateNameByPrefix('custom-compute');
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithToken,
                            hfToken
                        },
                        computeSection: {
                            newComputeResource: {
                                assetName: computeName,
                                projectName: project.name,
                                gpu: {
                                    devicesInput: 3
                                }
                            }
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: undefined,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal(
                        huggingFaceDefaultProtocol
                    );
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.be.equal(1);
                    expect(inferenceWorkload.spec.compute.gpuPortionRequest).to.be.equal(0.03);
                    expect(inferenceWorkload.spec.compute.gpuRequestType).to.be.equal('portion');

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);

                    // cleanup
                    const computeId = await computeResourceUtils.getId(computeName);
                    await computeResourceUtils.delete(computeId);
                    await computeResourceUtils.waitForComputeResourceToBeDeleted(computeName);
                });

                test('should create hugging face workload model with nfs storage', async function () {
                    const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithoutToken
                        },
                        dataSourceOptions: {
                            nfs: dataSourceName
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithoutToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: '/test/sample', // this value is taken from the data storage path
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);
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

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should create hugging face workload model with pvc storage', async function () {
                    const dataSourceOptions: IDataSourceOptions = {
                        volume: {
                            accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                            sizeUnit: 'MB',
                            path: '/home'
                        }
                    };

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

                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithoutToken
                        },
                        dataSourceOptions: {
                            assetNames: [pvc.meta.name]
                        }
                    });

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // HF specific data
                    expect(inferenceWorkload.spec.image).to.be.equal(vllmImage);
                    expect(inferenceWorkload.spec.args).to.be.equal(
                        `--model ${huggingFaceModelNameWithoutToken}`
                    );
                    expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                        {
                            name: 'HF_HOME',
                            value: dataSourceOptions.volume.path, // this value is taken from the data storage path
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ]);

                    expect(inferenceWorkload.spec.servingPort.container).to.be.equal(8000);
                    expect(inferenceWorkload.spec.servingPort.protocol).to.be.equal('http');
                    expect(inferenceWorkload.spec.nodePools).to.deep.equal(['default']);
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equal(1);
                    expect(inferenceWorkload.spec.storage.pvc).to.deep.equal([
                        {
                            claimInfo: undefined,
                            claimName: pvcSpec.claimName,
                            dataSharing: undefined,
                            ephemeral: false,
                            exclude: undefined,
                            existingPvc: true,
                            name: undefined,
                            path: dataSourceOptions.volume.path,
                            readOnly: false
                        }
                    ]);

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);

                    await dataSourcesUtils.deletePvcAsset(pvc.meta.id);
                    await dataSourcesUtils.waitForDataSourceToBeDeleted(pvcName);
                });

                test('should continue to advanced setup when create hugging face workload', async function () {
                    const servingPort = {
                        access: {
                            option: IServingPortAccessOptions.Users
                        }
                    };

                    const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithToken,
                            hfToken,
                            servingPort,
                            shouldOpenServingPortSection: true
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
                        {
                            name: 'HF_HOME',
                            value: '/test/sample', // this value is taken from the data storage path
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: hfToken,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ];
                    await huggingFaceElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                    await inference.clickOnCreateInferenceBtn();

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expectHuggingFaceWorkload(
                        inferenceWorkload,
                        inferenceName,
                        vllmImage,
                        `--model ${huggingFaceModelNameWithToken}`,
                        environmentVariables
                    );
                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should continue to advanced setup and create new data source for hugging face', async function () {
                    const servingPort = {
                        access: {
                            option: IServingPortAccessOptions.Users
                        }
                    };

                    const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithToken,
                            hfToken,
                            servingPort,
                            shouldOpenServingPortSection: true
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
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue, // default value when no data source is provided
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: hfToken,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ];
                    await huggingFaceElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                    // go to a different page
                    await inferenceFlatUIUtil.fillDataSourceData({ nfs: dataSourceName });
                    await inference.clickOnCreateInferenceBtn();

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    // we want to make sure that all the data is still there
                    expectHuggingFaceWorkload(
                        inferenceWorkload,
                        inferenceName,
                        vllmImage,
                        `--model ${huggingFaceModelNameWithToken}`,
                        environmentVariables // here the HF_HOME is still the default value
                    );

                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });

                test('should support modifying advanced setup fields, navigating back to hugging face configuration, and completing submission with the origin data', async function () {
                    const servingPort = {
                        access: {
                            option: IServingPortAccessOptions.Users
                        }
                    };

                    const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithToken,
                            hfToken,
                            servingPort,
                            shouldOpenServingPortSection: true
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
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue, // default value when no data source is provided
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: hfToken,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ];

                    await huggingFaceElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                    // change the environment and compute data
                    await inferenceFlatUIUtil.fillEnvironmentData({ assetName: 'gpt2' });
                    await inferenceFlatUIUtil.fillComputeData({ assetName: 'two-gpus' });
                    await inferenceFlatUIUtil.fillDataSourceData({ nfs: dataSourceName });

                    // go back to the hugging face page
                    await inference.locator('[aid="model-section"] [aid="back-btn"]').click();

                    // submit the inference
                    await inference.clickOnCreateInferenceBtn();

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    expectHuggingFaceWorkload(
                        inferenceWorkload,
                        inferenceName,
                        vllmImage,
                        `--model ${huggingFaceModelNameWithToken}`,
                        [
                            {
                                name: 'HF_HOME',
                                value: hfHomeDefaultValue, // default value when no data source is provided
                                secret: undefined,
                                configMap: undefined,
                                podFieldRef: undefined,
                                userCredential: undefined,
                                description: undefined,
                                exclude: undefined
                            },
                            {
                                name: 'HF_TOKEN',
                                value: hfToken,
                                secret: undefined,
                                configMap: undefined,
                                podFieldRef: undefined,
                                userCredential: undefined,
                                description: undefined,
                                exclude: undefined
                            }
                        ],
                        false
                    );
                    logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                });


                test('should clone hugging face workload correctly', async function () {
                    const copiedInferenceName = objectGenerator.generateShortNameByPrefix('copied-inference');
                    const servingPort = {
                        access: {
                            option: IServingPortAccessOptions.Users
                        }
                    };

                    const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                    await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                        clusterName: flowRootState.cluster.name,
                        projectName: project.name,
                        workloadName: inferenceName,
                        inferenceType: IInferenceType.VLLM,
                        modelSection: {
                            modelName: huggingFaceModelNameWithToken,
                            hfToken,
                            servingPort,
                            shouldOpenServingPortSection: true
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

                    const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );

                    const environmentVariables = [
                        {
                            name: 'HF_HOME',
                            value: '/test/sample', // this value is taken from the data storage path
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: hfToken,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ];

                    expectHuggingFaceWorkload(
                        inferenceWorkload,
                        inferenceName,
                        vllmImage,
                        `--model ${huggingFaceModelNameWithToken}`,
                        environmentVariables
                    );

                    await workload.toggleWorkload(inferenceName);
                    await workload.clickOnCopyAndEditWorkload();

                    await inference.fillInferenceName(copiedInferenceName);

                    await huggingFaceElementValidationUtils.expectEnvironmentVariables([
                        {
                            name: 'HF_HOME',
                            value: '/test/sample', // this value is taken from the data storage path
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: '', // should be empty
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                    ]);

                    // fill the HF_TOKEN value
                    const environmentVariableBoxes = inference.locator('.environment-variable-box');
                    const hfTokenKeyIndex = environmentVariables.findIndex((env) => env.name === 'HF_TOKEN');
                    const hfTokenKeyBox = environmentVariableBoxes.nth(hfTokenKeyIndex);
                    const hfTokenKeyValueInput = hfTokenKeyBox.locator(
                        hybridForm.hybridEnvironmentSection.envVarValue
                    );
                    await hfTokenKeyValueInput.fill(hfToken);
                    await inference.clickOnCreateInferenceBtn();

                    const copiedInferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                        copiedInferenceName,
                        project.id.toString()
                    );

                    expectHuggingFaceWorkload(
                        copiedInferenceWorkload,
                        copiedInferenceName,
                        vllmImage,
                        `--model ${huggingFaceModelNameWithToken}`,
                        environmentVariables
                    );

                    await workloadUtils.deleteMultipleWorkloads([copiedInferenceWorkload.workloadId]);
                    logger.debug(`Hugging face inference: ${copiedInferenceName} copied successfully!`);
                });

                supportedFrom(SupportedVersions.V2_23)(function () {
                    test.describe('hugging face advanced setup request timeout', function () {
                        test('should continue to advanced setup when create hugging face workload with custom request timeout values', async function () {
                            await inferenceFlatUIUtil.createHuggingFaceFlatWorkload({
                                clusterName: flowRootState.cluster.name,
                                projectName: project.name,
                                workloadName: inferenceName,
                                inferenceType: IInferenceType.VLLM,
                                modelSection: {
                                    modelName: huggingFaceModelNameWithoutToken
                                },
                                advancedSetup: true
                            });

                            logger.debug('Filling general section (initialization timeout 5m 7s, request timeout 2m 30s)');
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
                            logger.debug('General section filled');

                            // submit the inference
                            logger.debug('Waiting for Create Inference button to exist');
                            await waitUtils.waitForElementToExist(inference, inference.createInferenceBtn);
                            logger.debug('Create Inference button found, clicking');
                            await inference.clickOnCreateInferenceBtn();
                            logger.debug('Create Inference button clicked');

                            const inferenceWorkload: Inference = await workloadsCrudUtils.getInferenceByName(
                                inferenceName,
                                project.id.toString()
                            );
                            expect(inferenceWorkload).to.not.be.undefined;
                            logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                            expect(
                                inferenceWorkload.spec.servingConfiguration?.initializationTimeoutSeconds
                            ).to.equal(307); // 5 minutes * 60 (seconds) + 7 seconds
                            expect(
                                inferenceWorkload.spec.servingConfiguration?.requestTimeoutSeconds
                            ).to.equal(150); // 2 minutes * 60 (seconds) + 30 seconds

                            logger.debug(`Hugging face inference: ${inferenceName} created successfully!`);
                        });
                    });
                });
            });
        });
    }
);
