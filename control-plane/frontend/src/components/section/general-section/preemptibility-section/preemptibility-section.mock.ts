/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { PreemptibilityRules, PreemptibilityOptionsOptionsInner } from "@/swagger-models/workloads-client";
import { ScopeType } from "@/swagger-models/authorization-client";
import { Preemptibility } from "@/models/workload.model";

export const mockPolicyRules = {
  editable: {
    sourceOfRule: null,
    required: false,
    canEdit: true,
    options: [
      {
        value: Preemptibility.Preemptible,
        displayed: "Preemptible Workload",
      },
      {
        value: Preemptibility.NonPreemptible,
        displayed: "Non-Preemptible Workload",
      },
    ] as PreemptibilityOptionsOptionsInner[],
  } as PreemptibilityRules,

  nonEditable: {
    sourceOfRule: {
      scope: ScopeType.Cluster,
      clusterId: "test-cluster-id",
      isFallback: false,
    },
    required: true,
    canEdit: false,
    options: [
      {
        value: Preemptibility.NonPreemptible,
        displayed: "Fixed Non-Preemptible",
      },
    ] as PreemptibilityOptionsOptionsInner[],
  } as PreemptibilityRules,

  unrestricted: {
    sourceOfRule: null,
    required: false,
    canEdit: true,
    options: [],
  } as PreemptibilityRules,

  limitedOptions: {
    sourceOfRule: null,
    required: false,
    canEdit: true,
    options: [
      {
        value: Preemptibility.Preemptible,
      },
    ] as PreemptibilityOptionsOptionsInner[],
  } as PreemptibilityRules,
};

export const mockDefaultFactoryProps = {
  modelValue: Preemptibility.Preemptible,
  disabled: false,
  policyRules: mockPolicyRules.editable,
  showQuotaWarning: false,
};

export const preemptibilitySectionMock = {
  mockPolicyRules,
  mockDefaultFactoryProps,
};
