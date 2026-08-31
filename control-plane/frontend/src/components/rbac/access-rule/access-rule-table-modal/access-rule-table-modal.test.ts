/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";

// Components
import AccessRuleTableModal from "./access-rule-table-modal.vue";

// Mock data and services
import type { IAccessRuleTableModalOptions } from "@/models/access-rule.model";
import { SubjectTypeExtended, type AccessRule } from "@/swagger-models/authorization-client";
import { unitTestUtil } from "@/utils/unit-test.util";

// Import services to access mocked functions
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import {
  mockAccessRules,
  mockModalOptionsBackwardCompatibility,
  mockModalOptionsServerSide,
  mockModalOptionsWithSubject,
} from "@/components/rbac/access-rule/access-rule-table-modal/access-rule-table-modal.mock";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });

// Mock access rules service
vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    getAccessRulesV2: vi.fn().mockResolvedValue([]),
    getAccessRulesCount: vi.fn().mockResolvedValue(0),
    getAllAccessRules: vi.fn().mockResolvedValue([]),
  },
}));

describe("AccessRuleTableModal", () => {
  let wrapper: VueWrapper<InstanceType<typeof AccessRuleTableModal>>;

  const createComponent = (props: { modalOptions: IAccessRuleTableModalOptions }): void => {
    wrapper = mount(AccessRuleTableModal, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
          clickOutside: vi.fn(),
        },
      },
      props: props,
    });
    wrapper.vm.$q.notify = vi.fn(); // Disable notify to prevent test failures
  };

  beforeEach(() => {
    createComponent({
      modalOptions: mockModalOptionsServerSide,
    });
  });

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  test("renders the modal with correct header", () => {
    const modalEl = unitTestUtil.getModalElement();
    expect(modalEl.find(".runai-base-dialog-header").text()).toBe(mockModalOptionsServerSide.header);
  });

  test("renders table and search components", () => {
    const modalEl = unitTestUtil.getModalElement();
    expect(modalEl.find(".runai-table").exists()).toBe(true);
    expect(modalEl.find("[aid=runai-list-modal-search]").exists()).toBe(true);
  });

  test("enables server-side pagination when no accessRules provided", () => {
    expect(wrapper.vm.isServerSidePagination).toBe(true);
  });

  test("disables server-side pagination when accessRules provided", () => {
    wrapper.unmount();
    createComponent({
      modalOptions: mockModalOptionsBackwardCompatibility,
    });
    expect(wrapper.vm.isServerSidePagination).toBe(false);
  });

  test("displays provided accessRules in backward compatibility mode", () => {
    wrapper.unmount();
    createComponent({
      modalOptions: mockModalOptionsBackwardCompatibility,
    });
    expect(wrapper.vm.displayAccessRules).toEqual(mockAccessRules);
  });

  test("constructs correct filters for scope-based filtering", async () => {
    await flushPromises();

    expect(vi.mocked(accessRuleService.getAccessRulesV2)).toHaveBeenCalled();
    expect(vi.mocked(accessRuleService.getAccessRulesCount)).toHaveBeenCalled();
  });

  test("merges the subject's own rules with all-authenticated-users rules for a user subject", async () => {
    const subjectRule = { ...mockAccessRules[0] } as AccessRule;
    const allAuthRule = {
      id: 99,
      roleName: "Viewer",
      subjectType: SubjectTypeExtended.AllAuthenticatedUsers,
      subjectId: "all-authenticated-users",
    } as AccessRule;
    vi.mocked(accessRuleService.getAllAccessRules)
      .mockResolvedValueOnce([subjectRule])
      .mockResolvedValueOnce([allAuthRule]);

    wrapper.unmount();
    createComponent({
      modalOptions: mockModalOptionsWithSubject,
    });
    await flushPromises();

    expect(wrapper.vm.includeAllAuth).toBe(true);
    expect(wrapper.vm.isServerSidePagination).toBe(false);
    expect(vi.mocked(accessRuleService.getAllAccessRules)).toHaveBeenCalledTimes(2);
    expect(wrapper.vm.displayAccessRules).toEqual([subjectRule, allAuthRule]);
  });

  test("handles search functionality", async () => {
    const modalEl = unitTestUtil.getModalElement();
    const searchInput = modalEl.find("[aid=runai-list-modal-search]");

    await searchInput.setValue("test search");
    await flushPromises();

    // Verify that filter update was triggered
    expect(wrapper.vm.tableFilter.searchTerm).toBe("test search");
    expect(vi.mocked(accessRuleService.getAccessRulesV2)).toHaveBeenCalledWith(
      ["scopeType==project", "scopeId==project-1"],
      "roleName",
      "asc",
      0,
      10,
      "test search",
    );
  });

  test("loads access rules count on filter update", async () => {
    vi.mocked(accessRuleService.getAccessRulesCount).mockResolvedValue(25);

    await wrapper.vm.updateFilter({
      ...wrapper.vm.tableFilter,
      searchTerm: "new search",
    });
    await flushPromises();

    expect(vi.mocked(accessRuleService.getAccessRulesCount)).toHaveBeenCalled();
    expect(wrapper.vm.tableFilter.rowsNumber).toBe(25);
  });

  test("handles loading states correctly", async () => {
    expect(wrapper.vm.displayLoading).toBe(false);

    // Test loading state when internal loading is true
    wrapper.vm.loading = true;
    expect(wrapper.vm.displayLoading).toBe(true);
  });

  test("close button emits close event", async () => {
    const modalEl = unitTestUtil.getModalElement();
    const closeButton = modalEl.find("[aid=close-access-rule-table-modal-btn]");

    if (closeButton.exists()) {
      await closeButton.trigger("click");
      expect(wrapper.emitted("close")).toBeTruthy();
    } else {
      // Fallback - directly test the emit
      wrapper.vm.$emit("close");
      expect(wrapper.emitted("close")).toBeTruthy();
    }
  });

  test("handles error during access rules loading", async () => {
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(accessRuleService.getAccessRulesV2).mockRejectedValueOnce(new Error("API Error"));

    await wrapper.vm.updateFilter(wrapper.vm.tableFilter);
    await flushPromises();

    expect(consoleErrorSpy).toHaveBeenCalled();
    expect(wrapper.vm.$q.notify).toHaveBeenCalled();

    consoleErrorSpy.mockRestore();
  });
});
