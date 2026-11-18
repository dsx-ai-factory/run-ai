/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../modules/helpers';
import { expect, test } from '@playwright/test';
import { assetsApi } from '../../../../../../modules/api';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { PVCAsset, Scope } from '../../../../../../modules/api/apis/assetsApi';

test.describe(`Cluster-wide pvcs tests ${tags.teams.viper}`, () => {
    const RunaiNamespace = 'runai';
    const numberOfProjects = 3;
    const notReplicatedPvcName = objectGenerator.generateShortNameByPrefix('pvc-not-replicated');
    const ClusterScopedPvcName = objectGenerator.generateShortNameByPrefix('pvc');

    let clusterWidePvc: PVCAsset, projectScopedPvc: PVCAsset;
    const projects: Project[] = [];
    let deletedProjectId: number;
    let projectOfNonReplicatedPvc: Project;

    async function doesPvcExist(name: string, namespace: string): Promise<boolean> {
        const res = await kubeCli.CoreV1Api.listNamespacedPersistentVolumeClaim({ namespace });
        const pvcs = res.items;
        const foundPvc = pvcs.find((pvc) => pvc.metadata.name === name);
        return foundPvc !== undefined;
    }

    test.beforeAll(async function () {
        assetsApi.credentialsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        // create projects
        for (let index = 0; index < numberOfProjects; index++) {
            const project = await projectResourcesUtils.createProjectAndNamespace('p');
            projects.push(project);
        }
        projectOfNonReplicatedPvc = projects[0];
    });

    test.afterAll(async () => {
        logger.debug('Deleting all pvcs from the cluster');

        const deletePvc = async (id: string, pvcName: string, namespace: string): Promise<void> => {
            try {
                await dataSourcesUtils.deletePvcAsset(id);
                await waitUtils.waitForPvc(pvcName, namespace, 60, false);
                logger.debug(`Deleted PVC ${pvcName} from namespace ${namespace}`);
            } catch (e: any) {
                logger.warn(
                    `Failed to delete or confirm deletion of PVC ${pvcName} in ${namespace}: ${e.message}`
                );
            }
        };

        const deleteProject = async (project: Project): Promise<void> => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.debug(`Successfully deleted project and namespace: ${project.name}`);
            } catch (e: any) {
                logger.warn(`Failed to delete project ${project.name} (ID: ${project.id}): ${e.message}`);
            }
        };

        await Promise.all([
            deletePvc(projectScopedPvc.meta.id, notReplicatedPvcName, projectOfNonReplicatedPvc.namespace),
            deletePvc(clusterWidePvc.meta.id, ClusterScopedPvcName, RunaiNamespace)
        ]);

        for (const project of projects) {
            if (project.id !== deletedProjectId) {
                await deleteProject(project);
            }
        }
    });

    test('Cluster-wide PVC behavior end-to-end', async ({}, testInfo) => {
        assetsApi.credentialsApi.accessToken = flowRootState.users.system_administrator.accessToken;

        await test.step(`Create pvc in project namespace, ${tags.teams.viper}`, async () => {
            const pvcPayload = dataSourcesUtils.getPvcCreationPayload({
                name: notReplicatedPvcName,
                scope: Scope.Project,
                scopeId: projectOfNonReplicatedPvc.id.toString()
            });
            projectScopedPvc = await dataSourcesUtils.createPvcAsset(pvcPayload);
            expect(projectScopedPvc).toBeDefined();

            await waitUtils.waitForPvc(projectScopedPvc.meta.name, projectOfNonReplicatedPvc.namespace);
        });

        await test.step(`Create a pvc with clusterWideLabel and validate it is in runai namespace, ${tags.teams.viper}`, async () => {
            clusterWidePvc = await dataSourcesUtils.createPVC(ClusterScopedPvcName, Scope.Cluster, undefined);
            expect(clusterWidePvc).toBeDefined();

            await waitUtils.waitForPvc(ClusterScopedPvcName, RunaiNamespace);
        });

        await test.step(`Validate PVC replicated to all project namespaces (if needed), ${tags.teams.viper}`, async () => {
            for (const project of projects) {
                if (project.name === projectOfNonReplicatedPvc.name) continue;

                await waitUtils.waitForPvc(clusterWidePvc.meta.name, project.namespace, 60, true);
                const replicatedPvc = await waitUtils.getPvc(clusterWidePvc.meta.name, project.namespace);

                expect(replicatedPvc.metadata.name).toBe(clusterWidePvc.meta.name);
                expect(replicatedPvc.spec.accessModes[0]).toBe('ReadWriteOnce');
                expect(replicatedPvc.spec.resources.requests?.storage).toBe(
                    clusterWidePvc.spec.claimInfo.size
                );
                expect(replicatedPvc.spec.storageClassName).toBe(clusterWidePvc.spec.claimInfo.storageClass);
                expect(replicatedPvc.spec.volumeName).toBeUndefined();
            }
        });

        await test.step(`Validate PVC without clusterWideLabel is not replicated to project namespaces, ${tags.teams.viper}`, async () => {
            const [pvc1Exists, pvc2Exists] = await Promise.all([
                doesPvcExist(notReplicatedPvcName, projects[1].namespace),
                doesPvcExist(notReplicatedPvcName, projects[2].namespace)
            ]);
            expect(pvc1Exists).toBeFalsy();
            expect(pvc2Exists).toBeFalsy();
        });

        await test.step(`Delete a project and validate the cluster-wide PVC still exists in runai namespace, ${tags.teams.viper}`, async () => {
            deletedProjectId = projects[1].id;
            await operations.deleteProjectAndNamespace(deletedProjectId);
            const pvcExists = await doesPvcExist(ClusterScopedPvcName, RunaiNamespace);
            expect(pvcExists).toBeTruthy();
        });

        await test.step(`Delete created pvcs in runai namespace, ${tags.teams.viper}`, async () => {
            await dataSourcesUtils.deletePvcAsset(clusterWidePvc.meta.id);
            await waitUtils.waitForPvc(clusterWidePvc.meta.name, RunaiNamespace, 60, false);
        });

        await test.step(`Validate PVC deleted from all project namespaces, ${tags.teams.viper}`, async () => {
            const prms: Promise<void>[] = [];
            for (const project of projects) {
                if (project.name === projectOfNonReplicatedPvc.name) {
                    prms.push(waitUtils.waitForPvc(notReplicatedPvcName, project.namespace, 60, true));
                } else {
                    prms.push(waitUtils.waitForPvc(clusterWidePvc.meta.name, project.namespace, 60, false));
                }
            }
            await Promise.all(prms);
        });
    });
});
