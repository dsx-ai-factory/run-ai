import { describe, it, expect, vi, beforeEach } from "vitest";
import { gvkPolicyService } from "./gvk-policy.service";
import { policyServiceApi } from "@/services/infra/client-apis/policy-service-api/policy-service-api";
import { ScopeType } from "@/swagger-models/policy-service-client";
import type { GVK, Mapping } from "@/swagger-models/policy-service-client";

vi.mock("@/services/infra/client-apis/policy-service-api/policy-service-api", () => ({
  policyServiceApi: {
    specMappingApi: {
      listMappings: vi.fn(),
    },
    policyApi: {
      getEffectivePolicy: vi.fn(),
    },
  },
}));

const listMappingsApiMock = vi.mocked(policyServiceApi.specMappingApi.listMappings);
const getEffectiveApiMock = vi.mocked(policyServiceApi.policyApi.getEffectivePolicy);
type GetEffectiveResult = Awaited<ReturnType<typeof policyServiceApi.policyApi.getEffectivePolicy>>;
type ListMappingsResult = Awaited<ReturnType<typeof policyServiceApi.specMappingApi.listMappings>>;

function mapping(group: string, version: string, kind: string): Mapping {
  return { workloadType: { group, version, kind } } as Mapping;
}

function mockMappings(items: Mapping[]): void {
  listMappingsApiMock.mockResolvedValue({ data: { items } } as ListMappingsResult);
}

const ALLOWED = [
  mapping("apps", "v1", "Deployment"),
  mapping("batch", "v1", "Job"),
  mapping("", "v1", "Pod"), // bare Pod is now offered in the create picker (product decision)
];
// The kinds the create picker must hide (multi-version Ray*/MPIJob).
const EXCLUDED = [
  mapping("kubeflow.org", "v2beta1", "MPIJob"),
  mapping("ray.io", "v1", "RayJob"),
  mapping("ray.io", "v1", "RayCluster"),
  mapping("ray.io", "v1", "RayService"),
];

describe("gvkPolicyService.listMappedWorkloadTypes", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockMappings([...ALLOWED, ...EXCLUDED]);
  });

  it("hides the multi-version kinds (Ray* / MPIJob) but offers Pod in the create picker", async () => {
    const kinds = (await gvkPolicyService.listMappedWorkloadTypes()).map((o) => o.gvk.kind);
    expect(kinds).toEqual(["Deployment", "Job", "Pod"]); // sorted, multi-version kinds gone, Pod admitted
    for (const hidden of ["MPIJob", "RayJob", "RayCluster", "RayService"]) {
      expect(kinds).not.toContain(hidden);
    }
  });

  it("re-admits only the kind currently being edited (keepGvk), keeping the rest hidden", async () => {
    const editing: GVK = { group: "kubeflow.org", version: "v2beta1", kind: "MPIJob" };
    const kinds = (await gvkPolicyService.listMappedWorkloadTypes(editing)).map((o) => o.gvk.kind);
    expect(kinds).toContain("MPIJob"); // exempted so the edit flow can load it
    expect(kinds).toContain("Deployment");
    expect(kinds).toContain("Pod"); // Pod is no longer excluded
    for (const stillHidden of ["RayJob", "RayCluster", "RayService"]) {
      expect(kinds).not.toContain(stillHidden);
    }
  });

  it("exempts by exact GVK key — a different version of the same kind stays hidden", async () => {
    // Only MPIJob/v2beta1 is mapped; editing v1 must NOT re-admit the v2beta1 option.
    const editingV1: GVK = { group: "kubeflow.org", version: "v1", kind: "MPIJob" };
    const kinds = (await gvkPolicyService.listMappedWorkloadTypes(editingV1)).map((o) => o.gvk.kind);
    expect(kinds).not.toContain("MPIJob");
  });

  it("returns the surviving options sorted by label", async () => {
    mockMappings([
      mapping("apps", "v1", "StatefulSet"),
      mapping("apps", "v1", "Deployment"),
      mapping("batch", "v1", "Job"),
    ]);
    const labels = (await gvkPolicyService.listMappedWorkloadTypes()).map((o) => o.label);
    expect(labels).toEqual(["Deployment", "Job", "StatefulSet"]);
  });
});

describe("gvkPolicyService.getEffectivePolicy", () => {
  beforeEach(() => vi.clearAllMocks());

  const req = {
    workloadType: { group: "kubeflow.org", version: "v1", kind: "JAXJob" },
    scope: ScopeType.Project,
    projectId: "123",
  };
  const rules = [
    {
      field: "replicas",
      specSelector: "replicaSpecByType",
      enforcements: [],
      sourceOfRule: { scope: ScopeType.Project },
    },
  ];

  it("returns the rules array from the { effectiveRules } response", async () => {
    getEffectiveApiMock.mockResolvedValue({ data: { effectiveRules: rules } } as unknown as GetEffectiveResult);
    const result = await gvkPolicyService.getEffectivePolicy(req);
    expect(getEffectiveApiMock).toHaveBeenCalledWith(req);
    expect(result).toEqual(rules);
  });

  it("returns an empty array when effectiveRules is absent", async () => {
    getEffectiveApiMock.mockResolvedValue({ data: {} } as unknown as GetEffectiveResult);
    expect(await gvkPolicyService.getEffectivePolicy(req)).toEqual([]);
  });
});
