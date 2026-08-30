import { mount } from "@vue/test-utils";
import NimProfileSelector from "./nim-profile-selector.vue";
import { EProfileNameOrHashOptions } from "./nim-profile.model";
import { describe, it, expect, vi } from "vitest";
import { Quasar } from "quasar";
import type { TTestEmit } from "@/models/unit-test.model";
import { errorMessages } from "@/common/error-message.constant";

describe("nim-profile-selector.vue", () => {
  const createWrapper = function (
    props: { profileNameOrHash?: string; isProfileRequired?: boolean } = { profileNameOrHash: "" },
  ) {
    return mount(NimProfileSelector, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  };

  it("renders correctly", () => {
    const wrapper = createWrapper();

    // Check for RunaiRadioOptions
    const radioOptions = wrapper.findComponent({ name: "RunaiRadioOptions" });
    expect(radioOptions.exists()).toBe(true);
    expect(radioOptions.props("modelValue")).toBe(EProfileNameOrHashOptions.Automatically);

    // Check for PolicyStringField
    const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
    expect(policyStringField.exists()).toBe(false); // Should not show initially
  });

  it("shows PolicyStringField when 'Manually' option is selected", async () => {
    const wrapper = createWrapper();
    const radioOptions = wrapper.findComponent({ name: "RunaiRadioOptions" });

    // Simulate selection change
    await radioOptions.vm.$emit("update:model-value", EProfileNameOrHashOptions.Manually);

    expect(wrapper.findComponent({ name: "PolicyStringField" }).exists()).toBe(true);
  });

  it("validates empty profile input", () => {
    const wrapper = createWrapper();
    const result = wrapper.vm.profileInputNotEmpty("");

    expect(result).toBe(errorMessages.PROFILE_NAME_OR_HASH_NOT_EMPTY);
  });

  it("validates invalid profile input format", () => {
    const wrapper = createWrapper();
    const result = wrapper.vm.profileIsValid("invalid!");

    expect(result).toBe(errorMessages.ENTER_A_VALID_PROFILE_NAME_OR_HASH);
  });

  it("emits 'nim-profile-changed' and 'is-valid' when profile is updated manually", async () => {
    const wrapper = createWrapper({ profileNameOrHash: "test-profile" });

    // Set profile to "Manually"
    await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

    // Simulate input change
    const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
    await policyStringField.vm.$emit("update:model-value", "test-profile");

    const nimProfileChangedResult: TTestEmit<string[]> = wrapper.emitted("nim-profile-changed") || [];
    expect(nimProfileChangedResult).toHaveLength(1);
    expect(nimProfileChangedResult[0]).toEqual(["test-profile"]);

    const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
    expect(validationResult).toHaveLength(1);
    expect(validationResult[0]).toEqual([true]);
  });

  it("emits 'is-valid' as false for an invalid profile name", async () => {
    const wrapper = createWrapper({ profileNameOrHash: "invalid!" });

    // Set profile to "Manually"
    await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

    // Simulate input change
    const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
    await policyStringField.vm.$emit("update:model-value", "invalid!");

    const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
    expect(validationResult).toHaveLength(1);
    expect(validationResult[0]).toEqual([false]);
  });

  it("resets profile when 'Automatically' is selected", async () => {
    const wrapper = createWrapper({ profileNameOrHash: "manual-profile" });

    // Set profile to "Automatically"
    const radioOptions = wrapper.findComponent({ name: "RunaiRadioOptions" });
    await radioOptions.vm.$emit("update:model-value", EProfileNameOrHashOptions.Automatically.toString());

    const nimProfileChangedResult: TTestEmit<string[]> = wrapper.emitted("nim-profile-changed") || [];
    expect(nimProfileChangedResult).toHaveLength(1);
    expect(nimProfileChangedResult[0]).toEqual([""]);

    const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
    expect(validationResult).toHaveLength(2);
    expect(validationResult).toEqual([[false], [true]]); // the first emit is from the emit function that validates the empty string
  });

  it("initializes localProfileNameOrHash from profileNameOrHash", () => {
    const wrapper = createWrapper({ profileNameOrHash: "manual-profile" });

    expect(wrapper.vm.localProfileNameOrHash).toBe("manual-profile");
    expect(wrapper.vm.profileMethod).toBe(EProfileNameOrHashOptions.Manually);
  });

  it("defaults localProfileNameOrHash to an empty string when profileNameOrHash is not provided", () => {
    const wrapper = createWrapper({ profileNameOrHash: "" });

    expect(wrapper.vm.localProfileNameOrHash).toBe("");
    expect(wrapper.vm.profileMethod).toBe(EProfileNameOrHashOptions.Automatically);
  });

  it("updates localProfileNameOrHash and emits changes on input change", async () => {
    const wrapper = createWrapper({ profileNameOrHash: "initial-hash" });
    const emitSpy = vi.spyOn(wrapper.vm, "emitChanges");

    // Simulate input change
    wrapper.vm.onProfileInputChanged("updated-hash");

    await wrapper.vm.$nextTick();

    expect(wrapper.vm.localProfileNameOrHash).toBe("updated-hash");
    expect(emitSpy).toHaveBeenCalledWith("updated-hash");
  });

  it("calls emitChanges with an empty string when profileMethod changes to Automatically", async () => {
    const wrapper = createWrapper({ profileNameOrHash: "some-hash" });
    const emitSpy = vi.spyOn(wrapper.vm, "emitChanges");

    // Simulate profile selection change
    wrapper.vm.onProfileSelectionChanged(EProfileNameOrHashOptions.Automatically);

    await wrapper.vm.$nextTick();

    expect(wrapper.vm.profileMethod).toBe(EProfileNameOrHashOptions.Automatically);
    expect(wrapper.vm.localProfileNameOrHash).toBe("some-hash"); // localProfileNameOrHash should remain unchanged
    expect(emitSpy).toHaveBeenCalledWith("");
  });

  it("retains localProfileNameOrHash when switching back to Manually after selecting Automatically", async () => {
    const wrapper = createWrapper({ profileNameOrHash: "" });
    const emitSpy = vi.spyOn(wrapper.vm, "emitChanges");

    // Step 1: Start with an empty profileNameOrHash
    expect(wrapper.vm.localProfileNameOrHash).toBe("");
    expect(wrapper.vm.profileMethod).toBe(EProfileNameOrHashOptions.Automatically);

    // Step 2: Change to Manually
    wrapper.vm.onProfileSelectionChanged(EProfileNameOrHashOptions.Manually);
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.profileMethod).toBe(EProfileNameOrHashOptions.Manually);
    expect(emitSpy).toHaveBeenCalledWith("");

    // Step 3: Enter a text
    wrapper.vm.onProfileInputChanged("test-input");
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.localProfileNameOrHash).toBe("test-input");
    expect(emitSpy).toHaveBeenCalledWith("test-input");

    // Step 4: Select Automatically
    wrapper.vm.onProfileSelectionChanged(EProfileNameOrHashOptions.Automatically);
    expect(emitSpy).toHaveBeenCalledWith("");
    await wrapper.vm.$nextTick();

    // Step 5: Select Manually again
    wrapper.vm.onProfileSelectionChanged(EProfileNameOrHashOptions.Manually);
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.profileMethod).toBe(EProfileNameOrHashOptions.Manually);
    expect(wrapper.vm.localProfileNameOrHash).toBe("test-input"); // Text from step 3 should still be present
    expect(emitSpy).toHaveBeenCalledWith("test-input");
  });

  describe("isProfileRequired prop", () => {
    it("defaults to true when not provided", () => {
      const wrapper = createWrapper({ profileNameOrHash: "" });
      expect(wrapper.props("isProfileRequired")).toBe(true);
      expect(wrapper.vm.profileValidationRules).toHaveLength(2);
    });

    it("uses profileInputNotEmpty and profileIsValid validation when isProfileRequired is true", () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: true });

      // Should require non-empty value
      expect(wrapper.vm.profileInputNotEmpty("")).toBe(errorMessages.PROFILE_NAME_OR_HASH_NOT_EMPTY);
      expect(wrapper.vm.profileInputNotEmpty("some-value")).toBe(true);

      // Should validate format
      expect(wrapper.vm.profileIsValid("valid-profile")).toBe(true);
      expect(wrapper.vm.profileIsValid("invalid!")).toBe(errorMessages.ENTER_A_VALID_PROFILE_NAME_OR_HASH);

      expect(wrapper.vm.profileValidationRules).toHaveLength(2);

      // Should validate format when value is provided
      expect(wrapper.vm.profileInputNotRequired("valid-profile")).toBe(true);
      expect(wrapper.vm.profileInputNotRequired("invalid!")).toBe(errorMessages.ENTER_A_VALID_PROFILE_NAME_OR_HASH);
    });

    it("uses profileInputNotRequired validation when isProfileRequired is false", () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: false });

      // Should return true for empty string (not required)
      expect(wrapper.vm.profileInputNotRequired("")).toBe(true);

      // Should validate format when value is provided
      expect(wrapper.vm.profileInputNotRequired("valid-profile")).toBe(true);
      expect(wrapper.vm.profileInputNotRequired("invalid!")).toBe(errorMessages.ENTER_A_VALID_PROFILE_NAME_OR_HASH);
    });

    it("sets correct validation rules when isProfileRequired is true", () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: true });

      expect(wrapper.vm.profileValidationRules).toHaveLength(2);
      expect(wrapper.vm.profileValidationRules[0]).toBe(wrapper.vm.profileInputNotEmpty);
      expect(wrapper.vm.profileValidationRules[1]).toBe(wrapper.vm.profileIsValid);
    });

    it("sets correct validation rules when isProfileRequired is false", () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: false });

      expect(wrapper.vm.profileValidationRules).toHaveLength(1);
      expect(wrapper.vm.profileValidationRules[0]).toBe(wrapper.vm.profileInputNotRequired);
    });

    it("passes validation rules to PolicyStringField when isProfileRequired is true", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: true });

      // Set profile to "Manually" to show the PolicyStringField
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      expect(policyStringField.props("rules")).toEqual([wrapper.vm.profileInputNotEmpty, wrapper.vm.profileIsValid]);
    });

    it("passes validation rules to PolicyStringField when isProfileRequired is false", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: false });

      // Set profile to "Manually" to show the PolicyStringField
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      expect(policyStringField.props("rules")).toEqual([wrapper.vm.profileInputNotRequired]);
    });

    it("emits 'is-valid' as false for empty string when isProfileRequired is true", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: true });

      // Set profile to "Manually"
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      // Simulate empty input
      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      await policyStringField.vm.$emit("update:model-value", "");

      const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
      expect(validationResult).toHaveLength(1);
      expect(validationResult[0]).toEqual([false]); // emitChanges always validates with isNotEmpty
    });

    it("emits 'is-valid' as false for empty string when isProfileRequired is false", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: false });

      // Set profile to "Manually"
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      // Simulate empty input
      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      await policyStringField.vm.$emit("update:model-value", "");

      const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
      expect(validationResult).toHaveLength(1);
      expect(validationResult[0]).toEqual([false]); // Should be invalid when required
    });

    it("emits 'is-valid' as true for valid profile when isProfileRequired is true", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: true });

      // Set profile to "Manually"
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      // Simulate valid input
      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      await policyStringField.vm.$emit("update:model-value", "valid-profile");

      const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
      expect(validationResult).toHaveLength(1);
      expect(validationResult[0]).toEqual([true]);
    });

    it("emits 'is-valid' as true for valid profile when isProfileRequired is false", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: false });

      // Set profile to "Manually"
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      // Simulate valid input
      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      await policyStringField.vm.$emit("update:model-value", "valid-profile");

      const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
      expect(validationResult).toHaveLength(1);
      expect(validationResult[0]).toEqual([true]);
    });

    it("emits 'is-valid' as false for invalid profile when isProfileRequired is true", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: true });

      // Set profile to "Manually"
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      // Simulate invalid input
      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      await policyStringField.vm.$emit("update:model-value", "invalid!");

      const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
      expect(validationResult).toHaveLength(1);
      expect(validationResult[0]).toEqual([false]);
    });

    it("emits 'is-valid' as false for invalid profile when isProfileRequired is false", async () => {
      const wrapper = createWrapper({ profileNameOrHash: "", isProfileRequired: false });

      // Set profile to "Manually"
      await wrapper.setData({ profileMethod: EProfileNameOrHashOptions.Manually });

      // Simulate invalid input
      const policyStringField = wrapper.findComponent({ name: "PolicyStringField" });
      await policyStringField.vm.$emit("update:model-value", "invalid!");

      const validationResult: TTestEmit<string[]> = wrapper.emitted("is-valid") || [];
      expect(validationResult).toHaveLength(1);
      expect(validationResult[0]).toEqual([false]);
    });
  });
});
