/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Personal Credentials Plugin — UI trigger surface (workload v2 submission form): credentials
 * attached via the "My credentials" env-var source and the private image-pull picker.
 * Same contract as the API/CLI suites: ANY_WORKLOAD_SUBMISSION routes the submission through a
 * WorkloadRunner CRD and every dest Secret must be ownerRef'd to it (a native fallback owns nothing).
 * YAML-editor submissions stay out of scope (docs/personal-credentials-ui-manual-test.zh-CN.md).
 */
import { test, expect } from '@playwright/test';
import {
    config,
    credentialUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    personalCredentialsUtils,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import {
    DockerRegistry,
    GenericSecret,
    NgcApiKey,
    WorkloadType
} from '../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadType as CrudWorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { ECredentialsSource } from '../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { ECredentialType } from '../../../../../../modules/helpers/models/credential.models';
import { kubeCli } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    enableAnyWorkloadSubmission,
    RestoreAnyWorkload
} from '../../regression/viper/anyWorkloadSubmission';

// §1 P1 — v2-form trigger: generic secret (multi-key) + docker registry on a training, NGC API key on a workspace.
test.describe(
    `Personal Credentials Plugin via UI v2 submission form, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const HF_KEY = 'HF_TOKEN';
            const HF_VALUE = 'hf-personal-token-value';
            const WANDB_KEY = 'WANDB_API_KEY';
            const WANDB_VALUE = 'wandb-personal-key-value';
            const NGC_KEY = 'NGC_API_KEY';
            const NGC_VALUE = 'ngc-personal-key-value';
            // Created via the asset API (no shell), so the literal '$oauthtoken' is safe.
            const DOCKER_USER = '$oauthtoken';
            const DOCKER_PASSWORD = 'dummy-ngc-pull-key';
            const DOCKER_URL = 'https://nvcr.io';

            // The bogus nvcr.io credential can't break this pull — the kubelet only consults
            // imagePullSecrets whose registry matches the image host.
            const IMAGE = config.quickStartImage.training;
            const COMMAND = 'sleep infinity';
            const COMPUTE_RESOURCE = 'cpu-only';
            const RUNTIME_SETTINGS = 'Runtime settings';
            const IMAGE_PULL = 'Image pull';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericSecret: GenericSecret;
            let ngcApiKey: NgcApiKey;
            let dockerRegistry: DockerRegistry;
            let genericCredName: string;
            let ngcCredName: string;
            let dockerCredName: string;

            let browserEngine: BrowserEngine;
            let authPage: AuthPage;
            let workloadPage: WorkloadPage;
            let hybridForm: HybridForm;

            let workloadName: string;
            let workloadCreated: boolean;
            let currentCrudType: CrudWorkloadType;
            let videoName: string;

            // Page 1 + image + command; leaves the form open for credential-specific sections.
            const startFormSubmission = async (type: WorkloadType): Promise<void> => {
                await workloadPage.navigateToWorkloads();
                await hybridForm.fillWorkloadCreationFirstPage(
                    flowRootState.cluster.name,
                    project.name,
                    workloadName,
                    type
                );
                await hybridForm.fillImageInput(IMAGE);
                await hybridForm.clickOnSubSectionByLabel(RUNTIME_SETTINGS);
                await hybridForm.clickOnCommandAndArguments();
                await hybridForm.fillCmdInput(COMMAND);
            };

            const submitAndWaitForRunning = async (): Promise<void> => {
                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel(COMPUTE_RESOURCE);
                await hybridForm.clickOnSubmit();
                workloadCreated = true;
                try {
                    await workloadPage.waitForWorkloadUIStatus(workloadName);
                } catch (error) {
                    await personalCredentialsUtils.logSubmissionDiagnostics(
                        project.namespace,
                        workloadName
                    );
                    throw error;
                }
            };

            const expectWorkloadRunnerExists = async (): Promise<void> => {
                await waitUtils.waitForCondition(
                    async () =>
                        (await personalCredentialsUtils.findRunnerByWorkloadName(
                            project.namespace,
                            workloadName
                        )) !== undefined,
                    `WorkloadRunner CRD for ${workloadName} (absence => native fallback, plugin never ran)`,
                    true,
                    90,
                    5
                );
            };

            test.beforeAll('enable v2 path, create project and the three personal credentials', async () => {
                // The plugin only runs on the v2 path gated by ANY_WORKLOAD_SUBMISSION — same contract as API/CLI.
                restoreAnyWorkload = await enableAnyWorkloadSubmission();
                try {
                    project = await projectResourcesUtils.createProjectAndNamespace('pc-form');

                    genericCredName = objectGenerator.generateShortNameByPrefix('pc-form-gs', 4);
                    ngcCredName = objectGenerator.generateShortNameByPrefix('pc-form-ngc', 4);
                    dockerCredName = objectGenerator.generateShortNameByPrefix('pc-form-dr', 4);

                    genericSecret = await credentialUtils.createGenericSecretAsset(
                        personalCredentialsUtils.genericSecretPayload(
                            genericCredName,
                            [
                                { key: HF_KEY, value: HF_VALUE },
                                { key: WANDB_KEY, value: WANDB_VALUE }
                            ],
                            'Personal multi-key generic secret (HF_TOKEN + WANDB_API_KEY)'
                        )
                    );
                    ngcApiKey = await credentialUtils.createNgcApiKeyAsset(
                        personalCredentialsUtils.ngcApiKeyPayload(ngcCredName, NGC_VALUE)
                    );
                    dockerRegistry = await credentialUtils.createDockerRegistryAsset(
                        personalCredentialsUtils.dockerRegistryPayload(
                            dockerCredName,
                            DOCKER_USER,
                            DOCKER_PASSWORD,
                            DOCKER_URL
                        )
                    );

                    await credentialUtils.waitForPersonalCredentialSecrets(
                        genericCredName,
                        personalCredentialsUtils.systemAdminAccessToken()
                    );
                    await credentialUtils.waitForPersonalCredentialSecrets(
                        ngcCredName,
                        personalCredentialsUtils.systemAdminAccessToken()
                    );
                    await credentialUtils.waitForPersonalCredentialSecrets(
                        dockerCredName,
                        personalCredentialsUtils.systemAdminAccessToken()
                    );
                } catch (error) {
                    if (restoreAnyWorkload) {
                        await restoreAnyWorkload();
                        restoreAnyWorkload = undefined;
                    }
                    throw error;
                }
            });

            test.beforeEach(async () => {
                workloadName = objectGenerator.generateShortNameByPrefix('pc-form-wl', 4);
                workloadCreated = false;
                currentCrudType = CrudWorkloadType.TRAINING;
                videoName = `pc-form-${workloadName}`;

                browserEngine = new BrowserEngine();
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                workloadPage = new WorkloadPage(browserEngine.currentPage);
                hybridForm = new HybridForm(browserEngine.currentPage);

                // Private credentials are visible only to their creator — sign in as the admin that created them.
                await authPage.signIn(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
            });

            test.afterEach('archive browser, delete leftover workload', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
                // API-side safety net for mid-test failures; the happy path deletes via the UI.
                if (workloadCreated) {
                    try {
                        await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                            workloadName,
                            project.id,
                            currentCrudType
                        );
                    } catch (e) {
                        logger.error(`Failed to delete workload ${workloadName}`, e);
                    }
                }
            });

            test.afterAll('delete personal credentials, project, and restore v2 switch', async () => {
                await Promise.allSettled([
                    credentialUtils
                        .deleteGenericSecretAsset(genericSecret.meta.id)
                        .catch((e) => logger.error('Failed to delete generic-secret credential', e)),
                    credentialUtils
                        .deleteNgcApiKeyAsset(ngcApiKey.meta.id)
                        .catch((e) => logger.error('Failed to delete ngc-api-key credential', e)),
                    credentialUtils
                        .deleteDockerRegistryAsset(dockerRegistry.meta.id)
                        .catch((e) => logger.error('Failed to delete docker-registry credential', e))
                ]);
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (e) {
                    logger.error(`Failed to delete project ${project.name}`, e);
                }
                if (restoreAnyWorkload) {
                    await restoreAnyWorkload();
                }
            });

            test('T6141006: submit a training via the v2 form with generic-secret (multi-key) env vars + docker-registry image pull secret', async () => {
                // Sign-in + form + Running wait + UI delete + GC polls exceed the 150s self-hosted default.
                test.setTimeout(300_000);
                currentCrudType = CrudWorkloadType.TRAINING;
                videoName = `pc-form-multi-${workloadName}`;

                await test.step('Fill the v2 form: image, command, two env vars from the personal generic secret, private image pull secret', async () => {
                    await startFormSubmission(WorkloadType.Training);
                    // Two env vars off the same credential, different keys — must merge into ONE dest Secret.
                    await hybridForm.addEnvVarFromPersonalCredential(HF_KEY, genericCredName, {
                        key: HF_KEY
                    });
                    await hybridForm.addEnvVarFromPersonalCredential(WANDB_KEY, genericCredName, {
                        key: WANDB_KEY
                    });
                    await hybridForm.clickOnSubSectionByLabel(IMAGE_PULL);
                    await hybridForm.selectImagePullSecretWithSource(
                        ECredentialsSource.Private,
                        dockerCredName
                    );
                });

                await test.step('Submit and wait for Running in the workloads grid', async () => {
                    await submitAndWaitForRunning();
                });

                await test.step('v2-path discriminator: a WorkloadRunner CRD exists (not a native fallback)', async () => {
                    await expectWorkloadRunnerExists();
                });

                let genericDestSecretName: string;
                let dockerDestSecretName: string;

                await test.step('Generic multi-key: two env vars -> one dest Secret, different keys, decoded values, owned by WorkloadRunner', async () => {
                    const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                    const podEnv = pod.spec?.containers?.[0]?.env ?? [];
                    const hfEnv = podEnv.find((e) => e.name === HF_KEY);
                    const wandbEnv = podEnv.find((e) => e.name === WANDB_KEY);
                    expect(hfEnv?.valueFrom?.secretKeyRef, `env ${HF_KEY} secretKeyRef`).toBeDefined();
                    expect(wandbEnv?.valueFrom?.secretKeyRef, `env ${WANDB_KEY} secretKeyRef`).toBeDefined();

                    // Multi-key proof: same dest Secret name, different secretKeyRef.key.
                    expect(hfEnv!.valueFrom!.secretKeyRef!.key).toBe(HF_KEY);
                    expect(wandbEnv!.valueFrom!.secretKeyRef!.key).toBe(WANDB_KEY);
                    expect(hfEnv!.valueFrom!.secretKeyRef!.name).toBe(
                        wandbEnv!.valueFrom!.secretKeyRef!.name
                    );
                    expect(hfEnv!.valueFrom!.secretKeyRef!.name!.toLowerCase()).toContain(
                        genericCredName.toLowerCase()
                    );

                    genericDestSecretName = hfEnv!.valueFrom!.secretKeyRef!.name!;
                    const genSecret = await kubeCli.getSecret(genericDestSecretName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(genSecret.data[HF_KEY])).toBe(HF_VALUE);
                    expect(personalCredentialsUtils.decodeSecret(genSecret.data[WANDB_KEY])).toBe(
                        WANDB_VALUE
                    );
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(genSecret.metadata?.ownerReferences),
                        'generic dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
                    ).toBe(true);
                });

                await test.step('dockerRegistry wired as a dockerconfigjson imagePullSecret owned by WorkloadRunner', async () => {
                    const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                    const imagePullSecrets = pod.spec?.imagePullSecrets ?? [];
                    const dockerIps = imagePullSecrets.find((s) =>
                        s.name?.toLowerCase().includes(dockerCredName.toLowerCase())
                    );
                    expect(
                        dockerIps,
                        `imagePullSecret for ${dockerCredName}; actual: ${imagePullSecrets
                            .map((s) => s.name)
                            .join(', ')}`
                    ).toBeDefined();

                    dockerDestSecretName = dockerIps!.name!;
                    const dockerSecret = await kubeCli.getSecret(dockerDestSecretName, project.namespace);
                    expect(dockerSecret.type).toBe('kubernetes.io/dockerconfigjson');
                    expect(dockerSecret.data['.dockerconfigjson']).toBeDefined();
                    const dockerConfig = JSON.parse(
                        personalCredentialsUtils.decodeSecret(dockerSecret.data['.dockerconfigjson'])
                    ) as {
                        auths: Record<string, { username: string; password: string; auth: string }>;
                    };
                    const registryAuth = dockerConfig.auths[DOCKER_URL];
                    expect(
                        registryAuth,
                        `dockerconfigjson must carry the ${DOCKER_URL} registry from the personal cred`
                    ).toBeDefined();
                    expect(registryAuth.username).toBe(DOCKER_USER);
                    expect(registryAuth.password).toBe(DOCKER_PASSWORD);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(
                            dockerSecret.metadata?.ownerReferences
                        ),
                        'docker dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
                    ).toBe(true);
                });

                await test.step('Delete the workload via the UI and confirm both dest Secrets are garbage-collected', async () => {
                    await workloadPage.deleteWorkload(workloadName);
                    workloadCreated = false;
                    for (const destName of [genericDestSecretName, dockerDestSecretName]) {
                        await personalCredentialsUtils.waitForSecretGone(
                            project.namespace,
                            destName,
                            `dest Secret ${destName} garbage-collected after UI workload delete`
                        );
                    }
                });
            });

            test('T6141021: submit a workspace via the v2 form with an NGC API key env var', async () => {
                test.setTimeout(300_000);
                currentCrudType = CrudWorkloadType.WORKSPACE;
                videoName = `pc-form-ngc-${workloadName}`;

                await test.step('Fill the v2 form: image, command, one env var from the personal NGC API key credential', async () => {
                    await startFormSubmission(WorkloadType.Workspace);
                    // The Type filter defaults to Generic secret; NGC credentials render no key input.
                    await hybridForm.addEnvVarFromPersonalCredential(NGC_KEY, ngcCredName, {
                        credentialType: ECredentialType.NgcApiKey
                    });
                });

                await test.step('Submit and wait for Running in the workloads grid', async () => {
                    await submitAndWaitForRunning();
                });

                await test.step('v2-path discriminator: a WorkloadRunner CRD exists (not a native fallback)', async () => {
                    await expectWorkloadRunnerExists();
                });

                let ngcDestSecretName = '';

                await test.step('ngcApiKey env var wired to its own dest Secret, decoded value, owned by WorkloadRunner', async () => {
                    const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                    const podEnv = pod.spec?.containers?.[0]?.env ?? [];
                    const ngcEnv = podEnv.find((e) => e.name === NGC_KEY);
                    expect(ngcEnv?.valueFrom?.secretKeyRef, `env ${NGC_KEY} secretKeyRef`).toBeDefined();
                    expect(ngcEnv!.valueFrom!.secretKeyRef!.key).toBe(NGC_KEY);
                    expect(ngcEnv!.valueFrom!.secretKeyRef!.name!.toLowerCase()).toContain(
                        ngcCredName.toLowerCase()
                    );

                    ngcDestSecretName = ngcEnv!.valueFrom!.secretKeyRef!.name!;
                    const ngcSecret = await kubeCli.getSecret(ngcDestSecretName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(ngcSecret.data[NGC_KEY])).toBe(NGC_VALUE);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(ngcSecret.metadata?.ownerReferences),
                        'ngc dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
                    ).toBe(true);
                });

                await test.step('Delete the workload via the UI and confirm the dest Secret is garbage-collected', async () => {
                    await workloadPage.deleteWorkload(workloadName);
                    workloadCreated = false;
                    await personalCredentialsUtils.waitForSecretGone(
                        project.namespace,
                        ngcDestSecretName,
                        `dest Secret ${ngcDestSecretName} garbage-collected after UI workload delete`
                    );
                });
            });
        });
    }
);
