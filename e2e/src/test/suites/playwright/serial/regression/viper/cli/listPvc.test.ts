/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    dataSourcesUtils,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import {
    cliLoginAsSystemAdmin,
    listFlags
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Scope } from '../../../../../../../modules/api/apis/assetsApi';
import { test, expect } from '@playwright/test';
import * as yaml from 'js-yaml';

function findLineContaining(target: string, text: string): string {
    const lines = text.split('\n');
    const foundLine = lines.find((line) => line.includes(target));

    if (!foundLine) {
        throw new Error(`No line contains the string: "${target}"`);
    }

    return foundLine;
}

const checkList = (pvcList: any, pvcUnManaged: string, pvcManaged: string): void => {
    let managedOk = false;
    let unmanagedOk = false;
    for (const pvc of pvcList.pvcs) {
        if (pvc.name === pvcManaged && pvc.managed === true) {
            managedOk = true;
        }
        if (pvc.name.includes(pvcUnManaged) && pvc.managed === false) {
            unmanagedOk = true;
        }
    }
    expect(managedOk, 'manged pvc wasnt found, or contained invalid params').toBe(true);
    expect(unmanagedOk, 'unmanged pvc wasnt found, or contained invalid params').toBe(true);
};

test.describe(`Validate PVC list   ${tags.teams.viper}`, function () {
    let project: Project;
    const pvcManaged = objectGenerator.generateShortNameByPrefix('pvc-managed');
    const pvcUnManaged = objectGenerator.generateShortNameByPrefix('pvc-unmanaged');

    async function generatePvc(): Promise<void> {
        const pvcPayload = dataSourcesUtils.getPvcCreationPayload({
            name: pvcManaged,
            scopeId: project.id.toString(),
            scope: Scope.Project
        });
        await dataSourcesUtils.createClusterPvc(pvcUnManaged, project.namespace);
        await dataSourcesUtils.createPvcAsset(pvcPayload);
    }

    test.beforeAll(async function () {
        console.log('starting PVC list tests');
        cliLoginAsSystemAdmin();
        project = await projectResourcesUtils.createProjectAndNamespace('cli-pvc');
        runaiCliV2.projectSet(project.name);
        await generatePvc();
        await waitUtils.waitForPvc(pvcManaged, project.namespace, 20, true);
    });

    test.afterAll(async function () {
        await operations.deleteProjectAndNamespace(project.id);
        console.log('ending PVC list tests');
    });

    test('should show PVC - Table', () => {
        let lookup: string;
        const listFlagsParams = {} as Partial<typeof listFlags>;
        const result = runaiCliV2.list('pvc', listFlagsParams);
        expect(result.status).toEqual(0);
        console.log(result.stdout);
        lookup = findLineContaining(pvcManaged, result.stdout);
        expect(lookup).toContain(' Yes');
        lookup = findLineContaining(pvcUnManaged, result.stdout);
        expect(lookup).toContain(' No');
    });

    test('should show PVC - JSON', () => {
        const listFlagsParams = {
            [listFlags.json]: ''
        } as Partial<typeof listFlags>;
        const result = runaiCliV2.list('pvc', listFlagsParams);
        expect(result.status).toEqual(0);
        console.log(result.stdout);
        const pvcListJson = JSON.parse(result.stdout);
        checkList(pvcListJson, pvcUnManaged, pvcManaged);
    });

    test('should show PVC - YAML', () => {
        const listFlagsParams = {
            [listFlags.yaml]: ''
        } as Partial<typeof listFlags>;
        const result = runaiCliV2.list('pvc', listFlagsParams);
        expect(result.status).toEqual(0);
        console.log(result.stdout);
        const pvcListYaml = yaml.load(result.stdout);
        checkList(pvcListYaml, pvcUnManaged, pvcManaged);
    });
});
