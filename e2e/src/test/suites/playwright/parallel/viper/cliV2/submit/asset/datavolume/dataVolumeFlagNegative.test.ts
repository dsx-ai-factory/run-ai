/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { assert } from 'chai';

import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    datavolumeUtils,
    dataSourcesUtils,
    flowRootState,
    objectGenerator,
    projectResourcesUtils,
    tags,
    waitUtils,
    orgUnitUtils
} from '../../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { PVCCreationRequest, Scope } from '../../../../../../../../../modules/api/apis/assetsApi';

test.describe(`CLI V2 Submit with DataVolume Flag - Negative Tests, ${tags.teams.viper},`, { tag: [tags.services.datavolume, tags.runOn.parallelCI] }, () => {
    let project1: Project;
    let project2: Project;
    let workloadName: string;

    let pvcProject2Name: string;
    let pvcProject2Id: string;

    let dvProject2Id: string;
    let dvProject2Name: string;

    test.beforeAll(async () => {
        project1 = await projectResourcesUtils.createProjectAndNamespace('cli-dv-neg1');
        project2 = await projectResourcesUtils.createProjectAndNamespace('cli-dv-neg2');

        pvcProject2Name = objectGenerator.generateNameByPrefix('pvc-neg');
        const pvcProject2Request: PVCCreationRequest = {
            meta: {
                name: pvcProject2Name,
                scope: Scope.Project,
                projectId: project2.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: '/mnt/dv-neg',
                claimName: pvcProject2Name,
                existingPvc: false,
                claimInfo: {
                    size: '1Gi',
                    accessModes: { readWriteOnce: true },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };
        const pvcProject2Asset = await dataSourcesUtils.createPvcAsset(pvcProject2Request);
        pvcProject2Id = pvcProject2Asset.meta.id;
        await waitUtils.waitForPvc(pvcProject2Name, project2.namespace);

        dvProject2Name = objectGenerator.generateNameByPrefix('dv-neg');
        const dvProject2 = await datavolumeUtils.createDatavolume({
            name: dvProject2Name,
            description: 'Project2-scoped datavolume for CLI negative test',
            originPvcName: pvcProject2Name,
            projectId: String(project2.id)
        });
        dvProject2Id = dvProject2.id;
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl-neg');
    });

    test.afterAll(async () => {
        try {
            if (dvProject2Id) {
                await datavolumeUtils.deleteDatavolume(dvProject2Id);
                await datavolumeUtils.waitForApiDatavolumeToBeDeleted(dvProject2Id);
            }
        } catch (error: any) {
            if (error?.statusCode !== 404) {
                console.error('Datavolume cleanup failed:', error);
            }
        }

        try {
            if (pvcProject2Id) await dataSourcesUtils.deletePvcAsset(pvcProject2Id);
        } catch (error: any) {
            if (error?.statusCode !== 404) {
                console.error('PVC cleanup failed:', error);
            }
        }

        try {
            await orgUnitUtils.deleteProjectAndNamespace(project1.id.toString(), project1.name);
            await orgUnitUtils.deleteProjectAndNamespace(project2.id.toString(), project2.name);
        } catch (error) {
            console.error('Project cleanup failed:', error);
        }
    });

    test('fail when datavolume does not exist', () => {
        const nonExistentDv = 'non-existent-dv-12345';

        const submitFlags = {
            [SubmitFlagsMerged.datavolume]: [`name=${nonExistentDv},mountpath=/mnt/data`],
            [SubmitFlagsMerged.project]: project1.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(() => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training));
    });

    test('fail when datavolume is scoped to a different project', () => {
        const submitFlags = {
            [SubmitFlagsMerged.datavolume]: [`name=${dvProject2Name},mountpath=/mnt/data`],
            [SubmitFlagsMerged.project]: project1.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(() => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training));
    });

    test('fail when datavolume format is missing mountpath', () => {
        const submitFlags = {
            [SubmitFlagsMerged.datavolume]: ['name=some-datavolume'],
            [SubmitFlagsMerged.project]: project1.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(() => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training));
    });

    test('fail when datavolume format is missing name', () => {
        const submitFlags = {
            [SubmitFlagsMerged.datavolume]: ['mountpath=/mnt/data'],
            [SubmitFlagsMerged.project]: project1.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(() => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training));
    });

    test('fail when datavolume name is empty', () => {
        const submitFlags = {
            [SubmitFlagsMerged.datavolume]: ['name=,mountpath=/mnt/data'],
            [SubmitFlagsMerged.project]: project1.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(() => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training));
    });

    test('fail when datavolume mountpath is empty', () => {
        const submitFlags = {
            [SubmitFlagsMerged.datavolume]: ['name=some-dv,mountpath='],
            [SubmitFlagsMerged.project]: project1.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(() => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training));
    });
});
