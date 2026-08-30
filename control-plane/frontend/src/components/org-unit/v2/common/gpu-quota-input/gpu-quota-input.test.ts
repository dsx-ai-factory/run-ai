import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import GpuQuotaInput from "./gpu-quota-input.vue";
import { EQuotaEntity } from "@/models/resource.model";
import { EQuotaInputState } from "@/models/org-unit.model";

// Mock the QuotaInput component to capture props passed to it
vi.mock("@/components/org-unit/v2/common/quota-input", () => ({
  QuotaInput: {
    name: "QuotaInput",
    template: `
      <div class="mock-quota-input">
        <input 
          type="number" 
          :value="modelValue" 
          @input="$emit('update:model-value', $event.target.value === '' ? null : Number($event.target.value))"
          data-testid="quota-input"
        />
      </div>
    `,
    props: [
      "modelValue",
      "state",
      "message",
      "loading",
      "label",
      "rules",
      "preventNull",
      "defaultValue",
      "placeholder",
      "treatMinusOneAsUnlimited",
    ],
    emits: ["update:model-value"],
  },
}));

describe("GpuQuotaInput", () => {
  let wrapper: VueWrapper<InstanceType<typeof GpuQuotaInput>>;

  const mountComponent = (props: {
    modelValue: number | null;
    state?: EQuotaInputState;
    message?: string;
    loading?: boolean;
    entityType?: EQuotaEntity;
  }) => {
    wrapper = mount(GpuQuotaInput, {
      global: {
        plugins: [Quasar],
      },
      props: {
        modelValue: props.modelValue,
        state: props.state ?? EQuotaInputState.Default,
        message: props.message ?? "",
        loading: props.loading ?? false,
        entityType: props.entityType ?? EQuotaEntity.project,
      },
    });
    return wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    it("renders the component correctly", () => {
      const wrapper = mountComponent({ modelValue: 10 });
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.findComponent({ name: "QuotaInput" }).exists()).toBe(true);
    });

    it("passes modelValue to QuotaInput", () => {
      const wrapper = mountComponent({ modelValue: 25 });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("modelValue")).toBe(25);
    });

    it("passes state to QuotaInput", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        state: EQuotaInputState.Error,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("state")).toBe(EQuotaInputState.Error);
    });

    it("passes message to QuotaInput", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        message: "Custom error message",
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("message")).toBe("Custom error message");
    });

    it("passes loading state to QuotaInput", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        loading: true,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("loading")).toBe(true);
    });

    it("displays 'GPUs' as the label", () => {
      const wrapper = mountComponent({ modelValue: 10 });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("label")).toBe("GPUs");
    });
  });

  describe("Entity Type Behavior - Projects", () => {
    it("does not treat -1 as unlimited for projects", () => {
      const wrapper = mountComponent({
        modelValue: -1,
        entityType: EQuotaEntity.project,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("treatMinusOneAsUnlimited")).toBe(false);
    });

    it("prevents null values for projects", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        entityType: EQuotaEntity.project,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("preventNull")).toBe(true);
    });

    it("does not show Unlimited placeholder for projects", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        entityType: EQuotaEntity.project,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("placeholder")).toBeUndefined();
    });

    it("uses 0 as default value for projects", () => {
      const wrapper = mountComponent({
        modelValue: null,
        entityType: EQuotaEntity.project,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("defaultValue")).toBe(0);
    });
  });

  describe("Entity Type Behavior - Departments", () => {
    it("treats -1 as unlimited for departments", () => {
      const wrapper = mountComponent({
        modelValue: -1,
        entityType: EQuotaEntity.department,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("treatMinusOneAsUnlimited")).toBe(true);
    });

    it("allows null values for departments", () => {
      const wrapper = mountComponent({
        modelValue: null,
        entityType: EQuotaEntity.department,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("preventNull")).toBe(false);
    });

    it("shows Unlimited placeholder for departments", () => {
      const wrapper = mountComponent({
        modelValue: null,
        entityType: EQuotaEntity.department,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("placeholder")).toBe("Unlimited");
    });
  });

  describe("Validation Rules", () => {
    it("returns true when state is not error", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        state: EQuotaInputState.Default,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      const rules = quotaInput.props("rules") as Array<() => boolean | string>;
      expect(rules).toHaveLength(1);
      expect(rules[0]()).toBe(true);
    });

    it("returns empty string when state is error", () => {
      const wrapper = mountComponent({
        modelValue: -5,
        state: EQuotaInputState.Error,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      const rules = quotaInput.props("rules") as Array<() => boolean | string>;
      expect(rules[0]()).toBe("");
    });

    it("returns true when state is warning", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        state: EQuotaInputState.Warning,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      const rules = quotaInput.props("rules") as Array<() => boolean | string>;
      expect(rules[0]()).toBe(true);
    });
  });

  describe("Value Emission", () => {
    it("emits update:model-value when value changes", async () => {
      const wrapper = mountComponent({ modelValue: 10 });

      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      await quotaInput.vm.$emit("update:model-value", 20);

      expect(wrapper.emitted("update:model-value")).toBeTruthy();
      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([20]);
    });

    it("emits null when input is cleared", async () => {
      const wrapper = mountComponent({ modelValue: 10 });

      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      await quotaInput.vm.$emit("update:model-value", null);

      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([null]);
    });

    it("emits -1 for unlimited department quota", async () => {
      const wrapper = mountComponent({
        modelValue: 10,
        entityType: EQuotaEntity.department,
      });

      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      await quotaInput.vm.$emit("update:model-value", -1);

      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([-1]);
    });
  });

  describe("Reactivity", () => {
    it("updates QuotaInput when modelValue prop changes", async () => {
      const wrapper = mountComponent({ modelValue: 10 });

      await wrapper.setProps({ modelValue: 50 });
      await flushPromises();

      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("modelValue")).toBe(50);
    });

    it("updates QuotaInput when state prop changes", async () => {
      const wrapper = mountComponent({
        modelValue: 10,
        state: EQuotaInputState.Default,
      });

      await wrapper.setProps({ state: EQuotaInputState.Error });
      await flushPromises();

      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("state")).toBe(EQuotaInputState.Error);
    });

    it("updates treatMinusOneAsUnlimited when entityType changes", async () => {
      const wrapper = mountComponent({
        modelValue: 10,
        entityType: EQuotaEntity.project,
      });

      let quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("treatMinusOneAsUnlimited")).toBe(false);

      await wrapper.setProps({ entityType: EQuotaEntity.department });
      await flushPromises();

      quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("treatMinusOneAsUnlimited")).toBe(true);
    });
  });

  describe("Default Props", () => {
    it("uses default state when not provided", () => {
      const wrapper = mount(GpuQuotaInput, {
        global: { plugins: [Quasar] },
        props: { modelValue: 10 },
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("state")).toBe(EQuotaInputState.Default);
    });

    it("uses empty message when not provided", () => {
      const wrapper = mount(GpuQuotaInput, {
        global: { plugins: [Quasar] },
        props: { modelValue: 10 },
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("message")).toBe("");
    });

    it("uses false for loading when not provided", () => {
      const wrapper = mount(GpuQuotaInput, {
        global: { plugins: [Quasar] },
        props: { modelValue: 10 },
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("loading")).toBe(false);
    });

    it("defaults to project entity type", () => {
      const wrapper = mount(GpuQuotaInput, {
        global: { plugins: [Quasar] },
        props: { modelValue: 10 },
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      // Projects don't treat -1 as unlimited
      expect(quotaInput.props("treatMinusOneAsUnlimited")).toBe(false);
    });
  });

  describe("Edge Cases", () => {
    it("handles zero value correctly", () => {
      const wrapper = mountComponent({ modelValue: 0 });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("modelValue")).toBe(0);
    });

    it("handles large values correctly", () => {
      const wrapper = mountComponent({ modelValue: 999999 });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("modelValue")).toBe(999999);
    });

    it("handles decimal values correctly", () => {
      const wrapper = mountComponent({ modelValue: 10.5 });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("modelValue")).toBe(10.5);
    });

    it("handles null value for department", () => {
      const wrapper = mountComponent({
        modelValue: null,
        entityType: EQuotaEntity.department,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("modelValue")).toBeNull();
      expect(quotaInput.props("treatMinusOneAsUnlimited")).toBe(true);
    });
  });
});
