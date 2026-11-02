/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { createPinia } from "pinia";
import { Quasar } from "quasar";
import WorkloadActivityTab from "./workload-activity-tab.vue";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { intervalUtil } from "@/utils/interval.util";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiAdvancedFilters } from "@/components/common/runai-advanced-filters";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import { EIntervalLabels } from "@/models/interval.model";
import type { GetWorkloadActionsResponse, WorkloadActionRecord } from "@/swagger-models/workloads-client";

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: { getWorkloadActions: vi.fn() },
}));
vi.mock("@/utils/interval.util", () => ({ intervalUtil: { startInterval: vi.fn(), stopInterval: vi.fn() } }));
vi.mock("@/stores/user-display-name.store", () => ({
  useUserDisplayNameStore: () => ({
    isPending: () => false,
    getDisplayLabel: (name: string) => ({ primaryLabel: name }),
    resolveUsernames: vi.fn(),
  }),
}));

const action = (id = "action-1"): WorkloadActionRecord => ({
  id,
  workloadId: "w1",
  workloadName: "workload",
  subject: "admin@example.com",
  subjectType: "user",
  action: "create",
  acceptedAt: "2026-11-02T08:00:00Z",
});
const deferred = () => {
  let resolve!: (value: GetWorkloadActionsResponse) => void;
  const promise = new Promise<GetWorkloadActionsResponse>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};
let wrapper: VueWrapper | undefined;
const render = () => {
  wrapper = mount(WorkloadActivityTab, {
    props: { workloadId: "w1" },
    global: { plugins: [createPinia(), Quasar], directives: { permission: {}, "click-outside": {} } },
  });
  return wrapper;
};
const refresh = async () => {
  const call = vi
    .mocked(intervalUtil.startInterval)
    .mock.calls.find(([label]) => label === EIntervalLabels.WorkloadActivityTab);
  expect(call).toBeDefined();
  if (!call) throw new Error("Activity refresh interval was not started");
  await call[1]();
  await flushPromises();
};
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(workloadService.getWorkloadActions).mockResolvedValue({ actions: [action()], next: 50 });
});
afterEach(() => {
  wrapper?.unmount();
  wrapper = undefined;
  vi.useRealTimers();
});

describe("WorkloadActivityTab", () => {
  it("shows the initial loader, then exactly three columns and no actor-type column", async () => {
    const pending = deferred();
    vi.mocked(workloadService.getWorkloadActions).mockReturnValueOnce(pending.promise);
    const view = render();
    expect(view.find('[aid="workload-activity-loading"]').exists()).toBe(true);
    pending.resolve({ actions: [action()], next: 50 });
    await flushPromises();
    expect(view.findAll("thead th").map((cell) => cell.text().replace("arrow_upward", ""))).toEqual([
      "Date & time",
      "Event",
      "Subject",
    ]);
    expect(view.text()).toContain("Created");
    expect(view.text()).toContain("admin@example.com");
    expect(view.find('[aid="workload-activity-loading"]').exists()).toBe(false);
  });

  it("resets the server cursor when sorting and retains the sort on the next page", async () => {
    const view = render();
    await flushPromises();
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    view.findComponent(RunaiTable).vm.$emit("update-filters", { sortBy: "subject", descending: false });
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "subject",
      sortOrder: "asc",
      filterBy: [],
    });
    expect(view.get('[aid="workload-activity-range"]').text()).toBe("1–1");
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 50, 50, {
      sortBy: "subject",
      sortOrder: "asc",
      filterBy: [],
    });
  });

  it("sends combined filters to the server and distinguishes an empty filtered result", async () => {
    const view = render();
    await flushPromises();
    vi.mocked(workloadService.getWorkloadActions).mockResolvedValue({ actions: [] });
    view.findComponent(RunaiAdvancedFilters).vm.$emit("update-filter-by", {
      advancedFilters: [
        {
          name: "subject",
          label: "Subject",
          type: EColumnFilterType.FreeText,
          value: EFilterOperator.Contains,
          term: "missing",
        },
        {
          name: "action",
          label: "Event",
          type: EColumnFilterType.EnumBased,
          selectedValues: ["create"],
          selectOptions: [
            { label: "Created", value: "create" },
            { label: "Updated", value: "update" },
          ],
        },
      ],
    });
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: ["subject=@missing", "action!=update"],
    });
    expect(view.text()).toContain("No results");
    expect(view.text()).not.toContain("No activity has been recorded");
    expect(view.get('[aid="workload-activity-next"]').attributes("disabled")).toBeDefined();
  });

  it("reloads when adding a second filter or editing a saved filter", async () => {
    const view = render();
    await flushPromises();
    const filters = view.findComponent(RunaiAdvancedFilters);
    filters.vm.saveFilter({
      name: "subject",
      label: "Subject",
      type: EColumnFilterType.FreeText,
      value: EFilterOperator.Contains,
      term: "first",
    });
    await flushPromises();
    const calls = vi.mocked(workloadService.getWorkloadActions).mock.calls.length;
    filters.vm.saveFilter({
      name: "action",
      label: "Event",
      type: EColumnFilterType.EnumBased,
      selectedValues: ["create"],
      selectOptions: [
        { label: "Created", value: "create" },
        { label: "Updated", value: "update" },
      ],
    });
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenCalledTimes(calls + 1);
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: ["subject=@first", "action!=update"],
    });
    filters.vm.saveFilter({
      name: "subject",
      label: "Subject",
      type: EColumnFilterType.FreeText,
      value: EFilterOperator.Contains,
      term: "edited",
    });
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: ["subject=@edited", "action!=update"],
    });
  });

  it("searches from page one, combines filters, preserves search while paging and clears it", async () => {
    const view = render();
    await flushPromises();
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    view.findComponent(RunaiGeneralSearch).vm.$emit("update-search-term", "Created");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
      search: "Created",
    });
    expect(view.get('[aid="workload-activity-range"]').text()).toBe("1–1");
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 50, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
      search: "Created",
    });
    view.findComponent(RunaiAdvancedFilters).vm.saveFilter({
      name: "subject",
      label: "Subject",
      type: EColumnFilterType.FreeText,
      value: EFilterOperator.Contains,
      term: "admin",
    });
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: ["subject=@admin"],
      search: "Created",
    });
    await refresh();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: ["subject=@admin"],
      search: "Created",
    });
    view.findComponent(RunaiGeneralSearch).vm.$emit("update-search-term", "");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: ["subject=@admin"],
    });
    expect(view.findComponent(RunaiGeneralSearch).props("searchTerm")).toBe("");
  });

  it("discards a pending search when switching workloads", async () => {
    vi.useFakeTimers();
    const view = render();
    await flushPromises();
    await view.get('input[type="search"]').setValue("Created");
    await view.setProps({ workloadId: "w2" });
    await flushPromises();
    vi.advanceTimersByTime(400);
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w2", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
    expect(view.findComponent(RunaiGeneralSearch).props("searchTerm")).toBe("");
  });

  it("uses the next cursor for one page, supports going back and disables next at the end", async () => {
    const view = render();
    await flushPromises();
    vi.mocked(workloadService.getWorkloadActions).mockResolvedValueOnce({ actions: [action("page-two")] });
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 50, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
    expect(view.get('[aid="workload-activity-range"]').text()).toBe("51–51");
    expect(view.get('[aid="workload-activity-next"]').attributes("disabled")).toBeDefined();
    await view.get('[aid="workload-activity-previous"]').trigger("click");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
  });

  it("keeps the current page visible while refreshing and refreshes its offset", async () => {
    const view = render();
    await flushPromises();
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    const pending = deferred();
    vi.mocked(workloadService.getWorkloadActions).mockReturnValueOnce(pending.promise);
    const tick = refresh();
    await flushPromises();
    expect(view.text()).toContain("admin@example.com");
    expect(view.find('[aid="workload-activity-loading"]').exists()).toBe(false);
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 50, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
    pending.resolve({ actions: [{ ...action(), subject: "updated@example.com" }] });
    await tick;
    expect(view.text()).toContain("updated@example.com");
  });

  it("shows the empty state and disables pagination when no actions are available", async () => {
    vi.mocked(workloadService.getWorkloadActions).mockResolvedValue({ actions: [] });
    const view = render();
    await flushPromises();
    expect(view.text()).toContain("No activity has been recorded for this workload.");
    expect(view.get('[aid="workload-activity-next"]').attributes("disabled")).toBeDefined();
    expect(view.get('[aid="workload-activity-previous"]').attributes("disabled")).toBeDefined();
  });

  it("shows an initial error and recovers through Retry", async () => {
    vi.mocked(workloadService.getWorkloadActions).mockRejectedValueOnce(new Error("unavailable"));
    const view = render();
    await flushPromises();
    expect(view.get('[role="alert"]').text()).toContain("Activity could not be loaded.");
    await view.get('[aid="workload-activity-retry"]').trigger("click");
    await flushPromises();
    expect(view.text()).toContain("admin@example.com");
    expect(view.find('[role="alert"]').exists()).toBe(false);
  });

  it("retains rows on refresh failure and retries the failed page", async () => {
    const view = render();
    await flushPromises();
    vi.mocked(workloadService.getWorkloadActions).mockRejectedValueOnce(new Error("offline"));
    await refresh();
    expect(view.text()).toContain("admin@example.com");
    expect(view.get('[role="alert"]').text()).toContain("Activity could not be loaded.");
    await view.get('[aid="workload-activity-retry"]').trigger("click");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
    expect(view.find('[role="alert"]').exists()).toBe(false);
  });

  it("does not mislabel or clear the previous page when next fails", async () => {
    const view = render();
    await flushPromises();
    vi.mocked(workloadService.getWorkloadActions).mockRejectedValueOnce(new Error("offline"));
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    expect(view.get('[aid="workload-activity-range"]').text()).toBe("1–1");
    expect(view.text()).toContain("admin@example.com");
    await view.get('[aid="workload-activity-retry"]').trigger("click");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 50, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
    expect(view.get('[aid="workload-activity-range"]').text()).toBe("51–51");
  });

  it("resets pagination on workload changes and ignores the previous workload response", async () => {
    const pending = deferred();
    vi.mocked(workloadService.getWorkloadActions).mockReturnValueOnce(pending.promise);
    const view = render();
    vi.mocked(workloadService.getWorkloadActions).mockResolvedValueOnce({
      actions: [{ ...action("new"), subject: "new@example.com" }],
    });
    await view.setProps({ workloadId: "w2" });
    await flushPromises();
    pending.resolve({ actions: [action()] });
    await flushPromises();
    expect(view.text()).toContain("new@example.com");
    expect(view.text()).not.toContain("admin@example.com");
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w2", 0, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
  });

  it("preserves a failed page navigation for Retry when the refresh interval fires", async () => {
    const view = render();
    await flushPromises();
    vi.mocked(workloadService.getWorkloadActions).mockRejectedValueOnce(new Error("offline"));
    await view.get('[aid="workload-activity-next"]').trigger("click");
    await flushPromises();
    const calls = vi.mocked(workloadService.getWorkloadActions).mock.calls.length;
    await refresh();
    expect(workloadService.getWorkloadActions).toHaveBeenCalledTimes(calls);
    expect(view.find('[aid="workload-activity-retry"]').exists()).toBe(true);
    await view.get('[aid="workload-activity-retry"]').trigger("click");
    await flushPromises();
    expect(workloadService.getWorkloadActions).toHaveBeenLastCalledWith("w1", 50, 50, {
      sortBy: "acceptedAt",
      sortOrder: "desc",
      filterBy: [],
    });
  });

  it("provides a named keyboard-focusable region for scrolling the action table", async () => {
    const view = render();
    await flushPromises();
    const region = view.get('[role="region"][aria-label="Workload activity history"]');
    expect(region.attributes("tabindex")).toBe("0");
  });

  it("stops polling and discards pending data when unmounted", async () => {
    const pending = deferred();
    vi.mocked(workloadService.getWorkloadActions).mockReturnValueOnce(pending.promise);
    render().unmount();
    wrapper = undefined;
    pending.resolve({ actions: [action()] });
    await flushPromises();
    expect(intervalUtil.stopInterval).toHaveBeenCalledWith(EIntervalLabels.WorkloadActivityTab);
    expect(intervalUtil.startInterval).not.toHaveBeenCalled();
  });
});
