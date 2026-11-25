/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    aiApplicationUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    projectResourcesUtils,
    tags
} from '../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import { listFlags } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { AIApplication } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/aIApplication';

const COLUMNS = 'name,status';
const TABLE_HEADER = 'AI Application';
const COLUMN_STATUS = 'Status';
const COLUMN_GPU = 'GPU Compute Req';

const CHART_REF = {
    repo: 'https://charts.bitnami.com/bitnami',
    chartName: 'nginx',
    chartVersion: '18.1.11'
};

test.describe(`CLI V2 AI Application List Positive Tests, ${tags.teams.viper},`, () => {
    let projectA: Project;
    let projectB: Project;
    let appOne: AIApplication;
    let appTwo: AIApplication;
    let appInProjectB: AIApplication;

    const createAiApplication = (projectId: number): ReturnType<typeof aiApplicationUtils.createAiApplication> =>
        aiApplicationUtils.createAiApplication({
            clusterId: flowRootState.cluster.uuid,
            projectId: String(projectId),
            name: objectGenerator.generateNameByPrefix('app-list'),
            spec: { chartRef: CHART_REF }
        });

    test.beforeAll(async () => {
        [projectA, projectB] = await Promise.all([
            projectResourcesUtils.createProjectAndNamespace('cli-app-list-a'),
            projectResourcesUtils.createProjectAndNamespace('cli-app-list-b')
        ]);

        const aiApplications = await Promise.all([
            createAiApplication(projectA.id),
            createAiApplication(projectA.id),
            createAiApplication(projectB.id)
        ]);
        appOne = aiApplications[0].body;
        appTwo = aiApplications[1].body;
        appInProjectB = aiApplications[2].body;
    });

    test.afterAll(async () => {
        for (const app of [appOne, appTwo, appInProjectB]) {
            try {
                if (app?.id) await aiApplicationUtils.deleteAiApplication(app.id);
            } catch (error: any) {
                if (error?.statusCode !== 404) {
                    logger.error('Application cleanup failed:', error);
                }
            }
        }

        for (const proj of [projectA, projectB]) {
            try {
                if (proj) await orgUnitUtils.deleteProjectAndNamespace(proj.id.toString(), proj.name);
            } catch (error) {
                logger.error('Project cleanup failed:', error);
            }
        }
    });

    test('application list scoped to a project returns only its applications', () => {
        const listOutput = runaiCliV2.list('application', {
            [listFlags.project]: projectA.name
        } as Partial<typeof listFlags>).stdout;

        expect(listOutput).toContain(TABLE_HEADER);
        expect(listOutput).toContain(appOne.name);
        expect(listOutput).toContain(appTwo.name);
        expect(listOutput).not.toContain(appInProjectB.name);
    });

    test('application list with -A returns applications across projects', () => {
        const listOutput = runaiCliV2.list('application', {
            [listFlags.allScopes]: ''
        } as Partial<typeof listFlags>).stdout;

        expect(listOutput).toContain(appOne.name);
        expect(listOutput).toContain(appTwo.name);
        expect(listOutput).toContain(appInProjectB.name);
    });

    test('application list with --json outputs parseable JSON containing the created apps', () => {
        const listOutput = runaiCliV2.list('application', {
            [listFlags.project]: projectA.name,
            [listFlags.json]: ''
        } as Partial<typeof listFlags>).stdout;

        const parsed = JSON.parse(listOutput);
        const found = parsed.aiApplications?.find((a: any) => a.name === appOne.name);
        expect(found).toBeDefined();
        expect(found.projectName).toBe(projectA.name);
    });

    test('application list with --columns limits the table columns', () => {
        const listOutput = runaiCliV2.list('application', {
            [listFlags.project]: projectA.name,
            [listFlags.columns]: COLUMNS
        } as Partial<typeof listFlags>).stdout;

        expect(listOutput).toContain(TABLE_HEADER);
        expect(listOutput).toContain(COLUMN_STATUS);
        expect(listOutput).not.toContain(COLUMN_GPU);
        expect(listOutput).toContain(appOne.name);
    });

});
