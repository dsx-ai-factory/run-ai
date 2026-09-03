/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import PvcCreateForm from "./pvc-create-form.vue";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Scope } from "@/swagger-models/assets-service-client";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import type { IUIPVCCreationRequest } from "@/models/data-source.model";
import { useClusterStore } from "@/stores/cluster.store";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

vi.mock("@/services/control-plane/asset.service/asset.service");
vi.mock("@/services/control-plane/rbac/org-tree.service/org-tree.service.ts");
vi.mock("@/services/control-plane/data-source.service/data-source.service.ts");

describe("pvc-create-form.vue", () => {
  let pvcCreateFormCmp: VueWrapper<InstanceType<typeof PvcCreateForm>>;

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;
  mockIsExist.mockResolvedValue(false);

  const mockGetOrgTree = orgTreeService.getOrgTree as MockedFunction<typeof orgTreeService.getOrgTree>;
  mockGetOrgTree.mockReturnValue(orgTreeServiceMock.orgTree);

  const mockGetNodePathById = orgTreeService.getNodePathById as MockedFunction<typeof orgTreeService.getNodePathById>;
  mockGetNodePathById.mockReturnValue("runai");

  const mockListExistingPvc = dataSourceService.listExistingPVCs as MockedFunction<
    typeof dataSourceService.listExistingPVCs
  >;
  mockListExistingPvc.mockResolvedValue([]);

  const mockListStorageClassV2 = dataSourceService.listStorageClassV2 as MockedFunction<
    typeof dataSourceService.listStorageClassV2
  >;
  mockListStorageClassV2.mockResolvedValue([
    { storageClassName: "test-class1", clusterId: "cluster-1", spec: { is_default: true } },
  ]);

  const mockGetStorageClassConfigurationByName = dataSourceService.getStorageClassConfigurationByName as MockedFunction<
    typeof dataSourceService.getStorageClassConfigurationByName
  >;
  mockGetStorageClassConfigurationByName.mockResolvedValue([
    {
      id: "test-class1",
      name: "test-class1",
      scopeType: "cluster",
      scopeId: "cluster-1",
      tenantId: 1001,
    },
  ]);

  let clusterStore = useClusterStore(pinia);

  const validPvcModel: IUIPVCCreationRequest = {
    meta: {
      name: "test-pvc",
      projectId: 25,
      clusterId: null,
      scope: Scope.Project,
    },
    spec: {
      existingPvc: false,
      claimName: "test-claim",
      claimInfo: {
        storageClass: "test-class1",
        size: "1G",
        volumeMode: "Filesystem",
        accessModes: {
          readWriteOnce: true,
          readOnlyMany: false,
          readWriteMany: false,
        },
      },
      path: "/test",
    },
  };

  beforeEach(() => {
    clusterStore = useClusterStore(pinia);

    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);

    pvcCreateFormCmp = mount(PvcCreateForm, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        submitting: false,
        dataSourceModel: validPvcModel as IUIPVCCreationRequest,
        isScopeSelectable: true as boolean,
        initClaimInfoFromScratch: true as boolean,
        allowedScopes: {},
      },
    });

    setQuasarMock(pvcCreateFormCmp);
  });
  afterEach(() => {
    vi.clearAllMocks();
  });

  test("the form validation should fail if name is invalid", async () => {
    validPvcModel.meta.name = "";
    await pvcCreateFormCmp.setData({ pvcModel: validPvcModel });
    expect(await pvcCreateFormCmp.vm.validate()).toBe(false);

    validPvcModel.meta.name = "111";
    await pvcCreateFormCmp.setData({ pvcModel: validPvcModel });
    expect(await pvcCreateFormCmp.vm.validate()).toBe(false);

    validPvcModel.meta.name = "test-pvc";
    await pvcCreateFormCmp.setData({ pvcModel: validPvcModel });
    expect(await pvcCreateFormCmp.vm.validate()).toBe(true);
  });

  test("the form validation should fail if data mount section is invalid", async () => {
    if (!validPvcModel.spec.claimInfo) return;
    validPvcModel.spec.claimInfo.storageClass = "";
    await pvcCreateFormCmp.setData({ pvcModel: validPvcModel });
    expect(await pvcCreateFormCmp.vm.validate()).toBe(false);

    validPvcModel.spec.claimInfo.storageClass = "test-class1";
    await pvcCreateFormCmp.setData({ pvcModel: validPvcModel });
    expect(await pvcCreateFormCmp.vm.validate()).toBe(true);

    validPvcModel.spec.claimInfo.size = "0M";
    await pvcCreateFormCmp.setData({ pvcModel: validPvcModel });
    expect(await pvcCreateFormCmp.vm.validate()).toBe(false);

    validPvcModel.spec.claimInfo.size = "1G";
    await pvcCreateFormCmp.setData({ pvcModel: validPvcModel });
    expect(await pvcCreateFormCmp.vm.validate()).toBe(true);
  });

  test("selecting a shared volume clears an existing PVC selection", async () => {
    await pvcCreateFormCmp.setData({
      pvcModel: {
        ...validPvcModel,
        spec: {
          ...validPvcModel.spec,
          claimName: "existing-claim",
          claimInfo: undefined,
          dataSharing: false,
          existingPvc: true,
        },
      },
    });

    pvcCreateFormCmp.vm.updateDataSharing(true);

    expect(pvcCreateFormCmp.vm.pvcModel.spec.dataSharing).toBe(true);
    expect(pvcCreateFormCmp.vm.pvcModel.spec.existingPvc).toBe(false);
    expect(pvcCreateFormCmp.vm.pvcModel.spec.claimName).toBe("");
  });
});
