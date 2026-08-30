import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import CleanPodPolicySection from "./clean-pod-policy-section.vue";
import { CleanPodPolicy } from "@/swagger-models/workloads-client";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";

describe("clean-pod-policy-section.vue", () => {
  const createWrapper = (props = {}) => {
    return mount(CleanPodPolicySection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: null,
        sectionDisabled: false,
        ...props,
      },
    }) as VueWrapper<InstanceType<typeof CleanPodPolicySection>>;
  };

  it("renders with default props", () => {
    const wrapper = createWrapper();
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.findComponent({ name: "runai-section" }).exists()).toBe(true);
    expect(wrapper.findComponent({ name: "runai-select" }).exists()).toBe(true);
  });

  it("displays correct options", () => {
    const wrapper = createWrapper();
    const options = wrapper.vm.cleanPodPolicyOptions;

    expect(options).toHaveLength(3);
    expect(options[0].label).toBe("None");
    expect(options[1].label).toBe("All");
    expect(options[2].label).toBe("Running");
  });

  it("emits update event when selection changes", async () => {
    const wrapper = createWrapper();
    const select = wrapper.findComponent({ name: "runai-select" });

    await select.vm.$emit("update:model-value", { value: CleanPodPolicy.All });

    expect(wrapper.emitted("update:model-value")).toBeTruthy();
    expect(wrapper.emitted("update:model-value")?.[0]).toEqual([CleanPodPolicy.All]);
  });

  it("disables section when sectionDisabled is true", () => {
    const wrapper = createWrapper({ sectionDisabled: true });
    const select = wrapper.findComponent({ name: "runai-select" });

    expect(select.props("disable")).toBe(true);
  });

  it("applies policy rules when provided", () => {
    const policyRules = { isClosedList: true };
    const wrapper = createWrapper({ policyRules });
    const select = wrapper.findComponent({ name: "runai-select" });

    expect(select.props("policyRules")).toEqual({ ...policyRules, isClosedList: true });
  });

  it("handles null modelValue correctly", () => {
    const wrapper = createWrapper({ modelValue: null });
    const select = wrapper.findComponent({ name: "runai-select" });

    expect(select.props("modelValue")).toBe("");
  });

  it("handles non-null modelValue correctly", () => {
    const wrapper = createWrapper({ modelValue: CleanPodPolicy.All });
    const select = wrapper.findComponent({ name: "runai-select" });

    expect(select.props("modelValue")).toBe(CleanPodPolicy.All);
  });
});
