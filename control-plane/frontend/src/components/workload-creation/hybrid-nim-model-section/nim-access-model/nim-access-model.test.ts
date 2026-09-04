/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import NimAccessModel from "./nim-access-model.vue";
import { ENimAccessModelMethod, ENimAccessModelEvents } from "./nim-access-model.model";
import {
  ERunaiCredentialSource,
  ERunaiCredentialType,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { Quasar } from "quasar";
import type { INimAccessModelModel } from "./nim-access-model.model";
import { MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS } from "@/common/version.constant";
import { errorMessages } from "@/common/error-message.constant";
import { Scope } from "@/swagger-models/assets-service-client";
import { setActivePinia, createPinia } from "pinia";
import { useClusterStore } from "@/stores/cluster.store";

const pinia = createPinia();
setActivePinia(pinia);

vi.mock("@/services/control-plane/credential.service/credential.service");

vi.mock("@/components/common/runai-credential-selector-by-source", () => ({
  RunaiCredentialSelectorBySource: {
    name: "RunaiCredentialSelectorBySource",
    template: '<div class="mock-credential-selector"></div>',
  },
}));

const MockDockerRegistrySelector = {
  name: "docker-registry-selector",
  template: "<div class='mock-docker-registry-selector'></div>",
  props: {
    modelValue: {
      type: Object,
      required: true,
    },
    clusterId: {
      type: String,
      required: true,
    },
    scopeModel: {
      type: Object,
      required: true,
    },
  },
};

describe("nim-access-model.vue", () => {
  const clusterStore = useClusterStore(pinia);

  beforeEach(() => {
    vi.clearAllMocks();
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
  });
  const baseModelValue: INimAccessModelModel = {
    apiKey: "",
    selectedCredentials: undefined,
    credentialsSource: undefined,
    accessFromStorage: false,
  };

  const projectId = 1;

  const createWrapper = (modelValue = baseModelValue) =>
    mount(NimAccessModel, {
      global: {
        plugins: [Quasar],
        stubs: {
          "docker-registry-selector": MockDockerRegistrySelector,
        },
      },
      props: {
        modelValue: { ...modelValue },
        clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        scopeModel: {
          scope: Scope.Project,
          projectId,
        },
      },
    });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("emits update when onApiKeyChanged is called", () => {
    const wrapper = createWrapper();
    wrapper.vm.onApiKeyChanged("test-key");
    const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
    expect(emitted).toBeTruthy();
    expect(emitted[0]).toEqual([
      {
        apiKey: "test-key",
        selectedCredentials: undefined,
        credentialsSource: undefined,
        accessFromStorage: false,
      },
    ]);
  });

  it("onCredentialChanged emits correct model with shared secret", () => {
    const wrapper = createWrapper();
    const credential = {
      credentialName: "cred",
      source: ERunaiCredentialSource.SharedSecret,
      type: ERunaiCredentialType.GenericSecret,
    };
    wrapper.vm.onCredentialChanged(credential);
    const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
    expect(emitted).toBeTruthy();
    expect(emitted[0]).toEqual([
      {
        apiKey: "",
        selectedCredentials: "cred",
        credentialsSource: ERunaiCredentialSource.SharedSecret,
        credentialType: ERunaiCredentialType.GenericSecret,
        dockerRegistry: {
          dockerRegistryName: "",
        },
        accessFromStorage: false,
      },
    ]);

    expect(wrapper.vm.localSelectedCredentials).toBe("cred");
    expect(wrapper.vm.localCredentialsSource).toBe(ERunaiCredentialSource.SharedSecret);
  });

  it("onCredentialChanged emits correct model with NGC API key from My Credentials", () => {
    const wrapper = createWrapper();
    const credential = {
      credentialName: "my-ngc-cred",
      source: ERunaiCredentialSource.MyCredentials,
      type: ERunaiCredentialType.NgcApiKey,
    };
    wrapper.vm.onCredentialChanged(credential);
    const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
    expect(emitted).toBeTruthy();
    expect(emitted[0]).toEqual([
      {
        apiKey: "",
        selectedCredentials: "my-ngc-cred",
        credentialsSource: ERunaiCredentialSource.MyCredentials,
        credentialType: ERunaiCredentialType.NgcApiKey,
        dockerRegistry: {
          dockerRegistryName: "my-ngc-cred",
        },
        accessFromStorage: false,
      },
    ]);

    expect(wrapper.vm.localSelectedCredentials).toBe("my-ngc-cred");
    expect(wrapper.vm.localCredentialsSource).toBe(ERunaiCredentialSource.MyCredentials);
  });

  it("onCredentialChanged sets docker registry name to empty string when not NGC API key", () => {
    const wrapper = createWrapper();
    const credential = {
      credentialName: "generic-cred",
      source: ERunaiCredentialSource.SharedSecret,
      type: ERunaiCredentialType.GenericSecret,
    };
    wrapper.vm.onCredentialChanged(credential);
    const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
    expect(emitted).toBeTruthy();
    expect(emitted[0]).toEqual([
      {
        apiKey: "",
        selectedCredentials: "generic-cred",
        credentialsSource: ERunaiCredentialSource.SharedSecret,
        credentialType: ERunaiCredentialType.GenericSecret,
        dockerRegistry: {
          dockerRegistryName: "",
        },
        accessFromStorage: false,
      },
    ]);
  });

  describe("onAccessMethodChanged", () => {
    it("emits update when switching to manual mode", () => {
      const initialModelValue: INimAccessModelModel = {
        apiKey: "",
        selectedCredentials: "existing-cred",
        credentialsSource: ERunaiCredentialSource.SharedSecret,
        accessFromStorage: false,
      };
      const wrapper = createWrapper(initialModelValue);

      wrapper.vm.onAccessMethodChanged(ENimAccessModelMethod.Manually);

      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      expect(emitted[0]).toEqual([
        {
          apiKey: "",
          selectedCredentials: undefined,
          credentialsSource: undefined,
          accessFromStorage: false,
        },
      ]);

      // Verify local state is preserved
      expect(wrapper.vm.localSelectedCredentials).toBe("existing-cred");
      expect(wrapper.vm.localCredentialsSource).toBe(ERunaiCredentialSource.SharedSecret);
    });

    it("emits update when switching to credential selection mode", () => {
      const initialModelValue: INimAccessModelModel = {
        apiKey: "existing-key",
        selectedCredentials: undefined,
        credentialsSource: undefined,
        accessFromStorage: false,
      };
      const wrapper = createWrapper(initialModelValue);

      // Set some previous local state
      wrapper.vm.localSelectedCredentials = "previous-cred";
      wrapper.vm.localCredentialsSource = ERunaiCredentialSource.SharedSecret;

      wrapper.vm.onAccessMethodChanged(ENimAccessModelMethod.FromSecret);

      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      expect(emitted[0]).toEqual([
        {
          apiKey: undefined,
          selectedCredentials: "previous-cred",
          credentialsSource: ERunaiCredentialSource.SharedSecret,
          accessFromStorage: false,
        },
      ]);

      // Verify API key is preserved in local state
      expect(wrapper.vm.localApiKey).toBe("existing-key");
    });
  });

  describe("onAccessFromStorageChanged", () => {
    it("emits correct model when enabling access from storage with manual method", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        apiKey: "test-key",
        accessFromStorage: false,
      });

      wrapper.vm.onAccessFromStorageChanged(true);

      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      expect(emitted[0]).toEqual([
        {
          apiKey: undefined,
          selectedCredentials: undefined,
          credentialsSource: undefined,
          accessFromStorage: true,
        },
      ]);

      // Verify previous API key is preserved in local state
      expect(wrapper.vm.localApiKey).toBe("test-key");
    });

    it("emits correct model when enabling access from storage with credential from secret method", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        accessFromStorage: false,
        selectedCredentials: "test-cred",
        credentialsSource: ERunaiCredentialSource.SharedSecret,
      });

      wrapper.vm.onAccessFromStorageChanged(true);

      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      expect(emitted[0]).toEqual([
        {
          apiKey: undefined,
          selectedCredentials: undefined,
          credentialsSource: undefined,
          accessFromStorage: true,
        },
      ]);

      // Verify previous API key is preserved in local state
      expect(wrapper.vm.localApiKey).toBe("");
      expect(wrapper.vm.localSelectedCredentials).toBe("test-cred");
      expect(wrapper.vm.localCredentialsSource).toBe(ERunaiCredentialSource.SharedSecret);
    });

    it("emits correct model when disabling access from storage with manual method", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        accessFromStorage: true,
      });

      // Set up manual access method
      wrapper.vm.selectedAccessMethod = ENimAccessModelMethod.Manually;
      wrapper.vm.localApiKey = "preserved-key";

      wrapper.vm.onAccessFromStorageChanged(false);

      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      expect(emitted[0]).toEqual([
        {
          apiKey: "preserved-key",
          selectedCredentials: undefined,
          credentialsSource: undefined,
          accessFromStorage: false,
        },
      ]);
    });

    it("emits correct model when disabling access from storage with credential method", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        accessFromStorage: true,
      });

      // Set up credential access method
      wrapper.vm.selectedAccessMethod = ENimAccessModelMethod.FromSecret;
      wrapper.vm.localSelectedCredentials = "preserved-cred";
      wrapper.vm.localCredentialsSource = ERunaiCredentialSource.SharedSecret;

      wrapper.vm.onAccessFromStorageChanged(false);

      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      expect(emitted[0]).toEqual([
        {
          apiKey: undefined,
          selectedCredentials: "preserved-cred",
          credentialsSource: ERunaiCredentialSource.SharedSecret,
          accessFromStorage: false,
        },
      ]);
    });

    it("preserves local states when toggling access from storage multiple times", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        apiKey: "test-key",
      });

      // Set initial manual access state
      wrapper.vm.selectedAccessMethod = ENimAccessModelMethod.Manually;

      // Enable access from storage
      wrapper.vm.onAccessFromStorageChanged(true);

      // Verify local state is preserved
      expect(wrapper.vm.localApiKey).toBe("test-key");

      // Disable access from storage
      wrapper.vm.onAccessFromStorageChanged(false);

      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue] as Array<[INimAccessModelModel]>;
      const lastEmission = emitted[emitted.length - 1][0];

      // Should restore the previous API key
      expect(lastEmission.apiKey).toBe("test-key");
      expect(lastEmission.accessFromStorage).toBe(false);
    });
  });

  describe("isAccessRequired prop", () => {
    it("should set required validation rules when isAccessRequired is true", () => {
      const wrapper = createWrapper();
      expect(wrapper.vm.apiKeyValidationRules).toHaveLength(2);
      expect(wrapper.vm.apiKeyValidationRules).toContain(wrapper.vm.apiKeyNotEmpty);
      expect(wrapper.vm.apiKeyValidationRules).toContain(wrapper.vm.apiKeyValid);
    });

    it("should set optional validation rules when isAccessRequired is false", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: { ...baseModelValue },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          isAccessRequired: false,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });

      expect(wrapper.vm.apiKeyValidationRules).toHaveLength(1);
      expect(wrapper.vm.apiKeyValidationRules).toContain(wrapper.vm.apiKeyNotRequired);
    });

    it("should default to true when isAccessRequired is not provided", () => {
      const wrapper = createWrapper();
      expect(wrapper.vm.apiKeyValidationRules).toHaveLength(2);
      expect(wrapper.vm.apiKeyValidationRules).toContain(wrapper.vm.apiKeyNotEmpty);
      expect(wrapper.vm.apiKeyValidationRules).toContain(wrapper.vm.apiKeyValid);
    });

    it("should validate empty API key as valid when isAccessRequired is false", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: { ...baseModelValue },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          isAccessRequired: false,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });

      const result = wrapper.vm.apiKeyNotRequired("");
      expect(result).toBe(true);
    });

    it("should validate non-empty API key with isAccessRequired false", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: { ...baseModelValue },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          isAccessRequired: false,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });

      // Mock the apiKeyValid method to return true for a valid key
      vi.spyOn(wrapper.vm, "apiKeyValid").mockReturnValue(true);

      const result = wrapper.vm.apiKeyNotRequired("valid-key");
      expect(result).toBe(true);
    });

    it("should validate invalid API key with isAccessRequired false", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: { ...baseModelValue },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          isAccessRequired: false,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });

      // Mock the apiKeyValid method to return error message for invalid key
      vi.spyOn(wrapper.vm, "apiKeyValid").mockReturnValue(errorMessages.ENTER_A_VALID_NIM_API_KEY);

      const result = wrapper.vm.apiKeyNotRequired("invalid-key");
      expect(result).toBe(errorMessages.ENTER_A_VALID_NIM_API_KEY);
    });
  });

  describe("onDockerRegistryChanged", () => {
    it("emits update with docker registry changes", () => {
      const wrapper = createWrapper();
      const dockerRegistry = {
        dockerRegistryName: "my-registry",
      };
      wrapper.vm.onDockerRegistryChanged(dockerRegistry);
      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      expect(emitted[0]).toEqual([
        {
          ...baseModelValue,
          dockerRegistry: {
            dockerRegistryName: "my-registry",
          },
        },
      ]);
    });

    it("emits update with empty docker registry name", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        dockerRegistry: {
          dockerRegistryName: "existing-registry",
        },
      });
      const dockerRegistry = {
        dockerRegistryName: "",
      };
      wrapper.vm.onDockerRegistryChanged(dockerRegistry);
      const emitted = wrapper.emitted()[ENimAccessModelEvents.UpdateModelValue];
      expect(emitted).toBeTruthy();
      const lastEmission = emitted[0] as [INimAccessModelModel];
      expect(lastEmission[0].dockerRegistry).toEqual({
        dockerRegistryName: "",
      });
    });
  });

  describe("showDockerRegistry computed property", () => {
    it("should be false when supportsDockerRegistry is false", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: { ...baseModelValue },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          supportsDockerRegistry: false,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });

      expect(wrapper.vm.showDockerRegistry).toBe(false);
    });

    it("should be true when access method is Manually", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: {
            ...baseModelValue,
            apiKey: "test-key",
          },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          supportsDockerRegistry: true,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });
      wrapper.vm.selectedAccessMethod = ENimAccessModelMethod.Manually;

      expect(wrapper.vm.showDockerRegistry).toBe(true);
    });

    it("should be false when credential type is NgcApiKey", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: {
            ...baseModelValue,
            credentialType: ERunaiCredentialType.NgcApiKey,
          },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          supportsDockerRegistry: true,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });
      wrapper.vm.selectedAccessMethod = ENimAccessModelMethod.FromSecret;

      expect(wrapper.vm.showDockerRegistry).toBe(false);
    });

    it("should be false when credential type is undefined", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: {
            ...baseModelValue,
            credentialType: undefined,
          },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          supportsDockerRegistry: true,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });
      wrapper.vm.selectedAccessMethod = ENimAccessModelMethod.FromSecret;

      expect(wrapper.vm.showDockerRegistry).toBe(false);
    });

    it("should be true when credential type is GenericSecret", () => {
      const wrapper = mount(NimAccessModel, {
        global: {
          plugins: [Quasar],
          stubs: {
            "docker-registry-selector": MockDockerRegistrySelector,
          },
        },
        props: {
          modelValue: {
            ...baseModelValue,
            credentialType: ERunaiCredentialType.GenericSecret,
          },
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
          supportsDockerRegistry: true,
          scopeModel: {
            scope: Scope.Project,
            projectId,
          },
        },
      });
      wrapper.vm.selectedAccessMethod = ENimAccessModelMethod.FromSecret;

      expect(wrapper.vm.showDockerRegistry).toBe(true);
    });
  });

  describe("dockerRegistrySelectorModel computed property", () => {
    it("returns correct model when dockerRegistry is set", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        dockerRegistry: {
          dockerRegistryName: "test-registry",
        },
      });

      expect(wrapper.vm.dockerRegistrySelectorModel).toEqual({
        dockerRegistryName: "test-registry",
      });
    });

    it("returns model with empty string when dockerRegistry is not set", () => {
      const wrapper = createWrapper(baseModelValue);

      expect(wrapper.vm.dockerRegistrySelectorModel).toEqual({
        dockerRegistryName: "",
      });
    });
  });

  describe("credentialSelectorModel computed property", () => {
    it("returns correct model with all properties", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        selectedCredentials: "test-cred",
        credentialsSource: ERunaiCredentialSource.SharedSecret,
        credentialType: ERunaiCredentialType.GenericSecret,
      });

      expect(wrapper.vm.credentialSelectorModel).toEqual({
        source: ERunaiCredentialSource.SharedSecret,
        type: ERunaiCredentialType.GenericSecret,
        credentialName: "test-cred",
      });
    });

    it("returns model with empty credential name when not set", () => {
      const wrapper = createWrapper(baseModelValue);

      expect(wrapper.vm.credentialSelectorModel).toEqual({
        source: ERunaiCredentialSource.MyCredentials,
        type: undefined,
        credentialName: "",
      });
    });
  });
});
