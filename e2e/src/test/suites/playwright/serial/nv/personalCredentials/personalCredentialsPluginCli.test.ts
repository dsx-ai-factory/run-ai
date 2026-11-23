/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Personal Credentials Plugin — CLI trigger surface (runai CLI, PyTorchJob).
 *
 * §1 P1 — same three-credential + multi-key contract as the API case, driven via the runai CLI
 * (my-credential create → login → training pytorch submit → delete) on a PyTorchJob (Master + Worker)
 * to prove the plugin injects credentials into BOTH nested replica containers; also pins the
 * personalcreds SSA field-manager. ANY_WORKLOAD_SUBMISSION is ON so the native submission routes
 * through the v2 WorkloadRunner + plugin (dest-Secret ownerRef is the discriminator).
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
    waitUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import {
    cliV2TrainingType,
    cliV2WorkloadType,
    DeleteFlagsMerged,
    MyCredentialType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    enableAnyWorkloadSubmission,
    RestoreAnyWorkload
} from '../../regression/viper/anyWorkloadSubmission';
test.describe.skip(
    `Personal Credentials Plugin multi-credential via CLI, tags: ${tags.tool.runaiCli} ${tags.resource.credential} ${tags.teams.loong}`,
    () => {
        supportedFrom(SupportedVersions.V2_26)(() => {
            const HF_KEY = 'HF_TOKEN';
            const HF_VALUE = 'hf-personal-token-value';
            const WANDB_KEY = 'WANDB_API_KEY';
            const WANDB_VALUE = 'wandb-personal-key-value';
            const NGC_KEY = 'NGC_API_KEY';
            const NGC_VALUE = 'ngc-personal-key-value';
            // Username must not contain '$' — `runai my-credential create` items are interpolated into a shell
            // command, so the API case's literal '$oauthtoken' would expand to empty. Asserts wiring, not a real pull.
            const DOCKER_USER = 'oauthtoken';
            const DOCKER_PASSWORD = 'dummy-ngc-pull-key';
            const DOCKER_URL = 'https://nvcr.io';

            // 1 worker yields a Master + a Worker replica — enough to prove injection into BOTH nested containers.
            const IMAGE = config.quickStartImage.distributedPytorch;
            const COMMAND = '-- sleep infinity';
            const WORKERS = '1';

            let restoreAnyWorkload: RestoreAnyWorkload | undefined;
            let project: Project;
            let genericCredName: string;
            let ngcCredName: string;
            let dockerCredName: string;
            let workloadName: string;
            let workloadCreated: boolean;

            // CLI authenticates as the v1 system_administrator (loginAsUser below); read that same identity's
            // token so waitForPersonalCredentialSecrets resolves the credentials this user just created.
            const idToken = (): string => flowRootState.users.system_administrator.idToken;

            const deleteWorkloadViaCli = (): void => {
                runaiCliV2.delete(
                    { [DeleteFlagsMerged.project]: project.name } as Partial<typeof DeleteFlagsMerged>,
                    workloadName,
                    cliV2TrainingType.pyTorch
                );
            };

            test.beforeAll(
                'enable v2 path, log the CLI in, create project and the three personal credentials via CLI',
                async () => {
                    restoreAnyWorkload = await enableAnyWorkloadSubmission();
                    try {
                        const sysAdmin = flowRootState.users.system_administrator;
                        runaiCliV2.loginAsUser(sysAdmin.email, sysAdmin.password);

                        project = await projectResourcesUtils.createProjectAndNamespace('pc-cli-multi');
                        runaiCliV2.projectSet(project.name);

                        genericCredName = objectGenerator.generateShortNameByPrefix('pc-cli-gs', 4);
                        ngcCredName = objectGenerator.generateShortNameByPrefix('pc-cli-ngc', 4);
                        dockerCredName = objectGenerator.generateShortNameByPrefix('pc-cli-dr', 4);

                        // genericSecret with TWO keys (multi-key): one --item per key.
                        expect(
                            runaiCliV2.myCredentialCreate(genericCredName, MyCredentialType.genericSecret, [
                                `key=${HF_KEY},value=${HF_VALUE}`,
                                `key=${WANDB_KEY},value=${WANDB_VALUE}`
                            ])
                        ).toContain(`Credential "${genericCredName}" created.`);
                        expect(
                            runaiCliV2.myCredentialCreate(ngcCredName, MyCredentialType.ngcApiKey, [
                                `ngc-api-key=${NGC_VALUE}`
                            ])
                        ).toContain(`Credential "${ngcCredName}" created.`);
                        expect(
                            runaiCliV2.myCredentialCreate(dockerCredName, MyCredentialType.dockerRegistry, [
                                `user=${DOCKER_USER},password=${DOCKER_PASSWORD},url=${DOCKER_URL}`
                            ])
                        ).toContain(`Credential "${dockerCredName}" created.`);

                        // Wait for the agent sync loop to materialize each encrypted source Secret.
                        await credentialUtils.waitForPersonalCredentialSecrets(genericCredName, idToken());
                        await credentialUtils.waitForPersonalCredentialSecrets(ngcCredName, idToken());
                        await credentialUtils.waitForPersonalCredentialSecrets(dockerCredName, idToken());
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
                workloadName = objectGenerator.generateShortNameByPrefix('pc-cli-wl', 4);
                workloadCreated = false;
            });

            test.afterEach('delete workload if still present', () => {
                if (!workloadCreated) {
                    return;
                }
                try {
                    deleteWorkloadViaCli();
                } catch (e) {
                    logger.error(`Failed to delete workload ${workloadName}`, e);
                }
            });

            test.afterAll('delete personal credentials, project, and restore v2 switch', async () => {
                const credsToDelete: [string, MyCredentialType][] = [
                    [genericCredName, MyCredentialType.genericSecret],
                    [ngcCredName, MyCredentialType.ngcApiKey],
                    [dockerCredName, MyCredentialType.dockerRegistry]
                ];
                for (const [name, type] of credsToDelete) {
                    try {
                        runaiCliV2.myCredentialDelete(name, type, false);
                    } catch (e) {
                        logger.error(`Failed to delete personal credential ${name}`, e);
                    }
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

            test('T6141005: submit one PyTorchJob referencing genericSecret(multi-key) + ngcApiKey + dockerRegistry via the runai CLI', async () => {
                // --env-my-credentials: type=...,name=<ENV>,credential-name=<cred>[,key=<KEY>]
                //   key is required for genericSecret; ngcApiKey auto-targets NGC_API_KEY.
                const envMyCredentials = [
                    `type=${MyCredentialType.genericSecret},name=${HF_KEY},credential-name=${genericCredName},key=${HF_KEY}`,
                    `type=${MyCredentialType.genericSecret},name=${WANDB_KEY},credential-name=${genericCredName},key=${WANDB_KEY}`,
                    `type=${MyCredentialType.ngcApiKey},name=${NGC_KEY},credential-name=${ngcCredName}`
                ];
                const imagePullMyCredentials = [
                    `type=${MyCredentialType.dockerRegistry},name=${dockerCredName}`
                ];

                await test.step('Submit the multi-credential PyTorchJob via the runai CLI', () => {
                    const stdout = runaiCliV2.submit(
                        {
                            [SubmitFlagsMerged.project]: project.name,
                            [SubmitFlagsMerged.image]: IMAGE,
                            [SubmitFlagsMerged.workers]: WORKERS,
                            [SubmitFlagsMerged.envMyCredentials]: envMyCredentials,
                            [SubmitFlagsMerged.imagePullMyCredentials]: imagePullMyCredentials,
                            [SubmitFlagsMerged.command]: COMMAND
                        } as Partial<typeof SubmitFlagsMerged>,
                        workloadName,
                        cliV2WorkloadType.training,
                        cliV2TrainingType.pyTorch
                    );
                    workloadCreated = true;
                    expect(stdout, 'CLI submit output should reference the workload').toContain(workloadName);
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

                // PyTorch yields a Master + Worker(s). Assert the plugin injected into EVERY replica container.
                await test.step('Wait for the Master + Worker replica pods to be created', async () => {
                    await waitUtils.waitForCondition(
                        async () =>
                            (await kubeCli.listPodsOfJob(project.namespace, workloadName, true)).length >= 2,
                        `Master + Worker pods for ${workloadName} (PyTorch distributed -> >= 2 replica pods)`,
                        true,
                        300
                    );
                    const pods = await kubeCli.listPodsOfJob(project.namespace, workloadName, true);
                    expect(
                        pods.filter((p) => p.metadata?.name?.includes('master')).length,
                        'expected a PyTorch Master replica pod'
                    ).toBeGreaterThanOrEqual(1);
                    expect(
                        pods.filter((p) => p.metadata?.name?.includes('worker')).length,
                        'expected a PyTorch Worker replica pod'
                    ).toBeGreaterThanOrEqual(1);
                });

                let genericDestSecretName = '';
                let ngcDestSecretName = '';
                let dockerDestSecretName = '';

                await test.step('Generic multi-key: both env vars wired on every replica -> one dest Secret, decoded values, owned by WorkloadRunner', async () => {
                    const pods = await kubeCli.listPodsOfJob(project.namespace, workloadName, true);
                    for (const pod of pods) {
                        const podEnv = pod.spec?.containers?.[0]?.env ?? [];
                        const hfEnv = podEnv.find((e) => e.name === HF_KEY);
                        const wandbEnv = podEnv.find((e) => e.name === WANDB_KEY);
                        expect(
                            hfEnv?.valueFrom?.secretKeyRef,
                            `${pod.metadata?.name}: env ${HF_KEY} secretKeyRef`
                        ).toBeDefined();
                        expect(
                            wandbEnv?.valueFrom?.secretKeyRef,
                            `${pod.metadata?.name}: env ${WANDB_KEY} secretKeyRef`
                        ).toBeDefined();
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
                    }

                    const genSecret = await kubeCli.getSecret(genericDestSecretName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(genSecret.data[HF_KEY])).toBe(HF_VALUE);
                    expect(personalCredentialsUtils.decodeSecret(genSecret.data[WANDB_KEY])).toBe(
                        WANDB_VALUE
                    );
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(
                            genSecret.metadata?.ownerReferences,
                            true
                        ),
                        'generic dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
                    ).toBe(true);
                });

                await test.step('ngcApiKey env wired on every replica -> own dest Secret, decoded value, owned by WorkloadRunner', async () => {
                    const pods = await kubeCli.listPodsOfJob(project.namespace, workloadName, true);
                    for (const pod of pods) {
                        const podEnv = pod.spec?.containers?.[0]?.env ?? [];
                        const ngcEnv = podEnv.find((e) => e.name === NGC_KEY);
                        expect(
                            ngcEnv?.valueFrom?.secretKeyRef,
                            `${pod.metadata?.name}: env ${NGC_KEY} secretKeyRef`
                        ).toBeDefined();
                        expect(ngcEnv!.valueFrom!.secretKeyRef!.key).toBe(NGC_KEY);
                        expect(ngcEnv!.valueFrom!.secretKeyRef!.name!.toLowerCase()).toContain(
                            ngcCredName.toLowerCase()
                        );
                        ngcDestSecretName = ngcEnv!.valueFrom!.secretKeyRef!.name!;
                    }
                    expect(ngcDestSecretName).not.toBe(genericDestSecretName);
                    const ngcSecret = await kubeCli.getSecret(ngcDestSecretName, project.namespace);
                    expect(personalCredentialsUtils.decodeSecret(ngcSecret.data[NGC_KEY])).toBe(NGC_VALUE);
                    expect(
                        personalCredentialsUtils.isWorkloadRunnerOwned(
                            ngcSecret.metadata?.ownerReferences,
                            true
                        ),
                        'ngc dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
                    ).toBe(true);
                });

                await test.step('dockerRegistry wired as a dockerconfigjson imagePullSecret on every replica, owned by WorkloadRunner', async () => {
                    const pods = await kubeCli.listPodsOfJob(project.namespace, workloadName, true);
                    for (const pod of pods) {
                        const imagePullSecrets = pod.spec?.imagePullSecrets ?? [];
                        const dockerIps = imagePullSecrets.find((s) =>
                            s.name?.toLowerCase().includes(dockerCredName.toLowerCase())
                        );
                        expect(
                            dockerIps,
                            `${pod.metadata?.name}: imagePullSecret for ${dockerCredName}; actual: ${imagePullSecrets
                                .map((s) => s.name)
                                .join(', ')}`
                        ).toBeDefined();
                        dockerDestSecretName = dockerIps!.name!;
                    }

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
                            dockerSecret.metadata?.ownerReferences,
                            true
                        ),
                        'docker dest Secret must be owned by WorkloadRunner (plugin path, not native fallback)'
                    ).toBe(true);
                });

                await test.step('SSA field-manager pin: the generic dest Secret was server-side-applied by the personalcreds plugin', async () => {
                    const genSecret = await kubeCli.getSecret(genericDestSecretName, project.namespace);
                    const managers = (genSecret.metadata?.managedFields ?? []).map((f) => f.manager ?? '');
                    expect(
                        managers.some((m) => m.toLowerCase().includes('personalcreds')),
                        `dest Secret should carry a personalcreds-plugin field-manager; actual managers: ${managers.join(
                            ', '
                        )}`
                    ).toBe(true);
                });

                await test.step('Delete via the runai CLI and confirm all three dest Secrets are garbage-collected', async () => {
                    deleteWorkloadViaCli();
                    workloadCreated = false;

                    for (const destName of [genericDestSecretName, ngcDestSecretName, dockerDestSecretName]) {
                        await personalCredentialsUtils.waitForSecretGone(
                            project.namespace,
                            destName,
                            `dest Secret ${destName} garbage-collected after CLI workload delete`
                        );
                    }

                    // Sources untouched: the plugin GCs only the dest copies; the encrypted source Secrets remain.
                    const runaiSecrets = (await kubeCli.listSecretsByNamepsace('runai')).items;
                    for (const credName of [genericCredName, ngcCredName, dockerCredName]) {
                        expect(
                            runaiSecrets.some((s) => s.metadata?.name?.includes(credName)),
                            `source Secret for ${credName} must remain in the runai namespace after workload delete`
                        ).toBe(true);
                    }
                });
            });
        });
    }
);
