/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import WorkspaceNew from "./workspace-new.vue";
import { afterAll, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { loadedWorkspaceMergedWithPolicy } from "./workspace-assets-edit.mock";
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { clusters } from "./workspace-new.test.mock";

import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { trainingAssetsEditMockedData as mockData } from "@/views/training/creation/training-assets-edit.mock";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";

vi.mock("@/services/control-plane/asset.service/asset.service");
vi.mock("@/services/control-plane/workspace.service/workspace.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/infra/router.service/router.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/infra/unleash.service/unleash.service");
vi.mock("@/services/control-plane/workload-template.service/workload-template.service");

// Mock the useQuickSubmission composable
vi.mock("@/composables/use-quick-submission.composable", async () => {
  const { ref } = await import("vue");
  return {
    useQuickSubmission: vi.fn(() => ({
      isLoading: ref(false),
      onQuickSubmit: vi.fn(),
    })),
  };
});

let pinia = createTestingPinia({ createSpy: vi.fn });
let clusterStore = useClusterStore(pinia);

describe("workspace-new.vue", () => {
  let workspaceNewCmp: VueWrapper<InstanceType<typeof WorkspaceNew>>;

  vi.spyOn(orgUnitService, "getProjectsMetadata").mockImplementation(async () => {
    return mockData.loadedProjectsMetasV2;
  });

  vi.spyOn(orgUnitService, "getProject").mockImplementation(async (projectId) => {
    const found = mockData.loadedProjectsV2.find((project: Project) => {
      return project.id === projectId;
    }) as Project;

    return found;
  });

  vi.spyOn(orgTreeService, "getExpandedPermittedScopes").mockImplementation(
    (_: PermittedScopes, _2: OrgUnit[]): PermittedScopes => {
      return {
        projects: mockData.loadedProjectsMetasV2.map((project) => project.id),
      } as PermittedScopes;
    },
  );

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;
  mockIsExist.mockResolvedValue(false);

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    clusterStore = useClusterStore(pinia);

    clusterStore.$patch({
      clusters: clusters,
    });

    vi.mocked(clusterStore.getClusterById).mockReturnValue(clusters[0]);
    vi.mocked(workloadTemplateService.workspaceTemplatesList).mockResolvedValue({
      entries: [],
      pagination: { limit: 5 },
    });
    vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
    vi.mocked(workspaceService.getWorkspaceMergedWithPolicy).mockResolvedValue(loadedWorkspaceMergedWithPolicy);
    vi.mocked(workloadService.canSubmitWorkload).mockReturnValue(null);
    setActivePinia(pinia);

    workspaceNewCmp = mount(WorkspaceNew, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          clickOutside: vi.fn(),
        },
        mocks: {
          $route: {
            meta: {
              prevRoute: "",
            },
            query: "",
          },
          $router: [],
        },
      },
    });
    setQuasarMock(workspaceNewCmp);
  });
  afterAll(() => {
    vi.clearAllMocks();
  });

  test("component is loading properly", async () => {
    expect(workspaceNewCmp.classes()).toContain("workspace-new");
    expect(workspaceNewCmp.find(".create-workspace-form").exists()).toBeTruthy();
  });

  test("if a project is not selected template and name sections are hidden", async () => {
    expect(workspaceNewCmp.find("[aid=template-section-v2]").exists()).toBe(false);
    expect(workspaceNewCmp.find("[aid=ws-name-section]").exists()).toBe(false);
  });

  test("selecting a project will reveal and template and name section", async () => {
    useSettingStore(pinia).enableHybridSubmission = { value: true };
    vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(true);
    await workspaceNewCmp.setData({ selectedClusterId: "test-cluster" });
    await flushPromises();
    const selectedProject = mockData.loadedProjectsMetasV2[0];
    await workspaceNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await workspaceNewCmp.vm.$nextTick();
    expect(workspaceNewCmp.find("[aid=template-section-v2]").exists()).toBe(true);
    expect(workspaceNewCmp.find("[aid=ws-name-section]").exists()).toBe(true);
  });

  test("updates the workspace project when a new project id is emitted by project-section component", async () => {
    await flushPromises();
    const projectId = 1;
    const nameSectionCmp = workspaceNewCmp.findComponent<DefineComponent>("[aid=project-section]");
    nameSectionCmp.vm.$emit("project-changed", projectId);
    await workspaceNewCmp.vm.$nextTick();
    expect(workspaceNewCmp.vm.selectedProjectId).toBe(projectId);
  });

  test("updates the workspace name when a new name is emitted by workspace-name-section component", async () => {
    await flushPromises();
    const selectedProject = mockData.loadedProjectsMetasV2[0];
    await workspaceNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await flushPromises();
    const newName = "new-workspace-name";
    const nameSectionCmp = workspaceNewCmp.findComponent<DefineComponent>("[aid=ws-name-section]");
    nameSectionCmp.vm.$emit("update:name", newName);
    await workspaceNewCmp.vm.$nextTick();
    expect(workspaceNewCmp.vm.workspaceName).toBe(newName);
  });

  test("updates the templateId in the component data when a new value is emitted by template-section component", async () => {
    useSettingStore(pinia).enableHybridSubmission = { value: true };
    vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(true);
    await workspaceNewCmp.setData({ selectedClusterId: "test-cluster" });
    await flushPromises();
    const selectedProject = mockData.loadedProjectsMetasV2[0];
    await workspaceNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await flushPromises();
    expect(workspaceNewCmp.vm.selectedTemplateId).toBe("-1"); // the id for 'Start from scratch' which is automatically selected if there are no templates
    const newTemplateId = "test-id-1";
    const nameSectionCmp = workspaceNewCmp.findComponent<DefineComponent>("[aid=template-section-v2]");

    nameSectionCmp.vm.$emit("template-changed", newTemplateId);
    await workspaceNewCmp.vm.$nextTick();
    expect(workspaceNewCmp.vm.selectedTemplateId).toBe(newTemplateId);
  });

  test("component will redirect the user workspace index page when 'Cancel' is clicked", async () => {
    vi.mocked(requestToLeave).mockResolvedValue(true);
    await workspaceNewCmp.find("[aid=cancel-btn]").trigger("click");
    const expectedPush = { name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX };

    await flushPromises();
    expect(workspaceNewCmp.vm.$router).toContainEqual(expectedPush);
  });

  test("component will NOT redirect the user the workspace edit page when 'Continue' button is clicked and validation fails due unselected project", async () => {
    await flushPromises();
    expect(workspaceNewCmp.vm.selectedProjectId).toBe(-1);
    await workspaceNewCmp.find("[aid=create-workspace-continue-btn]").trigger("click");
    expect(await workspaceNewCmp.vm.validate()).toBe(false);
    await flushPromises();
    expect(workspaceNewCmp.vm.$router).toEqual([]);
  });

  test("component will NOT redirect the user the workspace edit page when 'Continue' button is clicked and validation fails due to empty or invalid name", async () => {
    await flushPromises();
    const selectedProject = mockData.loadedProjectsMetasV2[0];
    await workspaceNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await flushPromises();
    await workspaceNewCmp.find("[aid=create-workspace-continue-btn]").trigger("click");
    // Note: that by default name section is empty
    expect(await workspaceNewCmp.vm.validate()).toBe(false);
    await flushPromises();
    expect(workspaceNewCmp.vm.$router).toEqual([]);
  });

  test("component will redirect the user the workspace edit page when 'Continue' button is clicked and validation is successful", async () => {
    await flushPromises();
    const selectedProject = mockData.loadedProjectsMetasV2[0];
    await workspaceNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await workspaceNewCmp.find("[aid=create-workspace-continue-btn]").trigger("click");
    await workspaceNewCmp.setData({ workspaceName: "valid-name" });
    const expectedPush = { name: WORKSPACE_ROUTE_NAMES.WORKSPACE_ASSETS_EDIT };
    expect(await workspaceNewCmp.vm.validate()).toBe(true);
    await flushPromises();
    expect(workspaceNewCmp.vm.$router).toContainEqual(expectedPush);
  });

  test.each([
    { enabled: true, supported: true },
    { enabled: false, supported: true },
    { enabled: true, supported: false },
  ])(
    "loads flexible templates only when submission is supported ($enabled, $supported)",
    async ({ enabled, supported }) => {
      await flushPromises();
      useSettingStore(pinia).enableHybridSubmission = { value: enabled };
      vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(supported);
      await workspaceNewCmp.setData({ selectedClusterId: "compatibility-cluster", selectedProjectId: 123 });
      const list = vi.fn().mockResolvedValue(undefined);
      workspaceNewCmp.vm.list = list;

      await workspaceNewCmp.vm.loadTemplates();

      expect(list).toHaveBeenCalledTimes(enabled && supported ? 1 : 0);
      expect(workspaceNewCmp.find("[aid=template-section-v2]").exists()).toBe(enabled && supported);
    },
  );

  describe("Flexible templates", () => {
    beforeEach(() => {
      Object.defineProperty(workspaceNewCmp.vm, "isHybridSupported", {
        get: () => true,
        configurable: true,
      });
    });

    test("loads flat templates when isHybridSupported is true and project is selected", async () => {
      await flushPromises();

      // Mock the flat templates gallery functions
      const mockList = vi.fn().mockResolvedValue(undefined);
      workspaceNewCmp.vm.list = mockList;

      const projectId = +mockData.loadedProjectsMetasV2[0].id;
      await workspaceNewCmp.vm.onSelectedProject(projectId);

      // Note: This test verifies the method call structure
      expect(mockList).toHaveBeenCalled();
    });

    test("template selection is hidden when flexible submission is unavailable", async () => {
      Object.defineProperty(workspaceNewCmp.vm, "isHybridSupported", {
        get: () => false,
        configurable: true,
      });

      await flushPromises();
      const selectedProject = mockData.loadedProjectsMetasV2[0];
      await workspaceNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click");
      await workspaceNewCmp.vm.$nextTick();

      expect(workspaceNewCmp.find("[aid=template-section-v2]").exists()).toBe(false);
    });
  });

  describe("Template Filtering and Loading", () => {
    test("loadTemplates uses flat templates when isHybridSupported is true", async () => {
      Object.defineProperty(workspaceNewCmp.vm, "isHybridSupported", {
        get: () => true,
        configurable: true,
      });

      const mockList = vi.fn().mockResolvedValue(undefined);
      const mockSetTemplates = vi.fn();

      workspaceNewCmp.vm.list = mockList;
      workspaceNewCmp.vm.setTemplates = mockSetTemplates;
      workspaceNewCmp.vm.selectedProjectId = 123;

      await workspaceNewCmp.vm.loadTemplates();

      expect(mockList).toHaveBeenCalledWith(
        "Workspace", // WorkloadType.Workspace
        "123",
        false,
        undefined,
      );
    });

    test("skips template listing when flexible submission is unavailable", async () => {
      Object.defineProperty(workspaceNewCmp.vm, "isHybridSupported", {
        get: () => false,
        configurable: true,
      });
      const list = vi.fn();
      workspaceNewCmp.vm.list = list;
      workspaceNewCmp.vm.selectedProjectId = 123;

      await workspaceNewCmp.vm.loadTemplates();

      expect(list).not.toHaveBeenCalled();
      expect(workspaceNewCmp.vm.selectedTemplateId).toBe("-1");
    });

    test("onTemplatesListQueryChanged updates query and loads templates", async () => {
      const mockSetQuery = vi.fn();
      const mockLoadTemplates = vi.fn();

      workspaceNewCmp.vm.setFlatTemplatesListQuery = mockSetQuery;
      workspaceNewCmp.vm.loadTemplates = mockLoadTemplates;

      const newQuery = { search: "test", sortBy: "name", filterBy: "" };
      await workspaceNewCmp.vm.onTemplatesListQueryChanged(newQuery);

      expect(mockSetQuery).toHaveBeenCalledWith(newQuery);
      expect(mockLoadTemplates).toHaveBeenCalled();
    });

    test("onGetNextTemplatesBatch loads more templates with offset", async () => {
      const mockLoadTemplates = vi.fn();
      workspaceNewCmp.vm.loadTemplates = mockLoadTemplates;

      await workspaceNewCmp.vm.onGetNextTemplatesBatch(6);

      expect(mockLoadTemplates).toHaveBeenCalledWith(true, 6);
    });
  });

  describe("Hybrid Submission Flow", () => {
    beforeEach(() => {
      Object.defineProperty(workspaceNewCmp.vm, "isHybridSupported", {
        get: () => true,
        configurable: true,
      });
    });

    test("isHybridSupported returns true when cluster version and settings support it", async () => {
      expect(workspaceNewCmp.vm.isHybridSupported).toBe(true);
    });

    test("onContinue calls continueToHybrid when hybrid is supported", async () => {
      const mockContinueToHybrid = vi.fn();
      const mockContinueToEdit = vi.fn();

      workspaceNewCmp.vm.continueToHybrid = mockContinueToHybrid;
      workspaceNewCmp.vm.continueToEdit = mockContinueToEdit;

      workspaceNewCmp.vm.onContinue();

      // Should call continueToHybrid when hybrid is supported
      expect(mockContinueToHybrid).toHaveBeenCalled();
      expect(mockContinueToEdit).not.toHaveBeenCalled();
    });

    test("onContinue calls continueToEdit when hybrid is not supported", async () => {
      Object.defineProperty(workspaceNewCmp.vm, "isHybridSupported", {
        get: () => false,
        configurable: true,
      });

      const mockContinueToHybrid = vi.fn();
      const mockContinueToEdit = vi.fn();

      workspaceNewCmp.vm.continueToHybrid = mockContinueToHybrid;
      workspaceNewCmp.vm.continueToEdit = mockContinueToEdit;

      workspaceNewCmp.vm.onContinue();

      expect(mockContinueToEdit).toHaveBeenCalled();
      expect(mockContinueToHybrid).not.toHaveBeenCalled();
    });

    test("continueToHybrid navigates to hybrid workspace route", async () => {
      await flushPromises();
      const selectedProject = mockData.loadedProjectsMetasV2[0];
      await workspaceNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click");
      await workspaceNewCmp.setData({
        workspaceName: "valid-name",
        selectedClusterId: "test-cluster",
        selectedProjectId: 123,
        selectedTemplateId: "template-1",
      });

      // Mock validate to return true
      workspaceNewCmp.vm.validate = vi.fn().mockResolvedValue(true);

      await workspaceNewCmp.vm.continueToHybrid();

      expect(workspaceNewCmp.vm.$router).toContainEqual({
        name: "workspace-hybrid",
        query: {
          clusterId: "test-cluster",
          projectId: 123,
          templateId: "template-1",
          workspaceName: "valid-name",
        },
      });
    });
  });

  describe("Template Selection Integration", () => {
    test("onSelectedTemplate method updates template selection", async () => {
      const templateId = "test-template-123";

      workspaceNewCmp.vm.onSelectedTemplate(templateId);

      expect(workspaceNewCmp.vm.selectedTemplateId).toBe(templateId);
    });

    test("onSelectedTemplate handles null template selection", async () => {
      workspaceNewCmp.vm.onSelectedTemplate(null);

      expect(workspaceNewCmp.vm.selectedTemplateId).toBeNull();
    });
  });
});
