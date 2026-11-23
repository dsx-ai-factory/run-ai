/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workloadsService } from '../../api';
import { flowRootState } from '../config/config';
import { logger } from '../logger/logger';
import { AIApplicationCreationRequest } from '../../api/apis/workloadsApi/generated/model/aIApplicationCreationRequest';
import { AIApplication } from '../../api/apis/workloadsApi/generated/model/aIApplication';
import { AIApplicationVerbosity } from '../../api/apis/workloadsApi/generated/model/aIApplicationVerbosity';
import { AIApplicationsSortFilterFields } from '../../api/apis/workloadsApi/generated/model/aIApplicationsSortFilterFields';
import { ListApplications200Response } from '../../api/apis/workloadsApi/generated/model/listApplications200Response';
import http from 'http';

export class AiApplicationUtils {
    private setAccessToken(accessToken?: string): void {
        workloadsService.aiApplicationsApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
    }

    async createAiApplication(
        request: AIApplicationCreationRequest,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: AIApplication }> {
        this.setAccessToken(accessToken);
        const result = await workloadsService.aiApplicationsApi.createApplication(request);
        logger.info(`Created AI application: ${result.body.name}, id: ${result.body.id}`);
        return result;
    }

    async getAiApplication(
        applicationId: string,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: AIApplication }> {
        this.setAccessToken(accessToken);
        return workloadsService.aiApplicationsApi.getApplication(applicationId);
    }

    async deleteAiApplication(
        applicationId: string,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        this.setAccessToken(accessToken);
        const result = await workloadsService.aiApplicationsApi.deleteApplication(applicationId);
        logger.info(`Deleted AI application: ${applicationId}`);
        return result;
    }

    async listAiApplications(
        options: {
            accessToken?: string;
            verbosity?: AIApplicationVerbosity;
            offset?: number;
            limit?: number;
            sortOrder?: 'asc' | 'desc';
            sortBy?: AIApplicationsSortFilterFields;
            filterBy?: Array<string>;
            search?: string;
        } = {}
    ): Promise<{ response: http.IncomingMessage; body: ListApplications200Response }> {
        this.setAccessToken(options.accessToken);
        return workloadsService.aiApplicationsApi.listApplications(
            options.verbosity,
            options.offset,
            options.limit,
            options.sortOrder,
            options.sortBy,
            options.filterBy,
            options.search
        );
    }

    async countAiApplications(
        options: {
            accessToken?: string;
            filterBy?: Array<string>;
            search?: string;
        } = {}
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        this.setAccessToken(options.accessToken);
        return workloadsService.aiApplicationsApi.countApplications(options.filterBy, options.search);
    }
}
