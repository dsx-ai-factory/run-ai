/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ReportPageActions } from ".";
import { reportIndexColumns, type IReportTable } from "@/table-models/report.table-model";
import type { IFilterBy } from "@/models/filter.model";
import { ReportStatus } from "@/swagger-models/org-unit-service-client";

function mountActions(selectedRowsAmount: number, status: ReportStatus = ReportStatus.Ready) {
  return mount(ReportPageActions, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { permission: vi.fn(), clickOutside: vi.fn() },
    },
    props: {
      selectedReport: { id: "id-1", name: "report-one", status } as IReportTable,
      selectedRowsAmount,
      filterBy: {} as IFilterBy,
      columns: reportIndexColumns,
    },
  });
}

describe("report-page-actions download button", () => {
  beforeEach(() => {
    setActivePinia(createTestingPinia({ createSpy: vi.fn }));
  });

  it("is enabled for a single Ready report", () => {
    const vm = mountActions(1).vm as unknown as { isDownloadBtnDisabled: boolean };
    expect(vm.isDownloadBtnDisabled).toBe(false);
  });

  it("is disabled once more than one report is selected", () => {
    const vm = mountActions(3).vm as unknown as {
      isDownloadBtnDisabled: boolean;
      downloadBtnTooltip: string;
    };
    expect(vm.isDownloadBtnDisabled).toBe(true);
    expect(vm.downloadBtnTooltip).toBe("To download, select a single report.");
  });

  it("stays disabled for a single report that is not Ready", () => {
    const vm = mountActions(1, ReportStatus.Pending).vm as unknown as {
      isDownloadBtnDisabled: boolean;
      downloadBtnTooltip: string;
    };
    expect(vm.isDownloadBtnDisabled).toBe(true);
    expect(vm.downloadBtnTooltip).toBe("To download, the selected report must be in status Ready.");
  });
});
