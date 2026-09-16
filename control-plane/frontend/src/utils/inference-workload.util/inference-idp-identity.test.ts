/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { UidGidSource, WorkloadDesiredPhase, type UnifiedWorkload } from "@/swagger-models/workloads-client";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { policyUtil } from "@/utils/policy.util/policy.util";
import { clearWorkloadBlockedFields } from "@/utils/policy.util/policy-field-iterator.util";
import { inferenceWorkloadUtil } from "./inference-workload.util";

const resolvedIdentity = { runAsUid: 2001, runAsGid: 2002, supplementalGroups: "2003,2004" };
const policyIdentity = { runAsUid: 10001, runAsGid: 10002, supplementalGroups: "10003" };
const policy = {
  defaults: { security: { uidGidSource: UidGidSource.FromIdpToken, ...policyIdentity } },
  rules: {
    security: {
      runAsUid: { canEdit: false },
      runAsGid: { canEdit: false },
      supplementalGroups: { canEdit: false },
    },
  },
};

function unifiedWorkload(identity = resolvedIdentity): UnifiedWorkload {
  return {
    meta: {
      name: "original",
      requestedName: "original",
      workloadId: "original-id",
      projectId: "project",
      clusterId: "cluster",
      createdBy: "original-user",
      createdAt: "2026-09-01T00:00:00Z",
      desiredPhase: WorkloadDesiredPhase.Running,
    },
    spec: {
      worker: {
        image: "example-image",
        servingPort: { container: 8080, protocol: "http" },
        autoscaling: { minReplicas: 1, maxReplicas: 1 },
        security: { uidGidSource: UidGidSource.FromIdpToken, hostIpc: true, ...identity },
      },
    },
    policyInfo: { worker: policy },
  };
}

const forms = [
  { name: "custom inference", applyDefaults: inferenceWorkloadUtil.setInferenceUIDefaults },
  { name: "Hugging Face vLLM/TGI", applyDefaults: inferenceWorkloadUtil.setInferenceHuggingFaceUIDefaults },
  { name: "NIM", applyDefaults: inferenceWorkloadUtil.setInferenceNIMUIDefaults },
];

describe("inference creation with IdP identity", () => {
  describe.each(forms)("$name", ({ applyDefaults }) => {
    it.each([
      { flow: "copy", identity: resolvedIdentity },
      { flow: "template", identity: { runAsUid: 3001, runAsGid: 3002, supplementalGroups: "3003" } },
      { flow: "new workload with policy defaults", identity: policyIdentity },
      { flow: "zero and empty defaults", identity: { runAsUid: 0, runAsGid: 0, supplementalGroups: "" } },
    ])("omits hidden identity in a $flow request without changing the form", ({ flow, identity }) => {
      const original = unifiedWorkload(identity);
      const model = applyDefaults(inferenceWorkloadUtil.getHybridInferenceModel(original), true, policy);
      clearWorkloadBlockedFields(model.spec, policy);
      model.name = "new-inference";
      if (flow === "template") model.templateId = "expanded-template";
      model.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(model.spec, policy);
      const before = JSON.stringify(model);

      const request = JSON.parse(JSON.stringify(inferenceWorkloadUtil.prepareInferenceFlatRequest(model)));

      expect(request.name).toBe("new-inference");
      expect(request.spec.security).toEqual({
        uidGidSource: UidGidSource.FromIdpToken,
        hostIpc: true,
        capabilities: [],
      });
      expect(request).not.toHaveProperty("templateId");
      expect(JSON.stringify(model)).toBe(before);
      expect(original.spec?.worker?.security).toMatchObject(identity);
      expect(policy.defaults.security).toMatchObject(policyIdentity);
    });

    it.each([UidGidSource.Custom, UidGidSource.FromTheImage, null, undefined])(
      "preserves the form's identity when the final source is %s",
      (uidGidSource) => {
        const model = applyDefaults(inferenceWorkloadUtil.getHybridInferenceModel(unifiedWorkload()), true, policy);
        model.spec.security = { uidGidSource, ...resolvedIdentity, hostIpc: true };
        const request = JSON.parse(JSON.stringify(inferenceWorkloadUtil.prepareInferenceFlatRequest(model)));
        expect(request.spec.security).toEqual(JSON.parse(JSON.stringify(model.spec.security)));
      },
    );
  });

  it("omits NIM's automatic UID/GID 1000 for a new IdP workload", () => {
    const original = unifiedWorkload();
    original.spec = { worker: { ...original.spec?.worker, security: { uidGidSource: UidGidSource.FromIdpToken } } };
    const model = inferenceWorkloadUtil.setInferenceNIMUIDefaults(
      inferenceWorkloadUtil.getHybridInferenceModel(original),
      true,
    );
    expect(model.spec.security).toMatchObject({ runAsUid: 1000, runAsGid: 1000 });
    const request = JSON.parse(JSON.stringify(inferenceWorkloadUtil.prepareInferenceFlatRequest(model)));
    expect(request.spec.security).toEqual({ uidGidSource: UidGidSource.FromIdpToken, capabilities: [] });
  });

  it.each([
    [UidGidSource.FromIdpToken, UidGidSource.FromIdpToken],
    [UidGidSource.FromIdpToken, UidGidSource.Custom],
    [UidGidSource.Custom, UidGidSource.FromIdpToken],
    [UidGidSource.Custom, UidGidSource.FromTheImage],
  ])("normalizes distributed leader=%s and worker=%s independently", (leaderSource, workerSource) => {
    const model = inferenceWorkloadUtil.getDistributedInferenceModel({
      ...unifiedWorkload(),
      spec: {
        worker: {
          servingPort: { port: 8080, protocol: "http" },
          leader: {
            image: "leader-image",
            security: { uidGidSource: leaderSource, hostIpc: true, ...resolvedIdentity },
          },
          worker: { image: "worker-image", security: { uidGidSource: workerSource, ...policyIdentity } },
        },
      },
    });
    const initialized = inferenceWorkloadUtil.setDistributedUIDefaults(model, true);
    const before = JSON.stringify(initialized);
    const request = JSON.parse(
      JSON.stringify(inferenceWorkloadUtil.prepareDistributedInferenceFlatRequest(initialized, 1)),
    );
    expect(request.spec.leader.security).toEqual({
      uidGidSource: leaderSource,
      hostIpc: true,
      capabilities: [],
      ...(leaderSource === UidGidSource.FromIdpToken ? {} : resolvedIdentity),
    });
    expect(request.spec.worker.security).toEqual({
      uidGidSource: workerSource,
      capabilities: [],
      ...(workerSource === UidGidSource.FromIdpToken ? {} : policyIdentity),
    });
    expect(JSON.stringify(initialized)).toBe(before);
  });

  it("does not clear stored template identity when saving a template", () => {
    const model = inferenceWorkloadUtil.getHybridInferenceModel(unifiedWorkload());
    const template = inferenceWorkloadUtil.prepareInferenceTemplateFlatRequest(model);
    expect(template.spec?.security).toMatchObject({ uidGidSource: UidGidSource.FromIdpToken, ...resolvedIdentity });
  });

  it.each([null, undefined, {}])("preserves absent security sections: %j", (security) => {
    expect(workloadHybridUtil._getCleanedSecurityFieldsFromUI(security)).toBeUndefined();
  });
});
