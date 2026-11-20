/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BatchDeleteReportsResponse,
    Report,
    ReportGroupByOptions,
    ReportRequestFields,
    ReportStatus
} from '../../api/apis/orgUnitApi';
import { flowRootState, settings } from '../config/config';
import { orgUnitService } from '../../api';
import { HttpError } from '../../api/apis/orgUnitApi/generated/api/httpError';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { logger, objectGenerator, waitUtils } from '../index';
import http from 'http';

export class ReportApiUtils {
    reportStartTime = '2024-11-10T00:00:00Z';

    reportEndTime = '2024-11-10T12:33:44Z';

    // Derived from flowRootState, not held on this object: global setup runs in its own
    // process and only flowRootState reaches the workers through globalState.json.
    get consumptionReportsGenerated(): boolean {
        const {
            consumptionReportIdByNodePool,
            consumptionReportIdByProject,
            consumptionReportIdByDepartment
        } = flowRootState.reportIds;
        return Boolean(
            consumptionReportIdByNodePool && consumptionReportIdByProject && consumptionReportIdByDepartment
        );
    }

    get consumptionReportsError(): string | undefined {
        return flowRootState.reportIds.generationError;
    }

    async generateReportGroupedBy(
        reportName: string,
        groupBy: ReportGroupByOptions,
        startTime: Date | string,
        endTime: Date | string,
        filterBy = [`clusterId==${flowRootState.cluster.uuid}`],
        accessToken = flowRootState.users.system_administrator.accessToken,
        expectedStatusCode: number = StatusCodes.ACCEPTED
    ): Promise<{ response: http.IncomingMessage; body: Report }> {
        orgUnitService.reportsApi.accessToken = accessToken;

        const reportRequestFields = {
            name: reportName,
            // The generated model types these as Date and the serializer calls toISOString
            // on them, so a string reaches it as an object without that method.
            start: new Date(startTime),
            end: new Date(endTime),
            groupBy: groupBy,
            filterBy: filterBy
        } as ReportRequestFields;

        // The generated client rejects on any non-2xx, so an expected failure arrives as a
        // thrown HttpError rather than a response the caller can assert on.
        const res = await orgUnitService.reportsApi
            .createReport(reportRequestFields)
            .catch((error: unknown) => {
                if (!(error instanceof HttpError)) {
                    throw error;
                }
                return { response: error.response, body: error.body as Report };
            });

        expect(res.response.statusCode).to.equal(expectedStatusCode);
        if (expectedStatusCode === StatusCodes.ACCEPTED) {
            expect(res.body).to.not.be.undefined;
        }

        return res;
    }

    async listReportsAndValidate(
        expectedCount: number,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        orgUnitService.reportsApi.accessToken = accessToken;
        const listResponse = await orgUnitService.reportsApi.listReports([
            `createdBy==${flowRootState.users.system_administrator.email}`
        ]);

        expect(listResponse.response.statusCode).to.equal(
            StatusCodes.OK,
            `Failed to list reports, with status code ${listResponse.response.statusCode}`
        );
        expect(listResponse.body).to.not.be.undefined;
        expect(listResponse.body.reports.length).to.equal(
            expectedCount,
            `Report list count does not equal ${expectedCount}, count is ${listResponse.body.reports.length}`
        );

        const countResponse = await orgUnitService.reportsApi.countReports([
            `createdBy==${flowRootState.users.system_administrator.email}`
        ]);

        expect(countResponse.response.statusCode).to.equal(
            StatusCodes.OK,
            `Failed to count reports, with status code ${countResponse.response.statusCode}`
        );
        expect(countResponse.body).to.not.be.undefined;
        expect(countResponse.body.count).to.equal(
            expectedCount,
            `Reports count does not equal ${expectedCount}, count is ${countResponse.body.count}`
        );
    }

    async getReportAndValidate(
        reportId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<string> {
        orgUnitService.reportsApi.accessToken = accessToken;

        const res = await orgUnitService.reportsApi.getReportById(reportId);

        expect(res.response.statusCode).to.equal(
            StatusCodes.OK,
            `Failed to get report by id ${reportId}, with status code ${res.response.statusCode}`
        );
        expect(res.body).to.not.be.undefined;
        expect(res.body.status).to.equal(
            'Ready',
            `Report ${reportId} is not ready, status is ${res.body.status}`
        );

        const reportData = await orgUnitService.reportsApi.downloadReportById(reportId);

        expect(reportData.response.statusCode).to.equal(
            StatusCodes.OK,
            `Failed to download report by id ${reportId}, with status code ${reportData.response.statusCode}`
        );
        expect(reportData.body).to.not.be.undefined;

        return reportData.body;
    }

    async deleteReport(
        reportId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<http.IncomingMessage> {
        orgUnitService.reportsApi.accessToken = accessToken;
        const res = await orgUnitService.reportsApi.deleteReportById(reportId);
        return res.response;
    }

    async deleteReportAndValidate(
        reportId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        orgUnitService.reportsApi.accessToken = accessToken;

        const res = await orgUnitService.reportsApi.deleteReportById(reportId);

        expect(res.response.statusCode).to.equal(
            StatusCodes.NO_CONTENT,
            `Failed to delete report by id ${reportId}, with status code ${res.response.statusCode}`
        );
    }

    async batchDeleteReports(
        reportIds: string[],
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: BatchDeleteReportsResponse }> {
        orgUnitService.reportsApi.accessToken = accessToken;

        // The spec marks ids uniqueItems, so the generator types it as a Set - but the
        // serializer has no Set branch, and a Set stringifies to {}. Send the array.
        const ids = [...new Set(reportIds)] as unknown as Set<string>;

        return orgUnitService.reportsApi.batchDeleteReports({ ids });
    }

    // Null once the report is gone, which is what deletion is waiting for.
    async reportDeletionOutcome(
        reportId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<ReportStatus | null> {
        orgUnitService.reportsApi.accessToken = accessToken;
        try {
            return (await orgUnitService.reportsApi.getReportById(reportId)).body.status;
        } catch (error) {
            if (error?.statusCode === StatusCodes.NOT_FOUND) {
                return null;
            }
            throw error;
        }
    }

    async waitForReportsDeleted(
        reportIds: string[],
        accessToken = flowRootState.users.system_administrator.accessToken,
        timeoutSeconds = 300
    ): Promise<void> {
        // Ask per report rather than listing: the list is filtered and paginated, so a
        // report can drop out of it while its deletion is still running.
        await waitUtils.waitForCondition(
            async () => {
                let outcomes: (ReportStatus | null)[];
                try {
                    outcomes = await Promise.all(
                        reportIds.map((id) => this.reportDeletionOutcome(id, accessToken))
                    );
                } catch (error) {
                    // waitForCondition does not catch predicate throws, so one bad response
                    // would end the whole wait rather than being retried on the next tick.
                    logger.debug(`Failed to read report status while waiting for deletion: ${error}`);
                    return false;
                }
                expect(outcomes).to.not.include(
                    ReportStatus.DeletionFailed,
                    'A report failed to delete, so waiting for it to disappear would never finish'
                );
                return outcomes.every((outcome) => outcome === null);
            },
            `reports ${reportIds.join(', ')} to be deleted`,
            true,
            timeoutSeconds
        );
    }

    async batchDeleteReportsAndValidate(
        reportIds: string[],
        expectedFailedIds: string[] = [],
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<BatchDeleteReportsResponse> {
        const requestedIds = [...new Set(reportIds)];
        const res = await this.batchDeleteReports(requestedIds, accessToken);

        expect(res.response.statusCode).to.equal(
            StatusCodes.ACCEPTED,
            `Failed to batch delete reports, with status code ${res.response.statusCode}`
        );
        expect(res.body).to.not.be.undefined;

        const expectedAcceptedIds = requestedIds.filter((id) => !expectedFailedIds.includes(id));
        const failedIds = res.body.failed.map((failure) => failure.id);

        expect(res.body.accepted).to.have.members(
            expectedAcceptedIds,
            `Accepted ids ${res.body.accepted} do not match expected ${expectedAcceptedIds}`
        );
        expect(failedIds).to.have.members(
            expectedFailedIds,
            `Failed ids ${failedIds} do not match expected ${expectedFailedIds}`
        );
        expect(res.body.totalRequested).to.equal(requestedIds.length);
        expect(res.body.totalAccepted).to.equal(expectedAcceptedIds.length);
        expect(res.body.totalFailed).to.equal(expectedFailedIds.length);

        await this.waitForReportsDeleted(res.body.accepted, accessToken);

        return res.body;
    }

    async waitForReportsReady(
        reportIds: string[],
        accessToken = flowRootState.users.system_administrator.accessToken,
        timeoutSeconds = 300
    ): Promise<void> {
        orgUnitService.reportsApi.accessToken = accessToken;
        await waitUtils.waitForCondition(
            async () => {
                let statuses: ReportStatus[];
                try {
                    statuses = await Promise.all(
                        reportIds.map(
                            async (id) => (await orgUnitService.reportsApi.getReportById(id)).body.status
                        )
                    );
                } catch (error) {
                    logger.debug(`Failed to read report status while waiting for Ready: ${error}`);
                    return false;
                }
                expect(statuses).to.not.include(
                    ReportStatus.Failed,
                    'A report failed to generate, cannot continue'
                );
                return statuses.every((status) => status === ReportStatus.Ready);
            },
            `reports ${reportIds.join(', ')} to be Ready`,
            true,
            timeoutSeconds
        );
    }

    // Failures are recorded rather than thrown so one broken area cannot take down every
    // other suite in the run.
    async generateConsumptionReportsForSuites(): Promise<void> {
        if (settings.isSelfHosted || settings.isLocalRun) {
            logger.info('Skipping consumption report generation, reports are SaaS only');
            return;
        }

        try {
            [
                flowRootState.reportIds.consumptionReportIdByNodePool,
                flowRootState.reportIds.consumptionReportIdByProject,
                flowRootState.reportIds.consumptionReportIdByDepartment
            ] = await this.generateConsumptionReport();

            await this.waitForReportsReady([
                flowRootState.reportIds.consumptionReportIdByNodePool,
                flowRootState.reportIds.consumptionReportIdByProject,
                flowRootState.reportIds.consumptionReportIdByDepartment
            ]);
        } catch (error) {
            logger.error(`Failed to generate consumption reports: ${error}`);
            flowRootState.reportIds.generationError = `${error}`;
        }
    }

    async generateConsumptionReport(): Promise<string[]> {
        const nodePoolReportName = objectGenerator.generateNameByPrefix('reportByNP');
        const projectReportName = objectGenerator.generateNameByPrefix('reportByProject');
        const departmentReportName = objectGenerator.generateNameByPrefix('reportByDepartment');
        const startTime = this.reportStartTime;
        const endTime = this.reportEndTime;

        const nodePoolResult = (
            await this.generateReportGroupedBy(
                nodePoolReportName,
                ReportGroupByOptions.Nodepool,
                startTime,
                endTime
            )
        ).body;
        const projectResult = (
            await this.generateReportGroupedBy(
                projectReportName,
                ReportGroupByOptions.Project,
                startTime,
                endTime
            )
        ).body;
        const departmentResult = (
            await this.generateReportGroupedBy(
                departmentReportName,
                ReportGroupByOptions.Department,
                startTime,
                endTime
            )
        ).body;
        logger.info(
            `Generated reports with ids: ${nodePoolResult.id}, ${projectResult.id}, ${departmentResult.id}`
        );

        return [nodePoolResult.id, projectResult.id, departmentResult.id];
    }
}
