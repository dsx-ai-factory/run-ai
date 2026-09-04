/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import HuggingFaceAccessSection from "./hugging-face-access-section.vue";
import { Quasar } from "quasar";
import {
  EHuggingFaceAccessSectionEvents,
  ETokenProvisionMethods,
  type IHuggingFaceAccessSectionModel,
} from "./hugging-face-access-section.model";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { errorMessages } from "@/common/error-message.constant";
import { MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS } from "@/common/version.constant";
import { Scope } from "@/swagger-models/assets-service-client";

vi.mock("@/services/control-plane/inference.service/inference.service");
vi.mock("@/components/common/runai-credential-selector-by-source", () => ({
  RunaiCredentialSelectorBySource: {
    name: "RunaiCredentialSelectorBySource",
    template:
      "<div class=\"mock-credential-selector\" @click=\"$emit('update:model-value', { source: 'my-credentials', credentialName: 'cred1' })\"></div>",
    props: ["modelValue", "scopeModel"],
  },
}));

const defaultModel: IHuggingFaceAccessSectionModel = {
  modelName: "test-model",
  token: undefined,
  selectedCredentials: undefined,
  credentialsSource: undefined,
  selectedTokenProvisionMethod: ETokenProvisionMethods.Manually,
};

describe("hugging-face-access-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof HuggingFaceAccessSection>>;
  const projectId = 42;

  const mountComponent = (modelValue = defaultModel) =>
    mount(HuggingFaceAccessSection, {
      global: { plugins: [Quasar] },
      props: {
        modelValue,
        clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        scopeModel: {
          scope: Scope.Project,
          projectId: projectId,
        },
        modelName: "test-model",
      },
    });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  function getEmittedPayload(
    wrapper: VueWrapper<InstanceType<typeof HuggingFaceAccessSection>>,
  ): IHuggingFaceAccessSectionModel {
    return (
      wrapper.emitted()[EHuggingFaceAccessSectionEvents.UpdateModelValue][0] as IHuggingFaceAccessSectionModel[]
    )[0];
  }

  it("renders radio options, plaintext warning, and token input when manual provision is selected", () => {
    wrapper = mountComponent();
    expect(wrapper.findComponent({ name: "runai-radio-options" }).exists()).toBe(true);
    expect(wrapper.findComponent({ name: "NotificationAlertBar" }).exists()).toBe(true);
    expect(wrapper.text()).toContain("The token will be visible in the workload spec and via the API.");
    expect(wrapper.findComponent({ name: "policy-string-field" }).exists()).toBe(true);
    expect(wrapper.find(".mock-credential-selector").exists()).toBe(false);
  });

  it("lists the credential-backed option before the manual token option", () => {
    wrapper = mountComponent();
    const radioOptions = wrapper.findComponent({ name: "runai-radio-options" });
    expect(radioOptions.props("options")).toEqual([
      { value: ETokenProvisionMethods.FromSecret, label: "Select credential" },
      { value: ETokenProvisionMethods.Manually, label: "Provide a token" },
    ]);
  });

  it("renders credential selector when provision method is 'from_secret'", async () => {
    const model = { ...defaultModel, selectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret };
    wrapper = mountComponent(model);
    await flushPromises();
    expect(wrapper.find(".mock-credential-selector").exists()).toBe(true);
    expect(wrapper.findComponent({ name: "NotificationAlertBar" }).exists()).toBe(false);
    expect(wrapper.findComponent({ name: "policy-string-field" }).exists()).toBe(false);
  });

  it("emits event when token provision method changes", async () => {
    wrapper = mountComponent();
    await wrapper
      .findComponent({ name: "runai-radio-options" })
      .vm.$emit("update:model-value", ETokenProvisionMethods.FromSecret);

    const emitted = getEmittedPayload(wrapper);
    expect(emitted).toBeDefined();
    expect(emitted.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.FromSecret);
    expect(emitted.token).toBeUndefined();
    expect(emitted.selectedCredentials).toBeUndefined();
  });

  it("emits event when token is updated", async () => {
    wrapper = mountComponent();
    wrapper.vm.onTokenChanged("my-token");
    const emitted = getEmittedPayload(wrapper);
    expect(emitted).toBeDefined();
    expect(emitted.token).toBe("my-token");
    expect(emitted.selectedCredentials).toBeUndefined();
    expect(emitted.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.Manually);
  });

  it("emits event when credential is selected", async () => {
    const model = { ...defaultModel, selectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret };
    wrapper = mountComponent(model);
    await flushPromises();
    await wrapper.find(".mock-credential-selector").trigger("click");
    const emitted = getEmittedPayload(wrapper);
    expect(emitted).toBeDefined();
    expect(emitted.selectedCredentials).toBe("cred1");
    expect(emitted.credentialsSource).toBe("my-credentials");
    expect(emitted.token).toBeUndefined();
    expect(emitted.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.FromSecret);
  });

  it("passes correct props to credential selector", async () => {
    const model = {
      ...defaultModel,
      selectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret,
      selectedCredentials: "cred2",
      credentialsSource: ERunaiCredentialSource.MyCredentials,
    };
    wrapper = mountComponent(model);
    await flushPromises();
    const credentialCmp = wrapper.findComponent({ name: "RunaiCredentialSelectorBySource" });
    expect(credentialCmp.props("modelValue")).toEqual({
      source: ERunaiCredentialSource.MyCredentials,
      credentialName: "cred2",
    });

    expect(credentialCmp.props("scopeModel")).toEqual({
      scope: Scope.Project,
      projectId: projectId,
    });
  });

  it("passes source as undefined", async () => {
    const model = {
      ...defaultModel,
      selectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret,
      selectedCredentials: "cred2",
    };
    wrapper = mountComponent(model);
    await flushPromises();
    const credentialCmp = wrapper.findComponent({ name: "RunaiCredentialSelectorBySource" });
    expect(credentialCmp.props("modelValue")).toEqual({
      source: undefined, // localCredentialsSource is undefined by default
      credentialName: "cred2",
    });
    expect(credentialCmp.props("scopeModel")).toEqual({
      scope: Scope.Project,
      projectId: projectId,
    });
  });

  it("calls invokeTokenValidation in created hook if modelValue.token is set", () => {
    const modelWithToken: IHuggingFaceAccessSectionModel = {
      ...defaultModel,
      token: "sometoken",
    };
    const methods = (HuggingFaceAccessSection as unknown as { methods: Record<string, (...args: unknown[]) => unknown> })
      .methods;
    const invokeTokenValidationSpy = vi.spyOn(methods, "invokeTokenValidation");
    mountComponent(modelWithToken);
    expect(invokeTokenValidationSpy).toHaveBeenCalled();
  });

  describe("validateToken", () => {
    it("calls validateToken rule and handles validation result", async () => {
      wrapper = mountComponent();
      const verifyModelWithTokenSpy = vi.spyOn(wrapper.vm, "verifyModelWithToken");
      await wrapper.vm.validateToken("test-token");
      expect(verifyModelWithTokenSpy).toHaveBeenCalledWith({
        modelName: "test-model",
        hfToken: "test-token",
      });
    });

    it("validateToken returns error message if token is invalid", async () => {
      wrapper = mountComponent();
      vi.spyOn(wrapper.vm, "verifyModelWithToken").mockResolvedValueOnce({
        modelName: "test-model",
        hfToken: "hf_validtoken",
        gated: true,
        errorMessage: "Gain access to the model via Hugging Face",
      });
      const result = await wrapper.vm.validateToken("hf_validtoken");
      expect(result).toBe("Gain access to the model via Hugging Face");
    });

    it("validateToken returns true if token is valid and authorized", async () => {
      wrapper = mountComponent();
      vi.spyOn(wrapper.vm, "verifyModelWithToken").mockResolvedValueOnce({
        modelName: "test-model",
        hfToken: "hf_validtoken",
        gated: true,
        errorMessage: undefined,
      });
      const result = await wrapper.vm.validateToken("hf_validtoken");
      expect(result).toBe(true);
    });
  });

  describe("verifyModelWithToken", () => {
    const baseRequest = { modelName: "test-model", hfToken: "hf_token" };
    afterEach(() => {
      vi.clearAllMocks();
    });

    it("returns ENTER_A_TOKEN if token is empty", async () => {
      vi.mocked(inferenceService.getHuggingFaceModelInfo);
      wrapper = mountComponent();
      const result = await wrapper.vm.verifyModelWithToken({ ...baseRequest, hfToken: "" });
      expect(result.errorMessage).toBe(errorMessages.ENTER_A_TOKEN);
      expect(inferenceService.getHuggingFaceModelInfo).not.toHaveBeenCalled();
    });

    it("returns ENTER_A_VALID_TOKEN if token is invalid", async () => {
      vi.mocked(inferenceService.getHuggingFaceModelInfo);
      wrapper = mountComponent();
      const result = await wrapper.vm.verifyModelWithToken({ ...baseRequest, hfToken: "badtoken" });
      expect(result.errorMessage).toBe(errorMessages.ENTER_A_VALID_TOKEN);
      expect(inferenceService.getHuggingFaceModelInfo).not.toHaveBeenCalled();
    });

    it("returns TOKEN_IS_NOT_AUTHORIZED if backend says not authorized", async () => {
      vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
        modelName: baseRequest.modelName,
        authorized: false,
      });
      wrapper = mountComponent();

      const result = await wrapper.vm.verifyModelWithToken({ ...baseRequest, hfToken: "hf_validtoken" });
      expect(result.errorMessage).toBe(errorMessages.TOKEN_IS_NOT_AUTHORIZED);
      expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
    });

    it("returns no errorMessage if token is valid and authorized", async () => {
      wrapper = mountComponent();
      vi.mocked(inferenceService.getHuggingFaceModelInfo).mockResolvedValue({
        modelName: baseRequest.modelName,
        authorized: true,
      });
      const result = await wrapper.vm.verifyModelWithToken({ ...baseRequest, hfToken: "hf_validtoken" });
      expect(result.errorMessage).toBeUndefined();
      expect(inferenceService.getHuggingFaceModelInfo).toHaveBeenCalled();
    });
  });
});
