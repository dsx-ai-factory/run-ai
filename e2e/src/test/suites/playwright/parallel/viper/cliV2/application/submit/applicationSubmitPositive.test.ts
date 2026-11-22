/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import fs from 'fs';
import os from 'os';
import path from 'path';

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
import { applicationSubmitFlags } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';

const CHART_URL = 'https://charts.bitnami.com/bitnami/nginx-18.1.11.tgz';

test.describe(`CLI V2 AI Application Submit Positive Tests, ${tags.teams.viper},`, () => {
    let project: Project;
    const createdAppIds: string[] = [];

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-app-submit');
    });

    test.afterAll(async () => {
        for (const id of createdAppIds) {
            try {
                await aiApplicationUtils.deleteAiApplication(id);
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

    test('application submit creates the application and prints success message', async () => {
        const appName = objectGenerator.generateNameByPrefix('app-submit');

        const output = runaiCliV2.applicationSubmit(appName, {
            [applicationSubmitFlags.project]: project.name,
            [applicationSubmitFlags.chartUrl]: CHART_URL
        });

        expect(output).toContain(`Creating application ${appName}`);
        expect(output).toContain(`runai application describe ${appName} -p ${project.name}`);

        const listed = await aiApplicationUtils.listAiApplications();
        const created = listed.body.applications?.find((a) => a.name === appName);
        expect(created).toBeDefined();
        if (created?.id) createdAppIds.push(created.id);
    });

    test('application submit with --values applies values from file', async () => {
        const appName = objectGenerator.generateNameByPrefix('app-submit-vals');
        const valuesFile = path.join(os.tmpdir(), `runai-test-values-${appName}.yaml`);
        fs.writeFileSync(valuesFile, 'replicaCount: 1\n');

        try {
            const output = runaiCliV2.applicationSubmit(appName, {
                [applicationSubmitFlags.project]: project.name,
                [applicationSubmitFlags.chartUrl]: CHART_URL,
                [applicationSubmitFlags.values]: valuesFile
            });

            expect(output).toContain(`Creating application ${appName}`);

            const listed = await aiApplicationUtils.listAiApplications();
            const created = listed.body.applications?.find((a) => a.name === appName);
            expect(created).toBeDefined();
            if (created?.id) createdAppIds.push(created.id);
        } finally {
            fs.unlinkSync(valuesFile);
        }
    });

});
