import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, vi, test, expect } from "vitest";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import GvkPolicyYamlSection from "./gvk-policy-yaml-section.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("gvk-policy-yaml-section.vue", () => {
  const permissionStore = usePermissionStore(pinia);

  // Empty state (no rules) so the read-only editor (monaco) is not mounted.
  function mountComponent(permittedActions: Action[]): VueWrapper {
    permissionStore.hasPermission = vi
      .fn()
      .mockImplementation(
        (resource: ResourceType, action: Action) =>
          resource === ResourceType.Policies && permittedActions.includes(action),
      );
    return mount(GvkPolicyYamlSection, {
      global: {
        plugins: [Quasar],
      },
      props: { modelValue: [] },
    });
  }

  test("add-policy-yaml button is enabled with Policies/Update permission", () => {
    const wrapper = mountComponent([Action.Update]);
    const button = wrapper.find('[aid="add-gvk-policy-btn"]').element as HTMLButtonElement;
    expect(button.disabled).toBe(false);
  });

  test("add-policy-yaml button is enabled with Policies/Create permission only (create flow → POST)", () => {
    const wrapper = mountComponent([Action.Create]);
    const button = wrapper.find('[aid="add-gvk-policy-btn"]').element as HTMLButtonElement;
    expect(button.disabled).toBe(false);
  });

  test("add-policy-yaml button is disabled without Policies create/update permission", () => {
    const wrapper = mountComponent([]);
    const button = wrapper.find('[aid="add-gvk-policy-btn"]').element as HTMLButtonElement;
    expect(button.disabled).toBe(true);
  });
});
