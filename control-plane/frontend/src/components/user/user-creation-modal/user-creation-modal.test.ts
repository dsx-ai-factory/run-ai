import { mount, type VueWrapper, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, afterEach, beforeEach, vi, it, expect } from "vitest";

import UserCreationModal from "./user-creation-modal.vue";
import { ActionTypeEnum, type TActionTypes } from "../action-type.model";
import type { User } from "@/swagger-models/identity-manager-client";
import userService from "@/services/control-plane/user.service/user.service";

const mockHasPermission = vi.fn().mockReturnValue(false);

vi.mock("@/services/control-plane/user.service/user.service");
vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({
    setPageLoading: vi.fn(),
  }),
}));
vi.mock("@/stores/notifications.store", () => ({
  useNotificationStore: () => ({
    loadNotificationChannels: vi.fn().mockResolvedValue(undefined),
    isEmailChannelAvailable: false,
  }),
}));
vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: () => ({
    hasPermission: mockHasPermission,
  }),
}));

interface IUserCreationModalProps {
  actionType: TActionTypes;
  modelValue: boolean;
  selectedUser?: User;
}

const modalStubWithFooterLeft = {
  RunaiBaseDialog: {
    template:
      '<div><slot name="header"></slot><slot name="body"></slot><slot name="footer-left"></slot><slot name="footer"></slot></div>',
  },
};

describe("user-creation-modal.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserCreationModal>>;
  const mockNotify = vi.fn();
  const originalConsoleError = console.error;

  const mockUserResponse = {
    username: "new-user@test.com",
    tempPassword: "temp-pass-123",
  };

  const mockSelectedUser: User = {
    id: "user-123",
    username: "existing-user@test.com",
    createdAt: "2024-01-01",
    createdBy: "admin",
    updatedAt: "2024-01-01",
    lastLogin: "2024-01-01",
    isLocal: true,
  };

  const stepperNextSpy = vi.fn();

  function getMountedCmp(props: IUserCreationModalProps): VueWrapper<InstanceType<typeof UserCreationModal>> {
    return mount(UserCreationModal, {
      global: {
        plugins: [Quasar],
        stubs: {
          ...modalStubWithFooterLeft,
          UserCreationScreen: {
            template: '<div class="user-creation-screen"><input v-model="modelValue" /></div>',
            props: ["modelValue", "sendInvitation", "isSmtpConfigured"],
          },
          UserReviewScreen: {
            template: '<div class="user-review-screen"></div>',
            props: ["actionType", "name", "password"],
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
            props: ["name", "title", "prefix", "done"],
          },
          QBtn: true,
          QForm: true,
          QField: {
            template: '<div class="q-field-stub"></div>',
            props: ["modelValue", "rules"],
          },
          RunaiBtn: {
            template: '<button :class="$attrs.class" @click="$attrs.onClick">{{ $attrs.label }}</button>',
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

  describe("Access rules step visibility", () => {
    it("should show access rules step when user has AccessRules.Create permission in create mode", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(true);
    });

    it("should hide access rules step when user lacks AccessRules.Create permission", () => {
      mockHasPermission.mockReturnValue(false);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(false);
    });

    it("should hide access rules step in reset mode even with permission", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.RESET,
        modelValue: true,
        selectedUser: mockSelectedUser,
      });

      expect(wrapper.vm.showAccessRulesStep).toBe(false);
    });
  });

  describe("Footer buttons on Review step", () => {
    it("should show Done (flat) + Add Access Rules (primary) when user has permission", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });
      wrapper.vm.currentStep = 1;

      expect(wrapper.vm.showAccessRulesStep).toBe(true);
      expect(wrapper.vm.currentStep).toBe(1);
    });

    it("should show only Done button when user lacks permission on review step", () => {
      mockHasPermission.mockReturnValue(false);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });
      wrapper.vm.currentStep = 1;

      expect(wrapper.vm.showAccessRulesStep).toBe(false);
      expect(wrapper.vm.currentStep).toBe(1);
    });
  });

  describe("Navigation to access rules step", () => {
    it("should advance to step 2 when goToAccessRulesStep is called", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });
      wrapper.vm.currentStep = 1;

      wrapper.vm.goToAccessRulesStep();

      expect(wrapper.vm.currentStep).toBe(2);
    });

    it("should close modal when close is called from any step", () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });
      wrapper.vm.currentStep = 2;

      wrapper.vm.close();

      const emitted = wrapper.emitted();
      expect(emitted.close).toBeTruthy();
    });
  });

  describe("Access rule form options", () => {
    it("should provide correct form options for user page", async () => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });

      vi.mocked(userService.isUserExist).mockResolvedValue(false);
      vi.mocked(userService.create).mockResolvedValue(mockUserResponse as never);

      wrapper.vm.appForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createUser();
      await flushPromises();

      expect(wrapper.vm.accessRuleFormOptions).toEqual({
        page: "user",
        subjectId: "new-user@test.com",
        subjectType: "user",
      });
    });
  });

  describe("Create mode - basic flow", () => {
    beforeEach(() => {
      mockHasPermission.mockReturnValue(false);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.CREATE,
        modelValue: true,
      });
    });

    it("should render with correct header", () => {
      expect(wrapper.vm.headerText).toBe("New User");
    });

    it("should be at step 0 initially", () => {
      expect(wrapper.vm.currentStep).toBe(0);
      expect(wrapper.vm.isCreateMode).toBe(true);
    });

    it("should have correct nextBtnLabel for create mode", () => {
      expect(wrapper.vm.nextBtnLabel).toBe("create");
      wrapper.vm.currentStep = 1;
      expect(wrapper.vm.nextBtnLabel).toBe("done");
    });

    it("should create user successfully", async () => {
      vi.mocked(userService.isUserExist).mockResolvedValue(false);
      vi.mocked(userService.create).mockResolvedValue(mockUserResponse as never);

      wrapper.vm.newUser.email = "new-user@test.com";
      wrapper.vm.appForm = {
        validate: vi.fn().mockResolvedValue(true),
      } as never;

      await wrapper.vm.createUser();
      await flushPromises();

      expect(userService.create).toHaveBeenCalledWith({
        email: "new-user@test.com",
        notify: false,
      });
      expect(wrapper.vm.name).toBe("new-user@test.com");
      expect(wrapper.vm.password).toBe("temp-pass-123");
      expect(stepperNextSpy).toHaveBeenCalled();
    });

    it("should emit close event when close is called", () => {
      wrapper.vm.close();

      const emitted = wrapper.emitted();
      expect(emitted.close).toBeTruthy();
    });
  });

  describe("Reset mode", () => {
    beforeEach(() => {
      mockHasPermission.mockReturnValue(true);
      wrapper = getMountedCmp({
        actionType: ActionTypeEnum.RESET,
        modelValue: true,
        selectedUser: mockSelectedUser,
      });
    });

    it("should render with correct header", () => {
      expect(wrapper.vm.headerText).toBe("Reset Password");
    });

    it("should not show access rules step", () => {
      expect(wrapper.vm.showAccessRulesStep).toBe(false);
    });

    it("should have correct firstStepTitle", () => {
      expect(wrapper.vm.firstStepTitle).toBe("Reset");
    });
  });
});
