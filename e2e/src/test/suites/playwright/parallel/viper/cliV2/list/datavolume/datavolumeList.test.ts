/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    datavolumeUtils,
    dataSourcesUtils,
    flowRootState,
    objectGenerator,
    projectResourcesUtils,
    tags,
    waitUtils,
    orgUnitUtils
} from '../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import { listFlags } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { PVCCreationRequest, Scope } from '../../../../../../../../modules/api/apis/assetsApi';

test.describe(`CLI V2 Datavolume List Tests, ${tags.teams.viper},`, { tag: [tags.services.datavolume, tags.runOn.parallelCI] }, () => {
    let project1: Project;

    let pvcProjectName: string;
    let pvcProjectId: string;

    let dvProjectId: string;
    let dvProjectName: string;

    test.beforeAll(async () => {
        project1 = await projectResourcesUtils.createProjectAndNamespace('cli-dv-list');

        pvcProjectName = objectGenerator.generateNameByPrefix('pvc-proj');

        const pvcProjectRequest: PVCCreationRequest = {
            meta: {
                name: pvcProjectName,
                scope: Scope.Project,
                projectId: project1.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: '/mnt/dv-project',
                claimName: pvcProjectName,
                existingPvc: false,
                claimInfo: {
                    size: '1Gi',
                    accessModes: { readWriteOnce: true },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };
        const pvcProjectAsset = await dataSourcesUtils.createPvcAsset(pvcProjectRequest);
        pvcProjectId = pvcProjectAsset.meta.id;
        await waitUtils.waitForPvc(pvcProjectName, project1.namespace);

        dvProjectName = objectGenerator.generateNameByPrefix('dv-proj');
        const dvProject = await datavolumeUtils.createDatavolume({
            name: dvProjectName,
            description: 'Project-scoped datavolume for CLI list test',
            originPvcName: pvcProjectName,
            projectId: String(project1.id)
        });
        dvProjectId = dvProject.id;

        await datavolumeUtils.waitForApiDatavolumeToExist(dvProjectId);
    });

    test.afterAll(async () => {
        try {
            if (dvProjectId) {
                await datavolumeUtils.deleteDatavolume(dvProjectId);
                await datavolumeUtils.waitForApiDatavolumeToBeDeleted(dvProjectId);
            }
        } catch (error: any) {
            if (error?.statusCode !== 404) {
                console.error('Datavolume cleanup failed:', error);
            }
        }

        try {
            if (pvcProjectId) await dataSourcesUtils.deletePvcAsset(pvcProjectId);
        } catch (error: any) {
            if (error?.statusCode !== 404) {
                console.error('PVC cleanup failed:', error);
            }
        }

        try {
            await orgUnitUtils.deleteProjectAndNamespace(project1.id.toString(), project1.name);
        } catch (error) {
            console.error('Project cleanup failed:', error);
        }
    });

    test('CLI datavolume list returns project-scoped datavolume', () => {
        const listResult = runaiCliV2.list('datavolume', {
            [listFlags.project]: project1.name,
            [listFlags.json]: ''
        } as Partial<typeof listFlags>);

        const listOutput = listResult.stdout;
        expect(listOutput).toContain(dvProjectName);

        const jsonOutput = JSON.parse(listOutput);
        expect(jsonOutput.datavolumes.length).toBeGreaterThanOrEqual(1);

        const projectDv = jsonOutput.datavolumes.find((dv: any) => dv.name === dvProjectName);
        expect(projectDv).toBeDefined();
    });

    test('CLI datavolume list with --all flag returns datavolumes', () => {
        const listResult = runaiCliV2.list('datavolume', {
            [listFlags.allScopes]: ''
        } as Partial<typeof listFlags>);

        const listOutput = listResult.stdout;
        expect(listOutput).toContain(dvProjectName);
    });
});
