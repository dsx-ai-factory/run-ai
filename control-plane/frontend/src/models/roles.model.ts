/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Action } from "@/swagger-models/authorization-client";

export type TRolesTypes =
  | "admin"
  | "researcher"
  | "research_manager"
  | "editor"
  | "viewer"
  | "ml_engineer"
  | "department_admin";

export enum ERolesLabels {
  "admin" = "Admin",
  "researcher" = "Researcher",
  "research_manager" = "Research manager",
  "editor" = "Editor",
  "viewer" = "Viewer",
  "ml_engineer" = "ML engineer",
  "department_admin" = "Department admin",
}

export const Actions: Array<IAction> = [
  {
    id: Action.Read,
    title: "View",
    description:
      "If checked, an assigned user with this role can view instances of this type of entity within their defined scope",
    tooltipWidth: "330px",
  },
  {
    id: Action.Update,
    title: "Edit",
    description:
      "If checked, an assigned user with this role can change the settings of instances of this type of entity within their defined scope",
    tooltipWidth: "300px",
  },
  {
    id: Action.Create,
    title: "Create",
    description:
      "If checked, an assigned user with this role can create new instances of this type of entity within their defined scope",
    tooltipWidth: "300px",
  },
  {
    id: Action.Delete,
    title: "Delete",
    description:
      "If checked, an assigned user with this role can delete instances of this type of entity within their defined scope",
    tooltipWidth: "300px",
  },
];

export interface IAction {
  id: Action;
  title: string;
  description: string;
  tooltipWidth: string;
}

export enum ERoleType {
  System = "System",
  Custom = "Custom",
}

export enum ERoleStage {
  New = "New",
  Legacy = "Legacy",
}

export interface IRoleReplacement {
  replacementRole?: string;
  replacementLink?: string; // For future use
}

export interface IRoleStageConfig {
  stage: ERoleStage;
  replacement?: IRoleReplacement;
}

export const ROLE_STAGE_CONFIG: Record<string, IRoleStageConfig> = {
  // New roles
  "AI practitioner": { stage: ERoleStage.New },
  "Data & storage administrator": { stage: ERoleStage.New },
  "Project administrator": { stage: ERoleStage.New },

  // Legacy with replacement
  "L1 researcher": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "AI practitioner" },
  },
  "L2 researcher": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "AI practitioner" },
  },
  "ML engineer": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "AI practitioner" },
  },
  "Research manager": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "Project administrator" },
  },
  "Data source administrator": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "Data & storage administrator" },
  },
  "Data volume administrator": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "Data & storage administrator" },
  },

  "Department administrator (Legacy)": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "Department administrator" },
  },
  "Department administrator-DGXC (Legacy)": {
    stage: ERoleStage.Legacy,
    replacement: { replacementRole: "Department administrator-DGXC" },
  },

  // Legacy without replacement
  "Environment administrator": { stage: ERoleStage.Legacy },
  "Compute resource administrator": { stage: ERoleStage.Legacy },
  "Credentials administrator": { stage: ERoleStage.Legacy },
  "Template administrator": { stage: ERoleStage.Legacy },
};

// Maps config role names to their actual API names when they differ.
export const ROLE_NAME_API_MAP: Record<string, string> = {
  "Data & storage administrator": "Data and storage administrator",
};
