/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { HybridEnvironmentSection } from ".";
import { useAuthStore } from "@/stores/auth.store";
import { Quasar } from "quasar";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import type { ISecurityModel, IUIConnection } from "@/models/environment.model";
import { ToolType, UidGidSource } from "@/swagger-models/assets-service-client";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";

vi.mock("@/stores/auth.store");

vi.mock("@/services/control-plane/catalog.service/catalog.service", () => ({
  catalogService: {
    getRegistries: vi.fn().mockResolvedValue([]),
  },
}));

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("hybrid-environment-section.vue", () => {
  const workloadSpec: IUISuperWorkloadSpec = {
    image: "test-image",
    imagePullPolicy: "Always",
    connections: [],
    command: null,
    args: null,
    workingDir: null,
    environmentVariables: [],
    security: {
      uidGidSource: undefined,
      runAsUid: null,
      runAsGid: null,
      supplementalGroups: null,
      capabilities: null,
    },
  };

  const authStore = {
    isSSO: false,
  };

  const factory = (props = {}) => {
    const wrapper = mount(HybridEnvironmentSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        workloadSpec,
        selectedName: "test-name",
        label: "test-label",
        policy: {},
        environmentVariableSourceValueOptions: [],
        clusterId: "test-cluster-id",
        renderKey: "test-render-key",
        ...props,
      },
    });
    wrapper.vm.$q.notify = vi.fn();
    return wrapper;
  };

  beforeEach(() => {
    vi.mocked(useAuthStore).mockReturnValue(authStore as unknown as ReturnType<typeof useAuthStore>);
  });

  it("renders correctly", () => {
    const wrapper = factory();

    expect(wrapper.exists()).toBe(true);
  });

  it("emits update event when updateModel is called", async () => {
    const wrapper = factory();

    await wrapper.vm.updateModel("image", "new-image");
    expect(wrapper.emitted().update[0]).toEqual([{ ...workloadSpec, image: "new-image" }]);
  });

  it("emits update event when updateRuntimeSettings is called", async () => {
    const wrapper = factory();

    const runtimeSettings = {
      command: "new-command",
      args: "new-args",
      workingDir: "new-working-dir",
      hybridEnvironmentVariables: [],
    };

    await wrapper.vm.updateRuntimeSettings(runtimeSettings);
    expect(wrapper.emitted().update[0]).toEqual([
      {
        ...workloadSpec,
        environmentVariables: runtimeSettings.hybridEnvironmentVariables,
        command: runtimeSettings.command,
        args: runtimeSettings.args,
        workingDir: runtimeSettings.workingDir,
      },
    ]);
  });

  it("emits update event when updateSecuritySettings is called", async () => {
    const wrapper = factory();

    const securityModel: ISecurityModel = {
      uidGidSource: UidGidSource.Custom,
      runAsUid: 1000,
      runAsGid: 1000,
      supplementalGroups: "111111",
      capabilities: [],
    };

    await wrapper.vm.updateSecuritySettings(securityModel);
    expect(wrapper.emitted().update[0]).toEqual([
      {
        ...workloadSpec,
        security: { ...workloadSpec.security, ...securityModel },
      },
    ]);
  });

  it("emits update event when updateTools is called", async () => {
    const wrapper = factory();

    const tools: IUIConnection[] = [
      { id: "1", name: "tool1", toolType: ToolType.JupyterNotebook },
      { id: "2", name: "tool2", toolType: ToolType.Tensorboard },
    ];

    await wrapper.vm.updateTools(tools);
    expect(wrapper.emitted().update[0]).toEqual([
      {
        ...workloadSpec,
        connections: tools,
      },
    ]);
  });
});
