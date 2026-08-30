import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { describe, it, expect, vi } from "vitest";
import { Quasar } from "quasar";
import { UserProfileAccessRulesTable } from "@/components/user-settings/sections/user-profile-section/user-profile-access-rules-table/index";
import { createTestingPinia } from "@pinia/testing";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service");
let wrapper: VueWrapper<InstanceType<typeof UserProfileAccessRulesTable>>;
const pinia = createTestingPinia({ createSpy: vi.fn });

const mockAccessRules = [
  { id: 1, name: "Rule 1" },
  { id: 2, name: "Rule 2" },
];

accessRuleService.getMeAccessRules = vi.fn().mockResolvedValue(mockAccessRules);
accessRuleService.countMeAccessRules = vi.fn().mockResolvedValue(mockAccessRules.length);

function createComponent() {
  wrapper = mount(UserProfileAccessRulesTable, {
    global: {
      plugins: [Quasar, pinia],
      directives: {
        clickOutside: vi.fn(),
        permission: vi.fn(),
      },
    },
  });
  wrapper.vm.$q.notify = vi.fn();
}

describe("UserProfileAccessRulesTable.vue", () => {
  it("renders correctly", () => {
    createComponent();
    expect(wrapper.exists()).toBe(true);
  });

  it("should get access rules on first load", async () => {
    createComponent();
    await flushPromises();
    expect(accessRuleService.getMeAccessRules).toHaveBeenCalled();
  });

  it("displays loading state", async () => {
    createComponent();
    expect(wrapper.vm.loading).toBe(true);
    await flushPromises();
    expect(wrapper.vm.loading).toBe(false);
  });

  it("updates search term and fetches access rules", async () => {
    createComponent();
    const searchTerm = "test";
    await wrapper.vm.updateSearch(searchTerm);
    expect(wrapper.vm.filterBy.searchTerm).toBe(searchTerm);
    await flushPromises();
    expect(accessRuleService.getMeAccessRules).toHaveBeenCalled();
  });

  it("updates filter and fetches access rules", async () => {
    createComponent();
    const newFilter = { searchTerm: "new", rowsNumber: 5 };
    await wrapper.vm.updateFilter(newFilter);
    expect(wrapper.vm.filterBy).toEqual(newFilter);
    expect(accessRuleService.getMeAccessRules).toHaveBeenCalled();
  });

  it("handles errors in fetching access rules", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    accessRuleService.getMeAccessRules = vi.fn().mockRejectedValue(new Error("Failed to fetch"));
    createComponent();
    await wrapper.vm.getAccessRules();
    expect(errorSpy).toHaveBeenCalledWith("Failed to load access rules", expect.any(Error));
    expect(wrapper.vm.$q.notify).toHaveBeenCalled();
    expect(wrapper.vm.loading).toBe(false);
    errorSpy.mockRestore();
  });

  it("should handle empty state", async () => {
    accessRuleService.getMeAccessRules = vi.fn().mockResolvedValue([]);
    createComponent();
    await wrapper.vm.getAccessRules();
    expect(wrapper.vm.loading).toBe(false);
    expect(wrapper.vm.accessRules).toEqual([]);
    expect(wrapper.find(".empty-state-message").text()).toBe("You don't have any access rules yet.");
  });

  it("should emit access-rules-fetched on initial load", async () => {
    createComponent();
    await flushPromises();
    expect(wrapper.emitted("access-rules-fetched")).toBeTruthy();
  });
});
