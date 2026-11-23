/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import { expect, test } from '@playwright/test';
import {
    apiUtils,
    authUtils,
    BaseWorkload,
    config,
    crds,
    flowRootState,
    settings,
    tags,
    usersManager,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { projectsDepartmentsApi, workloadsService } from '../../../../../../modules/api';
import { SoftAssert } from '../../../../../../modules/helpers/SoftAssert';
import {
    shouldRunNativeWorkloadCrdTests,
    shouldRunTests
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    expectedV2SeedKinds,
    podsId,
    podsYaml,
    runningWlTestName,
    v2WorkloadsKey,
    V2WorkloadSeed,
    verifyV2WorkloadSeedHealthy,
    wlTestPhases,
    workloadCategory,
    workloadGen
} from './workloadUpgradeCommon';
import { resetNodePoolCrdDetection } from '../../../../../../modules/helpers/utils/NodePoolCrdFields';

const UPGRADE_ARTIFACT_PATH = '/tmp/upgradeArtifact.txt';

shouldRunTests(settings.isUpgradesTests)(() => {
    test.describe(`workload post upgrade validation ${tags.runOn.workloadPostUpgrade} ${tags.teams.viper}`, () => {
        let upgradeArtifact: Record<string, any>;
        let project: any;

        function createWorkloadYamlMessage(preworkload: BaseWorkload, postworkloadYaml: object): string {
            let message = '___________________________\n';
            message += `workload generation is not consistent for ${preworkload.name}\n`;
            message += '__________Pre_________________\n';
            message += `Pre workload yaml\n ________\n${JSON.stringify(preworkload['yaml'], null, 2)}\n`;
            message += '__________Post_________________\n';
            message += `Post workload yaml\n ________\n${JSON.stringify(postworkloadYaml, null, 2)}\n`;
            message += '___________________________\n';
            return message;
        }

        async function updateWorkloadControllerAdditionalLabelsAndAnnotations(): Promise<void> {
            // Pre-existing helper used a `for...in` loop on the conditions array,
            // which iterates indices (strings) rather than the condition objects.
            // Use `for...of` so the type/status comparisons actually work.
            const getAvailableCond = (runaiConfig: any): any => {
                const conditions = runaiConfig?.['status']?.['conditions'] ?? [];
                for (const condition of conditions) {
                    if (condition?.['type'] === 'Available') {
                        return condition;
                    }
                }
                return undefined;
            };

            let runaiConfig = await kubeCli.getCrd(crds.runaiConfigCrd, 'runai', 'runai');
            const generation = runaiConfig['metadata']['generation'];

            const patch = {
                spec: {
                    'workload-controller': {
                        additionalPodLabels: {
                            label: 'value'
                        },
                        additionalPodAnnotations: {
                            annotation: 'value'
                        }
                    }
                }
            };

            runaiConfig = await kubeCli.patchCrd(crds.runaiConfigCrd, patch, 'runai', 'runai');

            // if no changes were made, we can keep going
            if (runaiConfig['metadata']['generation'] === generation) {
                return;
            }

            await waitUtils.waitForCondition(
                () => {
                    const availableCond = getAvailableCond(runaiConfig);
                    if (!availableCond) {
                        return false;
                    }
                    return (
                        availableCond['status'] === 'True' && availableCond['observedGeneration'] > generation
                    );
                },
                'wait for configurations to be updated',
                true
            );
        }

        test.beforeAll(async () => {
            const adminTokens = await authUtils.authenticateKeycloakAndGetTokenOld(
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword
            );
            flowRootState.cluster = await apiUtils.getClusterByPrefix(
                config.clusterNamePrefix,
                adminTokens.idToken
            );
            await usersManager.initLocalUsers(adminTokens.accessToken);
            await kubeCli.initKubeContext(config.kubeconfigClusterName, false);
            kubeCli.useAdminContext();
            await waitUtils.waitForCondition(
                async () => (await kubeCli.getNumOfGpus()) > 0,
                'wait for gpus to be available in the cluster',
                true,
                300
            );
            flowRootState.nodesResources = await kubeCli.getNodesResources();
            flowRootState.numOfGpus = flowRootState.nodesResources.reduce((sum, node) => sum + node.gpu, 0);
            crds.initCrds(flowRootState.cluster.version);
            resetNodePoolCrdDetection();

            await updateWorkloadControllerAdditionalLabelsAndAnnotations();
            const fileContent = fs.readFileSync(UPGRADE_ARTIFACT_PATH, 'utf8');
            upgradeArtifact = JSON.parse(fileContent);
            project = upgradeArtifact['project'];
        });

        shouldRunNativeWorkloadCrdTests(() =>
        test('validating LastReconcileSucceeded condition', async () => {
            const softAssert = new SoftAssert();
            const workloads = upgradeArtifact[workloadCategory];
            for (const wlPhase in workloads) {
                for (const workload of workloads[wlPhase]) {
                    try {
                        const namespace = workload.namespace
                            ? workload.namespace
                            : workload.testingSpecificFields.namespace;

                        const workloadCrd = await kubeCli.getCrd(workload.crd, workload.name, namespace);
                        const reconcileStatus = workloadCrd['status']['conditions'].filter(
                            (x: any) => x.type === 'LastReconcileSucceeded' && x.status === 'True'
                        );
                        expect(reconcileStatus.length).toBeGreaterThan(0);
                    } catch (error) {
                        softAssert.assert(
                            false,
                            `unable to get LastReconcileSucceeded condition from ${workload.name} \n${error}`
                        );
                    }
                }
                softAssert.throwErrorIfFailed();
            }
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test('validating workload generation consistency', async () => {
            const softAssert = new SoftAssert();
            const workloads = upgradeArtifact[workloadCategory];

            for (const phase of wlTestPhases) {
                for (const workload of workloads[phase.name]) {
                    const postUpWlYaml = await kubeCli.getWorkload(
                        workload.name,
                        project.namespace,
                        workload['crd']
                    );
                    const isGenConsist = postUpWlYaml.metadata.generation === workload[workloadGen];
                    let message = '';
                    if (!isGenConsist) {
                        message = createWorkloadYamlMessage(workload, postUpWlYaml);
                    }
                    softAssert.assert(isGenConsist, message);
                }
            }
            softAssert.throwErrorIfFailed();
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test('validating workload pods consistency', async () => {
            const softAssert = new SoftAssert();
            const workloads = upgradeArtifact[workloadCategory];
            for (const workload of workloads[runningWlTestName]) {
                try {
                    const workloadPods = await workloadsUtilsV2.getWorkloadPods(workload['id']);
                    const postWorkloadPodsIds = workloadPods.body.pods.map((obj: any) => obj.id);
                    const preWorkloadPodsIds = workload[podsId];
                    const isPodsIdSame =
                        JSON.stringify(preWorkloadPodsIds.sort()) ===
                        JSON.stringify(postWorkloadPodsIds.sort());
                    let message = '';
                    if (!isPodsIdSame) {
                        const postUpWlYaml = await kubeCli.getWorkload(
                            workload.name,
                            project.namespace,
                            workload['crd']
                        );
                        message = '___________________________\n';
                        message += `workload pods id are not consistent for ${workload.name}\n`;
                        message += '__________Pre_________________\n';
                        message += `Pre pods Yaml\n ________\n${workload[podsYaml]}\n`;
                        message += '__________Post_________________\n';
                        message += 'Post pods Yaml\n ________\n';
                        message += createWorkloadYamlMessage(workload, postUpWlYaml);
                        for (const pod in workloadPods) {
                            const podYaml = await kubeCli.getPodYaml(
                                workloadPods[pod].name,
                                project.namespace
                            );
                            message += podYaml;
                        }
                        message += '___________________________\n';
                    }
                    softAssert.assert(isPodsIdSame, message);
                } catch (e) {
                    softAssert.assert(
                        false,
                        `validating workload consistency failed for ${runningWlTestName} job type ${workload.jobType}\n exception: ${e}`
                    );
                }
            }

            softAssert.throwErrorIfFailed();
            })
        );

        test.skip('validating workload controller working properly', async () => {
            const wlControllerPodName = (await kubeCli.listPodsByNamepsace('runai')).filter((x: any) =>
                x.metadata.name.startsWith('workload-controller')
            )[0].metadata.name;
            const isWlControllerRunning = await kubeCli.checkPodRunningTime(wlControllerPodName);
            expect(isWlControllerRunning).toBe(true);
        });

        shouldRunNativeWorkloadCrdTests(() => {
        wlTestPhases.forEach((params) => {
            test(`validating ${params.name} workloads stayed in the same state`, async () => {
                const softAssert = new SoftAssert();
                const adminTokens = await authUtils.authenticateUserOld(
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword
                );
                const wlDict = upgradeArtifact[workloadCategory][params.name];
                projectsDepartmentsApi.projectsApi.accessToken = adminTokens.idToken;
                workloadsService.workloadsApi.accessToken = adminTokens.idToken;
                for (const wl in wlDict) {
                    try {
                        await waitUtils.waitForBackendWorkloadPhase(
                            wlDict[wl].name,
                            upgradeArtifact['project'].id,
                            params.phase,
                            5
                        );
                    } catch (e) {
                        softAssert.assert(false, `${e}`);
                    }
                }
                softAssert.throwErrorIfFailed();
            });
        });
        });

        // ---------------- v2 workloads post-upgrade ----------------
        // Uses the same verifyV2WorkloadSeedHealthy helper the pre-upgrade test
        // calls as a baseline check, so "healthy" is defined in exactly one place.
        // Skip on the SAME gate the pre-upgrade tests use (operator + version),
        // not on an empty seed list - skipping on seeds.length === 0 would hide a
        // real bug where the pre-upgrade ran but its seed went missing.
        test('validating v2 workloads survived the upgrade', async () => {
            const expectedKinds = await expectedV2SeedKinds();
            if (expectedKinds.length === 0) {
                test.skip();
            }
            const seeds = (upgradeArtifact[v2WorkloadsKey] ?? []) as V2WorkloadSeed[];
            // every kind the pre-upgrade was expected to seed must be present
            for (const kind of expectedKinds) {
                expect(
                    seeds.some((s) => s.kind === kind),
                    `expected a ${kind} v2 seed from the pre-upgrade run but none was found in the artifact`
                ).toBe(true);
            }
            const softAssert = new SoftAssert();
            for (const seed of seeds) {
                const hydrated: V2WorkloadSeed = {
                    ...seed,
                    projectId: seed.projectId ?? String(project.id)
                };
                try {
                    await verifyV2WorkloadSeedHealthy(hydrated, softAssert);
                } catch (e) {
                    softAssert.assert(false, `v2 seed ${seed.kind} (${seed.name}) verification threw: ${e}`);
                }
            }
            softAssert.throwErrorIfFailed();
        });

        test('validating v2 workloads inner manifest preserved', async () => {
            const expectedKinds = await expectedV2SeedKinds();
            if (expectedKinds.length === 0) {
                test.skip();
            }
            const seeds = (upgradeArtifact[v2WorkloadsKey] ?? []) as V2WorkloadSeed[];
            for (const kind of expectedKinds) {
                expect(
                    seeds.some((s) => s.kind === kind),
                    `expected a ${kind} v2 seed from the pre-upgrade run but none was found in the artifact`
                ).toBe(true);
            }
            const softAssert = new SoftAssert();
            for (const seed of seeds) {
                try {
                    const wrapperName = seed.externalWorkloadWrapperName ?? seed.name;
                    const live = (await kubeCli.getCrd(crds[seed.crdKey], wrapperName, seed.namespace)) as {
                        spec?: any;
                    };

                    if (seed.submissionPath === 'workloads-v2-api') {
                        const tpl = live.spec?.template ?? {};
                        const innerRef = seed.innerManifest as {
                            apiVersion: string;
                            kind: string;
                            metadata: { name: string };
                        };
                        softAssert.assertNonStrictEqual(
                            tpl.kind,
                            innerRef.kind,
                            `${seed.kind} (${seed.name}) WorkloadRunner template.kind drift`
                        );
                        softAssert.assertNonStrictEqual(
                            tpl.apiVersion,
                            innerRef.apiVersion,
                            `${seed.kind} (${seed.name}) WorkloadRunner template.apiVersion drift`
                        );
                        softAssert.assertNonStrictEqual(
                            tpl.metadata?.name,
                            innerRef.metadata.name,
                            `${seed.kind} (${seed.name}) WorkloadRunner template.metadata.name drift`
                        );
                    } else {
                        // ExternalWorkload.spec.type is set by EWI to the
                        // original CRD kind it picked up (e.g. 'PyTorchJob').
                        // If the upgrade drops/changes it, EWI would
                        // re-classify or stop reconciling.
                        const expectedType = 'PyTorchJob';
                        softAssert.assertNonStrictEqual(
                            live.spec?.type,
                            expectedType,
                            `${seed.kind} (${seed.name}) ExternalWorkload type drift`
                        );
                    }
                } catch (err) {
                    softAssert.assert(
                        false,
                        `v2 seed ${seed.kind} (${seed.name}) inner-manifest check threw: ${err}`
                    );
                }
            }
            softAssert.throwErrorIfFailed();
        });
    });
});
