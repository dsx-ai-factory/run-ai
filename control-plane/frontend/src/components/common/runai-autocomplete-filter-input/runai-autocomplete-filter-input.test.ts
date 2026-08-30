import { describe, expect, it, vi, beforeEach } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import RunaiAutocompleteFilterInput from "./runai-autocomplete-filter-input.vue";

function mountComponent(props: Record<string, unknown> = {}): VueWrapper {
  return mount(RunaiAutocompleteFilterInput, {
    global: {
      plugins: [Quasar],
    },
    props,
  });
}

describe("RunaiAutocompleteFilterInput", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("plain input fallback (no autocomplete/search fn)", () => {
    it("renders a q-input when no fn props are provided", () => {
      const wrapper = mountComponent({ label: "Name", aid: "test-input" });

      expect(wrapper.find(".q-input").exists()).toBe(true);
      expect(wrapper.find(".q-select").exists()).toBe(false);
    });

    it("passes label and aid to q-input", () => {
      const wrapper = mountComponent({ label: "Name", aid: "test-input" });

      expect(wrapper.find("[aid='test-input']").exists()).toBe(true);
      expect(wrapper.html()).toContain("Name");
    });

    it("passes dense prop to q-input", () => {
      const wrapper = mountComponent({ dense: true });

      expect(wrapper.find(".q-field--dense").exists()).toBe(true);
    });

    it("emits update:model-value on input change", async () => {
      const wrapper = mountComponent({ modelValue: "" });

      const input = wrapper.find("input");
      await input.setValue("hello");

      expect(wrapper.emitted("update:model-value")).toBeTruthy();
      expect(wrapper.emitted("update:model-value")?.[0]).toEqual(["hello"]);
    });
  });

  describe("string-based autocomplete mode (autocompleteOptionsFn)", () => {
    const mockAutocompleteFn = vi.fn().mockResolvedValue(["Alpha", "Beta", "Gamma"]);

    it("renders a q-select when autocompleteOptionsFn is provided", () => {
      const wrapper = mountComponent({
        autocompleteOptionsFn: mockAutocompleteFn,
        label: "Filter",
        aid: "autocomplete-test",
      });

      expect(wrapper.find(".q-select").exists()).toBe(true);
      expect(wrapper.find(".q-input").exists()).toBe(false);
    });

    it("passes label and aid to q-select", () => {
      const wrapper = mountComponent({
        autocompleteOptionsFn: mockAutocompleteFn,
        label: "Filter",
        aid: "autocomplete-test",
      });

      expect(wrapper.find("[aid='autocomplete-test']").exists()).toBe(true);
      expect(wrapper.html()).toContain("Filter");
    });

    it("passes dense prop to q-select", () => {
      const wrapper = mountComponent({
        autocompleteOptionsFn: mockAutocompleteFn,
        dense: true,
      });

      expect(wrapper.find(".q-field--dense").exists()).toBe(true);
    });

    it("calls autocompleteOptionsFn on filter event", async () => {
      const wrapper = mountComponent({
        autocompleteOptionsFn: mockAutocompleteFn,
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const updateFn = vi.fn((cb: () => void) => cb());
      await qSelect.vm.$emit("filter", "test", updateFn);
      await flushPromises();

      expect(mockAutocompleteFn).toHaveBeenCalledWith("test");
    });

    it("calls autocompleteOptionsFn with empty string when filter query is empty", async () => {
      const wrapper = mountComponent({
        autocompleteOptionsFn: mockAutocompleteFn,
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const updateFn = vi.fn((cb: () => void) => cb());
      await qSelect.vm.$emit("filter", "", updateFn);
      await flushPromises();

      expect(mockAutocompleteFn).toHaveBeenCalledWith("");
    });

    it("handles autocompleteOptionsFn errors gracefully", async () => {
      const errorFn = vi.fn().mockRejectedValue(new Error("Network error"));
      const wrapper = mountComponent({
        autocompleteOptionsFn: errorFn,
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const updateFn = vi.fn((cb: () => void) => cb());
      await qSelect.vm.$emit("filter", "test", updateFn);
      await flushPromises();

      // Should not throw, component handles error internally
      expect(errorFn).toHaveBeenCalled();
      expect(wrapper.find(".q-select").exists()).toBe(true);
    });
  });

  describe("object-based search mode (searchOptionsFn)", () => {
    const mockOptions = [
      { label: "Engineering", value: "dept-1" },
      { label: "Research", value: "dept-2" },
      { label: "Marketing", value: "dept-3" },
    ];
    const mockSearchFn = vi.fn().mockResolvedValue(mockOptions);

    it("renders a q-select when searchOptionsFn is provided", () => {
      const wrapper = mountComponent({
        searchOptionsFn: mockSearchFn,
        label: "Department",
        aid: "search-test",
      });

      expect(wrapper.find(".q-select").exists()).toBe(true);
      expect(wrapper.find(".q-input").exists()).toBe(false);
    });

    it("passes label and aid to q-select", () => {
      const wrapper = mountComponent({
        searchOptionsFn: mockSearchFn,
        label: "Department",
        aid: "search-test",
      });

      expect(wrapper.find("[aid='search-test']").exists()).toBe(true);
      expect(wrapper.html()).toContain("Department");
    });

    it("passes dense prop to q-select in search mode", () => {
      const wrapper = mountComponent({
        searchOptionsFn: mockSearchFn,
        dense: true,
      });

      expect(wrapper.find(".q-field--dense").exists()).toBe(true);
    });

    it("calls searchOptionsFn on filter event", async () => {
      const wrapper = mountComponent({
        searchOptionsFn: mockSearchFn,
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const updateFn = vi.fn((cb: () => void) => cb());
      await qSelect.vm.$emit("filter", "eng", updateFn);
      await flushPromises();

      expect(mockSearchFn).toHaveBeenCalledWith("eng");
    });

    it("calls searchOptionsFn with empty string when filter query is empty", async () => {
      const wrapper = mountComponent({
        searchOptionsFn: mockSearchFn,
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const updateFn = vi.fn((cb: () => void) => cb());
      await qSelect.vm.$emit("filter", "", updateFn);
      await flushPromises();

      expect(mockSearchFn).toHaveBeenCalledWith("");
    });

    it("handles searchOptionsFn errors gracefully", async () => {
      const errorFn = vi.fn().mockRejectedValue(new Error("API error"));
      const wrapper = mountComponent({
        searchOptionsFn: errorFn,
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const updateFn = vi.fn((cb: () => void) => cb());
      await qSelect.vm.$emit("filter", "test", updateFn);
      await flushPromises();

      expect(errorFn).toHaveBeenCalled();
      expect(wrapper.find(".q-select").exists()).toBe(true);
    });

    it("seeds options with current selection so fill-input shows the label", () => {
      const wrapper = mountComponent({
        searchOptionsFn: mockSearchFn,
        modelValue: "dept-1",
        displayValue: "Engineering",
      });

      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const options = qSelect.props("options") as Array<{ label: string; value: string }>;
      expect(options.some((o) => o.value === "dept-1" && o.label === "Engineering")).toBe(true);
    });

    it("emits update:model-value when an option is selected", async () => {
      const wrapper = mountComponent({
        searchOptionsFn: mockSearchFn,
      });

      // Simulate q-select emitting the selected value by calling the component's handler directly
      const qSelect = wrapper.findComponent({ name: "QSelect" });
      qSelect.vm.$emit("update:modelValue", "dept-2");
      await flushPromises();

      expect(wrapper.emitted("update:model-value")).toBeTruthy();
      expect(wrapper.emitted("update:model-value")?.[0]).toEqual(["dept-2"]);
    });
  });

  describe("priority: searchOptionsFn takes precedence over autocompleteOptionsFn", () => {
    it("uses search mode when both fns are provided", async () => {
      const searchFn = vi.fn().mockResolvedValue([{ label: "Result", value: "r-1" }]);
      const autocompleteFn = vi.fn().mockResolvedValue(["Result"]);

      const wrapper = mountComponent({
        searchOptionsFn: searchFn,
        autocompleteOptionsFn: autocompleteFn,
      });

      // Trigger filter and verify searchOptionsFn is called (not autocompleteOptionsFn)
      const qSelect = wrapper.findComponent({ name: "QSelect" });
      const updateFn = vi.fn((cb: () => void) => cb());
      await qSelect.vm.$emit("filter", "test", updateFn);
      await flushPromises();

      expect(searchFn).toHaveBeenCalledWith("test");
      expect(autocompleteFn).not.toHaveBeenCalled();
    });
  });
});
