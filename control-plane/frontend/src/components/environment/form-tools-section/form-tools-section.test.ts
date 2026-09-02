/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test, vi } from "vitest";
import { FormToolsSection } from ".";
import type { IUIConnection } from "@/models/environment.model";

import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

// stores
import { useClusterStore } from "@/stores/cluster.store";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("form-tools-section.vue", () => {
  let formToolsSectionCmp: VueWrapper<InstanceType<typeof FormToolsSection>>;

  beforeEach(() => {
    formToolsSectionCmp = mount(FormToolsSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        tools: [],
      },
    });

    const clusterStore = useClusterStore(pinia);
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);
  });

  const tool: IUIConnection = { id: "1", name: "tool-1", connectionType: "ExternalUrl", toolType: "chatbot-ui" };

  test("component is loading properly", () => {
    expect(formToolsSectionCmp.html()).toContain("Tools");
  });

  // Summary tests
  test("section summary should be 'None' when there are no tools", () => {
    expect(formToolsSectionCmp.vm.subheaderSummary).toContain("None");
  });

  test("section summary should be 'Tool(s) 1' when there is 1 tool", async () => {
    await formToolsSectionCmp.setProps({ tools: [tool] });
    expect(formToolsSectionCmp.vm.subheaderSummary).toBe("Tool(s) 1");
  });

  test("section should emit update when there adding a tool", async () => {
    const elAddToolButton = formToolsSectionCmp.find("[aid=add-tool-button]");
    expect(formToolsSectionCmp.emitted("update")).toBeFalsy();
    await elAddToolButton.trigger("click");
    expect(formToolsSectionCmp.emitted("update")).toBeTruthy();
  });

  // Tools amount tests
  test("should show the correct amount tool-box when the tools prop contains 2 tools", async () => {
    await formToolsSectionCmp.setProps({ tools: [tool, tool] });
    expect(formToolsSectionCmp.findAll(".environment-tool-box").length).toBe(2);
  });

  // Emits tests
  test("should not emit 'update' when the user didn't add or remove a tool", async () => {
    expect(formToolsSectionCmp.emitted("update")).toBeFalsy();
  });
  test("should emit 'update' when the user add a tool", async () => {
    const elAddToolButton = formToolsSectionCmp.find("[aid=add-tool-button]");
    await elAddToolButton.trigger("click");
    expect(formToolsSectionCmp.emitted("update")).toBeTruthy();
  });
  test("should emit 'update' twice when the user add a tool and remove tool", async () => {
    const elAddToolButton = formToolsSectionCmp.find("[aid=add-tool-button]");
    await elAddToolButton.trigger("click");
    expect(formToolsSectionCmp.emitted("update")?.length).toBe(1);
    await formToolsSectionCmp.setProps({ tools: [tool, tool] });
    const elRemoveToolButton = formToolsSectionCmp.find("[aid=remove-tool-button]");
    await elRemoveToolButton.trigger("click");
    expect(formToolsSectionCmp.emitted("update")?.length).toBe(2);
  });
});
