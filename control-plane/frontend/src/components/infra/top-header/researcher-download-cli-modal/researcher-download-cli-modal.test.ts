/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { ResearcherDownloadCliModal } from ".";
import RunaiCopyableCodeBlock from "@/components/common/runai-copyable-code-block/runai-copyable-code-block.vue";
import { useClusterStore } from "@/stores/cluster.store";
import { ClusterState, type ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { EOperatingSystems, cliService } from "@/services/control-plane/cli-exposer.service/cli-exposer.service";

const linuxCommand = "bash -c install-linux";

vi.spyOn(cliService, "getUILinuxCommand").mockResolvedValue(linuxCommand);
vi.spyOn(cliService, "getUIMacCommand").mockResolvedValue("bash -c install-mac");
vi.spyOn(cliService, "getUIWindowsCommand").mockResolvedValue("");

async function factory(state: ClusterState) {
  const pinia = createTestingPinia({ createSpy: vi.fn });
  useClusterStore(pinia).clusters = [
    { uuid: "cluster-1", name: "cluster-1", version: "2.24", domain: "https://cluster-1", state } as ClustersMinimalData,
  ];

  const wrapper = mount(ResearcherDownloadCliModal, {
    props: { open: true },
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        RunaiBaseDialog: {
          template: '<div><slot name="header"></slot><slot name="body"></slot><slot name="footer"></slot></div>',
        },
        RunaiCopyableCodeBlock: true,
        RunaiSubExpansionTextItem: true,
        DownloadButtonWithProgress: true,
      },
    },
  });
  await flushPromises();
  await wrapper.vm.onSelectOS(EOperatingSystems.LINUX);
  await flushPromises();
  return wrapper;
}

describe("researcher-download-cli-modal.vue", () => {
  it.each([ClusterState.Connected, ClusterState.MissingPrerequisites, ClusterState.ServiceIssues])(
    "shows the install command when the only cluster is %s",
    async (state) => {
      const wrapper = await factory(state);
      expect(wrapper.findComponent(RunaiCopyableCodeBlock).props("code")).toBe(linuxCommand);
    },
  );

  it.each([ClusterState.WaitingToConnect, ClusterState.Disconnected, ClusterState.Unknown])(
    "hides the install command when the only cluster is %s",
    async (state) => {
      const wrapper = await factory(state);
      expect(wrapper.findComponent(RunaiCopyableCodeBlock).exists()).toBe(false);
    },
  );
});
