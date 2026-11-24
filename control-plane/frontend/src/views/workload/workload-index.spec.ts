/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assert, describe, expect, it, vi } from "vitest";
import { EWorkloadType } from "@/models/workload.model";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { Source } from "@/swagger-models/workloads-client";
import { workloadIndexMock } from "./workload-index.mock";
import WorkloadIndex from "./workload-index.vue";

const assetBasedWorkload = workloadIndexMock.workloads[0];

function copyWorkload(type: EWorkloadType, isHybridSubmissionEnabled: boolean, source: Source = Source.ControlPlane) {
  const methods = WorkloadIndex.methods;
  assert.isDefined(methods);
  const page = {
    ...methods,
    selectedWorkload: { ...assetBasedWorkload, type, source },
    settingStore: { isHybridSubmissionEnabled },
    isCopyCliModalOpen: false,
    $router: { push: vi.fn() },
  };
  page.cloneWorkload();
  return page;
}

describe.each([false, true])("Copy with flexible submission enabled = %s", (enabled) => {
  it.each([
    [EWorkloadType.Inference, INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID],
    [EWorkloadType.Workspace, WORKSPACE_ROUTE_NAMES.WORKSPACE_HYBRID],
    [EWorkloadType.Training, TRAINING_ROUTE_NAMES.TRAINING_HYBRID],
    [EWorkloadType.Distributed, TRAINING_ROUTE_NAMES.DISTRIBUTED_HYBRID],
  ] as const)("opens the flexible form for an asset-based %s", (type, route) => {
    const page = copyWorkload(type, enabled);

    expect(page.$router.push).toHaveBeenCalledExactlyOnceWith({
      name: route,
      query: {
        fromCopyId: assetBasedWorkload.id,
        projectId: assetBasedWorkload.projectId,
        clusterId: assetBasedWorkload.clusterId,
      },
    });
  });

  it("opens the CLI command dialog for a CLI workload", () => {
    const page = copyWorkload(EWorkloadType.Workspace, enabled, Source.Cli);

    expect(page.isCopyCliModalOpen).toBe(true);
    expect(page.$router.push).not.toHaveBeenCalled();
  });
});
