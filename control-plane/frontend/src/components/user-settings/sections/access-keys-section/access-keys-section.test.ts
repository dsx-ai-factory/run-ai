import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { accessKeysService } from "@/services/control-plane/access-keys.service/access-keys.service";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { AccessKeysSection } from "./";
import type { AccessKey } from "@/swagger-models/identity-manager-client";

vi.mock("@/services/control-plane/access-keys.service/access-keys.service");

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const mockedAccessKeysService = vi.mocked(accessKeysService);

describe("AccessKeysSection", () => {
  let wrapper: VueWrapper<InstanceType<typeof AccessKeysSection>>;

  const mockApplications: AccessKey[] = [
    {
      id: "1",
      name: "App 1",
      createdAt: "2021-01-01T00:00:00Z",
      lastLogin: "2021-01-01T00:00:00Z",
      clientId: "test_client_id",
    },
    {
      id: "2",
      name: "App 2",
      createdAt: "2021-01-01T00:00:00Z",
      lastLogin: "2021-01-01T00:00:00Z",
      clientId: "test_client_id",
    },
    { id: "3", name: "App 3", createdAt: "2021-01-01T00:00:00Z", lastLogin: "", clientId: "test_client_id" },
  ];

  beforeEach(async () => {
    mockedAccessKeysService.getAccessKeys.mockResolvedValue(mockApplications);
    wrapper = mount(AccessKeysSection, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
          $router: [],
        },
      },
    });
    wrapper.vm.$q.notify = vi.fn();
    await flushPromises();
  });

  afterEach(() => {
    vi.resetAllMocks(); // Reset all mocks after each test
    wrapper.unmount(); // Clean up the mounted component
  });

  it("renders the correct subheader when access keys are present", () => {
    expect(wrapper.find(".access-keys-container").exists()).toBe(true);
    const subheader = wrapper.vm.accessKeysSubheader;
    expect(subheader).toBe("3 access keys");
  });

  it("disables the 'Add Access Key' button when 20 access keys exist", async () => {
    await wrapper.setData({
      accessKeys: Array.from({ length: 20 }, (_, index) => ({
        id: `x${index}-${Math.random()}`, // Combines index and random string
        name: "Test",
        createdAt: "2021-01-01T00:00:00Z",
        lastLogin: "2021-01-01T00:00:00Z",
      })),
    });
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.disableAddAccessKey).toBe(true);
    const button = wrapper.find('[aid="aidAddApplication"]');
    expect(button.attributes("disabled")).toBeDefined();
  });

  it("'Add Access Key' button enabled when 19 access keys exist", async () => {
    await wrapper.setData({
      accessKeys: Array.from({ length: 19 }, (_, index) => ({
        id: `x${index}-${Math.random()}`, // Combines index and random string
        name: "Test",
        createdAt: "2021-01-01T00:00:00Z",
        lastLogin: "2021-01-01T00:00:00Z",
      })),
    });
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.disableAddAccessKey).toBe(false);
    const button = wrapper.find('[aid="aidAddApplication"]');
    expect(button.attributes("disabled")).toBeUndefined();
  });

  it("opens the creation modal when the 'Add Access Key' button is clicked", async () => {
    const button = wrapper.find('[aid="aidAddApplication"]');
    await button.trigger("click");
    expect(wrapper.vm.isCreationModalOpen).toBe(true);
  });

  it("opens the delete modal when a delete event is triggered", async () => {
    const firstAppBox = wrapper.findComponent({ name: "access-key-box" });
    await firstAppBox.vm.$emit("remove-access-key", mockApplications[0]);
    expect(wrapper.vm.isDeleteModalOpen).toBe(true);
    expect(wrapper.vm.selectedAccessKey).toEqual(mockApplications[0]);
  });

  it("fetches access keys on component creation", async () => {
    expect(mockedAccessKeysService.getAccessKeys).toHaveBeenCalledOnce();
    expect(wrapper.vm.accessKeys).toEqual(mockApplications);
  });

  it("removes an access key when confirmed in the delete modal", async () => {
    await wrapper.setData({ selectedAccessKey: mockApplications[0], isDeleteModalOpen: true });
    mockedAccessKeysService.removeAccessKey.mockResolvedValue(undefined);
    await wrapper.vm.removeAccessKey();
    expect(mockedAccessKeysService.removeAccessKey).toHaveBeenCalledWith(mockApplications[0].id);
    expect(mockedAccessKeysService.getAccessKeys).toHaveBeenCalledTimes(2); // First during creation, then after removal
    expect(wrapper.vm.isDeleteModalOpen).toBe(false);
  });

  it("handles API errors gracefully when removing an access key", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    await wrapper.setData({ selectedAccessKey: mockApplications[0], isDeleteModalOpen: true });
    mockedAccessKeysService.removeAccessKey.mockRejectedValue(new Error("API error"));
    await wrapper.vm.removeAccessKey();
    expect(mockedAccessKeysService.removeAccessKey).toHaveBeenCalledWith(mockApplications[0].id);
    expect(wrapper.vm.deleting).toBe(false);
    expect(wrapper.vm.isDeleteModalOpen).toBe(false);
    errorSpy.mockRestore();
  });
});
