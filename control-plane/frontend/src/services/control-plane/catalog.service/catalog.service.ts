/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { catalogServiceApi } from "@/services/infra/client-apis/catalog-service-api/catalog-service-api";
import type {
  HelmChartListResponse,
  HelmRepo,
  HuggingFaceModelItem,
  ImagesListResponse,
  RegistryIntegration,
  RegistryIntegrationCreationRequest,
  RegistryIntegrationUpdateRequest,
} from "@/swagger-models/catalog-service-client";

export const catalogService = {
  createRegistry,
  updateRegistry,
  deleteRegistry,
  getRegistries,
  getRegistryById,
  getImages,
  getImageTags,
  searchHuggingFaceCatalog,
  getHelmRepos,
  browseHelmRepoCharts,
};

async function createRegistry(registry: RegistryIntegrationCreationRequest): Promise<RegistryIntegration> {
  const response = await catalogServiceApi.registriesApi.createRegistryIntegration(registry);
  return response.data;
}

async function updateRegistry(
  registryId: string,
  registry: RegistryIntegrationUpdateRequest,
): Promise<RegistryIntegration> {
  const response = await catalogServiceApi.registriesApi.updateRegistryIntegration(registryId, registry);
  return response.data;
}

async function deleteRegistry(registryId: string): Promise<void> {
  await catalogServiceApi.registriesApi.deleteRegistryIntegration(registryId);
}

async function getRegistries(): Promise<Array<RegistryIntegration>> {
  const response = await catalogServiceApi.registriesApi.listRegistryIntegrations();
  return response.data.registries;
}

async function getRegistryById(registryIntegrationId: string): Promise<RegistryIntegration> {
  const response = await catalogServiceApi.registriesApi.getRegistryIntegrationById(registryIntegrationId);
  return response.data;
}

async function getImages(registryIntegrationId: string, search?: string, assetId?: string): Promise<ImagesListResponse> {
  const response = await catalogServiceApi.imagesApi.listImages(registryIntegrationId, search, assetId);
  return response.data;
}

async function getImageTags(registryIntegrationId: string, imageName: string, assetId?: string): Promise<Array<string>> {
  const response = await catalogServiceApi.imagesApi.getImageTags(registryIntegrationId, imageName, assetId);
  return response.data.tags || [];
}

async function searchHuggingFaceCatalog(query?: string): Promise<Array<HuggingFaceModelItem>> {
  const response = await catalogServiceApi.inferencesApi.searchHuggingFaceCatalog(query);
  return response.data?.resources || [];
}

async function getHelmRepos(): Promise<Array<HelmRepo>> {
  const response = await catalogServiceApi.helmReposApi.listHelmRepos();
  return response.data.helmRepos;
}

async function browseHelmRepoCharts(helmRepoId: string, assetId?: string): Promise<HelmChartListResponse> {
  const response = await catalogServiceApi.helmReposApi.browseHelmRepoCharts(helmRepoId, assetId);
  return response.data;
}
