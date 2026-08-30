import { mount, type VueWrapper, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, afterEach, beforeEach, vi, it, expect } from "vitest";

import CredentialCreationModal from "./credential-creation-modal.vue";
import type { TActionTypes, TEntityType } from "./credential-creation-modal.vue";
import type { AccessKey, ServiceAccount } from "@/swagger-models/identity-manager-client";
import { accessKeysService } from "@/services/control-plane/access-keys.service/access-keys.service";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";

const mockHasPermission = vi.fn().mockReturnValue(false);

vi.mock("@/services/control-plane/access-keys.service/access-keys.service");
vi.mock("@/services/control-plane/service-accounts.service/service-accounts.service");
vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({
    setPageLoading: vi.fn(),
  }),
}));
vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: () => ({
    hasPermission: mockHasPermission,
  }),
}));

interface ICredentialCreationModalProps {
  entityType: TEntityType;
  actionType: TActionTypes;
  modelValue: boolean;
  credential?: AccessKey | ServiceAccount;
}

describe("credential-creation-modal.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof CredentialCreationModal>>;
  const mockNotify = vi.fn();
  const originalConsoleError = console.error;

  const mockAccessKey: AccessKey = {
    id: "access-key-123",
    name: "Test Access Key",
    clientId: "client-123",
    createdAt: "2024-01-01",
    lastLogin: "2024-01-01",
  };

  const mockServiceAccount: ServiceAccount = {
    createdBy: "test-user",
    updatedAt: "2024-01-01",
    lastLogin: "2024-01-01",
    id: "service-account-456",
    name: "Test Service Account",
    clientId: "client-456",
    enabled: true,
    createdAt: "2024-01-01",
  };

  const mockAccessKeyResponse = {
    id: "new-access-key",
    name: "New Access Key",
    clientId: "new-client-id",
    clientSecret: "secret-123",
  };

  const mockServiceAccountResponse = {
    id: "new-service-account",
    name: "New Service Account",
    clientId: "new-client-id",
    clientSecret: "secret-456",
  };

  const mockSecretResponse = {
    clientSecret: "regenerated-secret-789",
  };

  const stepperNextSpy = vi.fn();

  function getMountedCmp(
    props: ICredentialCreationModalProps,
  ): VueWrapper<InstanceType<typeof CredentialCreationModal>> {
    return mount(CredentialCreationModal, {
      global: {
        plugins: [Quasar],
        stubs: {
          RunaiBaseDialog: {
            template:
              '<div><slot name="header"></slot><slot name="body"></slot><slot name="footer-left"></slot><slot name="footer"></slot></div>',
          },
          ServiceAccountsCreationScreen: {
            template: '<div class="service-accounts-creation-screen"><input v-model="modelValue" /></div>',
            props: ["modelValue", "isAccessKey"],
          },
          ServiceAccountsViewScreen: {
            template: '<div class="service-accounts-view-screen"></div>',
            props: ["isCreateMode", "isAccessKey", "name", "secret", "client"],
          },
          AccessRuleManagement: {
            template: '<div class="access-rule-management"></div>',
            props: ["formOptions"],
          },
          QStepper: {
            template: '<div class="q-stepper"><slot /></div>',
            methods: {
              next: stepperNextSpy,
            },
          },
          QStep: {
            template: '<div class="q-step"><slot /></div>',
            props: ["name", "title", "prefix", "done", "icon"],
          },
          QBtn: true,
          QForm: true,
          QField: {
            template: '<div class="q-field-stub"></div>',
            props: ["modelValue", "rules"],
          },
        },
        mocks: {
          $q: {
            notify: mockNotify,
          },
        },
      },
      props,
    });
  }

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
    vi.clearAllMocks();
    stepperNextSpy.mockClear();
    console.error = originalConsoleError;
  });

  describe("Access Key - Create Mode", () => {
    beforeEach(() => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "create",
        modelValue: true,
      });
    });

    it("should render with correct header for creating access key", () => {
      expect(wrapper.vm.headerText).toBe("New Access Key");
    });

    it("should be at step 0 initially", () => {
      expect(wrapper.vm.currentStep).toBe(0);
      expect(wrapper.vm.isCreateMode).toBe(true);
    });

    it("should have correct computed properties for access key", () => {
      expect(wrapper.vm.isAccessKey).toBe(true);
      expect(wrapper.vm.entityDisplayName).toBe("Access Key");
      expect(wrapper.vm.entityDisplayNameLower).toBe("access key");
      expect(wrapper.vm.isCreateMode).toBe(true);
    });

    it("should create access key successfully", async () => {
      vi.mocked(accessKeysService.createAccessKey).mockResolvedValue(mockAccessKeyResponse as never);

      wrapper.vm.newCredential.name = "New Access Key";
      wrapper.vm.credentialForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createCredential();
      await flushPromises();

      expect(accessKeysService.createAccessKey).toHaveBeenCalledWith({
        name: "New Access Key",
      });
      expect(wrapper.vm.name).toBe("New Access Key");
      expect(wrapper.vm.secret).toBe("secret-123");
      expect(wrapper.vm.clientId).toBe("new-client-id");
      expect(stepperNextSpy).toHaveBeenCalled();
    });

    it("should emit load-credential event after successful creation", async () => {
      vi.mocked(accessKeysService.createAccessKey).mockResolvedValue(mockAccessKeyResponse as never);

      wrapper.vm.newCredential.name = "New Access Key";
      wrapper.vm.credentialForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createCredential();
      await flushPromises();

      const emitted = wrapper.emitted();
      expect(emitted["load-credential"]).toBeTruthy();
      expect(emitted["load-credential"][0]).toEqual(["new-access-key"]);
    });

    it("should handle creation error gracefully", async () => {
      console.error = vi.fn();
      const mockError = new Error("Creation failed");
      vi.mocked(accessKeysService.createAccessKey).mockRejectedValue(mockError);

      wrapper.vm.newCredential.name = "New Access Key";
      wrapper.vm.credentialForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createCredential();
      await flushPromises();

      expect(wrapper.vm.loading).toBe(false);
      expect(wrapper.vm.currentStep).toBe(0);
    });

    it("should not create if validation fails", async () => {
      wrapper.vm.credentialForm = {
        validate: vi.fn().mockResolvedValue(false),
      } as never;

      await wrapper.vm.createCredential();

      expect(accessKeysService.createAccessKey).not.toHaveBeenCalled();
    });
  });

  describe("Access Key - Regenerate Mode", () => {
    beforeEach(() => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "regenerate",
        modelValue: true,
        credential: mockAccessKey,
      });
    });

    it("should render with correct header for regenerating secret", () => {
      expect(wrapper.vm.headerText).toBe("Regenerate Client Secret");
    });

    it("should be in regenerate mode", () => {
      expect(wrapper.vm.isCreateMode).toBe(false);
      expect(wrapper.vm.credential).toEqual(mockAccessKey);
    });

    it("should regenerate access key secret successfully", async () => {
      vi.mocked(accessKeysService.generateAccessKeySecret).mockResolvedValue(mockSecretResponse as never);

      await wrapper.vm.regenerateSecret();
      await flushPromises();

      expect(accessKeysService.generateAccessKeySecret).toHaveBeenCalledWith("access-key-123");
      expect(wrapper.vm.name).toBe("Test Access Key");
      expect(wrapper.vm.clientId).toBe("client-123");
      expect(wrapper.vm.secret).toBe("regenerated-secret-789");
      expect(stepperNextSpy).toHaveBeenCalled();
    });

    it("should handle regenerate error gracefully", async () => {
      console.error = vi.fn();
      const mockError = new Error("Regenerate failed");
      vi.mocked(accessKeysService.generateAccessKeySecret).mockRejectedValue(mockError);

      await wrapper.vm.regenerateSecret();
      await flushPromises();

      expect(wrapper.vm.loading).toBe(false);
      expect(wrapper.vm.currentStep).toBe(0);
    });

    it("should not regenerate if credential is missing", async () => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "regenerate",
        modelValue: true,
      });

      await wrapper.vm.regenerateSecret();

      expect(accessKeysService.generateAccessKeySecret).not.toHaveBeenCalled();
    });
  });

  describe("Service Account - Create Mode", () => {
    beforeEach(() => {
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "create",
        modelValue: true,
      });
    });

    it("should render with correct header for creating service account", () => {
      expect(wrapper.vm.headerText).toBe("New Service Account");
    });

    it("should have correct computed properties for service account", () => {
      expect(wrapper.vm.isAccessKey).toBe(false);
      expect(wrapper.vm.entityDisplayName).toBe("Service Account");
      expect(wrapper.vm.entityDisplayNameLower).toBe("service account");
      expect(wrapper.vm.isCreateMode).toBe(true);
    });

    it("should create service account successfully", async () => {
      vi.mocked(serviceAccountsService.createServiceAccount).mockResolvedValue(mockServiceAccountResponse as never);

      wrapper.vm.newCredential.name = "New Service Account";
      wrapper.vm.credentialForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createCredential();
      await flushPromises();

      expect(serviceAccountsService.createServiceAccount).toHaveBeenCalledWith({
        name: "New Service Account",
      });
      expect(wrapper.vm.name).toBe("New Service Account");
      expect(wrapper.vm.secret).toBe("secret-456");
      expect(wrapper.vm.clientId).toBe("new-client-id");
      expect(stepperNextSpy).toHaveBeenCalled();
    });

    it("should emit load-credential event after successful creation", async () => {
      vi.mocked(serviceAccountsService.createServiceAccount).mockResolvedValue(mockServiceAccountResponse as never);

      wrapper.vm.newCredential.name = "New Service Account";
      wrapper.vm.credentialForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createCredential();
      await flushPromises();

      const emitted = wrapper.emitted();
      expect(emitted["load-credential"]).toBeTruthy();
      expect(emitted["load-credential"][0]).toEqual(["new-service-account"]);
    });
  });

  describe("Service Account - Regenerate Mode", () => {
    beforeEach(() => {
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "regenerate",
        modelValue: true,
        credential: mockServiceAccount,
      });
    });

    it("should be in regenerate mode for service account", () => {
      expect(wrapper.vm.isCreateMode).toBe(false);
      expect(wrapper.vm.credential).toEqual(mockServiceAccount);
    });

    it("should regenerate service account secret successfully", async () => {
      vi.mocked(serviceAccountsService.regenerateSecret).mockResolvedValue(mockSecretResponse as never);

      await wrapper.vm.regenerateSecret();
      await flushPromises();

      expect(serviceAccountsService.regenerateSecret).toHaveBeenCalledWith("service-account-456");
      expect(wrapper.vm.name).toBe("Test Service Account");
      expect(wrapper.vm.clientId).toBe("client-456");
      expect(wrapper.vm.secret).toBe("regenerated-secret-789");
      expect(stepperNextSpy).toHaveBeenCalled();
    });
  });

  describe("Stepper Navigation", () => {
    beforeEach(() => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "create",
        modelValue: true,
      });

      Object.defineProperty(wrapper.vm.$refs, "stepper", {
        value: { next: vi.fn() },
        writable: true,
        configurable: true,
      });
      Object.defineProperty(wrapper.vm, "$q", {
        value: { notify: mockNotify },
        writable: true,
        configurable: true,
      });
      Object.defineProperty(wrapper.vm, "credentialForm", {
        value: { validate: vi.fn().mockResolvedValue(true) },
        writable: true,
        configurable: true,
      });
    });

    it("should compute isLastStep correctly", () => {
      expect(wrapper.vm.isLastStep).toBe(false);
      wrapper.vm.currentStep = 1;
      expect(wrapper.vm.isLastStep).toBe(true);
    });

    it("should compute nextBtnLabel correctly for create mode", () => {
      expect(wrapper.vm.nextBtnLabel).toBe("create");
      wrapper.vm.currentStep = 1;
      expect(wrapper.vm.nextBtnLabel).toBe("done");
    });

    it("should compute nextBtnLabel correctly for regenerate mode", async () => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "regenerate",
        modelValue: true,
        credential: mockAccessKey,
      });

      expect(wrapper.vm.nextBtnLabel).toBe("regenerate");
      wrapper.vm.currentStep = 1;
      expect(wrapper.vm.nextBtnLabel).toBe("done");
    });

    it("should call createCredential on nextStepAction when in create mode at step 0", async () => {
      console.error = vi.fn();
      const createSpy = vi.spyOn(wrapper.vm, "createCredential");
      await wrapper.vm.nextStepAction();
      expect(createSpy).toHaveBeenCalled();
    });

    it("should call regenerateSecret on nextStepAction when in regenerate mode at step 0", async () => {
      console.error = vi.fn();
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "regenerate",
        modelValue: true,
        credential: mockAccessKey,
      });

      Object.defineProperty(wrapper.vm.$refs, "stepper", {
        value: { next: vi.fn() },
        writable: true,
        configurable: true,
      });

      const regenerateSpy = vi.spyOn(wrapper.vm, "regenerateSecret");
      await wrapper.vm.nextStepAction();
      expect(regenerateSpy).toHaveBeenCalled();
    });

    it("should not call createCredential or regenerateSecret on nextStepAction when not at step 0", async () => {
      wrapper.vm.currentStep = 1;
      const createSpy = vi.spyOn(wrapper.vm, "createCredential");

      await wrapper.vm.nextStepAction();

      expect(createSpy).not.toHaveBeenCalled();
    });
  });

  describe("Modal Actions", () => {
    beforeEach(() => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "create",
        modelValue: true,
      });

      Object.defineProperty(wrapper.vm.$refs, "stepper", {
        value: { next: vi.fn() },
        writable: true,
        configurable: true,
      });
      Object.defineProperty(wrapper.vm, "$q", {
        value: { notify: mockNotify },
        writable: true,
        configurable: true,
      });
    });

    it("should emit close event when close is called", () => {
      wrapper.vm.close();

      const emitted = wrapper.emitted();
      expect(emitted.close).toBeTruthy();
    });

    it("should show cancel button when not at last step", () => {
      expect(wrapper.vm.isLastStep).toBe(false);
    });

    it("should hide cancel button when at last step", () => {
      wrapper.vm.currentStep = 1;
      expect(wrapper.vm.isLastStep).toBe(true);
    });
  });

  describe("dialogCustomClass computed property", () => {
    it("should return service-account modifier when entityType is service-account", () => {
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "create",
        modelValue: true,
      });

      expect(wrapper.vm.dialogCustomClass).toBe(
        "credential-creation-dialog credential-creation-dialog--service-account",
      );
    });

    it("should return access-key modifier when entityType is access-key", () => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "create",
        modelValue: true,
      });

      expect(wrapper.vm.dialogCustomClass).toBe("credential-creation-dialog credential-creation-dialog--access-key");
    });
  });

  describe("firstStepTitle computed property", () => {
    it("should return 'Create' for create mode", () => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "create",
        modelValue: true,
      });

      expect(wrapper.vm.firstStepTitle).toBe("Create");
    });

    it("should return 'Regenerate' for regenerate mode", () => {
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "regenerate",
        modelValue: true,
        credential: mockAccessKey,
      });

      expect(wrapper.vm.firstStepTitle).toBe("Regenerate");
    });
  });

  describe("Access rules step visibility", () => {
    it("should show access rules step when user has permission in create mode", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "create",
        modelValue: true,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(true);
    });

    it("should hide access rules step when user lacks permission", () => {
      mockHasPermission.mockReturnValue(false);
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "create",
        modelValue: true,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(false);
    });

    it("should hide access rules step in regenerate mode even with permission", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "regenerate",
        modelValue: true,
        credential: mockServiceAccount,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(false);
    });

    it("should hide access rules step for access key regenerate mode", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "regenerate",
        modelValue: true,
        credential: mockAccessKey,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(false);
    });

    it("should hide access rules step for access key in create mode even with permission", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        entityType: "access-key",
        actionType: "create",
        modelValue: true,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(false);
    });
  });

  describe("Access rules step navigation", () => {
    it("should advance to step 2 when goToAccessRulesStep is called", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "create",
        modelValue: true,
      });
      wrapper.vm.currentStep = 1;

      wrapper.vm.goToAccessRulesStep();

      expect(wrapper.vm.currentStep).toBe(2);
    });

    it("should close modal when close is called from access rules step", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "create",
        modelValue: true,
      });
      wrapper.vm.currentStep = 2;

      wrapper.vm.close();

      const emitted = wrapper.emitted();
      expect(emitted.close).toBeTruthy();
    });

    it("should provide correct form options for service account", async () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        entityType: "service-account",
        actionType: "create",
        modelValue: true,
      });

      vi.mocked(serviceAccountsService.createServiceAccount).mockResolvedValue(mockServiceAccountResponse as never);

      wrapper.vm.newCredential.name = "New Service Account";
      wrapper.vm.credentialForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createCredential();
      await flushPromises();

      expect(wrapper.vm.accessRuleFormOptions).toEqual({
        page: "application",
        subjectId: "New Service Account",
        subjectType: "service-account",
      });
    });
  });
});
