/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    objectGenerator,
    reportApiUtils,
    settings,
    tags
} from '../../../../../../../modules/helpers';
import { test, expect } from '@playwright/test';
import { SoftAssert } from '../../../../../../../modules/helpers/SoftAssert';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { StatusCodes } from 'http-status-codes';
import { ReportGroupByOptions } from '../../../../../../../modules/api/apis/orgUnitApi';
import { v4 as uuidv4 } from 'uuid';

function parseCSVToDict(csv: string): Record<string, any> {
    const rows = csv.trim().split('\n');
    const reportData = rows.map((row) => row.split(','));

    const headers = reportData[0];
    const data: Record<string, any> = {};

    for (let i = 1; i < reportData.length; i++) {
        // Start from 1 to skip headers
        const row = reportData[i];
        if (row.length < headers.length) continue; // Skip incomplete rows

        const rowData: any = {};
        headers.forEach((header, index) => {
            const value = row[index];
            rowData[header] = isNaN(parseFloat(value)) ? value : parseFloat(value);
        });

        const key = `${rowData.Cluster}-${rowData.Department}-${rowData.Project}-${rowData.Nodepool}`;
        data[key] = rowData;
    }

    return data;
}

function sumMetricsByDepartment(projectReport: Record<string, any>): Record<string, any> {
    const departmentMetrics: Record<string, any> = {};

    Object.values(projectReport).forEach((project: any) => {
        const department = project.Department;
        if (!departmentMetrics[department]) {
            departmentMetrics[department] = { ...project };
        } else {
            Object.keys(project).forEach((key) => {
                if (typeof project[key] === 'number') {
                    departmentMetrics[department][key] += project[key];
                }
            });
        }
    });

    return departmentMetrics;
}

test.describe(`Consumption report tests, tags: ${tags.teams.viper}, ${tags.tool.api}`, () => {
    test.beforeAll(async () => {
        // Reports are SaaS only, so not generating them there is expected. Anywhere else
        // it means setup broke, and skipping would hide it the way it was hidden before.
        // No skip here: the self hosted describe below is the one case that has to run
        // there, and skipping the group would take it with it.
        if (settings.isSelfHosted || settings.isLocalRun) {
            return;
        }
        if (!reportApiUtils.consumptionReportsGenerated) {
            throw new Error(
                `Consumption reports were not generated in global setup: ${reportApiUtils.consumptionReportsError}`
            );
        }
        await reportApiUtils.listReportsAndValidate(3);
    });

    test.afterAll(async () => {
        if (settings.isSelfHosted || !reportApiUtils.consumptionReportsGenerated) {
            return;
        }
        await reportApiUtils.deleteReportAndValidate(flowRootState.reportIds.consumptionReportIdByProject);
        await reportApiUtils.deleteReportAndValidate(flowRootState.reportIds.consumptionReportIdByDepartment);
        await reportApiUtils.deleteReportAndValidate(flowRootState.reportIds.consumptionReportIdByNodePool);
    });

    test.describe('Validate reports response in self hosted', () => {
        shouldRunTests(settings.isSelfHosted)(() => {
            test('Try to generate response in self-hosted and get 400 status code', async () => {
                const reportName = objectGenerator.generateNameByPrefix('consumptionReportByNodePool');
                const endTime = reportApiUtils.reportEndTime;
                const startTime = reportApiUtils.reportStartTime;
                const response = await reportApiUtils.generateReportGroupedBy(
                    reportName,
                    ReportGroupByOptions.Nodepool,
                    startTime,
                    endTime,
                    [`clusterId==${flowRootState.cluster.uuid}`],
                    flowRootState.users.system_administrator.accessToken,
                    StatusCodes.BAD_REQUEST
                );
                expect(response.response.statusCode).toBe(StatusCodes.BAD_REQUEST);
            });
        });
    });

    test.describe('Reports comparison', () => {
        shouldRunTests(!settings.isSelfHosted)(() => {
            let projectReport: Record<string, any>;
            let departmentReport: Record<string, any>;
            let nodepoolReport: Record<string, any>;

            test.beforeAll(async () => {
                projectReport = parseCSVToDict(
                    await reportApiUtils.getReportAndValidate(
                        flowRootState.reportIds.consumptionReportIdByProject
                    )
                );
                departmentReport = parseCSVToDict(
                    await reportApiUtils.getReportAndValidate(
                        flowRootState.reportIds.consumptionReportIdByDepartment
                    )
                );
                nodepoolReport = parseCSVToDict(
                    await reportApiUtils.getReportAndValidate(
                        flowRootState.reportIds.consumptionReportIdByNodePool
                    )
                );
            });

            test('Validate reports comparison', () => {
                const softAssert = new SoftAssert();

                [projectReport, nodepoolReport].forEach((report, index) => {
                    const reportType = index === 0 ? 'Project' : 'Nodepool';
                    const summedMetrics = sumMetricsByDepartment(report);

                    Object.keys(departmentReport).forEach((key) => {
                        const department = departmentReport[key].Department;
                        const departmentMetrics = departmentReport[key];
                        const summedDepartmentMetrics = summedMetrics[department];

                        softAssert.assert(
                            summedDepartmentMetrics !== undefined,
                            `[${reportType}] Summed metrics for department ${department} should not be undefined`
                        );

                        Object.keys(departmentMetrics).forEach((metric) => {
                            if (!isNaN(parseFloat(departmentMetrics[metric]))) {
                                const departmentMetricValue = parseFloat(departmentMetrics[metric]);
                                softAssert.assert(
                                    Math.abs(summedDepartmentMetrics[metric] - departmentMetricValue) <=
                                        departmentMetricValue * 0.01,
                                    `[${reportType}] Metric ${metric} for department ${department} should be close to ${departmentMetricValue}, instead got ${summedDepartmentMetrics[metric]}`
                                );
                            }
                        });
                    });
                });

                softAssert.throwErrorIfFailed();
            });
        });
    });

    test.describe('Batch delete reports', () => {
        shouldRunTests(!settings.isSelfHosted)(() => {
            async function createReports(count: number): Promise<string[]> {
                const ids: string[] = [];
                for (let i = 0; i < count; i++) {
                    const name = objectGenerator.generateNameByPrefix('batchDeleteReport');
                    const res = await reportApiUtils.generateReportGroupedBy(
                        name,
                        ReportGroupByOptions.Nodepool,
                        reportApiUtils.reportStartTime,
                        reportApiUtils.reportEndTime
                    );
                    ids.push(res.body.id);
                }
                return ids;
            }

            test('Batch delete removes every report in the batch', async () => {
                const reportIds = await createReports(3);

                await reportApiUtils.batchDeleteReportsAndValidate(reportIds);
            });

            test('Batch delete reports a missing id as failed and still deletes the rest', async () => {
                const reportIds = await createReports(2);
                const missingReportId = uuidv4();

                const response = await reportApiUtils.batchDeleteReportsAndValidate(
                    [...reportIds, missingReportId],
                    [missingReportId]
                );

                expect(response.failed[0].code).toBe(StatusCodes.NOT_FOUND);
            });

            test('Batch delete refuses a report the user does not own', async () => {
                const reportIds = await createReports(1);

                const response = await reportApiUtils.batchDeleteReports(
                    reportIds,
                    flowRootState.users.l1_researcher.accessToken
                );

                expect(response.response.statusCode).toBe(StatusCodes.ACCEPTED);
                expect(response.body.accepted).toHaveLength(0);
                expect(response.body.failed).toHaveLength(1);
                expect(response.body.failed[0].id).toBe(reportIds[0]);
                expect(response.body.failed[0].code).toBe(StatusCodes.FORBIDDEN);

                // the owner can still remove it, so the test leaves nothing behind
                await reportApiUtils.batchDeleteReportsAndValidate(reportIds);
            });
        });
    });

    test.describe('Reports permissions', () => {
        shouldRunTests(!settings.isSelfHosted)(() => {
            test('Validate user with no permissions to the cluster cannot get data in the report', async () => {
                const reportName = objectGenerator.generateNameByPrefix('consumptionReportByNodePool');
                const endTime = reportApiUtils.reportEndTime;
                const startTime = reportApiUtils.reportStartTime;
                const response = await reportApiUtils.generateReportGroupedBy(
                    reportName,
                    ReportGroupByOptions.Nodepool,
                    startTime,
                    endTime,
                    [`clusterId==${flowRootState.cluster.uuid}`],
                    flowRootState.users.l1_researcher.accessToken,
                    StatusCodes.FORBIDDEN
                );
                expect(response.response.statusCode).toBe(StatusCodes.FORBIDDEN);
            });
        });
    });
}); 