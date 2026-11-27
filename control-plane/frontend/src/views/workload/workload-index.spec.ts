/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assert, describe, expect, it, vi } from "vitest";
import { EWorkloadType } from "@/models/workload.model";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { Source, type Workload } from "@/swagger-models/workloads-client";
import { workloadIndexMock } from "./workload-index.mock";
import WorkloadIndex from "./workload-index.vue";

const assetBasedWorkload = workloadIndexMock.workloads[0];

function copyWorkload(type: EWorkloadType, overrides: Partial<Workload> = {}) {
  const methods = WorkloadIndex.methods;
  assert.isDefined(methods);
  const page = {
    ...methods,
    selectedWorkload: { ...assetBasedWorkload, type, ...overrides },
    isCopyCliModalOpen: false,
    $router: { push: vi.fn() },
  };
  page.cloneWorkload();
  return page;
}

describe("Copy a workload", () => {
  it.each([
    [EWorkloadType.Inference, INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID],
    [EWorkloadType.Workspace, WORKSPACE_ROUTE_NAMES.WORKSPACE_HYBRID],
    [EWorkloadType.Training, TRAINING_ROUTE_NAMES.TRAINING_HYBRID],
    [EWorkloadType.Distributed, TRAINING_ROUTE_NAMES.DISTRIBUTED_HYBRID],
  ] as const)("opens the flexible form for an asset-based %s", (type, route) => {
    const page = copyWorkload(type);

    expect(page.$router.push).toHaveBeenCalledExactlyOnceWith({
      name: route,
      query: {
        fromCopyId: assetBasedWorkload.id,
        projectId: assetBasedWorkload.projectId,
        clusterId: assetBasedWorkload.clusterId,
      },
    });
  });

  it("opens the distributed inference editor for a flat workload", () => {
    const page = copyWorkload(EWorkloadType.DistributedInference, { environments: [] });

    expect(page.$router.push).toHaveBeenCalledExactlyOnceWith({
      name: INFERENCE_ROUTE_NAMES.INFERENCE_DISTRIBUTED_EDIT,
      query: {
        fromCopyId: assetBasedWorkload.id,
        projectId: assetBasedWorkload.projectId,
        clusterId: assetBasedWorkload.clusterId,
      },
    });
  });

  it("does not route asset-based distributed inference to the flat editor", () => {
    const page = copyWorkload(EWorkloadType.DistributedInference);

    expect(page.$router.push).not.toHaveBeenCalled();
  });

  it("opens the CLI command dialog for a CLI workload", () => {
    const page = copyWorkload(EWorkloadType.Workspace, { source: Source.Cli });

    expect(page.isCopyCliModalOpen).toBe(true);
    expect(page.$router.push).not.toHaveBeenCalled();
  });
});
