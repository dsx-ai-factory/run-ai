/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import ClusterAuthenticationSection from "./cluster-autentications-section.vue";
import { useAuthStore } from "@/stores/auth.store";

async function createWrapper() {
  const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
  const authStore = useAuthStore(pinia);
  authStore.loadTenantConfigInfo = vi.fn().mockResolvedValue({
    serverConfigInfo: { authUrl: "https://auth.example.com", realm: "runai" },
  });

  const wrapper = mount(ClusterAuthenticationSection, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        RunaiExpansionItem: { template: "<div><slot /></div>" },
        ServerConfig: true,
      },
    },
  });
  await flushPromises();
  return wrapper;
}

describe("cluster-authentication-section.vue", () => {
  it("mounts and loads tenant server config without reading the permission store", async () => {
    const wrapper = await createWrapper();

    expect(wrapper.findComponent({ name: "ServerConfig" }).exists()).toBe(true);
  });

  it("never wraps its content in a setting-readonly-input (parent controls visibility)", async () => {
    const wrapper = await createWrapper();

    const readonlyWrappers = wrapper.findAllComponents({ name: "setting-readonly-input" });
    expect(readonlyWrappers).toHaveLength(0);
  });
});
