import type { ISecurityModel } from "@/models/environment.model";
import { UidGidSource } from "@/swagger-models/assets-service-client";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import type { DefineComponent } from "vue";
import { FormSecuritySection } from ".";
import { FormSecuritySettings } from "./form-security-settings";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { errorMessages } from "@/common/error-message.constant";
describe("form-security-section.vue", () => {
  let formSecuritySectionCmp: VueWrapper<InstanceType<typeof FormSecuritySection>>;

  beforeEach(() => {
    formSecuritySectionCmp = mount(FormSecuritySection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        isSSO: false,
        securityModel: {
          uidGidSource: UidGidSource.FromTheImage,
          runAsGid: null,
          runAsUid: null,
          supplementalGroups: "",
          capabilities: [],
        } as ISecurityModel,
      },
    });
  });

  test("component is loading properly", () => {
    expect(formSecuritySectionCmp.html()).toContain("Security");
  });

  test("should have 2 options when isSSO is false", () => {
    expect(formSecuritySectionCmp.findAll("input[type=radio]").length).toBe(2);
  });

  test("should have 3 options when isSSO is true", async () => {
    await formSecuritySectionCmp.setProps({ isSSO: true });
    expect(formSecuritySectionCmp.findAll("input[type=radio]").length).toBe(3);
  });

  test("should display the SSO security options in the expected order", async () => {
    await formSecuritySectionCmp.setProps({ isSSO: true });
    const labels = formSecuritySectionCmp.findAll(".q-radio__label").map((label) => label.text());
    expect(labels).toEqual(["From the IdP token", "From the image", "Custom"]);
  });

  test("should not render custom options section when uidGidSource is not custom", async () => {
    expect(formSecuritySectionCmp.find(".custom-option-section").exists()).toBe(false);
  });

  test("should render custom options section when uidGidSource is custom", async () => {
    await formSecuritySectionCmp.setProps({
      securityModel: { uidGidSource: UidGidSource.Custom },
    });
    expect(formSecuritySectionCmp.find(".custom-option-section").exists()).toBe(true);
  });

  test("should emit securityModel when uidGidSource is changed", async () => {
    await formSecuritySectionCmp.findAll("input[type=radio]")[1].trigger("click");
    expect(formSecuritySectionCmp.emitted("update")).toBeTruthy();
  });

  test("should emit securityModel when runAsUid is changed", async () => {
    await formSecuritySectionCmp.setProps({
      securityModel: { uidGidSource: UidGidSource.Custom },
    });
    await formSecuritySectionCmp.find("[aid=uid-input]").setValue(1000);
    expect(formSecuritySectionCmp.emitted("update")).toBeTruthy();
  });

  test("should emit securityModel when runAsGid is changed", async () => {
    await formSecuritySectionCmp.setProps({
      securityModel: { uidGidSource: UidGidSource.Custom },
    });
    await formSecuritySectionCmp.find("[aid=gid-input]").setValue(1000);
    expect(formSecuritySectionCmp.emitted("update")).toBeTruthy();
  });

  test("should emit securityModel when supplementalGroups is changed", async () => {
    await formSecuritySectionCmp.setProps({
      securityModel: { uidGidSource: UidGidSource.Custom, runAsGid: 1000 },
    });
    await formSecuritySectionCmp.find("[aid=supplemental-groups-input]").setValue("1000, 1001");
    expect(formSecuritySectionCmp.emitted("update")?.length).toBeTruthy();
  });

  test("should explain how to enable supplementary groups", async () => {
    await formSecuritySectionCmp.setProps({
      securityModel: { uidGidSource: UidGidSource.Custom, runAsGid: null },
    });
    const settings = formSecuritySectionCmp.findComponent(FormSecuritySettings);
    expect(settings.vm.supplementalGroupsTooltip).toBe("To modify this value, enter a GID first.");
  });

  test("should expose validation messages for UID, GID, and supplementary groups", async () => {
    await formSecuritySectionCmp.setProps({
      securityModel: { uidGidSource: UidGidSource.Custom, runAsGid: 1000 },
    });
    const settings = formSecuritySectionCmp.findComponent(FormSecuritySettings);
    const numberInputs = settings.findAllComponents(PolicyNumberInput);
    expect(numberInputs).toHaveLength(2);
    expect(numberInputs.every((input) => input.props("isInteger"))).toBe(true);
    expect(numberInputs[0].vm.minRule(-1)).toBe("Enter a value equal or above 0");
    expect(numberInputs[1].vm.minRule(-1)).toBe("Enter a value equal or above 0");
    expect(settings.vm.isSupplementalGroupsValid("group-one")).toBe(errorMessages.SUPPLEMENTAL_GROUPS_NOT_VALID);
  });

  test("should have all the values in the inputs as got in the props", async () => {
    await formSecuritySectionCmp.setProps({
      securityModel: {
        uidGidSource: UidGidSource.Custom,
        runAsGid: 1000,
        runAsUid: 1000,
        supplementalGroups: "1000, 1001",
      },
    });
    expect((formSecuritySectionCmp.find("[aid=uid-input]").element as HTMLInputElement).value).toBe("1000");
    expect((formSecuritySectionCmp.find("[aid=gid-input]").element as HTMLInputElement).value).toBe("1000");
    expect((formSecuritySectionCmp.find("[aid=supplemental-groups-input]").element as HTMLInputElement).value).toBe(
      "1000, 1001",
    );
  });

  test("should emit securityModel when capabilities are updated", async () => {
    formSecuritySectionCmp.findComponent<DefineComponent>(".capabilities-select").vm.$emit("update:modelValue", [
      { label: "AUDIT_CONTROL", value: "AUDIT_CONTROL" },
      { label: "AUDIT_READ", value: "AUDIT_READ" },
      { label: "AUDIT_WRITE", value: "AUDIT_WRITE" },
      { label: "BLOCK_SUSPEND", value: "BLOCK_SUSPEND" },
    ]);
    await formSecuritySectionCmp.vm.$nextTick();
    expect(formSecuritySectionCmp.emitted("update")?.length).toBeTruthy();
  });

  test("the capabilities select should render the existing capabilities loaded from the model props", async () => {
    const loadedModel: ISecurityModel = {
      capabilities: ["AUDIT_CONTROL", "AUDIT_READ", "BLOCK_SUSPEND"],
    };
    // Mounting comopnent again, because this logic only happens inside created
    formSecuritySectionCmp = mount(FormSecuritySection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        isSSO: false,
        securityModel: { ...loadedModel } as ISecurityModel,
      },
    });
    const capabilitiesSelect = formSecuritySectionCmp.findComponent<DefineComponent>(".capabilities-select");
    expect(capabilitiesSelect.vm.modelValue).toEqual(
      loadedModel.capabilities?.map((c) => ({ label: c, value: c })) || [],
    );
    expect(formSecuritySectionCmp.findComponent<DefineComponent>(".capabilities-select").text()).toContain(
      "AUDIT_CONTROL",
    );
    expect(formSecuritySectionCmp.findComponent<DefineComponent>(".capabilities-select").text()).toContain("AUDIT_READ");
    expect(formSecuritySectionCmp.findComponent<DefineComponent>(".capabilities-select").text()).toContain(
      "BLOCK_SUSPEND",
    );
  });
});
