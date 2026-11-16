/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import InferenceNew from "./inference-new.vue";

// stores
import { useSettingStore } from "@/stores/setting.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";

// services
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// mocks
import { clusters, emptyMergedWorkload, projects } from "./inference-new.test.mock";

// utils
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { EServerType } from "@/models/inference.model";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION } from "@/common/version.constant";

// Mock services
vi.mock("@/services/infra/unleash.service/unleash.service");
vi.mock("@/services/control-plane/inference.service/inference.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/workload-template.service/workload-template.service");

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const mockRouter = {
  push: vi.fn(),
  back: vi.fn(),
};

const mockNotify = vi.fn();

const PROJECT_AS_UNSELECTED = -1;

describe("inference-new.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof InferenceNew>>;
  const inferenceStore = useInferenceStore(pinia);
  const appStore = useAppStore(pinia);
  const clusterStore = useClusterStore(pinia);
  const settingStore = useSettingStore(pinia);
  const authStore = useAuthStore(pinia);

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(orgUnitService.getProjectsMetadata).mockResolvedValue(projects);
    clusterStore.$patch({ clusters });
    clusterStore.$patch({ selectedCluster: clusters[0] });
    vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation(
      (_, version) => version === MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION,
    );
    settingStore.$patch({ enableHybridSubmission: { value: true } });

    inferenceStore.$patch({
      inference: inferenceWorkloadUtil.getEmptyUIInferenceCreation(),
    });

    appStore.$patch({ pageLoading: false });
    authStore.$patch({ orgUnits: [] });
  });

  afterEach(() => {
    wrapper?.unmount();
  });

  function createWrapper() {
    return mount(InferenceNew, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $router: mockRouter,
          $route: {
            query: {},
            params: {},
          },
          $q: {
            notify: mockNotify,
          },
        },
        stubs: {
          RunaiClusterSection: {
            template: "<div>Cluster Section</div>",
            props: ["entityName", "clusterId", "disabledClusterOptions"],
            emits: ["update:cluster-id"],
          },
          ProjectSection: {
            template: "<div aid='project-section'>Project Section</div>",
            props: ["aid", "loading", "selectedProjectId", "projects", "clusterId"],
            emits: ["project-changed"],
          },
          TemplateSectionV2: {
            template: "<div :aid='aid'>Template Section V2</div>",
            props: ["aid", "templates", "totalTemplatesCount", "loading", "listQuery", "templateId"],
            emits: ["template-changed", "query-changed", "get-next-batch"],
          },
          InferenceServer: {
            template: "<div>Inference Server</div>",
            props: ["isSelectable", "modelValue", "clusterId"],
            emits: ["update-section-model"],
          },
          TitleAndBackSection: {
            template: "<div>Title and Back Section</div>",
            props: ["label", "summary"],
          },
          WorkloadNameSection: {
            template: "<div>Workload Name Section</div>",
            props: ["entityType", "disableClosing", "projectId", "clusterId"],
          },
          RunaiBtn: {
            template: "<div>RunaiBtn</div>",
            props: ["aid", "flat", "color", "label", "loading"],
          },
        },
      },
    });
  }

  describe("Created", () => {
    it("should handle new inference creation from workload page", async () => {
      vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
      vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
        pagination: { limit: 5 },
        entries: [],
      });
      vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
      vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);

      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.inferenceNewModel.name).toBe("");
      expect(wrapper.vm.inferenceNewModel.serverType).toBe(EServerType.Custom);
      expect(wrapper.vm.inferenceNewModel.templateId).toBe("-1");
      expect(wrapper.vm.showTemplates).toBe(true);
      expect(wrapper.vm.inferenceNewModel.projectId.toString()).toBe(projects[0].id);
      expect(wrapper.vm.showInferenceTypeSelection).toBe(false);
    });

    it("should handle new inference creation from workload page with templates enabled", async () => {
      vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
      vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
        pagination: { limit: 5 },
        entries: [],
      });
      vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);

      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.showTemplates).toBe(true);
    });
  });

  describe("Flexible template availability", () => {
    beforeEach(() => {
      vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
      vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
        pagination: { limit: 5 },
        entries: [],
      });
      vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
    });

    it.each([
      { enabled: true, supported: true },
      { enabled: false, supported: true },
      { enabled: true, supported: false },
    ])(
      "only offers and loads templates when submission is supported ($enabled, $supported)",
      async ({ enabled, supported }) => {
        settingStore.$patch({ enableHybridSubmission: { value: enabled } });
        vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation(
          (_, version) => supported && version === MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION,
        );

        wrapper = createWrapper();
        await flushPromises();

        expect(wrapper.find("[aid=template-section]").exists()).toBe(enabled && supported);
        if (enabled && supported) {
          expect(workloadTemplateService.inferenceTemplatesList).toHaveBeenCalled();
        } else {
          expect(workloadTemplateService.inferenceTemplatesList).not.toHaveBeenCalled();
        }
      },
    );

    it("passes the selected template to flexible inference submission", async () => {
      wrapper = createWrapper();
      await flushPromises();
      wrapper.vm.onSelectedTemplate("template-123");

      await wrapper.vm.continueToCustom();

      expect(mockRouter.push).toHaveBeenCalledWith({
        name: INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID,
        query: {
          clusterId: clusters[0].uuid,
          projectId: projects[0].id,
          inferenceName: "",
          templateId: "template-123",
        },
      });
      expect(inferenceService.getInferenceMergedWithPolicy).not.toHaveBeenCalled();
    });

    it("clears a selected template when switching to a cluster without flexible submission", async () => {
      wrapper = createWrapper();
      await flushPromises();
      wrapper.vm.onSelectedTemplate("template-123");
      const unsupportedCluster = { ...clusters[0], uuid: "unsupported-cluster", version: "2.17.0" };
      clusterStore.$patch({ clusters: [...clusters, unsupportedCluster] });
      vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation(
        (id, version) => id === clusters[0].uuid && version === MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION,
      );
      vi.mocked(workloadTemplateService.inferenceTemplatesList).mockClear();

      await wrapper.vm.updateClusterId(unsupportedCluster.uuid, { disabled: false });
      await flushPromises();

      expect(wrapper.vm.inferenceNewModel.templateId).toBeNull();
      expect(wrapper.find("[aid=template-section]").exists()).toBe(false);
      expect(workloadTemplateService.inferenceTemplatesList).not.toHaveBeenCalled();

      await wrapper.vm.updateClusterId(clusters[0].uuid, { disabled: false });
      await flushPromises();

      expect(wrapper.vm.inferenceNewModel.templateId).toBe("-1");
      expect(wrapper.find("[aid=template-section]").exists()).toBe(true);
      expect(workloadTemplateService.inferenceTemplatesList).toHaveBeenCalledOnce();
    });

    it("clears a restored template when flexible submission is disabled", async () => {
      settingStore.$patch({ enableHybridSubmission: { value: false } });
      inferenceStore.$patch({ inference: { templateId: "template-123", clusterId: clusters[0].uuid } });

      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.inferenceNewModel.templateId).toBeNull();
      expect(wrapper.find("[aid=template-section]").exists()).toBe(false);
      expect(workloadTemplateService.inferenceTemplatesList).not.toHaveBeenCalled();
    });
  });

  describe("Methods", () => {
    describe("onContinue", () => {
      describe("validation scenarios", () => {
        it("should handle validation failure", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);

          wrapper = createWrapper();
          wrapper.vm.validate = vi.fn().mockResolvedValue(false);
          wrapper.vm.showHint = vi.fn();

          await wrapper.vm.onContinue();

          expect(wrapper.vm.submitting).toBe(false);
          expect(wrapper.vm.displayFormHint).toBe(false);
          expect(wrapper.vm.showHint).toHaveBeenCalled();
        });

        it("should handle missing project ID", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.projectId = PROJECT_AS_UNSELECTED;

          const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {
            // do nothing
          });

          await wrapper.vm.onContinue();
          expect(consoleSpy).toHaveBeenCalledWith("ProjectId is missing");
          expect(mockNotify).toHaveBeenCalled();
          expect(wrapper.vm.submitting).toBe(false);

          consoleSpy.mockRestore();
        });
      });

      describe("serving port access handling", () => {
        it("should navigate to assets edit when serving port access not supported", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);
          wrapper = createWrapper();
          await flushPromises();

          vi.spyOn(wrapper.vm, "supportingServingPortAccess", "get").mockReturnValue(false);
          vi.spyOn(wrapper.vm, "isHybridSupported", "get").mockReturnValue(false);

          wrapper.vm.inferenceNewModel.serverType = EServerType.Custom;
          await wrapper.vm.onContinue();
          expect(mockRouter.push).toHaveBeenCalledWith(
            expect.objectContaining({ name: INFERENCE_ROUTE_NAMES.INFERENCE_ASSETS_EDIT }),
          );
          expect(inferenceStore.inference).toBeDefined();
        });

        it("should keep servingPortAccess when supported", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.serverType = EServerType.Custom;
          vi.spyOn(wrapper.vm, "supportingServingPortAccess", "get").mockReturnValue(true);
          vi.spyOn(wrapper.vm, "isHybridSupported", "get").mockReturnValue(false);

          await wrapper.vm.onContinue();
          expect(inferenceStore.inference?.specificEnv?.servingPortAccess).toBeDefined();
        });
      });

      describe("template handling", () => {
        it("should call continueToCustom when template is selected and serverType is Custom", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.templateId = "template-123";
          wrapper.vm.inferenceNewModel.serverType = EServerType.Custom;
          wrapper.vm.continueToCustom = vi.fn();

          await wrapper.vm.onContinue();

          expect(wrapper.vm.continueToCustom).toHaveBeenCalled();
        });

        it("should route by serverType when no template is selected", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.templateId = null;
          wrapper.vm.inferenceNewModel.serverType = EServerType.VLLM;
          wrapper.vm.continueToVLLM = vi.fn();

          await wrapper.vm.onContinue();

          expect(wrapper.vm.continueToVLLM).toHaveBeenCalled();
        });
      });

      describe("inference type routing", () => {
        it("should route to NIM inference", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.serverType = EServerType.Nim;
          wrapper.vm.continueToNim = vi.fn();

          await wrapper.vm.onContinue();

          expect(wrapper.vm.continueToNim).toHaveBeenCalled();
        });

        it("should route to VLLM inference", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.serverType = EServerType.VLLM;
          wrapper.vm.continueToVLLM = vi.fn();

          await wrapper.vm.onContinue();

          expect(wrapper.vm.continueToVLLM).toHaveBeenCalled();
        });

        it("should route to custom inference", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.serverType = EServerType.Custom;
          wrapper.vm.continueToCustom = vi.fn();

          await wrapper.vm.onContinue();

          expect(wrapper.vm.continueToCustom).toHaveBeenCalled();
        });

        it("should route to custom inference when template is selected and serverType is Custom", async () => {
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(true);
          vi.mocked(workloadTemplateService.inferenceTemplatesList).mockResolvedValue({
            pagination: { limit: 5 },
            entries: [],
          });
          vi.mocked(workloadTemplateService.templatesCount).mockResolvedValue(0);
          vi.mocked(settingStore.isFeatureEnabled).mockReturnValue(false);
          vi.mocked(inferenceService.getInferenceMergedWithPolicy).mockResolvedValue(emptyMergedWorkload);

          wrapper = createWrapper();
          await flushPromises();

          wrapper.vm.inferenceNewModel.serverType = EServerType.Custom;
          wrapper.vm.inferenceNewModel.templateId = "template-123";
          wrapper.vm.continueToCustom = vi.fn();

          await wrapper.vm.onContinue();

          expect(wrapper.vm.continueToCustom).toHaveBeenCalled();
        });
      });
    });

    describe("continueToDistributed", () => {
      it("preserves the selected server type in the distributed route", async () => {
        wrapper = createWrapper();
        wrapper.vm.inferenceNewModel.clusterId = "cluster-1";
        wrapper.vm.inferenceNewModel.projectId = 42;
        wrapper.vm.inferenceNewModel.name = "distributed-nim";
        wrapper.vm.inferenceNewModel.serverType = EServerType.Nim;

        await wrapper.vm.continueToDistributed();

        expect(mockRouter.push).toHaveBeenCalledWith({
          name: INFERENCE_ROUTE_NAMES.INFERENCE_DISTRIBUTED_EDIT,
          query: {
            clusterId: "cluster-1",
            projectId: "42",
            inferenceName: "distributed-nim",
            serverType: EServerType.Nim,
          },
        });
      });
    });
  });
});
