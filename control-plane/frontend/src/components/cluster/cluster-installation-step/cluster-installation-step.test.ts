/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import ClusterInstallationStep from "./cluster-installation-step.vue";
import { afterEach, beforeEach, describe, expect, test, vi as vitestVi } from "vitest";
import { Quasar } from "quasar";
import { EClusterDistIds, EClusterLocations, type IClusterInstalltionModel } from "@/models/cluster.model";
import { createTestingPinia } from "@pinia/testing";
import { useAuthStore } from "@/stores/auth.store";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { FAKE_ACCESS_TOKEN_A } from "@/tests/fixtures/auth.fixtures";

const accessToken = FAKE_ACCESS_TOKEN_A;

const pinia = createTestingPinia({ createSpy: vitestVi.fn });
const authStore = useAuthStore(pinia);
describe("cluster-installation-step.vue", () => {
  let clusterInstallationStepCmp: VueWrapper<InstanceType<typeof ClusterInstallationStep>>;
  function mockWindowObject(): void {
    global.window = Object.create(window);
    global.window.location = {
      // @ts-expect-error
      ancestorOrigins: [""],
      hash: "",
      host: "dummy.com",
      port: "80",
      protocol: "http:",
      hostname: "dummy.com",
      href: "http://dummy.com?page=1&name=testing",
      origin: "http://dummy.com",
      pathname: "",
    };
  }

  beforeEach(() => {
    authStore.$patch({ accessToken: accessToken });
    clusterService.getClusterInstallationInfo = vitestVi.fn().mockResolvedValue({
      data: { installationStr: "backend-generated-installation-command" },
    });

    mockWindowObject();

    clusterInstallationStepCmp = mount(ClusterInstallationStep, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        isMultiTenant: false as boolean,
        createdCluster: { uuid: "abcd" } as DisplayedCluster | null,
        clusterFormModel: {
          name: "",
          version: "",
          prerequisites: {
            location: "",
            clusterUrl: "",
            k8sDistribution: "",
          },
        } as IClusterInstalltionModel,
        versionsMap: {
          "2.9": "2.9",
          "2.10": "2.10",
          "2.11": "2.11",
          "2.12": "2.12",
          "2.14": "2.14",
        } as Record<string, string>,
        isHigherClusterVersion: false,
      },
    });
  });
  afterEach(() => {
    vitestVi.clearAllMocks();
    vitestVi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(clusterInstallationStepCmp.text()).toContain("Run:ai version");
    expect(clusterInstallationStepCmp.text()).toContain("Cluster name");
  });

  test("renders the installation command returned by Cluster Service", async () => {
    await flushPromises();
    expect(clusterInstallationStepCmp.text()).toContain("backend-generated-installation-command");
  });

  test("renders cluster name summary correctly", async () => {
    await clusterInstallationStepCmp.setProps({
      clusterFormModel: {
        name: "test-cluster",
        version: undefined,
        prerequisites: {
          location: "",
          clusterUrl: "",
          k8sDistribution: "",
        },
      },
    });
    expect(clusterInstallationStepCmp.find(".name-section").text()).toContain("test-cluster");
  });

  test("renders run:ai version summary correctly", async () => {
    await clusterInstallationStepCmp.setProps({
      clusterFormModel: {
        name: "test-cluster",
        version: "2.10",
        prerequisites: {
          location: "",
          clusterUrl: "",
          k8sDistribution: "",
        },
      },
    });
    expect(clusterInstallationStepCmp.find(".version").text()).toContain("2.10");
  });

  test("renders settings summary correctly", async () => {
    await clusterInstallationStepCmp.setProps({
      clusterFormModel: {
        name: "test-cluster",
        version: "2.10",
        prerequisites: {
          location: EClusterLocations.CONTROL_PLANE,
          clusterUrl: "",
          k8sDistribution: EClusterDistIds.OPEN_SHIFT,
        },
      },
    });
    expect(clusterInstallationStepCmp.find(".settings").text()).toContain("Same as control plane");
    expect(clusterInstallationStepCmp.find(".settings").text()).toContain("OpenShift");
  });

  test("showCertificateInstructions certificate instructions should be true if location is set to 'remote' ", async () => {
    expect(clusterInstallationStepCmp.vm.showCertificateInstructions).toBe(false);
    await clusterInstallationStepCmp.setProps({
      clusterFormModel: {
        name: "test-cluster",
        version: "2.10",
        prerequisites: {
          location: EClusterLocations.REMOTE,
          clusterUrl: "",
          k8sDistribution: EClusterDistIds.OPEN_SHIFT,
        },
      },
    });
    expect(clusterInstallationStepCmp.vm.showCertificateInstructions).toBe(true);
  });

  test("showCertificateInstructions certificate instructions should be true if isMultiTenant", async () => {
    expect(clusterInstallationStepCmp.vm.showCertificateInstructions).toBe(false);
    await clusterInstallationStepCmp.setProps({ isMultiTenant: true });
    expect(clusterInstallationStepCmp.vm.showCertificateInstructions).toBe(true);
  });

  test("showCertificateInstructions certificate instructions should be false if cluster version is equal or higher than 2.14 is true and distribution is openshift ", async () => {
    await clusterInstallationStepCmp.setProps({
      clusterFormModel: {
        name: "test-cluster",
        version: "2.14",
        prerequisites: {
          location: EClusterLocations.REMOTE,
          clusterUrl: "",
          k8sDistribution: EClusterDistIds.OPEN_SHIFT,
        },
      },
      isHigherClusterVersion: true,
    });
    expect(clusterInstallationStepCmp.vm.showCertificateInstructions).toBe(false);
  });
});
