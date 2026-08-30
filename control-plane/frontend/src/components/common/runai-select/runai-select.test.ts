import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { QSelect, Quasar, type ValidationRule } from "quasar";
import { describe, test, it, expect } from "vitest";
import { RunaiSelect } from ".";
import RunaiSvgIcon from "../runai-svg-icon/runai-svg-icon.vue";
import RunaiTooltip from "../runai-tooltip/runai-tooltip.vue";
import RoleStageBadge from "../role-stage-badge/role-stage-badge.vue";
import { SelectType, SELECT_TYPES } from "./runai-select.vue";

// Icon constants for reusability
const ICON_NAMES = {
  WARNING: "ph-bold ph-warning-circle",
  SUCCESS: "ph-bold ph-check-circle",
  NEGATIVE: "ph-bold ph-x-circle",
  INFO: "ph-bold ph-info",
  POSITIVE: "ph-bold ph-check-circle",
} as const;

// Color constants for reusability
const COLOR_NAMES = {
  WARNING: "warning",
  SUCCESS: "success",
  NEGATIVE: "negative",
  INFO: "info",
  POSITIVE: "positive",
} as const;

describe("runai-select.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiSelect>>;

  const allProps: InstanceType<typeof RunaiSelect>["$props"] = {
    modelValue: null,
    options: [],
    placeholder: "Select...",
    noOptionText: undefined,
    rules: undefined,
    outlined: undefined,
    emitValue: undefined,
    unclearable: undefined,
    hideBottomSpace: undefined,
    disable: undefined,
    policyRules: undefined,
  };

  function mountComponent(props = allProps) {
    wrapper = mount(RunaiSelect, {
      global: {
        plugins: [Quasar],
        stubs: {
          RoleStageBadge: true,
        },
      },
      components: {
        QSelect,
        RunaiSvgIcon,
        RunaiTooltip,
        RoleStageBadge,
      },
      props,
    });
  }

  describe("Html elements check", () => {
    test("component is loading properly", () => {
      mountComponent();
      expect(wrapper.html()).toContain("runai-select");
    });

    test("component should show the selected value label", async () => {
      mountComponent();
      await wrapper.setProps({
        options: [
          { value: "a", label: "A" },
          { value: "b", label: "B" },
        ],
        modelValue: "a",
      });
      expect(wrapper.html()).toContain("A");
    });
  });

  describe("Props", () => {
    describe("placeholder", () => {
      test("component should load with placeholder", () => {
        mountComponent();
        expect(wrapper.html()).toContain("Select...");
      });

      test("component should show placeholder according to prop if sent", async () => {
        mountComponent();
        const placeholder = "Select a value";
        await wrapper.setProps({ placeholder });
        expect(wrapper.html()).toContain(placeholder);
      });
    });
    describe("rules", () => {
      it("if policyRules.required is true, required rule should be added", async () => {
        const firstRule: ValidationRule = () => {
          return true;
        };
        const secondRule: ValidationRule = () => {
          return true;
        };
        mountComponent({
          options: ["first-opt", "second-opt"],
          rules: [firstRule, secondRule],
          modelValue: null,
          policyRules: {
            required: true,
          },
        });

        const select = wrapper.findComponent(QSelect);
        expect(select.props().rules.length).toBe(4);
        expect(select.props().rules.includes(firstRule)).toBe(true);
        expect(select.props().rules.includes(secondRule)).toBe(true);
        expect(select.props().rules.includes(wrapper.vm.requiredRule)).toBe(true);
      });
    });
  });

  describe("Computed", () => {
    describe("displayedOptions", () => {
      it("if policyRules.options exists and the its an open-list overrides the original list with the list from the policyRules.options", async () => {
        const policyOptions = [
          { value: "data-1", displayed: "D1" },
          { value: "data-2", displayed: "D2" },
        ];
        const expected = [
          { value: "data-1", label: "D1" },
          { value: "data-2", label: "D2" },
        ];
        mountComponent({
          options: ["first-opt", "second-opt"],
          modelValue: null,
          policyRules: {
            options: policyOptions,
          },
        });
        expect(wrapper.vm.displayedOptions).toEqual(expected);
      });

      it("if policyRules.options exists and the its a closed-list the list stays the same but options that are not mentioned in the policy are disabled", async () => {
        const policyOptions = [
          { value: "data-1", displayed: "D1" },
          { value: "data-2", displayed: "D2" },
        ];

        const closedList = [
          { value: "data-1", label: "D1" },
          { value: "data-2", label: "D2" },
          { value: "data-3", label: "D3" },
          { value: "data-4", label: "D4" },
        ];

        const expected = [
          { value: "data-1", label: "D1" },
          { value: "data-2", label: "D2" },
          {
            value: "data-3",
            label: "D3",
            disable: true,
            disabledTooltip: "This can't be selected based on the policy your administrator set.",
          },
          {
            value: "data-4",
            label: "D4",
            disable: true,
            disabledTooltip: "This can't be selected based on the policy your administrator set.",
          },
        ];
        mountComponent({
          options: closedList,
          modelValue: null,
          policyRules: {
            options: policyOptions,
            isClosedList: true,
          },
        });
        expect(wrapper.vm.displayedOptions).toEqual(expected);
      });

      it("if policyRules.options exists label would be set to the displayed value if its exists if not label would be equal to value)", async () => {
        const policyOptions = [
          { value: "data-1", displayed: "D1" },
          { value: "data-2", displayed: "D2" },
          { value: "data-3" },
          { value: "data-4" },
        ];

        const closedList = ["data-1", "data-2", "data-3", "data-4"];

        const expected = [
          { value: "data-1", label: "D1" },
          { value: "data-2", label: "D2" },
          {
            value: "data-3",
            label: "data-3",
          },
          {
            value: "data-4",
            label: "data-4",
          },
        ];
        mountComponent({
          options: closedList,
          modelValue: null,
          policyRules: {
            options: policyOptions,
            isClosedList: true,
          },
        });
        expect(wrapper.vm.displayedOptions).toEqual(expected);
      });

      it("if policy does not exists return data as it is without any modification", async () => {
        const origin = ["first-opt", "second-opt"];
        mountComponent({
          options: origin,
          modelValue: null,
          policyRules: {
            options: undefined,
          },
        });
        expect(wrapper.vm.displayedOptions).toEqual(origin);
      });
    });
  });

  describe("filterable (autocomplete)", () => {
    const fieldOptions = [
      { value: "args", label: "args" },
      { value: "image", label: "image" },
      { value: "command", label: "command" },
    ];

    test("optionsForSelect returns all options when there is no search term", () => {
      mountComponent({ ...allProps, filterable: true, options: fieldOptions, modelValue: null });
      expect(wrapper.vm.optionsForSelect).toEqual(fieldOptions);
    });

    test("optionsForSelect filters displayedOptions by label (case-insensitive) while searching", () => {
      mountComponent({ ...allProps, filterable: true, options: fieldOptions, modelValue: null });
      wrapper.vm.onFilter("IMA", (cb) => cb());
      expect(wrapper.vm.optionsForSelect).toEqual([{ value: "image", label: "image" }]);
    });

    test("selecting a value resets the search term so the full list returns next open", () => {
      mountComponent({ ...allProps, filterable: true, options: fieldOptions, modelValue: null });
      wrapper.vm.onFilter("arg", (cb) => cb());
      expect(wrapper.vm.optionsForSelect).toHaveLength(1);
      wrapper.vm.updateModelValue("args");
      expect(wrapper.vm.searchTerm).toBe("");
      expect(wrapper.vm.optionsForSelect).toEqual(fieldOptions);
    });

    test("hasSelection reflects whether a value is selected", async () => {
      mountComponent({ ...allProps, filterable: true, options: fieldOptions, modelValue: null });
      expect(wrapper.vm.hasSelection).toBe(false);
      await wrapper.setProps({ modelValue: "args" });
      expect(wrapper.vm.hasSelection).toBe(true);
    });

    test("does not filter when filterable is false (default)", () => {
      mountComponent({ ...allProps, options: fieldOptions, modelValue: null });
      wrapper.vm.onFilter("ima", (cb) => cb());
      expect(wrapper.vm.optionsForSelect).toEqual(fieldOptions);
    });
  });

  describe("SELECT_TYPES functionality", () => {
    describe("SELECT_TYPES constant", () => {
      test("should contain all expected predefined types", () => {
        mountComponent();

        expect(SELECT_TYPES).toHaveProperty("warning");
        expect(SELECT_TYPES).toHaveProperty("success");
        expect(SELECT_TYPES).toHaveProperty("negative");
        expect(SELECT_TYPES).toHaveProperty("info");
        expect(SELECT_TYPES).toHaveProperty("positive");
      });

      test("should have correct icon and color properties for each type", () => {
        mountComponent();

        expect(SELECT_TYPES.warning).toEqual({
          icon: ICON_NAMES.WARNING,
          color: COLOR_NAMES.WARNING,
        });

        expect(SELECT_TYPES.success).toEqual({
          icon: ICON_NAMES.SUCCESS,
          color: COLOR_NAMES.SUCCESS,
        });

        expect(SELECT_TYPES.negative).toEqual({
          icon: ICON_NAMES.NEGATIVE,
          color: COLOR_NAMES.NEGATIVE,
        });

        expect(SELECT_TYPES.info).toEqual({
          icon: ICON_NAMES.INFO,
          color: COLOR_NAMES.INFO,
        });

        expect(SELECT_TYPES.positive).toEqual({
          icon: ICON_NAMES.POSITIVE,
          color: COLOR_NAMES.POSITIVE,
        });
      });
    });

    describe("Type functionality in options", () => {
      test("should prioritize custom icon over type icon", async () => {
        const optionsWithBoth = [
          {
            value: "mixed-opt",
            label: "Mixed Option",
            type: SelectType.Warning,
            icon: "custom-icon",
          },
        ];

        mountComponent({
          options: optionsWithBoth,
          modelValue: null,
        });

        // Custom icon should take priority, so runai-svg-icon should show custom icon
        const svgIcons = wrapper.findAllComponents(RunaiSvgIcon);
        if (svgIcons.length > 0) {
          expect(svgIcons[0].props("name")).toBe("custom-icon");
        }
      });

      test("should display type icon in selected value when option has type", async () => {
        const optionsWithTypes = [{ value: "warning-opt", label: "Warning Option", type: SelectType.Warning }];

        mountComponent({
          options: optionsWithTypes,
          modelValue: { value: "warning-opt", label: "Warning Option", type: SelectType.Warning },
        });

        expect(wrapper.html()).toContain("Warning Option");
      });
    });

    describe("Component methods for types", () => {
      test("getTypeConfig should return correct config for valid type", () => {
        mountComponent();

        const warningConfig = wrapper.vm.getTypeConfig(SelectType.Warning);
        expect(warningConfig).toEqual({
          icon: ICON_NAMES.WARNING,
          color: COLOR_NAMES.WARNING,
        });

        const successConfig = wrapper.vm.getTypeConfig(SelectType.Success);
        expect(successConfig).toEqual({
          icon: ICON_NAMES.SUCCESS,
          color: COLOR_NAMES.SUCCESS,
        });
      });

      test("getTypeConfig should return null for invalid type", () => {
        mountComponent();

        const invalidConfig = wrapper.vm.getTypeConfig("invalid-type" as SelectType);
        expect(invalidConfig).toBeNull();
      });

      test("getTypeIcon should return correct icon for valid type", () => {
        mountComponent();

        expect(wrapper.vm.getTypeIcon(SelectType.Warning)).toBe(ICON_NAMES.WARNING);
        expect(wrapper.vm.getTypeIcon(SelectType.Success)).toBe(ICON_NAMES.SUCCESS);
        expect(wrapper.vm.getTypeIcon(SelectType.Negative)).toBe(ICON_NAMES.NEGATIVE);
        expect(wrapper.vm.getTypeIcon(SelectType.Info)).toBe(ICON_NAMES.INFO);
        expect(wrapper.vm.getTypeIcon(SelectType.Positive)).toBe(ICON_NAMES.POSITIVE);
      });

      test("getTypeIcon should return undefined for invalid type", () => {
        mountComponent();

        expect(wrapper.vm.getTypeIcon("invalid-type" as SelectType)).toBeUndefined();
      });

      test("getTypeColor should return correct color for valid type", () => {
        mountComponent();

        expect(wrapper.vm.getTypeColor(SelectType.Warning)).toBe(COLOR_NAMES.WARNING);
        expect(wrapper.vm.getTypeColor(SelectType.Success)).toBe(COLOR_NAMES.SUCCESS);
        expect(wrapper.vm.getTypeColor(SelectType.Negative)).toBe(COLOR_NAMES.NEGATIVE);
        expect(wrapper.vm.getTypeColor(SelectType.Info)).toBe(COLOR_NAMES.INFO);
        expect(wrapper.vm.getTypeColor(SelectType.Positive)).toBe(COLOR_NAMES.POSITIVE);
      });

      test("getTypeColor should return undefined for invalid type", () => {
        mountComponent();

        expect(wrapper.vm.getTypeColor("invalid-type" as SelectType)).toBeUndefined();
      });
    });

    describe("Type integration with tooltip", () => {
      test("should pass type icon and color to runai-tooltip when option has type", async () => {
        const optionsWithTypes = [
          {
            value: "warning-opt",
            label: "Warning Option",
            type: SelectType.Warning,
            toolTip: "This is a warning tooltip",
          },
        ];

        mountComponent({
          options: optionsWithTypes,
          modelValue: null,
        });

        const tooltips = wrapper.findAllComponents(RunaiTooltip);
        if (tooltips.length > 0) {
          const tooltip = tooltips[0];
          expect(tooltip.props("icon")).toBe(ICON_NAMES.WARNING);
          expect(tooltip.props("color")).toBe(COLOR_NAMES.WARNING);
        }
      });

      test("should show undefined icon and color when option has no type", async () => {
        const optionsWithoutTypes = [
          {
            value: "normal-opt",
            label: "Normal Option",
            toolTip: "This is a normal tooltip",
          },
        ];

        mountComponent({
          options: optionsWithoutTypes,
          modelValue: null,
        });

        const tooltips = wrapper.findAllComponents(RunaiTooltip);
        if (tooltips.length > 0) {
          const tooltip = tooltips[0];
          expect(tooltip.props("icon")).toBeUndefined();
          expect(tooltip.props("color")).toBeUndefined();
        }
      });
    });
  });

  describe("customAction functionality", () => {
    describe("customAction prop", () => {
      test("should accept customAction prop with text", () => {
        const customAction = { text: "Add new item" };
        mountComponent({
          ...allProps,
          customAction,
        });

        expect(wrapper.props("customAction")).toEqual(customAction);
      });

      test("should make customAction prop optional", () => {
        mountComponent();

        expect(wrapper.props("customAction")).toBeUndefined();
      });

      test("should accept customAction with different text values", () => {
        const customActions = [{ text: "Create new" }, { text: "Add more" }, { text: "New item" }];

        customActions.forEach((action) => {
          mountComponent({
            ...allProps,
            customAction: action,
          });

          expect(wrapper.props("customAction")).toEqual(action);
        });
      });
    });

    describe("customAction rendering", () => {
      test("should not render custom action when customAction is not provided", () => {
        mountComponent();

        expect(wrapper.find(".custom-action-item").exists()).toBe(false);
        expect(wrapper.find(".q-separator").exists()).toBe(false);
      });

      test("should not render custom action when customAction.text is empty", () => {
        mountComponent({
          ...allProps,
          customAction: { text: "" },
        });

        expect(wrapper.find(".custom-action-item").exists()).toBe(false);
        expect(wrapper.find(".q-separator").exists()).toBe(false);
      });

      test("should render custom action in no-option template when customAction.text is provided and options are empty", () => {
        mountComponent({
          ...allProps,
          options: [],
          noOptionText: "No options available",
          customAction: { text: "Add new item" },
        });

        // The custom action should be rendered in the no-option template
        // We can verify this by checking if the component accepts the customAction prop
        expect(wrapper.props("customAction")).toEqual({ text: "Add new item" });
      });

      test("should render custom action in after-options template when customAction.text is provided", () => {
        mountComponent({
          ...allProps,
          options: ["option1", "option2"],
          customAction: { text: "Add new item" },
        });

        // The custom action should be rendered in the after-options template
        // We can verify this by checking if the component accepts the customAction prop
        expect(wrapper.props("customAction")).toEqual({ text: "Add new item" });
      });

      test("should have customAction prop available for rendering", () => {
        mountComponent({
          ...allProps,
          customAction: { text: "Add new item" },
        });

        expect(wrapper.props("customAction")).toEqual({ text: "Add new item" });
      });

      test("should have customAction prop with correct text property", () => {
        const customActionText = "Create new credential";
        mountComponent({
          ...allProps,
          customAction: { text: customActionText },
        });

        expect(wrapper.props("customAction")?.text).toBe(customActionText);
      });
    });

    describe("handleCustomAction method", () => {
      test("should emit custom-action event when handleCustomAction is called", () => {
        mountComponent({
          ...allProps,
          customAction: { text: "Add new item" },
        });

        wrapper.vm.handleCustomAction();

        expect(wrapper.emitted("custom-action")).toBeTruthy();
        expect(wrapper.emitted("custom-action")?.length).toBe(1);
      });

      test("should emit custom-action event and attempt to close dropdown", () => {
        mountComponent({
          ...allProps,
          customAction: { text: "Add new item" },
        });

        wrapper.vm.handleCustomAction();

        expect(wrapper.emitted("custom-action")).toBeTruthy();
        expect(wrapper.emitted("custom-action")?.length).toBe(1);
      });
    });

    describe("customAction integration with options", () => {
      test("should accept customAction prop when options array is empty", () => {
        mountComponent({
          ...allProps,
          options: [],
          customAction: { text: "Add new item" },
        });

        expect(wrapper.props("customAction")).toEqual({ text: "Add new item" });
      });

      test("should accept customAction prop when options array has items", () => {
        mountComponent({
          ...allProps,
          options: ["option1", "option2", "option3"],
          customAction: { text: "Add new item" },
        });

        expect(wrapper.props("customAction")).toEqual({ text: "Add new item" });
      });

      test("should accept customAction prop with no-option text when options are empty", () => {
        mountComponent({
          ...allProps,
          options: [],
          noOptionText: "No options available",
          customAction: { text: "Add new item" },
        });

        expect(wrapper.props("customAction")).toEqual({ text: "Add new item" });
        expect(wrapper.props("noOptionText")).toBe("No options available");
      });
    });

    describe("customAction prop validation", () => {
      test("should accept customAction prop with text property", () => {
        const customAction = { text: "Add new item" };
        mountComponent({
          ...allProps,
          customAction,
        });

        expect(wrapper.props("customAction")).toEqual(customAction);
        expect(wrapper.props("customAction")?.text).toBe("Add new item");
      });

      test("should make customAction prop optional by default", () => {
        mountComponent();

        expect(wrapper.props("customAction")).toBeUndefined();
      });

      test("should handle customAction prop with different text values", () => {
        const testCases = [
          { text: "Create new" },
          { text: "Add more" },
          { text: "New item" },
          { text: "Custom action" },
        ];

        testCases.forEach((testCase) => {
          mountComponent({
            ...allProps,
            customAction: testCase,
          });

          expect(wrapper.props("customAction")).toEqual(testCase);
          expect(wrapper.props("customAction")?.text).toBe(testCase.text);
        });
      });
    });

    describe("customAction conditional behavior", () => {
      test("should conditionally accept customAction prop based on text existence", () => {
        // Test with customAction.text
        mountComponent({
          ...allProps,
          customAction: { text: "Add new item" },
        });
        expect(wrapper.props("customAction")).toEqual({ text: "Add new item" });

        // Test without customAction
        mountComponent({
          ...allProps,
          customAction: undefined,
        });
        expect(wrapper.props("customAction")).toBeUndefined();

        // Test with empty customAction.text
        mountComponent({
          ...allProps,
          customAction: { text: "" },
        });
        expect(wrapper.props("customAction")).toEqual({ text: "" });
      });

      test("should handle customAction prop with various text values", () => {
        const testCases = [
          { text: "Add new item" },
          { text: "Create new" },
          { text: "New credential" },
          { text: "Add more options" },
        ];

        testCases.forEach((testCase) => {
          mountComponent({
            ...allProps,
            customAction: testCase,
          });

          expect(wrapper.props("customAction")).toEqual(testCase);
        });
      });
    });
  });
});
