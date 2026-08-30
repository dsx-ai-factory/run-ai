import { mount } from "@vue/test-utils";
import GenericFormatSecretSection from "./generic-format-secret-section.vue";
import { beforeEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { IUIGenericSecretSpecDetails } from "@/models/credential.model";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { TYPE_SSH_KEY } from "@/models/credential.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";

describe("generic-format-secret-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof GenericFormatSecretSection>>;

  const existingSecrets: SecretListInfo[] = [
    { name: "secret-1", type: "generic" },
    { name: "secret-2", type: "generic" },
  ] as SecretListInfo[];

  const defaultProps = {
    secretDetails: {} as IUIGenericSecretSpecDetails,
    existingSecrets,
    loading: false,
    disabledFormats: [],
  };

  beforeEach(() => {
    wrapper = mount(GenericFormatSecretSection, {
      global: {
        plugins: [Quasar],
        stubs: {
          RunaiExpansionItem: true,
          RunaiSelect: true,
          RunaiRadioOptions: true,
          RunaiSection: true,
          GenericSecretKeyValueSection: true,
          GenericSecretInputPair: true,
        },
      },
      props: defaultProps,
    });
  });

  describe("Radio Options - Existing vs New Secret", () => {
    test("renders radio options for existing and new secret", () => {
      const radioOptions = wrapper.vm.radioOptions;
      expect(radioOptions).toHaveLength(2);
      expect(radioOptions[0].label).toBe("Existing secret");
      expect(radioOptions[0].value).toBe(true);
      expect(radioOptions[1].label).toBe("New secret");
      expect(radioOptions[1].value).toBe(false);
    });

    test("defaults to existing secret mode", () => {
      expect(wrapper.vm.isExistingSecret).toBe(true);
    });

    test("switches to new secret mode when updateFormState is called with false", async () => {
      wrapper.vm.updateFormState(false);
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.isExistingSecret).toBe(false);
      expect(wrapper.emitted("update-secret-details")).toBeTruthy();
      expect(wrapper.emitted("update-secret-details")?.[0]).toEqual([{}]);
    });

    test("resets selected format when switching modes", async () => {
      wrapper.vm.selectedFormat = AssetKind.AccessKey;
      wrapper.vm.updateFormState(true);
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.selectedFormat).toBe(null);
    });
  });

  describe("Existing Secret Selection", () => {
    test("converts existing secrets to options array", () => {
      const options = wrapper.vm.existingSecretOptions;
      expect(options).toEqual(["secret-1", "secret-2"]);
    });

    test("emits updated secret details when existing secret is selected", async () => {
      wrapper.vm.updateExistingSecret("secret-1");
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([{ existingSecretName: "secret-1" }]);
    });

    test("displays loading state correctly", async () => {
      await wrapper.setProps({ loading: true });
      expect(wrapper.vm.loading).toBe(true);
    });
  });

  describe("New Secret - Format Selection", () => {
    beforeEach(async () => {
      wrapper.vm.isExistingSecret = false;
      await wrapper.vm.$nextTick();
    });

    test("provides correct format options", () => {
      const options = wrapper.vm.supportedFormatOptions;
      expect(options).toHaveLength(4);
      expect(options.map((o) => o.value)).toEqual([
        AssetKind.GenericSecret,
        AssetKind.AccessKey,
        AssetKind.Password,
        TYPE_SSH_KEY,
      ]);
    });

    test("filters out disabled formats", async () => {
      await wrapper.setProps({ disabledFormats: [TYPE_SSH_KEY, AssetKind.GenericSecret] });
      const options = wrapper.vm.supportedFormatOptions;

      expect(options).toHaveLength(2);
      expect(options.map((o) => o.value)).toEqual([AssetKind.AccessKey, AssetKind.Password]);
    });

    test("emits correct structure when GenericSecret format is selected", async () => {
      wrapper.vm.updateSelectedFormat(AssetKind.GenericSecret);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted).toBeTruthy();
      expect(emitted?.[emitted.length - 1]).toEqual([{ keyValuePairs: [{ key: "", value: "" }] }]);
    });

    test("emits correct structure when AccessKey format is selected", async () => {
      wrapper.vm.updateSelectedFormat(AssetKind.AccessKey);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted).toBeTruthy();
      expect(emitted?.[emitted.length - 1]).toEqual([{ accessKeyId: "", secretAccessKey: "" }]);
    });

    test("emits correct structure when Password format is selected", async () => {
      wrapper.vm.updateSelectedFormat(AssetKind.Password);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted).toBeTruthy();
      expect(emitted?.[emitted.length - 1]).toEqual([{ user: "", password: "" }]);
    });

    test("emits correct structure when SSH Key format is selected", async () => {
      wrapper.vm.updateSelectedFormat(TYPE_SSH_KEY);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted).toBeTruthy();
      expect(emitted?.[emitted.length - 1]).toEqual([{ keyValuePairs: [{ key: "ssh", value: "" }] }]);
    });
  });

  describe("Secret Pair Details Updates", () => {
    beforeEach(async () => {
      wrapper.vm.isExistingSecret = false;
      await wrapper.vm.$nextTick();
    });

    test("updates access key details correctly", async () => {
      wrapper.vm.selectedFormat = AssetKind.AccessKey;
      await wrapper.vm.$nextTick();

      wrapper.vm.updateSecretPairDetails({ identifier: "my-access-key", secret: "my-secret" });
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted?.[emitted.length - 1]).toEqual([{ accessKeyId: "my-access-key", secretAccessKey: "my-secret" }]);
    });

    test("updates password details correctly", async () => {
      wrapper.vm.selectedFormat = AssetKind.Password;
      await wrapper.vm.$nextTick();

      wrapper.vm.updateSecretPairDetails({ identifier: "admin", secret: "pass123" });
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted?.[emitted.length - 1]).toEqual([{ user: "admin", password: "pass123" }]);
    });

    test("handles empty identifier for access key", async () => {
      wrapper.vm.selectedFormat = AssetKind.AccessKey;
      await wrapper.vm.$nextTick();

      wrapper.vm.updateSecretPairDetails({ identifier: "", secret: "my-secret" });
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted?.[emitted.length - 1]).toEqual([{ accessKeyId: "", secretAccessKey: "my-secret" }]);
    });
  });

  describe("SSH Key Updates", () => {
    beforeEach(async () => {
      wrapper.vm.isExistingSecret = false;
      wrapper.vm.selectedFormat = TYPE_SSH_KEY;
      await wrapper.vm.$nextTick();
    });

    test("updates SSH key value correctly", async () => {
      const sshKeyValue = "ssh-rsa AAAAB3...";
      await wrapper.setProps({
        secretDetails: { keyValuePairs: [{ key: "ssh", value: "" }] },
      });

      wrapper.vm.updateSshKeyValue(sshKeyValue);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted?.[emitted.length - 1]).toEqual([{ keyValuePairs: [{ key: "ssh", value: sshKeyValue }] }]);
    });

    test("handles null SSH key value", async () => {
      await wrapper.setProps({
        secretDetails: { keyValuePairs: [{ key: "ssh", value: "existing" }] },
      });

      wrapper.vm.updateSshKeyValue(null);
      await wrapper.vm.$nextTick();

      const emitted = wrapper.emitted("update-secret-details");
      expect(emitted?.[emitted.length - 1]).toEqual([{ keyValuePairs: [{ key: "ssh", value: "" }] }]);
    });
  });

  describe("Computed Properties - Secret Pair Details Props", () => {
    test("provides correct props for AccessKey format", async () => {
      wrapper.vm.selectedFormat = AssetKind.AccessKey;
      await wrapper.setProps({
        secretDetails: { accessKeyId: "test-id", secretAccessKey: "test-secret" },
      });
      await wrapper.vm.$nextTick();

      const props = wrapper.vm.secretPairDetailsProps;
      expect(props.identifierInput.value).toBe("test-id");
      expect(props.identifierInput.label).toBe("Access key");
      expect(props.secretInput.value).toBe("test-secret");
      expect(props.secretInput.label).toBe("Access secret");
    });

    test("provides correct props for Password format", async () => {
      wrapper.vm.selectedFormat = AssetKind.Password;
      await wrapper.setProps({
        secretDetails: { user: "admin", password: "pass123" },
      });
      await wrapper.vm.$nextTick();

      const props = wrapper.vm.secretPairDetailsProps;
      expect(props.identifierInput.value).toBe("admin");
      expect(props.identifierInput.label).toBe("Username");
      expect(props.secretInput.value).toBe("pass123");
      expect(props.secretInput.label).toBe("Password");
    });

    test("returns empty values when format is not selected", async () => {
      wrapper.vm.selectedFormat = null;
      await wrapper.vm.$nextTick();

      const props = wrapper.vm.secretPairDetailsProps;
      expect(props.identifierInput.value).toBe("");
      expect(props.identifierInput.label).toBe("");
      expect(props.secretInput.value).toBe("");
      expect(props.secretInput.label).toBe("");
    });
  });

  describe("Summary Display", () => {
    test("displays existing secret name in summary", async () => {
      wrapper.vm.isExistingSecret = true;
      await wrapper.setProps({
        secretDetails: { existingSecretName: "my-secret" },
      });
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.summary).toBe("my-secret");
    });

    test("displays 'None' when no existing secret is selected", async () => {
      wrapper.vm.isExistingSecret = true;
      await wrapper.setProps({
        secretDetails: {},
      });
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.summary).toBe("None");
    });

    test("displays format label for new secret", async () => {
      wrapper.vm.isExistingSecret = false;
      wrapper.vm.selectedFormat = AssetKind.AccessKey;
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.summary).toBe("Access key");
    });

    test("displays 'None' when no format is selected for new secret", async () => {
      wrapper.vm.isExistingSecret = false;
      wrapper.vm.selectedFormat = null;
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.summary).toBe("None");
    });
  });

  describe("Validation Rules", () => {
    test("provides correct validation rules for AccessKey identifier", async () => {
      wrapper.vm.selectedFormat = AssetKind.AccessKey;
      await wrapper.vm.$nextTick();

      const rules = wrapper.vm.identifierInputRules;
      expect(rules).toHaveLength(1);
    });

    test("provides correct validation rules for Password identifier", async () => {
      wrapper.vm.selectedFormat = AssetKind.Password;
      await wrapper.vm.$nextTick();

      const rules = wrapper.vm.identifierInputRules;
      expect(rules).toHaveLength(1);
    });

    test("provides correct validation rules for AccessKey secret", async () => {
      wrapper.vm.selectedFormat = AssetKind.AccessKey;
      await wrapper.vm.$nextTick();

      const rules = wrapper.vm.secretInputRules;
      expect(rules).toHaveLength(1);
    });

    test("provides correct validation rules for Password secret", async () => {
      wrapper.vm.selectedFormat = AssetKind.Password;
      await wrapper.vm.$nextTick();

      const rules = wrapper.vm.secretInputRules;
      expect(rules).toHaveLength(1);
    });
  });
});
