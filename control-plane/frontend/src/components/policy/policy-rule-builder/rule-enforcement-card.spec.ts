import { shallowMount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, it, expect } from "vitest";
import RuleEnforcementCard from "./rule-enforcement-card.vue";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select";
import { RunaiSelect } from "@/components/common/runai-select";
import EnforcementValueInput from "./enforcement-value-input.vue";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";
import { EEnforcementType, type IEnforcementDraft, type IFieldInfo } from "@/models/gvk-policy.model";
import type { MappingFieldType } from "@/swagger-models/policy-service-client";
import { errorMessages } from "@/common/error-message.constant";

function fieldInfo(overrides: Partial<IFieldInfo> = {}): IFieldInfo {
  return {
    type: "string" as MappingFieldType,
    builtin: null,
    setDefaultGroupingArg: null,
    itemSelectors: [],
    allowedEnforcements: [EEnforcementType.Options],
    setParams: [],
    specSelectorParams: [],
    itemSelectorParams: [],
    ...overrides,
  };
}

function optionsDraft(options: string[]): IEnforcementDraft {
  return { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Options, options };
}

function mountCard(options: string[], showValidation = false) {
  return shallowMount(RuleEnforcementCard, {
    props: { modelValue: optionsDraft(options), fieldInfo: fieldInfo(), index: 0, showValidation },
    global: { plugins: [Quasar] },
  });
}

describe("rule-enforcement-card.vue — Options enforcement", () => {
  // Regression: the "Allowed values" dropdown was empty after adding values because :options was
  // hard-coded to []. The already-added values must populate the dropdown, not just the chips.
  it("feeds the already-added values into the dropdown options", () => {
    const select = mountCard(["test", "option1"]).findComponent(RunaiAddableMultiSelect);
    expect(select.exists()).toBe(true);
    expect((select.props("options") as { value: string }[]).map((o) => o.value)).toEqual(["test", "option1"]);
    // selected mirrors options so the listed values render as already-selected.
    expect(select.props("selectedOptions")).toHaveLength(2);
  });

  it("renders an empty options list when nothing has been added yet", () => {
    const select = mountCard([]).findComponent(RunaiAddableMultiSelect);
    expect(select.props("options")).toHaveLength(0);
  });
});

describe("rule-enforcement-card.vue — inline value validation (shown inside the box)", () => {
  it("hides the value error until a save attempt (showValidation)", () => {
    expect(mountCard([], false).find('[aid="enforcement-value-error-0"]').exists()).toBe(false);
  });

  it("shows 'select a value' inside the card for empty Options on a save attempt", () => {
    const err = mountCard([], true).find('[aid="enforcement-value-error-0"]');
    expect(err.exists()).toBe(true);
    expect(err.text()).toContain(errorMessages.SELECT_A_VALUE);
  });

  it("clears the inline error once a value is added", () => {
    expect(mountCard(["x"], true).find('[aid="enforcement-value-error-0"]').exists()).toBe(false);
  });
});

describe("rule-enforcement-card.vue — API mapping-type validation", () => {
  function rulesForCard(overrides: Partial<IFieldInfo>) {
    const wrapper = shallowMount(RuleEnforcementCard, {
      props: {
        modelValue: { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Set },
        fieldInfo: fieldInfo(overrides),
        index: 0,
      },
      global: { plugins: [Quasar] },
    });
    return wrapper.findComponent(EnforcementValueInput).props("rules") as Array<(value: string) => boolean | string>;
  }

  it("applies ipaddress validation to a direct Set value from fieldInfo.type", () => {
    const rules = rulesForCard({ type: "ipaddress", allowedEnforcements: [EEnforcementType.Set] });
    expect(rules.some((rule) => rule("asd") === errorMessages.ENTER_VALID_IP_ADDRESS)).toBe(true);
    expect(rules.every((rule) => rule("192.168.1.1") === true)).toBe(true);
  });

  it("applies ipaddress validation to a typed formatter parameter from argsInfo", () => {
    const rules = rulesForCard({
      allowedEnforcements: [EEnforcementType.Set],
      setParams: [{ name: "hostAliasIp", type: "ipaddress", required: true }],
    });
    expect(rules.some((rule) => rule("asd") === errorMessages.ENTER_VALID_IP_ADDRESS)).toBe(true);
    expect(rules.every((rule) => rule("2001:db8::1") === true)).toBe(true);
  });
});

describe("rule-enforcement-card.vue — enforcement compatibility", () => {
  const allowedEnforcements = [
    EEnforcementType.Set,
    EEnforcementType.Default,
    EEnforcementType.Required,
    EEnforcementType.Blocked,
    EEnforcementType.Options,
  ];

  function enforcement(type: EEnforcementType, value = ""): IEnforcementDraft {
    return { ...gvkPolicyUtil.createEmptyEnforcement(), type, value };
  }

  function typeOptionsFor(usedEnforcements: IEnforcementDraft[]) {
    const wrapper = shallowMount(RuleEnforcementCard, {
      props: {
        modelValue: gvkPolicyUtil.createEmptyEnforcement(),
        fieldInfo: fieldInfo({ allowedEnforcements }),
        index: 0,
        usedEnforcements,
      },
      global: { plugins: [Quasar] },
    });
    return wrapper.findComponent(RunaiSelect).props("options") as {
      value: EEnforcementType;
      disable: boolean;
      disabledTooltip?: string;
    }[];
  }

  function optionByType(options: ReturnType<typeof typeOptionsFor>, type: EEnforcementType) {
    return options.find((option) => option.value === type);
  }

  it("disables enforcement types that cannot coexist with blocked=true", () => {
    const options = typeOptionsFor([enforcement(EEnforcementType.Blocked, "true")]);

    expect(optionByType(options, EEnforcementType.Set)).toMatchObject({
      disable: true,
      disabledTooltip: errorMessages.INCOMPATIBLE_ENFORCEMENTS,
    });
    expect(optionByType(options, EEnforcementType.Default)).toMatchObject({
      disable: true,
      disabledTooltip: errorMessages.INCOMPATIBLE_ENFORCEMENTS,
    });
    expect(optionByType(options, EEnforcementType.Options)).toMatchObject({
      disable: true,
      disabledTooltip: errorMessages.INCOMPATIBLE_ENFORCEMENTS,
    });
    expect(optionByType(options, EEnforcementType.Required)?.disable).toBe(false);
  });

  it("does not disable peer enforcement types when blocked=false", () => {
    const options = typeOptionsFor([enforcement(EEnforcementType.Blocked, "false")]);
    expect(optionByType(options, EEnforcementType.Set)?.disable).toBe(false);
    expect(optionByType(options, EEnforcementType.Default)?.disable).toBe(false);
    expect(optionByType(options, EEnforcementType.Required)?.disable).toBe(false);
    expect(optionByType(options, EEnforcementType.Options)?.disable).toBe(false);
  });

  it("keeps Blocked selectable with another enforcement because Not blocked is valid", () => {
    const blocked = optionByType(typeOptionsFor([enforcement(EEnforcementType.Options)]), EEnforcementType.Blocked);
    expect(blocked?.disable).toBe(false);
  });

  it("disables both Set and Default after either is selected for a flat field", () => {
    const options = typeOptionsFor([enforcement(EEnforcementType.Set), enforcement(EEnforcementType.Default)]);
    expect(optionByType(options, EEnforcementType.Set)).toMatchObject({
      disable: true,
      disabledTooltip: errorMessages.DUPLICATE_SET_DEFAULT_TARGET,
    });
    expect(optionByType(options, EEnforcementType.Default)).toMatchObject({
      disable: true,
      disabledTooltip: errorMessages.DUPLICATE_SET_DEFAULT_TARGET,
    });
  });

  it("keeps Set and Default enabled for a field grouped by an API-declared array key", () => {
    const wrapper = shallowMount(RuleEnforcementCard, {
      props: {
        modelValue: gvkPolicyUtil.createEmptyEnforcement(),
        fieldInfo: fieldInfo({ allowedEnforcements, setDefaultGroupingArg: "name" }),
        index: 0,
        usedEnforcements: [enforcement(EEnforcementType.Set), enforcement(EEnforcementType.Default)],
      },
      global: { plugins: [Quasar] },
    });
    const options = wrapper.findComponent(RunaiSelect).props("options") as ReturnType<typeof typeOptionsFor>;
    expect(optionByType(options, EEnforcementType.Set)?.disable).toBe(false);
    expect(optionByType(options, EEnforcementType.Default)?.disable).toBe(false);
  });
});
