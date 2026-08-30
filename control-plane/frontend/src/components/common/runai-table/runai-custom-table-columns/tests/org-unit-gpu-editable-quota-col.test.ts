import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, type VueWrapper, flushPromises } from "@vue/test-utils";
import { nextTick } from "vue";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import OrgUnitGpuEditableQuotaCol from "../org-unit-gpu-editable-quota-col.vue";
import { EditableCellPopup } from "../editable-cell-popup";
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell } from "@/models/table.model";
import type { ICustomCellGpuQuotaColConfig, IQuotaTableRow } from "@/table-models/org-unit.table-model";
import { EQuotaInputState } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";
import type { Department } from "@/swagger-models/org-unit-service-client";

// Mock orgUnitService
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    getDepartment: vi.fn(),
  },
}));

import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// Mock the table format util
vi.mock("@/utils/table-format.util/table-format.util", () => ({
  getCustomCellConfig: (column: ITableColumn, cellType: ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL) => {
    return column.customCellConfig?.[cellType] as ICustomCellGpuQuotaColConfig | undefined;
  },
  tableFormatUtil: {
    getCustomCellConfig: (column: ITableColumn, cellType: ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL) => {
      return column.customCellConfig?.[cellType] as ICustomCellGpuQuotaColConfig | undefined;
    },
  },
}));

describe("OrgUnitGpuEditableQuotaCol", () => {
  let onSaveMock: ReturnType<typeof vi.fn>;

  const createColumn = (config: Partial<ICustomCellGpuQuotaColConfig> = {}): ITableColumn => {
    onSaveMock = vi.fn();
    const defaultConfig: ICustomCellGpuQuotaColConfig = {
      availableQuota: 50,
      totalQuota: 100,
      nodePoolName: "default",
      nodePoolId: "pool-1",
      entityType: EQuotaEntity.project,
      onSave: onSaveMock,
      ...config,
    };

    return {
      name: "gpuQuota",
      label: "GPUs",
      field: "gpuQuota",
      customCell: ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL,
      customCellConfig: {
        [ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL]: defaultConfig,
      },
    };
  };

  const createRow = (overrides: Partial<IQuotaTableRow> = {}): IQuotaTableRow => ({
    id: "project-1",
    name: "project-1",
    gpuQuota: 10,
    quotaDistribution: 10,
    allocatedNonPreemptibleGpu: 5,
    serverGpuQuota: 10,
    serverGpuLimit: 10, // Default to same as quota
    serverOverQuotaWeight: 1, // Default over-quota weight
    ...overrides,
  });

  const createWrapper = (
    row: IQuotaTableRow = createRow(),
    column: ITableColumn = createColumn(),
    isOverSubscriptionEnabled = false,
  ): VueWrapper<InstanceType<typeof OrgUnitGpuEditableQuotaCol>> => {
    return mount(OrgUnitGpuEditableQuotaCol, {
      global: {
        plugins: [
          Quasar,
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              Setting: {
                limitQuotaOverSubscription: { value: isOverSubscriptionEnabled },
              },
            },
          }),
        ],
        components: {
          EditableCellPopup,
        },
      },
      props: {
        row,
        column,
      },
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    it("renders quota button with display value", () => {
      const wrapper = createWrapper();
      const button = wrapper.find(".cell-trigger");

      expect(button.exists()).toBe(true);
      expect(button.text()).toBe("10");
    });

    it("displays dash when quota is null", () => {
      const row = createRow({ gpuQuota: null });
      const wrapper = createWrapper(row);

      expect(wrapper.find(".cell-trigger").text()).toBe("Unlimited");
    });

    it("opens menu when button is clicked", async () => {
      const wrapper = createWrapper();
      const button = wrapper.find(".cell-trigger");

      await button.trigger("click");
      await nextTick();

      expect(wrapper.findComponent({ name: "q-menu" }).exists()).toBe(true);
      // isOpen is now on the child EditableCellPopup component
      const editablePopup = wrapper.findComponent({ name: "EditableCellPopup" });
      expect(editablePopup.vm.isOpen).toBe(true);
    });
  });

  describe("Parent Limit Calculation (via validation messages)", () => {
    it("uses correct max assignable quota in validation message", async () => {
      // Available: 50, Current (gpuQuota): 10 -> Max: 60
      const row = createRow({ gpuQuota: 10, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Message should reference the max limit of 60
      expect(wrapper.vm.quotaMessage).toContain("60");
    });

    it("uses total quota as fallback when available is null", async () => {
      const column = createColumn({ availableQuota: null });
      const row = createRow({ allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row, column);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Message should reference the total of 100
      expect(wrapper.vm.quotaMessage).toContain("100");
    });

    it("uses current gpuQuota (not serverGpuQuota) for calculation", async () => {
      // Current (gpuQuota): 15, availableQuota: 50 -> Max: 65
      // serverGpuQuota: 20 (original) should NOT be used
      const row = createRow({ serverGpuQuota: 20, gpuQuota: 15, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Message should reference the max of 65 (not 70)
      expect(wrapper.vm.quotaMessage).toContain("65");
    });

    it("handles zero available quota correctly", async () => {
      // Available: 0, Current (gpuQuota): 10 -> Max: 10
      const column = createColumn({ availableQuota: 0 });
      const row = createRow({ gpuQuota: 10, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row, column);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Message should reference the max of 10
      expect(wrapper.vm.quotaMessage).toContain("10");
    });

    it("handles negative available with current quota correctly", async () => {
      // Available: -2, Current (gpuQuota): 4 -> Max: 2
      // This tests the bug fix: should use gpuQuota (4) not serverGpuQuota (2)
      const column = createColumn({ availableQuota: -2 });
      const row = createRow({ serverGpuQuota: 2, gpuQuota: 4, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row, column);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Message should reference the max of 2 (Math.max(0, -2 + 4))
      expect(wrapper.vm.quotaMessage).toContain("2");
    });
  });

  describe("Reactive Updates", () => {
    it("updates validation message when availableQuota changes", async () => {
      const column = createColumn({ availableQuota: 50 });
      const wrapper = createWrapper(createRow({ gpuQuota: 10, allocatedNonPreemptibleGpu: 0 }), column);

      // Open the menu to see the validation message
      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Initial message should reference 60 (50 + 10)
      expect(wrapper.vm.quotaMessage).toContain("60");

      // Update availableQuota to 20
      await wrapper.setProps({
        column: createColumn({ availableQuota: 20 }),
      });
      await nextTick();

      // New message should reference 30 (20 + 10)
      expect(wrapper.vm.quotaMessage).toContain("30");
    });

    it("updates row quota value when prop changes", async () => {
      const wrapper = createWrapper();

      expect(wrapper.vm.editableValue).toBe(10);

      await wrapper.setProps({
        row: createRow({ gpuQuota: 25 }),
      });
      await nextTick();

      expect(wrapper.vm.editableValue).toBe(25);
    });

    it("updates display value when row quota changes", async () => {
      const wrapper = createWrapper();

      expect(wrapper.find(".cell-trigger").text()).toBe("10");

      await wrapper.setProps({
        row: createRow({ gpuQuota: 35 }),
      });
      await nextTick();

      expect(wrapper.find(".cell-trigger").text()).toBe("35");
    });
  });

  describe("Validation", () => {
    it("shows default state for valid quota", async () => {
      const wrapper = createWrapper();
      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
    });

    it("shows warning state when quota exceeds parent limit", async () => {
      const column = createColumn({ availableQuota: 10 });
      const row = createRow({ gpuQuota: 5, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row, column);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Set to 20 (exceeds max of 15 = 10 + 5)
      wrapper.vm.editableValue = 20;
      await nextTick();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Warning);
    });

    it("shows warning state when quota is less than allocated non-preemptible", async () => {
      const row = createRow({ gpuQuota: 10, allocatedNonPreemptibleGpu: 15 });
      const wrapper = createWrapper(row);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Warning);
    });

    it("applies correct wrapper class based on state", async () => {
      const wrapper = createWrapper();
      const cellWrapper = wrapper.find(".editable-cell-popup");
      const trigger = wrapper.find(".cell-trigger");

      // Default state - wrapper should not have error/warning classes
      expect(cellWrapper.classes()).not.toContain("cell--error");
      expect(cellWrapper.classes()).not.toContain("cell--warning");

      // Open menu (active state)
      await trigger.trigger("click");
      await nextTick();

      expect(cellWrapper.classes()).toContain("cell--active");
    });
  });

  describe.only("Apply Functionality", () => {
    it("calls onSave with correct payload when apply is clicked", async () => {
      const wrapper = createWrapper();

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Change value
      wrapper.vm.editableValue = 25;
      await nextTick();

      // Directly call the exposed method (more reliable than finding the button in q-menu)
      wrapper.vm.handleApply();
      await nextTick();

      expect(onSaveMock).toHaveBeenCalledWith({
        projectId: "project-1",
        entityName: "project-1",
        nodePoolName: "default",
        nodePoolId: "pool-1",
        gpuQuota: 25,
        originalValue: 10,
        originalLimit: 10,
        originalOverQuotaWeight: 1,
      });
    });

    it("does not call onSave when apply is disabled", async () => {
      const column = createColumn({ availableQuota: 5 });
      const row = createRow({ gpuQuota: 0, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row, column, true); // Over-subscription enabled = error state

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Set to invalid value (exceeds max of 5 = 5 + 0)
      wrapper.vm.editableValue = 10;
      await nextTick();

      expect(wrapper.vm.hasError).toBe(true);

      // Try to apply
      wrapper.vm.handleApply();
      await nextTick();

      expect(onSaveMock).not.toHaveBeenCalled();
    });

    it("closes menu after successful apply", async () => {
      const wrapper = createWrapper();
      const editablePopup = wrapper.findComponent({ name: "EditableCellPopup" });

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      expect(editablePopup.vm.isOpen).toBe(true);

      wrapper.vm.editableValue = 20;
      await nextTick();

      // Call the popup's handleApplyClick which emits apply and closes the popup
      editablePopup.vm.handleApplyClick();
      await nextTick();

      // The popup should close after apply
      expect(editablePopup.vm.isOpen).toBe(false);
    });

    it("calls onSave with correct originalValue", async () => {
      const row = createRow({
        gpuQuota: 15,
        serverGpuQuota: 12,
        serverGpuLimit: 15,
        serverOverQuotaWeight: 2,
      });
      const wrapper = createWrapper(row);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      wrapper.vm.editableValue = 30;
      await nextTick();

      wrapper.vm.handleApply();
      await nextTick();

      expect(onSaveMock).toHaveBeenCalledWith({
        projectId: "project-1",
        entityName: "project-1",
        nodePoolName: "default",
        nodePoolId: "pool-1",
        gpuQuota: 30,
        originalValue: 12, // Uses serverGpuQuota
        originalLimit: 15,
        originalOverQuotaWeight: 2,
      });
    });
  });

  describe("Reset Functionality", () => {
    it("resets to original value when menu is closed without apply", async () => {
      const wrapper = createWrapper();
      const editablePopup = wrapper.findComponent({ name: "EditableCellPopup" });

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Change value
      wrapper.vm.editableValue = 50;
      await nextTick();

      // Close without applying - isOpen is on child component
      editablePopup.vm.isOpen = false;
      await nextTick();

      // Manually trigger handleReset (simulating menu hide event)
      wrapper.vm.handleReset();
      await nextTick();

      expect(wrapper.vm.editableValue).toBe(10);
    });
  });

  describe("Edge Cases", () => {
    it("handles undefined allocatedNonPreemptibleGpu", async () => {
      const row = createRow({ allocatedNonPreemptibleGpu: undefined });
      const wrapper = createWrapper(row);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Should not throw error and should have a valid state
      expect(wrapper.vm.quotaState).toBeDefined();
    });

    it("handles null gpuQuota gracefully", async () => {
      const row = createRow({ gpuQuota: null, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Should use 0 as fallback - message should reference 50 (50 + 0)
      expect(wrapper.vm.quotaMessage).toContain("50");
    });

    it("handles negative availableQuota with Math.max", async () => {
      const column = createColumn({ availableQuota: -10 });
      const row = createRow({ gpuQuota: 5, allocatedNonPreemptibleGpu: 0 });
      const wrapper = createWrapper(row, column);

      await wrapper.find(".cell-trigger").trigger("click");
      await nextTick();

      // Math.max(0, -10 + 5) = 0, so validation should show that
      // When setting to 1, it should be over the limit (0)
      wrapper.vm.editableValue = 1;
      await nextTick();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Warning);
    });
  });

  describe("Department Subordinate Quota Validation", () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    const createDepartmentMock = (subordinateGpuQuota: number | null): Department => ({
      id: "dept-1",
      name: "Department 1",
      clusterId: "cluster-1",
      description: "",
      createdAt: "2024-01-01",
      updatedAt: "2024-01-01",
      projectsAggregatedResources: [
        {
          nodePool: { id: "pool-1", name: "default" },
          gpu: { deserved: subordinateGpuQuota },
        },
      ],
      resources: [],
      totalResources: { gpuQuota: 0 },
    });

    it("fetches subordinate quota when popup opens for departments", async () => {
      const mockDepartment = createDepartmentMock(20);
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment);

      const column = createColumn({ entityType: EQuotaEntity.department });
      const row = createRow({ id: "dept-1", gpuQuota: 30 });

      const wrapper = createWrapper(row, column);

      // Should not have fetched yet
      expect(orgUnitService.getDepartment).not.toHaveBeenCalled();

      // Open the popup
      await wrapper.find(".cell-trigger").trigger("click");
      // Manually trigger the opened event (QMenu's @show might not fire in tests)
      wrapper.findComponent({ name: "EditableCellPopup" }).vm.$emit("opened");
      await flushPromises();

      expect(orgUnitService.getDepartment).toHaveBeenCalledWith("dept-1");
      expect(wrapper.vm.subordinateGpuQuota).toBe(20);
      expect(wrapper.vm.isLoadingDepartment).toBe(false);
    });

    it("does not fetch for projects", async () => {
      const column = createColumn({ entityType: EQuotaEntity.project });
      const row = createRow({ id: "project-1", gpuQuota: 10 });

      const wrapper = createWrapper(row, column);

      // Open the popup
      await wrapper.find(".cell-trigger").trigger("click");
      await flushPromises();

      // Should still not fetch for projects
      expect(orgUnitService.getDepartment).not.toHaveBeenCalled();
    });

    it("shows error when department quota is below subordinate quota", async () => {
      const mockDepartment = createDepartmentMock(30);
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment);

      const column = createColumn({ entityType: EQuotaEntity.department });
      const row = createRow({ id: "dept-1", gpuQuota: 20 });

      // Enable over-subscription to get ERROR state for subordinate quota violations
      const wrapper = createWrapper(row, column, true);

      // Open the popup - this triggers the fetch
      await wrapper.find(".cell-trigger").trigger("click");
      // Manually trigger the opened event (QMenu's @show might not fire in tests)
      wrapper.findComponent({ name: "EditableCellPopup" }).vm.$emit("opened");
      await flushPromises();

      // Set quota below subordinate (20 < 30)
      wrapper.vm.editableValue = 20;
      await nextTick();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.hasError).toBe(true);
      expect(wrapper.vm.quotaMessage).toContain("subordinate");
    });

    it("allows saving when department quota is above subordinate quota", async () => {
      const mockDepartment = createDepartmentMock(20);
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment);

      const column = createColumn({ entityType: EQuotaEntity.department });
      const row = createRow({ id: "dept-1", gpuQuota: 30 });

      const wrapper = createWrapper(row, column);

      // Open the popup - this triggers the fetch
      await wrapper.find(".cell-trigger").trigger("click");
      // Manually trigger the opened event (QMenu's @show might not fire in tests)
      wrapper.findComponent({ name: "EditableCellPopup" }).vm.$emit("opened");
      await flushPromises();

      // Set quota above subordinate (30 > 20)
      wrapper.vm.editableValue = 30;
      await nextTick();

      expect(wrapper.vm.quotaState).not.toBe(EQuotaInputState.Error);
      expect(wrapper.vm.hasError).toBe(false);
    });

    it("handles fetch failure gracefully", async () => {
      vi.mocked(orgUnitService.getDepartment).mockRejectedValue(new Error("Failed to fetch department"));

      const column = createColumn({ entityType: EQuotaEntity.department });
      const row = createRow({ id: "dept-1", gpuQuota: 30 });

      const wrapper = createWrapper(row, column);

      // Open the popup - this triggers the fetch which will fail
      await wrapper.find(".cell-trigger").trigger("click");
      // Manually trigger the opened event (QMenu's @show might not fire in tests)
      wrapper.findComponent({ name: "EditableCellPopup" }).vm.$emit("opened");
      await flushPromises();

      expect(wrapper.vm.isLoadingDepartment).toBe(false);
      expect(wrapper.vm.subordinateGpuQuota).toBeUndefined();

      // Should still allow editing
      wrapper.vm.editableValue = 20;
      await nextTick();

      // Should not show error since subordinateQuota is undefined
      expect(wrapper.vm.quotaState).not.toBe(EQuotaInputState.Error);
    });

    it("treats -1 as unlimited (null) for subordinate quota", async () => {
      const mockDepartment = createDepartmentMock(-1);
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment);

      const column = createColumn({ entityType: EQuotaEntity.department });
      const row = createRow({ id: "dept-1", gpuQuota: 10 });

      const wrapper = createWrapper(row, column);

      // Open the popup - this triggers the fetch
      await wrapper.find(".cell-trigger").trigger("click");
      // Manually trigger the opened event (QMenu's @show might not fire in tests)
      wrapper.findComponent({ name: "EditableCellPopup" }).vm.$emit("opened");
      await flushPromises();

      expect(wrapper.vm.subordinateGpuQuota).toBeNull();

      // Should not show error when subordinate is unlimited
      expect(wrapper.vm.quotaState).not.toBe(EQuotaInputState.Error);
    });

    it("handles missing projectsAggregatedResources", async () => {
      const mockDepartment: Department = {
        id: "dept-1",
        name: "Department 1",
        clusterId: "cluster-1",
        projectsAggregatedResources: null,
        resources: [],
        totalResources: { gpuQuota: 0 },
        description: "",
        createdAt: "2024-01-01",
        updatedAt: "2024-01-01",
      };
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment);

      const column = createColumn({ entityType: EQuotaEntity.department });
      const row = createRow({ id: "dept-1", gpuQuota: 30 });

      const wrapper = createWrapper(row, column);

      // Open the popup - this triggers the fetch
      await wrapper.find(".cell-trigger").trigger("click");
      // Manually trigger the opened event (QMenu's @show might not fire in tests)
      wrapper.findComponent({ name: "EditableCellPopup" }).vm.$emit("opened");
      await flushPromises();

      expect(wrapper.vm.subordinateGpuQuota).toBeUndefined();
    });

    it("filters projectsAggregatedResources by node pool name", async () => {
      const mockDepartment: Department = {
        id: "dept-1",
        name: "Department 1",
        clusterId: "cluster-1",
        projectsAggregatedResources: [
          {
            nodePool: { id: "pool-1", name: "default" },
            gpu: { deserved: 20 },
          },
          {
            nodePool: { id: "pool-2", name: "gpu-pool" },
            gpu: { deserved: 50 },
          },
        ],
        resources: [],
        totalResources: { gpuQuota: 0 },
        description: "",
        createdAt: "2024-01-01",
        updatedAt: "2024-01-01",
      };
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue(mockDepartment);

      const column = createColumn({ entityType: EQuotaEntity.department, nodePoolName: "gpu-pool" });
      const row = createRow({ id: "dept-1", gpuQuota: 60 });

      const wrapper = createWrapper(row, column);

      // Open the popup - this triggers the fetch
      await wrapper.find(".cell-trigger").trigger("click");
      // Manually trigger the opened event (QMenu's @show might not fire in tests)
      wrapper.findComponent({ name: "EditableCellPopup" }).vm.$emit("opened");
      await flushPromises();

      // Should use the gpu-pool's subordinate quota (50), not default's (20)
      expect(wrapper.vm.subordinateGpuQuota).toBe(50);
    });
  });
});
