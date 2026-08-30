import { flushPromises, mount } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import RunaiCredentialSelectorBySource from "./runai-credential-selector-by-source.vue";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import {
  ECredentialSelectorEvents,
  ERunaiCredentialSource,
  ERunaiCredentialType,
  createResourceNameResolver,
} from "./runai-credential-selector-by-source.model";
import { Quasar } from "quasar";
import { mockConsoleError, restoreConsoleError } from "@/utils/unit-test.util/unit-test.util";
import { MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS } from "@/common/version.constant";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { useClusterStore } from "@/stores/cluster.store";
import {
  mockCredentialModel,
  mockDefaultGenericSecretItems,
  mockGenericSecret,
  mockAllCredentials,
  mockNgcApiKeyCreationRequest,
  mockNgcApiKey,
  mockNgcApiKeyWithClusterResources,
  mockPrivateNgcApiKeyWithClusterResources,
} from "./runai-credential-selector-by-source.mocks";

// Mock the CredentialCreationModal component
const MockCredentialCreationModal = {
  name: "credential-creation-modal",
  template: "<div class='mock-credential-creation-modal'></div>",
  props: {
    credentialType: {
      type: String,
      required: true,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    supportMultipleGenericSecretItems: {
      type: Boolean,
      required: false,
      default: true,
    },
    defaultGenericSecretItems: {
      type: Array,
      required: false,
    },
    disabledKeysList: {
      type: Array,
      required: false,
    },
  },
  emits: ["cancel", "create"],
};

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

vi.mock("@/services/control-plane/credential.service/credential.service");

// Mock Quasar notify
const mockNotify = vi.fn();
const $q = {
  notify: mockNotify,
};

describe("runai-credential-selector-by-source.vue", () => {
  const clusterStore = useClusterStore(pinia);
  const defaultProps = {
    modelValue: {
      source: ERunaiCredentialSource.SharedSecret,
      type: undefined,
      credentialName: "",
    },
    clusterId: "",
    scopeModel: {
      scope: Scope.Project,
      projectId: 123,
      clusterId: undefined,
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();

    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    vi.mocked(credentialService.listCredentials).mockResolvedValue(mockAllCredentials);
  });

  const createWrapper = (props = {}) => {
    const finalProps = {
      ...defaultProps,
      typeSelectConfig: {
        hidden: true,
        disable: false,
        allowedTypes: [ERunaiCredentialType.GenericSecret],
      },
      sourceSelectConfig: {
        hidden: false,
        disable: false,
        showPersonalCredentialsOption: true,
      },
      credentialSelectConfig: {
        required: true,
        showCreateNewBtn: false,
        supportMultipleGenericSecretItems: false,
      },
      ...props,
    };

    return mount(RunaiCredentialSelectorBySource, {
      global: {
        plugins: [Quasar],
        mocks: {
          $q,
        },
        stubs: {
          "credential-creation-modal": MockCredentialCreationModal,
        },
      },
      props: finalProps,
    });
  };

  it("should mount successfully", async () => {
    const wrapper = createWrapper();
    await flushPromises();
    expect(wrapper.find(".credential-selector").exists()).toBe(true);
    expect(wrapper.find("[aid='credential-source-select']").exists()).toBe(true);
    expect(wrapper.find("[aid='credential-list-select']").exists()).toBe(true);
  });

  it("should load all credentials on created", async () => {
    const wrapper = createWrapper();
    await flushPromises();
    expect(credentialService.listCredentials).toHaveBeenCalledWith({
      projectId: 123,
      clusterUuid: undefined,
      departmentId: undefined,
      includePrivate: true,
      statusInfo: true,
    });

    // Should have loaded all supported credential types
    expect(wrapper.vm.allCredentialsOptions).toHaveLength(6); // 4 shared + 2 private
  });

  it("should filter credentials by source (SharedSecret)", async () => {
    const wrapper = createWrapper({
      modelValue: {
        source: ERunaiCredentialSource.SharedSecret,
        type: ERunaiCredentialType.GenericSecret,
        credentialName: "",
      },
    });
    await flushPromises();

    // filteredCredentialsOptions should only show SharedSecret + GenericSecret credentials
    const options = wrapper.vm.filteredCredentialsOptions;
    expect(options).toHaveLength(2);
    expect(options[0].label).toBe("shared-generic-secret-1");
    expect(options[1].label).toBe("shared-generic-secret-2");
  });

  it("should filter credentials by source (MyCredentials)", async () => {
    const wrapper = createWrapper({
      modelValue: {
        source: ERunaiCredentialSource.MyCredentials,
        type: ERunaiCredentialType.GenericSecret,
        credentialName: "",
      },
    });
    await flushPromises();

    // filteredCredentialsOptions should only show MyCredentials + GenericSecret credentials
    const options = wrapper.vm.filteredCredentialsOptions;
    expect(options).toHaveLength(1);
    expect(options[0].label).toBe("my-generic-secret-1");
  });

  it("should handle credential source change from SharedSecret to MyCredentials", async () => {
    const wrapper = createWrapper();
    await flushPromises();

    wrapper.vm.onCredentialSourceChanged(ERunaiCredentialSource.MyCredentials);

    // Should NOT reload credentials - just filter client-side
    expect(credentialService.listCredentials).toHaveBeenCalledTimes(1); // Only initial load

    const emitted = wrapper.emitted(ECredentialSelectorEvents.UpdateModelValue);
    expect(emitted?.[0]?.[0]).toEqual({
      source: ERunaiCredentialSource.MyCredentials,
      type: ERunaiCredentialType.GenericSecret,
      credentialName: "",
      id: undefined,
    });
  });

  it("should handle credential source change from MyCredentials to SharedSecret", async () => {
    const wrapper = createWrapper({
      modelValue: {
        source: ERunaiCredentialSource.MyCredentials,
        type: undefined,
        credentialName: "",
      },
    });
    await flushPromises();

    wrapper.vm.onCredentialSourceChanged(ERunaiCredentialSource.SharedSecret);

    // Should NOT reload credentials - just filter client-side
    expect(credentialService.listCredentials).toHaveBeenCalledTimes(1); // Only initial load

    const emitted = wrapper.emitted(ECredentialSelectorEvents.UpdateModelValue);
    expect(emitted?.[0]?.[0]).toEqual({
      source: ERunaiCredentialSource.SharedSecret,
      type: ERunaiCredentialType.GenericSecret,
      credentialName: "",
      id: undefined,
    });
  });

  it("should handle credential selection", async () => {
    const wrapper = createWrapper({
      modelValue: {
        source: ERunaiCredentialSource.SharedSecret,
        type: ERunaiCredentialType.GenericSecret,
        credentialName: "",
      },
    });
    await flushPromises();

    wrapper.vm.onCredentialChanged({
      value: "shared-gen1",
      label: "shared-generic-secret-1",
    });

    const emitted = wrapper.emitted(ECredentialSelectorEvents.UpdateModelValue);
    expect(emitted?.[0]?.[0]).toEqual({
      source: ERunaiCredentialSource.SharedSecret,
      type: ERunaiCredentialType.GenericSecret,
      credentialName: "shared-generic-secret-1",
      id: "shared-gen1",
    });
  });

  it("should handle error when loading credentials", async () => {
    const errorLog = mockConsoleError();
    const error = new Error("Failed to load credentials");
    vi.mocked(credentialService.listCredentials).mockRejectedValue(error);
    createWrapper();
    await flushPromises();

    expect(mockNotify).toHaveBeenCalled();
    const notifyArgs = mockNotify.mock.calls[0][0];
    expect(notifyArgs.message).toContain("Secrets couldn't be loaded");
    expect(console.error).toHaveBeenCalled();
    restoreConsoleError(errorLog);
  });

  describe("credential source options based on cluster version", () => {
    it("should show only Shared Secret option when cluster version is insufficient", async () => {
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);
      const wrapper = createWrapper({
        clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        scopeModel: {
          scope: Scope.Project,
          projectId: 123,
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        },
      });
      await flushPromises();
      expect(wrapper.vm.credentialSourceOptions).toHaveLength(1);
      expect(wrapper.vm.credentialSourceOptions[0]).toEqual({
        value: ERunaiCredentialSource.SharedSecret,
        label: "Shared secret",
        toolTip: "Select a predefined Kubernetes secret",
      });
    });
    it("should show both Shared Secret and My Credentials options when cluster version is sufficient", async () => {
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
      const wrapper = createWrapper({
        clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        scopeModel: {
          scope: Scope.Project,
          projectId: 123,
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        },
      });
      await flushPromises();
      expect(wrapper.vm.credentialSourceOptions).toHaveLength(2);
      expect(wrapper.vm.credentialSourceOptions[0]).toEqual({
        value: ERunaiCredentialSource.SharedSecret,
        label: "Shared secret",
        toolTip: "Select a predefined Kubernetes secret",
      });
      expect(wrapper.vm.credentialSourceOptions[1]).toEqual({
        value: ERunaiCredentialSource.MyCredentials,
        label: "My credentials",
        toolTip: "These are only available to you. To add new credentials, go to User settings.",
      });
    });
    it("should maintain My Credentials selection when cluster version is sufficient", async () => {
      const clusterStore = useClusterStore(pinia);
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
      const wrapper = createWrapper({
        clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
        scopeModel: {
          scope: Scope.Project,
          projectId: 123,
          clusterId: MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialSource).toBe(ERunaiCredentialSource.MyCredentials);
      expect(wrapper.vm.credentialSourceOptions).toHaveLength(2);
    });
  });
  describe("isRequired prop functionality", () => {
    it("should set validation rules when isRequired is true (default)", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(wrapper.vm.validationRules).toHaveLength(2);
      expect(wrapper.vm.validationRules[0]).toBe(wrapper.vm.required);
    });
    it("should not set validation rules when isRequired is false", async () => {
      const wrapper = createWrapper({ credentialSelectConfig: { required: false } });
      await flushPromises();
      expect(wrapper.vm.validationRules).toHaveLength(1);
    });
    it("should set validation rules when isRequired is explicitly true", async () => {
      const wrapper = createWrapper({ credentialSelectConfig: { required: true } });
      await flushPromises();
      expect(wrapper.vm.validationRules).toHaveLength(2);
      expect(wrapper.vm.validationRules[0]).toBe(wrapper.vm.required);
    });
    describe("required validation function", () => {
      it("should return true when value is not empty", async () => {
        const wrapper = createWrapper();
        await flushPromises();
        const result = wrapper.vm.required("some-credential");
        expect(result).toBe(true);
      });
      it("should return error message when value is empty string", async () => {
        const wrapper = createWrapper();
        await flushPromises();
        const result = wrapper.vm.required("");
        expect(result).toBe("Select credential");
      });
      it("should return error message when value is undefined", async () => {
        const wrapper = createWrapper();
        await flushPromises();
        const result = wrapper.vm.required(undefined as unknown as string);
        expect(result).toBe("Select credential");
      });
      it("should return error message when value is null", async () => {
        const wrapper = createWrapper();
        await flushPromises();
        const result = wrapper.vm.required(null as unknown as string);
        expect(result).toBe("Select credential");
      });
    });
  });
  describe("defaultGenericSecretItems prop functionality", () => {
    it("should accept defaultGenericSecretItems prop", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { defaultGenericSecretItems: mockDefaultGenericSecretItems },
      });
      await flushPromises();
      expect(wrapper.props("credentialSelectConfig")?.defaultGenericSecretItems).toEqual(mockDefaultGenericSecretItems);
    });
    it("should make defaultGenericSecretItems optional", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(wrapper.props("credentialSelectConfig")?.defaultGenericSecretItems).toBeUndefined();
    });
  });

  describe("showCreateNewPersonalCredentialsBtn prop functionality", () => {
    it("should default showCreateNewPersonalCredentialsBtn to false", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(wrapper.props("credentialSelectConfig")?.showCreateNewBtn).toBe(false);
    });
    it("should accept showCreateNewPersonalCredentialsBtn prop as true", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
      });
      await flushPromises();
      expect(wrapper.props("credentialSelectConfig")?.showCreateNewBtn).toBe(true);
    });
  });
  describe("customAction computed property", () => {
    it("should return undefined when showCreateNewPersonalCredentialsBtn is false", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: false },
      });
      await flushPromises();
      expect(wrapper.vm.customAction).toBeUndefined();
    });
    it("should return undefined when selectedCredentialSource is not MyCredentials", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.customAction).toBeUndefined();
    });
    it("should return custom action when both conditions are met", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.customAction).toEqual({
        text: "new credential",
      });
    });
    it("should update customAction when credential source changes", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
      });
      await flushPromises();
      // Initially should be undefined (SharedSecret)
      expect(wrapper.vm.customAction).toBeUndefined();
      // Change to MyCredentials
      wrapper.vm.onCredentialSourceChanged(ERunaiCredentialSource.MyCredentials);
      await flushPromises();
      expect(wrapper.vm.customAction).toEqual({
        text: "new credential",
      });
    });
  });
  describe("modal state management", () => {
    it("should initialize modal state correctly", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(wrapper.vm.showPersonalCredentialModal).toBe(false);
      expect(wrapper.vm.isCreating).toBe(false);
      expect(wrapper.vm.personalCredentialType).toBe(AssetKind.GenericSecret);
    });
    it("should show modal when custom action is triggered", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      // Initially modal should be hidden
      expect(wrapper.vm.showPersonalCredentialModal).toBe(false);
      // Trigger custom action
      wrapper.vm.showPersonalCredentialModal = true;
      expect(wrapper.vm.showPersonalCredentialModal).toBe(true);
    });
    it("should hide modal when cancel is triggered", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      // Show modal
      wrapper.vm.showPersonalCredentialModal = true;
      expect(wrapper.vm.showPersonalCredentialModal).toBe(true);
      // Hide modal
      wrapper.vm.showPersonalCredentialModal = false;
      expect(wrapper.vm.showPersonalCredentialModal).toBe(false);
    });
  });
  describe("createCredential method", () => {
    beforeEach(() => {
      vi.mocked(credentialService.prepareGenericSecretForServerRequest).mockReturnValue(mockCredentialModel);
      vi.mocked(credentialService.createGeneralCredential).mockResolvedValue(mockGenericSecret);
    });
    it("should create generic secret credential successfully", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      // Mock the response to include the new credential after creation
      vi.mocked(credentialService.listCredentials).mockResolvedValueOnce([
        ...mockAllCredentials,
        {
          meta: {
            id: "new-gen1",
            name: "test-credential",
            kind: AssetKind.GenericSecret,
            scope: Scope.Project,
            createdBy: "user1",
            createdAt: "2023-01-01T00:00:00Z",
            updatedBy: "user1",
            updatedAt: "2023-01-01T00:00:00Z",
            private: true,
          },
          spec: {
            genericSecret: { keyId: 0 },
          },
        },
      ]);
      await wrapper.vm.createCredential(mockCredentialModel);
      expect(credentialService.prepareGenericSecretForServerRequest).toHaveBeenCalledWith(mockCredentialModel);
      expect(credentialService.createGeneralCredential).toHaveBeenCalledWith(mockCredentialModel);
      expect(mockNotify).toHaveBeenCalledWith({
        message: "Credential test-credential created",
        color: "positive",
      });
      expect(wrapper.vm.showPersonalCredentialModal).toBe(false);
    });
    it("should create NGC API Key credential successfully", async () => {
      vi.mocked(credentialService.createNgcApiKey).mockResolvedValue(mockNgcApiKey);
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        typeSelectConfig: {
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.NgcApiKey,
          credentialName: "",
        },
      });
      await flushPromises();
      // Mock the response to include the new credential after creation
      vi.mocked(credentialService.listCredentials).mockResolvedValueOnce([
        ...mockAllCredentials,
        {
          meta: {
            id: "new-ngc1",
            name: "test-ngc-credential",
            kind: AssetKind.NgcApiKey,
            scope: Scope.Project,
            createdBy: "user1",
            createdAt: "2023-01-01T00:00:00Z",
            updatedBy: "user1",
            updatedAt: "2023-01-01T00:00:00Z",
            private: true,
          },
          spec: {
            ngcApiKey: { keyId: 0 },
          },
        },
      ]);
      await wrapper.vm.createCredential(mockNgcApiKeyCreationRequest);
      expect(credentialService.createNgcApiKey).toHaveBeenCalledWith(mockNgcApiKeyCreationRequest);
      expect(mockNotify).toHaveBeenCalledWith({
        message: "Credential test-ngc-credential created",
        color: "positive",
      });
      expect(wrapper.vm.showPersonalCredentialModal).toBe(false);
    });
    it("should set isCreating flag during credential creation", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      // Initially should be false
      expect(wrapper.vm.isCreating).toBe(false);
      // Start creation process
      const createPromise = wrapper.vm.createCredential(mockCredentialModel);
      // Should be true during creation
      expect(wrapper.vm.isCreating).toBe(true);
      await createPromise;
      // Should be false after completion
      expect(wrapper.vm.isCreating).toBe(false);
    });
    it("should handle generic secret creation error", async () => {
      const errorLog = mockConsoleError();
      const error = new Error("Failed to create credential");
      vi.mocked(credentialService.createGeneralCredential).mockRejectedValue(error);
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      await wrapper.vm.createCredential(mockCredentialModel);
      expect(mockNotify).toHaveBeenCalled();
      const notifyArgs = mockNotify.mock.calls[0][0];
      expect(notifyArgs.message).toContain("Failed to create credential");
      expect(console.error).toHaveBeenCalled();
      expect(wrapper.vm.isCreating).toBe(false);
      restoreConsoleError(errorLog);
    });
    it("should handle NGC API Key creation error", async () => {
      const errorLog = mockConsoleError();
      const error = new Error("Failed to create NGC API Key");
      vi.mocked(credentialService.createNgcApiKey).mockRejectedValue(error);
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        typeSelectConfig: {
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.NgcApiKey,
          credentialName: "",
        },
      });
      await flushPromises();
      await wrapper.vm.createCredential(mockNgcApiKeyCreationRequest);
      expect(mockNotify).toHaveBeenCalled();
      const notifyArgs = mockNotify.mock.calls[0][0];
      expect(notifyArgs.message).toContain("Failed to create credential");
      expect(console.error).toHaveBeenCalled();
      expect(wrapper.vm.isCreating).toBe(false);
      restoreConsoleError(errorLog);
    });
    it("should select newly created credential after successful creation", async () => {
      // Mock the response to include the new credential after creation
      vi.mocked(credentialService.listCredentials).mockResolvedValue([
        ...mockAllCredentials,
        {
          meta: {
            id: "new-gen1",
            name: "test-credential",
            kind: AssetKind.GenericSecret,
            scope: Scope.Project,
            createdBy: "user1",
            createdAt: "2023-01-01T00:00:00Z",
            updatedBy: "user1",
            updatedAt: "2023-01-01T00:00:00Z",
            private: true,
          },
          spec: {
            genericSecret: { keyId: 0 },
          },
        },
      ]);
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      await wrapper.vm.createCredential(mockCredentialModel);
      // Should emit the credential selection
      const emitted = wrapper.emitted(ECredentialSelectorEvents.UpdateModelValue);
      const lastEmitted = emitted?.[emitted.length - 1]?.[0];
      expect(lastEmitted).toEqual({
        source: ERunaiCredentialSource.MyCredentials,
        type: ERunaiCredentialType.GenericSecret,
        credentialName: "test-credential",
        id: "new-gen1",
      });
    });
    it("should reload credentials after successful creation", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      // Clear the mock calls to track new calls
      vi.clearAllMocks();
      vi.mocked(credentialService.prepareGenericSecretForServerRequest).mockReturnValue(mockCredentialModel);
      vi.mocked(credentialService.createGeneralCredential).mockResolvedValue(mockGenericSecret);
      vi.mocked(credentialService.listCredentials).mockResolvedValue(mockAllCredentials);
      await wrapper.vm.createCredential(mockCredentialModel);
      expect(credentialService.listCredentials).toHaveBeenCalledWith({
        projectId: 123,
        clusterUuid: undefined,
        departmentId: undefined,
        includePrivate: true,
        statusInfo: true,
      });
    });
  });
  describe("CredentialCreationModal integration", () => {
    it("should render CredentialCreationModal when showPersonalCredentialModal is true", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      // Initially modal should not be rendered
      expect(wrapper.findComponent({ name: "credential-creation-modal" }).exists()).toBe(false);
      // Show modal
      await wrapper.setData({ showPersonalCredentialModal: true });
      expect(wrapper.findComponent({ name: "credential-creation-modal" }).exists()).toBe(true);
    });
    it("should pass correct props to CredentialCreationModal", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { defaultGenericSecretItems: mockDefaultGenericSecretItems },
      });
      await flushPromises();
      await wrapper.setData({ showPersonalCredentialModal: true });
      const modal = wrapper.findComponent({ name: "credential-creation-modal" });
      expect(modal.props("credentialType")).toBe(AssetKind.GenericSecret);
      expect(modal.props("loading")).toBe(false);
      expect(modal.props("supportMultipleGenericSecretItems")).toBe(false);
      expect(modal.props("defaultGenericSecretItems")).toEqual(mockDefaultGenericSecretItems);
      expect(modal.props("disabledKeysList")).toEqual(["API_KEY", "SECRET_TOKEN"]);
    });
    it("should pass disabledKeysList to CredentialCreationModal", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { defaultGenericSecretItems: mockDefaultGenericSecretItems },
      });
      await flushPromises();
      await wrapper.setData({ showPersonalCredentialModal: true });
      const modal = wrapper.findComponent({ name: "credential-creation-modal" });
      expect(modal.props("disabledKeysList")).toEqual(["API_KEY", "SECRET_TOKEN"]);
    });
    it("should pass empty disabledKeysList when no defaultGenericSecretItems", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      await wrapper.setData({ showPersonalCredentialModal: true });
      const modal = wrapper.findComponent({ name: "credential-creation-modal" });
      expect(modal.props("disabledKeysList")).toEqual([]);
    });
    it("should pass NgcApiKey as credentialType when credential type is NgcApiKey", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.NgcApiKey,
          credentialName: "",
        },
      });
      await flushPromises();
      await wrapper.setData({ showPersonalCredentialModal: true });
      const modal = wrapper.findComponent({ name: "credential-creation-modal" });
      expect(modal.props("credentialType")).toBe(AssetKind.NgcApiKey);
    });
    it("should handle modal cancel event", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      await wrapper.setData({ showPersonalCredentialModal: true });
      expect(wrapper.vm.showPersonalCredentialModal).toBe(true);
      // Trigger cancel event
      const modal = wrapper.findComponent({ name: "credential-creation-modal" });
      await modal.vm.$emit("cancel");
      expect(wrapper.vm.showPersonalCredentialModal).toBe(false);
    });
    it("should handle modal create event", async () => {
      vi.mocked(credentialService.prepareGenericSecretForServerRequest).mockReturnValue(mockCredentialModel);
      vi.mocked(credentialService.createGeneralCredential).mockResolvedValue(mockGenericSecret);
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      await wrapper.setData({ showPersonalCredentialModal: true });
      // Trigger create event
      const modal = wrapper.findComponent({ name: "credential-creation-modal" });
      await modal.vm.$emit("create", mockCredentialModel);
      expect(credentialService.prepareGenericSecretForServerRequest).toHaveBeenCalledWith(mockCredentialModel);
      expect(credentialService.createGeneralCredential).toHaveBeenCalledWith(mockCredentialModel);
    });
  });
  describe("RunaiSelect custom action integration", () => {
    it("should pass customAction to RunaiSelect when conditions are met", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      const runaiSelects = wrapper.findAllComponents({ name: "runai-select" });
      const credentialListSelect = runaiSelects[runaiSelects.length - 1]; // Last runai-select is the credential list
      expect(credentialListSelect.props("customAction")).toEqual({
        text: "new credential",
      });
    });
    it("should handle custom-action event from RunaiSelect", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: { showCreateNewBtn: true },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      const runaiSelects = wrapper.findAllComponents({ name: "runai-select" });
      const credentialListSelect = runaiSelects[runaiSelects.length - 1]; // Last runai-select is the credential list
      // Initially modal should be hidden
      expect(wrapper.vm.showPersonalCredentialModal).toBe(false);
      // Trigger custom action
      await credentialListSelect.vm.$emit("custom-action");
      expect(wrapper.vm.showPersonalCredentialModal).toBe(true);
    });
  });
  describe("disabledKeysList computed property", () => {
    it("should return array of keys from defaultGenericSecretItems", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: {
          defaultGenericSecretItems: mockDefaultGenericSecretItems,
        },
      });
      await flushPromises();
      expect(wrapper.vm.disabledKeysList).toEqual(["API_KEY", "SECRET_TOKEN"]);
    });
    it("should return empty array when defaultGenericSecretItems is undefined", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      expect(wrapper.vm.disabledKeysList).toEqual([]);
    });
    it("should filter out items with null, undefined, or empty keys", async () => {
      const wrapper = createWrapper({
        credentialSelectConfig: {
          defaultGenericSecretItems: [
            { key: "API_KEY", value: "" },
            { key: null, value: "" },
            { key: undefined, value: "" },
            { key: "", value: "" },
            { key: "SECRET_TOKEN", value: "" },
          ],
        },
      });
      await flushPromises();
      expect(wrapper.vm.disabledKeysList).toEqual(["API_KEY", "SECRET_TOKEN"]);
    });
  });
  describe("personalCredentialType computed property", () => {
    it("should return NgcApiKey when selectedCredentialType is NgcApiKey", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.NgcApiKey,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.personalCredentialType).toBe(AssetKind.NgcApiKey);
    });
    it("should return GenericSecret when selectedCredentialType is GenericSecret", async () => {
      const wrapper = createWrapper({
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.personalCredentialType).toBe(AssetKind.GenericSecret);
    });
    it("should return GenericSecret when selectedCredentialType is null", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      wrapper.vm.selectedCredentialType = null;
      expect(wrapper.vm.personalCredentialType).toBe(AssetKind.GenericSecret);
    });
  });
  describe("initializeCredentialType method", () => {
    it("should initialize with NgcApiKey when NGC is not disabled and no type in modelValue", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.NgcApiKey);
    });
    it("should initialize with GenericSecret when NGC is disabled and no type in modelValue", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: { disabledTypes: { [ERunaiCredentialType.NgcApiKey]: {} } },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: undefined,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.GenericSecret);
    });
    it("should use type from modelValue when provided, regardless of disabledTypes", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {},
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.GenericSecret);
    });
    it("should use NgcApiKey from modelValue when provided, even if NGC is disabled", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {
          disabledTypes: { [ERunaiCredentialType.NgcApiKey]: {} },
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials,
          type: ERunaiCredentialType.NgcApiKey,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.NgcApiKey);
    });
    it("should default to GenericSecret when NGC is disabled and modelValue has no type", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: { disabledTypes: { [ERunaiCredentialType.NgcApiKey]: {} } },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.GenericSecret);
    });
  });
  describe("Credential Type functionality", () => {
    describe("showCredentialType prop", () => {
      it("should not render credential type select when showCredentialType is false", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: { hidden: true },
        });
        await flushPromises();
        const credentialTypeSelect = wrapper.find("[aid='credential-type-select']");
        expect(credentialTypeSelect.exists()).toBe(false);
      });
      it("should render credential type select when showCredentialType is true", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: { hidden: false },
        });
        await flushPromises();
        const credentialTypeSelect = wrapper.find("[aid='credential-type-select']");
        expect(credentialTypeSelect.exists()).toBe(true);
      });
    });
    describe("credential type options", () => {
      it("should have correct options when no types are disabled", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
          },
        });
        await flushPromises();
        expect(wrapper.vm.credentialTypeOptions).toHaveLength(2);
        const ngcOption = wrapper.vm.credentialTypeOptions.find((opt) => opt.value === ERunaiCredentialType.NgcApiKey);
        expect(ngcOption).toMatchObject({
          value: ERunaiCredentialType.NgcApiKey,
          label: "NGC API key",
          disable: false,
          disabledTooltip: undefined,
        });
        const genericOption = wrapper.vm.credentialTypeOptions.find(
          (opt) => opt.value === ERunaiCredentialType.GenericSecret,
        );
        expect(genericOption).toMatchObject({
          value: ERunaiCredentialType.GenericSecret,
          label: "Generic secret",
          disable: false,
          disabledTooltip: undefined,
        });
      });
      it("should disable type with tooltip when in disabledTypes", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
            disabledTypes: {
              [ERunaiCredentialType.NgcApiKey]: {
                tooltip:
                  "This credential type is supported only for NIM inference workloads created on clusters version 2.23 or higher",
              },
            },
          },
        });
        await flushPromises();
        expect(wrapper.vm.credentialTypeOptions).toHaveLength(2);
        const ngcOption = wrapper.vm.credentialTypeOptions.find((opt) => opt.value === ERunaiCredentialType.NgcApiKey);
        expect(ngcOption).toMatchObject({
          value: ERunaiCredentialType.NgcApiKey,
          label: "NGC API key",
          disable: true,
          disabledTooltip:
            "This credential type is supported only for NIM inference workloads created on clusters version 2.23 or higher",
        });
        const genericOption = wrapper.vm.credentialTypeOptions.find(
          (opt) => opt.value === ERunaiCredentialType.GenericSecret,
        );
        expect(genericOption).toMatchObject({
          value: ERunaiCredentialType.GenericSecret,
          label: "Generic secret",
          disable: false,
          disabledTooltip: undefined,
        });
      });
    });
    describe("filtering credentials by type", () => {
      it("should filter to show only NGC API Key credentials when type is NgcApiKey", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();
        const options = wrapper.vm.filteredCredentialsOptions;
        expect(options).toHaveLength(1);
        expect(options[0].label).toBe("shared-ngc-key-1");
      });
      it("should filter to show only Generic Secret credentials when type is GenericSecret", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: { hidden: false },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.GenericSecret,
            credentialName: "",
          },
        });
        await flushPromises();
        const options = wrapper.vm.filteredCredentialsOptions;
        expect(options).toHaveLength(2);
        expect(options[0].label).toBe("shared-generic-secret-1");
        expect(options[1].label).toBe("shared-generic-secret-2");
      });
    });
    describe("switching credential types", () => {
      it("should filter credentials when switching from Generic Secret to NGC API Key", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.GenericSecret,
            credentialName: "",
          },
        });
        await flushPromises();
        // Initially should have generic secrets
        expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.GenericSecret);
        expect(wrapper.vm.filteredCredentialsOptions).toHaveLength(2);
        expect(wrapper.vm.filteredCredentialsOptions[0].label).toBe("shared-generic-secret-1");
        // Change to NGC API Key
        wrapper.vm.onCredentialTypeChanged(ERunaiCredentialType.NgcApiKey);
        await flushPromises();
        // Should NOT reload credentials - just filter client-side
        expect(credentialService.listCredentials).toHaveBeenCalledTimes(1); // Only initial load
        expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.NgcApiKey);
        expect(wrapper.vm.filteredCredentialsOptions).toHaveLength(1);
        expect(wrapper.vm.filteredCredentialsOptions[0].label).toBe("shared-ngc-key-1");
      });
      it("should filter credentials when switching from NGC API Key to Generic Secret", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();
        // Initially should have NGC API Keys
        expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.NgcApiKey);
        expect(wrapper.vm.filteredCredentialsOptions).toHaveLength(1);
        expect(wrapper.vm.filteredCredentialsOptions[0].label).toBe("shared-ngc-key-1");
        // Change to Generic Secret
        wrapper.vm.onCredentialTypeChanged(ERunaiCredentialType.GenericSecret);
        await flushPromises();
        // Should NOT reload credentials - just filter client-side
        expect(credentialService.listCredentials).toHaveBeenCalledTimes(1); // Only initial load
        expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.GenericSecret);
        expect(wrapper.vm.filteredCredentialsOptions).toHaveLength(2);
        expect(wrapper.vm.filteredCredentialsOptions[0].label).toBe("shared-generic-secret-1");
      });
      it("should emit type change when credential type is changed", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.GenericSecret,
            credentialName: "",
          },
        });
        await flushPromises();
        wrapper.vm.onCredentialTypeChanged(ERunaiCredentialType.NgcApiKey);
        const emitted = wrapper.emitted(ECredentialSelectorEvents.UpdateModelValue);
        expect(emitted?.[0]?.[0]).toEqual({
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.NgcApiKey,
          credentialName: "",
          id: undefined,
        });
      });
    });
    describe("emitting credential type changes", () => {
      it("should emit type when credential is selected with NGC API Key type", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();
        wrapper.vm.onCredentialChanged({
          value: "shared-ngc1",
          label: "shared-ngc-key-1",
        });
        const emitted = wrapper.emitted(ECredentialSelectorEvents.UpdateModelValue);
        expect(emitted?.[0]?.[0]).toEqual({
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.NgcApiKey,
          credentialName: "shared-ngc-key-1",
          id: "shared-ngc1",
        });
      });
      it("should emit type when credential is selected with Generic Secret type", async () => {
        const wrapper = createWrapper({
          typeSelectConfig: { hidden: false },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.GenericSecret,
            credentialName: "",
          },
        });
        await flushPromises();
        wrapper.vm.onCredentialChanged({
          value: "shared-gen1",
          label: "shared-generic-secret-1",
        });
        const emitted = wrapper.emitted(ECredentialSelectorEvents.UpdateModelValue);
        expect(emitted?.[0]?.[0]).toEqual({
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "shared-generic-secret-1",
          id: "shared-gen1",
        });
      });
    });
  });
  describe("validateSourceAndType method", () => {
    it("should update source and type based on credential name", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {
          hidden: false,
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      const credential = await wrapper.vm.getCredentialByName("shared-ngc-key-1");
      if (!credential) {
        throw new Error("Credential not found");
      }
      wrapper.vm.validateSourceAndType(credential);
      expect(wrapper.vm.selectedCredentialSource).toBe(ERunaiCredentialSource.SharedSecret);
      expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.NgcApiKey);
    });
    it("should update to MyCredentials source when credential is private", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {
          hidden: false,
          allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
        },
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      const credential = await wrapper.vm.getCredentialByName("my-generic-secret-1");
      if (!credential) {
        throw new Error("Credential not found");
      }
      wrapper.vm.validateSourceAndType(credential);
      expect(wrapper.vm.selectedCredentialSource).toBe(ERunaiCredentialSource.MyCredentials);
      expect(wrapper.vm.selectedCredentialType).toBe(ERunaiCredentialType.GenericSecret);
    });
    it("should reload credentials if credential not found initially", async () => {
      const wrapper = createWrapper();
      await flushPromises();
      // Reset mock to track new calls
      vi.clearAllMocks();
      // Mock that after reload, the credential is found
      vi.mocked(credentialService.listCredentials).mockResolvedValueOnce([
        ...mockAllCredentials,
        {
          meta: {
            id: "new-cred",
            name: "new-credential",
            kind: AssetKind.GenericSecret,
            scope: Scope.Project,
            createdBy: "user1",
            createdAt: "2023-01-01T00:00:00Z",
            updatedBy: "user1",
            updatedAt: "2023-01-01T00:00:00Z",
            private: false,
          },
          spec: {
            genericSecret: { keyId: 0 },
          },
        },
      ]);
      const credential = await wrapper.vm.getCredentialByName("new-credential");
      if (!credential) {
        throw new Error("Credential not found");
      }
      wrapper.vm.validateSourceAndType(credential);
      expect(credentialService.listCredentials).toHaveBeenCalledTimes(1);
    });
  });
  describe("selectedCredentialValue computed property", () => {
    it("should return id when modelValue has id", async () => {
      const wrapper = createWrapper({
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "shared-generic-secret-1",
          id: "shared-gen1",
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialValue).toBe("shared-gen1");
    });
    it("should find id from credentials list when only name is provided", async () => {
      const wrapper = createWrapper({
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "shared-generic-secret-1",
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialValue).toBe("shared-gen1");
    });
    it("should return null when no credential is selected", async () => {
      const wrapper = createWrapper({
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      expect(wrapper.vm.selectedCredentialValue).toBe(null);
    });
  });

  describe("modelValue watcher", () => {
    it("should call validateSourceAndType when credentialName changes", async () => {
      const wrapper = createWrapper({
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      const validateSpy = vi.spyOn(wrapper.vm, "validateSourceAndType");
      await wrapper.setProps({
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "shared-generic-secret-1",
        },
      });
      await flushPromises();
      const credential = await wrapper.vm.getCredentialByName("shared-generic-secret-1");
      if (!credential) {
        throw new Error("Credential not found");
      }
      expect(validateSpy).toHaveBeenCalledWith(credential);
    });
    it("should not call validateSourceAndType when credentialName does not change", async () => {
      const wrapper = createWrapper({
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "test-cred",
        },
      });
      await flushPromises();
      const validateSpy = vi.spyOn(wrapper.vm, "validateSourceAndType");
      validateSpy.mockClear();
      await wrapper.setProps({
        modelValue: {
          source: ERunaiCredentialSource.MyCredentials, // Change source but not name
          type: ERunaiCredentialType.NgcApiKey, // Change type but not name
          credentialName: "test-cred", // Same name
        },
      });
      await flushPromises();
      expect(validateSpy).not.toHaveBeenCalled();
    });
  });
  describe("filteredCredentialsOptions computed property", () => {
    it("should filter by both source and type", async () => {
      const wrapper = createWrapper({
        typeSelectConfig: {
          hidden: false,
          allowedTypes: [
            ERunaiCredentialType.NgcApiKey,
            ERunaiCredentialType.GenericSecret,
            ERunaiCredentialType.DockerRegistry,
          ],
        },
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.DockerRegistry,
          credentialName: "",
        },
      });
      await flushPromises();
      const options = wrapper.vm.filteredCredentialsOptions;
      expect(options).toHaveLength(1);
      expect(options[0].label).toBe("shared-docker-registry-1");
    });
    it("should return options with value (id) and label (name)", async () => {
      const wrapper = createWrapper({
        modelValue: {
          source: ERunaiCredentialSource.SharedSecret,
          type: ERunaiCredentialType.GenericSecret,
          credentialName: "",
        },
      });
      await flushPromises();
      const options = wrapper.vm.filteredCredentialsOptions;
      expect(options[0]).toEqual({
        disable: true,
        disabledTooltip:
          "This setup is being created. If it isn't created after a few seconds, contact your administrator.",
        value: "shared-gen1",
        label: "shared-generic-secret-1",
      });
    });
  });

  describe("nameResolver prop", () => {
    const credentialsWithNgcResources = [...mockAllCredentials, mockNgcApiKeyWithClusterResources];

    describe("when nameResolver is provided", () => {
      it("should use docker registry resource name with default resolver", async () => {
        vi.mocked(credentialService.listCredentials).mockResolvedValue(credentialsWithNgcResources);

        const wrapper = createWrapper({
          nameResolver: createResourceNameResolver(),
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();

        const ngcCredentialOption = wrapper.vm.allCredentialsOptions.find((opt) => opt.id === "ngc-with-resources");
        expect(ngcCredentialOption?.name).toBe("ngc-api-key-dockerregistry-xyz789");
      });

      it("should use generic secret resource name with GenericSecret resolver", async () => {
        vi.mocked(credentialService.listCredentials).mockResolvedValue(credentialsWithNgcResources);

        const wrapper = createWrapper({
          nameResolver: createResourceNameResolver(AssetKind.GenericSecret),
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();

        const ngcCredentialOption = wrapper.vm.allCredentialsOptions.find((opt) => opt.id === "ngc-with-resources");
        expect(ngcCredentialOption?.name).toBe("ngc-api-key-genericsecret-abc123");
      });

      it("should use credential name for private NGC API key credentials regardless of preference", async () => {
        const credentialsWithPrivateNgc = [...mockAllCredentials, mockPrivateNgcApiKeyWithClusterResources];
        vi.mocked(credentialService.listCredentials).mockResolvedValue(credentialsWithPrivateNgc);

        const wrapper = createWrapper({
          nameResolver: createResourceNameResolver(AssetKind.GenericSecret),
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey],
          },
          modelValue: {
            source: ERunaiCredentialSource.MyCredentials,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();

        const privateNgcOption = wrapper.vm.allCredentialsOptions.find((opt) => opt.id === "private-ngc-with-resources");
        expect(privateNgcOption?.name).toBe("my-ngc-api-key-with-resources");
      });

      it("should fallback to credential name when resource is not found", async () => {
        const ngcWithoutMatchingResource = {
          ...mockNgcApiKeyWithClusterResources,
          clusterInfo: {
            resources: [{ name: "some-unrelated-resource" }],
          },
        };
        vi.mocked(credentialService.listCredentials).mockResolvedValue([
          ...mockAllCredentials,
          ngcWithoutMatchingResource,
        ]);

        const wrapper = createWrapper({
          nameResolver: createResourceNameResolver(AssetKind.GenericSecret),
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();

        const ngcCredentialOption = wrapper.vm.allCredentialsOptions.find((opt) => opt.id === "ngc-with-resources");
        expect(ngcCredentialOption?.name).toBe("ngc-api-key-with-resources");
      });

      it("should not affect non-NGC credentials", async () => {
        vi.mocked(credentialService.listCredentials).mockResolvedValue(credentialsWithNgcResources);

        const wrapper = createWrapper({
          nameResolver: createResourceNameResolver(AssetKind.GenericSecret),
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.GenericSecret,
            credentialName: "",
          },
        });
        await flushPromises();

        const genericSecretOption = wrapper.vm.allCredentialsOptions.find((opt) => opt.id === "shared-gen1");
        expect(genericSecretOption?.name).toBe("shared-generic-secret-1");
      });
    });

    describe("when nameResolver is not provided", () => {
      it("should use credential name by default", async () => {
        vi.mocked(credentialService.listCredentials).mockResolvedValue(credentialsWithNgcResources);

        const wrapper = createWrapper({
          typeSelectConfig: {
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey],
          },
          modelValue: {
            source: ERunaiCredentialSource.SharedSecret,
            type: ERunaiCredentialType.NgcApiKey,
            credentialName: "",
          },
        });
        await flushPromises();

        const ngcCredentialOption = wrapper.vm.allCredentialsOptions.find((opt) => opt.id === "ngc-with-resources");
        expect(ngcCredentialOption?.name).toBe("ngc-api-key-with-resources");
      });
    });
  });
});
