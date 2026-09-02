/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import ClusterPrereqSection from "./cluster-prerequisites-section.vue";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import { EClusterDistIds, EClusterLocations, type IPrerequisites } from "@/models/cluster.model";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
describe("cluster-prerequisites-section.vue", () => {
  let clusterPrereqSectionCmp: VueWrapper<InstanceType<typeof ClusterPrereqSection>>;

  beforeEach(() => {
    clusterPrereqSectionCmp = mount(ClusterPrereqSection, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        isHigherClusterVersion: false as boolean,
        isMultiTenant: false as boolean,
        prerequisites: {
          location: EClusterLocations.CONTROL_PLANE,
          clusterUrl: "",
          k8sDistribution: "",
        } as IPrerequisites,
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(clusterPrereqSectionCmp.html()).toContain("Settings");
    expect(clusterPrereqSectionCmp.classes()).toContain("settings-section");
  });

  describe("other k8s distribution section", () => {
    test("will not render the k8s distribtion radio buttons when isHigherClusterVersion is false", async () => {
      expect(clusterPrereqSectionCmp.find(".open-shift-options").exists()).toBe(false);
    });
  });

  describe("k8s distribution cards carousel", () => {
    test("will not render a k8s distribution cards carousel when isHigherClusterVersion is true", async () => {
      await clusterPrereqSectionCmp.setProps({ isHigherClusterVersion: true });
      expect(clusterPrereqSectionCmp.find(".dist-selection").exists()).toBe(false);
    });

    test("renders a k8s distribution cards carousel when isHigherClusterVersion is false", async () => {
      expect(clusterPrereqSectionCmp.find(".dist-selection").exists()).toBe(true);
    });

    test("emits an update:model-value event when selected distribution changes by the card carousel", async () => {
      //todo: validate why this test is failing, due to import of DEPARTMENT_ROUTE_NAMES in gpu-allocation-ratio
      await clusterPrereqSectionCmp.find("[aid=vanilla-kubernetes-card]").trigger("click");
      const expectedEmit: IPrerequisites[][] = [
        [{ location: EClusterLocations.CONTROL_PLANE, clusterUrl: "", k8sDistribution: EClusterDistIds.VANILLA }],
      ];
      expect(clusterPrereqSectionCmp.emitted("update:model-value")).toBeTruthy();
      expect(clusterPrereqSectionCmp.emitted("update:model-value")).toEqual(expectedEmit);
    });
  });

  describe("cluster location section tests", () => {
    test("will not render location options if isMultiTenant is true", async () => {
      expect(clusterPrereqSectionCmp.find(".cluster-location").exists()).toBeTruthy();
      await clusterPrereqSectionCmp.setProps({ isMultiTenant: true });
      expect(clusterPrereqSectionCmp.find(".cluster-location").exists()).toBeFalsy();
    });

    test("emits an update:model-value event when cluster location changes", async () => {
      await (clusterPrereqSectionCmp.findComponent(".location-options") as VueWrapper).vm.$emit(
        "update:model-value",
        EClusterLocations.REMOTE,
      );
      const expectedEmit: IPrerequisites[][] = [
        [{ location: EClusterLocations.REMOTE, clusterUrl: "", k8sDistribution: "" }],
      ];
      expect(clusterPrereqSectionCmp.emitted("update:model-value")).toBeTruthy();
      expect(clusterPrereqSectionCmp.emitted("update:model-value")).toEqual(expectedEmit);
    });
  });

  describe("cluster url section tests", () => {
    test("will render the cluster url section if cluster location is set to 'remote' OR isMultiTenant is set to true", async () => {
      expect(clusterPrereqSectionCmp.find(".k8s-url-cluster").exists()).toBeFalsy();

      await clusterPrereqSectionCmp.setProps({ isMultiTenant: true });
      expect(clusterPrereqSectionCmp.find(".k8s-url-cluster").exists()).toBeTruthy();

      await clusterPrereqSectionCmp.setProps({
        isMultiTenant: false,
        prerequisites: {
          location: EClusterLocations.REMOTE,
          clusterUrl: "",
          k8sDistribution: "",
        },
      });
      expect(clusterPrereqSectionCmp.find(".k8s-url-cluster").exists()).toBeTruthy();
    });

    test("emits an update:model-value event when cluster URL changes", async () => {
      await clusterPrereqSectionCmp.setProps({
        // revealing the url section
        prerequisites: {
          location: EClusterLocations.REMOTE,
          clusterUrl: "",
          k8sDistribution: "",
        },
      });

      await clusterPrereqSectionCmp.find("[aid=cluster-url-input]").setValue("https://test.com");
      const expectedEmit: IPrerequisites[][] = [
        [{ location: EClusterLocations.REMOTE, clusterUrl: "https://test.com", k8sDistribution: "" }],
      ];
      expect(clusterPrereqSectionCmp.emitted("update:model-value")).toBeTruthy();
      expect(clusterPrereqSectionCmp.emitted("update:model-value")).toEqual(expectedEmit);
    });
  });

  describe("sectionInvalid tests", () => {
    test("sets sectionInvalid to true when k8sDistribution is empty", () => {
      expect(clusterPrereqSectionCmp.vm.sectionInvalid).toBe(true);
    });

    test("sets sectionInvalid to false when k8sDistribution is selected", async () => {
      await clusterPrereqSectionCmp.setProps({
        prerequisites: {
          location: EClusterLocations.CONTROL_PLANE,
          clusterUrl: "",
          k8sDistribution: EClusterDistIds.VANILLA,
        },
      });
      expect(clusterPrereqSectionCmp.vm.sectionInvalid).toBe(false);
    });

    test("sets sectionInvalid to true when location is remote and clusterUrl is empty", async () => {
      await clusterPrereqSectionCmp.setProps({
        prerequisites: {
          location: EClusterLocations.REMOTE,
          clusterUrl: "",
          k8sDistribution: EClusterDistIds.VANILLA,
        },
      });
      expect(clusterPrereqSectionCmp.vm.sectionInvalid).toBe(true);
    });

    test("sets sectionInvalid to false when all prerequisites are provided", async () => {
      await clusterPrereqSectionCmp.setProps({
        prerequisites: {
          location: EClusterLocations.REMOTE,
          clusterUrl: "https://test.com",
          k8sDistribution: EClusterDistIds.VANILLA,
        },
      });
      expect(clusterPrereqSectionCmp.vm.sectionInvalid).toBe(false);
    });
  });
});
