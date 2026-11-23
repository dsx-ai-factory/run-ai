/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../api';
import type { WorkspaceListResponse, Workspace } from '../../api/apis/assetsApi';
import { Workspace as WorkspaceModel } from '../../api/apis/workloadsApi/generated/model/workspace';
import { flowRootState } from '../config/config';
import { logger, policyUtils, workloadsCrudUtils } from '../../helpers';
import http from 'http';
import { MatchExpressionOperator } from '../../api/apis/policyApi';
import { WorkspaceSpecSpecAllOfCompute } from '../../api/apis/workloadsApi/generated/api';

export class WorkspaceUtils {
    async getWorkspaceByName(clusterId: string, name: string): Promise<Workspace | undefined> {
        assetsApi.workspaceApi.defaultHeaders = {
            Authorization: 'Bearer ' + flowRootState.users.system_administrator.accessToken
        };

        logger.debug('getting workspace by name ', name);
        const workspaceListResponse: WorkspaceListResponse = (
            await assetsApi.workspaceApi.listWorkspaces(clusterId)
        ).body;
        const workspace: Workspace = workspaceListResponse.entries.find(
            (workspace: Workspace) => workspace.meta.name === name
        );
        return workspace;
    }

    async createWorkspace(
        projectId: string,
        workloadName: string,
        workspaceSpec: {
            image: string;
            command: string;
            compute?: WorkspaceSpecSpecAllOfCompute;
        },
        excludeNodeNames?: string[],
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: WorkspaceModel }> {
        if (excludeNodeNames) {
            workspaceSpec['nodeAffinityRequired'] = policyUtils.getNodeSelectorTerm(
                MatchExpressionOperator.NotIn,
                excludeNodeNames
            ).nodeAffinityRequired;
        }

        if (accessToken) {
            workloadsCrudUtils.accessToken = accessToken;
        } else {
            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
        }

        return workloadsCrudUtils.createWorkspace(
            workloadName,
            false,
            flowRootState.cluster.uuid,
            projectId,
            workspaceSpec
        );
    }
}
