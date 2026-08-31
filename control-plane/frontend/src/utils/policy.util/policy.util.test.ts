/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ISelectOption } from "@/models/global.model";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import {
  getDistributedPolicyViewDetails,
  getInferenceDistributedPolicyViewDetails,
  getPolicyViewDetails,
  type IPolicyViewDetails,
} from "./policy-view.util";
import { policyUtilMock } from "./policy.util.mock";
import { policyUtil } from "./policy.util";
import { errorMessages } from "@/common/error-message.constant";
import { ScopeType } from "@/swagger-models/policy-service-client";
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import { describe, expect, it, test } from "vitest";

// this will require additional testing once view policy has matured
describe("policy-view.util", () => {
  test("should return the expected policy structure - simple", () => {
    const expectedResult = {
      compute: {},
      general: {},
      storage: {},
      environment: {
        allowPrivilegeEscalation: {
          name: "allowPrivilegeEscalation",
          rules: {
            required: true,
            canEdit: true,
          },
          sourceOfRule: {
            scope: "project",
            projectId: 3,
          },
          default: true,
        },
      },
    };
    const actualResult = getPolicyViewDetails(policyUtilMock.simplePolicy);
    expect(actualResult).toEqual(expectedResult);
  });
  test("should return the expected policy structure - with no rules details", () => {
    const expectedResult = {
      compute: {},
      general: {},
      storage: {},
      environment: {
        allowPrivilegeEscalation: {
          name: "allowPrivilegeEscalation",
          default: true,
          rules: null,
          sourceOfRule: null,
          type: undefined,
        },
      },
    };
    const actualResult = getPolicyViewDetails(policyUtilMock.policyWithNoRules);
    expect(actualResult).toEqual(expectedResult);
  });
  test("should return the expected policy structure - with no defaults", () => {
    const expectedResult = {
      compute: {},
      general: {},
      storage: {},
      environment: {
        allowPrivilegeEscalation: {
          default: undefined,
          name: "allowPrivilegeEscalation",
          rules: {
            required: true,
            canEdit: true,
          },
          sourceOfRule: {
            projectId: 3,
            scope: "project",
          },
          type: undefined,
        },
      },
    };
    const actualResult = getPolicyViewDetails(policyUtilMock.policyWithNoDefaults);
    expect(actualResult).toEqual(expectedResult);
  });

  test("should return the expected policy structure - with no defaults", () => {
    const expectedResult = {
      general: {},
      storage: {},
      environment: {
        "environmentVariables.instances": {
          type: undefined,
          name: "environmentVariables.instances",
          rules: {
            canAdd: true,
            locked: ["a", "b"],
          },
          sourceOfRule: {
            projectId: 3,
            scope: "project",
          },
          default: [
            { name: "a", value: "aaa" },
            { name: "b", value: "bbb" },
            { name: "c", value: "ccc" },
          ],
        },
      },
      compute: {
        gpuMemoryRequest: {
          type: undefined,
          default: undefined,
          name: "gpuMemoryRequest",
          sourceOfRule: {
            projectId: 3,
            scope: "project",
          },
          rules: {
            min: "4G",
            max: "8G",
          },
        },
        gpuDevicesRequest: {
          type: undefined,
          default: undefined,
          name: "gpuDevicesRequest",
          sourceOfRule: {
            projectId: 3,
            scope: "project",
          },
          rules: {
            min: 10,
            max: 20,
          },
        },
      },
    };
    const actualResult = getPolicyViewDetails(policyUtilMock.complexPolicy);
    expect(actualResult).toEqual(expectedResult);
  });

  test("should return the expected policy structure - distributed", () => {
    const expectedResult: IPolicyViewDetails = {
      compute: {},
      general: {},
      storage: {},
      environment: {
        "master.command": {
          name: "command",
          rules: null,
          default: "cmd-default-master",
          sourceOfRule: null,
          replica: "master",
        },
        "master.image": {
          name: "image",
          rules: {
            options: [
              {
                displayed: "master-image-1",
                value: "image-default-master",
              },
              {
                displayed: "master-image-2",
                value: "image-default-master-2",
              },
            ],
          },
          default: "image-default-master",
          sourceOfRule: {
            clusterId: "clusteruuid-123123123",
            departmentId: "44",
            projectId: 12,
            scope: "project",
          },
          replica: "master",
        },
        "worker.command": {
          name: "command",
          rules: null,
          default: "cmd-default-worker",
          sourceOfRule: null,
          replica: "worker",
        },
        "worker.image": {
          name: "image",
          rules: null,
          default: "image-default-worker",
          sourceOfRule: null,
          replica: "worker",
        },
        "worker.supplementalGroups": {
          name: "supplementalGroups",
          rules: {
            options: [
              {
                displayed: "display-this-supplementalGroups",
                value: "supplementalGroups-opt-1",
              },
            ],
            canEdit: true,
            required: true,
          },
          default: "supplementalGroups-opt-1",
          sourceOfRule: {
            clusterId: "clusteruuid-123123123",
            departmentId: "2",
            scope: "department",
          },
          replica: "worker",
        },
        "worker.capabilities": {
          name: "capabilities",
          rules: {
            canEdit: true,
            required: true,
          },
          default: ["BLOCK_SUSPEND", "DAC_OVERRIDE", "LINUX_IMMUTABLE", "AUDIT_WRITE", "IPC_LOCK"],
          sourceOfRule: {
            clusterId: "clusteruuid-123123123",
            departmentId: "2",
            scope: "department",
          },
          replica: "worker",
        },
      },
    };
    const actualResult = getDistributedPolicyViewDetails(policyUtilMock.distributedPolicy);
    expect(actualResult).toEqual(expectedResult);
  });

  test("should categorize privileged field under environment section", () => {
    const policyWithPrivileged = {
      defaults: {
        security: {
          privileged: false,
        },
      },
      rules: {
        security: {
          privileged: {
            sourceOfRule: {
              scope: ScopeType.System,
            },
            canEdit: false,
          },
        },
      },
    };

    const expectedResult = {
      compute: {},
      general: {},
      storage: {},
      environment: {
        privileged: {
          name: "privileged",
          rules: {
            canEdit: false,
          },
          default: false,
          sourceOfRule: {
            scope: ScopeType.System,
          },
          type: undefined,
        },
      },
    };

    const actualResult = getPolicyViewDetails(policyWithPrivileged);
    expect(actualResult).toEqual(expectedResult);
  });

  test("should categorize privileged field under environment section for distributed policy", () => {
    const distributedPolicyWithPrivileged = {
      defaults: {
        master: {
          security: {
            privileged: false,
          },
        },
        worker: {
          security: {
            privileged: false,
          },
        },
      },
      rules: {
        master: {
          security: {
            privileged: {
              sourceOfRule: {
                scope: ScopeType.System,
              },
              canEdit: false,
            },
          },
        },
        worker: {
          security: {
            privileged: {
              sourceOfRule: {
                scope: ScopeType.System,
              },
              canEdit: false,
            },
          },
        },
      },
    };

    const expectedResult = {
      compute: {},
      general: {},
      storage: {},
      environment: {
        "master.privileged": {
          name: "privileged",
          rules: {
            canEdit: false,
          },
          default: false,
          sourceOfRule: {
            scope: ScopeType.System,
          },
          replica: "master",
        },
        "worker.privileged": {
          name: "privileged",
          rules: {
            canEdit: false,
          },
          default: false,
          sourceOfRule: {
            scope: ScopeType.System,
          },
          replica: "worker",
        },
      },
    };

    const actualResult = getDistributedPolicyViewDetails(distributedPolicyWithPrivileged);
    expect(actualResult).toEqual(expectedResult);
  });

  test("getInferenceDistributedPolicyViewDetails merges leader and worker from policy.worker with leader/worker keys", () => {
    const actualResult = getInferenceDistributedPolicyViewDetails(policyUtilMock.inferenceDistributedPolicy);

    expect(actualResult.environment).toHaveProperty("leader.image");
    expect(actualResult.environment).toHaveProperty("worker.image");
    expect(actualResult.environment["leader.image"].replica).toBe("leader");
    expect(actualResult.environment["worker.image"].replica).toBe("worker");
    expect(actualResult.environment["leader.image"].default).toBe("leader-default-image");
    expect(actualResult.environment["worker.image"].default).toBe("worker-default-image");
    expect(actualResult.environment["leader.allowPrivilegeEscalation"].default).toBe(true);
    expect(actualResult.environment["worker.allowPrivilegeEscalation"].default).toBe(true);
  });
});
describe("getOptionsWithRules", () => {
  const sampleOptions: ISelectOption[] = [
    { label: "Option 1", value: "1" },
    { label: "Option 2", value: "2" },
    { label: "Option 3", value: "3" },
  ];

  const sampleRules = [
    { value: "1", displayed: "Custom Option 1" },
    { value: "3", displayed: "Custom Option 3" },
  ];

  it("should return original options if ruleOptions is null or empty", () => {
    const result = policyUtil.getOptionsWithRules(sampleOptions, null, true);
    expect(result).toEqual(sampleOptions);

    const resultWithEmptyRules = policyUtil.getOptionsWithRules(sampleOptions, [], true);
    expect(resultWithEmptyRules).toEqual(sampleOptions);
  });

  it("should disable options not present in ruleOptions when isClosedList is true", () => {
    const result = policyUtil.getOptionsWithRules(sampleOptions, sampleRules, true);

    expect(result).toEqual([
      { label: "Custom Option 1", value: "1" },
      { label: "Option 2", value: "2", disable: true, disabledTooltip: errorMessages.CANT_BE_SELECTED_POLICY },
      { label: "Custom Option 3", value: "3" },
    ]);
  });

  it("should return only ruleOptions when isClosedList is false", () => {
    const result = policyUtil.getOptionsWithRules(sampleOptions, sampleRules, false);

    expect(result).toEqual([
      { value: "1", label: "Custom Option 1" },
      { value: "3", label: "Custom Option 3" },
    ]);
  });

  it("should handle ruleOptions with a displayed label if it exist in the list", () => {
    const ruleWithoutDisplayed = [{ value: "2", displayed: null }];

    const result = policyUtil.getOptionsWithRules(sampleOptions, ruleWithoutDisplayed, false);

    expect(result).toEqual([{ value: "2", label: "Option 2" }]);
  });

  it("should handle ruleOptions without a displayed label in open list when it is not in the list", () => {
    const ruleWithoutDisplayed = [{ value: "55", displayed: null }];

    const result = policyUtil.getOptionsWithRules(sampleOptions, ruleWithoutDisplayed, false);

    expect(result).toEqual([{ value: "55", label: "55" }]);
  });
});

describe("cleanupItemizeExcludeDuplicates", () => {
  it("should remove duplicates and keep the non-excluded item", () => {
    const items = [
      { name: "item1", exclude: true },
      { name: "item1", exclude: false },
      { name: "item2", exclude: false },
    ];

    const result = policyUtil.cleanupItemizeExcludeDuplicates(items, "name");

    expect(result).toEqual([
      { name: "item1", exclude: false },
      { name: "item2", exclude: false },
    ]);
  });

  it("should not prevent duplicates which are not excluded", () => {
    const items = [
      { name: "item1", exclude: false },
      { name: "item1", exclude: false },
    ];

    const result = policyUtil.cleanupItemizeExcludeDuplicates(items, "name");

    expect(result).toEqual([
      { name: "item1", exclude: false },
      { name: "item1", exclude: false },
    ]);
  });

  it("should return empty array if no items are provided", () => {
    const result = policyUtil.cleanupItemizeExcludeDuplicates([], "name");
    expect(result).toEqual([]);
  });
});

describe("cleanupLockedItems", () => {
  it("should remove items with keys found in locked array", () => {
    const items = [{ name: "item1" }, { name: "item2" }, { name: "item3" }];
    const locked = ["item1", "item3"];

    const result = policyUtil.cleanupLockedItems(items, "name", locked);

    expect(result).toEqual([{ name: "item2" }]);
  });

  it("should return the same array if no locked items are present", () => {
    const items = [{ name: "item1" }, { name: "item2" }];
    const locked: string[] = [];

    const result = policyUtil.cleanupLockedItems(items, "name", locked);

    expect(result).toEqual(items);
  });

  it("should return an empty array if all items are locked", () => {
    const items = [{ name: "item1" }, { name: "item2" }];
    const locked = ["item1", "item2"];

    const result = policyUtil.cleanupLockedItems(items, "name", locked);

    expect(result).toEqual([]);
  });
});

describe("cleanupFlatItemizeBeforeSubmission", () => {
  it("should remove locked annotations, labels, env vars, and tolerations", () => {
    const flatSpec: Partial<IUISuperWorkloadSpec> = {
      annotations: [
        { name: "anno1", value: "v1" },
        { name: "anno2", value: "v2" },
      ],
      labels: [
        { name: "label1", value: "v1" },
        { name: "label2", value: "v2" },
      ],
      environmentVariables: [
        { name: "env1", value: "v1" },
        { name: "env2", value: "v2" },
      ],
      tolerations: [{ name: "tol1" }, { name: "tol2" }],
    };

    const policy: Partial<UnifiedPolicyInfoPerReplicaType> = {
      rules: {
        annotations: { instances: { locked: ["anno1"] } },
        labels: { instances: { locked: ["label2"] } },
        environmentVariables: { instances: { locked: ["env1"] } },
        tolerations: { instances: { locked: ["tol2"] } },
      },
    };

    const result = policyUtil.cleanupFlatItemizeBeforeSubmission(
      flatSpec as IUISuperWorkloadSpec,
      policy as UnifiedPolicyInfoPerReplicaType,
    );

    expect(result.annotations).toEqual([{ name: "anno2", value: "v2" }]);
    expect(result.labels).toEqual([{ name: "label1", value: "v1" }]);
    expect(result.environmentVariables).toEqual([{ name: "env2", value: "v2" }]);
    expect(result.tolerations).toEqual([{ name: "tol1" }]);
  });

  it("should remove locked extended resources", () => {
    const flatSpec: Partial<IUISuperWorkloadSpec> = {
      compute: {
        extendedResources: [
          { resource: "res1", quantity: "1" },
          { resource: "res2", quantity: "2" },
        ],
      },
    };

    const policy: Partial<UnifiedPolicyInfoPerReplicaType> = {
      rules: {
        compute: {
          extendedResources: { instances: { locked: ["res1"] } },
        },
      },
    };

    const result = policyUtil.cleanupFlatItemizeBeforeSubmission(
      flatSpec as IUISuperWorkloadSpec,
      policy as UnifiedPolicyInfoPerReplicaType,
    );

    expect(result.compute?.extendedResources).toEqual([{ resource: "res2", quantity: "2" }]);
  });

  it("should remove locked storage items across all storage types", () => {
    const flatSpec = {
      storage: {
        pvc: [{ name: "pvc1" }, { name: "pvc2" }],
        nfs: [{ name: "nfs1" }, { name: "nfs2" }],
        s3: [{ name: "s3-locked" }, { name: "s3-kept" }],
        hostPath: [{ name: "hp1" }],
        git: [{ name: "git1" }, { name: "git2" }],
        configMapVolume: [{ name: "cm1" }],
        secretVolume: [{ name: "sec1" }, { name: "sec2" }],
        emptyDirVolume: [{ name: "ed1" }],
      },
    } as unknown as IUISuperWorkloadSpec;

    const policy: Partial<UnifiedPolicyInfoPerReplicaType> = {
      rules: {
        storage: {
          pvc: { instances: { locked: ["pvc1"] } },
          nfs: { instances: { locked: ["nfs2"] } },
          s3: { instances: { locked: ["s3-locked"] } },
          hostPath: { instances: { locked: ["hp1"] } },
          git: { instances: { locked: ["git1"] } },
          configMapVolume: { instances: { locked: ["cm1"] } },
          secretVolume: { instances: { locked: ["sec1"] } },
          emptyDirVolume: { instances: { locked: ["ed1"] } },
        },
      },
    };

    const result = policyUtil.cleanupFlatItemizeBeforeSubmission(flatSpec, policy as UnifiedPolicyInfoPerReplicaType);

    expect(result.storage?.pvc).toEqual([{ name: "pvc2" }]);
    expect(result.storage?.nfs).toEqual([{ name: "nfs1" }]);
    expect(result.storage?.s3).toEqual([{ name: "s3-kept" }]);
    expect(result.storage?.hostPath).toEqual([]);
    expect(result.storage?.git).toEqual([{ name: "git2" }]);
    expect(result.storage?.configMapVolume).toEqual([]);
    expect(result.storage?.secretVolume).toEqual([{ name: "sec2" }]);
    expect(result.storage?.emptyDirVolume).toEqual([]);
  });

  it("should handle missing/undefined fields and null policy gracefully", () => {
    const flatSpec = {} as IUISuperWorkloadSpec;

    const result = policyUtil.cleanupFlatItemizeBeforeSubmission(flatSpec, null);

    expect(result.annotations).toEqual([]);
    expect(result.labels).toEqual([]);
    expect(result.environmentVariables).toEqual([]);
    expect(result.tolerations).toEqual([]);
  });

  it("should deduplicate excluded items before removing locked items", () => {
    const flatSpec: Partial<IUISuperWorkloadSpec> = {
      annotations: [
        { name: "anno1", value: "v-excluded", exclude: true },
        { name: "anno1", value: "v-kept", exclude: false },
        { name: "anno2", value: "v2", exclude: false },
      ],
    };

    const policy: Partial<UnifiedPolicyInfoPerReplicaType> = {
      rules: {
        annotations: { instances: { locked: ["anno2"] } },
      },
    };

    const result = policyUtil.cleanupFlatItemizeBeforeSubmission(
      flatSpec as IUISuperWorkloadSpec,
      policy as UnifiedPolicyInfoPerReplicaType,
    );

    expect(result.annotations).toEqual([{ name: "anno1", value: "v-kept", exclude: false }]);
  });

  it("should keep all items when no policy is provided", () => {
    const flatSpec: Partial<IUISuperWorkloadSpec> = {
      annotations: [{ name: "anno1", value: "v1" }],
      labels: [{ name: "label1", value: "v1" }],
      environmentVariables: [{ name: "env1", value: "v1" }],
      tolerations: [{ name: "tol1" }],
      compute: {
        extendedResources: [{ resource: "res1", quantity: "1" }],
      },
    };

    const result = policyUtil.cleanupFlatItemizeBeforeSubmission(flatSpec as IUISuperWorkloadSpec, null);

    expect(result.annotations).toEqual([{ name: "anno1", value: "v1" }]);
    expect(result.labels).toEqual([{ name: "label1", value: "v1" }]);
    expect(result.environmentVariables).toEqual([{ name: "env1", value: "v1" }]);
    expect(result.tolerations).toEqual([{ name: "tol1" }]);
    expect(result.compute?.extendedResources).toEqual([{ resource: "res1", quantity: "1" }]);
  });
});
