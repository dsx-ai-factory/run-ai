/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    externalWorkloadUtils,
    flowRootState,
    inferenceClientCreator,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { supportedFrom, SupportedVersions } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';

const normalizeIds = (ids: string[]): string[] => [...ids].sort();

test.describe(
    `Element structure e2e (Deployment), tags: ${tags.teams.viper}, ${tags.resource.workloads}`,
    { tag: [tags.services.externalWorkloadIntegrator, tags.runOn.parallelCI] },
    () => {
        test.describe.configure({ mode: 'serial', timeout: 180_000 });

        supportedFrom(SupportedVersions.V2_25)(() => {
            let project: Project;
            let namespace: string;
            let projectName: string;
            let workloadId: string;
            let ewName: string;
            let initialElementIds: string[];

            const deploymentName = objectGenerator.generateNameByPrefix('struct');
            const initialReplicas = 2;

            test.beforeAll(async () => {
                projectName = objectGenerator.generateNameByPrefix('struct-proj');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await orgUnitUtils.createProjectAndValidate({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });

                const deployment = inferenceClientCreator.getDeployment(
                    namespace,
                    projectName,
                    deploymentName,
                    initialReplicas,
                    config.quickStartImage.ubuntu
                );
                await kubeCli.createDeployment(deployment);

                ewName = `ew-deployment-${deploymentName}`;

                workloadId = await workloadsUtilsV2.waitForWorkloadIdByWorkloadNameAndProjectId(
                    deploymentName,
                    project.id.toString()
                );

                await waitUtils.waitForCondition(
                    () => externalWorkloadUtils.hasExpectedPodCount(namespace, deploymentName, initialReplicas),
                    `wait for ${initialReplicas} pods to exist`,
                    true,
                    30
                );
            });

            test.afterAll(async () => {
                try {
                    await kubeCli.deleteDeployment(deploymentName, namespace);
                } catch (e) {
                    logger.warn(`Failed to delete deployment ${deploymentName}: ${e}`);
                }
                try {
                    if (project?.id) {
                        await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                    }
                } catch (e) {
                    logger.warn(`Failed to delete project ${projectName}: ${e}`);
                }
            });

            test('ExternalWorkload structureElements should be populated', async () => {
                await waitUtils.waitForCondition(
                    () => externalWorkloadUtils.hasStructureElements(ewName, namespace),
                    'structureElements to be populated on ExternalWorkload CRD',
                    true,
                    40
                );

                const elements = await externalWorkloadUtils.getStructureElements(ewName, namespace);

                expect(elements.length).toBeGreaterThan(0);
                for (const el of elements) {
                    expect(el.id).toBeDefined();
                    expect(typeof el.id).toBe('string');
                    expect(el.name).toBeDefined();
                    expect(typeof el.name).toBe('string');
                    expect(typeof el.level).toBe('number');
                }

                const physicalElements = elements.filter((el: any) => el.k8sRef?.uid);
                expect(physicalElements.length).toBeGreaterThan(0);

                initialElementIds = elements.map((el: any) => el.id);
                logger.info(`Initial element IDs: ${JSON.stringify(initialElementIds)}`);
            });

            test('Pods should be labeled with element IDs', async () => {
                await waitUtils.waitForCondition(
                    () => externalWorkloadUtils.allPodsHaveElementLabels(namespace, deploymentName, initialReplicas),
                    'all pods to have element labels',
                    true,
                    30
                );

                const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, deploymentName);
                for (const pod of pods) {
                    const elementLabel = pod.metadata?.labels?.['run.ai/element-1-id'];
                    expect(elementLabel).toBeDefined();
                    expect(initialElementIds).toContain(elementLabel);
                }
            });

            test('getWorkloadElements API should return correct elements', async () => {
                await waitUtils.waitForCondition(
                    () => workloadsUtilsV2.hasExpectedElementCount(workloadId, initialElementIds.length),
                    `getWorkloadElements to return ${initialElementIds.length} elements`,
                    true,
                    40
                );

                const res = await workloadsUtilsV2.getWorkloadElements(workloadId);
                for (const el of res.body.elements) {
                    expect(el.id).toBeDefined();
                    expect(typeof el.hierarchyLevel).toBe('number');
                    expect(el.name).toBeDefined();
                }

                const apiElementIds = res.body.elements.map((el: any) => el.id);
                expect(normalizeIds(apiElementIds)).toEqual(normalizeIds(initialElementIds));
            });

            test('countWorkloadElements API should return correct counts', async () => {
                const res = await workloadsUtilsV2.countWorkloadElements(workloadId);
                const { storedElementsCount, totalElementsCount } = res.body;

                expect(storedElementsCount).toBe(initialElementIds.length);
                expect(totalElementsCount).toBe(initialElementIds.length);
            });

            test('getWorkloadPods should return pod elements matching the workload elements', async () => {
                await waitUtils.waitForCondition(
                    () => workloadsUtilsV2.allPodsHaveElements(workloadId, initialReplicas),
                    'all pods to have elements in getWorkloadPods API',
                    true,
                    30
                );

                const podsRes = await workloadsUtilsV2.getWorkloadPods(workloadId);
                for (const pod of podsRes.body.pods) {
                    for (const podElement of pod.elements) {
                        expect(initialElementIds).toContain(podElement.elementId);
                        expect(podElement.hierarchyLevel).toBeGreaterThanOrEqual(1);
                    }
                }
            });

            test('Element IDs should remain stable after scaling', async () => {
                const scaledReplicas = 4;
                await inferenceClientCreator.setDeploymentReplica(
                    deploymentName,
                    scaledReplicas,
                    namespace
                );

                await waitUtils.waitForCondition(
                    () => externalWorkloadUtils.allPodsHaveElementLabels(namespace, deploymentName, scaledReplicas),
                    'all scaled pods to have element labels',
                    true,
                    40
                );

                // Verify element IDs on CRD are unchanged
                const elements = await externalWorkloadUtils.getStructureElements(ewName, namespace);
                expect(elements.length).toBeGreaterThan(0);

                const currentElementIds = elements.map((el: any) => el.id);
                expect(normalizeIds(currentElementIds)).toEqual(normalizeIds(initialElementIds));

                // Verify via API too
                await waitUtils.waitForCondition(
                    () => workloadsUtilsV2.elementIdsMatch(workloadId, initialElementIds),
                    'API element IDs to match after scaling',
                    true,
                    40
                );

                // Verify all pods have correct labels
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, deploymentName);
                expect(pods.length).toBeGreaterThanOrEqual(scaledReplicas);
                for (const pod of pods) {
                    const elementLabel = pod.metadata?.labels?.['run.ai/element-1-id'];
                    expect(elementLabel).toBeDefined();
                    expect(initialElementIds).toContain(elementLabel);
                }
            });
        });
    }
);
