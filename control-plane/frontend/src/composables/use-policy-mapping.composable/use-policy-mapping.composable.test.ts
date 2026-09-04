/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { usePolicyMapping } from "./use-policy-mapping.composable";
import { gvkPolicyService } from "@/services/control-plane/gvk-policy.service/gvk-policy.service";
import type { Breakdown, GetMappingMinimalResp, GVK } from "@/swagger-models/policy-service-client";

vi.mock("@/services/control-plane/gvk-policy.service/gvk-policy.service", () => ({
  gvkPolicyService: {
    getMappingMinimal: vi.fn(),
    getMappingFieldInfo: vi.fn(),
  },
}));

describe("usePolicyMapping — field metadata", () => {
  const workloadType: GVK = { group: "batch", version: "v1", kind: "CronJob" };

  beforeEach(() => vi.clearAllMocks());

  it("loads the full field breakdown once and resolves selector parameters from the cache", async () => {
    vi.mocked(gvkPolicyService.getMappingMinimal).mockResolvedValue({
      fields: {
        "ports-HostPort": {
          allAppContainers: { allPorts: "all ports", portByContainerPort: "port by container port" },
          appContainerByName: { allPorts: "all ports", portByContainerPort: "port by container port" },
        },
      },
      specSelectors: {
        allAppContainers: "all app containers",
        appContainerByName: "app container with a given name",
      },
    } as GetMappingMinimalResp);
    vi.mocked(gvkPolicyService.getMappingFieldInfo).mockResolvedValue({
      breakdown: {
        fields: {
          "ports-HostPort": [
            {
              specSelectors: ["allAppContainers", "appContainerByName"],
              formatter: { builtin: "KeyValue" },
              type: "integer",
              itemSelectors: {
                allPorts: { query: "(.ports // [])[]" },
                portByContainerPort: {
                  query: "(.ports // [])[] | select(.containerPort == $ARGS.named.containerPort)",
                  argsInfo: { containerPort: "integer" },
                },
              },
              possibleEnforcements: ["set", "items"],
            },
          ],
        },
        specSelectors: {
          appContainerByName: {
            query: ".containers[] | select(.name == $ARGS.named.containerName)",
            argsInfo: { containerName: "string" },
          },
        },
      },
    } as Breakdown);
    const mapping = usePolicyMapping();

    await mapping.loadMinimal(workloadType);
    await mapping.loadFieldInfo("ports-HostPort");
    const specInfo = await mapping.resolveFieldInfo("ports-HostPort", "appContainerByName", null);
    const itemInfo = await mapping.resolveFieldInfo("ports-HostPort", "appContainerByName", "portByContainerPort");

    expect(gvkPolicyService.getMappingFieldInfo).toHaveBeenCalledTimes(1);
    expect(gvkPolicyService.getMappingFieldInfo).toHaveBeenCalledWith({
      workloadType,
      fieldsRef: [{ fieldName: "ports-HostPort" }],
    });
    expect(specInfo?.specSelectorParams).toEqual([
      { name: "containerName", type: "string", enumValues: undefined, required: false },
    ]);
    expect(specInfo?.setDefaultGroupingArg).toBeNull();
    expect(specInfo?.itemSelectorParams).toEqual([]);
    expect(itemInfo?.itemSelectorParams).toEqual([
      { name: "containerPort", type: "integer", enumValues: undefined, required: false },
    ]);
  });

  it.each([
    {
      field: "environmentVariables-Value",
      formatter: {
        setExpr: ".env[$ARGS.named.name] = $ARGS.named.value",
        arrayKey: { argName: "name" },
        argsInfo: { name: "string", value: "string" },
      },
      expectedGroupingArg: "name",
    },
    {
      field: "annotations-value",
      formatter: { builtin: "KeyValueMap" },
      expectedGroupingArg: "key",
    },
  ])("derives the Set/Default grouping target from $field formatter metadata", async (testCase) => {
    const gvk: GVK = { group: "batch", version: "v1", kind: testCase.field };
    vi.mocked(gvkPolicyService.getMappingMinimal).mockResolvedValue({
      fields: { [testCase.field]: { podMetadata: { "": "" } } },
      specSelectors: { podMetadata: "pod metadata" },
    } as GetMappingMinimalResp);
    vi.mocked(gvkPolicyService.getMappingFieldInfo).mockResolvedValue({
      breakdown: {
        fields: {
          [testCase.field]: [
            {
              specSelectors: ["podMetadata"],
              formatter: testCase.formatter,
              type: "string",
              possibleEnforcements: ["set", "default"],
            },
          ],
        },
        specSelectors: { podMetadata: { query: ".metadata" } },
      },
    } as Breakdown);
    const mapping = usePolicyMapping();

    await mapping.loadMinimal(gvk);
    const info = await mapping.resolveFieldInfo(testCase.field, "podMetadata", null);

    expect(info?.setDefaultGroupingArg).toBe(testCase.expectedGroupingArg);
  });
});
