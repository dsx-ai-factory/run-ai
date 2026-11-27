/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import ClusterInstallationSection from "./cluster-installation-section.vue";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getClusterInstallationInfo: vi.fn() },
}));

function createWrapper() {
  return mount(ClusterInstallationSection, {
    props: {
      createdCluster: { uuid: "cluster-uuid", name: "c1" } as never,
      showCertificateInstructions: false,
      isHigherClusterVersion: true,
      isMultiTenant: false,
      selectedInstallationModel: { name: "c1", version: "v1", prerequisites: {} } as never,
      versionsMap: { v1: "3.0.0" },
      preloadedInstructions: "helm upgrade -i runai-cluster",
    },
    global: { plugins: [Quasar, createTestingPinia({ stubActions: false })] },
  });
}

describe("cluster-installation-section", () => {
  it("tells the user the bootstrap token in the install command expires after 2 hours", () => {
    const notice = createWrapper().find('[aid="bootstrap-token-expiry-notice"]');

    expect(notice.exists()).toBe(true);
    expect(notice.text()).toContain("Bootstrap token expires after 2 hours.");
    expect(notice.text()).toContain('click "Get installation instructions"');
  });
});
