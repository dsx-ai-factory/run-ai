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
    tags,
    waitUtils
} from '../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import {
    cliV2workloadNoun,
    DeleteFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { AIApplication } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/aIApplication';

const CHART_REF = {
    repo: 'https://charts.bitnami.com/bitnami',
    chartName: 'nginx',
    chartVersion: '18.1.11'
};

const applicationNoun = 'application' as unknown as typeof cliV2workloadNoun;

test.describe(`CLI V2 AI Application Delete Positive Tests, ${tags.teams.viper},`, () => {
    let project: Project;
    let appToDelete: AIApplication;
    let appToForceDelete: AIApplication;

    const createAiApplication = (): ReturnType<typeof aiApplicationUtils.createAiApplication> =>
        aiApplicationUtils.createAiApplication({
            clusterId: flowRootState.cluster.uuid,
            projectId: String(project.id),
            name: objectGenerator.generateNameByPrefix('app-delete'),
            spec: { chartRef: CHART_REF }
        });

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-app-delete');

        const aiApplications = await Promise.all([createAiApplication(), createAiApplication()]);
        appToDelete = aiApplications[0].body;
        appToForceDelete = aiApplications[1].body;

        for (const app of [appToDelete, appToForceDelete]) {
            await waitUtils.waitForCondition(
                async () => (await aiApplicationUtils.getAiApplication(app.id)).body.actionStatus !== 'creating',
                `application ${app.name} left the creating phase`
            );
        }
    });

    test.afterAll(async () => {
        for (const app of [appToDelete, appToForceDelete]) {
            try {
                if (app?.id) await aiApplicationUtils.deleteAiApplication(app.id);
            } catch (error: any) {
                if (error?.statusCode !== 404) {
                    logger.error('Application cleanup failed:', error);
                }
            }
        }

        try {
            await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), project.name);
        } catch (error) {
            logger.error('Project cleanup failed:', error);
        }
    });

    test('application delete with -y deletes the application', async () => {
        const output = runaiCliV2.delete(
            {
                [DeleteFlagsMerged.project]: project.name,
                [DeleteFlagsMerged.byPassDialog]: ''
            } as Partial<typeof DeleteFlagsMerged>,
            appToDelete.name,
            applicationNoun
        );

        expect(output).toContain(`Application "${appToDelete.name}" deleted successfully.`);

        const app = await aiApplicationUtils.getAiApplication(appToDelete.id);
        expect(app.body.desiredPhase).toBe('deleted');
    });

    test('application delete with --force deletes the application', () => {
        const output = runaiCliV2.delete(
            {
                [DeleteFlagsMerged.project]: project.name,
                ['--force']: ''
            } as Partial<typeof DeleteFlagsMerged>,
            appToForceDelete.name,
            applicationNoun
        );

        expect(output).toContain(`Application "${appToForceDelete.name}" deleted successfully.`);
    });

});
