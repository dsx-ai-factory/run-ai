/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  ClusterPriority,
  UnifiedPolicyInfo,
  UnifiedPolicyInfoPerReplicaType,
} from "@/swagger-models/workloads-client";
import type { IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import type { StringOption, AssetKind } from "@/swagger-models/assets-service-client";
import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";

import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IAssetsFilter } from "@/models/filter.model";

export enum EInferenceSubmitFormDrawerActionType {
  Select = "select",
  SelectMultiple = "select-multiple",
  CreateNew = "create-new",
  Close = "close",
}

// Data interfaces for each drawer action type
export interface IInferenceSubmitFormDrawerActionSelectData {
  id: string;
  name: string;
}

export interface IInferenceSubmitFormDrawerActionSelectMultipleData {
  id: string;
  name: string;
  kind: AssetKind;
}

export interface IInferenceSubmitFormDrawerAction {
  action: EInferenceSubmitFormDrawerActionType;
  selectData?: IInferenceSubmitFormDrawerActionSelectData;
  selectMultipleData?: IInferenceSubmitFormDrawerActionSelectMultipleData[];
  createNewData?: string;
}

export enum EInferenceSubmitFormSectionActionType {
  ResetEnvironmentToDefault = "reset-environment-to-default",
  ClearEnvironmentTemplate = "clear-environment-template",
  ResetComputeToDefault = "reset-compute-to-default",
  ClearComputeTemplate = "clear-compute-template",
}

export interface IInferenceSubmitFormSectionAction {
  action: EInferenceSubmitFormSectionActionType;
}

export enum EInferenceSubmitFormNavigationActionType {
  ChangeSection = "change-section",
}

export enum EInferenceSubmitFormNavigationSection {
  Cluster = "cluster",
  Project = "project",
  Scope = "scope",
  WorkloadArchitecture = "workload-architecture",
  InferenceType = "inference-type",
  Model = "model",
}

export interface IInferenceSubmitFormNavigationAction {
  action: EInferenceSubmitFormNavigationActionType;
  section?: EInferenceSubmitFormNavigationSection;
}

export interface IInferenceSubmitFormNavigation {
  clusterName?: string;
  projectName?: string;
  showInferenceType: boolean;
  typeSummary: string;
  defaultOpenNameSection: boolean;
  modelName?: string;
  canChangeModel?: boolean;
}

export interface IInferenceSubmitFormPolicy {
  policyData: UnifiedPolicyInfoPerReplicaType;
  policyViewDetails: IPolicyViewDetails;
}

export interface IDistributedInferenceSubmitFormPolicy {
  distributedPolicy: UnifiedPolicyInfo;
  policyViewDetails: IPolicyViewDetails;
}

export interface IInferenceSubmitFormSectionTemplates {
  selectedTemplates: THybridFormState["selectedTemplates"];
  selectedEnvironmentName: string;
  selectedComputeName: string;
}

export interface IInferenceSubmitFormClusterConfig {
  selectedProjectNodeTypes?: string[];
  projectNodePoolsResources?: StringOption[];
  autoScaleOptions: IAutoScaleOptions;
  workloadPriorities?: ClusterPriority[];
}

export interface IInferenceSubmitFormDataSources {
  secrets: SecretListInfo[];
  allowedDataSourceOptions?: Set<TDataSourceKinds>;
}

export interface IInferenceSubmitFormFeatures {
  supportImagePullSecrets: boolean;
  isVolumesEnabled: boolean;
  showRequestTimeout: boolean;
  supportPriorityClass: boolean;
  supportPreemptibility: boolean;
}

export interface IInferenceSubmitFormUIState {
  isDrawerOpen?: boolean;
  drawerDataEntity: WorkloadCreationDrawerEntities | null;
  recentlyCreatedEntityId?: string | null;
  submitting: boolean;
  isWorkloadTemplate?: boolean;
  originalTemplateName?: string;
  isEditMode?: boolean;
  scopeChangeInProgress?: boolean;
  baseDrawerFilters?: IAssetsFilter;
  isValidScopeSelected?: boolean;
}

export enum EInferenceSubmitFormLinkActionType {
  Environment = "environment",
  Compute = "compute",
}
