/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    nodePoolUtils,
    flowRootState,
    auditLogsUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { NodepoolUpdateFields, PlacementStrategy } from '../../../../../../modules/api/apis/clusterApi';
import { StatusCodes } from 'http-status-codes';
import http from 'http';

test.describe(`Audit log Tests, tags: ${tags.resource.audit} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let defaultNodePoolId: string;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-audit-log', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);
            const updateFields: NodepoolUpdateFields = {
                schedulingConfiguration: {
                    gpuPodsPlacementStrategy: {
                        node: PlacementStrategy.Spread
                    }
                }
            };
            await nodePoolUtils.updateClusterNodePoolAndValidate(defaultNodePoolId, updateFields);
        });

        test.afterAll('deleting project and namespace', async () => {
            const updateFields: NodepoolUpdateFields = {
                schedulingConfiguration: {
                    gpuPodsPlacementStrategy: {
                        node: PlacementStrategy.Binpack
                    }
                }
            };
            await nodePoolUtils.updateClusterNodePoolAndValidate(defaultNodePoolId, updateFields);

            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test('T3738098: test project and node pool operation audit log', async function () {
            const startTime = new Date(new Date().getTime() - 1000 * 60 * 5);

            await test.step('Verify create project operation is saved in audit log', async function () {
                const filters = new Array<string>();
                filters.push('action=@create,entity_type=@project');
                await waitUtils.waitForCondition(
                    async () => {
                        const response = await auditLogsUtils.getAuditLogs(
                            flowRootState.users.system_administrator.accessToken,
                            startTime,
                            new Date(),
                            filters
                        );

                        const auditLogs = response.body.auditLogs;
                        if (!auditLogs) return false;

                        for (const auditLog of auditLogs) {
                            if (auditLog.entityName === projectName) return true;
                        }
                        return false;
                    },
                    'waiting for audit logs to be available',
                    true,
                    60,
                    5
                );
            });

            await test.step('Verify update node pool operation audit log', async () => {
                const filters = new Array<string>();
                filters.push('entity_type=@node pool,action=@update');

                await waitUtils.waitForCondition(
                    async () => {
                        const response = await auditLogsUtils.getAuditLogs(
                            flowRootState.users.system_administrator.accessToken,
                            startTime,
                            new Date(),
                            filters
                        );

                        const auditLogs = response.body.auditLogs;
                        if (!auditLogs) return false;

                        for (const auditLog of auditLogs) {
                            if (auditLog.entityName === 'default') return true;
                        }
                        return false;
                    },
                    'waiting for audit logs to be available',
                    true,
                    60,
                    5
                );
            });

            await test.step('Verify downloading audit log', async () => {
                const numberOfLogs = 5;
                let response: {
                    response: http.IncomingMessage;
                    body?: any;
                };

                logger.info('test downloading json log.');
                response = await auditLogsUtils.downloadAuditLogs(
                    flowRootState.users.system_administrator.accessToken,
                    startTime,
                    new Date(),
                    'JSON',
                    numberOfLogs,
                    undefined
                );
                expect(response.response.statusCode).toEqual(StatusCodes.OK);
                expect(response.body).not.toBeUndefined();
                expect(response.body.length).toEqual(numberOfLogs);

                logger.info('test downloading csv log.');
                response = await auditLogsUtils.downloadAuditLogs(
                    flowRootState.users.system_administrator.accessToken,
                    startTime,
                    new Date(),
                    'CSV',
                    numberOfLogs,
                    undefined
                );
                expect(response.response.statusCode).toEqual(StatusCodes.OK);
                expect(response.body).not.toBeUndefined();

                // including an additional row to store audit log table headers
                expect(
                    response.body.split(/\r?\n/).filter((line: string) => line.trim() !== '').length
                ).toEqual(numberOfLogs + 1);
            });
        });
    });
});
