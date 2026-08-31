/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import ClusterCreationStep from "./cluster-creation-step.vue";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";

import { Quasar } from "quasar";
import {
  clusterNewInstallerThreshold,
  EClusterDistIds,
  EClusterLocations,
  type IClusterInstalltionModel,
  type IPrerequisites,
} from "@/models/cluster.model";
import type { DefineComponent } from "vue";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
describe("cluster-creation-step.vue", () => {
  let clusterCreationStepCmp: VueWrapper<InstanceType<typeof ClusterCreationStep>>;

  async function updateComponentPrerequisites(newPrereq: IPrerequisites): Promise<void> {
    const model = clusterCreationStepCmp.vm.$props.clusterFormModel;
    await clusterCreationStepCmp.setProps({
      clusterFormModel: {
        ...model,
        prerequisites: newPrereq,
      },
    });
  }
  async function updateComponentVersion(newVersion: string): Promise<void> {
    const model = clusterCreationStepCmp.vm.$props.clusterFormModel;
    await clusterCreationStepCmp.setProps({
      clusterFormModel: {
        ...model,
        version: newVersion,
        prerequisites: { ...model.prerequisites, location: EClusterLocations.CONTROL_PLANE },
      },
    });
  }

  beforeEach(() => {
    clusterCreationStepCmp = mount(ClusterCreationStep, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        isMultiTenant: false as boolean,
        isExistingCluster: false as boolean,
        isClusterEditMode: false,
        clusterFormModel: {
          name: "",
          version: "",
          prerequisites: {
            location: "",
            clusterUrl: "",
            k8sDistribution: "",
          },
        } as IClusterInstalltionModel,
        versions: ["2.7", "2.8", "2.9", "2.10"] as string[],
        isHigherClusterVersion: false,
      },
    });
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(clusterCreationStepCmp.text()).toContain("Run:ai version");
    expect(clusterCreationStepCmp.text()).toContain("Cluster name");
  });

  test("component updates cluster name when a new name is emitted by the name section", async () => {
    clusterCreationStepCmp.findComponent<DefineComponent>(".cluster-name").vm.$emit("update:modelValue", "new-name");
    const expectedEmit: IClusterInstalltionModel[][] = [
      [
        {
          name: "new-name",
          version: "",
          prerequisites: {
            location: "",
            clusterUrl: "",
            k8sDistribution: "",
          },
        },
      ],
    ];
    expect(clusterCreationStepCmp.emitted("update-form-model")).toBeTruthy();
    expect(clusterCreationStepCmp.emitted("update-form-model")).toEqual(expectedEmit);
  });

  test("updates version when a new version is emitted by the version section and will set location to 'controlPlane' if isMultiTenant is false", () => {
    clusterCreationStepCmp.findComponent<DefineComponent>(".runai-version").vm.$emit("update:model-value", "2.10");

    const expectedEmit: IClusterInstalltionModel[][] = [
      [
        {
          name: "",
          version: "2.10",
          prerequisites: {
            location: EClusterLocations.CONTROL_PLANE,
            clusterUrl: "",
            k8sDistribution: "",
          },
        },
      ],
    ];
    expect(clusterCreationStepCmp.emitted("update-form-model")).toBeTruthy();
    expect(clusterCreationStepCmp.emitted("update-form-model")).toEqual(expectedEmit);
  });

  test("when a new version is emitted by the version section and will set location to 'remote' if isMultiTenant is true", async () => {
    await clusterCreationStepCmp.setProps({ isMultiTenant: true });
    clusterCreationStepCmp.findComponent<DefineComponent>(".runai-version").vm.$emit("update:model-value", "2.10");

    const expectedEmit: IClusterInstalltionModel[][] = [
      [
        {
          name: "",
          version: "2.10",
          prerequisites: {
            location: EClusterLocations.REMOTE,
            clusterUrl: "",
            k8sDistribution: "",
          },
        },
      ],
    ];
    expect(clusterCreationStepCmp.emitted("update-form-model")).toBeTruthy();
    expect(clusterCreationStepCmp.emitted("update-form-model")).toEqual(expectedEmit);
  });

  test("when a new version is emitted by the version section the component will reset the selected k8s distribution", async () => {
    await updateComponentPrerequisites({
      location: EClusterLocations.CONTROL_PLANE,
      clusterUrl: "",
      k8sDistribution: EClusterDistIds.VANILLA,
    });

    clusterCreationStepCmp.findComponent<DefineComponent>(".runai-version").vm.$emit("update:model-value", "2.10");

    const expectedEmit: IClusterInstalltionModel[][] = [
      [
        {
          name: "",
          version: "2.10",
          prerequisites: {
            location: EClusterLocations.CONTROL_PLANE,
            clusterUrl: "",
            k8sDistribution: "",
          },
        },
      ],
    ];
    expect(clusterCreationStepCmp.emitted("update-form-model")).toBeTruthy();
    expect(clusterCreationStepCmp.emitted("update-form-model")).toEqual(expectedEmit);
  });

  test("component will set k8sDistribution to 'openshift' by default if clusterVersion is equal or higher than clusterHigherVersion", async () => {
    await updateComponentPrerequisites({
      location: EClusterLocations.CONTROL_PLANE,
      clusterUrl: "",
      k8sDistribution: "",
    });

    clusterCreationStepCmp
      .findComponent<DefineComponent>(".runai-version")
      .vm.$emit("update:model-value", clusterNewInstallerThreshold);

    const expectedEmit: IClusterInstalltionModel[][] = [
      [
        {
          name: "",
          version: clusterNewInstallerThreshold,
          prerequisites: {
            location: EClusterLocations.CONTROL_PLANE,
            clusterUrl: "",
            k8sDistribution: EClusterDistIds.OPEN_SHIFT,
          },
        },
      ],
    ];
    expect(clusterCreationStepCmp.emitted("update-form-model")).toBeTruthy();
    expect(clusterCreationStepCmp.emitted("update-form-model")).toEqual(expectedEmit);
  });

  test("component will not render prerequisites section unless a version has been selected", async () => {
    expect(clusterCreationStepCmp.findComponent<DefineComponent>(".settings-section").exists()).toBeFalsy();

    await updateComponentVersion("2.10");

    expect(clusterCreationStepCmp.findComponent<DefineComponent>(".settings-section").exists()).toBeTruthy();
  });

  test("will emit the update prerequisites when it is emited by prerequisites section", async () => {
    await updateComponentVersion("2.10"); // revealing prerequsities section

    const newPrerequisites = {
      location: EClusterLocations.REMOTE,
      clusterUrl: "https://test.com",
      k8sDistribution: EClusterDistIds.OPEN_SHIFT,
    } as IPrerequisites;

    clusterCreationStepCmp
      .findComponent<DefineComponent>(".settings-section")
      .vm.$emit("update:model-value", newPrerequisites);

    const expectedEmit: IClusterInstalltionModel[][] = [
      [
        {
          name: "",
          version: "2.10",
          prerequisites: newPrerequisites,
        },
      ],
    ];

    expect(clusterCreationStepCmp.emitted("update-form-model")).toBeTruthy();
    expect(clusterCreationStepCmp.emitted("update-form-model")).toEqual(expectedEmit);
  });
});
