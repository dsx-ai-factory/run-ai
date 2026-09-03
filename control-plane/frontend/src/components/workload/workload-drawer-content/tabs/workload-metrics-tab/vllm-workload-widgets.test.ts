/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import VllmWorkloadWidgets from "./vllm-workload-widgets.vue";
import {
  vllmHistogramWidgets,
  vllmLinearWidgets,
} from "@/components/dashboard-v2/widgets/vllm-widgets/vllm-widget-configs";
import type { Workload } from "@/swagger-models/workloads-client";

describe("vllm-workload-widgets", () => {
  it("assembles every vLLM widget with workload and pod selection", () => {
    const wrapper = mount(VllmWorkloadWidgets, {
      props: {
        workload: { id: "workload-id" } as Workload,
        podId: "pod-id",
        filterByDates: { dateStart: new Date(), dateEnd: new Date(), label: "test" },
      },
      global: {
        stubs: {
          VllmMetricWidget: {
            props: ["workloadId", "podId"],
            template: '<div class="linear-widget" :data-workload-id="workloadId" :data-pod-id="podId" />',
          },
          VllmHistogramWidget: {
            props: ["workloadId", "podId"],
            template: '<div class="histogram-widget" :data-workload-id="workloadId" :data-pod-id="podId" />',
          },
        },
      },
    });

    expect(wrapper.findAll(".linear-widget")).toHaveLength(vllmLinearWidgets.length);
    expect(wrapper.findAll(".histogram-widget")).toHaveLength(vllmHistogramWidgets.length);
    expect(wrapper.findAll("[data-workload-id='workload-id']")).toHaveLength(
      vllmLinearWidgets.length + vllmHistogramWidgets.length,
    );
    expect(wrapper.findAll("[data-pod-id='pod-id']")).toHaveLength(
      vllmLinearWidgets.length + vllmHistogramWidgets.length,
    );
  });
});
