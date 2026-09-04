/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import HybridHuggingFaceModelSection from "./hybrid-hugging-face-model-section.vue";
import { describe, it, expect, beforeEach, afterEach, vi, beforeAll, afterAll } from "vitest";
import { Quasar } from "quasar";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { EAccessOptions } from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import {
  LLM_IMAGE_FOR_HUGGING_FACE,
  EHuggingFaceEnvVariableNames,
  ARGS_PREFIX_FOR_TGI,
  ARGS_PREFIX_FOR_VLLM,
} from "@/models/inference.model";
import { ETokenProvisionMethods } from "./hugging-face-access-section/hugging-face-access-section.model";
import { ServingPortAccessAuthorizationTypeEnum } from "@/swagger-models/policy-service-client";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import type { IHuggingFaceModelAndAccessModel } from "./hugging-face-model-and-access/hugging-face-model-and-access.model";
import type { IInferenceFormUIState } from "@/stores/inference.store";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS } from "@/common/version.constant";

import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { useClusterStore } from "@/stores/cluster.store";
import { Scope } from "@/swagger-models/assets-service-client";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const baseProps = {
  scopeModel: {
    scope: Scope.Project,
    projectId: 1,
  },
  workloadSpec: {},
};

const HuggingFaceModelAndAccessStub = {
  name: "HuggingFaceModelAndAccess",
  template: '<div data-testid="hugging-face-model-and-access-stub"></div>',
};

const HuggingFaceAccessSectionStub = {
  name: "HuggingFaceAccessSection",
  template: '<div data-testid="hugging-face-access-section-stub"></div>',
};

const ServingEndpointSectionStub = {
  name: "ServingEndpointSection",
  template: '<div data-testid="serving-endpoint-section-stub"></div>',
};

const HuggingFaceModelSelectorStub = {
  name: "HuggingFaceModelSelector",
  template: '<div data-testid="hugging-face-model-selector-stub"></div>',
};

function createWrapper(workloadSpec = {}, uiState = {}): VueWrapper<InstanceType<typeof HybridHuggingFaceModelSection>> {
  return mount(HybridHuggingFaceModelSection, {
    global: {
      plugins: [Quasar],
      components: {
        HuggingFaceModelAndAccess: HuggingFaceModelAndAccessStub,
        HuggingFaceAccessSection: HuggingFaceAccessSectionStub,
        ServingEndpointSection: ServingEndpointSectionStub,
        HuggingFaceModelSelector: HuggingFaceModelSelectorStub,
      },
    },
    props: {
      ...baseProps,
      workloadSpec,
      uiState,
      clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
    },
  });
}

describe("hybrid-hugging-face-model-section.vue", () => {
  beforeAll(() => {
    const clusterStore = useClusterStore(pinia);
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
  });

  afterAll(() => {
    vi.clearAllMocks();
  });

  describe("computed", () => {
    describe("summary", () => {
      it("returns model name when args contains --model", () => {
        const wrapper = createWrapper({ args: `${ARGS_PREFIX_FOR_VLLM} my-model` });
        expect(wrapper.vm.summary).toBe("my-model");
      });

      it("returns model name when args contains --model-id", () => {
        const wrapper = createWrapper({ args: `${ARGS_PREFIX_FOR_TGI} my-model-id` });
        expect(wrapper.vm.summary).toBe("my-model-id");
      });

      it("returns None if args does not contain --model or --model-id", () => {
        const wrapper = createWrapper({ args: "--other-flag value" });
        expect(wrapper.vm.summary).toBe("None");
      });

      it("returns None if args is undefined", () => {
        const wrapper = createWrapper({});
        expect(wrapper.vm.summary).toBe("None");
      });
    });

    describe("hfSpec", () => {
      it("returns modelName from args when args contains --model", () => {
        const wrapper = createWrapper({ args: `${ARGS_PREFIX_FOR_VLLM} my-model` });
        expect(wrapper.vm.hfSpec.modelName).toBe("my-model");
      });

      it("defaults to credential-backed token provision when no HF_TOKEN or UI selection exists", () => {
        const wrapper = createWrapper({});
        expect(wrapper.vm.hfSpec.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.FromSecret);
      });

      it("returns token and Manually when HF_TOKEN has value", () => {
        const wrapper = createWrapper({
          environmentVariables: [{ name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token-123" }],
        });
        expect(wrapper.vm.hfSpec.token).toBe("token-123");
        expect(wrapper.vm.hfSpec.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.Manually);
      });

      it("returns selectedCredentials and FromSecret when HF_TOKEN has secret", () => {
        const wrapper = createWrapper({
          environmentVariables: [
            { name: EHuggingFaceEnvVariableNames.HF_TOKEN, secret: { name: "secret-abc", key: "key" } },
          ],
        });
        expect(wrapper.vm.hfSpec.selectedCredentials).toBe("secret-abc");
        expect(wrapper.vm.hfSpec.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.FromSecret);
      });

      it("infers SharedSecret from HF_TOKEN secret when uiState has no credentialsSource", () => {
        const wrapper = createWrapper({
          environmentVariables: [
            { name: EHuggingFaceEnvVariableNames.HF_TOKEN, secret: { name: "secret-abc", key: "HF_TOKEN" } },
          ],
        });
        expect(wrapper.vm.hfSpec.credentialsSource).toBe(ERunaiCredentialSource.SharedSecret);
      });

      it("returns selectedCredentials and FromSecret when HF_TOKEN has userCredential", () => {
        const wrapper = createWrapper({
          environmentVariables: [
            {
              name: EHuggingFaceEnvVariableNames.HF_TOKEN,
              userCredential: { name: "my-uc", key: "HF_TOKEN" },
            },
          ],
        });
        expect(wrapper.vm.hfSpec.selectedCredentials).toBe("my-uc");
        expect(wrapper.vm.hfSpec.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.FromSecret);
      });

      it("infers MyCredentials from HF_TOKEN userCredential when uiState has no credentialsSource", () => {
        const wrapper = createWrapper({
          environmentVariables: [
            {
              name: EHuggingFaceEnvVariableNames.HF_TOKEN,
              userCredential: { name: "my-uc", key: "HF_TOKEN" },
            },
          ],
        });
        expect(wrapper.vm.hfSpec.credentialsSource).toBe(ERunaiCredentialSource.MyCredentials);
      });

      it("prefers uiState credentialsSource over inference from env", () => {
        const wrapper = createWrapper(
          {
            environmentVariables: [
              {
                name: EHuggingFaceEnvVariableNames.HF_TOKEN,
                userCredential: { name: "my-uc", key: "HF_TOKEN" },
              },
            ],
          },
          { credentialsSource: ERunaiCredentialSource.SharedSecret },
        );
        expect(wrapper.vm.hfSpec.credentialsSource).toBe(ERunaiCredentialSource.SharedSecret);
      });

      it("returns credentialsSource and accessFromStorage from uiState", () => {
        const wrapper = createWrapper(
          {},
          {
            credentialsSource: ERunaiCredentialSource.MyCredentials,
            accessFromStorage: true,
          },
        );
        expect(wrapper.vm.hfSpec.credentialsSource).toBe(ERunaiCredentialSource.MyCredentials);
        expect(wrapper.vm.hfSpec.accessFromStorage).toBe(true);
      });

      it("defaults accessFromStorage to false when not in uiState", () => {
        const wrapper = createWrapper({});
        expect(wrapper.vm.hfSpec.accessFromStorage).toBe(false);
      });

      it("uses hfSelectedTokenProvisionMethod from uiState when no HF_TOKEN in env", () => {
        const wrapper = createWrapper({}, { hfSelectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret });
        expect(wrapper.vm.hfSpec.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.FromSecret);
      });

      it("preserves a manual token selection from uiState when no HF_TOKEN is set yet", () => {
        const wrapper = createWrapper({}, { hfSelectedTokenProvisionMethod: ETokenProvisionMethods.Manually });
        expect(wrapper.vm.hfSpec.selectedTokenProvisionMethod).toBe(ETokenProvisionMethods.Manually);
      });
    });

    describe("servingPortModel", () => {
      afterEach(() => {
        vi.clearAllMocks();
      });

      it("returns correct model when servingPort is fully defined", () => {
        const wrapper = createWrapper({
          servingPort: { protocol: "http", container: 1234 },
        });
        expect(wrapper.vm.servingPortModel.protocol).toBe("http");
        expect(wrapper.vm.servingPortModel.port).toBe(1234);
        expect(wrapper.vm.servingPortModel.access).toBeDefined();
      });

      it("returns correct model when servingPort is partially defined (missing container)", () => {
        const wrapper = createWrapper({
          servingPort: { protocol: "grpc" },
        });
        expect(wrapper.vm.servingPortModel.protocol).toBe("grpc");
        expect(wrapper.vm.servingPortModel.port).toBeUndefined();
      });

      it("returns default values when servingPort is undefined", () => {
        const wrapper = createWrapper({});
        expect(wrapper.vm.servingPortModel.protocol).toBe("");
        expect(wrapper.vm.servingPortModel.port).toBeUndefined();
        expect(wrapper.vm.servingPortModel.access).toBeDefined();
      });
    });
  });

  describe("methods", () => {
    describe("onServingPortChanged", () => {
      const mockAuth = { authorizationType: ServingPortAccessAuthorizationTypeEnum.Public };

      it("emits update event with correct servingPort when protocol and port are provided", () => {
        const servingPortAccessAuthorizationSpied = vi
          .spyOn(inferenceWorkloadUtil, "servingPortAccessAuthorization")
          .mockReturnValue(mockAuth);
        const wrapper = createWrapper({ servingPort: { protocol: "http", container: 8080 } });
        const servingPortAccess = { protocol: "grpc", port: 1234, access: { accessOption: EAccessOptions.PUBLIC } };
        wrapper.vm.onServingPortChanged(servingPortAccess);
        const emitted = wrapper.emitted().update as Array<[IUISuperWorkloadSpec]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].servingPort).toEqual({
          protocol: "grpc",
          container: 1234,
          ...mockAuth,
        });

        servingPortAccessAuthorizationSpied.mockRestore();
      });

      it("emits update event with undefined port when port is missing", () => {
        const servingPortAccessAuthorizationSpied = vi
          .spyOn(inferenceWorkloadUtil, "servingPortAccessAuthorization")
          .mockReturnValue(mockAuth);
        const wrapper = createWrapper({ servingPort: { protocol: "http", container: 8080 } });
        const servingPortAccess = { protocol: "grpc", access: { accessOption: EAccessOptions.PUBLIC } };
        wrapper.vm.onServingPortChanged(servingPortAccess);
        const emitted = wrapper.emitted().update as Array<[IUISuperWorkloadSpec]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].servingPort).toEqual({
          protocol: "grpc",
          container: undefined,
          ...mockAuth,
        });

        servingPortAccessAuthorizationSpied.mockRestore();
      });

      it("emits update event with correct servingPort when access is undefined", () => {
        const servingPortAccessAuthorizationSpied = vi
          .spyOn(inferenceWorkloadUtil, "servingPortAccessAuthorization")
          .mockReturnValue(mockAuth);
        const wrapper = createWrapper({ servingPort: { protocol: "http", container: 8080 } });
        const servingPortAccess = { protocol: "grpc", port: 1234 };
        wrapper.vm.onServingPortChanged(servingPortAccess);
        const emitted = wrapper.emitted().update as Array<[IUISuperWorkloadSpec]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].servingPort).toEqual({
          protocol: "grpc",
          container: 1234,
          ...mockAuth,
        });

        servingPortAccessAuthorizationSpied.mockRestore();
      });

      it("emits update event with correct servingPort when accessOption is GROUPS and authorizedGroups are provided", () => {
        const mockAuthGroups = {
          authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
          authorizedGroups: ["group1", "group2"],
        };
        const servingPortAccessAuthorizationSpied = vi
          .spyOn(inferenceWorkloadUtil, "servingPortAccessAuthorization")
          .mockReturnValue(mockAuthGroups);
        const wrapper = createWrapper({ servingPort: { protocol: "http", container: 8080 } });
        const servingPortAccess = {
          port: 1234,
          protocol: "grpc",
          access: { accessOption: EAccessOptions.GROUPS, authorizedGroups: ["group1", "group2"] },
        };
        wrapper.vm.onServingPortChanged(servingPortAccess);
        const emitted = wrapper.emitted().update as Array<[IUISuperWorkloadSpec]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].servingPort).toEqual({
          protocol: "grpc",
          container: 1234,
          ...mockAuthGroups,
        });

        servingPortAccessAuthorizationSpied.mockRestore();
      });
    });

    describe("applyHfAccessToEnv", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("excludes HF_TOKEN when neither token nor selectedCredentials is set", () => {
        const excludeMock = vi
          .spyOn(inferenceWorkloadUtil, "excludeEnvironmentVariablesByName")
          .mockReturnValue([{ name: "OTHER_ENV", value: "something" }]);
        const wrapper = createWrapper({
          environmentVariables: [
            { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token-123" },
            { name: "OTHER_ENV", value: "something" },
          ],
        });
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "test-model",
          token: undefined,
          selectedCredentials: undefined,
          selectedTokenProvisionMethod: ETokenProvisionMethods.Manually,
        };
        const envVars = wrapper.props("workloadSpec").environmentVariables ?? [];
        const result = wrapper.vm.applyHfAccessToEnv(envVars, payload);
        expect(excludeMock).toHaveBeenCalledWith(EHuggingFaceEnvVariableNames.HF_TOKEN, expect.any(Array));
        expect(result).toEqual([{ name: "OTHER_ENV", value: "something" }]);
      });

      it("replaces/adds HF_TOKEN with value when provision method is Manually and token is set", () => {
        const replaceMock = vi.spyOn(inferenceWorkloadUtil, "replaceEnvironmentVariable").mockReturnValue([
          { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token-abc" },
          { name: "OTHER_ENV", value: "something" },
        ]);
        const wrapper = createWrapper({
          environmentVariables: [{ name: "OTHER_ENV", value: "something" }],
        });
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "test-model",
          token: "token-abc",
          selectedCredentials: undefined,
          selectedTokenProvisionMethod: ETokenProvisionMethods.Manually,
        };
        const envVars = wrapper.props("workloadSpec").environmentVariables ?? [];
        const result = wrapper.vm.applyHfAccessToEnv(envVars, payload);
        expect(replaceMock).toHaveBeenCalledWith(expect.any(Array), EHuggingFaceEnvVariableNames.HF_TOKEN, "token-abc");
        expect(result).toEqual([
          { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token-abc" },
          { name: "OTHER_ENV", value: "something" },
        ]);
      });

      it("replaces/adds HF_TOKEN with secret when provision method is FromSecret and selectedCredentials is set", () => {
        const replaceMock = vi.spyOn(inferenceWorkloadUtil, "replaceEnvironmentVariable").mockReturnValue([
          {
            name: EHuggingFaceEnvVariableNames.HF_TOKEN,
            secret: { name: "my-secret", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
          },
          { name: "OTHER_ENV", value: "something" },
        ]);
        const wrapper = createWrapper({
          environmentVariables: [{ name: "OTHER_ENV", value: "something" }],
        });
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "test-model",
          token: undefined,
          selectedCredentials: "my-secret",
          selectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret,
        };
        const envVars = wrapper.props("workloadSpec").environmentVariables ?? [];
        const result = wrapper.vm.applyHfAccessToEnv(envVars, payload);
        expect(replaceMock).toHaveBeenCalledWith(
          expect.any(Array),
          EHuggingFaceEnvVariableNames.HF_TOKEN,
          undefined,
          { name: "my-secret", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
          undefined,
        );
        expect(result).toEqual([
          {
            name: EHuggingFaceEnvVariableNames.HF_TOKEN,
            secret: { name: "my-secret", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
          },
          { name: "OTHER_ENV", value: "something" },
        ]);
      });

      it("replaces HF_TOKEN with userCredential when provision method is FromSecret and credentialsSource is MyCredentials", () => {
        const replaceMock = vi.spyOn(inferenceWorkloadUtil, "replaceEnvironmentVariable").mockReturnValue([
          {
            name: EHuggingFaceEnvVariableNames.HF_TOKEN,
            userCredential: { name: "uc-name", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
          },
          { name: "OTHER_ENV", value: "something" },
        ]);
        const wrapper = createWrapper({
          environmentVariables: [{ name: "OTHER_ENV", value: "something" }],
        });
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "test-model",
          token: undefined,
          selectedCredentials: "uc-name",
          selectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret,
          credentialsSource: ERunaiCredentialSource.MyCredentials,
        };
        const envVars = wrapper.props("workloadSpec").environmentVariables ?? [];
        const result = wrapper.vm.applyHfAccessToEnv(envVars, payload);
        expect(replaceMock).toHaveBeenCalledWith(
          expect.any(Array),
          EHuggingFaceEnvVariableNames.HF_TOKEN,
          undefined,
          undefined,
          { name: "uc-name", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
        );
        expect(result).toEqual([
          {
            name: EHuggingFaceEnvVariableNames.HF_TOKEN,
            userCredential: { name: "uc-name", key: EHuggingFaceEnvVariableNames.HF_TOKEN },
          },
          { name: "OTHER_ENV", value: "something" },
        ]);
      });
    });

    describe("onHfUpdate", () => {
      beforeEach(() => {
        vi.spyOn(inferenceWorkloadUtil, "getHFArgsForModel").mockReturnValue(`${ARGS_PREFIX_FOR_VLLM} test-model`);
      });

      afterEach(() => {
        vi.clearAllMocks();
      });

      it("emits update with args and environmentVariables from payload", () => {
        const replaceMock = vi.spyOn(inferenceWorkloadUtil, "replaceEnvironmentVariable").mockReturnValue([
          { name: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "token-abc" },
          { name: "OTHER_ENV", value: "something" },
        ]);
        const wrapper = createWrapper({
          image: LLM_IMAGE_FOR_HUGGING_FACE,
          environmentVariables: [{ name: "OTHER_ENV", value: "something" }],
        });
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "test-model",
          token: "token-abc",
          selectedCredentials: undefined,
          selectedTokenProvisionMethod: ETokenProvisionMethods.Manually,
          credentialsSource: undefined,
        };
        wrapper.vm.onHfUpdate(payload);
        const emitted = wrapper.emitted().update as Array<[IUISuperWorkloadSpec]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].args).toBe(`${ARGS_PREFIX_FOR_VLLM} test-model`);
        expect(replaceMock).toHaveBeenCalledWith(expect.any(Array), EHuggingFaceEnvVariableNames.HF_TOKEN, "token-abc");
      });

      it("emits update:ui-state when payload has credentialsSource", () => {
        vi.spyOn(inferenceWorkloadUtil, "replaceEnvironmentVariable").mockReturnValue([]);
        const wrapper = createWrapper({}, { credentialsSource: ERunaiCredentialSource.SharedSecret });
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "test-model",
          token: undefined,
          selectedCredentials: undefined,
          selectedTokenProvisionMethod: ETokenProvisionMethods.FromSecret,
          credentialsSource: ERunaiCredentialSource.MyCredentials,
        };
        wrapper.vm.onHfUpdate(payload);
        const emitted = wrapper.emitted("update:ui-state") as Array<[IInferenceFormUIState]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].credentialsSource).toBe(ERunaiCredentialSource.MyCredentials);
      });

      it("emits update:ui-state when payload has accessFromStorage", () => {
        vi.spyOn(inferenceWorkloadUtil, "excludeEnvironmentVariablesByName").mockReturnValue([]);
        const wrapper = createWrapper({});
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "test-model",
          token: undefined,
          selectedCredentials: undefined,
          selectedTokenProvisionMethod: ETokenProvisionMethods.Manually,
          credentialsSource: undefined,
          accessFromStorage: true,
        };
        wrapper.vm.onHfUpdate(payload);
        const emitted = wrapper.emitted("update:ui-state") as Array<[IInferenceFormUIState]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].accessFromStorage).toBe(true);
      });

      it("emits update:ui-state with hfSelectedTokenProvisionMethod when payload has selectedTokenProvisionMethod", () => {
        vi.spyOn(inferenceWorkloadUtil, "excludeEnvironmentVariablesByName").mockReturnValue([]);
        const wrapper = createWrapper({});
        const payload: IHuggingFaceModelAndAccessModel = {
          modelName: "new-model",
          token: undefined,
          selectedCredentials: undefined,
          selectedTokenProvisionMethod: ETokenProvisionMethods.Manually,
          credentialsSource: undefined,
        };
        wrapper.vm.onHfUpdate(payload);
        const emitted = wrapper.emitted("update:ui-state") as Array<[IInferenceFormUIState]>;
        expect(emitted).toBeTruthy();
        expect(emitted[0][0].hfSelectedTokenProvisionMethod).toBe(ETokenProvisionMethods.Manually);
      });
    });
  });
});
