import { describe, it, expect, vi, afterEach } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import AccessRuleDeletionModal from "./access-rule-deletion-modal.vue";
import type { AccessRule } from "@/swagger-models/authorization-client";
import { unitTestUtil } from "@/utils/unit-test.util";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service");
let wrapper: VueWrapper<InstanceType<typeof AccessRuleDeletionModal>>;

function mountComponent(props: { selectedAccessRules: AccessRule[]; deletableAccessRulesIds: number[] }) {
  wrapper = mount(AccessRuleDeletionModal, {
    global: {
      plugins: [Quasar],
    },
    props,
  });
  wrapper.vm.$q.notify = vi.fn();
}

describe("AccessRuleDeletionModal.vue", () => {
  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  it("should render correctly", () => {
    mountComponent({
      selectedAccessRules: [{ id: 1, subjectId: "user1" }] as AccessRule[],
      deletableAccessRulesIds: [1],
    });
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.findComponent(AccessRuleDeletionModal).exists()).toBe(true);
  });

  it("should display the correct title for single deletion", async () => {
    mountComponent({
      selectedAccessRules: [{ id: 1, subjectId: "user1" }] as AccessRule[],
      deletableAccessRulesIds: [1],
    });
    await wrapper.vm.$nextTick();
    const modalEl = unitTestUtil.getModalElement();
    const title = modalEl.find(".delete-modal-msg").text();
    expect(title).toBe("Permanently delete 1 access rule?");
  });

  it("should display the correct title for batch deletion", async () => {
    mountComponent({
      selectedAccessRules: [
        { id: 1, subjectId: "user1" },
        { id: 2, subjectId: "user2" },
      ] as AccessRule[],
      deletableAccessRulesIds: [1, 2],
    });
    await wrapper.vm.$nextTick();

    const modalEl = unitTestUtil.getModalElement();
    const title = modalEl.find(".delete-modal-msg").text();
    expect(title).toBe("Permanently delete 2 access rules?");
  });

  it("should emit 'cancel' event when cancel button is clicked", async () => {
    mountComponent({
      selectedAccessRules: [{ id: 1, subjectId: "user1" }] as AccessRule[],
      deletableAccessRulesIds: [1],
    });
    const modal = wrapper.findComponent({ name: "RunaiDeleteModal" });
    await modal.vm.$emit("cancel");
    expect(wrapper.emitted()).toHaveProperty("cancel");
  });

  it("should emit 'deletion-success' event after successful deletion", async () => {
    mountComponent({
      selectedAccessRules: [{ id: 1, subjectId: "user1" }] as AccessRule[],
      deletableAccessRulesIds: [1],
    });
    const modal = wrapper.findComponent({ name: "RunaiDeleteModal" });
    await modal.vm.$emit("delete");
    expect(wrapper.emitted()).toHaveProperty("deletion-success");
  });

  it("should handle errors during batch deletion", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const errorMessage = "Access rule/s couldn't be deleted";
    vi.mocked(accessRuleService.accessRulesBatch).mockRejectedValue(new Error(errorMessage));
    mountComponent({
      selectedAccessRules: [
        { id: 1, subjectId: "user1" },
        { id: 2, subjectId: "user2" },
      ] as AccessRule[],
      deletableAccessRulesIds: [1, 2],
    });

    const notifySpy = vi.spyOn(wrapper.vm, "notifyError");

    const modal = wrapper.findComponent({ name: "RunaiDeleteModal" });
    await modal.vm.$emit("delete");

    await flushPromises();

    expect(wrapper.emitted()).not.toHaveProperty("deletion-success");
    expect(notifySpy).toHaveBeenCalledWith(
      expect.objectContaining({
        message: errorMessage,
      }),
    );
    errorSpy.mockRestore();
  });

  it("should show tooltip when there are non-deletable access rules", async () => {
    mountComponent({
      selectedAccessRules: [
        { id: 1, subjectId: "user1" },
        { id: 2, subjectId: "user2" },
      ] as AccessRule[],
      deletableAccessRulesIds: [1],
    });
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.deletionTitle?.secondary).toBe("1 out of 2 access rules can't be deleted.");
  });

  it("should handle single deletion with non-deletable rules", async () => {
    mountComponent({
      selectedAccessRules: [
        { id: 1, subjectId: "user1" },
        { id: 2, subjectId: "user2" },
      ] as AccessRule[],
      deletableAccessRulesIds: [1],
    });
    await wrapper.vm.$nextTick();
    const modalEl = unitTestUtil.getModalElement();
    const title = modalEl.find(".delete-modal-msg").text();
    expect(title).toBe("Permanently delete 1 access rule?");
    expect(wrapper.vm.deletionTitle?.secondary).toBe("1 out of 2 access rules can't be deleted.");
  });
});
