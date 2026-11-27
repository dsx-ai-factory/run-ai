/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    tags,
    flowRootState,
    objectGenerator,
    orgUnitUtils,
    nodePoolUtils,
    waitUtils,
    auditLogsUtils,
    crds
} from '../../../../../../modules/helpers';

test.describe(
    `Audit api tests, tags: ${tags.resource.audit} ${tags.teams.phoenix}`,
    { tag: [tags.services.auditService] },
    () => {
        let projectName: string;
        let projectId: string;
        let defaultNodePoolId: string;

        test.beforeAll(async () => {
            projectName = objectGenerator.generateNameByPrefix('p');
            defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);

            const res = await orgUnitUtils.createProjectAndValidate({
                schedulingRules: {
                    interactiveJobTimeLimitSeconds: 2,
                    interactiveJobMaxIdleDurationSeconds: 3,
                    interactiveJobPreemptIdleDurationSeconds: 4,
                    trainingJobMaxIdleDurationSeconds: 5,
                    trainingJobTimeLimitSeconds: 6
                },
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 1 }
                    }
                ],
                name: projectName,
                clusterId: flowRootState.cluster.uuid
            });
            projectId = res.id;
        });

        test.afterAll(async () => {
            await orgUnitUtils.deleteProject(projectId);
            await waitUtils.waitForCrd(crds.queueCrd, projectName, undefined, 60, false);
        });

        test('Get audit logs- with entity_name filter', async () => {
            const e = new Date(new Date().getTime() + 1000 * 60);
            const s = new Date(new Date().getTime() - 1000 * 60 * 10);
            const filters = [`entity_name==${projectName}`];
            const { accessToken } = flowRootState.users.system_administrator;
            await waitUtils.waitForCondition(
                async () =>
                    (await auditLogsUtils.getAuditLogs(accessToken, s, e, filters)).body.auditLogs !==
                    undefined,
                'waiting for audit logs to be available',
                true,
                60,
                5
            );

            const {
                body: { auditLogs }
            } = await auditLogsUtils.getAuditLogs(accessToken, s, e, filters);
            expect(auditLogs.length).toBeGreaterThanOrEqual(1);
        });

        test('Download audit logs- with entity_name filter', async () => {
            const e = new Date(new Date().getTime() + 1000 * 60);
            const s = new Date(new Date().getTime() - 1000 * 60 * 10);
            const filters = [`entity_name==${projectName}`];
            const { accessToken } = flowRootState.users.system_administrator;
            await waitUtils.waitForCondition(
                async () =>
                    (await auditLogsUtils.downloadAuditLogs(accessToken, s, e, 'JSON', undefined, filters))
                        .body !== undefined,
                'waiting for audit logs to be available',
                true,
                60,
                5
            );

            const { body, response } = await auditLogsUtils.downloadAuditLogs(
                accessToken,
                s,
                e,
                'JSON',
                undefined,
                filters
            );
            expect(response.headers['content-disposition']).toContain('attachment');
            expect(body.length).toBeGreaterThanOrEqual(1);
        });
    }
);
