/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, type MockedFunction, vi, beforeEach, test, expect } from "vitest";
import { PolicyCreationForm } from ".";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import { mockHugePolicy, mockSimplePolicy } from "../policy-yaml-section/policy-yaml-section.mock";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { PolicyType, ScopeType } from "@/swagger-models/policy-service-client";
import type { DefineComponent } from "vue";
import type { IScopeModel } from "@/models/global.model";
import { useClusterStore } from "@/stores/cluster.store";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { policyService } from "@/services/control-plane/policy.service/policy.service";

vi.mock("@/services/control-plane/rbac/org-tree.service/org-tree.service.ts");
vi.mock("@/services/control-plane/policy.service/policy.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("policy-creation-form.vue", () => {
  let policyCreationFormWrapper: VueWrapper<InstanceType<typeof PolicyCreationForm>>;

  const mockGetOrgTree = orgTreeService.getOrgTree as MockedFunction<typeof orgTreeService.getOrgTree>;
  mockGetOrgTree.mockResolvedValue(orgTreeServiceMock.orgTree);

  const validatePolicyMock = policyService.validatePolicy as MockedFunction<typeof policyService.validatePolicy>;

  const clusterStore = useClusterStore(pinia);

  function mountComponent(initialPolicy = mockHugePolicy, existingPolicyType: PolicyType | null = null, isEdit = false) {
    policyCreationFormWrapper = mount(PolicyCreationForm, {
      global: {
        plugins: [Quasar],
        directives: {
          permission: vi.fn(),
        },
        mocks: {
          $route: {
            meta: {
              prevRoute: "",
            },
            query: {},
          },
          $router: [],
        },
      },
      props: {
        initialPolicy,
        existingPolicyType,
        submitting: false,
        allowedScopes: { tenant: "runai" } as unknown as Record<string, string>,
        loadingPolicy: false,
        isEdit,
      },
    });
    policyCreationFormWrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  }

  beforeEach(() => {
    validatePolicyMock.mockResolvedValue({});
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);
    mountComponent();
  });

  test("component is loading properly", () => {
    expect(policyCreationFormWrapper.html()).toContain("policy-creation-form");
  });

  test("leave-guard form state includes any-workload rules so the unsaved-changes prompt covers them", async () => {
    const formWrapper = policyCreationFormWrapper.findComponent(RunaiFormWrapper);
    let state = formWrapper.props("formState") as { gvkRules: unknown[] };
    expect(state.gvkRules).toEqual([]);
    await policyCreationFormWrapper.setData({
      gvkRules: [{ field: "image", specSelector: "mainContainer", enforcements: [] }],
    });
    state = formWrapper.props("formState") as { gvkRules: unknown[] };
    expect(state.gvkRules).toHaveLength(1);
  });

  test("policy yaml section will not be rendered if workloadType and Scope are not selected", () => {
    expect(policyCreationFormWrapper.find("[aid=policy-yaml-section]").exists()).toBeFalsy();
  });

  test("render yaml section when both Scope and workloadType are selected", async () => {
    await policyCreationFormWrapper.setData({ selectedWorkloadType: PolicyType.Workspace }); // setting a selected workloadType, no reason to select a scope as its defined on the mock policy while mounting the component
    expect(policyCreationFormWrapper.find("[aid=policy-yaml-section]").exists()).toBeTruthy();
  });

  test("update scope and trigger loadPolicy method when its emitted by scope section", () => {
    const spyLoadPolicy = vi.spyOn(policyCreationFormWrapper.vm, "loadPolicy");
    const newScope: IScopeModel = {
      scope: ScopeType.Department,
      departmentId: "4",
    };

    policyCreationFormWrapper
      .findComponent<DefineComponent>("[aid='runai-scope-section']")
      .vm.$emit("changed-scope", newScope);
    expect(policyCreationFormWrapper.vm.policyModel.meta?.scope).toBe(newScope.scope);
    expect(policyCreationFormWrapper.vm.policyModel.meta?.departmentId).toBe(newScope.departmentId);
    expect(spyLoadPolicy).toHaveBeenCalledOnce();
    spyLoadPolicy.mockRestore();
  });

  test("update workloadType and trigger loadPolicy method when its emitted by workload-type-select", async () => {
    const spyLoadPolicy = vi.spyOn(policyCreationFormWrapper.vm, "loadPolicy");

    policyCreationFormWrapper
      .findComponent<DefineComponent>(".workload-type-select")
      .vm.$emit("update:model-value", "training");
    expect(policyCreationFormWrapper.vm.selectedWorkloadType).toBe("training");

    expect(spyLoadPolicy).toHaveBeenCalledOnce();
    spyLoadPolicy.mockRestore();
  });

  test("update the policy when an updated policy is emitted by policy-yaml-section", async () => {
    await policyCreationFormWrapper.setData({ selectedWorkloadType: PolicyType.Workspace }); // revealing the policyYamlSection by selecting a workloadType (scope already selected)
    await policyCreationFormWrapper.vm.$nextTick();
    policyCreationFormWrapper
      .findComponent<DefineComponent>("[aid=policy-yaml-section]")
      .vm.$emit("update-policy", mockSimplePolicy);
    expect(policyCreationFormWrapper.vm.policyModel.policy).toEqual(mockSimplePolicy);
  });

  test("emitting 'load-policy' only if both workloadType and Scope are selected", async () => {
    const policyModelWithoutScope = {
      ...mockHugePolicy,
      meta: { ...mockHugePolicy.meta, scope: undefined, projectId: null },
    };
    // no scope and no selectedWorkloadType
    await policyCreationFormWrapper.setData({ selectedWorkloadType: null, policyModel: policyModelWithoutScope });
    policyCreationFormWrapper.vm.loadPolicy();

    expect(policyCreationFormWrapper.emitted("load-policy")).toBeFalsy();

    // type is selected but no scope
    await policyCreationFormWrapper.setData({ selectedWorkloadType: PolicyType.Workspace });
    policyCreationFormWrapper.vm.loadPolicy();

    expect(policyCreationFormWrapper.emitted("load-policy")).toBeFalsy();

    // scope is selected by no type
    await policyCreationFormWrapper.setData({ selectedWorkloadType: null, policyModel: mockHugePolicy });
    policyCreationFormWrapper.vm.loadPolicy();

    expect(policyCreationFormWrapper.emitted("load-policy")).toBeFalsy();

    // both selected
    await policyCreationFormWrapper.setData({ selectedWorkloadType: PolicyType.Workspace });
    policyCreationFormWrapper.vm.loadPolicy();

    expect(policyCreationFormWrapper.emitted("load-policy")).toBeTruthy();
  });

  test("run form validation when 'save policy' button is pressed", async () => {
    const spyValidate = vi.spyOn(policyCreationFormWrapper.vm, "validate");
    expect(spyValidate).not.toHaveBeenCalled();

    await policyCreationFormWrapper.find("[aid=submit-btn]").trigger("click");

    expect(spyValidate).toHaveBeenCalledOnce();
    spyValidate.mockRestore();
  });

  test("distributed radio section will reveal if training is the selected type", async () => {
    await policyCreationFormWrapper.setData({ selectedWorkloadType: PolicyType.Training });
    expect(policyCreationFormWrapper.find("[aid=distributed-radio-options]").exists()).toBeTruthy();
  });

  test("distributed radio section will be hidden if the selected type is anything other than training", async () => {
    await policyCreationFormWrapper.setData({ selectedWorkloadType: PolicyType.Workspace });
    expect(policyCreationFormWrapper.find("[aid=distributed-radio-options]").exists()).toBeFalsy();
  });

  test("component will init the selectedWorkloadType as Training if the existing policy type is Distributed", async () => {
    mountComponent(mockHugePolicy, PolicyType.Distributed);
    await policyCreationFormWrapper.vm.$nextTick();
    expect(policyCreationFormWrapper.vm.selectedWorkloadType).toBe(PolicyType.Training);
  });

  test("component's computed 'policyType' will return Distributed if selectedWorkloadType is 'Training' and isDistributed is 'true'", async () => {
    expect(policyCreationFormWrapper.vm.policyType).toBe(null);
    await policyCreationFormWrapper.setData({ selectedWorkloadType: PolicyType.Training, isDistributed: true });
    expect(policyCreationFormWrapper.vm.policyType).toBe(PolicyType.Distributed);
  });

  describe("any-workload cluster-version gate (>= 2.26)", () => {
    const getParentClusterUuid = orgTreeService.getParentClusterUuid as MockedFunction<
      typeof orgTreeService.getParentClusterUuid
    >;

    async function selectProjectWithGvkType(): Promise<void> {
      getParentClusterUuid.mockReturnValue("cluster-uuid");
      await policyCreationFormWrapper.setData({
        gvkWorkloadTypeOptions: [
          {
            key: "gvk:apps/v1/Deployment",
            label: "Deployment",
            gvk: { group: "apps", version: "v1", kind: "Deployment" },
          },
        ],
        policyModel: { meta: { scope: ScopeType.Project, projectId: 123 } },
      });
    }

    function gvkOption() {
      return policyCreationFormWrapper.vm.combinedWorkloadTypeOptions.find((o) => o.value === "gvk:apps/v1/Deployment");
    }

    test("hides the any-workload type when the project's cluster is below 2.26", async () => {
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);
      await selectProjectWithGvkType();
      expect(gvkOption()).toBeUndefined();
    });

    test("enables the any-workload type when the project's cluster supports 2.26", async () => {
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
      await selectProjectWithGvkType();
      expect(gvkOption()?.disable).toBe(false);
      expect(gvkOption()?.disabledTooltip).toBeUndefined();
    });

    function configModeToggleExists(): boolean {
      return policyCreationFormWrapper.find("[aid=config-mode-segmented-button]").exists();
    }

    test("hides the rule-builder/YAML toggle when the project's cluster is below 2.26 (version gate)", async () => {
      getParentClusterUuid.mockReturnValue("cluster-uuid");
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);
      await policyCreationFormWrapper.setData({
        selectedWorkloadType: PolicyType.Workspace,
        policyModel: { meta: { scope: ScopeType.Project, projectId: 123 } },
      });
      expect(configModeToggleExists()).toBe(false);
    });

    test("shows the rule-builder/YAML toggle on native create when the cluster supports 2.26", async () => {
      getParentClusterUuid.mockReturnValue("cluster-uuid");
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
      await policyCreationFormWrapper.setData({
        selectedWorkloadType: PolicyType.Workspace,
        policyModel: { meta: { scope: ScopeType.Project, projectId: 123 } },
      });
      expect(configModeToggleExists()).toBe(true);
    });

    test("non-project scope still shows the toggle — that's the scope gate (disables the type), not the version gate", async () => {
      clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(false);
      await policyCreationFormWrapper.setData({
        selectedWorkloadType: PolicyType.Workspace,
        policyModel: { meta: { scope: ScopeType.Department, departmentId: "4" } },
      });
      expect(configModeToggleExists()).toBe(true);
    });
  });

  describe("workload-type dropdown native/any-workload separator (#1)", () => {
    test("marks only the first any-workload option with a separator; native options have none", async () => {
      await policyCreationFormWrapper.setData({
        gvkWorkloadTypeOptions: [
          {
            key: "gvk:apps/v1/Deployment",
            label: "Deployment",
            gvk: { group: "apps", version: "v1", kind: "Deployment" },
          },
          { key: "gvk:batch/v1/Job", label: "Job", gvk: { group: "batch", version: "v1", kind: "Job" } },
        ],
      });
      const options = policyCreationFormWrapper.vm.combinedWorkloadTypeOptions;
      const gvkOptions = options.filter((o) => String(o.value).startsWith("gvk:"));
      const nativeOptions = options.filter((o) => !String(o.value).startsWith("gvk:"));

      expect(gvkOptions[0].separator).toBe(true);
      expect(gvkOptions[1].separator).toBeFalsy();
      expect(nativeOptions.every((o) => !o.separator)).toBe(true);
    });
  });

  describe("edit mode keeps the workload-type section usable (#7)", () => {
    const gvkOption = {
      key: "gvk:apps/v1/Deployment",
      label: "Deployment",
      gvk: { group: "apps", version: "v1", kind: "Deployment" },
    };

    test("disables only the workload-type select in edit (type can't change)", () => {
      mountComponent(mockHugePolicy, PolicyType.Workspace, true);
      expect(
        // eslint-disable-next-line @typescript-eslint/no-explicit-any -- generic wrapper just to read a prop
        (policyCreationFormWrapper.findComponent(".workload-type-select") as VueWrapper<any>).props("disable"),
      ).toBe(true);
    });

    test("hides the rule-builder/YAML toggle on a native policy edit", () => {
      mountComponent(mockHugePolicy, PolicyType.Workspace, true);
      expect(policyCreationFormWrapper.find("[aid=config-mode-segmented-button]").exists()).toBe(false);
    });

    test("shows the rule-builder/YAML toggle on an any-workload edit, and the toggle is functional", async () => {
      mountComponent(mockHugePolicy, null, true);
      await policyCreationFormWrapper.setData({ selectedGvkOption: gvkOption, selectedWorkloadType: null });
      expect(policyCreationFormWrapper.find("[aid=config-mode-segmented-button]").exists()).toBe(true);

      policyCreationFormWrapper.vm.updateConfigMode("yaml");
      await policyCreationFormWrapper.vm.$nextTick();
      expect(policyCreationFormWrapper.vm.effectiveConfigMode).toBe("yaml");
      expect(policyCreationFormWrapper.vm.showGvkYaml).toBe(true);
      expect(policyCreationFormWrapper.vm.showGvkRuleBuilder).toBe(false);
    });
  });
});
