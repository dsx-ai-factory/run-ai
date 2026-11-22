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
import { describeFlags } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { AIApplication } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/aIApplication';

const CHART_REF = {
    repo: 'https://charts.bitnami.com/bitnami',
    chartName: 'nginx',
    chartVersion: '18.1.11'
};

test.describe(`CLI V2 AI Application Describe Positive Tests, ${tags.teams.viper},`, () => {
    let project: Project;
    let app: AIApplication;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-app-describe');

        const created = await aiApplicationUtils.createAiApplication({
            clusterId: flowRootState.cluster.uuid,
            projectId: String(project.id),
            name: objectGenerator.generateNameByPrefix('app-describe'),
            spec: { chartRef: CHART_REF }
        });
        app = created.body;
    });

    test.afterAll(async () => {
        try {
            if (app?.id) await aiApplicationUtils.deleteAiApplication(app.id);
        } catch (error: any) {
            if (error?.statusCode !== 404) {
                logger.error('Application cleanup failed:', error);
            }
        }

        try {
            await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), project.name);
        } catch (error) {
            logger.error('Project cleanup failed:', error);
        }
    });

    test('application describe returns the correct project and app name in JSON output', () => {
        const output = runaiCliV2.applicationDescribe(app.name, {
            [describeFlags.project]: project.name,
            [describeFlags.output]: 'json'
        });

        const parsed = JSON.parse(output);
        expect(parsed.name).toBe(app.name);
        expect(parsed.projectName).toBe(project.name);
    });
});
