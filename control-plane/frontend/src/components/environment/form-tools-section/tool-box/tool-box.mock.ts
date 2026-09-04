/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IUIConnection } from "@/models/environment.model";
import { InternalConnectionType, ToolType } from "@/swagger-models/assets-service-client";

export const twoMatlabTools: Array<IUIConnection> = [
  {
    id: "1",
    name: "Jupyter Notebook",
    toolType: ToolType.Matlab,
    connectionType: InternalConnectionType.ExternalUrl,
    containerPort: 8888,
  },
  {
    id: "2",
    name: "Jupyter Lab",
    toolType: ToolType.Matlab,
    connectionType: InternalConnectionType.NodePort,
    containerPort: 8888,
  },
];

export const oneMatlabTools: Array<IUIConnection> = [
  {
    id: "1",
    name: "Jupyter 1",
    toolType: ToolType.Matlab,
    connectionType: InternalConnectionType.NodePort,
    containerPort: 8888,
  },
];

export const oneJupyterNodePortTool: Array<IUIConnection> = [
  {
    id: "1",
    name: "Jupyter Notebook - NodePort",
    toolType: ToolType.JupyterNotebook,
    connectionType: InternalConnectionType.NodePort,
    containerPort: 8888,
  },
];

export const twoJupyterTools: Array<IUIConnection> = [
  {
    id: "1",
    name: "Jupyter Notebook",
    toolType: ToolType.JupyterNotebook,
    connectionType: InternalConnectionType.ExternalUrl,
    containerPort: 8888,
  },
  {
    id: "2",
    name: "Jupyter Notebook - NodePort",
    toolType: ToolType.JupyterNotebook,
    connectionType: InternalConnectionType.NodePort,
    containerPort: 8888,
  },
];

export const threeMatlabTools: Array<IUIConnection> = [
  {
    id: "1",
    name: "Matlab",
    toolType: ToolType.Matlab,
    connectionType: InternalConnectionType.ExternalUrl,
    containerPort: 8888,
  },
  {
    id: "2",
    name: "Matlab - NodePort",
    toolType: ToolType.Matlab,
    connectionType: InternalConnectionType.NodePort,
    containerPort: 8888,
  },
  {
    id: "3",
    name: "Matlab - Load Balancer",
    toolType: ToolType.Matlab,
    connectionType: InternalConnectionType.LoadBalancer,
    containerPort: 8888,
  },
];

export const threeJupyterTools: Array<IUIConnection> = [
  {
    id: "1",
    name: "Jupyter Notebook",
    toolType: ToolType.JupyterNotebook,
    connectionType: InternalConnectionType.ExternalUrl,
    containerPort: 8888,
  },
  {
    id: "2",
    name: "Jupyter Notebook - NodePort",
    toolType: ToolType.JupyterNotebook,
    connectionType: InternalConnectionType.NodePort,
    containerPort: 8888,
  },
  {
    id: "3",
    name: "Jupyter Notebook - Load Balancer",
    toolType: ToolType.JupyterNotebook,
    connectionType: InternalConnectionType.LoadBalancer,
    containerPort: 8888,
  },
];
