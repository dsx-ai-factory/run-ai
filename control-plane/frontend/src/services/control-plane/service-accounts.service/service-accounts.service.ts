/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Service Accounts API Service
 *
 * This service provides methods to interact with the Service Accounts API endpoints.
 * It handles all CRUD operations for service accounts including:
 * - Listing service accounts with optional filters
 * - Creating new service accounts
 * - Updating existing service accounts
 * - Deleting service accounts
 * - Regenerating service account secrets
 */

import type { IAssetsFilter } from "@/models/filter.model";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type {
  ServiceAccount,
  ServiceAccountCreationRequest,
  ServiceAccountPostResponse,
  ServiceAccountPatchRequest,
  RegenerateServiceAccountSecretResponse,
} from "@/swagger-models/identity-manager-client";

/**
 * Service Accounts API Service
 * Provides methods to interact with service account endpoints
 */
export const serviceAccountsService = {
  list,
  getServiceAccountById,
  createServiceAccount,
  updateServiceAccountById,
  remove,
  regenerateSecret,
  getRunAiInstallerServiceAccountClient,
};

/**
 * Get a list of service accounts
 *
 * @param filterBy - Optional filter parameters
 * @returns Promise resolving to an array of service accounts
 * @throws Error if the request fails
 */
async function list(filterBy: IAssetsFilter = {}): Promise<ServiceAccount[]> {
  try {
    const response = await identityManagerApi.serviceAccountsApi.getServiceAccounts({
      params: {
        projectId: filterBy.projectId,
        scope: filterBy.scope,
        usageInfo: filterBy.usageInfo,
      },
    });
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * Get a specific service account by ID
 *
 * @param id - The service account ID
 * @returns Promise resolving to the service account
 * @throws Error if the request fails or service account not found
 */
async function getServiceAccountById(id: string): Promise<ServiceAccount> {
  try {
    const response = await identityManagerApi.serviceAccountsApi.getServiceAccountById(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * Create a new service account
 *
 * @param serviceAccountCreationRequest - The service account creation data
 * @returns Promise resolving to the created service account with credentials
 * @throws Error if the request fails or validation errors occur
 */
async function createServiceAccount(
  serviceAccountCreationRequest: ServiceAccountCreationRequest,
): Promise<ServiceAccountPostResponse> {
  try {
    const response = await identityManagerApi.serviceAccountsApi.createServiceAccount(serviceAccountCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * Update an existing service account
 *
 * @param id - The service account ID
 * @param patchRequest - The service account update data
 * @returns Promise resolving when update is complete
 * @throws Error if the request fails or service account not found
 */
async function updateServiceAccountById(id: string, patchRequest: ServiceAccountPatchRequest): Promise<void> {
  try {
    const response = await identityManagerApi.serviceAccountsApi.updateServiceAccountById(id, patchRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * Delete a service account
 *
 * @param id - The service account ID to delete
 * @returns Promise resolving when deletion is complete
 * @throws Error if the request fails or service account not found
 */
async function remove(id: string): Promise<void> {
  try {
    const response = await identityManagerApi.serviceAccountsApi.deleteServiceAccountById(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * Regenerate the secret for a service account
 *
 * @param id - The service account ID
 * @returns Promise resolving to the new credentials
 * @throws Error if the request fails or service account not found
 */
async function regenerateSecret(id: string): Promise<RegenerateServiceAccountSecretResponse> {
  try {
    const response = await identityManagerApi.serviceAccountsApi.regenerateServiceAccountSecret(id);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * Get the Run:AI installer service account client
 * This is a special system service account used for cluster installation
 *
 * The installer service account has a fixed name: "runai-internal.installer"
 *
 * @returns Promise resolving to the installer service account with credentials
 * @throws Error if the request fails or installer not found
 */
async function getRunAiInstallerServiceAccountClient(): Promise<ServiceAccountPostResponse> {
  try {
    const response = await identityManagerApi.serviceAccountsApi.getInstallerServiceAccount();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
