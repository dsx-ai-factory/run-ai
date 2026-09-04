/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { StringOption, AssetKind } from "@/swagger-models/assets-service-client";
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import type { IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import type { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { IInferenceFormUIState } from "@/stores/inference.store";
import type {
  EInferenceSubmitFormNavigationActionType,
  EInferenceSubmitFormNavigationSection,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import type { ISelectOption } from "@/models/global.model";

// Drawer action types matching inference-submit-form pattern
export enum EHfInferenceSubmitFormDrawerActionType {
  Select = "select",
  SelectMultiple = "select-multiple",
  CreateNew = "create-new",
  Close = "close",
}

// Data interfaces for each drawer action type
export interface IHfInferenceSubmitFormDrawerActionSelectData {
  id: string;
  name: string;
}

export interface IHfInferenceSubmitFormDrawerActionSelectMultipleData {
  id: string;
  name: string;
  kind: AssetKind;
}

export interface IHfInferenceSubmitFormDrawerAction {
  action: EHfInferenceSubmitFormDrawerActionType;
  selectData?: IHfInferenceSubmitFormDrawerActionSelectData;
  selectMultipleData?: IHfInferenceSubmitFormDrawerActionSelectMultipleData[];
  createNewData?: string;
}

// Grouped prop interfaces matching inference-submit-form pattern
export interface IHfInferenceSubmitFormNavigation {
  clusterName?: string;
  projectName?: string;
  showInferenceType: boolean;
  typeSummary: string;
  defaultOpenNameSection: boolean;
  modelName?: string;
  canChangeModel?: boolean;
}

export interface IHfInferenceSubmitFormPolicy {
  policyData: UnifiedPolicyInfoPerReplicaType;
  policyViewDetails: IPolicyViewDetails;
}

export interface IHfInferenceSubmitFormSectionTemplates {
  selectedTemplates: THybridFormState["selectedTemplates"];
  selectedComputeName: string;
}

export interface IHfInferenceSubmitFormClusterConfig {
  selectedProjectNodeTypes?: string[];
  projectNodePoolsResources?: StringOption[];
  autoScaleOptions: IAutoScaleOptions;
}

export interface IHfInferenceSubmitFormDataSources {
  secrets: SecretListInfo[];
  allowedDataSourceOptions?: Set<TDataSourceKinds>;
  inferenceSupportedDataSources: Set<TDataSourceKinds>;
  shouldOpenModelStoreSection: boolean;
}

export interface IHfInferenceSubmitFormFeatures {
  supportMultiFractionGpu: boolean;
  supportTolerations: boolean;
}

export interface IHfInferenceSubmitFormUIState {
  isDrawerOpen: boolean;
  drawerDataEntity: WorkloadCreationDrawerEntities | null;
  recentlyCreatedEntityId?: string | null;
  submitting: boolean;
  uiState: IInferenceFormUIState;
  drawerFilters: IAssetsFilter;
  isWorkloadTemplate?: boolean;
  originalTemplateName?: string;
  scopeChangeInProgress?: boolean;
  isValidScopeSelected?: boolean;
}

export interface IHfInferenceSubmitFormNavigationAction {
  action: EInferenceSubmitFormNavigationActionType;
  section?: EInferenceSubmitFormNavigationSection;
}

export const SUBMIT_OPTIONS: Record<string, ISelectOption> = {
  submit: {
    label: "Create Inference",
    value: "submit-inference",
    aid: "create-inference-btn",
  },
  advancedSetup: {
    label: "Advanced setup",
    value: "advanced-setup",
    toolTip: "Fine-tune your setup with additional preferences, including environment details and data sources.",
    aid: "advanced-setup-btn",
  },
  submitTemplate: {
    label: "Create Template",
    value: "submit-template",
    aid: "create-template-btn",
  },
};
