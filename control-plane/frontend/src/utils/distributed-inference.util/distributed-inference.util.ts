import { isEqual } from "@/utils/common.util";
import type { UnifiedPolicyInfo } from "@/swagger-models/workloads-client";
import { AssetKind } from "@/swagger-models/assets-service-client";
import type { TDataSourceKinds } from "@/models/data-source.model";

/**
 * Utilities for distributed inference workloads.
 */

/**
 * Returns the supported data sources for distributed inference workloads.
 * Based on DistributedInferenceLeaderWorkerSpecStorage interface which supports:
 * - configMapVolume (ConfigMap)
 * - pvc (PVC)
 * - secretVolume (Secret)
 */
function getSupportedDataSources(): TDataSourceKinds[] {
  return [AssetKind.ConfigMap, AssetKind.Pvc, AssetKind.SecretVolume];
}

/**
 * Policy handling logic:
 * 1. No policy or no worker policy => copy leader data into worker
 * 2. Policy is equal between leader and worker (defaults AND rules) => copy leader data into worker
 * 3. Policy is NOT equal between leader and worker => do not change worker data, show mismatch alert
 */

/**
 * Checks if the worker policy is missing or empty (no defaults and no rules).
 */
function hasNoWorkerPolicy(policy: UnifiedPolicyInfo | null | undefined): boolean {
  const workerPolicy = policy?.worker;
  return !workerPolicy || (!workerPolicy.defaults && !workerPolicy.rules);
}

/**
 * Compares leader and worker policies within the worker policy section.
 * Returns true if both defaults and rules are deeply equal between leader and worker.
 */
function areLeaderAndWorkerPoliciesEqual(policy: UnifiedPolicyInfo | null | undefined): boolean {
  const leaderDefaults = policy?.worker?.defaults?.leader;
  const workerDefaults = policy?.worker?.defaults?.worker;
  const leaderRules = policy?.worker?.rules?.leader;
  const workerRules = policy?.worker?.rules?.worker;

  const areDefaultsEqual = isEqual(leaderDefaults, workerDefaults);
  const areRulesEqual = isEqual(leaderRules, workerRules);

  return areDefaultsEqual && areRulesEqual;
}

/**
 * Determines whether leader data should be copied to the worker.
 * Returns true if:
 * - No policy exists
 * - No worker policy exists (missing or empty)
 * - Leader and worker policies are equal
 */
function shouldCopyLeaderToWorker(policy: UnifiedPolicyInfo | null | undefined): boolean {
  return !policy || hasNoWorkerPolicy(policy) || areLeaderAndWorkerPoliciesEqual(policy);
}

export const distributedInferenceUtil = {
  getSupportedDataSources,
  hasNoWorkerPolicy,
  areLeaderAndWorkerPoliciesEqual,
  shouldCopyLeaderToWorker,
};
