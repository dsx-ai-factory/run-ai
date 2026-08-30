import { afterEach, describe, expect, it, vi } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";

import SubjectAddableMultiSelect from "./subject-addable-multi-select.vue";
import { SubjectType } from "@/swagger-models/authorization-client";
import type { ISelectOption } from "@/models/global.model";

import userService from "@/services/control-plane/user.service/user.service";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import { usePermissionStore } from "@/stores/permissions.store";

vi.mock("@/services/control-plane/user.service/user.service", () => ({
  default: {
    getUsersMinimal: vi.fn().mockResolvedValue([]),
    getUsers: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    getSsoGroupSubjectIds: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("@/services/control-plane/service-accounts.service/service-accounts.service", () => ({
  serviceAccountsService: {
    list: vi.fn().mockResolvedValue([]),
  },
}));

const USER_SUBJECT_TYPE: ISelectOption = { label: "User", value: SubjectType.User };
const SERVICE_ACCOUNT_SUBJECT_TYPE: ISelectOption = { label: "Service account", value: SubjectType.ServiceAccount };
const SSO_SUBJECT_TYPE: ISelectOption = { label: "SSO Group", value: SubjectType.Group };

const MOCK_USERS = [
  { id: "1", username: "alice@run.ai", firstName: "Alice", lastName: "Smith" },
  { id: "2", username: "bob@run.ai", firstName: "Bob", lastName: null },
];

const MOCK_SERVICE_ACCOUNTS = [
  {
    id: "sa-1",
    name: "sa-test-1",
    createdBy: "",
    createdAt: null,
    updatedAt: null,
    enabled: true,
    lastLogin: null,
    clientId: "c1",
  },
  {
    id: "sa-2",
    name: "sa-test-2",
    createdBy: "",
    createdAt: null,
    updatedAt: null,
    enabled: true,
    lastLogin: null,
    clientId: "c2",
  },
];

const MOCK_SSO_GROUPS = ["group-a", "group-b"];

describe("SubjectAddableMultiSelect", () => {
  let wrapper: VueWrapper<InstanceType<typeof SubjectAddableMultiSelect>>;

  const factory = (subjectType: ISelectOption, selectedSubjectsOptions: ISelectOption[] = []) => {
    wrapper = mount(SubjectAddableMultiSelect, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      },
      props: { subjectType, selectedSubjectsOptions },
    });
    wrapper.vm.$q.notify = vi.fn();
  };

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  describe("label", () => {
    it("shows 'User(s) email' label for User subject type", () => {
      factory(USER_SUBJECT_TYPE);
      expect(wrapper.vm.subjectSelectLabel).toBe("User(s) email");
    });

    it("shows 'Service account(s) name' label for ServiceAccount subject type", () => {
      factory(SERVICE_ACCOUNT_SUBJECT_TYPE);
      expect(wrapper.vm.subjectSelectLabel).toBe("Service account(s) name");
    });

    it("shows 'Group(s) name' label for SSO Group subject type", () => {
      factory(SSO_SUBJECT_TYPE);
      expect(wrapper.vm.subjectSelectLabel).toBe("Group(s) name");
    });
  });

  describe("initial fetch on mount", () => {
    it("fetches users with username=@@ filter on mount for User type", async () => {
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      expect(userService.getUsersMinimal).toHaveBeenCalledWith(
        undefined,
        expect.arrayContaining([expect.stringContaining("username=@@")]),
        undefined,
        10,
      );
    });

    it("fetches service accounts on mount for ServiceAccount type", async () => {
      vi.mocked(serviceAccountsService.list).mockResolvedValue(MOCK_SERVICE_ACCOUNTS);
      factory(SERVICE_ACCOUNT_SUBJECT_TYPE);
      await flushPromises();

      expect(serviceAccountsService.list).toHaveBeenCalled();
    });

    it("fetches SSO groups on mount for Group type", async () => {
      vi.mocked(accessRuleService.getSsoGroupSubjectIds).mockResolvedValue(MOCK_SSO_GROUPS);
      factory(SSO_SUBJECT_TYPE);
      await flushPromises();

      expect(accessRuleService.getSsoGroupSubjectIds).toHaveBeenCalledWith("");
    });

    it("populates combinedEmailOptions from initial user fetch", async () => {
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      expect(wrapper.vm.combinedEmailOptions).toHaveLength(2);
      expect(wrapper.vm.combinedEmailOptions[0].value).toBe("alice@run.ai");
    });

    it("limits initial results to 10", async () => {
      const manyUsers = Array.from({ length: 20 }, (_, i) => ({
        id: `${i}`,
        username: `user${i}@run.ai`,
        firstName: null,
        lastName: null,
      }));
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(manyUsers);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      expect(wrapper.vm.combinedEmailOptions).toHaveLength(10);
    });

    it("limits initial service accounts to 10", async () => {
      const manySAs = Array.from({ length: 20 }, (_, i) => ({
        id: `sa-${i}`,
        name: `sa-test-${i}`,
        createdBy: "",
        createdAt: null,
        updatedAt: null,
        enabled: true,
        lastLogin: null,
        clientId: `c${i}`,
      }));
      vi.mocked(serviceAccountsService.list).mockResolvedValue(manySAs);
      factory(SERVICE_ACCOUNT_SUBJECT_TYPE);
      await flushPromises();

      expect(wrapper.vm.combinedEmailOptions).toHaveLength(10);
    });
  });

  describe("search", () => {
    it("calls getUsersMinimal with search term when user types", async () => {
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();
      vi.mocked(userService.getUsersMinimal).mockClear();

      await wrapper.vm.fetchUsers("alice");
      await flushPromises();

      expect(userService.getUsersMinimal).toHaveBeenCalledWith("alice");
    });

    it("does not call getUsersMinimal for 1-char queries (users)", async () => {
      factory(USER_SUBJECT_TYPE);
      await flushPromises();
      vi.mocked(userService.getUsersMinimal).mockClear();

      await wrapper.vm.fetchUsers("a");

      expect(userService.getUsersMinimal).not.toHaveBeenCalled();
    });

    it("does not add limit param when search has a query string", async () => {
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();
      vi.mocked(userService.getUsersMinimal).mockClear();

      await wrapper.vm.fetchUsers("alice");

      expect(userService.getUsersMinimal).toHaveBeenCalledWith("alice");
      expect(userService.getUsersMinimal).not.toHaveBeenCalledWith(
        expect.anything(),
        expect.anything(),
        expect.anything(),
        expect.any(Number),
      );
    });

    it("filters service accounts client-side by search query without re-fetching", async () => {
      vi.mocked(serviceAccountsService.list).mockResolvedValue(MOCK_SERVICE_ACCOUNTS);
      factory(SERVICE_ACCOUNT_SUBJECT_TYPE);
      await flushPromises();

      await wrapper.vm.fetchUsers("sa-test-1");
      await flushPromises();

      expect(wrapper.vm.combinedEmailOptions).toHaveLength(1);
      expect(wrapper.vm.combinedEmailOptions[0].value).toBe("sa-test-1");
      // list() should only have been called once (cached after initial fetch)
      expect(serviceAccountsService.list).toHaveBeenCalledTimes(1);
    });

    it("replaces fetchedEmailOptions on each new search (no accumulation)", async () => {
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      vi.mocked(userService.getUsersMinimal).mockResolvedValue([MOCK_USERS[0]]);
      await wrapper.vm.fetchUsers("alice");
      await flushPromises();

      expect(wrapper.vm.combinedEmailOptions).toHaveLength(1);
    });
  });

  describe("subject type change", () => {
    it("clears options and selected subjects when subject type changes", async () => {
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      await wrapper.setProps({ subjectType: SERVICE_ACCOUNT_SUBJECT_TYPE });
      await flushPromises();

      expect(wrapper.emitted("update:selected-subjects")?.[0]).toEqual([[]]);
    });

    it("fetches service accounts after switching to ServiceAccount type", async () => {
      vi.mocked(serviceAccountsService.list).mockResolvedValue(MOCK_SERVICE_ACCOUNTS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      await wrapper.setProps({ subjectType: SERVICE_ACCOUNT_SUBJECT_TYPE });
      await flushPromises();

      expect(serviceAccountsService.list).toHaveBeenCalled();
      expect(wrapper.vm.combinedEmailOptions[0].value).toBe("sa-test-1");
    });

    it("calls correct endpoint after switching from ServiceAccount to User", async () => {
      factory(SERVICE_ACCOUNT_SUBJECT_TYPE);
      await flushPromises();
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);

      await wrapper.setProps({ subjectType: USER_SUBJECT_TYPE });
      await flushPromises();

      expect(userService.getUsersMinimal).toHaveBeenCalled();
      expect(serviceAccountsService.list).toHaveBeenCalledTimes(1); // only from initial mount
    });
  });

  describe("manual options (addEmail)", () => {
    it("preserves manual options after a fetch", async () => {
      vi.mocked(userService.getUsersMinimal).mockResolvedValue(MOCK_USERS);
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      const doneCallback = vi.fn();
      wrapper.vm.addEmail("manual@run.ai", doneCallback);
      await flushPromises();

      await wrapper.vm.fetchUsers("al");
      await flushPromises();

      const values = wrapper.vm.combinedEmailOptions.map((o) => o.value);
      expect(values).toContain("manual@run.ai");
    });
  });

  describe("error handling", () => {
    it("shows notification on user fetch error", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(userService.getUsersMinimal).mockRejectedValue(new Error("network error"));
      factory(USER_SUBJECT_TYPE);
      await flushPromises();

      expect(wrapper.vm.$q.notify).toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("shows notification on service account fetch error", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(serviceAccountsService.list).mockRejectedValue(new Error("network error"));
      factory(SERVICE_ACCOUNT_SUBJECT_TYPE);
      await flushPromises();

      expect(wrapper.vm.$q.notify).toHaveBeenCalled();
      errorSpy.mockRestore();
    });
  });

  describe("loading state (anti-flicker fix)", () => {
    it("sets loadingEmails to true immediately when onSearchInput fires (before debounce)", () => {
      // SSO always returns canSearchSubjects=true regardless of permission store mocks
      factory(SSO_SUBJECT_TYPE);

      const event = { target: { value: "group-a" } } as unknown as Event;
      wrapper.vm.onSearchInput(event);

      // synchronous check — loadingEmails must be true before any await / debounce fires
      expect(wrapper.vm.loadingEmails).toBe(true);
    });

    it("sets loadingEmails to true immediately even for empty query (initial open)", () => {
      factory(SSO_SUBJECT_TYPE);

      const event = { target: { value: "" } } as unknown as Event;
      wrapper.vm.onSearchInput(event);

      expect(wrapper.vm.loadingEmails).toBe(true);
    });

    it("sets loadingEmails immediately for SSO regardless of query length", async () => {
      factory(SSO_SUBJECT_TYPE);
      await flushPromises();

      wrapper.vm.loadingEmails = false;
      const event = { target: { value: "g" } } as unknown as Event;
      wrapper.vm.onSearchInput(event);

      expect(wrapper.vm.loadingEmails).toBe(true);
    });

    it("clears loadingEmails for 1-char User queries even if a previous fetch left it true", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(userService.getUsersMinimal).mockResolvedValue([]);
      // hasPermission must return true BEFORE mount so canReadUsers/canSearchSubjects=true
      const pinia = createTestingPinia({ createSpy: vi.fn });
      const permStore = usePermissionStore(pinia);
      vi.mocked(permStore.hasPermission).mockReturnValue(true);
      wrapper = mount(SubjectAddableMultiSelect, {
        global: { plugins: [Quasar, pinia] },
        props: { subjectType: USER_SUBJECT_TYPE, selectedSubjectsOptions: [] },
      });
      wrapper.vm.$q.notify = vi.fn();
      await flushPromises();
      wrapper.vm.loadingEmails = true; // simulate an in-flight fetch

      const event = { target: { value: "a" } } as unknown as Event;
      wrapper.vm.onSearchInput(event);

      expect(wrapper.vm.loadingEmails).toBe(false);
      errorSpy.mockRestore();
    });
  });
});
