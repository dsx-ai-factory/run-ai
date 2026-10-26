/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import WorkloadYamlNew from "./workload-yaml-new.vue";

// stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { useWorkloadStore } from "@/stores/workload.store";

// services
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { permissionService } from "@/services/authorization/permission.service/permission.service";

// models
import { Preemptibility } from "@/models/workload.model";
import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";
import type { WorkloadV2 } from "@/swagger-models/workloads-manager-client";

// Mock services
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/authorization/permission.service/permission.service");

// Mock vue-router's useRoute
vi.mock("vue-router", async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...(actual as object),
    useRoute: () => ({
      query: { clusterId: "cluster-1" },
      params: {},
    }),
  };
});

const { mockLoadWorkloadPriorities } = vi.hoisted(() => ({
  mockLoadWorkloadPriorities: vi.fn().mockResolvedValue(undefined),
}));
vi.mock("@/composables/use-workload-priorities.composable", async () => {
  const { ref, computed } = await import("vue");
  return {
    useWorkloadPriorities: () => ({
      workloadPriorities: ref([]),
      loadWorkloadPriorities: mockLoadWorkloadPriorities,
      supportPriorityClass: computed(() => true),
    }),
  };
});

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const mockRouter = {
  push: vi.fn(),
  back: vi.fn(),
};

const mockNotify = vi.fn();

const mockClusters = [
  { uuid: "cluster-1", tenantId: 1, name: "Test Cluster", createdAt: "2024-01-01", version: "2.23" },
];

const mockProjects = [
  {
    id: "1",
    name: "project-1",
    namespace: "runai-project-1",
    phase: "Ready",
    clusterId: "cluster-1",
    createdAt: "2024-01-01",
    allocated: 0,
    deserved: 0,
  },
  {
    id: "2",
    name: "project-2",
    namespace: "runai-project-2",
    phase: "Ready",
    clusterId: "cluster-1",
    createdAt: "2024-01-01",
    allocated: 0,
    deserved: 0,
  },
] as ProjectMetadata[];

describe("workload-yaml-new.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadYamlNew>>;
  const appStore = useAppStore(pinia);
  const clusterStore = useClusterStore(pinia);
  const authStore = useAuthStore(pinia);

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(orgUnitService.getProjectsMetadata).mockResolvedValue(mockProjects);
    vi.mocked(permissionService.getPermittedScopesForAction).mockResolvedValue({} as never);
    vi.mocked(workloadService.canSubmitWorkload).mockReturnValue(null);

    clusterStore.$patch({ clusters: mockClusters as never });
    vi.spyOn(clusterStore, "getClusterById").mockImplementation(
      (id: string) => mockClusters.find((c) => c.uuid === id) as never,
    );
    appStore.$patch({ pageLoading: false });
    authStore.$patch({ orgUnits: [] });
  });

  function createWrapper() {
    return mount(WorkloadYamlNew, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $router: mockRouter,
          $route: {
            query: { clusterId: "cluster-1" },
            params: {},
          },
          $q: {
            notify: mockNotify,
          },
        },
        stubs: {
          RunaiClusterSection: {
            template: "<div aid='cluster-section'>Cluster Section</div>",
            props: ["entityName", "clusterId", "disabledClusterOptions"],
            emits: ["update:cluster-id"],
          },
          ProjectSection: {
            template: "<div aid='project-section'>Project Section</div>",
            props: ["loading", "selectedProjectId", "projects", "clusterId", "showOverrideHint", "disableClosing"],
            emits: ["project-changed"],
          },
          GeneralSection: {
            name: "GeneralSection",
            template: "<div aid='general-section'>General Section</div>",
            props: [
              "workloadFormType",
              "singlePodWorkload",
              "generalModel",
              "supportPreemptibility",
              "workloadPriorityData",
              "showAnnotationsAndLabels",
              "showAutoDeletionSection",
              "overrideHints",
            ],
            emits: ["general-model-changed"],
          },
          YamlWorkloadSection: {
            template: "<div aid='yaml-workload-section'>YAML Section</div>",
            props: ["yaml", "clusterId", "selectedFormValues"],
            emits: ["yaml-input", "validation-changed", "yaml-loaded", "dryrun-success"],
          },
          WorkloadFormWrapper: {
            template: "<div><slot></slot></div>",
            props: ["formState", "submitting", "submitBtnText", "cancelBtnText"],
            emits: ["leave-page", "canceled", "submit"],
          },
          RunaiExpansionWrapper: {
            template: "<div><slot></slot></div>",
          },
        },
      },
    });
  }

  describe("Component Initialization", () => {
    it("should render the component", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.classes()).toContain("workload-yaml-new");
    });

    it("should initialize with empty YAML content", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.yamlContent).toBe("");
      expect(wrapper.vm.isYamlValid).toBe(false);
    });

    it("should set default cluster from query params", async () => {
      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.selectedClusterId).toBe("cluster-1");
    });
  });

  describe("YAML Input Handling", () => {
    it("should update yamlContent on yaml-input event", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const yamlContent = "apiVersion: v1\nkind: Pod";
      wrapper.vm.onYamlInput(yamlContent);

      expect(wrapper.vm.yamlContent).toBe(yamlContent);
    });

    it("should load projects when YAML is valid and cluster is selected", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.isYamlValid = true;
      wrapper.vm.isClusterSectionValid = true;
      wrapper.vm.loadedProjects = [];
      wrapper.vm.yamlContent = "apiVersion: v1";

      await wrapper.vm.onYamlInput("apiVersion: v1\nkind: Pod");
      await flushPromises();

      expect(orgUnitService.getProjectsMetadata).toHaveBeenCalled();
    });
  });

  it("does not assume arbitrary YAML is single-pod based on the default training form type", async () => {
    wrapper = createWrapper();
    await flushPromises();
    wrapper.vm.selectedClusterId = "cluster-1";
    wrapper.vm.isClusterSectionValid = true;
    wrapper.vm.yamlContent = "apiVersion: apps/v1\nkind: Deployment\nspec:\n  replicas: 3";
    await flushPromises();

    expect(wrapper.findComponent({ name: "GeneralSection" }).props("singlePodWorkload")).toBe(false);
  });

  describe("YAML Loaded Handling (File Upload/Paste)", () => {
    it("should reset user overrides when YAML is loaded", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.userOverrides = { project: true, priority: true, preemptibility: true };

      wrapper.vm.onYamlLoaded("apiVersion: v1\nkind: Pod");

      expect(wrapper.vm.userOverrides).toEqual({
        project: false,
        priority: false,
        preemptibility: false,
      });
    });

    it("should extract values from YAML on bulk load", async () => {
      wrapper = createWrapper();
      await flushPromises();

      // Priority and preemptibility are extracted from metadata (not spec)
      const yamlContent = `
apiVersion: v1
kind: Pod
metadata:
  namespace: runai-project-1
  priority: high
  preemptibility: Preemptible
spec:
  containers: []
`;

      wrapper.vm.onYamlLoaded(yamlContent);

      expect(wrapper.vm.yamlExtractedValues.namespace).toBe("runai-project-1");
      expect(wrapper.vm.yamlExtractedValues.priority).toBe("high");
      expect(wrapper.vm.yamlExtractedValues.preemptibility).toBe("Preemptible");
    });
  });

  describe("Project Auto-Selection by Namespace", () => {
    it("should auto-select project matching YAML namespace", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;
      wrapper.vm.userOverrides.project = false;

      wrapper.vm.autoSelectProjectByNamespace("runai-project-1");

      expect(wrapper.vm.selectedProjectId).toBe(1);
    });

    it("should NOT auto-select project if user has overridden", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;
      wrapper.vm.userOverrides.project = true;
      wrapper.vm.selectedProjectId = 2;

      wrapper.vm.autoSelectProjectByNamespace("runai-project-1");

      expect(wrapper.vm.selectedProjectId).toBe(2); // unchanged
    });

    it("should NOT auto-select if no matching project found", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;
      wrapper.vm.selectedProjectId = null;

      wrapper.vm.autoSelectProjectByNamespace("runai-nonexistent");

      expect(wrapper.vm.selectedProjectId).toBeNull();
    });
  });

  describe("Dry-Run Validation Success Handling", () => {
    it("should update isDryRunValid on successful result", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      wrapper = createWrapper();
      await flushPromises();

      const workload = {
        metadata: {
          projectId: "1",
          priority: "high",
          preemptibility: "Preemptible",
        },
      } as unknown as WorkloadV2;

      await wrapper.vm.onDryRunSuccess(workload);
      await flushPromises();

      expect(wrapper.vm.isDryRunValid).toBe(true);
      errorSpy.mockRestore();
    });

    it("should apply project from validation response", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;

      const workload = {
        metadata: { projectId: "2" },
      } as unknown as WorkloadV2;

      await wrapper.vm.onDryRunSuccess(workload);
      await flushPromises();

      expect(wrapper.vm.selectedProjectId).toBe(2);
      errorSpy.mockRestore();
    });

    it("should apply priority from validation response", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const workload = {
        metadata: { priority: "high" },
      } as unknown as WorkloadV2;

      await wrapper.vm.onDryRunSuccess(workload);
      await flushPromises();

      expect(wrapper.vm.generalModel.priorityClass).toBe("high");
    });

    it("should apply preemptibility from validation response", async () => {
      // Mock cluster version check to enable preemptibility support
      vi.spyOn(clusterStore, "isClusterVersionSufficient").mockReturnValue(true);

      wrapper = createWrapper();
      await flushPromises();

      // Set selectedClusterId so supportPreemptibility computed works
      wrapper.vm.selectedClusterId = "cluster-1";

      const workload = {
        metadata: { preemptibility: "Preemptible" },
      } as unknown as WorkloadV2;

      await wrapper.vm.onDryRunSuccess(workload);
      await flushPromises();

      expect(wrapper.vm.generalModel.preemptibility).toBe("Preemptible");
    });

    // Note: apiOverrides.priority is no longer tracked
    // Override hints are now computed by comparing generalModel values vs yamlExtractedValues
  });

  describe("Override Hints", () => {
    it("should show project override hint when API returned different project", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.apiOverrides.project = true;

      expect(wrapper.vm.showProjectOverrideHint).toBe(true);
    });

    it("should show project override hint when user manually changed project", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;
      wrapper.vm.yamlExtractedValues.namespace = "runai-project-1";
      wrapper.vm.userOverrides.project = true;
      wrapper.vm.selectedProjectId = 2; // different from YAML

      expect(wrapper.vm.showProjectOverrideHint).toBe(true);
    });

    it("should NOT show project override hint when project matches YAML", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;
      wrapper.vm.yamlExtractedValues.namespace = "runai-project-1";
      wrapper.vm.userOverrides.project = true;
      wrapper.vm.selectedProjectId = 1; // matches YAML

      expect(wrapper.vm.showProjectOverrideHint).toBe(false);
    });

    it("should show priority override hint when current value differs from YAML", async () => {
      wrapper = createWrapper();
      await flushPromises();

      // YAML has "low" but current model has "high"
      wrapper.vm.yamlExtractedValues.priority = "low";
      wrapper.vm.generalModel.priorityClass = "high";

      expect(wrapper.vm.overrideHints.priority).toBe(true);
    });

    it("should NOT show priority override hint when values match", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.yamlExtractedValues.priority = "high";
      wrapper.vm.generalModel.priorityClass = "high";

      expect(wrapper.vm.overrideHints.priority).toBe(false);
    });

    it("should show preemptibility override hint when current value differs from YAML", async () => {
      wrapper = createWrapper();
      await flushPromises();

      // YAML has Preemptible but current model has NonPreemptible
      wrapper.vm.yamlExtractedValues.preemptibility = Preemptibility.Preemptible;
      wrapper.vm.generalModel.preemptibility = Preemptibility.NonPreemptible;

      expect(wrapper.vm.overrideHints.preemptibility).toBe(true);
    });

    it("should NOT show preemptibility override hint when values match", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.yamlExtractedValues.preemptibility = Preemptibility.Preemptible;
      wrapper.vm.generalModel.preemptibility = Preemptibility.Preemptible;

      expect(wrapper.vm.overrideHints.preemptibility).toBe(false);
    });
  });

  describe("General Model Changes", () => {
    it("should update generalModel when onGeneralModelChanged is called", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const newModel = {
        ...wrapper.vm.generalModel,
        priorityClass: "high",
      };

      wrapper.vm.onGeneralModelChanged(newModel);

      expect(wrapper.vm.generalModel.priorityClass).toBe("high");
    });

    // Note: userOverrides for priority/preemptibility are no longer tracked
    // Override hints are computed by comparing generalModel vs yamlExtractedValues
  });

  describe("Selected Form Values Computed", () => {
    it("should return current form values", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.selectedProjectId = 1;
      wrapper.vm.generalModel.priorityClass = "high";
      wrapper.vm.generalModel.preemptibility = Preemptibility.Preemptible;

      const formValues = wrapper.vm.selectedFormValues;

      expect(formValues.projectId).toBe(1);
      expect(formValues.priority).toBe("high");
      expect(formValues.preemptibility).toBe(Preemptibility.Preemptible);
    });

    it("should handle null values", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.selectedProjectId = null;
      wrapper.vm.generalModel.priorityClass = null;
      wrapper.vm.generalModel.preemptibility = null;

      const formValues = wrapper.vm.selectedFormValues;

      expect(formValues.projectId).toBeNull();
      expect(formValues.priority).toBeNull();
      expect(formValues.preemptibility).toBeNull();
    });
  });

  describe("Cluster Version Requirements", () => {
    it("should disable clusters below minimum version", async () => {
      clusterStore.$patch({
        clusters: [
          { uuid: "old-cluster", tenantId: 1, name: "Old Cluster", createdAt: "2024-01-01", version: "2.20" },
          { uuid: "new-cluster", tenantId: 1, name: "New Cluster", createdAt: "2024-01-01", version: "2.23" },
        ] as never,
      });

      // Mock the version check
      vi.spyOn(clusterStore, "isClusterVersionSufficient").mockImplementation(
        (clusterId: string | null) => clusterId === "new-cluster",
      );

      wrapper = createWrapper();
      await flushPromises();

      const disabledOptions = wrapper.vm.disabledClusterOptions;

      expect(disabledOptions["old-cluster"]).toBeDefined();
      expect(disabledOptions["new-cluster"]).toBeUndefined();
    });
  });

  describe("Form Sections Visibility", () => {
    it("should show form sections when YAML has content", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.selectedClusterId = "cluster-1";
      wrapper.vm.isClusterSectionValid = true;
      wrapper.vm.yamlContent = "apiVersion: v1";

      expect(wrapper.vm.showFormSections).toBe(true);
    });

    it("should hide form sections when YAML is empty", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.selectedClusterId = "cluster-1";
      wrapper.vm.isClusterSectionValid = true;
      wrapper.vm.yamlContent = "";

      expect(wrapper.vm.showFormSections).toBe(false);
    });

    it("should hide form sections when cluster is not valid", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.selectedClusterId = "cluster-1";
      wrapper.vm.isClusterSectionValid = false;
      wrapper.vm.yamlContent = "apiVersion: v1";

      expect(wrapper.vm.showFormSections).toBe(false);
    });
  });

  describe("Form State Preservation", () => {
    const workloadStore = useWorkloadStore(pinia);

    beforeEach(() => {
      workloadStore.clearYamlFormState();
    });

    describe("saveFormState", () => {
      it("should save YAML content and general settings to store", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.yamlContent = "apiVersion: v1\nkind: Pod";
        wrapper.vm.generalModel.priorityClass = "high";
        wrapper.vm.generalModel.preemptibility = Preemptibility.Preemptible;

        wrapper.vm.saveFormState();

        expect(workloadStore.setYamlFormState).toHaveBeenCalledWith({
          yamlContent: "apiVersion: v1\nkind: Pod",
          priorityClass: "high",
          preemptibility: Preemptibility.Preemptible,
          mnnvl: null,
        });
      });

      it("should save null values when fields are not set", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.yamlContent = "apiVersion: v1";
        wrapper.vm.generalModel.priorityClass = null;
        wrapper.vm.generalModel.preemptibility = null;

        wrapper.vm.saveFormState();

        expect(workloadStore.setYamlFormState).toHaveBeenCalledWith({
          yamlContent: "apiVersion: v1",
          priorityClass: null,
          preemptibility: null,
          mnnvl: null,
        });
      });
    });

    describe("loadFormState", () => {
      it("should restore state from store and return saved state", async () => {
        const savedState = {
          yamlContent: "apiVersion: v1\nkind: Pod",
          priorityClass: "low",
          preemptibility: Preemptibility.NonPreemptible,
          mnnvl: null,
        };

        vi.mocked(workloadStore.getYamlFormState).mockReturnValue(savedState);

        wrapper = createWrapper();
        await flushPromises();

        // Reset values first
        wrapper.vm.yamlContent = "";
        wrapper.vm.generalModel.priorityClass = null;
        wrapper.vm.generalModel.preemptibility = null;

        const result = wrapper.vm.loadFormState();

        expect(result).toEqual(savedState);
        expect(wrapper.vm.yamlContent).toBe("apiVersion: v1\nkind: Pod");
        expect(wrapper.vm.generalModel.priorityClass).toBe("low");
        expect(wrapper.vm.generalModel.preemptibility).toBe(Preemptibility.NonPreemptible);
        expect(workloadStore.clearYamlFormState).toHaveBeenCalled();
      });

      it("should return null when no saved state exists", async () => {
        vi.mocked(workloadStore.getYamlFormState).mockReturnValue(null);

        wrapper = createWrapper();
        await flushPromises();

        const result = wrapper.vm.loadFormState();

        expect(result).toBeNull();
      });

      it("should handle undefined priority and preemptibility", async () => {
        const savedState = {
          yamlContent: "apiVersion: v1",
          priorityClass: null,
          preemptibility: null,
          mnnvl: null,
        };

        vi.mocked(workloadStore.getYamlFormState).mockReturnValue(savedState);

        wrapper = createWrapper();
        await flushPromises();

        const result = wrapper.vm.loadFormState();

        expect(result).toEqual(savedState);
        expect(wrapper.vm.generalModel.priorityClass).toBeNull();
        expect(wrapper.vm.generalModel.preemptibility).toBeNull();
      });
    });

    describe("clearFormState", () => {
      it("should clear form state from store", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.clearFormState();

        expect(workloadStore.clearYamlFormState).toHaveBeenCalled();
      });
    });
  });

  describe("findProjectByNamespace", () => {
    it("should find project by exact namespace match", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;

      const result = wrapper.vm.findProjectByNamespace("runai-project-1");

      expect(result).toBeDefined();
      expect(result?.id).toBe("1");
      expect(result?.name).toBe("project-1");
    });

    it("should return undefined when no project matches namespace", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = mockProjects;

      const result = wrapper.vm.findProjectByNamespace("nonexistent-namespace");

      expect(result).toBeUndefined();
    });

    it("should return undefined for empty projects list", async () => {
      wrapper = createWrapper();
      await flushPromises();

      wrapper.vm.loadedProjects = [];

      const result = wrapper.vm.findProjectByNamespace("runai-project-1");

      expect(result).toBeUndefined();
    });

    it("should match custom namespace without runai prefix", async () => {
      wrapper = createWrapper();
      await flushPromises();

      const projectsWithCustomNamespace = [
        ...mockProjects,
        {
          id: "3",
          name: "custom-project",
          namespace: "my-custom-namespace",
          phase: "Ready",
          clusterId: "cluster-1",
          createdAt: "2024-01-01",
          allocated: 0,
          deserved: 0,
        },
      ] as ProjectMetadata[];

      wrapper.vm.loadedProjects = projectsWithCustomNamespace;

      const result = wrapper.vm.findProjectByNamespace("my-custom-namespace");

      expect(result).toBeDefined();
      expect(result?.id).toBe("3");
      expect(result?.name).toBe("custom-project");
    });
  });
});
