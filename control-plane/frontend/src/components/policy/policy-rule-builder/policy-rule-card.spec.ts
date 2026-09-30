/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import { flushPromises, shallowMount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import PolicyRuleCard from "./policy-rule-card.vue";
import RuleSelectorParams from "./rule-selector-params.vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { NotificationAlertBar } from "@/components/common/notification-alert-bar";
import { EEnforcementType, type IEnforcementDraft, type IFieldInfo, type IRuleDraft } from "@/models/gvk-policy.model";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";
import type { PolicyMapping } from "@/composables/use-policy-mapping.composable";
import { errorMessages } from "@/common/error-message.constant";

interface SelectorCase {
  field: string;
  specSelector: string;
  itemSelector: string;
}

const annotationsCase: SelectorCase = {
  field: "annotations-value",
  specSelector: "podMetadata",
  itemSelector: "annotationsByKey",
};
const environmentVariableCase: SelectorCase = {
  field: "environmentVariables-Value",
  specSelector: "allAppContainers",
  itemSelector: "envVarAsValue",
};

function fieldInfo(itemSelector: string, itemSelectorParams: IFieldInfo["itemSelectorParams"]): IFieldInfo {
  return {
    type: "string",
    builtin: "KeyValue",
    setDefaultGroupingArg: null,
    itemSelectors: [itemSelector],
    allowedEnforcements: [EEnforcementType.Set],
    setParams: [],
    specSelectorParams: [],
    itemSelectorParams,
  };
}

function initialDraft(selectorCase: SelectorCase): IRuleDraft {
  return {
    ...gvkPolicyUtil.createEmptyRuleDraft(),
    field: selectorCase.field,
    specSelector: selectorCase.specSelector,
    itemSelector: null,
  };
}

function mapping(info: IFieldInfo, selectorCase: SelectorCase): PolicyMapping {
  return {
    fields: vi.fn(() => [selectorCase.field]),
    specSelectorsForField: vi.fn(() => [selectorCase.specSelector]),
    itemSelectorsFor: vi.fn(() => [selectorCase.itemSelector]),
    loadFieldInfo: vi.fn().mockResolvedValue(undefined),
    resolveFieldInfo: vi.fn().mockResolvedValue(info),
  } as unknown as PolicyMapping;
}

describe("policy-rule-card.vue — selector metadata", () => {
  let wrapper: VueWrapper | null = null;

  afterEach(() => wrapper?.unmount());

  async function mountCard(
    info: IFieldInfo,
    selectorCase: SelectorCase,
  ): Promise<{ wrapper: VueWrapper; mapping: PolicyMapping }> {
    const policyMapping = mapping(info, selectorCase);
    wrapper = shallowMount(PolicyRuleCard, {
      props: { mapping: policyMapping, initialDraft: initialDraft(selectorCase) },
      global: {
        plugins: [Quasar],
        stubs: { QForm: { template: "<form><slot /></form>" } },
      },
    });
    await flushPromises();
    return { wrapper, mapping: policyMapping };
  }

  function itemSelector(wrapper: VueWrapper): VueWrapper {
    const selector = wrapper
      .findAllComponents(RunaiSelect)
      .find((component) => component.attributes("aid") === "rule-item-selector-select");

    expect(selector).toBeDefined();
    return selector as VueWrapper;
  }

  it("auto-selects and displays the sole item selector with its parameters", async () => {
    const { wrapper, mapping: policyMapping } = await mountCard(
      fieldInfo("annotationsByKey", [{ name: "key", type: "string", required: false }]),
      annotationsCase,
    );
    const selector = itemSelector(wrapper);
    const selectorProps = selector.props() as { modelValue: unknown; options: unknown };

    expect(selectorProps.modelValue).toBe("annotationsByKey");
    expect(selectorProps.options).toEqual([
      { label: "annotationsByKey", value: "annotationsByKey", aid: "rule-item-option-annotationsByKey" },
    ]);
    expect(wrapper.findComponent(RuleSelectorParams).exists()).toBe(true);
    expect(policyMapping.resolveFieldInfo).toHaveBeenCalledWith("annotations-value", "podMetadata", "annotationsByKey");
  });

  it("keeps the sole item-selector identity visible even when it has no parameters", async () => {
    const { wrapper } = await mountCard(fieldInfo("envVarAsValue", []), environmentVariableCase);
    const selectorProps = itemSelector(wrapper).props() as { modelValue: unknown };

    expect(selectorProps.modelValue).toBe("envVarAsValue");
    expect(wrapper.findComponent(RuleSelectorParams).exists()).toBe(false);
  });

  it("loads metadata after field selection and displays parameters under their selectors", async () => {
    const specInfo: IFieldInfo = {
      ...fieldInfo("allPorts", []),
      itemSelectors: ["allPorts", "portByContainerPort"],
      specSelectorParams: [{ name: "containerName", type: "string", required: false }],
    };
    const itemInfo: IFieldInfo = {
      ...specInfo,
      itemSelectorParams: [{ name: "containerPort", type: "integer", required: false }],
    };
    const loadFieldInfo = vi.fn().mockResolvedValue(undefined);
    const resolveFieldInfo = vi
      .fn()
      .mockImplementation((_field: string, _specSelector: string, itemSelector: string | null) =>
        Promise.resolve(itemSelector ? itemInfo : specInfo),
      );
    const policyMapping = {
      fields: vi.fn(() => ["ports-HostPort"]),
      specSelectorsForField: vi.fn((field: string | null) => (field ? ["allAppContainers", "appContainerByName"] : [])),
      itemSelectorsFor: vi.fn((_field: string | null, specSelector: string | null) =>
        specSelector === "appContainerByName" ? ["allPorts", "portByContainerPort"] : [],
      ),
      loadFieldInfo,
      resolveFieldInfo,
    } as unknown as PolicyMapping;
    wrapper = shallowMount(PolicyRuleCard, {
      props: { mapping: policyMapping, initialDraft: gvkPolicyUtil.createEmptyRuleDraft() },
      global: {
        plugins: [Quasar],
        stubs: { QForm: { template: "<form><slot /></form>" } },
      },
    });

    const selectByAid = (aid: string): VueWrapper => {
      const select = wrapper?.findAllComponents(RunaiSelect).find((component) => component.attributes("aid") === aid);
      expect(select).toBeDefined();
      return select as VueWrapper;
    };

    selectByAid("rule-field-select").vm.$emit("update:model-value", "ports-HostPort");
    await flushPromises();
    expect(loadFieldInfo).toHaveBeenCalledWith("ports-HostPort");

    selectByAid("rule-spec-selector-select").vm.$emit("update:model-value", "appContainerByName");
    await flushPromises();
    expect(resolveFieldInfo).toHaveBeenCalledWith("ports-HostPort", "appContainerByName", null);
    let parameterGroups = wrapper.findAllComponents(RuleSelectorParams);
    expect(parameterGroups).toHaveLength(1);
    expect(parameterGroups[0].props("params")).toEqual(specInfo.specSelectorParams);

    selectByAid("rule-item-selector-select").vm.$emit("update:model-value", "portByContainerPort");
    await flushPromises();
    parameterGroups = wrapper.findAllComponents(RuleSelectorParams);
    expect(parameterGroups).toHaveLength(2);
    expect(parameterGroups[0].props("params")).toEqual(specInfo.specSelectorParams);
    expect(parameterGroups[1].props("params")).toEqual(itemInfo.itemSelectorParams);
  });
});

describe("policy-rule-card.vue — enforcement combination validation", () => {
  function setOrDefault(type: EEnforcementType, target?: string): IEnforcementDraft {
    const setParams: Record<string, string> = target === undefined ? {} : { name: target };
    return {
      ...gvkPolicyUtil.createEmptyEnforcement(),
      type,
      value: "configured-value",
      setParams,
    };
  }

  function enforcement(type: EEnforcementType, value = ""): IEnforcementDraft {
    return {
      ...gvkPolicyUtil.createEmptyEnforcement(),
      type,
      value,
      options: type === EEnforcementType.Options ? ["allowed"] : [],
    };
  }

  async function mountValidationCard(
    info: IFieldInfo,
    enforcements: IRuleDraft["enforcements"],
    extraProps: Record<string, unknown> = {},
  ) {
    const policyMapping = {
      fields: vi.fn(() => ["test-field"]),
      specSelectorsForField: vi.fn(() => ["test-selector"]),
      itemSelectorsFor: vi.fn(() => []),
      loadFieldInfo: vi.fn().mockResolvedValue(undefined),
      resolveFieldInfo: vi.fn().mockResolvedValue(info),
    } as unknown as PolicyMapping;
    const wrapper = shallowMount(PolicyRuleCard, {
      props: {
        mapping: policyMapping,
        initialDraft: {
          ...gvkPolicyUtil.createEmptyRuleDraft(),
          field: "test-field",
          specSelector: "test-selector",
          enforcements,
        },
        ...extraProps,
      },
      global: {
        plugins: [Quasar],
        stubs: {
          NotificationAlertBar: {
            props: { title: String, showH1Content: Boolean, orientation: String },
            template: "<div><slot /></div>",
          },
          QForm: {
            template: "<form><slot /></form>",
            methods: { validate: () => Promise.resolve(true) },
          },
        },
      },
    });
    await flushPromises();
    return wrapper;
  }

  it("blocks save when Set and Default target the same flat field", async () => {
    const info: IFieldInfo = {
      ...fieldInfo("", []),
      itemSelectors: [],
      allowedEnforcements: [EEnforcementType.Set, EEnforcementType.Default],
    };
    const wrapper = await mountValidationCard(info, [
      setOrDefault(EEnforcementType.Set),
      setOrDefault(EEnforcementType.Default),
    ]);

    expect(await (wrapper.vm as unknown as { commit: () => Promise<boolean> }).commit()).toBe(false);
    expect(wrapper.find('[aid="rule-card-error"]').text()).toBe(errorMessages.DUPLICATE_SET_DEFAULT_TARGET);
    expect(wrapper.emitted("save")).toBeUndefined();
  });

  it("allows save when repeated Set/Default entries target different API-declared keys", async () => {
    const info: IFieldInfo = {
      ...fieldInfo("", []),
      setDefaultGroupingArg: "name",
      itemSelectors: [],
      allowedEnforcements: [EEnforcementType.Set, EEnforcementType.Default],
      setParams: [
        { name: "name", type: "string", required: true },
        { name: "value", type: "string", required: true },
      ],
    };
    const wrapper = await mountValidationCard(info, [
      setOrDefault(EEnforcementType.Set, "KEY1"),
      setOrDefault(EEnforcementType.Default, "KEY2"),
    ]);

    expect(await (wrapper.vm as unknown as { commit: () => Promise<boolean> }).commit()).toBe(true);
    expect(wrapper.emitted("save")).toHaveLength(1);
  });

  it("blocks save for blocked=true with Options before calling the API", async () => {
    const info: IFieldInfo = {
      ...fieldInfo("", []),
      itemSelectors: [],
      allowedEnforcements: [EEnforcementType.Blocked, EEnforcementType.Options],
    };
    const wrapper = await mountValidationCard(info, [
      enforcement(EEnforcementType.Blocked, "true"),
      enforcement(EEnforcementType.Options),
    ]);

    expect(await (wrapper.vm as unknown as { commit: () => Promise<boolean> }).commit()).toBe(false);
    expect(wrapper.find('[aid="rule-card-error"]').text()).toBe(errorMessages.INCOMPATIBLE_ENFORCEMENTS);
    expect(wrapper.emitted("save")).toBeUndefined();
  });

  it("allows save for blocked=false with Options", async () => {
    const info: IFieldInfo = {
      ...fieldInfo("", []),
      itemSelectors: [],
      allowedEnforcements: [EEnforcementType.Blocked, EEnforcementType.Options],
    };
    const wrapper = await mountValidationCard(info, [
      enforcement(EEnforcementType.Blocked, "false"),
      enforcement(EEnforcementType.Options),
    ]);

    expect(await (wrapper.vm as unknown as { commit: () => Promise<boolean> }).commit()).toBe(true);
    expect(wrapper.emitted("save")).toHaveLength(1);
  });

  it("allows save for blocked=true with required=false", async () => {
    const info: IFieldInfo = {
      ...fieldInfo("", []),
      itemSelectors: [],
      allowedEnforcements: [EEnforcementType.Blocked, EEnforcementType.Required],
    };
    const wrapper = await mountValidationCard(info, [
      enforcement(EEnforcementType.Blocked, "true"),
      enforcement(EEnforcementType.Required, "false"),
    ]);

    expect(await (wrapper.vm as unknown as { commit: () => Promise<boolean> }).commit()).toBe(true);
    expect(wrapper.emitted("save")).toHaveLength(1);
  });

  it("shows the Figma override radio choices and emits the selected value when enabled for the scope", async () => {
    const info: IFieldInfo = {
      ...fieldInfo("", []),
      itemSelectors: [],
      allowedEnforcements: [EEnforcementType.Set],
    };
    const wrapper = await mountValidationCard(
      info,
      [{ ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Set, value: "configured" }],
      { allowOverrideEnabled: true },
    );
    const radioOptions = wrapper.findComponent(RunaiRadioOptions);

    expect(radioOptions.exists()).toBe(true);
    expect(radioOptions.props("title")).toBe("Override by descendant scope");
    expect(radioOptions.props("titleTooltip")).toBe(
      "Sets whether scopes under the selected scope can define policies that take precedence over this rule within their own scope",
    );
    expect(radioOptions.props("modelValue")).toBe(false);
    expect(radioOptions.props("options")).toEqual([
      expect.objectContaining({
        label: "Not allowed",
        value: false,
        toolTip: "Descendant scopes cannot override this rule",
      }),
      expect.objectContaining({
        label: "Allowed",
        value: true,
        toolTip: "Any descendant scope can define a policy that overrides this rule",
      }),
    ]);
    radioOptions.vm.$emit("update:model-value", true);
    await wrapper.vm.$nextTick();

    expect(await (wrapper.vm as unknown as { commit: () => Promise<boolean> }).commit()).toBe(true);
    const save = wrapper.emitted("save")?.[0]?.[0] as { rule: { allowOverride?: boolean } };
    expect(save.rule.allowOverride).toBe(true);
  });

  it.each([
    { scope: "cluster", clusterId: "cluster-id" },
    { scope: "department", departmentId: "department-id" },
    { scope: "project", projectId: "project-id" },
  ])("names only the blocking $scope in the warning", async (sourceOfRule) => {
    const info: IFieldInfo = {
      ...fieldInfo("", []),
      itemSelectors: [],
      allowedEnforcements: [EEnforcementType.Set],
    };
    const wrapper = await mountValidationCard(
      info,
      [{ ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Set, value: "configured" }],
      {
        lockedRules: [
          {
            field: "test-field",
            specSelector: "test-selector",
            allowOverride: false,
            enforcements: [],
            sourceOfRule,
          },
        ],
        tenantName: "runai",
        orgUnits: [
          { id: "cluster-id", type: "cluster", name: "cluster-a" },
          { id: "department-id", type: "department", name: "department-a", parentId: "cluster-id" },
          { id: "project-id", type: "project", name: "project-a", parentId: "department-id" },
        ],
      },
    );

    const warning = wrapper.find('[aid="locked-rule-warning"]');
    expect(warning.exists()).toBe(true);
    expect(wrapper.findComponent(NotificationAlertBar).props()).toMatchObject({
      title: "Rule will be saved but inactive",
      showH1Content: true,
      orientation: "vertical",
    });
    const scope = sourceOfRule.scope;
    expect(warning.text()).toBe(
      `There is a policy rule for the same target in place in ${scope} ${scope}-a that prevents lower scopes from overriding it. The rule you're creating will be saved but remain inactive until the ${scope} rule is removed or changed to allow override.`,
    );
    expect(warning.findAll("strong").map((node) => node.text())).toEqual([`${scope} ${scope}-a`]);
  });
});
