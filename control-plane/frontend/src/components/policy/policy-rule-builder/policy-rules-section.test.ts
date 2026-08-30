import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, vi, test, expect } from "vitest";
import { ref } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import PolicyRulesSection from "./policy-rules-section.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { errorMessages } from "@/common/error-message.constant";

// The section drives the real mapping composable on mount; stub it so no service calls fire.
vi.mock("@/composables/use-policy-mapping.composable", () => ({
  usePolicyMapping: () => ({
    loadingMinimal: ref(false),
    loadMinimal: vi.fn().mockResolvedValue(undefined),
  }),
}));

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("policy-rules-section.vue", () => {
  const permissionStore = usePermissionStore(pinia);

  function mountComponent(permittedActions: Action[]): VueWrapper {
    permissionStore.hasPermission = vi
      .fn()
      .mockImplementation(
        (resource: ResourceType, action: Action) =>
          resource === ResourceType.Policies && permittedActions.includes(action),
      );
    return mount(PolicyRulesSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        gvk: { group: "apps", version: "v1", kind: "Deployment" },
        modelValue: [],
      },
    });
  }

  test("add-rule button is enabled with Policies/Update permission", () => {
    const wrapper = mountComponent([Action.Update]);
    const button = wrapper.find('[aid="add-rule-btn"]').element as HTMLButtonElement;
    expect(button.disabled).toBe(false);
  });

  test("add-rule button is enabled with Policies/Create permission only (create flow → POST)", () => {
    const wrapper = mountComponent([Action.Create]);
    const button = wrapper.find('[aid="add-rule-btn"]').element as HTMLButtonElement;
    expect(button.disabled).toBe(false);
  });

  test("add-rule button is disabled without Policies create/update permission", () => {
    const wrapper = mountComponent([]);
    const button = wrapper.find('[aid="add-rule-btn"]').element as HTMLButtonElement;
    expect(button.disabled).toBe(true);
  });

  // The inline reason stays hidden until a Save Policy attempt (revealValidation) so it never nags
  // mid-edit; once revealed it shows the specific reason in-section (the policy footer stays generic).
  test("hides the rules validation reason until a save attempt, then reveals it inline", async () => {
    const wrapper = mountComponent([Action.Update]);
    expect(wrapper.find('[aid="rules-validation-error"]').exists()).toBe(false);

    (wrapper.vm as unknown as { revealValidation: () => void }).revealValidation();
    await wrapper.vm.$nextTick();

    const error = wrapper.find('[aid="rules-validation-error"]');
    expect(error.exists()).toBe(true);
    expect(error.text()).toContain(errorMessages.ADD_AT_LEAST_ONE_RULE);
  });
});
