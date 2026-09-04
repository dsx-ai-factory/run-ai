/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { Quasar } from "quasar";
import AiApplicationSourceSection from "./ai-application-source-section.vue";
import { aiApplicationService } from "@/services/cluster/ai-application.service/ai-application.service";
import { errorMessages } from "@/common/error-message.constant";

vi.mock("@/services/control-plane/catalog.service/catalog.service", () => ({
  catalogService: {
    getHelmRepos: vi.fn().mockResolvedValue([]),
    browseHelmRepoCharts: vi.fn().mockResolvedValue({ charts: [] }),
  },
}));

vi.mock("@/services/cluster/ai-application.service/ai-application.service", () => ({
  aiApplicationService: {
    uploadChart: vi.fn().mockResolvedValue("chart-123"),
  },
}));

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Vm = any;

const mountComponent = () =>
  shallowMount(AiApplicationSourceSection, {
    props: {
      chartRef: { repo: "", chartName: "", chartVersion: "", chartURL: "" },
      yamlContent: "",
      clusterId: "cluster-1",
      projectId: 1,
    },
    global: {
      plugins: [Quasar],
      stubs: {
        RunaiExpansionItem: { template: "<div><slot /></div>" },
        RunaiRadioOptions: true,
        RunaiSection: { template: "<div><slot /></div>" },
        RunaiSelect: true,
        RunaiEditor: true,
        RunaiCredentialSelectorBySource: true,
        AiApplicationChartUploadField: true,
        QInput: true,
      },
    },
  });

function makeFile(name: string, size: number): File {
  const file = new File([new Uint8Array(1)], name);
  Object.defineProperty(file, "size", { value: size });
  return file;
}

function lastChartRefEmit(wrapper: ReturnType<typeof mountComponent>): Record<string, string> {
  const emits = wrapper.emitted("update:chartRef");
  return (emits?.[emits.length - 1]?.[0] ?? {}) as Record<string, string>;
}

describe("ai-application-source-section upload file", () => {
  beforeEach(() => vi.clearAllMocks());

  it("offers Upload file as a source option", () => {
    const vm = mountComponent().vm as Vm;
    expect(vm.aiApplicationSourceOptions.map((o: { value: string }) => o.value)).toContain("upload-file");
  });

  it("uploads a selected file and syncs the returned chart id", async () => {
    const wrapper = mountComponent();
    const vm = wrapper.vm as Vm;

    vm.handleSourceChange("upload-file");
    await vm.handleChartFileSelected(makeFile("chart.tgz", 2));

    expect(aiApplicationService.uploadChart).toHaveBeenCalledTimes(1);
    expect(vm.uploadedChartId).toBe("chart-123");
    expect(lastChartRefEmit(wrapper).chartId).toBe("chart-123");
  });

  it("surfaces a server upload failure without setting a chart id", async () => {
    vi.mocked(aiApplicationService.uploadChart).mockRejectedValueOnce(new Error("boom"));
    const wrapper = mountComponent();
    const vm = wrapper.vm as Vm;

    vm.handleSourceChange("upload-file");
    await vm.handleChartFileSelected(makeFile("chart.tgz", 2));

    expect(vm.uploadedChartId).toBe("");
    expect(vm.chartUploadError).toBe(errorMessages.CHART_UPLOAD_FAILED);
  });

  it("flags an oversized rejection as too large without uploading", () => {
    const wrapper = mountComponent();
    const vm = wrapper.vm as Vm;

    vm.handleChartFileRejected([{ failedPropValidation: "max-file-size", file: makeFile("big.tgz", 11 * 1024 * 1024) }]);

    expect(aiApplicationService.uploadChart).not.toHaveBeenCalled();
    expect(vm.chartUploadError).toBe(errorMessages.CHART_FILE_TOO_LARGE);
  });

  it("preserves the uploaded file and restores chartId when returning to upload-file", async () => {
    const wrapper = mountComponent();
    const vm = wrapper.vm as Vm;

    vm.handleSourceChange("upload-file");
    await vm.handleChartFileSelected(makeFile("chart.tgz", 2));
    expect(vm.uploadedChartId).toBe("chart-123");
    expect(lastChartRefEmit(wrapper).chartId).toBe("chart-123");

    vm.handleSourceChange("custom-url");
    expect(vm.selectedFile).not.toBeNull();
    expect(vm.uploadedChartId).toBe("chart-123");
    expect(lastChartRefEmit(wrapper).chartId).toBe("");

    vm.handleSourceChange("upload-file");
    expect(lastChartRefEmit(wrapper).chartId).toBe("chart-123");
  });
});
