/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, beforeEach, test, expect } from "vitest";
import { GeneralSection } from ".";
import { EWorkloadFormType, Preemptibility } from "@/models/workload.model";
import { PreemptibilitySection } from "./preemptibility-section";
import type { Project } from "@/swagger-models/org-unit-service-client";

describe("general-section.vue", () => {
  let generalSectionCmp: VueWrapper<InstanceType<typeof GeneralSection>>;

  beforeEach(() => {
    generalSectionCmp = mount(GeneralSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        workloadFormType: EWorkloadFormType.Workspace,
        generalModel: {
          allowOverQuota: false,
          autoDeletionTimeAfterCompletionSeconds: null,
          backoffLimit: 0,
        },
        workloadPriorityData: {
          supportPriorityClass: false,
          workloadPriorities: [],
          currProject: {} as Project,
          totalGpuRequest: 0,
        },
      },
    });
  });

  test("component is loading properly", () => {
    expect(generalSectionCmp.html()).toContain("General");
  });

  test("keeps the collapsed summary and warning icon on the same row", () => {
    const subheader = generalSectionCmp.find(".general-section__subheader");
    expect(subheader.classes()).toContain("no-wrap");
    expect(subheader.classes()).toContain("q-pr-md");
    expect(generalSectionCmp.find(".general-section__summary").exists()).toBeTruthy();
  });

  test.each([
    { workloadFormType: EWorkloadFormType.Workspace, offersSemiPreemptible: false },
    { workloadFormType: EWorkloadFormType.Training, offersSemiPreemptible: false },
    { workloadFormType: EWorkloadFormType.Distributed, offersSemiPreemptible: true },
    { workloadFormType: EWorkloadFormType.Inference, offersSemiPreemptible: true },
  ])("filters semi-preemptible for $workloadFormType forms", async ({ workloadFormType, offersSemiPreemptible }) => {
    await generalSectionCmp.setProps({
      workloadFormType,
      supportPreemptibility: true,
      workloadPriorityData: { supportSemiPreemptibility: true },
    });

    const options = generalSectionCmp.findComponent(PreemptibilitySection).vm.preemptibilityOptions;
    expect(options.some((option) => option.value === Preemptibility.SemiPreemptible)).toBe(offersSemiPreemptible);
    expect(options.some((option) => option.value === Preemptibility.Preemptible)).toBe(true);
    expect(options.some((option) => option.value === Preemptibility.NonPreemptible)).toBe(true);
  });

  test("keeps semi-preemptible for arbitrary YAML using the default training form type", async () => {
    await generalSectionCmp.setProps({
      workloadFormType: EWorkloadFormType.Training,
      singlePodWorkload: false,
      supportPreemptibility: true,
      workloadPriorityData: { supportSemiPreemptibility: true },
    });

    expect(generalSectionCmp.findComponent(PreemptibilitySection).vm.preemptibilityOptions).toHaveLength(3);
  });

  test("updates inference options when its single-pod configuration changes", async () => {
    await generalSectionCmp.setProps({
      workloadFormType: EWorkloadFormType.Inference,
      supportPreemptibility: true,
      workloadPriorityData: { supportSemiPreemptibility: true },
      singlePodWorkload: true,
    });
    const preemptibilitySection = generalSectionCmp.findComponent(PreemptibilitySection);
    expect(preemptibilitySection.vm.preemptibilityOptions).toHaveLength(2);

    await generalSectionCmp.setProps({ singlePodWorkload: false });

    expect(preemptibilitySection.vm.preemptibilityOptions).toHaveLength(3);
  });

  describe("over quota toggle", () => {
    test("should render the toggle when the workload type is workspace and supportPriorityClass is false", () => {
      expect(generalSectionCmp.find("[aid=allow-over-quota-input]").exists()).toBeTruthy();
    });

    test("should not render the toggle when the workload type is workspace but supportPriorityClass is true", async () => {
      await generalSectionCmp.setProps({
        workloadPriorityData: {
          supportPriorityClass: true,
          workloadPriorities: [],
          currProject: {} as Project,
          totalGpuRequest: 0,
        },
      });
      expect(generalSectionCmp.find("[aid=allow-over-quota-input]").exists()).toBeFalsy();
    });

    test("should not render the toggle when the workload type is not workspace", async () => {
      await generalSectionCmp.setProps({ workloadFormType: EWorkloadFormType.Training });
      expect(generalSectionCmp.find("[aid=allow-over-quota-input]").exists()).toBeFalsy();
    });

    test("should emit the new generalModel when the toggle is triggered", async () => {
      const overQuotaToggle: DOMWrapper<HTMLInputElement> = generalSectionCmp.find("[aid=allow-over-quota-input]");
      await overQuotaToggle.trigger("click");
      expect(generalSectionCmp.emitted("general-model-changed")).toBeTruthy();
    });
  });
});
