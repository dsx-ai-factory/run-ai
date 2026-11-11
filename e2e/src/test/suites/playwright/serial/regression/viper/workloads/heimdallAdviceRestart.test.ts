/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    distributedUtils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { kubeCli, bashCli, runaiCliV2 } from '../../../../../../../modules/cli';
import { crds } from '../../../../../../../modules/helpers/models/customResourceDefinition';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { DistributedFramework } from '../../../../../../../modules/api/apis/policyApi';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { flowRootState } from '../../../../../../../modules/helpers/config/config';
import { cliLoginAsSystemAdmin } from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    REQUEST_REF_LABEL,
    WORKFLOW_ID_LABEL,
    CREATED_BY_LABEL,
    WORKLOAD_NAME_LABEL,
    WORKLOAD_ID_LABEL,
    WORKLOAD_ID_ANNOTATION,
    NODE_AFFINITY_OVERRIDE_PATH,
    buildAdviceBody,
    createPytorchWorkload,
    ensureDistributedWlrSubmission,
    restoreDistributedWlrSubmission,
    ensureWorkload,
    skipIfFewerThanTwoSchedulableNodes,
    buildRestartRequestBody,
    childReplicaExclusions,
    deleteAllCrInstancesInNamespace,
    expectKubectlDenied,
    expectRestartRequestCountAfterSettle,
    getPytorchJob,
    getRestartRequestsForAdvice,
    getWlr,
    getWlrEvents,
    runRestartCycle,
    waitForCpVisibleAdviceOperand,
    waitForNodesOffExcluded,
    waitForScheduledNodes,
    waitForWlrAndChildBack,
    wlrInstructionExclusions,
    allHostnameExclusions,
    childReplicaTerms,
    getSchedulableNodeNames,
    clusterSupportsWlrRestart,
    waitForDeploymentRollout
} from '../../../../../../../modules/helpers/utils/wlrRestartUtils';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { fileURLToPath } from 'url';

test.describe(
    `Heimdall advice restart - WorkloadRunner ${tags.teams.viper}`,
    // intentionally NO runOn/component tags: the label-driven CI selection regex
    // requires BOTH a PR-label tag and a runOn tag, so without them this file never
    // auto-runs on component changes - it runs only when explicitly targeted or via
    // the changed-test-files PR flow
    { tag: [tags.resource.distributedTrainingWorkload, tags.resource.patchRunaiConfig] },
    function () {
        const __filename = fileURLToPath(import.meta.url);
        const __dirname = path.dirname(__filename);
        let project: Project;
        let namespace: string;
        const namePrefix = objectGenerator.generateShortNameByPrefix('wlr', 4);

        // Shared across the ordered tests below (Playwright runs tests of a
        // single file in declaration order on one worker). Chain map:
        //   wlHappy - tests of the restart lifecycle: ordered cap accounting
        //             (restart #1, then union/dedup, then the 3-restart cap)
        //   wlRules - restart-rules chain, self-healing via ensureWorkload
        //   wlSched - single-test: real-node exclusion scheduling
        //   wlMerge - single-test: merge with user-set node anti-affinity
        //   wlRace  - suspend/resume race chain. MUST stay a PLAIN workload (no user
        //             exclusions): typed suspend/resume 404s for workloads created
        //             with workerExcludeNodeNames (flagged product gap), so the race
        //             tests must not share wlMerge
        const wlHappy = `${namePrefix}-happy`;
        const wlRules = `${namePrefix}-rules`;
        const wlSched = `${namePrefix}-sched`;
        const wlMerge = `${namePrefix}-merge`;
        const wlRace = `${namePrefix}-race`;
        let wlHappyId: string;
        let wlRulesId: string;
        let wlRaceId: string;

        // runaiconfig restore bookkeeping: restore in afterAll ONLY if we changed it
        let adviceOperandPriorSpec: any;
        let adviceOperandChanged = false;
        let distributedToggle: { changed: boolean; priorValue?: string } = { changed: false };


        test.beforeAll(async () => {
            test.setTimeout(300 * 1000);

            // No released version carries the WLR restart flow yet, so instead of a
            // version gate we probe the WorkloadRunner CRD schema for the restart
            // fields (spec.managedBy + spec.instructions.nodeAffinity).
            test.skip(
                !(await clusterSupportsWlrRestart()),
                'WorkloadRunner CRD on this cluster has no restart fields (managedBy / instructions.nodeAffinity)'
            );

            // Must precede the CP restarts below: they stall runai-agent's sync
            // loop for longer than the project-ready wait.
            project = await projectResourcesUtils.createProjectAndNamespace(namePrefix);
            namespace = Utils.namespaceFromProjectName(project.name);
            await waitUtils.waitForNamespaceToExist(namespace);

            // The distributed WLR route ships dark (chart default off) - enable it
            // for this suite and restore in afterAll, so the tests run in CI
            // instead of skipping (serial project, nothing else runs concurrently).
            distributedToggle = await ensureDistributedWlrSubmission();
            if (distributedToggle.changed) {
                logger.info('Enabled ANY_WORKLOAD_DISTRIBUTED_SUBMISSION for this suite');
            }

            // Advice CRD must exist BEFORE the operand is enabled (else CrashLoop).
            const adviceCrdPath = path.resolve(__dirname, 'testdata/heimdall.nvidia.com_advices.yaml');
            await kubeCli.createAllFromYaml(fs.readFileSync(adviceCrdPath, 'utf8'));
            logger.info('Ensured Advice CRD is installed');

            // snapshot the current runaiconfig value so afterAll restores it
            // only if this suite actually changed it
            const runaiConfig = (await kubeCli.getCrd(crds.runaiConfigCrd, 'runai', 'runai')) as any;
            adviceOperandPriorSpec = runaiConfig?.spec?.['advice-restart-controller'];

            if (adviceOperandPriorSpec?.enabled !== true) {
                await kubeCli.patchRunaiConfig([
                    {
                        op: 'add',
                        path: '/spec/advice-restart-controller',
                        value: { enabled: true }
                    }
                ]);
                adviceOperandChanged = true;
                logger.info('Enabled advice-restart-controller via runaiconfig');
            }

            await waitUtils.waitForDeploymentToBeAvailable('advice-restart-controller', 'runai', 120);
            await waitForCpVisibleAdviceOperand();

            if (adviceOperandChanged) {
                // workloads-manager caches the DisplayedCluster (operands included)
                // for ClusterCacheExpirationSec=600s. A pre-flip cache entry makes it
                // silently skip the workflow-id stamp for up to 10 minutes - bust the
                // cache by restarting the deployment.
                await kubeCli.restartDeployment('runai-backend-workloads-manager', 'runai-backend');
                await waitForDeploymentRollout('runai-backend-workloads-manager', 'runai-backend');
            }
        });

        test.afterAll(async function () {
            if (!project) return;
            try {
                await deleteAllCrInstancesInNamespace(crds.adviceCrd, namespace);
                await deleteAllCrInstancesInNamespace(crds.restartRequestCrd, namespace);
            } catch (e) {
                logger.error('Failed to clean up Advice/RestartRequest CRs');
            }
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (e) {
                logger.error('Failed to delete project');
            }
            if (distributedToggle.changed) {
                try {
                    await restoreDistributedWlrSubmission(distributedToggle.priorValue);
                    logger.info('Restored ANY_WORKLOAD_DISTRIBUTED_SUBMISSION to its prior state');
                } catch (e) {
                    logger.error('Failed to restore ANY_WORKLOAD_DISTRIBUTED_SUBMISSION');
                }
            }
            if (adviceOperandChanged) {
                try {
                    await kubeCli.patchRunaiConfig([
                        adviceOperandPriorSpec === undefined
                            ? { op: 'remove', path: '/spec/advice-restart-controller' }
                            : {
                                  op: 'add',
                                  path: '/spec/advice-restart-controller',
                                  value: adviceOperandPriorSpec
                              }
                    ]);
                    logger.info('Restored prior advice-restart-controller runaiconfig value');
                } catch (e) {
                    logger.error('Failed to restore runaiconfig advice-restart-controller value');
                }
            }
        });

        test.describe.serial(
            'restart lifecycle and exclusion accounting (wlHappy)',
            { tag: [tags.runOn.basicCI, tags.runOn.serialCI] },
            function () {
            test('submit enrolls the WorkloadRunner and advice restarts it with node exclusions', async function () {
                test.setTimeout(300 * 1000);
                const adviceName = `advice-${wlHappy}-1`;

                await test.step('submit distributed PyTorch -> WorkloadRunner + child PyTorchJob', async () => {
                    wlHappyId = await createPytorchWorkload(String(project.id), wlHappy);
                    await waitUtils.waitForCrd(crds.workloadRunnerCrd, wlHappy, namespace, 90);
                    await waitUtils.waitForCrd(crds.pyTorchJobCrd, wlHappy, namespace, 90);
                });

                let baselineGeneration: number;

                await test.step('WLR carries identity + heimdall enrollment', async () => {
                    const wlr = await getWlr(wlHappy, namespace);
                    expect(wlr.metadata.annotations?.[WORKLOAD_ID_ANNOTATION]).toBe(wlHappyId);
                    expect(wlr.metadata.labels?.['run.ai/workload-kind']).toBe('PyTorchJob');

                    // workflow-id stamp gates on operand readiness reported via
                    // cluster-sync (~10s freshness) - poll instead of asserting once.
                    await waitUtils.waitForCondition(
                        async () => {
                            const w = await getWlr(wlHappy, namespace);
                            return w.spec?.template?.metadata?.labels?.[WORKFLOW_ID_LABEL] === wlHappyId;
                        },
                        'workflow-id label stamped on WLR template',
                        true,
                        60
                    );

                    const child = await getPytorchJob(wlHappy, namespace);
                    expect(child.metadata.labels?.[WORKFLOW_ID_LABEL]).toBe(wlHappyId);

                    baselineGeneration = wlr.metadata.generation;
                });

                await test.step('advice creates a fully-shaped RestartRequest and the restart completes: RR Succeeded, lock cleared, child recreated', async () => {
                    await runRestartCycle(wlHappy, namespace, adviceName, ['node-b', 'node-a'], {
                        onRequestCreated: (rr: any): void => {
                            expect(rr.metadata.labels[CREATED_BY_LABEL]).toBe('advice-restart-controller');
                            expect(rr.metadata.labels[WORKLOAD_NAME_LABEL]).toBe(wlHappy);
                            expect(rr.metadata.labels[WORKLOAD_ID_LABEL]).toBe(wlHappyId);
                            const ownerRef = rr.metadata.ownerReferences?.[0];
                            expect(ownerRef?.kind).toBe('WorkloadRunner');
                            expect(ownerRef?.name).toBe(wlHappy);
                            expect(rr.spec.workloadRef.gvk.kind).toBe('WorkloadRunner');
                            expect(rr.spec.workloadRef.gvk.group).toBe('run.ai');
                            expect(rr.spec.workloadRef.name).toBe(wlHappy);
                            const overridePaths = (rr.spec.specOverrides as any[]).map((o: any) => o.path);
                            expect(overridePaths).toEqual([NODE_AFFINITY_OVERRIDE_PATH]);
                        }
                    });
                });

                await test.step('exclusions persisted on the WLR and materialized on the new child', async () => {
                    const expected = ['node-a', 'node-b']; // sorted regardless of advice order
                    const wlr = await getWlr(wlHappy, namespace);
                    expect(wlrInstructionExclusions(wlr)).toEqual(expected);

                    const child = await getPytorchJob(wlHappy, namespace);
                    expect(childReplicaExclusions(child, 'Master')).toEqual(expected);
                    expect(childReplicaExclusions(child, 'Worker')).toEqual(expected);
                });

                await test.step('identity is stable across the restart (hard requirement)', async () => {
                    const wlr = await getWlr(wlHappy, namespace);
                    expect(wlr.metadata.annotations?.[WORKLOAD_ID_ANNOTATION]).toBe(wlHappyId);
                    expect(wlr.metadata.generation).toBeGreaterThanOrEqual(baselineGeneration + 3);

                    const child = await getPytorchJob(wlHappy, namespace);
                    expect(child.metadata.labels?.[WORKFLOW_ID_LABEL]).toBe(wlHappyId);

                    // NOTE: the TYPED read (GET /api/v1/workloads/distributed/{id})
                    // returns 404 'V2 workload snapshot not found' for WLR-routed
                    // workloads - flagged product gap. The untyped workloads API works.
                    const workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(wlHappyId)).body;
                    expect(workload.id).toBe(wlHappyId);
                    expect(workload.name).toBe(wlHappy);
                });

                await test.step('restart lifecycle events emitted on the WLR', async () => {
                    const events = await getWlrEvents(wlHappy, namespace);
                    const reasons = events.map((e: any) => e.reason);
                    expect(reasons).toContain('RestartInitiated');
                    expect(reasons).toContain('RestartSucceeded');
                });
            });

            test('exclusions accumulate across restarts and the cap then ignores further advices', async function () {
                test.setTimeout(300 * 1000);

                await test.step('second restart with a new exclusion unions into the set', async () => {
                    await runRestartCycle(wlHappy, namespace, `advice-${wlHappy}-2`, ['node-c']);
                    const wlr = await getWlr(wlHappy, namespace);
                    expect(wlrInstructionExclusions(wlr)).toEqual(['node-a', 'node-b', 'node-c']);
                });

                await test.step('third restart with a duplicate exclusion does not duplicate', async () => {
                    await runRestartCycle(wlHappy, namespace, `advice-${wlHappy}-3`, ['node-a']);
                    const wlr = await getWlr(wlHappy, namespace);
                    expect(wlrInstructionExclusions(wlr)).toEqual(['node-a', 'node-b', 'node-c']);

                    const child = await getPytorchJob(wlHappy, namespace);
                    expect(childReplicaExclusions(child, 'Master')).toEqual(['node-a', 'node-b', 'node-c']);
                    expect(childReplicaExclusions(child, 'Worker')).toEqual(['node-a', 'node-b', 'node-c']);
                });

                await test.step('a fourth advice is ignored: no RestartRequest, AdviceIgnored event', async () => {
                    const adviceName = `advice-${wlHappy}-4`;

                    const child = await getPytorchJob(wlHappy, namespace);
                    await kubeCli.createCrd(
                        crds.adviceCrd,
                        buildAdviceBody(
                            adviceName,
                            namespace,
                            wlHappy,
                            child.metadata.uid,
                            'ADVICE_STATE_RESTARTED'
                        ),
                        namespace
                    );

                    await expectRestartRequestCountAfterSettle(adviceName, namespace);

                    const events = await getWlrEvents(wlHappy, namespace);
                    const ignored = events.filter((e: any) => e.reason === 'AdviceIgnored');
                    expect(ignored.length).toBeGreaterThanOrEqual(1);
                });
            });
        });

        test.describe('restart rules on a shared workload (wlRules)', function () {
            test.beforeEach(async () => {
                wlRulesId = await ensureWorkload(String(project.id), namespace, wlRules, wlRulesId);
            });

            test('re-applying the same advice does not trigger a second restart', async function () {
                test.setTimeout(300 * 1000);
                const adviceName = `advice-${wlRules}-1`;

                await test.step('submit and restart once', async () => {
                    await runRestartCycle(wlRules, namespace, adviceName);
                });

                await test.step('touching the advice creates no second RestartRequest', async () => {
                    await kubeCli.patchCrd(
                        crds.adviceCrd,
                        { metadata: { annotations: { 'e2e/touched': `${Date.now()}` } } },
                        adviceName,
                        namespace
                    );
                    await expectRestartRequestCountAfterSettle(adviceName, namespace, 1);
                });
            });

            test('advice is ignored: TERMINATED state, and suspended workload', async function () {
                test.setTimeout(300 * 1000);
                const termAdvice = `advice-${wlRules}-term`;
                const suspAdvice = `advice-${wlRules}-susp`;

                // ordering is load-bearing: suspension removes the WLR in-cluster and
                // RestartRequests are owner-referenced to it, so the TERMINATED advice's
                // zero-count MUST be asserted BEFORE suspending - otherwise GC could
                // delete the evidence and the assertion would become vacuous.
                await test.step('TERMINATED advice against the running workload creates no RestartRequest', async () => {
                    const child = await getPytorchJob(wlRules, namespace);
                    await kubeCli.createCrd(
                        crds.adviceCrd,
                        buildAdviceBody(
                            termAdvice,
                            namespace,
                            wlRules,
                            child.metadata.uid,
                            'ADVICE_STATE_TERMINATED'
                        ),
                        namespace
                    );

                    await expectRestartRequestCountAfterSettle(termAdvice, namespace);
                });

                await test.step('suspend the workload and wait until fully Stopped', async () => {
                    await workloadsCrudUtils.suspendDistributedTraining(wlRulesId);
                    // resume is only permitted from Stopped, and the in-cluster WLR
                    // is REMOVED while the workload is suspended - wait on the CP phase.
                    await waitUtils.waitForBackendWorkloadPhase(wlRules, project.id, Phase.Stopped, 180);
                });

                await test.step('RESTARTED advice against the suspended workload creates no RestartRequest', async () => {
                    // while suspended neither the child nor the WLR exist in-cluster;
                    // the advice must resolve to nothing and create no RR.
                    await kubeCli.createCrd(
                        crds.adviceCrd,
                        buildAdviceBody(
                            suspAdvice,
                            namespace,
                            wlRules,
                            'suspended-child-uid',
                            'ADVICE_STATE_RESTARTED'
                        ),
                        namespace
                    );
                    await expectRestartRequestCountAfterSettle(suspAdvice, namespace);
                });

                await test.step('resume restores the WorkloadRunner and child', async () => {
                    await workloadsCrudUtils.resumeDistributedTraining(wlRulesId);
                    await waitForWlrAndChildBack(wlRules, namespace);
                });
            });

            test('RestartRequest admission rejects malformed requests', async function () {
                test.setTimeout(180 * 1000);
                const wlr = await getWlr(wlRules, namespace);
                const liveGeneration = wlr.metadata.generation;

                const validRr = (name: string): any =>
                    buildRestartRequestBody({
                        name,
                        namespace,
                        wlName: wlRules,
                        workloadId: wlRulesId,
                        generation: liveGeneration
                    });

                await test.step('workload-name label must match workloadRef.name', async () => {
                    const rr = validRr('e2e-rr-badname');
                    rr.metadata.labels[WORKLOAD_NAME_LABEL] = 'some-other-workload';
                    await expect(kubeCli.createCrd(crds.restartRequestCrd, rr, namespace)).rejects.toThrow();
                });

                await test.step('GVK allowlist rejects non-WorkloadRunner targets', async () => {
                    const rr = validRr('e2e-rr-badgvk');
                    rr.spec.workloadRef.gvk = {
                        group: 'run.ai',
                        version: 'v2alpha1',
                        kind: 'DistributedWorkload'
                    };
                    await expect(kubeCli.createCrd(crds.restartRequestCrd, rr, namespace)).rejects.toThrow();
                });

                await test.step('override path allowlist rejects foreign paths', async () => {
                    const rr = validRr('e2e-rr-badpath');
                    rr.spec.specOverrides = [{ op: 'add', path: '/spec/template', value: {} }];
                    await expect(kubeCli.createCrd(crds.restartRequestCrd, rr, namespace)).rejects.toThrow();
                });

                await test.step('stale pinned generation is rejected', async () => {
                    const rr = validRr('e2e-rr-stalegen');
                    rr.spec.workloadRef.generation = Math.max(1, liveGeneration - 1);
                    await expect(kubeCli.createCrd(crds.restartRequestCrd, rr, namespace)).rejects.toThrow();
                });

                await test.step('referenced workload must exist', async () => {
                    const rr = validRr('e2e-rr-nowl');
                    rr.metadata.labels[WORKLOAD_NAME_LABEL] = 'ghost-workload';
                    rr.spec.workloadRef.name = 'ghost-workload';
                    await expect(kubeCli.createCrd(crds.restartRequestCrd, rr, namespace)).rejects.toThrow();
                });
            });

            test('WLR restart lock fields are webhook-protected', async function () {
                test.setTimeout(180 * 1000);

                await test.step('foreign managedBy value is denied', () => {
                    expectKubectlDenied(
                        `kubectl patch workloadrunners.run.ai ${wlRules} -n ${namespace} --type=merge -p '{"spec":{"managedBy":"foreign-controller"}}'`,
                        'managedBy'
                    );
                });

                await test.step('request-ref label without managedBy is denied', () => {
                    expectKubectlDenied(
                        `kubectl label workloadrunners.run.ai ${wlRules} -n ${namespace} ${REQUEST_REF_LABEL}=fake-uid`,
                        'request-ref'
                    );
                });

                await test.step('workflow-id template label cannot be removed (VAP)', () => {
                    expectKubectlDenied(
                        `kubectl patch workloadrunners.run.ai ${wlRules} -n ${namespace} --type=json -p '[{"op":"remove","path":"/spec/template/metadata/labels/heimdall.nvidia.com~1workflow-id"}]'`,
                        'workflow-id'
                    );
                });
            });

            test('invalid node affinity is rejected before the child is deleted (dry-run fence)', async function () {
                test.setTimeout(300 * 1000);

                // multi-term affinity is a supported feature (maxTerms=16), so the
                // invalid payload here is an EMPTY required term - the validator
                // rejects it with "must have at least one matchExpressions or
                // matchFields entry".
                const invalidAffinity = {
                    nodeAffinity: {
                        requiredDuringSchedulingIgnoredDuringExecution: {
                            nodeSelectorTerms: [{}]
                        }
                    }
                };

                await test.step('direct invalid instruction is denied at admission', () => {
                    const patch = { spec: { instructions: { nodeAffinity: invalidAffinity } } };
                    expectKubectlDenied(
                        `kubectl patch workloadrunners.run.ai ${wlRules} -n ${namespace} --type=merge -p '${JSON.stringify(patch)}'`,
                        'must have at least one'
                    );
                });

                await test.step('RR carrying an invalid override fails without deleting the child', async () => {
                    const wlr = await getWlr(wlRules, namespace);
                    const child = await getPytorchJob(wlRules, namespace);
                    const childUid = child.metadata.uid;

                    const rr: any = buildRestartRequestBody({
                        name: 'e2e-rr-invalidaffinity',
                        namespace,
                        wlName: wlRules,
                        workloadId: wlRulesId,
                        generation: wlr.metadata.generation,
                        specOverrides: [
                            { op: 'add', path: NODE_AFFINITY_OVERRIDE_PATH, value: invalidAffinity }
                        ]
                    });
                    await kubeCli.createCrd(crds.restartRequestCrd, rr, namespace);

                    await waitUtils.waitForCondition(
                        async () => {
                            const got = (await kubeCli.getCrd(
                                crds.restartRequestCrd,
                                'e2e-rr-invalidaffinity',
                                namespace
                            )) as any;
                            return got?.status?.phase === 'Failed';
                        },
                        'invalid-affinity RR reached Failed',
                        true,
                        120
                    );

                    const survivingChild = await getPytorchJob(wlRules, namespace);
                    expect(survivingChild?.metadata?.uid, 'child must survive a failed dry-run').toBe(
                        childUid
                    );

                    await waitUtils.waitForCondition(
                        async () => {
                            const w = await getWlr(wlRules, namespace);
                            return !w.spec?.managedBy && !w.metadata?.labels?.[REQUEST_REF_LABEL];
                        },
                        'handback after failed RR',
                        true,
                        60
                    );
                });
            });

            // MUST stay last in this block - it deletes wlRules
            test('deleting the workload garbage-collects its restart requests', async function () {
                test.setTimeout(300 * 1000);

                await workloadsCrudUtils.deleteDistributed(wlRulesId);

                await waitUtils.waitForCondition(
                    async () => {
                        const rrs = await kubeCli.getCrds(crds.restartRequestCrd, namespace);
                        return (
                            (rrs || []).filter(
                                (rr: any) => rr.metadata?.labels?.[WORKLOAD_NAME_LABEL] === wlRules
                            ).length === 0
                        );
                    },
                    'RestartRequests garbage-collected after workload deletion',
                    true,
                    120
                );
            });
        });

        test.describe('enrollment scope', function () {
            test('raw v2 manifest submission is not heimdall-enrolled', async function () {
                test.setTimeout(300 * 1000);
                const wlName = `${namePrefix}-rawv2`;

                // Only native (v1-cast) distributed pytorch is stamped: raw v2
                // manifests are mutable via the v2 update endpoint and a restart
                // cannot race a user update, so the CP does not enroll them at all
                // The typed-API opt-out remains unreachable because
                // spec.annotations land on the replica PodTemplates, not on the
                // manifest metadata the opt-out gate inspects (flagged gap).
                await test.step('submit a raw v2 PyTorchJob manifest via the CLI', async () => {
                    const manifest = [
                        'apiVersion: kubeflow.org/v1',
                        'kind: PyTorchJob',
                        'metadata:',
                        `  name: ${wlName}`,
                        'spec:',
                        '  pytorchReplicaSpecs:',
                        '    Master:',
                        '      replicas: 1',
                        '      restartPolicy: OnFailure',
                        '      template:',
                        '        spec:',
                        '          containers:',
                        '            - name: pytorch',
                        '              image: alpine:latest',
                        '              command: ["sh", "-c", "sleep infinity"]',
                        ''
                    ].join('\n');
                    const manifestPath = `${os.tmpdir()}/wlr-rawv2-${Date.now()}.yaml`;
                    fs.writeFileSync(manifestPath, manifest);

                    cliLoginAsSystemAdmin();
                    bashCli.exec(
                        `${runaiCliV2.cliPath} workload submit ${wlName} --file ${manifestPath} -p ${project.name}`
                    );
                    await waitUtils.waitForCrd(crds.workloadRunnerCrd, wlName, namespace, 90);
                });

                await test.step('the WLR is never stamped with a workflow-id', async () => {
                    // give the stamp path its full propagation window before asserting absence
                    await Utils.delay(15000);
                    const wlr = await getWlr(wlName, namespace);
                    expect(
                        wlr.spec?.template?.metadata?.labels?.[WORKFLOW_ID_LABEL],
                        'raw v2 WLR must never be stamped with a workflow-id'
                    ).toBeUndefined();
                });
            });

            test('advice targeting a non-PyTorchJob workload is ignored', async function () {
                test.setTimeout(300 * 1000);
                const wlName = `${namePrefix}-tw`;
                const adviceName = `advice-${wlName}`;

                await test.step('submit a regular training workload through the CP', async () => {
                    // a plain training workload also routes through WorkloadRunner, but its
                    // child is a batch Job - only PyTorchJob is restartable, so an advice
                    // pointing at it must be ignored ("unsupported workload type")
                    await workloadsCrudUtils.createTraining(
                        wlName,
                        false,
                        flowRootState.cluster.uuid,
                        String(project.id),
                        {
                            image: 'alpine:latest',
                            command: 'sleep infinity',
                            compute: {
                                gpuDevicesRequest: 0,
                                cpuCoreRequest: 0.1,
                                cpuMemoryRequest: '100M'
                            }
                        },
                        flowRootState.users.system_administrator.accessToken
                    );
                    await waitUtils.waitForCrd(crds.workloadRunnerCrd, wlName, namespace, 90);
                });

                await test.step('advice at the batch-Job child creates no RestartRequest', async () => {
                    let childJob: any;
                    await waitUtils.waitForCondition(
                        async () => {
                            try {
                                childJob = await kubeCli.getBatchJob(wlName, namespace);
                                return childJob !== undefined;
                            } catch {
                                return false;
                            }
                        },
                        `batch Job child of ${wlName} created`,
                        true,
                        90
                    );

                    await kubeCli.createCrd(
                        crds.adviceCrd,
                        buildAdviceBody(
                            adviceName,
                            namespace,
                            wlName,
                            String(childJob.metadata?.uid),
                            'ADVICE_STATE_RESTARTED',
                            undefined,
                            { apiVersion: 'batch/v1', kind: 'Job' }
                        ),
                        namespace
                    );

                    await expectRestartRequestCountAfterSettle(adviceName, namespace);
                });
            });
        });

        test.describe('real-node scheduling', function () {
            test('excluding the live node reschedules the workload to another node', async function () {
                test.setTimeout(480 * 1000);

                const schedulable = await skipIfFewerThanTwoSchedulableNodes(test, 'real-node exclusion');

                let excluded: string[] = [];
                let wlSchedId = '';

                await test.step('submit and record the nodes the pods run on', async () => {
                    wlSchedId = await createPytorchWorkload(String(project.id), wlSched);
                    await waitUtils.waitForCrd(crds.workloadRunnerCrd, wlSched, namespace, 90);
                    await waitUtils.waitForCrd(crds.pyTorchJobCrd, wlSched, namespace, 90);
                });

                await test.step('hop 1: exclude the live node(s), pods move elsewhere', async () => {
                    const currentNodes = await waitForScheduledNodes(wlSchedId);
                    // never exclude everything - leave at least one schedulable node
                    const toExclude =
                        schedulable.length - currentNodes.length >= 1
                            ? currentNodes
                            : currentNodes.slice(0, 1);
                    excluded = [...toExclude];

                    await runRestartCycle(wlSched, namespace, `advice-${wlSched}-hop1`, toExclude);
                    await waitForNodesOffExcluded(wlSchedId, excluded);
                });

                await test.step('hop 2: exclude the new node too (only when capacity allows)', async () => {
                    const currentNodes = await waitForScheduledNodes(wlSchedId);
                    const union = [...new Set([...excluded, ...currentNodes])];
                    if (schedulable.length - union.length < 1) {
                        logger.info(
                            `skipping hop 2 - excluding ${union.length} nodes would leave no capacity`
                        );
                        return;
                    }

                    await runRestartCycle(wlSched, namespace, `advice-${wlSched}-hop2`, currentNodes);
                    excluded = union;
                    await waitForNodesOffExcluded(wlSchedId, excluded);
                });
            });

            test('advice exclusions merge with user-set node anti-affinity', async function () {
                test.setTimeout(480 * 1000);

                const schedulable = await skipIfFewerThanTwoSchedulableNodes(test, 'anti-affinity merge');

                const userExcluded = schedulable[0];

                let wlMergeId = '';

                await test.step('submit with user anti-affinity, pods avoid the user-excluded node', async () => {
                    wlMergeId = await createPytorchWorkload(String(project.id), wlMerge, [userExcluded]);
                    await waitUtils.waitForCrd(crds.workloadRunnerCrd, wlMerge, namespace, 90);
                    await waitUtils.waitForCrd(crds.pyTorchJobCrd, wlMerge, namespace, 90);

                    const nodes = await waitForScheduledNodes(wlMergeId);
                    expect(nodes).not.toContain(userExcluded);
                });

                await test.step('advice exclusion is added on top of the user affinity', async () => {
                    const currentNodes = await waitForScheduledNodes(wlMergeId);
                    const adviceExcluded = currentNodes[0];

                    await runRestartCycle(wlMerge, namespace, `advice-${wlMerge}-merge`, [adviceExcluded]);

                    // CR-level: user affinity and plugin exclusions may live in SEPARATE
                    // matchExpressions, so read the union across all of them. The
                    // workerExcludeNodeNames semantics are defined on the Worker replica.
                    const child = await getPytorchJob(wlMerge, namespace);
                    const workerExclusions = allHostnameExclusions(childReplicaTerms(child, 'Worker'));
                    expect(workerExclusions).toContain(adviceExcluded);
                    expect(workerExclusions).toContain(userExcluded);
                    logger.info(
                        `master-replica exclusion union: [${allHostnameExclusions(childReplicaTerms(child, 'Master')).join(',')}]`
                    );

                    // scheduling-level only when a free node remains
                    const blocked = [...new Set([userExcluded, adviceExcluded])];
                    if (schedulable.length - blocked.length >= 1) {
                        await waitForNodesOffExcluded(wlMergeId, blocked);
                    } else {
                        logger.info('no free node left for scheduling assertion - CR-level check only');
                    }
                });
            });
        });

        test.describe('suspend/resume races (wlRace)', function () {
            test('suspend/resume races never leave a stuck lock or fire a dormant advice', async function () {
                test.setTimeout(480 * 1000);
                const raceAdvice = `advice-${wlRace}-suspendrace`;
                const dormantAdvice = `advice-${wlRace}-resumerace`;

                await test.step('ensure the race workload exists and runs', async () => {
                    wlRaceId = await ensureWorkload(String(project.id), namespace, wlRace, wlRaceId);
                });

                await test.step('fire advice and suspend concurrently', async () => {
                    const child = await getPytorchJob(wlRace, namespace);
                    await kubeCli.createCrd(
                        crds.adviceCrd,
                        buildAdviceBody(
                            raceAdvice,
                            namespace,
                            wlRace,
                            child.metadata.uid,
                            'ADVICE_STATE_RESTARTED'
                        ),
                        namespace
                    );
                    // KNOWN GAP (EW recreation, design J3): while the restart
                    // recreates the child, cluster-sync drops the CP workload row (typed ops
                    // 404 "Failed to get workload") and then re-creates it passing through
                    // Creating/Initializing (suspend 400 "not permitted for Creating
                    // workloads"). A real user suspend in that window fails the same way -
                    // retry until the workload is actionable again and the suspend is
                    // accepted. The invariant under test is unchanged: suspend eventually
                    // wins and no restart lock is left behind.
                    await waitUtils.waitForCondition(
                        async () => {
                            // check the workload row is back and actionable BEFORE (re)trying
                            // the suspend - during the restart it vanishes and then comes
                            // back in Creating
                            try {
                                const workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(wlRaceId))
                                    .body;
                                if (
                                    !workload ||
                                    ['Creating', 'Initializing'].includes(String(workload.phase))
                                ) {
                                    return false;
                                }
                            } catch {
                                return false;
                            }
                            try {
                                await workloadsCrudUtils.suspendDistributedTraining(wlRaceId);
                                return true;
                            } catch (err: any) {
                                if (err?.statusCode === 404) return false;
                                if (
                                    err?.statusCode === 400 &&
                                    String(err?.body?.message ?? '').includes('not permitted for')
                                ) {
                                    return false;
                                }
                                throw err;
                            }
                        },
                        'workload actionable again and suspend accepted',
                        true,
                        180,
                        3
                    );
                });

                await test.step('suspend wins: workload reaches Stopped, no non-terminal RR remains', async () => {
                    await waitUtils.waitForBackendWorkloadPhase(wlRace, project.id, Phase.Stopped, 300);

                    // whatever the race produced, the RR must be terminal (or absent) -
                    // a Pending/InProgress RR against a suspended workload is a stuck lock
                    await waitUtils.waitForCondition(
                        async () => {
                            const rrs = await getRestartRequestsForAdvice(raceAdvice, namespace);
                            return rrs.every((rr: any) => ['Succeeded', 'Failed'].includes(rr.status?.phase));
                        },
                        'race RR reached a terminal phase (or none was created)',
                        true,
                        120
                    );
                });

                await test.step('fire a second advice against the suspended workload, then resume', async () => {
                    await kubeCli.createCrd(
                        crds.adviceCrd,
                        buildAdviceBody(
                            dormantAdvice,
                            namespace,
                            wlRace,
                            'suspended-child-uid',
                            'ADVICE_STATE_RESTARTED'
                        ),
                        namespace
                    );

                    // belt-and-braces: record the RR set we resume on top of
                    const rrs = ((await kubeCli.getCrds(crds.restartRequestCrd, namespace)) || []) as any[];
                    logger.info(
                        `RestartRequests before resume: ${JSON.stringify(
                            rrs.map((rr: any) => ({ name: rr.metadata?.name, phase: rr.status?.phase }))
                        )}`
                    );

                    await workloadsCrudUtils.resumeDistributedTraining(wlRaceId);
                });

                await test.step('resume restores WLR/child; dormant advice never fires; lock is clean', async () => {
                    await waitForWlrAndChildBack(wlRace, namespace);

                    // settle window: an ignored advice must stay dormant - the child
                    // existing again must NOT retroactively trigger a restart
                    await expectRestartRequestCountAfterSettle(dormantAdvice, namespace, 0, 20);

                    const wlr = await getWlr(wlRace, namespace);
                    expect(wlr.spec?.managedBy ?? '', 'WLR must be handed back: managedBy cleared').toBe('');
                    expect(
                        wlr.metadata?.labels?.[REQUEST_REF_LABEL],
                        'WLR must be handed back: request-ref label removed'
                    ).toBeUndefined();
                });
            });
        });
    }
);
