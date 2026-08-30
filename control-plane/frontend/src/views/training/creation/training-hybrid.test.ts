import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, shallowMount } from "@vue/test-utils";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";

import TrainingHybrid from "./training-hybrid.vue";
import { MIN_CLUSTER_VERSION_FOR_PRIORITY_CLASS } from "@/common/version.constant";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useTrainingStore } from "@/stores/training.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { mockQuasar } from "@/utils/unit-test.util/unit-test.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { deepCopy } from "@/utils/common.util";
import { EPolicyGroup } from "@/utils/policy.util/policy-view.util";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type { Project } from "@/swagger-models/org-unit-service-client";

vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/training.service/training.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/control-plane/policy.service/policy.service");

vi.mock("vue-router", () => ({
  useRoute: vi.fn(() => ({
    query: {
      clusterId: "priority-only-cluster",
      fromCopyId: "source-training-id",
    },
  })),
  useRouter: vi.fn(() => ({
    push: vi.fn(),
  })),
}));

describe("training-hybrid.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof TrainingHybrid>>;
  let pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });

  const copiedTrainingModel: IUISuperWorkloadCreationRequest = {
    name: "copy-training",
    projectId: "123",
    clusterId: "priority-only-cluster",
    spec: {
      image: "alpine",
      command: "sleep infinity",
      nodePools: ["default"],
      connections: [],
      priorityClass: "low",
      preemptibility: "preemptible",
    },
  };

  function seedFormState(): void {
    const trainingStore = useTrainingStore(pinia);
    const workloadModel = workloadHybridUtil.setTrainingUIDefaults(deepCopy(copiedTrainingModel), false, null, "2.23");

    trainingStore.setFlatFormState({
      projectInfo: { name: "project-1" },
      projectNodePoolsResources: [],
      selectedTemplates: workloadHybridUtil.getWorkloadHybridTemplatesSetup(),
      policy: null,
      policyViewDetails: {
        [EPolicyGroup.Environment]: {},
        [EPolicyGroup.Compute]: {},
        [EPolicyGroup.Storage]: {},
        [EPolicyGroup.General]: {},
      },
      selectedProjectNodeTypes: [],
      workloadModel,
    });
  }

  function factory(): void {
    wrapper = shallowMount(TrainingHybrid, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: {
              clusterId: "priority-only-cluster",
              fromCopyId: "source-training-id",
            },
          },
          $q: mockQuasar,
          $router: {
            push: vi.fn(),
          },
        },
      },
    });
  }

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
    setActivePinia(pinia);

    useAppStore(pinia);
    useAuthStore(pinia);
    useWorkloadStore(pinia);
    const clusterStore = useClusterStore(pinia);
    const dataSourceStore = useDataSourceStore(pinia);

    vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation((_, minVersion) => {
      return minVersion === MIN_CLUSTER_VERSION_FOR_PRIORITY_CLASS;
    });
    vi.mocked(dataSourceStore.getAvailableDataSources).mockResolvedValue(new Set());
    vi.mocked(dataSourceService.listExistingConfigMaps).mockResolvedValue([]);
    vi.mocked(credentialService.listCredentials).mockResolvedValue([]);
    vi.mocked(credentialService.listExistingSecrets).mockResolvedValue([]);
    vi.mocked(orgUnitService.getProject).mockResolvedValue({ name: "project-1" } as Project);
    vi.mocked(workloadService.canSubmitWorkload).mockResolvedValue(null);
    vi.mocked(workloadService.getWorkloadClusterPriorities).mockResolvedValue([{ id: "low", name: "low", value: 50 }]);
    vi.mocked(workloadService.getDefaultWorkloadPriorityByType).mockResolvedValue({ priorityName: "low" } as never);
    vi.mocked(trainingService.createTrainingFlat).mockResolvedValue({ workloadId: "created-id" } as never);

    seedFormState();
    factory();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("updateGeneralModel - should preserve priority and omit unsupported preemptibility", async () => {
    await flushPromises();

    wrapper.vm.updateGeneralModel({
      priorityClass: "low",
      preemptibility: "preemptible",
    });

    expect(wrapper.vm.trainingModel.spec.priorityClass).toBe("low");
    expect(wrapper.vm.trainingModel.spec.preemptibility).toBeUndefined();
  });

  it("onSubmit - should omit copied preemptibility while preserving priority on unsupported clusters", async () => {
    await flushPromises();

    await wrapper.vm.onSubmit();

    const request = vi.mocked(trainingService.createTrainingFlat).mock.calls[0][0];
    expect(request.spec?.priorityClass).toBe("low");
    expect(request.spec?.preemptibility).toBeUndefined();
  });
});
