/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { ref, computed } from "vue";
import { UidGidSource } from "@/swagger-models/workloads-client";
import InferenceHuggingFaceBased from "./inference-hugging-face-based.vue";

import { policyUtil } from "@/utils/policy.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { getPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";

import { useAppStore } from "@/stores/app.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";

vi.mock("@/services/infra/unleash.service/unleash.service");
vi.mock("@/services/control-plane/inference.service/inference.service");
vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/infra/router.service/router.service");
vi.mock("@/utils/inference-workload.util/inference-workload.util");

vi.mock("@/utils/policy.util", async (importOriginal) => {
  const original: Record<string, unknown> = await importOriginal();
  return {
    ...original,
    policyUtil: {
      ...(original.policyUtil as Record<string, unknown>),
      cleanupFlatItemizeBeforeSubmission: vi.fn((spec: unknown) => spec),
    },
  };
});

const mockInferenceModel = {
  name: "test-inference",
  projectId: "1",
  clusterId: "cluster-1",
  spec: {
    annotations: [{ name: "anno1", value: "v1" }],
    environmentVariables: [{ name: "env1", value: "v1" }],
    connections: [],
  },
};

vi.mock("@/composables/use-hybrid-templates.composable", () => ({
  useHybridTemplates: () => ({
    isDrawerOpen: ref(false),
    hybridWorkloadModel: ref({ ...mockInferenceModel, spec: { ...mockInferenceModel.spec } }),
    selectedTemplates: ref({}),
    drawerDataEntity: ref(null),
    recentlyCreatedEntityId: ref(null),
    closeDrawer: vi.fn(),
    getUnifiedHybridWorkload: vi.fn(),
    displayComputesList: vi.fn(),
    resetComputeToDefault: vi.fn(),
    clearComputeTemplate: vi.fn(),
    handleSelection: vi.fn(),
    handleMultiSelection: vi.fn(),
    onCreateNewSetup: vi.fn(),
    displayDataSourcesList: vi.fn(),
    displayModelStorageList: vi.fn(),
    handleCreatedAssets: vi.fn(),
    selectedComputeName: computed(() => ""),
  }),
}));

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const mockRouter = { push: vi.fn(), back: vi.fn() };
const mockNotify = vi.fn();

describe("inference-hugging-face-based.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof InferenceHuggingFaceBased>>;
  const appStore = useAppStore(pinia);
  const inferenceStore = useInferenceStore(pinia);
  const clusterStore = useClusterStore(pinia);
  const authStore = useAuthStore(pinia);

  beforeEach(() => {
    vi.clearAllMocks();
    appStore.$patch({ pageLoading: false });
    authStore.$patch({ orgUnits: [], tenant: { id: 1 } } as never);
    clusterStore.$patch({
      clusters: [{ uuid: "cluster-1", name: "test-cluster", tenantId: 1 }],
      selectedCluster: { uuid: "cluster-1", name: "test-cluster", tenantId: 1 },
    } as never);
    inferenceStore.$patch({
      inference: { serverType: "HuggingFace" } as never,
    });
    vi.mocked(inferenceStore.getFlatFormState).mockReturnValue(null);
    vi.mocked(inferenceStore.getInferenceFormUIState).mockReturnValue({} as never);
    vi.mocked(inferenceService.createInferenceFlatApi).mockResolvedValue({
      meta: { workloadId: "w-1", name: "test-inference" },
    } as never);
    vi.mocked(inferenceWorkloadUtil.prepareInferenceFlatRequest).mockReturnValue({} as never);
    vi.mocked(inferenceWorkloadUtil.convertInferenceToWorkload).mockReturnValue({} as never);
  });

  function createWrapper() {
    return mount(InferenceHuggingFaceBased, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $router: mockRouter,
          $route: { query: { clusterId: "cluster-1", projectId: "1" }, params: {} },
          $q: { notify: mockNotify },
        },
        stubs: {
          HfInferenceSubmitForm: { template: "<div />" },
        },
      },
    });
  }

  describe("submitInference", () => {
    it.each([UidGidSource.FromIdpToken, UidGidSource.Custom, UidGidSource.FromTheImage])(
      "serializes %s identity using the real request preparation",
      async (uidGidSource) => {
        const actual = await vi.importActual<{ inferenceWorkloadUtil: typeof inferenceWorkloadUtil }>(
          "@/utils/inference-workload.util/inference-workload.util",
        );
        vi.mocked(inferenceWorkloadUtil.prepareInferenceFlatRequest).mockImplementation(
          actual.inferenceWorkloadUtil.prepareInferenceFlatRequest,
        );
        const identity = { runAsUid: 2001, runAsGid: 2002, supplementalGroups: "2003,2004" };
        const model = {
          ...mockInferenceModel,
          spec: {
            image: "inference-image",
            servingPort: { protocol: "http", container: 8080 },
            autoscaling: { minReplicas: 1, maxReplicas: 1 },
            security: { uidGidSource, hostIpc: true, ...identity },
          },
        };
        vi.mocked(inferenceStore.getFlatFormState).mockReturnValue({
          projectInfo: {},
          projectNodePoolsResources: [],
          selectedTemplates: workloadHybridUtil.getWorkloadHybridTemplatesSetup(),
          policy: null,
          policyViewDetails: getPolicyViewDetails({}),
          selectedProjectNodeTypes: [],
          workloadModel: model,
        });
        wrapper = createWrapper();
        await flushPromises();
        const before = JSON.stringify(wrapper.vm.inference);

        await wrapper.vm.submitInference();

        expect(inferenceService.createInferenceFlatApi).toHaveBeenCalledOnce();
        const request = JSON.parse(JSON.stringify(vi.mocked(inferenceService.createInferenceFlatApi).mock.calls[0][0]));
        expect(request.spec.security).toEqual({
          uidGidSource,
          hostIpc: true,
          ...(uidGidSource === UidGidSource.FromIdpToken ? {} : identity),
        });
        expect(JSON.stringify(wrapper.vm.inference)).toBe(before);
        wrapper.unmount();
      },
    );

    it("should call cleanupFlatItemizeBeforeSubmission before submitting", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      wrapper = createWrapper();
      await flushPromises();

      const mockPolicy = { rules: { annotations: { instances: { locked: ["anno1"] } } } };
      wrapper.vm.policy = mockPolicy;

      await wrapper.vm.submitInference();

      expect(policyUtil.cleanupFlatItemizeBeforeSubmission).toHaveBeenCalledWith(
        expect.objectContaining({ annotations: expect.any(Array) }),
        mockPolicy,
      );
      expect(inferenceService.createInferenceFlatApi).toHaveBeenCalled();
      errorSpy.mockRestore();
    });
  });
});
