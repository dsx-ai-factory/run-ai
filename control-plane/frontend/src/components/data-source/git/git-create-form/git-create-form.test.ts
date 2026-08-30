import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import GitCreateForm from "./git-create-form.vue";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { AssetKind, Scope, type Password } from "@/swagger-models/assets-service-client";
import type { IProjectSectionModel } from "@/models/project.model";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { useClusterStore } from "@/stores/cluster.store";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import type { IUIGitCreationRequest } from "@/models/data-source.model";
import { useAuthStore } from "@/stores/auth.store";

vi.mock("@/services/control-plane/asset.service/asset.service");
vi.mock("@/services/control-plane/rbac/org-tree.service/org-tree.service.ts");

const pinia = createTestingPinia({ createSpy: vi.fn });

const listResolvedValue: Array<Password> = [
  {
    meta: {
      name: "mocked-name-101",
      id: "mocked-id-101",
      scope: Scope.Project,
      kind: AssetKind.Password,
      createdBy: "",
      createdAt: "",
      updatedBy: "",
      updatedAt: "",
    },
    spec: { keyId: 1, existingSecretName: "mocked-secret", user: "mocked-user", password: "mocked-password" },
  },
];

setActivePinia(pinia);

const validGitModel: IUIGitCreationRequest = {
  meta: {
    name: "test-name",
    scope: Scope.Project,
    projectId: 25,
    clusterId: null,
  },
  spec: {
    repository: "https://test-repo.com",
    branch: "test-branch",
    revision: "test-revision",
    path: "/test-path",
    passwordAssetId: null,
  },
};

describe("git-create-form.vue", () => {
  let gitCreateFormCmp: VueWrapper<InstanceType<typeof GitCreateForm>>;
  const mockGetOrgTree = orgTreeService.getOrgTree as MockedFunction<typeof orgTreeService.getOrgTree>;
  mockGetOrgTree.mockResolvedValue(orgTreeServiceMock.orgTree);
  const mockGetNodePathById = orgTreeService.getNodePathById as MockedFunction<typeof orgTreeService.getNodePathById>;
  mockGetNodePathById.mockReturnValue("runai");

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;

  const spyListPasswordsByScope = vi.spyOn(dataSourceService, "listPasswordsByScope");
  const spyListGenericSecretsByScope = vi.spyOn(dataSourceService, "listGenericSecretsByScope");

  let clusterStore = useClusterStore(pinia);
  const authStore = useAuthStore(pinia);

  const emptyGitModel = dataSourceService.getEmptyGitModel();

  function mountComponent(model: IUIGitCreationRequest) {
    // Set up mocks before mounting to prevent network calls during component creation
    spyListPasswordsByScope.mockResolvedValue(listResolvedValue);
    spyListGenericSecretsByScope.mockResolvedValue([]);

    gitCreateFormCmp = mount(GitCreateForm, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        submitting: false,
        dataSourceModel: model,
        isScopeSelectable: true,
        allowedScopes: {},
      },
    });
    gitCreateFormCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  }

  beforeEach(() => {
    clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);
    authStore.$patch({ tenant: { id: 1001 } });
  });
  afterEach(() => {
    vi.clearAllMocks();
  });

  test("component is loading properly", () => {
    mountComponent(emptyGitModel);
    expect(gitCreateFormCmp.html()).toContain("git-create-form");
  });

  test("form validation should fail when the name is invalid", async () => {
    mountComponent(validGitModel);
    validGitModel.meta.name = "";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(false);

    validGitModel.meta.name = "test-name";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);
  });

  test("form validation should fail when the data mount is invalid", async () => {
    mountComponent(validGitModel);
    const dataMountCmp = gitCreateFormCmp.findComponent<DefineComponent>(".git-data-mount-section");

    validGitModel.spec.repository = "";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(false);

    validGitModel.spec.repository = "https://test-repo.com";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);

    validGitModel.spec.passwordAssetId = "";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    await dataMountCmp.setData({ isPrivate: true });
    expect(await gitCreateFormCmp.vm.validate()).toBe(false);

    validGitModel.spec.passwordAssetId = "mocked-id-101";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    await dataMountCmp.setData({ isPrivate: true });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);

    validGitModel.spec.path = "";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(false);

    validGitModel.spec.path = "/test-path";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);

    validGitModel.spec.branch = "";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);

    validGitModel.spec.branch = "test-branch";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);

    validGitModel.spec.revision = "";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);

    validGitModel.spec.revision = "test-revision";
    await gitCreateFormCmp.setData({ gitModel: validGitModel });
    expect(await gitCreateFormCmp.vm.validate()).toBe(true);
  });

  test("updates the scope in the model when a new scope is emitted by scope-section", async () => {
    mountComponent(emptyGitModel);
    const newScopeSectionModel: IProjectSectionModel = {
      scope: Scope.Project,
      projectId: 12,
    };
    const scopeSectionCmp = gitCreateFormCmp.findComponent<DefineComponent>("[aid='runai-scope-section']");
    scopeSectionCmp.vm.$emit("changed-scope", newScopeSectionModel);
    await gitCreateFormCmp.vm.$nextTick();
    expect(gitCreateFormCmp.vm.gitModel.meta.projectId).toBe(newScopeSectionModel.projectId);
    expect(gitCreateFormCmp.vm.gitModel.meta.scope).toBe(newScopeSectionModel.scope);
  });

  test("failing to submit the form when button is clicked and the form is invalid", async () => {
    // // selecting scope so the rest of the form apears
    const gitModel = dataSourceService.getEmptyGitModel();
    gitModel.meta.projectId = 1;
    gitModel.meta.scope = Scope.Project;

    mountComponent(gitModel);

    gitCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(gitCreateFormCmp.emitted("submit")).toBeFalsy();
  });

  test("successfully submitting the form when button is clicked and the form is valid", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: IUIGitCreationRequest = {
      meta: {
        name: "test-name",
        scope: Scope.Tenant,
        clusterId: null,
      },
      spec: {
        repository: "https://test-repo.com",
        branch: "test-branch",
        revision: "test-revision",
        path: "/test-path",
        passwordAssetId: "mocked-id-101",
      },
    };

    mountComponent(validModel);

    await gitCreateFormCmp.vm.$nextTick();
    await flushPromises();

    gitCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await gitCreateFormCmp.vm.$nextTick();
    await flushPromises();
    expect(gitCreateFormCmp.emitted().submit).toBeTruthy();
  });

  test("isExists is called when submitting the form", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: IUIGitCreationRequest = {
      meta: {
        name: "test-check-name",
        scope: Scope.Tenant,
        clusterId: null,
      },
      spec: {
        repository: "",
        branch: null,
        revision: null,
        path: "",
        passwordAssetId: "",
      },
    };
    await gitCreateFormCmp.setData({ gitModel: validModel });
    gitCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(mockIsExist).toHaveBeenLastCalledWith("git", validModel.meta.name, {
      scopeType: Scope.Tenant,
    });
  });
});
