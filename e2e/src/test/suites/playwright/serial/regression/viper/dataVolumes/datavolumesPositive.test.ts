/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    datavolumeUtils,
    objectGenerator,
    logger,
    projectResourcesUtils,
    operations,
    settings,
    storageUtils,
    Utils,
    departmentUtils,
    flowRootState,
    tags
} from '../../../../../../../modules/helpers';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { DatavolumeRequestType } from '../../../../../../../modules/api/apis/datavolumesApi/generated/model/datavolumeRequestType';
import { ScopeType } from '../../../../../../../modules/api/apis/datavolumesApi/generated/model/scopeType';
import { V1PersistentVolume } from '@kubernetes/client-node/';
import { V1PersistentVolumeClaim } from '@kubernetes/client-node';
import { datavolumesService } from '../../../../../../../modules/api';

test.describe(
    `Datavolumes - positive tests ${tags.teams.viper}`,
    { tag: [tags.services.datavolume, tags.runOn.serialCI] },
    () => {
        shouldRunTests(settings.isSelfHosted)(function () {
            let proj1: Project;
            let proj2: Project;
            let defaultDepartmentId: number;

            test.beforeAll(async function () {
                proj1 = await projectResourcesUtils.createProjectAndNamespace('p-1');
                proj2 = await projectResourcesUtils.createProjectAndNamespace('p-2');

                defaultDepartmentId = (
                    await departmentUtils.getDefaultDepartmentByCluster(flowRootState.cluster.uuid)
                ).id;

                datavolumesService.datavolumesApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                await operations.deleteProjectAndNamespace(proj1.id);
                await operations.deleteProjectAndNamespace(proj2.id);
            });

            test('should create a datavolume successfully', async function () {
                const dvName = objectGenerator.generateNameByPrefix('dv');

                logger.info(`Creating the datavolume '${dvName}'`);

                const datavolume = await datavolumeUtils.createDatavolume({
                    name: dvName,
                    description: 'datavolume description',
                    originPvcName: '',
                    projectId: String(proj1.id)
                });

                expect(datavolume.name).toBe(dvName);

                // Clean up
                await datavolumeUtils.deleteDatavolume(datavolume.id);
            });

            test.describe('Datavolume Sharing Flow', () => {
                let dvName: string;
                let pvOfDatavolume: any;
                let pvcOfDatavolume: any;
                let testDatavolume: any;
                let projIdToShareWith: string;
                let namespaceOfProjToShareWith: string;
                let storageClassName: string;
                let existingPvToUseInTest: V1PersistentVolume;
                let existingPvcToUseInTest: V1PersistentVolumeClaim;
                const existingPvcNameToUse = 'data-runai-backend-postgresql-0';
                const existingPvcNamespaceToUse = 'runai-backend';

                test.beforeAll(async function () {
                    existingPvcToUseInTest = await storageUtils.getPvc(
                        existingPvcNameToUse,
                        existingPvcNamespaceToUse
                    );
                    logger.info(`Using PVC ${existingPvcNamespaceToUse}/${existingPvcNameToUse} for tests`);
                    storageClassName = existingPvcToUseInTest.spec.storageClassName;

                    logger.info(`Using storage class ${storageClassName} for tests`);
                    const existingPvToUseName = existingPvcToUseInTest.spec.volumeName;
                    existingPvToUseInTest = await storageUtils.getPv(existingPvToUseName);
                });

                test('should handle full lifecycle of sharing a datavolume', async () => {
                    await test.step('Create PV and PVC for the datavolume', async () => {
                        dvName = objectGenerator.generateNameByPrefix('dv');
                        pvOfDatavolume = storageUtils.buildNewPvObjFromPv(
                            dvName,
                            storageClassName,
                            existingPvToUseInTest
                        );
                        pvcOfDatavolume = storageUtils.buildPvcObjFromPv(
                            pvOfDatavolume,
                            dvName,
                            proj1.namespace
                        );

                        pvOfDatavolume = await storageUtils.createPv(pvOfDatavolume);
                        pvcOfDatavolume = await storageUtils.createPvc(pvcOfDatavolume);

                        await storageUtils.waitForPvcToBeBoundToPv(
                            pvcOfDatavolume.metadata.name,
                            pvcOfDatavolume.metadata.namespace,
                            pvOfDatavolume.metadata.name
                        );
                    });

                    await test.step('Create a datavolume based on the PVC', async () => {
                        testDatavolume = await datavolumeUtils.createDatavolume({
                            name: dvName,
                            description: 'datavolume description',
                            originPvcName: pvcOfDatavolume.metadata.name,
                            projectId: String(proj1.id)
                        });

                        await datavolumeUtils.waitForCrdDatavolumeToBeReady(testDatavolume.name);
                        await datavolumeUtils.waitForApiDatavolumeToBeReady(testDatavolume.id);

                        const dvsOriginated = await datavolumeUtils.getDatavolumes(
                            DatavolumeRequestType.Originated
                        );

                        // Check that at least one DataVolume exists
                        expect(dvsOriginated.length).toBeGreaterThan(0);

                        // Assert that the expected DataVolume exists
                        const originatedDvNames = dvsOriginated.map((dv) => dv.name);
                        expect(originatedDvNames).toContain(dvName);

                        await storageUtils.waitForPvcAndPvOfDatavolumeForNamespace(
                            pvcOfDatavolume.metadata.name,
                            proj1.namespace,
                            true
                        );
                    });

                    await test.step('Patch datavolume description', async () => {
                        await datavolumeUtils.patchDatavolume(testDatavolume.id, {
                            description: 'new description'
                        });
                        const updated = await datavolumeUtils.getDatavolume(testDatavolume.id);
                        expect(updated.description).toBe('new description');
                    });

                    await test.step('Share datavolume with another project', async () => {
                        projIdToShareWith = String(proj2.id);
                        namespaceOfProjToShareWith = proj2.namespace;

                        await datavolumeUtils.patchDatavolumeSharedScopes(testDatavolume.id, {
                            add: [{ scopeId: projIdToShareWith, scopeType: ScopeType.Project }]
                        });

                        const apiDv = await datavolumeUtils.getDatavolume(testDatavolume.id);
                        expect(apiDv.sharedScopes).toHaveLength(1);
                        expect(apiDv.sharedScopes[0].scopeId).toBe(projIdToShareWith);

                        const sharedScopes = await datavolumeUtils.getDatavolumeSharedScopes(
                            testDatavolume.id
                        );
                        expect(sharedScopes).toHaveLength(1);
                        expect(sharedScopes[0].scopeId).toBe(projIdToShareWith);

                        const usableDvs = await datavolumeUtils.getDatavolumes(
                            DatavolumeRequestType.UsableInProject,
                            projIdToShareWith
                        );
                        expect(usableDvs).toHaveLength(1);

                        const pvcInNs = await storageUtils.waitForPvcAndPvOfDatavolumeForNamespace(
                            testDatavolume.name,
                            namespaceOfProjToShareWith
                        );
                        await storageUtils.waitForPvcToBeBound(
                            pvcInNs.metadata.name,
                            namespaceOfProjToShareWith
                        );

                        const dvCrd = await datavolumeUtils.waitForDatavolumeCrd(testDatavolume.name);
                        expect(dvCrd['metadata'].name).toBe(apiDv.name);
                        expect(dvCrd['spec'].datavolumeId).toBe(apiDv.id);
                        expect(dvCrd['spec'].description).toBe(apiDv.description);
                        expect(dvCrd['spec'].originPvcName).toBe(apiDv.originPvcName);
                        expect(dvCrd['spec'].shouldDeleteOriginalVolume).toBe(
                            apiDv.shouldDeleteOriginalVolume
                        );
                        expect(dvCrd['spec'].sharedScopes).toHaveLength(1);
                        expect(dvCrd['spec'].sharedScopes[0].scopeName).toBe(proj2.name);
                        expect(dvCrd['spec'].sharedScopes[0].scopeType).toBe(ScopeType.Project);
                    });

                    await test.step('Un-share the datavolume', async () => {
                        await datavolumeUtils.patchDatavolumeSharedScopes(testDatavolume.id, {
                            remove: [{ scopeId: projIdToShareWith, scopeType: ScopeType.Project }]
                        });

                        const apiDv = await datavolumeUtils.getDatavolume(testDatavolume.id);
                        expect(apiDv.sharedScopes || []).toHaveLength(0);

                        const sharedScopes = await datavolumeUtils.getDatavolumeSharedScopes(
                            testDatavolume.id
                        );
                        expect(sharedScopes).toHaveLength(0);
                    });

                    await test.step('Share datavolume with a department', async () => {
                        await datavolumeUtils.patchDatavolumeSharedScopes(testDatavolume.id, {
                            add: [{ scopeId: String(defaultDepartmentId), scopeType: ScopeType.Department }]
                        });

                        const apiDv = await datavolumeUtils.getDatavolume(testDatavolume.id);
                        expect(apiDv.sharedScopes).toHaveLength(1);
                        expect(apiDv.sharedScopes[0].scopeId).toBe(String(defaultDepartmentId));

                        const deptNamespace = Utils.namespaceFromProjectName(proj2.name);
                        const pvcInNs = await storageUtils.waitForPvcAndPvOfDatavolumeForNamespace(
                            testDatavolume.name,
                            deptNamespace
                        );
                        await storageUtils.waitForPvcToBeBound(pvcInNs.metadata.name, deptNamespace);
                    });
                });
            });
        });
    }
);
