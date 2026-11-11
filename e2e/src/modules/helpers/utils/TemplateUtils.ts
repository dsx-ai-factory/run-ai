/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi, workloadsService } from '../../api';
import type { WorkloadTemplate, WorkloadTemplateCreationRequest } from '../../api/apis/assetsApi';

import { flowRootState } from '../config/config';

import { logger } from '..';
import http from 'http';
import {
    DistributedTemplate,
    DistributedTemplateCreationRequest,
    DistributedTemplatesListResponse,
    DistributedTemplateUpdateRequest,
    DistributedInferenceTemplateCreationRequest,
    InferenceTemplate,
    InferenceTemplateCreationRequest,
    InferenceTemplatesListResponse,
    InferenceTemplateUpdateRequest,
    ScopeType,
    SupersetTemplate,
    TemplatesListResponse,
    TrainingTemplate,
    TrainingTemplateCreationRequest,
    TrainingTemplatesListResponse,
    TrainingTemplateUpdateRequest,
    WorkloadType,
    WorkspaceTemplate,
    WorkspaceTemplateCreationRequest,
    WorkspaceTemplatesListResponse,
    WorkspaceTemplateUpdateRequest
} from '../../api/apis/workloadsApi/generated/api';
import { CountResponse } from '../../api/apis/orgUnitApi';

export class TemplateUtils {
    async getTemplateByName(name: string): Promise<WorkloadTemplate | undefined> {
        assetsApi.templateApi.accessToken = flowRootState.users.system_administrator.accessToken;

        logger.debug('getting workspace by name ', name);
        const templates: Array<WorkloadTemplate> = (
            await assetsApi.templateApi.listTemplates(flowRootState.cluster.uuid)
        ).body.entries;
        const template: WorkloadTemplate | undefined = templates.find(
            (template: WorkloadTemplate) => template.meta.name === name
        );
        return template;
    }

    async delete(id: string): Promise<void> {
        try {
            logger.info(`deleting workload-template ${id}`);
            assetsApi.templateApi.accessToken = flowRootState.users.system_administrator.accessToken;

            await assetsApi.templateApi.deleteTemplateById(id);
            logger.info(`workload-template ${id} deleted successfully!`);
        } catch (e) {
            logger.error(`Failed to delete workload-template ${id}: ${e}`);
            throw e;
        }
    }

    async create(
        workspaceTemplateCreationRequest: WorkloadTemplateCreationRequest
    ): Promise<WorkloadTemplate> {
        try {
            logger.info('creating workload-template', workspaceTemplateCreationRequest.meta.name);

            assetsApi.templateApi.accessToken = flowRootState.users.system_administrator.accessToken;

            const template: WorkloadTemplate = (
                await assetsApi.templateApi.createTemplate(workspaceTemplateCreationRequest)
            ).body;

            logger.info(
                `workload-template ${workspaceTemplateCreationRequest.meta.name} created successfully. id: ${template.meta.id}`
            );
            return template;
        } catch (e) {
            logger.error(
                `Failed to create workload-template ${workspaceTemplateCreationRequest.meta.name}: ${e}`
            );
            throw e;
        }
    }

    async deleteAllTemplates(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.templateApi.accessToken = accessToken;
            const templates: WorkloadTemplate[] = (await assetsApi.templateApi.listTemplates()).body.entries;
            await Promise.all(
                templates.map((template) => assetsApi.templateApi.deleteTemplateById(template.meta.id))
            );
            if (templates.length > 0) {
                logger.info('deleted all templates successfully');
            }
        } catch (error) {
            logger.warn(`Failed to delete templates: ${error}`);
        }
    }

    async createAsUser(
        workspaceTemplateCreationRequest: WorkloadTemplateCreationRequest,
        accessToken: string
    ): Promise<WorkloadTemplate> {
        try {
            logger.info('creating workload-template', workspaceTemplateCreationRequest.meta.name);
            assetsApi.templateApi.accessToken = accessToken;
            const template: WorkloadTemplate = (
                await assetsApi.templateApi.createTemplate(workspaceTemplateCreationRequest)
            ).body;

            logger.info(
                `workload-template ${workspaceTemplateCreationRequest.meta.name} created successfully. id: ${template.meta.id}`
            );
            return template;
        } catch (e) {
            logger.error(
                `Failed to create workload-template ${workspaceTemplateCreationRequest.meta.name}: ${e}`
            );
            throw e;
        }
    }

    async deleteAsUser(id: string, accessToken: string): Promise<http.IncomingMessage> {
        try {
            logger.info(`deleting workload-template ${id}`);
            assetsApi.templateApi.accessToken = accessToken;
            const res = await assetsApi.templateApi.deleteTemplateById(id);
            logger.info(`workload-template ${id} deleted successfully!`);
            return res.response;
        } catch (e) {
            logger.error(`Failed to delete workload-template ${id} by user: ${e}`);
            throw e;
        }
    }

    async createWorkloadTemplateV2(
        metaConfig: {
            workloadType: WorkloadType;
            scopeType: ScopeType;
            scopeId: string;
            templateName: string;
            description?: string;
        },
        spec: object,
        distConfig?: {
            masterSpecSameAsWorker?: boolean;
            masterSpec?: object;
        },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{
        response: http.IncomingMessage;
        body: TrainingTemplate | DistributedTemplate | WorkspaceTemplate | InferenceTemplate;
    }> {
        const templateRequest = {
            meta: {
                name: metaConfig.templateName,
                description: metaConfig.description ? metaConfig.description : 'default template description',
                scopeType: metaConfig.scopeType,
                scopeId: metaConfig.scopeId
            },
            spec: spec
        };

        logger.info(`template payload: ${JSON.stringify(templateRequest)}`);

        workloadsService.workloadTemplatesApi.accessToken = accessToken;

        switch (metaConfig.workloadType) {
            case WorkloadType.Training:
                return await workloadsService.workloadTemplatesApi.createTrainingTemplate(
                    templateRequest as TrainingTemplateCreationRequest
                );
            case WorkloadType.Distributed:
                // distributed 'JAX' framework doesn't have master
                if (distConfig) {
                    if (distConfig.masterSpecSameAsWorker) {
                        templateRequest['masterSpecSameAsWorker'] = true;
                    } else {
                        templateRequest['masterSpecSameAsWorker'] = false;
                        templateRequest['master'] = distConfig.masterSpec;
                    }
                }

                return await workloadsService.workloadTemplatesApi.createDistributedTemplate(
                    templateRequest as DistributedTemplateCreationRequest
                );
            case WorkloadType.Workspace:
                return await workloadsService.workloadTemplatesApi.createWorkspaceTemplate(
                    templateRequest as WorkspaceTemplateCreationRequest
                );
            case WorkloadType.Inference:
                return await workloadsService.workloadTemplatesApi.createInferenceTemplate(
                    templateRequest as InferenceTemplateCreationRequest
                );
            case WorkloadType.DistributedInference:
                return (await workloadsService.workloadTemplatesApi.createDistributedInferenceTemplate(
                    templateRequest as DistributedInferenceTemplateCreationRequest
                )) as unknown as { response: http.IncomingMessage; body: InferenceTemplate };
            default:
                throw new Error(`Unknown WorkloadType: ${metaConfig.workloadType}!`);
        }
    }

    async getWorkloadTemplateV2(
        templateId: string,
        workloadType: WorkloadType,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{
        response: http.IncomingMessage;
        body: TrainingTemplate | DistributedTemplate | WorkspaceTemplate | InferenceTemplate;
    }> {
        workloadsService.workloadTemplatesApi.accessToken = accessToken;

        switch (workloadType) {
            case WorkloadType.Training:
                return await workloadsService.workloadTemplatesApi.getTrainingTemplate(templateId);
            case WorkloadType.Distributed:
                return await workloadsService.workloadTemplatesApi.getDistributedTemplate(templateId);
            case WorkloadType.Workspace:
                return await workloadsService.workloadTemplatesApi.getWorkspaceTemplate(templateId);
            case WorkloadType.Inference:
                return await workloadsService.workloadTemplatesApi.getInferenceTemplate(templateId);
            case WorkloadType.DistributedInference:
                return (await workloadsService.workloadTemplatesApi.getDistributedInferenceTemplate(
                    templateId
                )) as unknown as { response: http.IncomingMessage; body: InferenceTemplate };
            default:
                throw new Error(`Unknown WorkloadType: ${workloadType}!`);
        }
    }

    async deleteWorkloadTemplateV2(
        templateId: string,
        templateType: WorkloadType,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{
        response: http.IncomingMessage;
        body?: any;
    }> {
        workloadsService.workloadTemplatesApi.accessToken = accessToken;

        switch (templateType) {
            case WorkloadType.Training:
                return await workloadsService.workloadTemplatesApi.deleteTrainingTemplate(templateId);
            case WorkloadType.Distributed:
                return await workloadsService.workloadTemplatesApi.deleteDistributedTemplate(templateId);
            case WorkloadType.Workspace:
                return await workloadsService.workloadTemplatesApi.deleteWorkspaceTemplate(templateId);
            case WorkloadType.Inference:
                return await workloadsService.workloadTemplatesApi.deleteInferenceTemplate(templateId);
            case WorkloadType.DistributedInference:
                return await workloadsService.workloadTemplatesApi.deleteDistributedInferenceTemplate(
                    templateId
                );
            default:
                throw new Error(`Unknown template type: ${templateType}!`);
        }
    }

    async listWorkloadTemplateV2(
        workloadType: WorkloadType,
        opts?: {
            offset?: number;
            limit?: number;
            sortOrder?: 'asc' | 'desc';
            sortBy?:
                | 'name'
                | 'workloadType'
                | 'runnable'
                | 'scopeType'
                | 'clusterId'
                | 'projectId'
                | 'departmentId'
                | 'createdAt'
                | 'updatedAt'
                | 'createdBy'
                | 'updatedBy';
            search?: string;
        },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{
        response: http.IncomingMessage;
        body:
            | TrainingTemplatesListResponse
            | DistributedTemplatesListResponse
            | WorkspaceTemplatesListResponse
            | InferenceTemplatesListResponse;
    }> {
        workloadsService.workloadTemplatesApi.accessToken = accessToken;

        switch (workloadType) {
            case WorkloadType.Training:
                return await workloadsService.workloadTemplatesApi.listTrainingTemplates(
                    opts?.offset,
                    opts?.limit,
                    opts?.sortOrder,
                    opts?.sortBy,
                    null,
                    opts?.search
                );
            case WorkloadType.Distributed:
                return await workloadsService.workloadTemplatesApi.listDistributedTemplates(
                    opts?.offset,
                    opts?.limit,
                    opts?.sortOrder,
                    opts?.sortBy,
                    null,
                    opts?.search
                );
            case WorkloadType.Workspace:
                return await workloadsService.workloadTemplatesApi.listWorkspaceTemplates(
                    opts?.offset,
                    opts?.limit,
                    opts?.sortOrder,
                    opts?.sortBy,
                    null,
                    opts?.search
                );
            case WorkloadType.Inference:
                return await workloadsService.workloadTemplatesApi.listInferenceTemplates(
                    opts?.offset,
                    opts?.limit,
                    opts?.sortOrder,
                    opts?.sortBy,
                    null,
                    opts?.search
                );
            default:
                throw new Error(`Unknown WorkloadType: ${workloadType}!`);
        }
    }

    async updateWorkloadTemplateV2(
        templateId: string,
        workloadType: WorkloadType,
        metaConfig: object,
        spec: object,
        distConfig?: {
            masterSpecSameAsWorker?: boolean;
            masterSpec?: object;
        },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{
        response: http.IncomingMessage;
        body: TrainingTemplate | DistributedTemplate | WorkspaceTemplate | InferenceTemplate;
    }> {
        const templateRequest = {
            meta: metaConfig,
            spec: spec
        };

        logger.info(`template payload: ${JSON.stringify(templateRequest)}`);

        workloadsService.workloadTemplatesApi.accessToken = accessToken;

        switch (workloadType) {
            case WorkloadType.Training:
                return await workloadsService.workloadTemplatesApi.updateTrainingTemplate(
                    templateId,
                    templateRequest as TrainingTemplateUpdateRequest
                );
            case WorkloadType.Distributed:
                // doesn't want to update master spec or it's 'JAX' which doesn't have master
                if (distConfig) {
                    if (distConfig.masterSpecSameAsWorker) {
                        templateRequest['masterSpecSameAsWorker'] = true;
                    } else {
                        templateRequest['masterSpecSameAsWorker'] = false;
                        templateRequest['master'] = distConfig.masterSpec;
                    }
                }

                return await workloadsService.workloadTemplatesApi.updateDistributedTemplate(
                    templateId,
                    templateRequest as DistributedTemplateUpdateRequest
                );
            case WorkloadType.Workspace:
                return await workloadsService.workloadTemplatesApi.updateWorkspaceTemplate(
                    templateId,
                    templateRequest as WorkspaceTemplateUpdateRequest
                );
            case WorkloadType.Inference:
                return await workloadsService.workloadTemplatesApi.updateInferenceTemplate(
                    templateId,
                    templateRequest as InferenceTemplateUpdateRequest
                );
            default:
                throw new Error(`Unknown WorkloadType: ${workloadType}!`);
        }
    }

    async patchWorkloadTemplateV2(
        templateId: string,
        workloadType: WorkloadType,
        metaConfig: object,
        spec: object,
        distConfig?: {
            masterSpecSameAsWorker?: boolean;
            masterSpec?: object;
        },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{
        response: http.IncomingMessage;
        body: TrainingTemplate | DistributedTemplate | WorkspaceTemplate | InferenceTemplate;
    }> {
        const templateRequest = {
            meta: metaConfig,
            spec: spec
        };

        logger.info(`template payload: ${JSON.stringify(templateRequest)}`);

        workloadsService.workloadTemplatesApi.accessToken = accessToken;

        switch (workloadType) {
            case WorkloadType.Training:
                return await workloadsService.workloadTemplatesApi.patchTrainingTemplate(
                    templateId,
                    templateRequest as TrainingTemplateUpdateRequest
                );
            case WorkloadType.Distributed:
                // doesn't want to patch master spec or it's 'JAX' which doesn't have master
                if (distConfig) {
                    if (distConfig.masterSpecSameAsWorker) {
                        templateRequest['masterSpecSameAsWorker'] = true;
                    } else {
                        templateRequest['masterSpecSameAsWorker'] = false;
                        templateRequest['master'] = distConfig.masterSpec;
                    }
                }

                return await workloadsService.workloadTemplatesApi.patchDistributedTemplate(
                    templateId,
                    templateRequest as DistributedTemplateUpdateRequest
                );
            case WorkloadType.Workspace:
                return await workloadsService.workloadTemplatesApi.patchWorkspaceTemplate(
                    templateId,
                    templateRequest as WorkspaceTemplateUpdateRequest
                );
            case WorkloadType.Inference:
                return await workloadsService.workloadTemplatesApi.patchInferenceTemplate(
                    templateId,
                    templateRequest as InferenceTemplateUpdateRequest
                );
            default:
                throw new Error(`Unknown WorkloadType: ${workloadType}!`);
        }
    }

    async getTemplateCountV2(
        filterBy?: string[],
        search?: string,
        includeAncestors?: boolean
    ): Promise<{
        response: http.IncomingMessage;
        body: CountResponse;
    }> {
        return await workloadsService.workloadTemplatesApi.countTemplates(filterBy, search, includeAncestors);
    }

    async listAllWorkloadTemplatesV2(
        offset?: number,
        limit?: number,
        sortOrder?: 'asc' | 'desc',
        sortBy?:
            | 'name'
            | 'workloadType'
            | 'runnable'
            | 'scopeType'
            | 'clusterId'
            | 'projectId'
            | 'departmentId'
            | 'createdAt'
            | 'updatedAt'
            | 'createdBy'
            | 'updatedBy',
        filterBy?: Array<string>,
        search?: string,
        includeAncestors?: boolean
    ): Promise<{
        response: http.IncomingMessage;
        body: TemplatesListResponse;
    }> {
        return await workloadsService.workloadTemplatesApi.listTemplates(
            offset,
            limit,
            sortOrder,
            sortBy,
            filterBy,
            search,
            includeAncestors
        );
    }

    async genericGetWorkloadTemplateV2(templateId: string): Promise<{
        response: http.IncomingMessage;
        body: SupersetTemplate;
    }> {
        return await workloadsService.workloadTemplatesApi.getTemplateById(templateId);
    }

    async getIdByName(name: string): Promise<string> {
        workloadsService.workloadTemplatesApi.accessToken =
            flowRootState.users.system_administrator.accessToken;
        const templates = (await this.listAllWorkloadTemplatesV2(0, 100)).body.entries;

        const template = templates.find((template) => template.meta.name === name);

        return template.meta.id;
    }

    async getTemplateV2ByName(
        workloadType: WorkloadType,
        name: string
    ): Promise<InferenceTemplate | TrainingTemplate | DistributedTemplate | WorkspaceTemplate> {
        const workloadsTemplates: {
            response: http.IncomingMessage;
            body:
                | TrainingTemplatesListResponse
                | DistributedTemplatesListResponse
                | WorkspaceTemplatesListResponse
                | InferenceTemplatesListResponse;
        } = await this.listWorkloadTemplateV2(workloadType, { search: name });

        // Type assertion to handle union type
        const entries = workloadsTemplates?.body?.entries as Array<
            InferenceTemplate | TrainingTemplate | DistributedTemplate | WorkspaceTemplate
        >;
        if (!entries || !Array.isArray(entries)) {
            return undefined;
        }

        const template:
            | InferenceTemplate
            | TrainingTemplate
            | DistributedTemplate
            | WorkspaceTemplate
            | undefined = entries.find(
            (template: InferenceTemplate | TrainingTemplate | DistributedTemplate | WorkspaceTemplate) =>
                template.meta.name === name
        );
        return template;
    }
}
