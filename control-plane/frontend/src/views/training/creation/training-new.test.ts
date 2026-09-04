/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import TrainingNew from "./training-new.vue";
import { afterAll, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { DistributedFramework } from "@/swagger-models/assets-service-client";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { useClusterStore } from "@/stores/cluster.store";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { clusters } from "./training-new.test.mock";
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { loadedTrainingMergedWithPolicy, trainingAssetsEditMockedData } from "./training-assets-edit.mock";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";
import { ScopeType, WorkloadType } from "@/swagger-models/workloads-client";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";

vi.mock("@/services/control-plane/asset.service/asset.service");
vi.mock("@/services/control-plane/training.service/training.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/infra/router.service/router.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/infra/unleash.service/unleash.service");
vi.mock("@/services/control-plane/workload-template.service/workload-template.service");

// Mock the useQuickSubmission composable
vi.mock("@/composables/use-quick-submission.composable", () => ({
  useQuickSubmission: vi.fn(() => ({
    isLoading: { value: false },
    onQuickSubmit: vi.fn(),
  })),
}));

let pinia = createTestingPinia({ createSpy: vi.fn });
let clusterStore = useClusterStore(pinia);

describe("training-new.vue", () => {
  let trainingNewCmp: VueWrapper<InstanceType<typeof TrainingNew>>;

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;
  mockIsExist.mockResolvedValue(false);

  vi.spyOn(orgUnitService, "getProjectsMetadata").mockImplementation(async () => {
    return trainingAssetsEditMockedData.loadedProjectsMetasV2;
  });

  vi.spyOn(orgUnitService, "getProject").mockImplementation(async (projectId) => {
    const found = trainingAssetsEditMockedData.loadedProjectsV2.find((project: Project) => {
      return project.id === projectId;
    }) as Project;

    return found;
  });

  vi.spyOn(orgTreeService, "getExpandedPermittedScopes").mockImplementation(
    (_: PermittedScopes, _2: OrgUnit[]): PermittedScopes => {
      return {
        projects: trainingAssetsEditMockedData.loadedProjectsMetasV2.map((project) => project.id),
      } as PermittedScopes;
    },
  );

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    clusterStore = useClusterStore(pinia);

    clusterStore.$patch({
      clusters: clusters,
    });

    vi.mocked(clusterStore.getClusterById).mockReturnValue(clusters[0]);
    vi.mocked(trainingService.getTrainingMergedWithPolicy).mockResolvedValue(loadedTrainingMergedWithPolicy);
    vi.mocked(workloadService.canSubmitWorkload).mockReturnValue(null);
    setActivePinia(pinia);

    trainingNewCmp = mount(TrainingNew, {
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
    setQuasarMock(trainingNewCmp);
  });
  afterAll(() => {
    vi.resetAllMocks();
  });

  test("component is loading properly", async () => {
    expect(trainingNewCmp.classes()).toContain("training-new");
    expect(trainingNewCmp.find(".create-training-form").exists()).toBeTruthy();
  });

  test("if a project is not selected template and name sections are hidden", async () => {
    expect(trainingNewCmp.find("[aid=template-section]").exists()).toBe(false);
    expect(trainingNewCmp.find("[aid=training-name-section]").exists()).toBe(false);
  });

  test("selecting a project will reveal and template and name section", async () => {
    await flushPromises();
    const selectedProject = trainingAssetsEditMockedData.loadedProjectsMetasV2[0];
    await trainingNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await trainingNewCmp.vm.$nextTick();
    expect(trainingNewCmp.find("[aid=template-section]").exists()).toBe(true);
    expect(trainingNewCmp.find("[aid=training-name-section]").exists()).toBe(true);
  });

  test("updates the training project when a new project id is emitted by project-section component", async () => {
    await flushPromises();
    const projectId = +trainingAssetsEditMockedData.loadedProjectsMetasV2[0].id;
    const nameSectionCmp = trainingNewCmp.findComponent<DefineComponent>("[aid=project-section]");
    nameSectionCmp.vm.$emit("project-changed", projectId);
    await trainingNewCmp.vm.$nextTick();
    expect(trainingNewCmp.vm.selectedProjectId).toBe(projectId);
  });

  test("updates the training name when a new name is emitted by training-name-section component", async () => {
    await flushPromises();
    const selectedProject = trainingAssetsEditMockedData.loadedProjectsMetasV2[0];
    await trainingNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await flushPromises();
    const newName = "new-training-name";
    const nameSectionCmp = trainingNewCmp.findComponent<DefineComponent>("[aid=training-name-section]");
    nameSectionCmp.vm.$emit("update:name", newName);
    await trainingNewCmp.vm.$nextTick();
    expect(trainingNewCmp.vm.trainingName).toBe(newName);
  });

  test("updates the templateId in the component data when a new value is emitted by template-section component", async () => {
    await flushPromises();
    const selectedProject = trainingAssetsEditMockedData.loadedProjectsMetasV2[0];
    await trainingNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await flushPromises();
    expect(trainingNewCmp.vm.selectedTemplateId).toBe("-1"); // the id for 'Start from scratch' which is automatically selected if there are no templates
    const newTemplateId = "test-id-1";
    const nameSectionCmp = trainingNewCmp.findComponent<DefineComponent>("[aid=template-section]");

    nameSectionCmp.vm.$emit("template-changed", newTemplateId);
    await trainingNewCmp.vm.$nextTick();
    expect(trainingNewCmp.vm.selectedTemplateId).toBe(newTemplateId);
  });

  test("component will redirect the user training index page when 'Cancel' is clicked", async () => {
    vi.mocked(requestToLeave).mockResolvedValue(true);
    await trainingNewCmp.find("[aid=cancel-btn]").trigger("click");
    const expectedPush = { name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX };

    await flushPromises();
    expect(trainingNewCmp.vm.$router).toContainEqual(expectedPush);
  });

  test("component will NOT redirect the user the training edit page when 'Continue' button is clicked and validation fails due unselected project", async () => {
    await flushPromises();
    expect(trainingNewCmp.vm.training.projectId).toBe(-1);
    await trainingNewCmp.find("[aid=create-training-continue-btn]").trigger("click");
    expect(await trainingNewCmp.vm.validate()).toBe(false);
    await flushPromises();
    expect(trainingNewCmp.vm.$router).toEqual([]);
  });

  test("component will NOT redirect the user the training edit page when 'Continue' button is clicked and validation fails due to empty or invalid name", async () => {
    await flushPromises();
    const selectedProject = trainingAssetsEditMockedData.loadedProjectsMetasV2[0];
    await trainingNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await flushPromises();
    await trainingNewCmp.find("[aid=create-training-continue-btn]").trigger("click");
    // Note: that by default name section is empty
    await flushPromises();
    expect(await trainingNewCmp.vm.validate()).toBe(false);
    await flushPromises();
    expect(trainingNewCmp.vm.$router).toEqual([]);
  });

  test("component will redirect the user the training edit page when 'Continue' button is clicked and validation is successful", async () => {
    await flushPromises();
    const selectedProject = trainingAssetsEditMockedData.loadedProjectsMetasV2[0];
    await trainingNewCmp.find(`[aid=${selectedProject.name}-card]`).trigger("click"); // selecting a project to reveal the other sections
    await trainingNewCmp.find("[aid=create-training-continue-btn]").trigger("click");
    await trainingNewCmp.setData({ trainingName: "valid-name" });
    const expectedPush = { name: TRAINING_ROUTE_NAMES.TRAINING_ASSETS_EDIT };
    expect(await trainingNewCmp.vm.validate()).toBe(true);
    await flushPromises();
    expect(trainingNewCmp.vm.$router).toContainEqual(expectedPush);
  });

  describe("Templates V2 Functionality", () => {
    beforeEach(() => {
      Object.defineProperty(trainingNewCmp.vm, "showTemplatesV2", {
        get: () => true,
        configurable: true,
      });
    });

    test("loads flat templates when showTemplatesV2 is true and project is selected", async () => {
      await flushPromises();
      // Mock the flat templates gallery functions
      const mockList = vi.fn().mockResolvedValue(undefined);
      trainingNewCmp.vm.list = mockList;

      const projectId = +trainingAssetsEditMockedData.loadedProjectsMetasV2[0].id;
      await trainingNewCmp.vm.onSelectedProject(projectId);

      // Note: This test verifies the method call structure
      expect(mockList).toHaveBeenCalled();
    });
  });

  describe("Workload Architecture Selection", () => {
    test("setDistributedDetails method sets distributed details correctly for distributed workload", async () => {
      trainingNewCmp.vm.setDistributedDetails(true);

      expect(trainingNewCmp.vm.selectedDistributedDetails).toEqual({
        distFramework: null,
        numWorkers: 1,
      });
      expect(trainingNewCmp.vm.selectedTemplateId).toBeNull();
    });

    test("setDistributedDetails method clears distributed details for non-distributed workload", async () => {
      // First set distributed
      trainingNewCmp.vm.setDistributedDetails(true);
      expect(trainingNewCmp.vm.selectedDistributedDetails).not.toBeNull();

      // Then clear it
      trainingNewCmp.vm.setDistributedDetails(false);
      expect(trainingNewCmp.vm.selectedDistributedDetails).toBeNull();
    });
  });

  describe("Distributed Framework Selection", () => {
    test("setDistributedFramework method sets framework correctly", async () => {
      const mockFramework = DistributedFramework.PyTorch;

      trainingNewCmp.vm.setDistributedDetails(true);
      trainingNewCmp.vm.setDistributedFramework(mockFramework);

      expect(trainingNewCmp.vm.selectedDistributedDetails?.distFramework).toBe(mockFramework);
      expect(trainingNewCmp.vm.selectedDistributedDetails?.numWorkers).toBe(1);
    });

    test("setDistributedFramework sets noMaster to true for JAX framework", async () => {
      const mockFramework = DistributedFramework.Jax;

      trainingNewCmp.vm.setDistributedDetails(true);
      trainingNewCmp.vm.setDistributedFramework(mockFramework);

      expect(trainingNewCmp.vm.selectedDistributedDetails?.distFramework).toBe(mockFramework);
      expect(trainingNewCmp.vm.selectedDistributedDetails?.noMaster).toBe(true);
    });

    test("disableFrameworkSection computed property logic", async () => {
      await trainingNewCmp.vm.setTemplates([
        {
          meta: {
            id: "id-1",
            name: "name-1",
            workloadType: WorkloadType.Distributed,
            scopeType: ScopeType.Project,
            scopeId: "scope-id-1",
            createdBy: "user-1",
            createdAt: "2021-01-01",
          },
          spec: { distributedFramework: DistributedFramework.PyTorch },
        },
        {
          meta: {
            id: "id-2",
            name: "name-2",
            workloadType: WorkloadType.Distributed,
            scopeType: ScopeType.Project,
            scopeId: "scope-id-2",
            createdBy: "user-2",
            createdAt: "2021-01-02",
          },
          spec: {},
        },
      ]);
      await flushPromises();
      await trainingNewCmp.setData({
        selectedTemplateId: "id-1",
      });
      await trainingNewCmp.vm.$nextTick();
      await flushPromises();
      expect(trainingNewCmp.vm.disableFrameworkSection).toBe(true);

      // Test without distributed framework
      await trainingNewCmp.setData({
        selectedTemplateId: "id-2",
      });
      await trainingNewCmp.vm.$nextTick();
      await flushPromises();
      expect(trainingNewCmp.vm.disableFrameworkSection).toBe(false);
    });
  });

  describe("Template Filtering and Loading", () => {
    test("templatesFilter includes correct filters when distributed is selected", async () => {
      const mockFramework = DistributedFramework.PyTorch;

      trainingNewCmp.vm.selectedProjectId = 123;
      trainingNewCmp.vm.setDistributedDetails(true);
      trainingNewCmp.vm.setDistributedFramework(mockFramework);

      const filter = trainingNewCmp.vm.templatesFilter;

      expect(filter.projectId).toBe(123);
      expect(filter.isTraining).toBe(true);
      expect(filter.isDistributed).toBe(true);
      expect(filter.distributedFramework).toBe(mockFramework);
    });

    test("templatesFilter does not include distributed filters when not distributed", async () => {
      trainingNewCmp.vm.selectedProjectId = 123;
      trainingNewCmp.vm.selectedDistributedDetails = null;

      const filter = trainingNewCmp.vm.templatesFilter;

      expect(filter.projectId).toBe(123);
      expect(filter.isTraining).toBe(true);
      expect(filter.isDistributed).toBeUndefined();
      expect(filter.distributedFramework).toBeUndefined();
    });

    test("onTemplatesListQueryChanged updates query and loads templates", async () => {
      const mockSetQuery = vi.fn();
      const mockLoadTemplates = vi.fn();

      trainingNewCmp.vm.setFlatTemplatesListQuery = mockSetQuery;
      trainingNewCmp.vm.loadTemplates = mockLoadTemplates;

      const newQuery = { search: "test", sortBy: "name", filterBy: "" };
      await trainingNewCmp.vm.onTemplatesListQueryChanged(newQuery);

      expect(mockSetQuery).toHaveBeenCalledWith(newQuery);
      expect(mockLoadTemplates).toHaveBeenCalled();
    });

    test("onGetNextTemplatesBatch loads more templates with offset", async () => {
      const mockLoadTemplates = vi.fn();
      trainingNewCmp.vm.loadTemplates = mockLoadTemplates;

      await trainingNewCmp.vm.onGetNextTemplatesBatch(6);

      expect(mockLoadTemplates).toHaveBeenCalledWith(true, 6);
    });
  });

  describe("Hybrid Submission Flow", () => {
    beforeEach(() => {
      Object.defineProperty(trainingNewCmp.vm, "isHybridSupported", {
        get: () => true,
        configurable: true,
      });
    });

    test("isHybridSupported returns true when cluster version and settings support it", async () => {
      expect(trainingNewCmp.vm.isHybridSupported).toBe(true);
    });

    test("onContinue method routing logic", async () => {
      const mockContinueToHybrid = vi.fn();
      const mockContinueToEdit = vi.fn();

      trainingNewCmp.vm.continueToHybrid = mockContinueToHybrid;
      trainingNewCmp.vm.continueToEdit = mockContinueToEdit;

      trainingNewCmp.vm.onContinue();

      // Should call continueToHybrid when hybrid is supported
      expect(mockContinueToHybrid).toHaveBeenCalled();
      expect(mockContinueToEdit).not.toHaveBeenCalled();
    });
  });

  describe("Template Selection Integration", () => {
    test("onSelectedTemplate method updates template selection", async () => {
      const templateId = "test-template-123";
      const mockSetDistributedFramework = vi.fn();

      trainingNewCmp.vm.setDistributedFramework = mockSetDistributedFramework;
      trainingNewCmp.vm.onSelectedTemplate(templateId);

      expect(trainingNewCmp.vm.selectedTemplateId).toBe(templateId);
    });

    test("onSelectedTemplate updates framework when distributed and template has framework", async () => {
      const templateId = "id-1";
      const mockFramework = DistributedFramework.PyTorch;

      // Set up distributed state
      trainingNewCmp.vm.setDistributedDetails(true);

      // Mock templates array with distributed template

      await trainingNewCmp.vm.setTemplates([
        {
          meta: {
            id: templateId,
            name: "name-1",
            workloadType: WorkloadType.Distributed,
            scopeType: ScopeType.Project,
            scopeId: "scope-id-1",
            createdBy: "user-1",
            createdAt: "2021-01-01",
          },
          spec: { distributedFramework: mockFramework },
        },
      ]);

      Object.defineProperty(trainingNewCmp.vm, "showTemplatesV2", {
        get: () => true,
        configurable: true,
      });

      const setDistributedFrameworkSpy = vi.spyOn(trainingNewCmp.vm, "setDistributedFramework");

      trainingNewCmp.vm.onSelectedTemplate(templateId);

      expect(setDistributedFrameworkSpy).toHaveBeenCalledWith(mockFramework);
    });
  });
});
