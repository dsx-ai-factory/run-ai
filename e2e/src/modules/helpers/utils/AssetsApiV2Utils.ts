/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../api';
import http from 'http';

import {
    TrainingCreationRequestV2,
    TrainingV2,
    WorkspaceCreationRequestV2,
    WorkspaceV2,
    InferenceCreationRequestV2,
    InferenceV2
} from '../../api/apis/assetsApi';

export class AssetsApiV2Utils {
    set accessToken(accessToken: string) {
        assetsApi.trainingApi.accessToken = accessToken;
        assetsApi.workspaceApi.accessToken = accessToken;
    }

    async createTraining(
        workloadCreationRequest: TrainingCreationRequestV2
    ): Promise<{ response: http.IncomingMessage; body: TrainingV2 }> {
        return assetsApi.trainingApi.createTrainingV2(workloadCreationRequest);
    }

    async createWorkspace(
        workloadCreationRequest: WorkspaceCreationRequestV2
    ): Promise<{ response: http.IncomingMessage; body: WorkspaceV2 }> {
        return assetsApi.workspaceApi.createWorkspaceV2(workloadCreationRequest);
    }

    async createInference(
        workloadCreationRequest: InferenceCreationRequestV2
    ): Promise<{ response: http.IncomingMessage; body: InferenceV2 }> {
        return assetsApi.inferenceApi.createInferenceV2(workloadCreationRequest);
    }
}
