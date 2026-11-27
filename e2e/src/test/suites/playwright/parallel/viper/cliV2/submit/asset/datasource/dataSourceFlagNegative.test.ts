/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { assert } from 'chai';
import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { PVCCreationRequest, Scope } from '../../../../../../../../../modules/api/apis/assetsApi';

test.describe(`CLI V2 Submit with Datasource Asset Negative Tests, ${tags.teams.viper}, ${tags.runOn.parallelCI}`, () => {
    let project: Project;
    let otherProject: Project;
    let otherProjectPvcId: string;
    let otherProjectPvcName: string;
    let workloadName: string;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-ds-neg');

        // Create another project with a PVC datasource scoped to it
        otherProject = await projectResourcesUtils.createProjectAndNamespace('cli-ds-other');
        otherProjectPvcName = objectGenerator.generateNameByPrefix('pvc-other');

        const pvcRequest: PVCCreationRequest = {
            meta: {
                name: otherProjectPvcName,
                scope: Scope.Project,
                projectId: otherProject.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: '/mnt/data',
                claimName: otherProjectPvcName,
                existingPvc: false,
                claimInfo: {
                    size: '1Gi',
                    accessModes: {
                        readWriteOnce: true
                    },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };

        const pvcAsset = await dataSourcesUtils.createPvcAsset(pvcRequest);
        otherProjectPvcId = pvcAsset.meta.id;

        // Wait for the actual PVC to be created in Kubernetes
        await waitUtils.waitForPvc(otherProjectPvcName, otherProject.namespace);
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl-neg');
    });

    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            await operations.deleteProjectAndNamespace(otherProject.id);
            if (otherProjectPvcId) {
                await dataSourcesUtils.deletePvcAsset(otherProjectPvcId);
            }
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    test('fail when datasource does not exist', () => {
        const dsName = 'non-existent-ds-12345';
        const nonExistentDatasource = [`type=pvc,name=${dsName}`];

        const submitFlags = {
            [SubmitFlagsMerged.datasource]: nonExistentDatasource,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            `failed to submit. datasource '${dsName}' of type 'pvc' not found`
        );
    });

    test('fail when datasource is scoped to a different project', () => {
        const submitFlags = {
            [SubmitFlagsMerged.datasource]: [`type=pvc,name=${otherProjectPvcName}`],
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            `failed to submit. datasource '${otherProjectPvcName}' of type 'pvc' not found`
        );
    });

    test('fail when datasource format is missing type', () => {
        const invalidDatasource = ['name=some-datasource'];

        const submitFlags = {
            [SubmitFlagsMerged.datasource]: invalidDatasource,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            // prettier-ignore
            'failed to submit. invalid datasource format \'name=some-datasource\': required format is type=TYPE,name=NAME'
        );
    });

    test('fail when datasource format is missing name', () => {
        const invalidDatasource = ['type=pvc'];

        const submitFlags = {
            [SubmitFlagsMerged.datasource]: invalidDatasource,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            // prettier-ignore
            'failed to submit. invalid datasource format \'type=pvc\': required format is type=TYPE,name=NAME'
        );
    });

    test('fail when datasource type is invalid', () => {
        const invalidDatasource = ['type=invalid_type,name=some-ds'];

        const submitFlags = {
            [SubmitFlagsMerged.datasource]: invalidDatasource,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            // prettier-ignore
            'failed to submit. datasource \'some-ds\' of type \'invalid_type\' not found'
        );
    });
});
