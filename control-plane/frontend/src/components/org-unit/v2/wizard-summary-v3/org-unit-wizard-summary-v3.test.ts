/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { defineComponent, h } from "vue";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { provideWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import { EQuotaEntity } from "@/models/resource.model";
import type { SchedulingRules } from "@/swagger-models/org-unit-service-client";
import OrgUnitWizardSummaryV3, { type IProps } from "./org-unit-wizard-summary-v3.vue";

const mountSummary = (props: Partial<IProps> = {}, previewRules?: SchedulingRules | null) => {
  const Host = defineComponent({
    setup() {
      const preview = provideWizardPreview();
      if (previewRules !== undefined) preview.updatePreview("schedulingRules", { schedulingRules: previewRules });
      return () => h(OrgUnitWizardSummaryV3, { entityType: EQuotaEntity.department, accessRulesCount: 0, ...props });
    },
  });
  return mount(Host, { global: { plugins: [Quasar] } });
};

const sectionValue = (wrapper: ReturnType<typeof mountSummary>, aid: string) =>
  wrapper.find(`[aid="wizard-summary-${aid}"] .summary-section-value`).text();

describe("OrgUnitWizardSummaryV3", () => {
  it("shows the design's sections in order", () => {
    const wrapper = mountSummary();

    expect(wrapper.findAll(".summary-section-title").map((title) => title.text())).toEqual([
      "Cluster",
      "Administrator(s)",
      "Resources",
      "Projects",
      "Scheduling rules",
      "Access rules",
    ]);
  });

  it("leaves the sections that later steps fill unset", () => {
    const wrapper = mountSummary();

    for (const aid of ["cluster", "administrators", "resources", "projects"]) {
      expect(sectionValue(wrapper, aid)).toBe("-");
    }
  });

  it("counts the scheduling rules that are set and shows the access rules count", () => {
    const wrapper = mountSummary({
      accessRulesCount: 3,
      schedulingRules: { interactiveJobTimeLimitSeconds: 60, trainingJobTimeLimitSeconds: null } as SchedulingRules,
    });

    expect(sectionValue(wrapper, "scheduling-rules")).toBe("1");
    expect(sectionValue(wrapper, "access-rules")).toBe("3");
  });

  it("counts the scheduling rules being edited before they are saved", () => {
    const wrapper = mountSummary({ schedulingRules: null }, {
      interactiveJobTimeLimitSeconds: 60,
      trainingJobTimeLimitSeconds: 120,
    } as SchedulingRules);

    expect(sectionValue(wrapper, "scheduling-rules")).toBe("2");
  });

  it("explains the scheduling rules and access rules sections for the entity", () => {
    const wrapper = mountSummary();

    expect(wrapper.findAllComponents(RunaiTooltip).map((tooltip) => tooltip.props("tooltipText"))).toEqual([
      "Apply resource and runtime restrictions to all workloads associated with the department",
      "Authorize access to this department",
    ]);
  });
});
