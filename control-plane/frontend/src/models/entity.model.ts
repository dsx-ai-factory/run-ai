/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { INode } from "./node.model";
import type { IWorkspaceList } from "./workspace.model";
import type { INodeGpu } from "@/models/node-gpu.model";

export type TEntity = INode | IWorkspaceList | INodeGpu;
