/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
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
import { SettingKeys } from "@/models/setting.model";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";

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
    vi.mocked(orgUnitService.getProjectsMetadata).mockResolvedValue(projects);
    clusterStore.$patch({ clusters });
    clusterStore.$patch({ selectedCluster: clusters[0] });

    inferenceStore.$patch({
      inference: inferenceWorkloadUtil.getEmptyUIInferenceCreation(),
    });

    appStore.$patch({ pageLoading: false });
    authStore.$patch({ orgUnits: [] });
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
            template: "<div>Template Section V2</div>",
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
      expect(wrapper.vm.inferenceNewModel.templateId).toBeNull();
      expect(wrapper.vm.showTemplates).toBe(false);
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
      vi.mocked(settingStore.isFeatureEnabled).mockImplementation(
        (value: SettingKeys) => value === SettingKeys.EnableTemplatesV2,
      );

      wrapper = createWrapper();
      await flushPromises();

      expect(wrapper.vm.showTemplates).toBe(true);
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
