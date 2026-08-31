/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IK8sDistributionCardData } from "@/models/cluster.model";
import type {
  ComputeAsset,
  DatasourceListResponseEntry,
  EnvironmentAsset,
  WorkloadTemplate,
} from "@/swagger-models/assets-service-client";
import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";
export interface ICreateNewCard {
  meta: {
    id: string;
  };
}
export interface ICardQuery {
  page?: number;
  search: string | number | null;
  sortBy: string | null;
  filterBy?: string | null;
}

export type TCardSortValues = "name" | "creation" | "recentUsage";

export enum TCardCmpName {
  CREATE_NEW = "create-new-card",
  TEMPLATE = "template-card",
  TEMPLATE_V2 = "template-card-v2",
  PROJECT = "project-card",
  ENVIRONMENT = "environment-card",
  COMPUTE_RESOURCE = "compute-resource-card",
  DATA_SOURCE = "data-source-card",
  K8S_DISTRIBUTION = "k8s-distribution-card",
  CREDENTIAL = "credential-card",
}

export type TCardListCmpType =
  | ProjectMetadata
  | EnvironmentAsset
  | ICreateNewCard
  | WorkloadTemplate
  | ComputeAsset
  | DatasourceListResponseEntry
  | IK8sDistributionCardData;

export interface ICardListItem {
  id: string;
  cardName: TCardCmpName;
  data: TCardListCmpType;
  searchValues?: string[];
  locked?: boolean;
  disabled?: boolean;
  tooltip?: string;
  showDisabledInfo?: boolean;
  sortInfo?: {
    name?: string;
    createdAt?: string;
    recentUsage?: string | null;
  };
}
