import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { QInput, QSelect, Quasar } from "quasar";
import { EnvironmentVariableBox } from ".";
import type { EnvironmentVariableOfAsset } from "@/swagger-models/assets-service-client";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { describe, expect, it } from "vitest";

enum Labels {
  Name = "Name",
  Value = "Value",
  Source = "Source",
  SecretName = "Credential name",
  SecretKey = "Secret key",
}

describe("environment-variable-box.vue", () => {
  const creds = [
    { label: "gc", value: "generic-secret", type: AssetKind.GenericSecret },
    { label: "ac", value: "access-key", type: AssetKind.AccessKey },
    { label: "up", value: "user-pass", type: AssetKind.Password },
  ];
  const customEnvironmentVariable = { name: "name", value: "value" };
  const credentialEnvironmentVariable = { name: "name", credential: { assetId: creds[0].value, key: "key" } };

  function factory(props = {}): VueWrapper<InstanceType<typeof EnvironmentVariableBox>> {
    return mount(EnvironmentVariableBox, {
      props: {
        environmentVariable: customEnvironmentVariable,
        sourceValueOptions: creds,
        ...props,
      },
      global: {
        plugins: [Quasar],
        components: { QInput, QSelect },
      },
    });
  }

  function findQInputByLabel(
    wrapper: VueWrapper<InstanceType<typeof EnvironmentVariableBox>>,
    label: string,
  ): VueWrapper<InstanceType<typeof QInput>> | undefined {
    return wrapper.findAllComponents(QInput).find((cmp) => cmp.props("label") === label);
  }

  function findQSelectByLabel(
    wrapper: VueWrapper<InstanceType<typeof EnvironmentVariableBox>>,
    label: string,
  ): VueWrapper<InstanceType<typeof QSelect>> | undefined {
    return wrapper.findAllComponents(QSelect).find((cmp) => cmp.props("label") === label);
  }

  describe("Html elements check", () => {
    it("Component exists", () => {
      const wrapper = factory();
      expect(wrapper.exists()).toBe(true);
    });

    it("Check if the name input exists", () => {
      const wrapper = factory();
      expect(wrapper.findComponent(QInput).exists()).toBe(true);
      expect(wrapper.findComponent(QInput).props("label")).toEqual(Labels.Name);
    });

    it("Check if the source select exists", () => {
      const wrapper = factory();
      expect(wrapper.findComponent(QSelect).exists()).toBe(true);
      expect(wrapper.findComponent(QSelect).props("label")).toEqual(Labels.Source);
    });

    it("Check if the X button exists", () => {
      const wrapper = factory();
      expect(wrapper.find("[aid=remove-environment-variable-button]").exists()).toBe(true);
    });
  });

  describe("Props", () => {
    describe("environmentVariable", () => {
      it("custom environment variable should show the value input", () => {
        const wrapper = factory();
        const valueInput = findQInputByLabel(wrapper, Labels.Value);
        expect(valueInput).toBeTruthy();
      });

      it("custom environment variable should show the source select with 'Custom' option selected", () => {
        const wrapper = factory();
        const sourceSelect = findQSelectByLabel(wrapper, Labels.Source);
        expect(sourceSelect?.props("modelValue")).toEqual("custom");
      });

      it("credential environment variable should show the source select with 'Credentials' option selected", () => {
        const wrapper = factory({ environmentVariable: credentialEnvironmentVariable });
        const sourceSelect = findQSelectByLabel(wrapper, Labels.Source);
        expect(sourceSelect).toBeTruthy();
        expect(sourceSelect?.props("modelValue")).toEqual("credentials");
      });

      it("credential environment variable should show the credential select", () => {
        const wrapper = factory({ environmentVariable: credentialEnvironmentVariable });
        const secretNameSelect = findQSelectByLabel(wrapper, Labels.SecretName);
        expect(secretNameSelect).toBeTruthy();
        expect(secretNameSelect?.props("modelValue")).toEqual(credentialEnvironmentVariable.credential.assetId);
      });

      it("credential environment variable should show the key INPUT element if its type is generic-secret", () => {
        const wrapper = factory({ environmentVariable: credentialEnvironmentVariable });
        const secretKeyInput = findQInputByLabel(wrapper, Labels.SecretKey);
        expect(secretKeyInput).toBeTruthy();
        expect(secretKeyInput?.props("modelValue")).toEqual(credentialEnvironmentVariable.credential.key);
      });

      it("credential environment variable should show the key SELECT element if its type is access-key", () => {
        const credentialEnvironmentVariableWithAccessKey = {
          ...credentialEnvironmentVariable,
          credential: { ...credentialEnvironmentVariable.credential, assetId: creds[1].value },
        };
        const wrapper = factory({ environmentVariable: credentialEnvironmentVariableWithAccessKey });
        const secretKeyInput = findQSelectByLabel(wrapper, Labels.SecretKey);
        expect(secretKeyInput).toBeTruthy();
        expect(secretKeyInput?.props("modelValue")).toEqual(credentialEnvironmentVariableWithAccessKey.credential.key);
      });

      it("credential environment variable should show the key SELECT element if its type is userpass", () => {
        const credentialEnvironmentVariableWithUserpass = {
          ...credentialEnvironmentVariable,
          credential: { ...credentialEnvironmentVariable.credential, assetId: creds[2].value },
        };
        const wrapper = factory({ environmentVariable: credentialEnvironmentVariableWithUserpass });
        const secretKeyInput = findQSelectByLabel(wrapper, Labels.SecretKey);
        expect(secretKeyInput).toBeTruthy();
        expect(secretKeyInput?.props("modelValue")).toEqual(credentialEnvironmentVariableWithUserpass.credential.key);
      });
    });
  });

  describe("Emits", () => {
    it("should emit 'remove-environment-variable' event when the X button is clicked", async () => {
      const wrapper = factory();
      await wrapper.find("[aid=remove-environment-variable-button]").trigger("click");
      expect(wrapper.emitted("remove-environment-variable")).toBeTruthy();
    });

    it("should emit 'update-environment-variable' event when the name input changes", async () => {
      const wrapper = factory();
      const nameInput = findQInputByLabel(wrapper, Labels.Name);
      await nameInput?.setValue("newName");
      const expectedEmit: EnvironmentVariableOfAsset = { ...customEnvironmentVariable, name: "newName" };
      expect(wrapper.emitted("update-environment-variable")?.[0][0]).toEqual(expectedEmit);
    });

    it("should emit 'update-environment-variable' event when the value input changes", async () => {
      const wrapper = factory();
      const valueInput = findQInputByLabel(wrapper, Labels.Value);
      await valueInput?.setValue("newValue");

      const expectedEmit: EnvironmentVariableOfAsset = { ...customEnvironmentVariable, value: "newValue" };
      expect(wrapper.emitted("update-environment-variable")?.[0][0]).toEqual(expectedEmit);
    });

    it("should emit 'update-environment-variable' event when the source select changes", async () => {
      const wrapper = factory();
      const sourceSelect = findQSelectByLabel(wrapper, Labels.Source);
      await sourceSelect?.setValue("credentials");
      const expectedEmit: EnvironmentVariableOfAsset = {
        ...credentialEnvironmentVariable,
        credential: { assetId: null, key: "" },
      };
      expect(wrapper.emitted("update-environment-variable")?.[0][0]).toEqual(expectedEmit);
    });

    it("should emit 'update-environment-variable' event when the secret name select changes", async () => {
      const wrapper = factory({ environmentVariable: credentialEnvironmentVariable });
      const secretNameCmp = findQSelectByLabel(wrapper, Labels.SecretName);
      await secretNameCmp?.setValue("secret2");
      const expectedEmit: EnvironmentVariableOfAsset = {
        ...credentialEnvironmentVariable,
        credential: { key: null, assetId: "secret2" },
      };
      expect(wrapper.emitted("update-environment-variable")?.[0][0]).toEqual(expectedEmit);
    });

    it("should emit 'update-environment-variable' event when the secret key INPUT changes", async () => {
      const wrapper = factory({ environmentVariable: credentialEnvironmentVariable });
      const secretKeyInput = findQInputByLabel(wrapper, Labels.SecretKey);
      await secretKeyInput?.setValue("newKey");
      const expectedEmit: EnvironmentVariableOfAsset = {
        ...credentialEnvironmentVariable,
        credential: { ...credentialEnvironmentVariable.credential, key: "newKey" },
      };
      expect(wrapper.emitted("update-environment-variable")?.[0][0]).toEqual(expectedEmit);
    });

    it("should emit 'update-environment-variable' event when the secret key SELECT changes", async () => {
      const credentialEnvironmentVariableWithUserpass = {
        ...credentialEnvironmentVariable,
        credential: { ...credentialEnvironmentVariable.credential, assetId: creds[2].value },
      };
      const wrapper = factory({ environmentVariable: credentialEnvironmentVariableWithUserpass });
      const secretKeyInput = findQSelectByLabel(wrapper, Labels.SecretKey);
      await secretKeyInput?.setValue("newKey");
      const expectedEmit: EnvironmentVariableOfAsset = {
        ...credentialEnvironmentVariableWithUserpass,
        credential: { ...credentialEnvironmentVariableWithUserpass.credential, key: "newKey" },
      };
      expect(wrapper.emitted("update-environment-variable")?.[0][0]).toEqual(expectedEmit);
    });
  });
});
