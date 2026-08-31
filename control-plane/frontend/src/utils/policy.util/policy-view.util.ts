/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable  @typescript-eslint/no-explicit-any */
import type {
  DistributedPolicyDefaultsAndRulesV2,
  DistributedPolicyDefaultsV2Worker,
  DistributedPolicyRulesV2Worker,
  ReplicaDefaultsV2,
  ReplicaRulesV2,
  SourceOfRule,
  TrainingPolicyDefaultsAndRulesV2,
  TrainingPolicyDefaultsAndRulesV2Defaults,
  TrainingPolicyDefaultsAndRulesV2Rules,
  WorkspacePolicyDefaultsAndRulesV2,
  WorkspacePolicyDefaultsAndRulesV2Defaults,
  WorkspacePolicyDefaultsAndRulesV2Rules,
} from "@/swagger-models/policy-service-client";

import type { EnvironmentAssetSpec } from "@/swagger-models/assets-service-client";
import type { UnifiedPolicyInfo } from "@/swagger-models/workloads-client";
import { omit } from "../common.util";

interface ExtendedEnvironmentAssetSpec extends EnvironmentAssetSpec {
  privileged?: boolean;
}

export interface IPolicyViewProperty {
  name: string;
  default: any;
  rules: any;
  sourceOfRule: SourceOfRule | null;
  type?: string;
  replica?: string;
}

export const enum EPolicyGroup {
  Compute = "compute",
  Environment = "environment",
  Storage = "storage",
  General = "general",
}

export interface IPolicyViewDetails {
  [EPolicyGroup.Environment]: Record<string, IPolicyViewProperty>;
  [EPolicyGroup.Compute]: Record<string, IPolicyViewProperty>;
  [EPolicyGroup.Storage]: Record<string, IPolicyViewProperty>;
  [EPolicyGroup.General]: Record<string, IPolicyViewProperty>;
}

type TDistributedTypes = "worker" | "master" | "leader";

type TAllWorkloadPolicyRules =
  | WorkspacePolicyDefaultsAndRulesV2Rules
  | TrainingPolicyDefaultsAndRulesV2Rules
  | DistributedPolicyRulesV2Worker
  | ReplicaRulesV2;

type TAllWorkloadPolicyDefaults =
  | TrainingPolicyDefaultsAndRulesV2Defaults
  | DistributedPolicyDefaultsV2Worker
  | ReplicaDefaultsV2
  | WorkspacePolicyDefaultsAndRulesV2Defaults;

const environmentKeys: (keyof ExtendedEnvironmentAssetSpec)[] = [
  "command",
  "args",
  "environmentVariables",
  "runAsUid",
  "runAsGid",
  "supplementalGroups",
  "image",
  "imagePullPolicy",
  "workingDir",
  "hostIpc",
  "hostNetwork",
  "connections",
  "createHomeDir",
  "allowPrivilegeEscalation",
  "privileged",
  "uidGidSource",
  "overrideUidGidInWorkspace",
  "capabilities",
  "seccompProfileType",
  "runAsNonRoot",
  "probes",
];

const environmentKeySet = new Set(environmentKeys);

export function getPolicyViewDetails(
  policy: WorkspacePolicyDefaultsAndRulesV2 | TrainingPolicyDefaultsAndRulesV2,
): IPolicyViewDetails {
  const defaultsBySections = (policy.defaults && setSections(policy.defaults)) || null;
  const rulesBySections = (policy.rules && setSections(policy.rules)) || null;

  const flattenPolicy = {
    rules: (rulesBySections && flattenPolicyObject(rulesBySections)) || null,
    defaults: (defaultsBySections && flattenPolicyObject(defaultsBySections)) || null,
    imposedAssets: policy.imposedAssets || null,
  };

  const matchedPolicy = matchFlattenedPolicyRulesAndDefaults(flattenPolicy);

  return finalizePolicyForView(matchedPolicy);
}

export function getDistributedPolicyViewDetails(policy: DistributedPolicyDefaultsAndRulesV2): IPolicyViewDetails {
  const readyState = {} as Record<TDistributedTypes, IPolicyViewDetails>;

  const masterPolicy = {
    defaults: { ...(policy.defaults?.master || {}) },
    rules: { ...(policy.rules?.master || {}) },
  };
  const workerPolicy = {
    defaults: { ...(policy.defaults?.worker || {}) },
    rules: { ...(policy.rules?.worker || {}) },
  };

  [
    { key: "master" as TDistributedTypes, policy: masterPolicy },
    { key: "worker" as TDistributedTypes, policy: workerPolicy },
  ].forEach((p) => {
    // going through the same process as a regular policy for each of the properties
    const defaultsBySections = (p.policy.defaults && setSections(p.policy.defaults)) || null;
    const rulesBySections = (p.policy.rules && setSections(p.policy.rules)) || null;

    const flattenPolicy = {
      rules: (rulesBySections && flattenPolicyObject(rulesBySections)) || null,
      defaults: (defaultsBySections && flattenPolicyObject(defaultsBySections)) || null,
      imposedAssets: policy.imposedAssets || null,
    };

    const matchedPolicy = matchFlattenedPolicyRulesAndDefaults(flattenPolicy);
    // making sure to add type to differentiate between the properties
    readyState[p.key as TDistributedTypes] = finalizePolicyForView(matchedPolicy, p.key);
  });
  //  combining the properties into a single view
  return {
    environment: { ...readyState.master.environment, ...readyState.worker.environment },
    compute: { ...readyState.master.compute, ...readyState.worker.compute },
    general: { ...readyState.master.general, ...readyState.worker.general },
    storage: { ...readyState.master.storage, ...readyState.worker.storage },
  };
}

export function getInferenceDistributedPolicyViewDetails(
  policy: UnifiedPolicyInfo | null | undefined,
): IPolicyViewDetails {
  if (!policy?.worker) {
    return { environment: {}, compute: {}, storage: {}, general: {} };
  }
  const w = policy.worker as Record<string, any>;
  const leaderPolicy = {
    defaults: { ...(w.defaults ?? {}), ...(w.defaults?.leader ?? {}) },
    rules: { ...(w.rules ?? {}), ...(w.rules?.leader ?? {}) },
  };
  const workerPolicy = {
    defaults: { ...(w.defaults ?? {}), ...(w.defaults?.worker ?? {}) },
    rules: { ...(w.rules ?? {}), ...(w.rules?.worker ?? {}) },
  };

  const readyState = {} as Record<string, IPolicyViewDetails>;
  [
    { key: "leader" as TDistributedTypes, policy: leaderPolicy },
    { key: "worker" as TDistributedTypes, policy: workerPolicy },
  ].forEach((p) => {
    const defaultsBySections = (p.policy.defaults && setSections(p.policy.defaults)) || null;
    const rulesBySections = (p.policy.rules && setSections(p.policy.rules)) || null;

    const flattenPolicy = {
      rules: (rulesBySections && flattenPolicyObject(rulesBySections)) || null,
      defaults: (defaultsBySections && flattenPolicyObject(defaultsBySections)) || null,
      imposedAssets: null,
    };

    const matchedPolicy = matchFlattenedPolicyRulesAndDefaults(flattenPolicy);
    readyState[p.key] = finalizePolicyForView(matchedPolicy, p.key);
  });
  return {
    environment: { ...readyState.leader.environment, ...readyState.worker.environment },
    compute: { ...readyState.leader.compute, ...readyState.worker.compute },
    general: { ...readyState.leader.general, ...readyState.worker.general },
    storage: { ...readyState.leader.storage, ...readyState.worker.storage },
  };
}

// This function is re-aligning the property in the defaults/rules for view policy
// because some sections such as security do not exist in the view page, and because environment properties are not bundled together
function setSections(data: TAllWorkloadPolicyDefaults | TAllWorkloadPolicyRules): Record<keyof IPolicyViewDetails, any> {
  const environment: Record<string, any> = {};
  const general: Record<string, any> = {};
  const { compute, storage, security, ...rest } = data;

  Object.entries({ ...rest, ...security }).forEach(([key, value]) => {
    if (environmentKeySet.has(key as any)) {
      environment[key] = value;
    } else {
      general[key] = value;
    }
  });

  return {
    compute,
    storage,
    environment,
    general,
  };
}

// This function flattens the data until it reaches a value of any type except for Objects that do not have the property "sourceOfRule", if so it goes deeper
function flattenPolicyObject(data: Record<string, any>, prefix?: string): Record<string, any> {
  let objectMap: Record<string, any> = {};

  objectMap = Object.entries(data).reduce((acc, [parameter, dataValue]: [string, any]) => {
    const key = prefix ? `${prefix}.${parameter}` : parameter;

    if (isObjectRule(dataValue)) {
      // if its a rule object you're done.
      if (hasValue(dataValue)) acc[key] = dataValue;
    } else if (isObject(dataValue)) {
      // if its an object but not rule object check whether or not it has attributes or instances.
      // attributes are treated as direct properties of said object, while instances are treated as a new property entirely
      // if it has neither -> flat the object to its values
      if (isAttributesObject(dataValue) || isInstancesObject(dataValue)) {
        if (isAttributesObject(dataValue)) acc = { ...acc, ...flattenPolicyObject(dataValue.attributes, key) };
        if (isInstancesObject(dataValue))
          acc = { ...acc, ...flattenPolicyObject({ instances: dataValue.instances }, key) };
      } else acc = { ...acc, ...flattenPolicyObject(dataValue, key) };
    } else {
      // if its not an object, we take the value directly
      if (hasValue(dataValue)) acc[key] = dataValue;
    }
    return acc;
  }, objectMap);

  return objectMap;
}

// This function matches default data and rules data that shares the same key (post flat process)
function matchFlattenedPolicyRulesAndDefaults(flattedPolicy: Record<string, any>): Record<string, any> {
  const matched: Record<string, any> = {};

  const uniqueKeys = new Set([...Object.keys(flattedPolicy.defaults || {}), ...Object.keys(flattedPolicy.rules || {})]);

  uniqueKeys.forEach((uniqueKey: string) => {
    if (hasValue(flattedPolicy.defaults?.[uniqueKey])) {
      matched[uniqueKey] ??= {};
      matched[uniqueKey].defaults = flattedPolicy.defaults[uniqueKey];
    }
    if (hasValue(flattedPolicy.rules?.[uniqueKey])) {
      matched[uniqueKey] ??= {};
      matched[uniqueKey].rules = flattedPolicy.rules[uniqueKey];
    }
  });

  return matched;
}

// this function finalize the policy before it reaches the view page, bundling it under sections and setting up property name, sourceOfRule, rules and default (post match process)
function finalizePolicyForView(matchedPolicy: Record<string, any>, forceType?: TDistributedTypes): IPolicyViewDetails {
  const finalViewPolicy: IPolicyViewDetails = {
    environment: {},
    storage: {},
    compute: {},
    general: {},
  };

  Object.keys(matchedPolicy).forEach((key) => {
    const location = key.split(".");
    const section = location[0];
    const property = location.slice(1).join(".");
    const keyToSet = forceType ? `${forceType}.${property}` : property;

    const rules = (hasValue(matchedPolicy[key]?.rules) && omit(matchedPolicy[key].rules, ["sourceOfRule"])) || null;
    const defaultVal = matchedPolicy[key]?.defaults;
    const sourceOfRule = matchedPolicy[key]?.rules?.sourceOfRule || null;

    switch (section) {
      case "environment":
        finalViewPolicy.environment[keyToSet] = {
          name: property,
          rules,
          default: defaultVal,
          sourceOfRule,
          replica: forceType,
        };
        break;
      case "compute":
        finalViewPolicy.compute[keyToSet] = {
          name: property,
          rules,
          default: defaultVal,
          sourceOfRule,
          replica: forceType,
        };
        break;
      case "general":
        finalViewPolicy.general[keyToSet] = {
          name: property,
          rules,
          default: defaultVal,
          sourceOfRule,
          replica: forceType,
        };
        break;
      case "storage":
        // Contrary to the other tables storage table is presented with a type
        // so we add the type here and remove the prefix in the name to avoid repetition
        finalViewPolicy.storage[keyToSet] = {
          name: location.slice(2).join("."),
          replica: forceType,
          type: location[1],
          rules,
          default: defaultVal,
          sourceOfRule,
          // type: forceType ? `${forceType}: ${location[1]}` : location[1],
        };
        break;
      default:
        break;
    }
  });

  return finalViewPolicy;
}

// helper functions
function hasValue(val: any): boolean {
  return val !== undefined && val !== null;
}

function isObject(val: any): boolean {
  return typeof val === "object" && !Array.isArray(val);
}

function isInstancesObject(val: any): boolean {
  return isObject(val) && Object.keys(val).includes("instances");
}

function isAttributesObject(val: any): boolean {
  return isObject(val) && Object.keys(val).includes("attributes");
}

function isObjectRule(val: any): boolean {
  return isObject(val) && Object.keys(val).includes("sourceOfRule");
}
