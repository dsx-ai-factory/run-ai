/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import type * as VueRouter from "vue-router";
import { describe, expect, it, vi } from "vitest";
import InferenceSubmitForm from "./inference-submit-form.vue";
import { GeneralSection } from "@/components/section/general-section";
import { PreemptibilitySection } from "@/components/section/general-section/preemptibility-section";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { Preemptibility, type IUISuperWorkloadCreationRequest } from "@/models/workload.model";

vi.mock("vue-router", async (importOriginal) => ({
  ...(await importOriginal<typeof VueRouter>()),
  useRoute: () => ({ query: {}, params: {} }),
}));

function factory(minReplicas?: number, maxReplicas?: number, supportSemiPreemptibility = true) {
  return shallowMount(InferenceSubmitForm, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      stubs: { GeneralSection: false, PreemptibilitySection: false },
      renderStubDefaultSlot: true,
      mocks: { $route: { query: {} } },
    },
    props: {
      modelValue: {
        name: "inference",
        clusterId: "cluster",
        projectId: "1",
        spec: { autoscaling: { minReplicas, maxReplicas } },
      },
      navigation: { showInferenceType: false, typeSummary: "", defaultOpenNameSection: true },
      sectionTemplates: {
        selectedTemplates: workloadHybridUtil.getWorkloadHybridTemplatesSetup(),
        selectedEnvironmentName: "",
        selectedComputeName: "",
      },
      clusterConfig: { autoScaleOptions: { isAutoScaleEnabled: true, latencyEnabled: false } },
      dataSources: { secrets: [] },
      features: {
        supportImagePullSecrets: false,
        isVolumesEnabled: false,
        showRequestTimeout: false,
        supportPriorityClass: false,
        supportPreemptibility: true,
        supportSemiPreemptibility,
      },
      uiState: { submitting: false, drawerDataEntity: null },
      project: { name: "project" } as Project,
    },
  });
}

describe("inference-submit-form preemptibility", () => {
  it.each([
    { min: 1, max: 1, singlePod: true },
    { min: 1, max: 5, singlePod: false },
    { min: 3, max: 3, singlePod: false },
    { min: 0, max: 1, singlePod: false },
    { min: undefined, max: undefined, singlePod: false },
  ])("identifies min=$min, max=$max as single-pod=$singlePod", ({ min, max, singlePod }) => {
    const wrapper = factory(min, max);

    expect(wrapper.findComponent(GeneralSection).props("singlePodWorkload")).toBe(singlePod);
    wrapper.unmount();
  });

  it("updates single-pod eligibility when the replica settings change", async () => {
    const wrapper = factory(1, 5);
    const generalSection = wrapper.findComponent(GeneralSection);
    expect(generalSection.props("singlePodWorkload")).toBe(false);

    await wrapper.setProps({
      modelValue: {
        ...wrapper.props("modelValue"),
        spec: { autoscaling: { minReplicas: 1, maxReplicas: 1 } },
      } as IUISuperWorkloadCreationRequest,
    });

    expect(generalSection.props("singlePodWorkload")).toBe(true);
    wrapper.unmount();
  });

  it("filters options without changing the model when inference becomes single-pod", async () => {
    const wrapper = factory(1, 5);
    await wrapper.setProps({
      modelValue: {
        ...wrapper.props("modelValue"),
        spec: { autoscaling: { minReplicas: 1, maxReplicas: 5 }, preemptibility: Preemptibility.SemiPreemptible },
      } as IUISuperWorkloadCreationRequest,
    });
    expect(wrapper.findComponent(PreemptibilitySection).vm.preemptibilityOptions).toHaveLength(3);

    await wrapper.setProps({
      modelValue: {
        ...wrapper.props("modelValue"),
        spec: { autoscaling: { minReplicas: 1, maxReplicas: 1 }, preemptibility: Preemptibility.SemiPreemptible },
      } as IUISuperWorkloadCreationRequest,
    });

    expect(wrapper.findComponent(PreemptibilitySection).vm.preemptibilityOptions).toHaveLength(2);
    expect(wrapper.props("modelValue").spec.preemptibility).toBe(Preemptibility.SemiPreemptible);
    expect(wrapper.emitted("update:model-value")).toBeUndefined();
    wrapper.unmount();
  });

  it("preserves the cluster capability check for multi-pod inference", () => {
    const wrapper = factory(1, 5, false);

    expect(wrapper.findComponent(GeneralSection).props("workloadPriorityData")?.supportSemiPreemptibility).toBe(false);
    wrapper.unmount();
  });
});
