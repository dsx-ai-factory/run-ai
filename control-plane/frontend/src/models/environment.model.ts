/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  ToolType,
  Capability,
  ServingPortInfoProtocolEnum,
  InternalConnectionType,
} from "@/swagger-models/assets-service-client";
import {
  type EnvironmentVariableOfAsset,
  type EnvironmentVariable,
  ImagePullPolicy,
  UidGidSource,
} from "@/swagger-models/assets-service-client";

import { toolTypeIconsMap } from "@/common/icons.constant";
import type { ISelectOption, IUIAssetCreationMeta } from "./global.model";

export const toolTypeOptions: Array<IToolTypeSelectOption> = [
  {
    label: "Jupyter",
    value: "jupyter-notebook",
    icon: toolTypeIconsMap["jupyter-notebook"],
  },
  {
    label: "VSCode",
    value: "visual-studio-code",
    icon: toolTypeIconsMap["visual-studio-code"],
  },
  {
    label: "RStudio",
    value: "rstudio",
    icon: toolTypeIconsMap.rstudio,
  },
  {
    label: "MLflow",
    value: "mlflow",
    icon: toolTypeIconsMap.mlflow,
  },
  {
    label: "Weights & Biases",
    value: "wandb",
    icon: toolTypeIconsMap.wandb,
  },
  {
    label: "Comet",
    value: "comet",
    icon: toolTypeIconsMap.comet,
  },
  {
    label: "TensorBoard",
    value: "tensorboard",
    icon: toolTypeIconsMap.tensorboard,
  },
  {
    label: "MATLAB",
    value: "matlab",
    icon: toolTypeIconsMap.matlab,
  },
  {
    label: "Chatbot UI",
    value: "chatbot-ui",
    icon: toolTypeIconsMap["chatbot-ui"],
  },
  {
    label: "Custom",
    value: "custom",
    icon: toolTypeIconsMap.custom,
  },
];

export const connectionTypeOptions: Array<IConnectionTypeSelectOption> = [
  {
    label: "External URL",
    value: "ExternalUrl",
  },
  {
    label: "NodePort",
    value: "NodePort",
  },
  {
    label: "Load Balancer",
    value: "LoadBalancer",
  },
];
export interface ISecurityModel {
  uidGidSource?: UidGidSource;
  overrideUidGidInWorkspace?: boolean;
  runAsUid?: number | null;
  runAsGid?: number | null;
  supplementalGroups?: string | null;
  capabilities?: Array<Capability> | null;
}

export interface IUidGidSupplementalGroupsModel {
  runAsUid?: number | null;
  runAsGid?: number | null;
  supplementalGroups?: string | null;
}

export interface ISecurityOption {
  label: string;
  value: UidGidSource;
  toolTip: string;
  display: boolean;
}

export interface IImage {
  image: string | null;
  imagePullPolicy: ImagePullPolicy;
}

export interface IRuntimeSettings {
  command: string | null;
  args: string | null;
  workingDir: string | null;
  assetEnvironmentVariables?: Array<EnvironmentVariableOfAsset>;
  hybridEnvironmentVariables?: Array<EnvironmentVariable>;
}

export interface IUICommandAndArgs {
  command?: string;
  args: string;
}

export interface IUIConnection {
  id: string;
  name: string;
  toolType: ToolType | null;
  connectionType?: InternalConnectionType | null;
  containerPort?: number;
  isCustomUrl?: boolean;
  isCustomPort?: boolean;
  nodePort?: number | null;
  externalUrl?: string | null;
  authorizedUsers?: Array<string> | null;
  authorizedGroups?: Array<string> | null;
  protocol?: ServingPortInfoProtocolEnum;
  accessOption?: string | null;
}

export interface IUIEnvironmentAsset {
  meta: IUIAssetCreationMeta;
  spec: IUIEnvironmentAssetSpec;
}

export interface IUIEnvironmentAssetSpec {
  command: string | null;
  args: string | null;
  environmentVariables: Array<EnvironmentVariableOfAsset>;
  image: string | null;
  imagePullPolicy: ImagePullPolicy;
  workingDir: string | null;
  connections: Array<IUIConnection>;
  uidGidSource?: UidGidSource;
  overrideUidGidInWorkspace?: boolean;
  runAsUid?: number | null;
  runAsGid?: number | null;
  supplementalGroups?: string | null;
  capabilities?: Array<Capability> | null;
}

export interface IToolTypeSelectOption extends ISelectOption {
  value: ToolType;
}

export interface IConnectionTypeSelectOption extends ISelectOption {
  value: InternalConnectionType;
}

export type TEnvVarNoneCustomSourceOption = {
  label: string;
  value: string;
  type: string;
  isPrivate?: boolean;
};

export enum EEnvVarSourceType {
  Custom = "custom",
  Credentials = "credentials",
  Secrets = "secrets",
  ConfigMap = "configMap",
  UserCredential = "userCredential",
}

export enum EAccessKeySecretKey {
  Key = "AccessKeyId",
  Secret = "SecretKey",
}

export enum EUserPassSecretKey {
  Pass = "Password",
  User = "User",
}

export enum EEnvVarFormType {
  Assets = "assets",
  Hybrid = "hybrid",
}

export interface IEnvironmentModalOptions {
  environmentIds: Array<string>;
  header: string;
}

export const IMAGE_PULL_POLICY_LABELS = {
  [ImagePullPolicy.IfNotPresent]: "Pull the image only if it's not already present on the host",
  [ImagePullPolicy.Always]: "Always pull the image from the registry",
  [ImagePullPolicy.Never]: "Never pull",
};

export const UID_GID_SOURCE_LABELS = {
  [UidGidSource.FromTheImage]: "Defined by the image",
  [UidGidSource.FromIdpToken]: "Defined by the IdP token",
  [UidGidSource.Custom]: "Custom",
};
