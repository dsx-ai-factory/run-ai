/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Workload } from "@/swagger-models/workloads-client";
import { ECustomCell, type ITableColumn } from "@/models/table.model";
import { Phase, Source } from "@/swagger-models/workloads-client";

export const mockWorkloadRow: Workload = {
  tenantId: 1,
  runningPods: 0,
  phaseUpdatedAt: "2025-03-12T10:00:00Z",
  k8sPhaseUpdatedAt: "2025-03-12T10:00:00Z",
  updatedAt: "2025-03-12T10:00:00Z",
  source: Source.Cli,
  deletedAt: null,
  type: "Training",
  name: "runai-0a5701bb402e",
  id: "df8f0794-b4e0-4111-9c3b-be60b0e1d93b",
  priority: null,
  priorityClassName: "",
  clusterId: "fc72c81d-0d70-416a-b6ff-5b48759ae215",
  projectName: "default",
  projectId: "1",
  departmentName: "",
  departmentId: "",
  namespace: "runai-default",
  createdAt: "2025-03-12T10:00:00Z",
  phase: Phase.Failed,
  conditions: [],
  phaseMessage: "RunaiJob has reached the specified backoff limit",
  k8sPhase: "",
  category: "",
  requestedPods: { number: 1 },
  requestedNodePools: ["node-pool-with-long-name"],
  currentNodePools: ["default"],
  images: ["ubuntu"],
  environments: [{ name: "", id: "" }],
  additionalFields: {
    cliCommand: "runai training submit -i ubuntu --backoff-limit 1 -c -- bash -c 'sleep 15; exit 1'",
    issues:
      '{"network":null,"storage":[{"kind":"RunaiJob","name":"runai-0a5701bb402e","description":"[RunaiJob:runai-0a5701bb402e] RunaiJob has reached the specified backoff limit"}],"compute":[{"kind":"RunaiJob","name":"runai-0a5701bb402e","description":"[RunaiJob:runai-0a5701bb402e] RunaiJob has reached the specified backoff limit"}],"internal":null}',
  },
  preemptible: true,
  fullySaasWorkload: true,
  environmentVariables: { RUNAI_JOB_NAME: "${RUNAI_JOB_NAME}", RUNAI_PROJECT: "${RUNAI_PROJECT}" },
  command: "bash -c 'sleep 15; exit 1'",
};

export const mockWorkloadColumn: ITableColumn = {
  name: "phase",
  label: "Status",
  field: "phase",
  sortable: true,
  align: "left",
  customCell: ECustomCell.WORKLOAD_STATUS_COL,
  customCellEvent: { emitName: "workload-status-clicked" },
  format: (val: string) => val,
};

export const mockStatusOptions = {
  status: "Failed",
  tooltipText: "",
  color: "negative",
  displayAnimation: false,
  statusUpdatedTimeInMs: 1741769567868,
};
