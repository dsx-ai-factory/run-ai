/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    externalWorkloadUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../../../modules/helpers';
import { NIMServiceWorkload } from '../../../../../../../../../modules/helpers/models/externalWorkload';
import { PodGroupCrd } from '../../../../../../../../../modules/helpers/models/podGroups';
import { kubeCli } from '../../../../../../../../../modules/cli';
import { Project } from '../../../../../../../../../modules/api/apis/orgUnitApi';
import { isNimOperatorInstalled } from '../../../../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    createNgcAuthSecret,
    NIM_TEST_TIMEOUT_MS
} from '../../../../../../../../../modules/helpers/utils/nimServiceUtils';
import { Phase } from '../../../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, TestInfo } from '@playwright/test';
import { expect } from 'chai';

/**
 * validate the KAI podgrouper's karta-driven recognition of a workload type that
 * has a karta but NO hand-written podgrouper handler.
 *
 * NIMService (`apps.nvidia.com/v1alpha1`) is applied raw via `kubectl apply` into a Run:ai project
 * namespace (external-workload path, `source: 'Other'`). Because the NIMService karta is installed,
 * the external-workload-integrator grants the pod-grouper RBAC to read NIMService objects, so the
 * podgrouper walks the pod owner chain up to the NIMService top owner and attributes the PodGroup to
 * it. We assert exactly that: the PodGroup's ownerReference kind is `NIMService` (not the
 * operator-created Deployment/ReplicaSet). This behaviour is only present from v2.26 onwards; prior
 * versions behaved differently and are excluded via the supportedFrom gate.
 */
test.describe(
    `NIMService external workload podgrouper karta test, tags: ${tags.teams.viper} ${tags.resource.workloads} ${tags.resource.mimService}`,
    { tag: [tags.services.externalWorkloadIntegrator, tags.runOn.serialCI] },
    () => {
        supportedFrom(SupportedVersions.V2_26)(function () {
            let project: Project;
            let namespace: string;
            // NIM-operator presence is a runtime condition, so it can't gate the suite at collection
            // time (shouldRunTests needs a synchronous flag). We resolve it in beforeAll and no-op the
            // body when absent, instead of test.skip() which adds a noisy skips-report entry.
            // The nightlies that run this (@serial-ci) set nim-operator-enabled: true, so it runs
            // for real there; the no-op only triggers on environments without the NIM operator.
            let nimOperatorInstalled = false;
            const nimWorkloads: NIMServiceWorkload[] = [];

            test.beforeAll(async () => {
                nimOperatorInstalled = await isNimOperatorInstalled();
                if (!nimOperatorInstalled) {
                    logger.warn('k8s-nim-operator is not installed on the cluster, skipping test body');
                    return;
                }

                const projectName = objectGenerator.generateShortNameByPrefix('nim-ext');
                project = await orgUnitUtils.createProjectAndValidate({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    enforceRunaiScheduler: true,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });
                if (!project.id) {
                    throw new Error('Project ID is required for the test');
                }
                namespace = project.status.namespace;

                // The NIM pod mounts the NGC key secret; a dummy value is enough for the CPU NIM image.
                await createNgcAuthSecret(namespace);
                logger.info(
                    `Created NIMService external-workload project ${project.name}, namespace ${namespace}`
                );
            });

            test.afterEach(async () => {
                if (!nimOperatorInstalled) {
                    return;
                }
                while (nimWorkloads.length > 0) {
                    const workload = nimWorkloads.pop();
                    if (!workload) {
                        continue;
                    }
                    try {
                        await kubeCli.deleteCrd(crds.nimServiceCrd, workload.name, namespace);
                        await kubeCli.waitForCrdToBeDeleted(crds.nimServiceCrd, workload.name, namespace, 60);
                        logger.info(`Cleaned up NIMService ${workload.name}`);
                    } catch (error) {
                        logger.error(`Failed to cleanup NIMService ${workload.name}: ${error}`);
                    }
                }
            });

            test.afterAll(async () => {
                if (!nimOperatorInstalled) {
                    return;
                }
                if (project?.id) {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                }
            });

            test('Should schedule a raw NIMService and attribute its PodGroup to the NIMService top owner', async ({}, testInfo: TestInfo) => {
                if (!nimOperatorInstalled) {
                    logger.warn('k8s-nim-operator is not installed on the cluster, skipping test body');
                    return;
                }
                test.setTimeout(NIM_TEST_TIMEOUT_MS);

                const nimName = objectGenerator.generateNameByPrefix('nim-workload');
                const nimService = new NIMServiceWorkload(nimName, namespace, config.quickStartImage.nim);
                nimWorkloads.push(nimService);

                let workloadId: string;

                await test.step('Apply the raw NIMService and wait for it to surface in GET /workloads', async () => {
                    workloadId = await externalWorkloadUtils.submitWorkloadAndGetWorkloadId(
                        nimService,
                        project.id
                    );
                });

                await test.step('Wait for the NIMService to reach Running', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                });

                await test.step('Validate the workload API response (external workload, NIMService)', async () => {
                    await workloadsUtilsV2.getWorkloadAndValidate(workloadId, {
                        expectedType: 'NIMService',
                        expectedSource: 'Other',
                        expectedPhase: Phase.Running,
                        expectedRequestedPods: 1
                    });
                });

                await test.step('Validate the PodGroup is attributed to the NIMService top owner (karta mechanism)', async () => {
                    const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, nimName);
                    expect(pods.length, 'At least one NIMService pod should exist').to.be.greaterThan(0);

                    const podGroupName = pods[0].metadata?.annotations?.['pod-group-name'];
                    expect(podGroupName, 'Pod should carry a pod-group-name annotation').to.exist;

                    const podGroup = await kubeCli.getCrd<PodGroupCrd>(
                        crds.podGroupsCrd,
                        podGroupName,
                        namespace
                    );
                    expect(podGroup, `PodGroup ${podGroupName} should exist`).to.exist;

                    // The DefaultGrouper stamps the PodGroup with the TOP owner it resolved. If the
                    // podgrouper recognized NIMService (via karta-driven RBAC) the owner is NIMService;
                    // if it had been blocked and stopped lower, it would be Deployment/ReplicaSet.
                    const ownerRefs = podGroup.metadata?.ownerReferences ?? [];
                    expect(ownerRefs.length, 'PodGroup should have an owner reference').to.be.greaterThan(0);
                    expect(
                        ownerRefs[0].kind,
                        'PodGroup should be owned by the NIMService top owner'
                    ).to.equal('NIMService');
                    expect(ownerRefs[0].apiVersion).to.equal('apps.nvidia.com/v1alpha1');
                    expect(podGroup.metadata?.name).to.match(new RegExp(`^pg-${nimName}-`));

                    // No PodGroup in the namespace should be attributed to an object owned by the
                    // NIMService (e.g. the operator-created Deployment/ReplicaSet).
                    const allPodGroups = await kubeCli.getCrds(crds.podGroupsCrd, namespace);
                    for (const pg of allPodGroups) {
                        const ownerKind = pg.metadata?.ownerReferences?.[0]?.kind;
                        expect(
                            ownerKind,
                            `PodGroup ${pg.metadata?.name} should be owned by NIMService, not ${ownerKind}`
                        ).to.equal('NIMService');
                    }

                    logger.info(
                        `PodGroup ${podGroup.metadata?.name} is owned by NIMService ${nimName} — karta-driven top-owner recognition validated`
                    );
                });
            });
        });
    }
);
