import { policyServiceApi } from "@/services/infra/client-apis/policy-service-api/policy-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";
import { allWorkloadsIconsMap } from "@/common/icons.constant";
import type { IGvkWorkloadTypeOption } from "@/models/gvk-policy.model";
import type {
  Breakdown,
  CreateGvkPolicyReq,
  EffectiveRulesEffectiveRulesInner,
  GetGvkPolicyResp,
  GetMappingFieldInfoReq,
  GetMappingMinimalResp,
  GVK,
  ListGvkPoliciesResp,
  Mapping,
  Policy,
  UpdateGvkPolicyReq,
  WorkloadTypeAndScope,
} from "@/swagger-models/policy-service-client";

// Hidden from CREATE picker: these kinds span multiple K8s versions but a policy is tied to one, so a version match isn't guaranteed (temporary).
const CREATE_PICKER_EXCLUDED_KINDS: ReadonlySet<string> = new Set(["MPIJob", "RayJob", "RayCluster", "RayService"]);

async function listMappings(): Promise<Mapping[]> {
  try {
    return await policyServiceApi.specMappingApi.listMappings().then((res) => res.data.items || []);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listMappedWorkloadTypes(keepGvk?: GVK): Promise<IGvkWorkloadTypeOption[]> {
  const mappings = await listMappings();
  const keepKey = keepGvk ? gvkPolicyUtil.gvkKey(keepGvk) : null;
  return mappings
    .map((mapping): IGvkWorkloadTypeOption => {
      const gvk: GVK = {
        group: mapping.workloadType.group,
        version: mapping.workloadType.version,
        kind: mapping.workloadType.kind,
      };
      return {
        key: gvkPolicyUtil.gvkKey(gvk),
        label: gvk.kind,
        icon: allWorkloadsIconsMap[gvk.kind],
        gvk,
      };
    })
    .filter((option) => {
      // Exempt the edited kind (gvkKey-exact, so a different version stays hidden).
      if (keepKey && option.key === keepKey) return true;
      return !CREATE_PICKER_EXCLUDED_KINDS.has(option.gvk.kind);
    })
    .sort((a, b) => a.label.localeCompare(b.label));
}

async function getMappingMinimal(workloadType: GVK): Promise<GetMappingMinimalResp> {
  try {
    return await policyServiceApi.policyApi.getMappingMinimal({ workloadType }).then((res) => res.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getMappingFieldInfo(req: GetMappingFieldInfoReq): Promise<Breakdown> {
  try {
    return await policyServiceApi.policyApi.getMappingFieldInfo(req).then((res) => res.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createGvkPolicy(req: CreateGvkPolicyReq): Promise<Policy> {
  try {
    return await policyServiceApi.policyApi.createGvkPolicy(req).then((res) => res.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function applyGvkPolicy(policyId: string, req: UpdateGvkPolicyReq): Promise<Policy> {
  try {
    return await policyServiceApi.policyApi.applyGvkPolicy(policyId, req).then((res) => res.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function patchGvkPolicy(policyId: string, req: UpdateGvkPolicyReq): Promise<Policy> {
  try {
    return await policyServiceApi.policyApi.patchGvkPolicy(policyId, req).then((res) => res.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getGvkPolicy(policyId: string, breakdown = false): Promise<GetGvkPolicyResp> {
  try {
    return await policyServiceApi.policyApi.getGvkPolicy(policyId, breakdown).then((res) => res.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEffectivePolicy(req: WorkloadTypeAndScope): Promise<EffectiveRulesEffectiveRulesInner[]> {
  try {
    return (await policyServiceApi.policyApi.getEffectivePolicy(req)).data.effectiveRules ?? [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listGvkPolicies(): Promise<ListGvkPoliciesResp> {
  try {
    return await policyServiceApi.policyApi.listGvkPolicies().then((res) => res.data);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteGvkPolicy(policyId: string): Promise<void> {
  try {
    await policyServiceApi.policyApi.deleteGvkPolicy(policyId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

export const gvkPolicyService = {
  listMappings,
  listMappedWorkloadTypes,
  getMappingMinimal,
  getMappingFieldInfo,
  createGvkPolicy,
  applyGvkPolicy,
  patchGvkPolicy,
  getGvkPolicy,
  getEffectivePolicy,
  listGvkPolicies,
  deleteGvkPolicy,
};
