/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { EnvironmentAsset } from "@/swagger-models/assets-service-client";

export const environments: Array<EnvironmentAsset> = [
  {
    meta: {
      name: "custom-tools",
      scope: "project",
      clusterId: "f9f177ea-6685-42c8-be1e-3d894d6386f2",
      projectId: 1,
      id: "8cc24924-3ec5-4f97-9f14-293989e7d80b",
      kind: "environment",
      tenantId: 1,
      createdBy: "test@run.ai",
      createdAt: "2023-03-05T12:22:42.750213Z",
      updatedBy: "test@run.ai",
      updatedAt: "2023-03-05T12:22:42.750213Z",
      projectName: "project-1",
    },
    spec: {
      environmentVariables: [],
      image: "ubuntu",
      imagePullPolicy: "IfNotPresent",
      connections: [
        {
          name: "Jupyter",
          internalToolInfo: {
            toolType: "jupyter-notebook",
            connectionType: "ExternalUrl",
            containerPort: 8888,
            externalUrlInfo: {
              isCustomUrl: true,
            },
          },
        },
        {
          name: "VSCode - NodePort",
          internalToolInfo: {
            toolType: "visual-studio-code",
            connectionType: "NodePort",
            containerPort: 8080,
            nodePortInfo: {
              isCustomPort: true,
            },
          },
        },
      ],
      uidGidSource: "fromIdpToken",
      overrideUidGidInWorkspace: true,
      capabilities: [],
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "jupyter-lab",
      scope: "tenant",
      id: "0122ab4e-44eb-4f11-a694-ea15f60504c6",
      kind: "environment",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-05T07:42:22.96088Z",
      updatedBy: "runai",
      updatedAt: "2023-03-05T07:42:22.96088Z",
    },
    spec: {
      command: "start-notebook.sh",
      args: "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
      environmentVariables: [],
      image: "jupyter/scipy-notebook",
      imagePullPolicy: "IfNotPresent",
      connections: [
        {
          name: "Jupyter",
          internalToolInfo: {
            toolType: "jupyter-notebook",
            connectionType: "ExternalUrl",
            containerPort: 8888,
            externalUrlInfo: {},
          },
        },
      ],
      uidGidSource: "fromTheImage",
      capabilities: [],
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "jupyter-tensorboard",
      scope: "tenant",
      id: "3f034c18-0eaa-4538-8556-758b7774ad6a",
      kind: "environment",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-05T07:42:22.961561Z",
      updatedBy: "runai",
      updatedAt: "2023-03-05T07:42:22.961561Z",
    },
    spec: {
      environmentVariables: [
        {
          name: "LOG_DIR",
          value: "./",
        },
      ],
      image: "quay.io/jupyter/scipy-notebook",
      imagePullPolicy: "IfNotPresent",
      connections: [
        {
          name: "Jupyter",
          internalToolInfo: {
            toolType: "jupyter-notebook",
            connectionType: "ExternalUrl",
            containerPort: 8888,
            externalUrlInfo: {},
          },
        },
        {
          name: "TensorBoard",
          internalToolInfo: {
            toolType: "tensorboard",
            connectionType: "ExternalUrl",
            containerPort: 6006,
            externalUrlInfo: {},
          },
        },
      ],
      uidGidSource: "fromTheImage",
      capabilities: [],
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "tensorboard",
      scope: "tenant",
      id: "52fffdff-8787-48f9-9e94-2e26065e3fe5",
      kind: "environment",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-05T07:42:22.959501Z",
      updatedBy: "runai",
      updatedAt: "2023-03-05T07:42:22.959502Z",
    },
    spec: {
      command: "tensorboard",
      args: "--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
      environmentVariables: [],
      image: "tensorflow/tensorflow:latest",
      imagePullPolicy: "IfNotPresent",
      connections: [
        {
          name: "TensorBoard",
          internalToolInfo: {
            toolType: "tensorboard",
            connectionType: "ExternalUrl",
            containerPort: 6006,
            externalUrlInfo: {},
          },
        },
      ],
      uidGidSource: "fromTheImage",
      capabilities: [],
    },
    compliance: {
      compliance: true,
    },
  },
];

export const environmentUidGidSourceTest: EnvironmentAsset = {
  meta: {
    name: "custom-tools",
    scope: "project",
    clusterId: "f9f177ea-6685-42c8-be1e-3d894d6386f2",
    projectId: 1,
    id: "8cc24924-3ec5-4f97-9f14-293989e7d80b",
    kind: "environment",
    tenantId: 1,
    createdBy: "test@run.ai",
    createdAt: "2023-03-05T12:22:42.750213Z",
    updatedBy: "test@run.ai",
    updatedAt: "2023-03-05T12:22:42.750213Z",
    projectName: "project-1",
  },
  spec: {
    environmentVariables: [],
    image: "ubuntu",
    imagePullPolicy: "IfNotPresent",
    connections: [],
    uidGidSource: "fromIdpToken",
    overrideUidGidInWorkspace: true,
    capabilities: [],
  },
};

export const environmentVariablesEnvTest: EnvironmentAsset = {
  meta: {
    name: "custom-tools",
    scope: "project",
    clusterId: "f9f177ea-6685-42c8-be1e-3d894d6386f2",
    projectId: 1,
    id: "8cc24924-3ec5-4f97-9f14-293989e7d80b",
    kind: "environment",
    tenantId: 1,
    createdBy: "test@run.ai",
    createdAt: "2023-03-05T12:22:42.750213Z",
    updatedBy: "test@run.ai",
    updatedAt: "2023-03-05T12:22:42.750213Z",
    projectName: "project-1",
  },
  spec: {
    environmentVariables: [
      { name: "project-id", value: "proj-1" },
      { name: "cluster-id", value: "cluster-1" },
    ],
    image: "ubuntu",
    imagePullPolicy: "IfNotPresent",
    connections: [],
    uidGidSource: "fromIdpToken",
    overrideUidGidInWorkspace: true,
    capabilities: [],
  },
};

export const commandArgsEnvTest: EnvironmentAsset = {
  meta: {
    name: "custom-tools",
    scope: "project",
    clusterId: "f9f177ea-6685-42c8-be1e-3d894d6386f2",
    projectId: 1,
    id: "8cc24924-3ec5-4f97-9f14-293989e7d80b",
    kind: "environment",
    tenantId: 1,
    createdBy: "test@run.ai",
    createdAt: "2023-03-05T12:22:42.750213Z",
    updatedBy: "test@run.ai",
    updatedAt: "2023-03-05T12:22:42.750213Z",
    projectName: "project-1",
  },
  spec: {
    environmentVariables: [
      { name: "project-id", value: "proj-1" },
      { name: "cluster-id", value: "cluster-1" },
    ],
    image: "ubuntu",
    imagePullPolicy: "IfNotPresent",
    connections: [],
    uidGidSource: "fromIdpToken",
    overrideUidGidInWorkspace: true,
    capabilities: [],
  },
};

export const customToolsEnvTest: EnvironmentAsset = {
  meta: {
    name: "custom-tools",
    scope: "tenant",
    id: "7f4d1993-ba90-4ec1-8eb4-24febfcfa62f",
    kind: "environment",
    tenantId: 1,
    createdBy: "test@run.ai",
    createdAt: "2023-03-09T08:25:47.243191Z",
    updatedBy: "test@run.ai",
    updatedAt: "2023-03-09T08:25:47.243191Z",
  },
  spec: {
    environmentVariables: [],
    image: "ubuntu",
    imagePullPolicy: "IfNotPresent",
    connections: [
      {
        name: "TensorBoard",
        internalToolInfo: {
          toolType: "tensorboard",
          connectionType: "ExternalUrl",
          containerPort: 6006,
          externalUrlInfo: {
            isCustomUrl: true,
          },
        },
      },
      {
        name: "Jupyter - NodePort",
        internalToolInfo: {
          toolType: "jupyter-notebook",
          connectionType: "NodePort",
          containerPort: 8888,
          nodePortInfo: {
            isCustomPort: true,
          },
        },
      },
      {
        name: "RStudio",
        internalToolInfo: {
          toolType: "rstudio",
          connectionType: "NodePort",
          containerPort: 8787,
          externalUrlInfo: {},
        },
      },
    ],
    uidGidSource: "fromTheImage",
    capabilities: [],
  },
  compliance: {
    compliance: true,
  },
  usedBy: {
    workspaces: [
      {
        id: "1053e2f1-d682-4714-9dc3-5e9d69ee2647",
        name: "wsws",
        status: "Pending",
      },
    ],
    assets: {},
  },
};
