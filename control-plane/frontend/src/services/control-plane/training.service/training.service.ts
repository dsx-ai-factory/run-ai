/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// services
import { storageUtil } from "@/utils/storage.util";
import { dateUtil } from "@/utils/date.util";

// models
import type { ILocalStatus } from "@/models/global.model";
import { type Training as TrainingFlat } from "@/swagger-models/workloads-client";
// constants
import { httpService } from "@/services/infra/https.service/http.service";
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import type {
  HttpResponse,
  TrainingCreationRequest,
  DistributedCreationRequest,
} from "@/swagger-models/workloads-client";

const TRAINING_STATUS = "trainingStatus";

export const trainingService = {
  resumeTrainingV2,
  stopTrainingV2,
  stop,
  createTrainingFlat,
  createDistributedFlat,
};

async function createTrainingFlat(trCreationRequest: TrainingCreationRequest): Promise<TrainingFlat> {
  try {
    const response = await workloadServiceApi.trainingsApi.createTraining(trCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createDistributedFlat(distCreationRequest: DistributedCreationRequest): Promise<TrainingFlat> {
  try {
    const response = await workloadServiceApi.distributedApi.createDistributed(distCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function resumeTrainingV2(trainingId: string): Promise<HttpResponse> {
  try {
    const response = await workloadServiceApi.trainingsApi.resumeTraining(trainingId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function stopTrainingV2(trainingId: string): Promise<HttpResponse> {
  try {
    const response = await workloadServiceApi.trainingsApi.suspendTraining(trainingId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function _cleanupStatus() {
  const currentStatus: Record<string, ILocalStatus> = _getTrainingStatus();
  if (!currentStatus || Object.keys(currentStatus).length === 0) return;
  Object.keys(currentStatus).forEach((key: string) => {
    if (_shouldCleanLocalStatus(currentStatus[key])) {
      delete currentStatus[key];
    }
  });
  _saveTrainingStatus(currentStatus);
}

// this should be called only once in a while,
// because it's a heavy operation and not needed to be called on every status update
_cleanupStatus();

function _shouldCleanLocalStatus(trainingToCheck: ILocalStatus): boolean {
  return dateUtil.moreThanHourFromNow(trainingToCheck.creationTimestamp);
}

function _getTrainingStatus(): Record<string, ILocalStatus> {
  return storageUtil.get<Record<string, ILocalStatus>>(TRAINING_STATUS) || {};
}

function _saveTrainingStatus(trainingStatus: Record<string, ILocalStatus>): void {
  storageUtil.save<Record<string, ILocalStatus>>(TRAINING_STATUS, trainingStatus);
}
