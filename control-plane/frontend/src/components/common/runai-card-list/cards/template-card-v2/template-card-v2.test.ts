/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import TemplateCardV2 from "./template-card-v2.vue";
import { EWorkloadType } from "@/models/workload.model";
import { LLM_IMAGE_FOR_HUGGING_FACE, LLM_IMAGE_FOR_TGI } from "@/models/inference.model";
import type { SupersetTemplate } from "@/swagger-models/workloads-client";

const DEPRECATION_SELECTOR = '[aid="template-card-tgi-deprecation"]';

function createTemplate(workloadType: EWorkloadType, image: string): SupersetTemplate {
  return { meta: { name: "my-template", workloadType }, spec: { image } } as unknown as SupersetTemplate;
}

function mountCard(data: SupersetTemplate) {
  return mount(TemplateCardV2, {
    global: { plugins: [Quasar] },
    props: { data, isSelected: false },
  });
}

describe("template-card-v2.vue", () => {
  describe("TGI deprecation indicator", () => {
    it.each([
      {
        name: "inference template with the TGI image",
        type: EWorkloadType.Inference,
        image: LLM_IMAGE_FOR_TGI,
        want: true,
      },
      {
        name: "inference template with the vLLM image",
        type: EWorkloadType.Inference,
        image: LLM_IMAGE_FOR_HUGGING_FACE,
        want: false,
      },
      {
        name: "training template with the TGI image",
        type: EWorkloadType.Training,
        image: LLM_IMAGE_FOR_TGI,
        want: false,
      },
    ])("should show the indicator: $want for $name", ({ type, image, want }) => {
      const wrapper = mountCard(createTemplate(type, image));

      expect(wrapper.find(DEPRECATION_SELECTOR).exists()).toBe(want);
    });

    it("should not select the card when the indicator is clicked", async () => {
      const wrapper = mountCard(createTemplate(EWorkloadType.Inference, LLM_IMAGE_FOR_TGI));

      await wrapper.find(DEPRECATION_SELECTOR).trigger("click");

      expect(wrapper.emitted("card-clicked")).toBeUndefined();
    });
  });
});
