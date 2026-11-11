/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Personal Credentials Plugin — API trigger surface (typed CRUD createTraining).
 *
 * The personalcreds plugin only runs on the Workload API v2 path, gated by the deployment master switch
 * ANY_WORKLOAD_SUBMISSION (default OFF). Each suite flips it ON in beforeAll (restored in afterAll) so
 * native createTraining routes through a WorkloadRunner CRD; every materialized dest Secret must be
 * ownerRef'd to that runner — the discriminator a silent native fallback cannot satisfy.
 */
import { test, expect } from '@playwright/test';
import {
    authorizationUtils,
    config,
    credentialUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    PC_ANNOTATION,
    PC_WLR,
    personalCredentialsUtils,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { DockerRegistry, GenericSecret, NgcApiKey } from '../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { TrainingSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/trainingSpecSpec';
import { ImagePullPolicy } from '../../../../../../modules/api/apis/workloadsApi/generated/model/imagePullPolicy';
import { WorkloadType as CrudWorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { kubeCli } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    enableAnyWorkloadSubmission,
    RestoreAnyWorkload
} from '../../regression/viper/anyWorkloadSubmission';
import { StatusCodes } from 'http-status-codes';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { AccessRule } from '../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
// Shared workload shape across every API suite below.
const IMAGE = config.quickStartImage.training;
const COMMAND = 'sleep infinity';
const CPU_COMPUTE = { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' };

// §1 P0 — one Job referencing genericSecret(multi-key HF_TOKEN+WANDB_API_KEY) + ngcApiKey + dockerRegistry.
// Multi-key proof: the two env vars share ONE dest Secret with different secretKeyRef.key.
test.describe(
    `Personal Credentials Plugin multi-credential, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const HF_KEY = 'HF_TOKEN';
            const HF_VALUE = 'hf-personal-token-value';
            const WANDB_KEY = 'WANDB_API_KEY';
            const WANDB_VALUE = 'wandb-personal-key-value';
            const NGC_KEY = 'NGC_API_KEY';
            const NGC_VALUE = 'ngc-personal-key-value';
            const DOCKER_USER = '$oauthtoken';
            const DOCKER_PASSWORD = 'dummy-ngc-pull-key';
            const DOCKER_URL = 'https://nvcr.io';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericSecret: GenericSecret;
            let ngcApiKey: NgcApiKey;
            let dockerRegistry: DockerRegistry;
            let genericCredName: string;
            let ngcCredName: string;
            let dockerCredName: string;
            let workloadName: string;
            let workloadCreated: boolean;

            const deleteWorkload = async (): Promise<void> => {
                await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                    workloadName,
                    project.id,
                    CrudWorkloadType.TRAINING
                );
            };

            test.beforeAll('enable v2 path, create project and the three personal credentials', async () => {
                restoreAnyWorkload = await enableAnyWorkloadSubmission();
                try {
                    project = await projectResourcesUtils.createProjectAndNamespace('pc-multi');
                    genericCredName = objectGenerator.generateShortNameByPrefix('pc-multi-gs', 4);
                    ngcCredName = objectGenerator.generateShortNameByPrefix('pc-multi-ngc', 4);
                    dockerCredName = objectGenerator.generateShortNameByPrefix('pc-multi-dr', 4);

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

                    // Wait for the agent sync loop to materialize each encrypted source Secret.
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

            test.beforeEach(() => {
                workloadName = objectGenerator.generateShortNameByPrefix('pc-multi-wl', 4);
                workloadCreated = false;
            });

            test.afterEach('delete workload if still present', async () => {
                if (!workloadCreated) {
                    return;
                }
                try {
                    await deleteWorkload();
                } catch (e) {
                    logger.error(`Failed to delete workload ${workloadName}`, e);
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

            test('T6141004: submit one workload referencing genericSecret(multi-key) + ngcApiKey + dockerRegistry', async () => {
                const spec: TrainingSpecSpec = {
                    image: IMAGE,
                    command: COMMAND,
                    compute: CPU_COMPUTE,
                    environmentVariables: [
                        { name: HF_KEY, userCredential: { name: genericCredName, key: HF_KEY } },
                        { name: WANDB_KEY, userCredential: { name: genericCredName, key: WANDB_KEY } },
                        { name: NGC_KEY, userCredential: { name: ngcCredName, key: NGC_KEY } }
                    ],
                    imagePullSecrets: [{ name: dockerCredName, userCredential: true }]
                };

                await test.step('Submit the multi-credential training workload via the typed API', async () => {
                    workloadsCrudUtils.accessToken = personalCredentialsUtils.systemAdminAccessToken();
                    const { body: training } = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        spec
                    );
                    workloadCreated = true;
                    expect(training.name).toBe(workloadName);
                });

                await test.step('Wait for the workload pod to reach Running', async () => {
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadName, 120);
                });

                await test.step('v2-path discriminator: a WorkloadRunner CRD exists (not a native fallback)', async () => {
                    let runner: Record<string, any> | undefined;
                    await waitUtils.waitForCondition(
                        async () => {
                            runner = await personalCredentialsUtils.findRunnerByWorkloadName(
                                project.namespace,
                                workloadName
                            );
                            return runner !== undefined;
                        },
                        `WorkloadRunner CRD for ${workloadName} (absence => silent native fallback, plugin never ran)`,
                        true,
                        60
                    );
                    expect(
                        runner,
                        'No WorkloadRunner CRD: the submission fell back to the native path and the personalcreds plugin never ran'
                    ).toBeDefined();
                });

                let genericDestSecretName: string;
                let ngcDestSecretName: string;
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
                    expect(ngcDestSecretName).not.toBe(genericDestSecretName);
                    const ngcSecret = await kubeCli.getSecret(ngcDestSecretName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(ngcSecret.data[NGC_KEY])).toBe(NGC_VALUE);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(ngcSecret.metadata?.ownerReferences),
                        'ngc dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
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

                await test.step('Delete the workload and confirm all three dest Secrets are garbage-collected', async () => {
                    await deleteWorkload();
                    workloadCreated = false;
                    for (const destName of [genericDestSecretName, ngcDestSecretName, dockerDestSecretName]) {
                        await personalCredentialsUtils.waitForSecretGone(
                            project.namespace,
                            destName,
                            `dest Secret ${destName} garbage-collected after workload delete`
                        );
                    }
                });
            });
        });
    }
);

// §1 P2 (rewrite of the multi-container case) — two workloads sharing the SAME three personal credentials each
// get their OWN dest Secrets (destName carries the runner → no cross-runner collision), with per-runner GC.
test.describe(
    `Personal Credentials Plugin multi-credential shared by two workloads, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const HF_KEY = 'HF_TOKEN';
            const HF_VALUE = 'hf-personal-token-value';
            const WANDB_KEY = 'WANDB_API_KEY';
            const WANDB_VALUE = 'wandb-personal-key-value';
            const NGC_KEY = 'NGC_API_KEY';
            const NGC_VALUE = 'ngc-personal-key-value';
            const DOCKER_USER = '$oauthtoken';
            const DOCKER_PASSWORD = 'dummy-ngc-pull-key';
            const DOCKER_URL = 'https://nvcr.io';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericSecret: GenericSecret;
            let ngcApiKey: NgcApiKey;
            let dockerRegistry: DockerRegistry;
            let genericCredName: string;
            let ngcCredName: string;
            let dockerCredName: string;
            let workloadA: string;
            let workloadB: string;
            let createdWorkloadNames: string[];

            const sharedSpec = (): TrainingSpecSpec => ({
                image: IMAGE,
                command: COMMAND,
                compute: CPU_COMPUTE,
                environmentVariables: [
                    { name: HF_KEY, userCredential: { name: genericCredName, key: HF_KEY } },
                    { name: WANDB_KEY, userCredential: { name: genericCredName, key: WANDB_KEY } },
                    { name: NGC_KEY, userCredential: { name: ngcCredName, key: NGC_KEY } }
                ],
                imagePullSecrets: [{ name: dockerCredName, userCredential: true }]
            });

            const submitTraining = async (name: string): Promise<void> => {
                workloadsCrudUtils.accessToken = personalCredentialsUtils.systemAdminAccessToken();
                const { body: training } = await workloadsCrudUtils.createTraining(
                    name,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    sharedSpec()
                );
                createdWorkloadNames.push(name);
                expect(training.name).toBe(name);
            };

            const deleteWorkload = async (name: string): Promise<void> => {
                await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                    name,
                    project.id,
                    CrudWorkloadType.TRAINING
                );
                createdWorkloadNames = createdWorkloadNames.filter((n) => n !== name);
            };

            // Resolve a workload's three dest Secret names from its first pod (env secretKeyRefs + imagePullSecret).
            const destsOf = async (
                workloadName: string
            ): Promise<{ generic: string; ngc: string; docker: string }> => {
                const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                const env = pod.spec?.containers?.[0]?.env ?? [];
                const ips = pod.spec?.imagePullSecrets ?? [];
                return {
                    generic: env.find((e) => e.name === HF_KEY)!.valueFrom!.secretKeyRef!.name!,
                    ngc: env.find((e) => e.name === NGC_KEY)!.valueFrom!.secretKeyRef!.name!,
                    docker: ips.find((s) => s.name?.toLowerCase().includes(dockerCredName.toLowerCase()))!
                        .name!
                };
            };

            test.beforeAll(
                'enable v2 path, create project and the three shared personal credentials',
                async () => {
                    restoreAnyWorkload = await enableAnyWorkloadSubmission();
                    try {
                        project = await projectResourcesUtils.createProjectAndNamespace('pc-shared');
                        genericCredName = objectGenerator.generateShortNameByPrefix('pc-shared-gs', 4);
                        ngcCredName = objectGenerator.generateShortNameByPrefix('pc-shared-ngc', 4);
                        dockerCredName = objectGenerator.generateShortNameByPrefix('pc-shared-dr', 4);

                        genericSecret = await credentialUtils.createGenericSecretAsset(
                            personalCredentialsUtils.genericSecretPayload(
                                genericCredName,
                                [
                                    { key: HF_KEY, value: HF_VALUE },
                                    { key: WANDB_KEY, value: WANDB_VALUE }
                                ],
                                'Personal multi-key generic secret shared by two workloads'
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
                }
            );

            test.beforeEach(() => {
                workloadA = objectGenerator.generateShortNameByPrefix('pc-shared-a', 4);
                workloadB = objectGenerator.generateShortNameByPrefix('pc-shared-b', 4);
                createdWorkloadNames = [];
            });

            test.afterEach('delete any workloads still present', async () => {
                for (const name of [...createdWorkloadNames]) {
                    try {
                        await deleteWorkload(name);
                    } catch (e) {
                        logger.error(`Failed to delete workload ${name}`, e);
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

            test('T6141007: two workloads sharing the same three personal credentials get independent per-runner dest Secrets', async () => {
                await test.step('Submit two workloads, both referencing the same three personal credentials', async () => {
                    await submitTraining(workloadA);
                    await submitTraining(workloadB);
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadA, 120);
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadB, 120);
                });

                let a: { generic: string; ngc: string; docker: string };
                let b: { generic: string; ngc: string; docker: string };

                await test.step('Each workload materializes its OWN three dest Secrets (destName carries the runner → no cross-runner collision)', async () => {
                    a = await destsOf(workloadA);
                    b = await destsOf(workloadB);
                    expect(a.generic, 'generic dest names must differ across runners').not.toBe(b.generic);
                    expect(a.ngc, 'ngc dest names must differ across runners').not.toBe(b.ngc);
                    expect(a.docker, 'docker dest names must differ across runners').not.toBe(b.docker);
                });

                await test.step('Every dest Secret decodes correctly and is owned only by its own runner', async () => {
                    const runnerA = (await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadA
                    ))!.metadata?.name as string;
                    const runnerB = (await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadB
                    ))!.metadata?.name as string;
                    const cases: {
                        dests: { generic: string; ngc: string; docker: string };
                        runner: string;
                    }[] = [
                        { dests: a, runner: runnerA },
                        { dests: b, runner: runnerB }
                    ];
                    for (const { dests, runner } of cases) {
                        const gen = await kubeCli.getSecret(dests.generic, project.namespace);
                        expect(personalCredentialsUtils.decodeSecret(gen.data[HF_KEY])).toBe(HF_VALUE);
                        expect(personalCredentialsUtils.decodeSecret(gen.data[WANDB_KEY])).toBe(WANDB_VALUE);
                        const ngc = await kubeCli.getSecret(dests.ngc, project.namespace);
                        expect(personalCredentialsUtils.decodeSecret(ngc.data[NGC_KEY])).toBe(NGC_VALUE);
                        const docker = await kubeCli.getSecret(dests.docker, project.namespace);
                        expect(docker.type).toBe('kubernetes.io/dockerconfigjson');
                        for (const secret of [gen, ngc, docker]) {
                            const owners = secret.metadata?.ownerReferences ?? [];
                            expect(owners.length, `${secret.metadata?.name} has exactly one ownerRef`).toBe(
                                1
                            );
                            expect(owners[0].kind).toBe(PC_WLR.kind);
                            expect(owners[0].name, `${secret.metadata?.name} owned by ${runner}`).toBe(
                                runner
                            );
                        }
                    }
                });

                await test.step('Delete only workload A → its three dest Secrets are GC’d; workload B and its Secrets are unaffected', async () => {
                    await deleteWorkload(workloadA);
                    for (const destName of [a.generic, a.ngc, a.docker]) {
                        await personalCredentialsUtils.waitForSecretGone(project.namespace, destName);
                    }
                    for (const destName of [b.generic, b.ngc, b.docker]) {
                        expect((await kubeCli.getSecret(destName, project.namespace)).metadata?.name).toBe(
                            destName
                        );
                    }
                    const podsB = await kubeCli.listPodsOfJob(project.namespace, workloadB, true);
                    expect(podsB[0]?.status?.phase).toBe('Running');
                });
            });
        });
    }
);

// §1 P2 — imagePull negative control, inverted. No genuinely-private image/credential exists on this env
// (every quickstart image is anon-pullable), so instead of "private image fails without the cred, succeeds
// with it" we prove the SAME runtime wiring in reverse: an anon-pullable Docker Hub image runs with no
// credential, then FAILS pull auth once a personal dockerRegistry credential carrying bogus creds for that
// registry is attached (imagePullPolicy Always defeats node-side image caching). Docker Hub's token service
// rejects invalid basic auth with 401 rather than falling back to anonymous (probed 2026-07-08), so a
// pull-time 401 proves the kubelet consulted the plugin's materialized dest Secret — strictly stronger than
// asserting pod-spec attachment. Inverted direction agreed with dev.
test.describe(
    `Personal Credentials Plugin imagePull negative control, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            // 'ubuntu' (no registry prefix) is pulled from Docker Hub; airgapped envs mirror it and the
            // 401-on-bogus-creds premise is only verified for Docker Hub, so skip there instead of failing
            // on unknown mirror auth semantics.
            const CONTROL_IMAGE = config.quickStartImage.ubuntu;
            // The canonical legacy Docker Hub key the kubelet credential keyring matches for docker.io images.
            const HUB_KEY = 'https://index.docker.io/v1/';
            const BOGUS_USER = 'e2e-bogus-user';
            const BOGUS_PASSWORD = 'e2e-bogus-password';
            const PULL_FAILURE_REASON = /ErrImagePull|ImagePullBackOff/;
            const PULL_AUTH_FAILURE =
                /401|unauthorized|authentication required|pull access denied|forbidden/i;

            test.skip(
                () => CONTROL_IMAGE.includes('/'),
                'control image is not pulled from Docker Hub, where the bogus-creds 401 premise was verified'
            );

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let dockerRegistry: DockerRegistry;
            let dockerCredName: string;
            let controlWorkload: string;
            let negativeWorkload: string;
            let createdWorkloadNames: string[];

            const trainingSpec = (withCredential: boolean): TrainingSpecSpec => ({
                image: CONTROL_IMAGE,
                command: COMMAND,
                compute: CPU_COMPUTE,
                // Always forces the kubelet to re-authenticate against the registry on every start, so a
                // node-cached image cannot mask the bogus-credential 401 with a silent local hit.
                imagePullPolicy: ImagePullPolicy.Always,
                ...(withCredential
                    ? { imagePullSecrets: [{ name: dockerCredName, userCredential: true }] }
                    : {})
            });

            const submitTraining = async (name: string, withCredential: boolean): Promise<void> => {
                workloadsCrudUtils.accessToken = personalCredentialsUtils.systemAdminAccessToken();
                const { body: training } = await workloadsCrudUtils.createTraining(
                    name,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    trainingSpec(withCredential)
                );
                createdWorkloadNames.push(name);
                expect(training.name).toBe(name);
            };

            const deleteWorkload = async (name: string): Promise<void> => {
                await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                    name,
                    project.id,
                    CrudWorkloadType.TRAINING
                );
                createdWorkloadNames = createdWorkloadNames.filter((n) => n !== name);
            };

            test.beforeAll(
                'enable v2 path, create project and the bogus docker-registry credential',
                async () => {
                    restoreAnyWorkload = await enableAnyWorkloadSubmission();
                    try {
                        project = await projectResourcesUtils.createProjectAndNamespace('pc-ipull');
                        dockerCredName = objectGenerator.generateShortNameByPrefix('pc-ip-dr', 4);
                        dockerRegistry = await credentialUtils.createDockerRegistryAsset(
                            personalCredentialsUtils.dockerRegistryPayload(
                                dockerCredName,
                                BOGUS_USER,
                                BOGUS_PASSWORD,
                                HUB_KEY,
                                'Personal docker-registry credential with bogus Docker Hub creds (negative control)'
                            )
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
                }
            );

            test.beforeEach(() => {
                controlWorkload = objectGenerator.generateShortNameByPrefix('pc-ip-ok', 4);
                negativeWorkload = objectGenerator.generateShortNameByPrefix('pc-ip-bad', 4);
                createdWorkloadNames = [];
            });

            test.afterEach('delete any workloads still present', async () => {
                for (const name of [...createdWorkloadNames]) {
                    try {
                        await deleteWorkload(name);
                    } catch (e) {
                        logger.error(`Failed to delete workload ${name}`, e);
                    }
                }
            });

            test.afterAll('delete the credential, project, and restore v2 switch', async () => {
                if (dockerRegistry) {
                    await credentialUtils
                        .deleteDockerRegistryAsset(dockerRegistry.meta.id)
                        .catch((e) => logger.error('Failed to delete docker-registry credential', e));
                }
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (e) {
                    logger.error(`Failed to delete project ${project.name}`, e);
                }
                if (restoreAnyWorkload) {
                    await restoreAnyWorkload();
                }
            });

            test('T6141008: imagePull negative control — image runs without the credential, fails pull auth with a bogus personal credential', async () => {
                await test.step('Control: the image runs with NO credential attached (anon-pullable premise holds on this env)', async () => {
                    await submitTraining(controlWorkload, false);
                    await kubeCli.waitForJobPodsRunning(project.namespace, controlWorkload, 120);
                    await deleteWorkload(controlWorkload);
                });

                await test.step('Submit the same workload WITH the bogus personal credential as imagePull secret', async () => {
                    await submitTraining(negativeWorkload, true);
                });

                await test.step('v2-path discriminator: a WorkloadRunner CRD exists (not a native fallback)', async () => {
                    await waitUtils.waitForCondition(
                        async () =>
                            (await personalCredentialsUtils.findRunnerByWorkloadName(
                                project.namespace,
                                negativeWorkload
                            )) !== undefined,
                        `WorkloadRunner CRD for ${negativeWorkload} (absence => silent native fallback, plugin never ran)`,
                        true,
                        60
                    );
                });

                let destSecretName: string;

                await test.step('The bogus creds are materialized as a WorkloadRunner-owned dockerconfigjson dest Secret on the pod', async () => {
                    await waitUtils.waitForCondition(
                        async () =>
                            (await kubeCli.listPodsOfJob(project.namespace, negativeWorkload, true)).length >= 1,
                        `pod of ${negativeWorkload} created`,
                        true,
                        120
                    );
                    const pod = (await kubeCli.listPodsOfJob(project.namespace, negativeWorkload, true))[0];
                    const dockerIps = (pod.spec?.imagePullSecrets ?? []).find((s) =>
                        s.name?.toLowerCase().includes(dockerCredName.toLowerCase())
                    );
                    expect(dockerIps, `pod imagePullSecret for ${dockerCredName}`).toBeDefined();

                    destSecretName = dockerIps!.name!;
                    const destSecret = await kubeCli.getSecret(destSecretName, project.namespace);
                    expect(destSecret.type).toBe('kubernetes.io/dockerconfigjson');
                    const dockerConfig = JSON.parse(
                        personalCredentialsUtils.decodeSecret(destSecret.data['.dockerconfigjson'])
                    ) as {
                        auths: Record<string, { username: string; password: string }>;
                    };
                    expect(dockerConfig.auths[HUB_KEY], `dockerconfigjson carries ${HUB_KEY}`).toBeDefined();
                    expect(dockerConfig.auths[HUB_KEY].username).toBe(BOGUS_USER);
                    expect(dockerConfig.auths[HUB_KEY].password).toBe(BOGUS_PASSWORD);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(destSecret.metadata?.ownerReferences),
                        'dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
                    ).toBe(true);
                });

                await test.step('The pull FAILS with an auth error: the runtime consulted the personal credential and the registry rejected it', async () => {
                    let lastWaiting: { reason?: string; message?: string } | undefined;
                    try {
                        await waitUtils.waitForCondition(
                            async () => {
                                const pod = (
                                    await kubeCli.listPodsOfJob(project.namespace, negativeWorkload, true)
                                )[0];
                                const waiting = (pod?.status?.containerStatuses ?? [])
                                    .map((cs) => cs.state?.waiting)
                                    .find((w) => PULL_FAILURE_REASON.test(w?.reason ?? ''));
                                if (waiting) {
                                    lastWaiting = waiting;
                                }
                                // ImagePullBackOff cycles carry a generic back-off message; the auth detail
                                // (401) surfaces on the recurring ErrImagePull attempts — keep polling for it.
                                return PULL_AUTH_FAILURE.test(waiting?.message ?? '');
                            },
                            `pod of ${negativeWorkload} stuck in ErrImagePull with a registry auth failure`,
                            true,
                            120,
                            5
                        );
                    } catch (error) {
                        logger.error(
                            `Last observed waiting state for ${negativeWorkload}: ${JSON.stringify(lastWaiting)}`
                        );
                        await personalCredentialsUtils.logSubmissionDiagnostics(
                            project.namespace,
                            negativeWorkload
                        );
                        throw error;
                    }

                    const pod = (await kubeCli.listPodsOfJob(project.namespace, negativeWorkload, true))[0];
                    expect(
                        pod.status?.phase,
                        'pod must be stuck Pending — the bogus credential blocked an otherwise anon-pullable image'
                    ).toBe('Pending');
                });

                await test.step('Delete the workload and confirm the dest Secret is garbage-collected', async () => {
                    await deleteWorkload(negativeWorkload);
                    await personalCredentialsUtils.waitForSecretGone(project.namespace, destSecretName);
                });
            });
        });
    }
);

// §2/§3/§4/§5 — four single generic-secret cases sharing one credential: GC-on-delete, no-cred no-op,
// idempotent reconcile, and two runners sharing one source (independent dests + per-runner GC).
test.describe(
    `Personal Credentials Plugin genericSecret lifecycle & reconcile, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const SRC_KEY = 'token';
            const SRC_VALUE = 'hf-personal-token-value';
            const ENV_NAME = 'HF_TOKEN';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericSecret: GenericSecret;
            let credName: string;
            let createdWorkloadNames: string[];

            const credEnvSpec = (): TrainingSpecSpec => ({
                image: IMAGE,
                command: COMMAND,
                compute: CPU_COMPUTE,
                environmentVariables: [{ name: ENV_NAME, userCredential: { name: credName, key: SRC_KEY } }]
            });

            const submitTraining = async (name: string, spec: TrainingSpecSpec): Promise<void> => {
                workloadsCrudUtils.accessToken = personalCredentialsUtils.systemAdminAccessToken();
                const { body: training } = await workloadsCrudUtils.createTraining(
                    name,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    spec
                );
                createdWorkloadNames.push(name);
                expect(training.name).toBe(name);
            };

            const deleteTraining = async (name: string): Promise<void> => {
                await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                    name,
                    project.id,
                    CrudWorkloadType.TRAINING
                );
                createdWorkloadNames = createdWorkloadNames.filter((n) => n !== name);
            };

            const destSecretNameFromPodEnv = async (workloadName: string): Promise<string> => {
                const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                const env = (pod.spec?.containers?.[0]?.env ?? []).find((e) => e.name === ENV_NAME);
                expect(env?.valueFrom?.secretKeyRef, `env ${ENV_NAME} secretKeyRef`).toBeDefined();
                expect(env!.valueFrom!.secretKeyRef!.key).toBe(SRC_KEY);
                return env!.valueFrom!.secretKeyRef!.name!;
            };

            const sourceSecretStillPresent = async (): Promise<boolean> => {
                const runaiSecrets = (await kubeCli.listSecretsByNamepsace('runai')).items;
                return runaiSecrets.some((s) => s.metadata?.name?.includes(credName.toLowerCase()));
            };

            test.beforeAll('enable v2 path, create project and one generic-secret credential', async () => {
                restoreAnyWorkload = await enableAnyWorkloadSubmission();
                try {
                    project = await projectResourcesUtils.createProjectAndNamespace('pc-life');
                    credName = objectGenerator.generateShortNameByPrefix('pc-life-gs', 4);
                    genericSecret = await credentialUtils.createGenericSecretAsset(
                        personalCredentialsUtils.genericSecretPayload(
                            credName,
                            [{ key: SRC_KEY, value: SRC_VALUE }],
                            'Personal generic secret for lifecycle/reconcile/GC cases'
                        )
                    );
                    await credentialUtils.waitForPersonalCredentialSecrets(
                        credName,
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

            test.beforeEach(() => {
                createdWorkloadNames = [];
            });

            test.afterEach('delete any workloads created by the test', async () => {
                for (const name of [...createdWorkloadNames]) {
                    try {
                        await deleteTraining(name);
                    } catch (e) {
                        logger.error(`Failed to delete workload ${name}`, e);
                    }
                }
            });

            test.afterAll('delete credential, project, restore v2 switch', async () => {
                try {
                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                } catch (e) {
                    logger.error('Failed to delete generic-secret credential', e);
                }
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (e) {
                    logger.error(`Failed to delete project ${project.name}`, e);
                }
                if (restoreAnyWorkload) {
                    await restoreAnyWorkload();
                }
            });

            test('T6141015: destination Secret is garbage-collected on workload deletion (no finalizer)', async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('pc-gc-wl', 4);
                let destName = '';

                await test.step('Submit a training referencing the generic secret', async () => {
                    await submitTraining(workloadName, credEnvSpec());
                });

                await test.step('Wait for the pod to run', async () => {
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadName, 120);
                });

                await test.step('Destination Secret exists, ownerRef → WorkloadRunner (controller + blockOwnerDeletion), decoded value', async () => {
                    destName = await destSecretNameFromPodEnv(workloadName);
                    const secret = await kubeCli.getSecret(destName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(secret.data[SRC_KEY])).toBe(SRC_VALUE);
                    const ownerRef = (secret.metadata?.ownerReferences ?? []).find(
                        (r) => r.kind === PC_WLR.kind
                    );
                    expect(ownerRef, 'dest Secret must be owned by a WorkloadRunner').toBeDefined();
                    expect(ownerRef!.controller).toBe(true);
                    expect(ownerRef!.blockOwnerDeletion).toBe(true);
                });

                await test.step('Delete the workload; dest Secret is GC’d (no finalizer), source untouched', async () => {
                    await deleteTraining(workloadName);
                    await personalCredentialsUtils.waitForSecretGone(project.namespace, destName);
                    expect(
                        await sourceSecretStillPresent(),
                        'source Secret in runai ns must remain after workload delete'
                    ).toBe(true);
                });
            });

            test('T6141011: workload without any personal credential → no annotation, plugin is a no-op', async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('pc-noop-wl', 4);
                let runnerName = '';

                await test.step('Submit a training with NO personal-credential reference', async () => {
                    await submitTraining(workloadName, {
                        image: IMAGE,
                        command: COMMAND,
                        compute: CPU_COMPUTE
                    });
                });

                await test.step('A WorkloadRunner is created for the workload', async () => {
                    await waitUtils.waitForCondition(
                        async () =>
                            (await personalCredentialsUtils.findRunnerByWorkloadName(
                                project.namespace,
                                workloadName
                            )) !== undefined,
                        `WorkloadRunner for ${workloadName}`,
                        true,
                        90
                    );
                    runnerName = (await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadName
                    ))!.metadata?.name as string;
                });

                await test.step('Runner carries NO run.ai/personal-credentials annotation', async () => {
                    const runner = await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadName
                    );
                    const annotations = (runner?.metadata?.annotations ?? {}) as Record<string, string>;
                    expect(annotations[PC_ANNOTATION]).toBeUndefined();
                });

                await test.step('No pc-* destination Secret was created for this workload', async () => {
                    const secrets = (await kubeCli.listSecretsByNamepsace(project.namespace)).items;
                    const pcSecrets = secrets.filter(
                        (s) => s.metadata?.name?.startsWith('pc-') && s.metadata.name.includes(workloadName)
                    );
                    expect(pcSecrets.map((s) => s.metadata?.name)).toEqual([]);
                });

                await test.step('personalcreds plugin reports Disabled with no issues (no-annotation no-op path)', async () => {
                    await waitUtils.waitForCondition(
                        async () =>
                            (
                                await personalCredentialsUtils.readPersonalCredsPluginStatus(
                                    project.namespace,
                                    runnerName
                                )
                            )?.phase === 'Disabled',
                        `personalcreds plugin reports Disabled (no annotation) for ${runnerName}`,
                        true,
                        60,
                        5
                    );
                    const status = await personalCredentialsUtils.readPersonalCredsPluginStatus(
                        project.namespace,
                        runnerName
                    );
                    expect(status?.issues ?? [], 'no-op path records no personalcreds issue').toEqual([]);
                });
            });

            test('T6141014: re-reconcile is idempotent (runner-owned destination Secret unchanged, no new issue)', async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('pc-idem-wl', 4);
                let destName = '';
                let runnerName = '';

                await test.step('Submit, wait running, resolve destination Secret + runner', async () => {
                    await submitTraining(workloadName, credEnvSpec());
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadName, 120);
                    destName = await destSecretNameFromPodEnv(workloadName);
                    runnerName = (await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadName
                    ))!.metadata?.name as string;
                });

                let beforeUid = '';
                let beforeValue = '';
                await test.step('Capture decoded token + Secret uid before reconcile', async () => {
                    const secret = await kubeCli.getSecret(destName, project.namespace);
                    beforeValue = personalCredentialsUtils.decodeSecret(secret.data[SRC_KEY]);
                    beforeUid = secret.metadata?.uid ?? '';
                    expect(beforeValue).toBe(SRC_VALUE);
                });

                await test.step('Force a reconcile; dest Secret stays the same object + value, plugin healthy (Ready, no issues)', async () => {
                    await personalCredentialsUtils.forceRunnerReconcile(project.namespace, runnerName);
                    // SSA re-apply rebuilds identical Data on the SAME object (uid unchanged), plugin stays Ready.
                    await waitUtils.waitForCondition(
                        async () => {
                            const secret = await kubeCli.getSecret(destName, project.namespace);
                            const status = await personalCredentialsUtils.readPersonalCredsPluginStatus(
                                project.namespace,
                                runnerName
                            );
                            return (
                                personalCredentialsUtils.decodeSecret(secret.data[SRC_KEY]) === beforeValue &&
                                (secret.metadata?.uid ?? '') === beforeUid &&
                                (status?.issues ?? []).length === 0 &&
                                status?.phase === 'Ready'
                            );
                        },
                        `dest Secret ${destName} unchanged (same uid + value) and personalcreds plugin Ready after reconcile`,
                        true,
                        30,
                        5
                    );
                });
            });

            test('T6141009: two WorkloadRunners share one source → independent destinations, independent ownerRefs, per-runner GC', async () => {
                const workloadA = objectGenerator.generateShortNameByPrefix('pc-sharea', 4);
                const workloadB = objectGenerator.generateShortNameByPrefix('pc-shareb', 4);
                let destA = '';
                let destB = '';

                await test.step('Submit two trainings both referencing the same generic secret', async () => {
                    await submitTraining(workloadA, credEnvSpec());
                    await submitTraining(workloadB, credEnvSpec());
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadA, 120);
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadB, 120);
                });

                await test.step('Each workload gets its OWN destination Secret, each owned only by its own runner', async () => {
                    destA = await destSecretNameFromPodEnv(workloadA);
                    destB = await destSecretNameFromPodEnv(workloadB);
                    expect(
                        destA,
                        'destName carries the runner name → cannot collide across runners'
                    ).not.toBe(destB);

                    const secretA = await kubeCli.getSecret(destA, project.namespace);
                    const secretB = await kubeCli.getSecret(destB, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(secretA.data[SRC_KEY])).toBe(SRC_VALUE);
                    expect(personalCredentialsUtils.decodeSecret(secretB.data[SRC_KEY])).toBe(SRC_VALUE);

                    const ownersA = secretA.metadata?.ownerReferences ?? [];
                    const ownersB = secretB.metadata?.ownerReferences ?? [];
                    expect(ownersA.length, 'dest A has exactly one ownerRef').toBe(1);
                    expect(ownersB.length, 'dest B has exactly one ownerRef').toBe(1);
                    const runnerA = await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadA
                    );
                    const runnerB = await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadB
                    );
                    expect(ownersA[0].kind).toBe(PC_WLR.kind);
                    expect(ownersB[0].kind).toBe(PC_WLR.kind);
                    expect(ownersA[0].name).toBe(runnerA!.metadata?.name);
                    expect(ownersB[0].name).toBe(runnerB!.metadata?.name);
                    expect(ownersA[0].name, 'no shared ownerRef across runners').not.toBe(ownersB[0].name);
                });

                await test.step('Delete only workload A → its dest is GC’d; workload B and its dest are unaffected', async () => {
                    await deleteTraining(workloadA);
                    await personalCredentialsUtils.waitForSecretGone(project.namespace, destA);
                    const secretB = await kubeCli.getSecret(destB, project.namespace);
                    expect(secretB.metadata?.name).toBe(destB);
                    const podsB = await kubeCli.listPodsOfJob(project.namespace, workloadB, true);
                    expect(podsB[0]?.status?.phase).toBe('Running');
                });
            });
        });
    }
);

// §5 P1 — delete the source credential while a workload runs: (A) sticky dest copy, no propagation;
// (B) missing-source webhook 422 on any runner update, pod survives; (C) new submit rejected; (D) dest still GC'd.
test.describe(
    `Personal Credentials Plugin delete-source-while-running, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const SRC_KEY = 'token';
            const SRC_VALUE = 'sticky-token-value';
            const ENV_NAME = 'HF_TOKEN';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericSecret: GenericSecret;
            let credName: string;
            let credDeleted: boolean;
            let workloadName: string;
            let workloadCreated: boolean;

            const credEnvSpec = (): TrainingSpecSpec => ({
                image: IMAGE,
                command: COMMAND,
                compute: CPU_COMPUTE,
                environmentVariables: [{ name: ENV_NAME, userCredential: { name: credName, key: SRC_KEY } }]
            });

            const sourceSecretPresent = async (): Promise<boolean> => {
                const runaiSecrets = (await kubeCli.listSecretsByNamepsace('runai')).items;
                return runaiSecrets.some((s) => s.metadata?.name?.includes(credName.toLowerCase()));
            };

            test.beforeAll('enable v2 path and create the project', async () => {
                restoreAnyWorkload = await enableAnyWorkloadSubmission();
                try {
                    project = await projectResourcesUtils.createProjectAndNamespace('pc-delsrc');
                } catch (error) {
                    if (restoreAnyWorkload) {
                        await restoreAnyWorkload();
                        restoreAnyWorkload = undefined;
                    }
                    throw error;
                }
            });

            // Credential created PER-TEST (not beforeAll) because this test deletes it mid-flight; a fresh
            // credential per attempt keeps the test retry-safe (beforeAll runs once).
            test.beforeEach(async () => {
                workloadName = objectGenerator.generateShortNameByPrefix('pc-sticky-wl', 4);
                workloadCreated = false;
                credDeleted = false;
                credName = objectGenerator.generateShortNameByPrefix('pc-delsrc-gs', 4);
                genericSecret = await credentialUtils.createGenericSecretAsset(
                    personalCredentialsUtils.genericSecretPayload(
                        credName,
                        [{ key: SRC_KEY, value: SRC_VALUE }],
                        'Personal generic secret to be deleted mid-flight'
                    )
                );
                await credentialUtils.waitForPersonalCredentialSecrets(
                    credName,
                    personalCredentialsUtils.systemAdminAccessToken()
                );
            });

            test.afterEach('delete workload + credential if still present', async () => {
                if (workloadCreated) {
                    try {
                        await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                            workloadName,
                            project.id,
                            CrudWorkloadType.TRAINING
                        );
                    } catch (e) {
                        logger.error(`Failed to delete workload ${workloadName}`, e);
                    }
                }
                if (!credDeleted) {
                    try {
                        await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                    } catch (e) {
                        logger.error('Failed to delete generic-secret credential', e);
                    }
                }
            });

            test.afterAll('delete project, restore v2 switch', async () => {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (e) {
                    logger.error(`Failed to delete project ${project.name}`, e);
                }
                if (restoreAnyWorkload) {
                    await restoreAnyWorkload();
                }
            });

            test('T6141016: deleting the source credential leaves a sticky copy, raises a missing-source issue, blocks new submits, and still GCs on delete', async () => {
                let destName = '';
                let runnerName = '';

                await test.step('Submit, run, and materialize the baseline', async () => {
                    workloadsCrudUtils.accessToken = personalCredentialsUtils.systemAdminAccessToken();
                    const { body: training } = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        credEnvSpec()
                    );
                    workloadCreated = true;
                    expect(training.name).toBe(workloadName);
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadName, 120);
                    const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                    const env = (pod.spec?.containers?.[0]?.env ?? []).find((e) => e.name === ENV_NAME);
                    expect(env?.valueFrom?.secretKeyRef).toBeDefined();
                    destName = env!.valueFrom!.secretKeyRef!.name!;
                    runnerName = (await personalCredentialsUtils.findRunnerByWorkloadName(
                        project.namespace,
                        workloadName
                    ))!.metadata?.name as string;
                    const secret = await kubeCli.getSecret(destName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(secret.data[SRC_KEY])).toBe(SRC_VALUE);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(secret.metadata?.ownerReferences)
                    ).toBe(true);
                });

                await test.step('Assert A — delete the credential; source Secret disappears, dest copy + pod stay (sticky, no propagation)', async () => {
                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                    credDeleted = true;
                    await waitUtils.waitForCondition(
                        async () => !(await sourceSecretPresent()),
                        `source Secret for ${credName} removed from runai ns after credential delete`,
                        true,
                        120,
                        5
                    );
                    const secret = await kubeCli.getSecret(destName, project.namespace);
                    expect(secret.metadata?.name).toBe(destName);
                    const owners = secret.metadata?.ownerReferences ?? [];
                    expect(owners.length).toBe(1);
                    expect(owners[0].kind).toBe(PC_WLR.kind);
                    const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                    expect(pod?.status?.phase).toBe('Running');
                });

                await test.step('Assert B — missing-source detection: the admission webhook rejects any runner update once the source is gone (pod survives)', async () => {
                    // The ValidatorPlugin re-validates the annotation on every runner update and rejects with a
                    // 422 naming the missing sourceSecret; the running pod is NOT torn down.
                    let caught: any;
                    try {
                        await personalCredentialsUtils.forceRunnerReconcile(project.namespace, runnerName);
                    } catch (e) {
                        caught = e;
                    }
                    expect(
                        caught,
                        'updating a runner whose source was deleted must be rejected by the validator webhook'
                    ).toBeDefined();
                    const errText = `${(caught as any)?.body ?? ''} ${(caught as any)?.message ?? ''} ${caught}`;
                    expect(errText, 'rejection is a 422 from the admission webhook').toMatch(/422|Invalid/);
                    expect(errText, 'rejection names the missing sourceSecret').toContain('sourceSecret');
                    const pod = (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0];
                    expect(pod?.status?.phase).toBe('Running');
                });

                await test.step('Assert C — a NEW submit referencing the deleted credential is rejected (CP 4xx, resolve-not-found)', async () => {
                    const ghostWorkload = objectGenerator.generateShortNameByPrefix('pc-sticky2-wl', 4);
                    let caught: any;
                    try {
                        await workloadsCrudUtils.createTraining(
                            ghostWorkload,
                            false,
                            flowRootState.cluster.uuid,
                            project.id.toString(),
                            credEnvSpec(),
                            personalCredentialsUtils.systemAdminAccessToken()
                        );
                    } catch (e) {
                        caught = e;
                    }
                    expect(
                        caught,
                        'submit referencing the deleted credential must be rejected'
                    ).toBeDefined();
                    // CP still resolves the just-deleted credential by name → rejection surfaces cluster-side
                    // (422). Tolerate 404 if asset-deletion has already propagated to CP resolve.
                    expect(
                        [StatusCodes.UNPROCESSABLE_ENTITY, StatusCodes.NOT_FOUND],
                        'deleted-credential submit rejected (cluster webhook 422, or CP 404 if propagated)'
                    ).toContain(personalCredentialsUtils.httpStatusOf(caught));
                    expect(
                        await personalCredentialsUtils.findRunnerByWorkloadName(
                            project.namespace,
                            ghostWorkload
                        )
                    ).toBeUndefined();
                });

                await test.step('Assert D — deleting the original runner still GCs the dest Secret', async () => {
                    await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                        workloadName,
                        project.id,
                        CrudWorkloadType.TRAINING
                    );
                    workloadCreated = false;
                    await personalCredentialsUtils.waitForSecretGone(
                        project.namespace,
                        destName,
                        `dest Secret ${destName} GC'd after runner delete (ownerRef intact regardless of source)`
                    );
                });
            });
        });
    }
);

// §3 P1 — a submission mixing valid + invalid credentials is rejected atomically; KILLER: even the valid
// creds do not materialize (all-or-nothing). A valid-only positive control proves the two valid creds resolve.
test.describe(
    `Personal Credentials Plugin multi-credential partial failure, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const HF_KEY = 'HF_TOKEN';
            const HF_VALUE = 'hf-valid-value';
            const NGC_KEY = 'NGC_API_KEY';
            const NGC_VALUE = 'ngc-valid-value';
            const GHOST_ENV = 'GHOST';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericSecret: GenericSecret;
            let ngcApiKey: NgcApiKey;
            let genericCredName: string;
            let ngcCredName: string;
            let ghostCredName: string;
            let workloadName: string;
            let workloadCreated: boolean;

            const pcSecretsFor = async (name: string): Promise<string[]> => {
                const secrets = (await kubeCli.listSecretsByNamepsace(project.namespace)).items;
                return secrets
                    .filter((s) => s.metadata?.name?.startsWith('pc-') && s.metadata.name.includes(name))
                    .map((s) => s.metadata!.name!);
            };

            test.beforeAll('enable v2 path, create project + two valid credentials', async () => {
                restoreAnyWorkload = await enableAnyWorkloadSubmission();
                try {
                    project = await projectResourcesUtils.createProjectAndNamespace('pc-partial');
                    genericCredName = objectGenerator.generateShortNameByPrefix('pc-partial-gs', 4);
                    ngcCredName = objectGenerator.generateShortNameByPrefix('pc-partial-ngc', 4);
                    ghostCredName = objectGenerator.generateShortNameByPrefix('pc-partial-ghost', 4); // never created
                    genericSecret = await credentialUtils.createGenericSecretAsset(
                        personalCredentialsUtils.genericSecretPayload(
                            genericCredName,
                            [{ key: HF_KEY, value: HF_VALUE }],
                            'valid generic secret'
                        )
                    );
                    ngcApiKey = await credentialUtils.createNgcApiKeyAsset(
                        personalCredentialsUtils.ngcApiKeyPayload(ngcCredName, NGC_VALUE, 'valid ngc-api-key')
                    );
                    await credentialUtils.waitForPersonalCredentialSecrets(
                        genericCredName,
                        personalCredentialsUtils.systemAdminAccessToken()
                    );
                    await credentialUtils.waitForPersonalCredentialSecrets(
                        ngcCredName,
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

            test.beforeEach(() => {
                workloadName = objectGenerator.generateShortNameByPrefix('pc-partial-wl', 4);
                workloadCreated = false;
            });

            test.afterEach('delete workload if created', async () => {
                if (!workloadCreated) {
                    return;
                }
                try {
                    await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                        workloadName,
                        project.id,
                        CrudWorkloadType.TRAINING
                    );
                } catch (e) {
                    logger.error(`Failed to delete workload ${workloadName}`, e);
                }
            });

            test.afterAll('delete credentials, project, restore v2 switch', async () => {
                await Promise.allSettled([
                    credentialUtils
                        .deleteGenericSecretAsset(genericSecret.meta.id)
                        .catch((e) => logger.error('Failed to delete generic-secret credential', e)),
                    credentialUtils
                        .deleteNgcApiKeyAsset(ngcApiKey.meta.id)
                        .catch((e) => logger.error('Failed to delete ngc-api-key credential', e))
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

            test('T6141013: a submission mixing valid + invalid credentials is rejected atomically — no partial materialization', async () => {
                const mixedSpec: TrainingSpecSpec = {
                    image: IMAGE,
                    command: COMMAND,
                    compute: CPU_COMPUTE,
                    environmentVariables: [
                        { name: HF_KEY, userCredential: { name: genericCredName, key: HF_KEY } },
                        { name: NGC_KEY, userCredential: { name: ngcCredName, key: NGC_KEY } },
                        { name: GHOST_ENV, userCredential: { name: ghostCredName, key: 'token' } }
                    ]
                };

                await test.step('Submit with one ghost credential among two valid → 404 at the asset layer', async () => {
                    let caught: any;
                    try {
                        await workloadsCrudUtils.createTraining(
                            workloadName,
                            false,
                            flowRootState.cluster.uuid,
                            project.id.toString(),
                            mixedSpec,
                            personalCredentialsUtils.systemAdminAccessToken()
                        );
                        workloadCreated = true;
                    } catch (e) {
                        caught = e;
                    }
                    expect(caught, 'mixed valid+invalid submission must be rejected').toBeDefined();
                    // The ghost credential fails to resolve, atomically rejecting the whole submission (404).
                    expect(
                        personalCredentialsUtils.httpStatusOf(caught),
                        'unresolved credential → 404 (not found)'
                    ).toBe(StatusCodes.NOT_FOUND);
                });

                await test.step('No workload reached the cluster (no runner)', async () => {
                    expect(
                        await personalCredentialsUtils.findRunnerByWorkloadName(
                            project.namespace,
                            workloadName
                        )
                    ).toBeUndefined();
                });

                await test.step('KILLER — neither VALID credential was materialized (all-or-nothing)', async () => {
                    expect(await pcSecretsFor(workloadName)).toEqual([]);
                });

                await test.step('Positive control — the same submission with only the two valid credentials succeeds and both materialize', async () => {
                    const validSpec: TrainingSpecSpec = {
                        image: IMAGE,
                        command: COMMAND,
                        compute: CPU_COMPUTE,
                        environmentVariables: [
                            { name: HF_KEY, userCredential: { name: genericCredName, key: HF_KEY } },
                            { name: NGC_KEY, userCredential: { name: ngcCredName, key: NGC_KEY } }
                        ]
                    };
                    workloadsCrudUtils.accessToken = personalCredentialsUtils.systemAdminAccessToken();
                    const { body: training } = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        validSpec
                    );
                    workloadCreated = true;
                    expect(training.name).toBe(workloadName);
                    await kubeCli.waitForJobPodsRunning(project.namespace, workloadName, 120);
                    const env =
                        (await kubeCli.listPodsOfJob(project.namespace, workloadName, true))[0].spec
                            ?.containers?.[0]?.env ?? [];
                    const hfDest = env.find((e) => e.name === HF_KEY)!.valueFrom!.secretKeyRef!.name!;
                    const ngcDest = env.find((e) => e.name === NGC_KEY)!.valueFrom!.secretKeyRef!.name!;
                    expect(hfDest).not.toBe(ngcDest);
                    const hfSecret = await kubeCli.getSecret(hfDest, project.namespace);
                    const ngcSecret = await kubeCli.getSecret(ngcDest, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(hfSecret.data[HF_KEY])).toBe(HF_VALUE);
                    expect(personalCredentialsUtils.decodeSecret(ngcSecret.data[NGC_KEY])).toBe(NGC_VALUE);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(hfSecret.metadata?.ownerReferences)
                    ).toBe(true);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(ngcSecret.metadata?.ownerReferences)
                    ).toBe(true);
                });
            });
        });
    }
);

// §3 P1 — the CP resolves only the caller's own private credentials: a non-owner with project access is
// rejected (404), the owner succeeds with the same reference; a no-cred researcher submit isolates the cause to ownership.
test.describe(
    `Personal Credentials Plugin non-owner rejection, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const SRC_KEY = 'token';
            const SRC_VALUE = 'owner-only-value';
            const ENV_NAME = 'HF_TOKEN';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericSecret: GenericSecret;
            let credName: string;
            let researcherRule: AccessRule | undefined;
            let createdWorkloadNames: string[];

            const credEnvSpec = (): TrainingSpecSpec => ({
                image: IMAGE,
                command: COMMAND,
                compute: CPU_COMPUTE,
                environmentVariables: [{ name: ENV_NAME, userCredential: { name: credName, key: SRC_KEY } }]
            });
            const noCredSpec = (): TrainingSpecSpec => ({
                image: IMAGE,
                command: COMMAND,
                compute: CPU_COMPUTE
            });

            test.beforeAll(
                'enable v2 path, create project + admin-owned private credential, grant researcher project access',
                async () => {
                    restoreAnyWorkload = await enableAnyWorkloadSubmission();
                    try {
                        project = await projectResourcesUtils.createProjectAndNamespace('pc-owner');
                        credName = objectGenerator.generateShortNameByPrefix('pc-owner-gs', 4);
                        genericSecret = await credentialUtils.createGenericSecretAsset(
                            personalCredentialsUtils.genericSecretPayload(
                                credName,
                                [{ key: SRC_KEY, value: SRC_VALUE }],
                                'private credential owned by the admin'
                            )
                        );
                        await credentialUtils.waitForPersonalCredentialSecrets(
                            credName,
                            personalCredentialsUtils.systemAdminAccessToken()
                        );
                        researcherRule = await authorizationUtils.createAccessRuleForUser(
                            personalCredentialsUtils.systemAdminAccessToken(),
                            flowRootState.users.l1_researcher.email,
                            flowRootState.users.l1_researcher.role.id,
                            SubjectType.User,
                            project.id.toString(),
                            ScopeType.Project
                        );
                    } catch (error) {
                        if (restoreAnyWorkload) {
                            await restoreAnyWorkload();
                            restoreAnyWorkload = undefined;
                        }
                        throw error;
                    }
                }
            );

            test.beforeEach(() => {
                createdWorkloadNames = [];
            });

            test.afterEach('delete any workloads created', async () => {
                for (const name of [...createdWorkloadNames]) {
                    try {
                        await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                            name,
                            project.id,
                            CrudWorkloadType.TRAINING
                        );
                    } catch (e) {
                        logger.error(`Failed to delete workload ${name}`, e);
                    }
                }
            });

            test.afterAll('revoke access rule, delete credential + project, restore v2 switch', async () => {
                if (researcherRule) {
                    try {
                        await authorizationUtils.deleteAccessRule(
                            researcherRule,
                            personalCredentialsUtils.systemAdminAccessToken()
                        );
                    } catch (e) {
                        logger.error('Failed to delete researcher access rule', e);
                    }
                }
                try {
                    await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id);
                } catch (e) {
                    logger.error('Failed to delete credential', e);
                }
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (e) {
                    logger.error(`Failed to delete project ${project.name}`, e);
                }
                if (restoreAnyWorkload) {
                    await restoreAnyWorkload();
                }
            });

            test('T6141012: the credential owner can use it; a non-owner with project access is rejected by the CP ownership filter', async () => {
                await test.step('Non-owner (researcher) referencing the admin-owned credential → CP 404', async () => {
                    const name = objectGenerator.generateShortNameByPrefix('pc-bob-wl', 4);
                    let caught: any;
                    try {
                        await workloadsCrudUtils.createTraining(
                            name,
                            false,
                            flowRootState.cluster.uuid,
                            project.id.toString(),
                            credEnvSpec(),
                            flowRootState.users.l1_researcher.accessToken
                        );
                        createdWorkloadNames.push(name);
                    } catch (e) {
                        caught = e;
                    }
                    expect(caught, 'non-owner submission must be rejected').toBeDefined();
                    // The asset layer resolves only the caller's own private credentials → 404 before composition.
                    expect(
                        personalCredentialsUtils.httpStatusOf(caught),
                        'non-owner credential reference → 404 (not found)'
                    ).toBe(StatusCodes.NOT_FOUND);
                    expect(
                        await personalCredentialsUtils.findRunnerByWorkloadName(project.namespace, name)
                    ).toBeUndefined();
                });

                await test.step('Baseline — researcher CAN submit WITHOUT the credential (proves project access; isolates the cause to ownership)', async () => {
                    const name = objectGenerator.generateShortNameByPrefix('pc-bobok-wl', 4);
                    const { body: training } = await workloadsCrudUtils.createTraining(
                        name,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        noCredSpec(),
                        flowRootState.users.l1_researcher.accessToken
                    );
                    createdWorkloadNames.push(name);
                    expect(training.name).toBe(name);
                });

                await test.step('Owner (admin) referencing the SAME credential by name → succeeds + materializes', async () => {
                    const name = objectGenerator.generateShortNameByPrefix('pc-alice-wl', 4);
                    const { body: training } = await workloadsCrudUtils.createTraining(
                        name,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        credEnvSpec(),
                        personalCredentialsUtils.systemAdminAccessToken()
                    );
                    createdWorkloadNames.push(name);
                    expect(training.name).toBe(name);
                    await kubeCli.waitForJobPodsRunning(project.namespace, name, 120);
                    const env = (
                        await kubeCli.listPodsOfJob(project.namespace, name, true)
                    )[0].spec?.containers?.[0]?.env?.find((e) => e.name === ENV_NAME);
                    const destName = env!.valueFrom!.secretKeyRef!.name!;
                    const secret = await kubeCli.getSecret(destName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(secret.data[SRC_KEY])).toBe(SRC_VALUE);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(secret.metadata?.ownerReferences)
                    ).toBe(true);
                });
            });
        });
    }
);
