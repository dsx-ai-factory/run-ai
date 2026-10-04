/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import {
  UidGidSource,
  WorkloadDesiredPhase,
  type UnifiedWorkload,
  type WorkspaceSpecSpecAllOfSecurity,
} from "@/swagger-models/workloads-client";
import { policyUtil } from "@/utils/policy.util/policy.util";
import { workloadHybridUtil } from "./workload-hybrid.util";

const policyIdentity = { runAsUid: 10001, runAsGid: 10002, supplementalGroups: "10003" };
const securityWithoutIdentity: WorkspaceSpecSpecAllOfSecurity = {
  uidGidSource: UidGidSource.FromIdpToken,
  allowPrivilegeEscalation: true,
  runAsNonRoot: true,
  hostIpc: false,
};
const resolvedSecurity: WorkspaceSpecSpecAllOfSecurity = {
  ...securityWithoutIdentity,
  runAsUid: 2001,
  runAsGid: 2002,
  supplementalGroups: "2002,2003,2004,2005,2006",
};

function copiedWorkload(canEdit = false) {
  return {
    meta: {
      name: "original-workload",
      requestedName: "original-workload",
      workloadId: "original-id",
      projectId: "project",
      clusterId: "cluster",
      createdBy: "user",
      createdAt: "2026-09-08T10:00:00Z",
      desiredPhase: WorkloadDesiredPhase.Running,
    },
    spec: { worker: { image: "ubuntu:22.04", security: { ...resolvedSecurity } } },
    policyInfo: {
      worker: {
        defaults: { security: { uidGidSource: UidGidSource.FromIdpToken, ...policyIdentity } },
        rules: {
          security: {
            uidGidSource: { canEdit: false },
            runAsUid: { canEdit },
            runAsGid: { canEdit },
            supplementalGroups: { canEdit },
          },
        },
      },
    },
  } satisfies UnifiedWorkload;
}

const forms = [
  {
    name: "workspace",
    getModel: workloadHybridUtil.getHybridWorkspaceModel,
    prepare: workloadHybridUtil.prepareWorkloadFlatRequest,
  },
  {
    name: "training",
    getModel: workloadHybridUtil.getHybridTrainingModel,
    prepare: workloadHybridUtil.prepareTrainingFlatRequest,
  },
  {
    name: "distributed",
    getModel: workloadHybridUtil.getHybridDistributedModel,
    prepare: workloadHybridUtil.prepareDistributedFlatRequest,
  },
];

describe("hybrid workload submission with IdP identity", () => {
  describe.each(forms)("$name submission", ({ getModel, prepare }) => {
    it.each([false, true])("omits resolved identity with canEdit=%s only in the outgoing request", (canEdit) => {
      const unified = copiedWorkload(canEdit);
      const model = getModel(unified);
      model.name = "copy";
      model.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(model.spec, unified.policyInfo.worker);
      expect(model.spec.security).toMatchObject(resolvedSecurity);
      const request = JSON.parse(JSON.stringify(prepare(model)));

      expect(request.name).toBe("copy");
      expect(request.spec?.security).toEqual(securityWithoutIdentity);
      expect(request).not.toHaveProperty("policy");
      expect(request).not.toHaveProperty("templateId");
      expect(request).not.toHaveProperty("assets");
      expect(request).not.toHaveProperty("masterAssets");
      expect(model.spec.security).toMatchObject(resolvedSecurity);
      expect(unified.spec.worker.security).toEqual(resolvedSecurity);
      expect(unified.policyInfo.worker.defaults.security).toEqual({
        uidGidSource: UidGidSource.FromIdpToken,
        ...policyIdentity,
      });
    });

    it.each([
      { name: "policy defaults", identity: policyIdentity },
      { name: "template values", identity: resolvedSecurity },
      { name: "zero and empty values", identity: { runAsUid: 0, runAsGid: 0, supplementalGroups: "" } },
    ])("omits $name when submitting a new IdP workload", ({ identity }) => {
      const { policyInfo } = copiedWorkload();
      const security = { ...securityWithoutIdentity, ...identity };
      const model = getModel({ spec: { worker: { image: "ubuntu:22.04", security } }, policyInfo });
      const request = JSON.parse(JSON.stringify(prepare(model)));
      expect(request.spec?.security).toEqual(securityWithoutIdentity);
      expect(model.spec.security).toMatchObject(security);
    });

    it.each([UidGidSource.Custom, UidGidSource.FromTheImage, null, undefined])(
      "preserves copied identity when uidGidSource is %s",
      (uidGidSource) => {
        const unified = copiedWorkload();
        unified.spec.worker.security.uidGidSource = uidGidSource;
        const model = getModel(unified);
        expect(prepare(model).spec?.security).toMatchObject({
          ...resolvedSecurity,
          uidGidSource,
        });
      },
    );

    it.each([
      [UidGidSource.FromIdpToken, UidGidSource.Custom],
      [UidGidSource.Custom, UidGidSource.FromIdpToken],
    ])("honors the source changed from %s to %s before submission", (initialSource, submittedSource) => {
      const unified = copiedWorkload();
      unified.spec.worker.security.uidGidSource = initialSource;
      const model = getModel(unified);
      model.spec.security = { ...resolvedSecurity, uidGidSource: submittedSource };
      const request = JSON.parse(JSON.stringify(prepare(model)));
      expect(request.spec?.security).toEqual({
        ...(submittedSource === UidGidSource.FromIdpToken ? securityWithoutIdentity : resolvedSecurity),
        uidGidSource: submittedSource,
      });
      expect(model.spec.security).toEqual({ ...resolvedSecurity, uidGidSource: submittedSource });
    });
  });

  it("omits resolved identity when the project has no policy defaults", () => {
    const unified: UnifiedWorkload = { ...copiedWorkload(), policyInfo: null };
    const model = workloadHybridUtil.getHybridWorkspaceModel(unified);
    const request = JSON.parse(JSON.stringify(workloadHybridUtil.prepareWorkloadFlatRequest(model)));
    expect(request.spec?.security).toEqual({
      uidGidSource: UidGidSource.FromIdpToken,
      allowPrivilegeEscalation: true,
      runAsNonRoot: true,
      hostIpc: false,
    });
  });

  it("omits identity fields even when the policy has zero, null, or empty defaults", () => {
    const unified: UnifiedWorkload = {
      ...copiedWorkload(),
      policyInfo: { worker: { defaults: { security: { runAsUid: 0, runAsGid: null, supplementalGroups: "" } } } },
    };
    const model = workloadHybridUtil.getHybridWorkspaceModel(unified);
    const request = workloadHybridUtil.prepareWorkloadFlatRequest(model);
    const security = JSON.parse(JSON.stringify(request.spec?.security));
    expect(security).toEqual(securityWithoutIdentity);
  });

  it.each([UidGidSource.FromIdpToken, UidGidSource.Custom])(
    "cleans each distributed replica according to its source (master=%s)",
    (uidGidSource) => {
      const original = copiedWorkload();
      const masterIdentity = { runAsUid: 20001, runAsGid: 20002, supplementalGroups: "20003" };
      const unified: UnifiedWorkload = {
        ...original,
        spec: {
          ...original.spec,
          masterSpecSameAsWorker: false,
          master: { image: "ubuntu:22.04", security: { ...resolvedSecurity, uidGidSource } },
        },
        policyInfo: { ...original.policyInfo, master: { defaults: { security: masterIdentity } } },
      };
      const model = workloadHybridUtil.getHybridDistributedModel(unified, false, true);
      const request = JSON.parse(JSON.stringify(workloadHybridUtil.prepareDistributedFlatRequest(model)));
      expect(request.spec?.security).toEqual(securityWithoutIdentity);
      expect(request.masterSpec?.security).toEqual({
        ...(uidGidSource === UidGidSource.FromIdpToken ? securityWithoutIdentity : resolvedSecurity),
        uidGidSource,
      });
      const masterModel = workloadHybridUtil.getHybridDistributedMasterModel(unified);
      expect(masterModel.spec.security).toEqual(model.masterSpec?.security);
      expect(model.masterSpec?.security).toMatchObject({ ...resolvedSecurity, uidGidSource });
      expect(unified.spec?.master?.security).toEqual({ ...resolvedSecurity, uidGidSource });
    },
  );

  it("keeps a distributed copy without a master", () => {
    const unified = copiedWorkload();
    const model = workloadHybridUtil.getHybridDistributedModel(unified, true, true);
    const request = JSON.parse(JSON.stringify(workloadHybridUtil.prepareDistributedFlatRequest(model)));
    expect(model.masterSpec).toBeUndefined();
    expect(request.masterSpec).toBeUndefined();
    expect(request.spec?.security).toEqual(securityWithoutIdentity);
    expect(model.spec.security).toMatchObject(resolvedSecurity);
  });

  it("uses the worker identity when the master shares its spec", () => {
    const original = copiedWorkload();
    const unified: UnifiedWorkload = {
      ...original,
      spec: { ...original.spec, master: original.spec.worker, masterSpecSameAsWorker: true },
      policyInfo: {
        ...original.policyInfo,
        master: { defaults: { security: { runAsUid: 20001, runAsGid: 20002, supplementalGroups: "20003" } } },
      },
    };
    const model = workloadHybridUtil.getHybridDistributedModel(unified, false, true);
    const request = JSON.parse(JSON.stringify(workloadHybridUtil.prepareDistributedFlatRequest(model)));
    expect(request.masterSpecSameAsWorker).toBe(true);
    expect(request.masterSpec).toBeUndefined();
    expect(request.spec?.security).toEqual(securityWithoutIdentity);
    expect(unified.spec?.master?.security).toEqual(resolvedSecurity);
  });

  it.each<UnifiedWorkload>([{}, { spec: null }, { spec: { worker: { security: null } } }])(
    "handles absent security sections: %j",
    (unified) => {
      const model = workloadHybridUtil.getHybridWorkspaceModel({ ...unified, meta: copiedWorkload().meta });
      expect(model.spec.security).toEqual({});
      const request = workloadHybridUtil.prepareWorkloadFlatRequest(model);
      expect(request.spec?.security).toEqual({});
    },
  );
});
