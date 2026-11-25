/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    storageUtils,
    tags
} from '../../../../../../modules/helpers';
import { expect } from 'chai';
import { V1StorageClass } from '@kubernetes/client-node/';
import { test } from '@playwright/test';

test.describe(`StorageClass API and sync, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`, { tag: [tags.services.k8sObjectsTracker, tags.runOn.parallelCI] }, function () {
    let projectA: Project;
    let storageClass: V1StorageClass;
    const RunaiNamespace = 'runai';

    let clusterId: string;

    test.beforeAll(async () => {
        clusterId = flowRootState.cluster.uuid;
        projectA = await projectResourcesUtils.createProjectAndNamespace('storage-class');
    });
    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(projectA.id);
        } catch (error) {
            logger.error(`failed to delete project "storage-class" and namespace, error: ${error}`);
        }
        if (storageClass !== undefined) {
            await storageUtils.deleteStorageClass(storageClass.metadata.name);
        }
    });

    test('create / delete a storage classes and verify sync', async function () {
        // instructing k8s to create the storage class
        storageClass = await storageUtils.createStorageClass(
            objectGenerator.generateNameByPrefix('storage-class')
        );

        // wait for the storage class to be synced to control-plane
        const sc = await dataSourcesUtils.waitForStorageClassInfo(storageClass.metadata.name, clusterId);
        expect(sc, 'expected SC to be synced to objects-tracker immediately').not.undefined;
        logger.info(`verified storage class: ${sc.storageClassName} synced to control-plane successfully`);

        // instructing k8s to delete the storage class
        await storageUtils.deleteStorageClass(storageClass.metadata.name);

        // wait for the storage class to be deleted from control-plane
        const deleted = await dataSourcesUtils.waitForStorageClassInfo(
            storageClass.metadata.name,
            clusterId,
            true
        );
        expect(deleted, 'expected SC to be deleted from objects-tracker').undefined;

        // verify the storage class is deleted from control-plane
        const res = await dataSourcesUtils.getStorageClassInfo(storageClass.metadata.name, clusterId);
        expect(res, 'expected SC to be deleted from objects-tracker').to.be.empty;

        logger.info(
            `verified storage class: ${storageClass.metadata.name} deleted from control-plane successfully`
        );

        storageClass = undefined;
    });
});
