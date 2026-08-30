import { shallowMount } from "@vue/test-utils";
import { describe, it, expect, vi } from "vitest";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import { accessKeysService } from "@/services/control-plane/access-keys.service/access-keys.service";
import { Quasar } from "quasar";
import { ServiceAccountsCreationScreen } from "@/components/service-accounts/service-accounts-creation-screen/index";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";
import { errorMessages } from "@/common/error-message.constant";

vi.mock("@/services/control-plane/service-accounts.service/service-accounts.service");
vi.mock("@/services/control-plane/access-keys.service/access-keys.service");

const mockedServiceAccountsService = vi.mocked(serviceAccountsService);
const mockedAccessKeysService = vi.mocked(accessKeysService);

describe("NameValidationComponent", () => {
  const mockApplications: ServiceAccount[] = [
    {
      name: "existing_app",
      createdBy: "test_user",
      createdAt: "2021-01-01",
      updatedAt: "2021-01-01",
      enabled: true,
      tenantId: "test_tenant",
      lastLogin: "2021-01-01",
      id: "test_id",
      clientId: "test_client_id",
    },
    {
      name: "test_app",
      createdBy: "test_user",
      createdAt: "2021-01-01",
      updatedAt: "2021-01-01",
      enabled: true,
      tenantId: "test_tenant",
      lastLogin: "2021-01-01",
      id: "test_id",
      clientId: "test_client_id",
    },
  ];

  const mountComponent = (props = {}) => {
    return shallowMount(ServiceAccountsCreationScreen, {
      plugins: [Quasar],
      props: {
        modelValue: "",
        isAccessKey: false,
        ...props,
      },
    });
  };

  it("should validate that the name is not empty", () => {
    const wrapper = mountComponent();
    const { isEmpty } = wrapper.vm;

    expect(isEmpty("")).toBe("Enter a name"); // Update with your errorMessages.NAME_NOT_EMPTY
    expect(isEmpty("valid_name")).toBe(true);
  });

  it("should validate the format of the name", () => {
    const wrapper = mountComponent();
    const { isValidFormat } = wrapper.vm;

    expect(isValidFormat("Valid_Name")).toBe(
      "Enter a name starting with a lowercase letter and including only a-z, 0-9, -, _",
    ); // Update with your errorMessages.VALID_APP_FORMAT
    expect(isValidFormat("valid_name")).toBe(true);
  });

  it("should validate the name is not taken", async () => {
    mockedServiceAccountsService.list.mockResolvedValue(mockApplications);
    const wrapper = mountComponent();
    const { isNameNotTaken } = wrapper.vm;

    expect(await isNameNotTaken("existing_app")).toBe(errorMessages.NAME_ALREADY_EXIST);
    expect(await isNameNotTaken("new_app")).toBe(true);
  });

  it("should fetch access keys if isAccessKey is true", async () => {
    mockedAccessKeysService.getAccessKeys.mockResolvedValue(mockApplications);
    const wrapper = mountComponent({ isAccessKey: true });
    const { isNameNotTaken } = wrapper.vm;

    await isNameNotTaken("existing_app");

    expect(mockedAccessKeysService.getAccessKeys).toHaveBeenCalled();
  });

  it("should emit update:model-value event on input change", async () => {
    const wrapper = mountComponent();
    const inputField = wrapper.findComponent({ name: "RunaiNameValidation" });
    await inputField.vm.$emit("update:model-value", "new_name");

    expect(wrapper.emitted("update:model-value")).toBeTruthy();
    expect(wrapper.emitted("update:model-value")?.[0]).toEqual(["new_name"]);
  });

  it("should handle fetch error gracefully", async () => {
    console.warn = vi.fn();
    mockedServiceAccountsService.list.mockRejectedValue("");
    const wrapper = mountComponent();
    const { isNameNotTaken } = wrapper.vm;
    const result = await isNameNotTaken("new_app");
    expect(result).toBe(true);
  });
});
