import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import RequestTimeoutSection, {
  EEmitRequestTimeoutSection,
  INITIALIZATION_DEFAULTS,
  REQUEST_TIMEOUT_DEFAULTS,
} from "./request-timeout-section.vue";
import { DurationInputSection } from "./duration-input";
import type { ServingConfiguration, ServingConfigurationRules } from "@/swagger-models/assets-service-client";

// Helper to mount the component
const defaultModelValue: ServingConfiguration = {
  initializationTimeoutSeconds: 100,
  requestTimeoutSeconds: 200,
};

const defaultPolicyRules: ServingConfigurationRules = {
  initializationTimeoutSeconds: { min: 5, max: 3600, step: 5, canEdit: true },
  requestTimeoutSeconds: { min: 5, max: 600, step: 5, canEdit: true },
};

const mountCmp = (props = {}) =>
  mount(RequestTimeoutSection, {
    global: { plugins: [Quasar] },
    props: {
      modelValue: defaultModelValue,
      policyRules: defaultPolicyRules,
      ...props,
    },
  });

describe("request-timeout-section.vue", () => {
  it("renders two DurationInputSection components", () => {
    const wrapper = mountCmp();
    const durationInputs = wrapper.findAllComponents(DurationInputSection);
    expect(durationInputs.length).toBe(2);
  });

  it("passes correct props to initialization DurationInputSection", () => {
    const wrapper = mountCmp();
    const initInput = wrapper.findAllComponents(DurationInputSection)[0];
    expect(initInput.html()).toContain(INITIALIZATION_DEFAULTS.label);
    const tooltip = initInput.findComponent({ name: "runai-tooltip" }).findComponent({ name: "q-tooltip" });
    expect(tooltip.vm.$slots.default()[0].children[0].children).toContain(INITIALIZATION_DEFAULTS.tooltipText);
    expect(initInput.props("modelValue")).toBe(defaultModelValue.initializationTimeoutSeconds);
    expect(initInput.props("disable")).toBe(false);
    expect(initInput.props("min")).toBe(INITIALIZATION_DEFAULTS.min);
    expect(initInput.props("max")).toBe(INITIALIZATION_DEFAULTS.max);
    expect(initInput.props("policyRules")).toEqual(defaultPolicyRules.initializationTimeoutSeconds);
  });

  it("passes correct props to request DurationInputSection", () => {
    const wrapper = mountCmp();
    const reqInput = wrapper.findAllComponents(DurationInputSection)[1];
    expect(reqInput.html()).toContain(REQUEST_TIMEOUT_DEFAULTS.label);
    const tooltip = reqInput.findComponent({ name: "runai-tooltip" }).findComponent({ name: "q-tooltip" });
    expect(tooltip.vm.$slots.default()[0].children[0].children).toContain(REQUEST_TIMEOUT_DEFAULTS.tooltipText);
    expect(reqInput.props("modelValue")).toBe(defaultModelValue.requestTimeoutSeconds);
    expect(reqInput.props("disable")).toBe(false);
    expect(reqInput.props("policyRules")).toEqual(defaultPolicyRules.requestTimeoutSeconds);
    expect(reqInput.props("min")).toBe(5);
    expect(reqInput.props("max")).toBe(600);
  });

  it("disables DurationInputSection if disable prop is true", () => {
    const wrapper = mountCmp({ disable: true });
    const durationInputs = wrapper.findAllComponents(DurationInputSection);
    expect(durationInputs[0].props("disable")).toBe(true);
    expect(durationInputs[1].props("disable")).toBe(true);
  });

  it("emits update:model-value with updated initializationTimeoutSeconds", () => {
    const wrapper = mountCmp();
    // Simulate child event
    wrapper.vm.updateInitializationTimeout(555);
    const emittedObj = wrapper.emitted() as Record<string, unknown[]>;
    const arr = emittedObj[EEmitRequestTimeoutSection.UpdateModelValue];
    expect(Array.isArray(arr)).toBe(true);
    const emitted = Array.isArray(arr) && Array.isArray(arr[0]) ? (arr[0][0] as ServingConfiguration) : undefined;
    expect(emitted?.initializationTimeoutSeconds).toBe(555);
    expect(emitted?.requestTimeoutSeconds).toBe(defaultModelValue.requestTimeoutSeconds);
  });

  it("emits update:model-value with updated requestTimeoutSeconds", () => {
    const wrapper = mountCmp();
    wrapper.vm.updateRequestTimeout(333);
    const emittedObj = wrapper.emitted() as Record<string, unknown[]>;
    const arr = emittedObj[EEmitRequestTimeoutSection.UpdateModelValue];
    expect(Array.isArray(arr)).toBe(true);
    const emitted = Array.isArray(arr) && Array.isArray(arr[0]) ? (arr[0][0] as ServingConfiguration) : undefined;
    expect(emitted?.requestTimeoutSeconds).toBe(333);
    expect(emitted?.initializationTimeoutSeconds).toBe(defaultModelValue.initializationTimeoutSeconds);
  });

  it("passes null to DurationInputSection if modelValue fields are missing", () => {
    const wrapper = mountCmp({ modelValue: {} });
    const durationInputs = wrapper.findAllComponents(DurationInputSection);
    expect(durationInputs[0].props("modelValue")).toBe(null);
    expect(durationInputs[1].props("modelValue")).toBe(null);
  });

  it("passes policyRules correctly when not provided", () => {
    const wrapper = mountCmp({ policyRules: undefined });
    const durationInputs = wrapper.findAllComponents(DurationInputSection);
    expect(durationInputs[0].props("policyRules")).toBeNull();
    expect(durationInputs[1].props("policyRules")).toBeNull();
  });
});
