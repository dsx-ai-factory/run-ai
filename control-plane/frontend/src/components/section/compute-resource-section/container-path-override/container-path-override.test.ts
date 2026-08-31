/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import ContainerPathOverride from "./container-path-override.vue";
import type { ISelectedDatasourcesContainerPathData } from "./container-path-override.model";
import { Quasar } from "quasar";
import { describe, expect, it, vi } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { AssetKind } from "@/swagger-models/assets-service-client";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("container-path-override.vue", () => {
  const dataSourcesContainerPathData: ISelectedDatasourcesContainerPathData[] = [
    { id: "1", kind: "git", name: "DataSource 1", origin: "/origin/path1" },
    { id: "2", kind: "nfs", name: "DataSource 2", origin: "/origin/path2" },
  ];

  const defaultProps = {
    dataSourcesContainerPathData,
  };

  const factory = (props = defaultProps) => {
    return mount(ContainerPathOverride, {
      global: {
        plugins: [Quasar],
      },
      props: { ...props },
    });
  };

  it("renders correctly", () => {
    const wrapper = factory();
    expect(wrapper.exists()).toBe(true);
  });

  it("sets the selected data source id on creation", () => {
    const wrapper = factory();
    expect(wrapper.vm.selectedDatasourceId).toBe(dataSourcesContainerPathData[0].id);
  });

  it("renders data source tabs correctly", () => {
    const wrapper = factory();
    const tabs = wrapper.findAll(".container-path-tabs .q-tab");
    expect(tabs.length).toBe(dataSourcesContainerPathData.length);
  });

  it("renders container path input with the correct value", async () => {
    const wrapper = factory();
    await wrapper.setData({ selectedDatasourceId: "1" });
    const input = wrapper.find("[aid='container-path-input']");
    expect((input.element as HTMLInputElement).value).toBe(dataSourcesContainerPathData[0].origin);
  });

  it("emits update-container-path-override event when input value changes", async () => {
    const wrapper = factory();
    await wrapper.setData({ selectedDatasourceId: "1" });
    const input = wrapper.find("[aid='container-path-input']");
    input.setValue("/new/path");
    expect(wrapper.emitted("update-container-path-override")).toBeTruthy();
    expect(wrapper.emitted("update-container-path-override")?.[0]?.[0]).toEqual({
      dataSourceId: "1",
      containerPathOverride: "/new/path",
    });
  });

  it("shows validation error icon when the path is invalid", async () => {
    const wrapper = factory({
      ...defaultProps,
      dataSourcesContainerPathData: [
        {
          id: "1",
          kind: "hostPath",
          name: "DataSource 1",
          origin: "/origin/path1",
          overrides: { containerPath: "not-a-valid-path" },
        },
      ],
    });
    await wrapper.setData({ selectedDatasourceId: "1" });
    const icon = wrapper.find(".ph-warning-circle");
    expect(icon.exists()).toBe(true);
  });

  it("computes isValidationError correctly", async () => {
    const wrapper = factory({
      ...defaultProps,
      dataSourcesContainerPathData: [
        { id: "1", kind: "hostPath", name: "DataSource 1", origin: "/origin/path1", overrides: { containerPath: "" } },
        {
          id: "2",
          kind: "hostPath",
          name: "DataSource 2",
          origin: "/origin/path2",
          overrides: { containerPath: "/valid/path" },
        },
      ],
    });
    expect(wrapper.vm.isValidationError).toBe(true);
  });

  it("emits is-section-invalid event when validation error changes", async () => {
    const wrapper = factory();
    await wrapper.setData({ selectedDatasourceId: "1" });
    expect(wrapper.emitted("is-section-invalid")).toBeTruthy();
  });

  it("emits is-section-invalid event when validation error changes", async () => {
    const wrapper = factory();
    await wrapper.setData({ selectedDatasourceId: "1" });
    expect(wrapper.emitted("is-section-invalid")).toBeTruthy();
  });

  it("updates selectedDatasourceId and dsMapByIds when dataSourcesContainerPathData changes", async () => {
    const wrapper = factory({ dataSourcesContainerPathData });

    const newDataSource = { id: "3", name: "DataSource 3", origin: "/origin/path3", kind: AssetKind.Git };
    await wrapper.setProps({
      dataSourcesContainerPathData: [...dataSourcesContainerPathData, newDataSource],
    });

    expect(wrapper.vm.selectedDatasourceId).toBe("3");
    expect(wrapper.vm.dsMapByIds).toHaveProperty("3", newDataSource);
  });

  it("selects the first data source if the selected data source is removed", async () => {
    const wrapper = factory({ dataSourcesContainerPathData });

    await wrapper.setData({ selectedDatasourceId: "2" });
    await wrapper.setProps({ dataSourcesContainerPathData: [dataSourcesContainerPathData[0]] });

    expect(wrapper.vm.selectedDatasourceId).toBe("1");
  });

  it("updates dsMapByIds correctly when dataSourcesContainerPathData changes", async () => {
    const wrapper = factory({ dataSourcesContainerPathData });

    const newDataSource = { id: "3", name: "DataSource 3", origin: "/origin/path3", kind: AssetKind.Git };
    await wrapper.setProps({
      dataSourcesContainerPathData: [...dataSourcesContainerPathData, newDataSource],
    });

    expect(wrapper.vm.dsMapByIds).toHaveProperty("3", newDataSource);
    expect(wrapper.vm.dsMapByIds).toHaveProperty("1", dataSourcesContainerPathData[0]);
    expect(wrapper.vm.dsMapByIds).toHaveProperty("2", dataSourcesContainerPathData[1]);
  });

  it("emits is-section-invalid event when isValidationError changes", async () => {
    const invalidDataSource: ISelectedDatasourcesContainerPathData = {
      id: "3",
      name: "DataSource 3",
      origin: "/origin/path3",
      kind: "hostPath",
      overrides: { containerPath: "" },
    };
    const wrapper = factory({
      dataSourcesContainerPathData: [invalidDataSource, ...dataSourcesContainerPathData],
    });

    expect(wrapper.emitted("is-section-invalid")).toBeTruthy();
    expect(wrapper.emitted("is-section-invalid")?.[0]).toEqual([true]);
  });
});
