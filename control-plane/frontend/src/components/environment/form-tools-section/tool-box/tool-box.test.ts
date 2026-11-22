/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// @vitest-environment-options {"url":"https://runai.example.com/"}

import { mount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { QForm, Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import { defineComponent } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";
import { errorMessages } from "@/common/error-message.constant";
import { InternalConnectionType, ToolType } from "@/swagger-models/assets-service-client";
import ToolBox from "./tool-box.vue";

describe("tool-box Custom URL validation", () => {
  it.each([
    ["registered route", "/workload-templates", false],
    ["registered route child", "/workload-templates/custom", false],
    ["control-plane root", "/", false],
    ["unrelated workload path", "/project/tool", true],
    ["another host", "https://tools.example.com/workload-templates", true],
  ])("validates the form for a %s", async (_, path, valid) => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: "/workload-templates", component: {} }],
    });
    const wrapper = mount(QForm, {
      slots: {
        default: defineComponent({
          components: { ToolBox },
          data: () => ({
            tool: {
              id: "custom-tool",
              name: "Custom tool",
              toolType: ToolType.Custom,
              connectionType: InternalConnectionType.ExternalUrl,
              isCustomUrl: true,
              externalUrl: new URL(path, window.location.href).href,
              containerPort: 8888,
            },
          }),
          template: '<tool-box :tool="tool" :tools="[tool]" />',
        }),
      },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn }), router],
      },
    });

    expect(await wrapper.vm.validate()).toBe(valid);
    if (!valid) {
      expect(wrapper.text()).toContain(errorMessages.URL_RESERVED_FOR_CONTROL_PLANE);
    }
    wrapper.unmount();
  });
});
