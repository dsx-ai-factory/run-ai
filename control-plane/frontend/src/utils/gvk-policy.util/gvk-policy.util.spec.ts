/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { gvkPolicyUtil } from "./gvk-policy.util";
import {
  EEnforcementType,
  type IEnforcementDraft,
  type IFieldInfo,
  type IMappingParam,
} from "@/models/gvk-policy.model";
import { ScopeType } from "@/swagger-models/policy-service-client";
import type { MappingFieldType, PolicyRule, SourceOfRule } from "@/swagger-models/policy-service-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import { errorMessages } from "@/common/error-message.constant";

/**
 * Executable documentation of the any-workload policy-builder data-layer findings
 * (QA hardening pass). Each `it` is the repro for a catalogued edge case; the assertion
 * captures the CURRENT behaviour so a future fix flips it deliberately. No product code is
 * changed by this file.
 */

function fieldInfo(type: MappingFieldType, setParams: IMappingParam[] = [], spec: IMappingParam[] = []): IFieldInfo {
  return {
    type,
    builtin: null,
    setDefaultGroupingArg: null,
    itemSelectors: [],
    // allowedEnforcements now comes from the backend (`possibleEnforcements`); these coercion/serialization
    // tests don't depend on it, so use a permissive static list.
    allowedEnforcements: Object.values(EEnforcementType),
    setParams,
    specSelectorParams: spec,
    itemSelectorParams: [],
  };
}

describe("gvk-policy.util — enforcement compatibility", () => {
  function enforcement(type: EEnforcementType, value = ""): IEnforcementDraft {
    return { ...gvkPolicyUtil.createEmptyEnforcement(), type, value };
  }

  it.each([
    EEnforcementType.Set,
    EEnforcementType.Default,
    EEnforcementType.Required,
    EEnforcementType.Match,
    EEnforcementType.Options,
    EEnforcementType.Length,
    EEnforcementType.Range,
    EEnforcementType.QuantityRange,
    EEnforcementType.Elements,
    EEnforcementType.Items,
  ])("rejects blocked=true with %s", (peerType) => {
    expect(
      gvkPolicyUtil.enforcementCombinationError([
        enforcement(EEnforcementType.Blocked, "true"),
        enforcement(peerType, "true"),
      ]),
    ).toBe(errorMessages.INCOMPATIBLE_ENFORCEMENTS);
  });

  it("allows blocked=false with another enforcement", () => {
    expect(
      gvkPolicyUtil.enforcementCombinationError([
        enforcement(EEnforcementType.Blocked, "false"),
        enforcement(EEnforcementType.Options),
      ]),
    ).toBeNull();
  });

  it("allows blocked=true with required=false", () => {
    expect(
      gvkPolicyUtil.enforcementCombinationError([
        enforcement(EEnforcementType.Blocked, "true"),
        enforcement(EEnforcementType.Required, "false"),
      ]),
    ).toBeNull();
  });

  it("checks the combination in either order", () => {
    expect(
      gvkPolicyUtil.enforcementCombinationError([
        enforcement(EEnforcementType.Options),
        enforcement(EEnforcementType.Blocked, "true"),
      ]),
    ).toBe(errorMessages.INCOMPATIBLE_ENFORCEMENTS);
  });

  it("leaves Set/Default target uniqueness to the field-aware validator", () => {
    expect(
      gvkPolicyUtil.enforcementCombinationError([
        enforcement(EEnforcementType.Set),
        enforcement(EEnforcementType.Set),
        enforcement(EEnforcementType.Default),
        enforcement(EEnforcementType.Default),
      ]),
    ).toBeNull();
  });
});

describe("gvk-policy.util — Set/Default target uniqueness", () => {
  function setOrDefault(type: EEnforcementType, target?: string): IEnforcementDraft {
    return {
      ...gvkPolicyUtil.createEmptyEnforcement(),
      type,
      value: target ?? "value",
      setParams: target === undefined ? {} : { name: target },
    };
  }

  const flatField = fieldInfo("string");
  const keyedField: IFieldInfo = {
    ...fieldInfo("string", [{ name: "name", type: "string", required: true }]),
    setDefaultGroupingArg: "name",
  };

  it.each([
    [EEnforcementType.Set, EEnforcementType.Set],
    [EEnforcementType.Set, EEnforcementType.Default],
    [EEnforcementType.Default, EEnforcementType.Default],
  ])("rejects %s plus %s on a flat field", (first, second) => {
    expect(gvkPolicyUtil.setDefaultCombinationError([setOrDefault(first), setOrDefault(second)], flatField)).toBe(
      errorMessages.DUPLICATE_SET_DEFAULT_TARGET,
    );
  });

  it("allows Set/Default entries that target different keys", () => {
    expect(
      gvkPolicyUtil.setDefaultCombinationError(
        [setOrDefault(EEnforcementType.Set, "KEY1"), setOrDefault(EEnforcementType.Default, "KEY2")],
        keyedField,
      ),
    ).toBeNull();
  });

  it("rejects Set/Default entries that target the same key", () => {
    expect(
      gvkPolicyUtil.setDefaultCombinationError(
        [setOrDefault(EEnforcementType.Set, "KEY1"), setOrDefault(EEnforcementType.Default, "KEY1")],
        keyedField,
      ),
    ).toBe(errorMessages.DUPLICATE_SET_DEFAULT_TARGET);
  });

  it("rejects a keyless fan-out combined with a concrete key", () => {
    expect(
      gvkPolicyUtil.setDefaultCombinationError(
        [setOrDefault(EEnforcementType.Set), setOrDefault(EEnforcementType.Default, "KEY1")],
        keyedField,
      ),
    ).toBe(errorMessages.DUPLICATE_SET_DEFAULT_TARGET);
  });

  it("normalizes typed grouping values before comparing them", () => {
    const integerKeyedField: IFieldInfo = {
      ...keyedField,
      setParams: [{ name: "name", type: "integer", required: true }],
    };
    expect(
      gvkPolicyUtil.setDefaultCombinationError(
        [setOrDefault(EEnforcementType.Set, "01"), setOrDefault(EEnforcementType.Default, "1")],
        integerKeyedField,
      ),
    ).toBe(errorMessages.DUPLICATE_SET_DEFAULT_TARGET);
  });
});

describe("gvk-policy.util — value coercion (convertValueByType)", () => {
  it("returns the raw string for empty / blank input (not null, not omitted)", () => {
    expect(gvkPolicyUtil.convertValueByType("", "string")).toBe("");
  });

  it("coerces boolean only from the exact literal 'true'", () => {
    expect(gvkPolicyUtil.convertValueByType("true", "boolean")).toBe(true);
    expect(gvkPolicyUtil.convertValueByType("false", "boolean")).toBe(false);
    expect(gvkPolicyUtil.convertValueByType("anything-else", "boolean")).toBe(false);
  });

  it("coerces scalar integer / number, falling back to the raw string on NaN", () => {
    expect(gvkPolicyUtil.convertValueByType("5", "integer")).toBe(5);
    expect(gvkPolicyUtil.convertValueByType("1.5", "number")).toBe(1.5);
    // NaN fallback keeps the raw string — only reachable via free-text inputs (arrays/params).
    expect(gvkPolicyUtil.convertValueByType("abc", "integer")).toBe("abc");
  });

  // FINDING #6 — array fields take free-text CSV and emit a MIXED-type array.
  it("FINDING #6: integers/numbers CSV with a non-numeric token yields a mixed-type array", () => {
    expect(gvkPolicyUtil.convertValueByType("1, abc, 3", "integers")).toEqual([1, "abc", 3]);
    expect(gvkPolicyUtil.convertValueByType("1, abc, 3", "numbers")).toEqual([1, "abc", 3]);
  });

  it("splits string arrays and drops empty CSV elements", () => {
    expect(gvkPolicyUtil.convertValueByType("a, , b", "strings")).toEqual(["a", "b"]);
  });

  // FINDING #7 — malformed JSON for an object field is silently passed through as a string.
  it("FINDING #7: malformed JSON for an object field is returned as the raw string", () => {
    expect(gvkPolicyUtil.convertValueByType('{"k":1}', "object")).toEqual({ k: 1 });
    expect(gvkPolicyUtil.convertValueByType("{not json", "object")).toBe("{not json");
  });

  // FINDING #8 — quantity scalars are passed through untouched (no k8s-quantity validation).
  it("FINDING #8: quantity scalars are passed through verbatim (no validation)", () => {
    expect(gvkPolicyUtil.convertValueByType("1Gi", "quantity")).toBe("1Gi");
    expect(gvkPolicyUtil.convertValueByType("definitely-not-a-quantity", "quantity")).toBe("definitely-not-a-quantity");
  });
});

describe("gvk-policy.util — enforcement serialization (enforcementDraftToApi)", () => {
  const STRING = fieldInfo("string");

  it("builds a single-value set/default", () => {
    const set = gvkPolicyUtil.enforcementDraftToApi(
      { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Set, value: "nginx:1.27" },
      STRING,
    );
    expect(set).toEqual({ set: { value: "nginx:1.27" } });
  });

  it("emits required/blocked booleans both ways (radio-controlled; false is explicit, not omitted)", () => {
    const reqDefault = gvkPolicyUtil.enforcementDraftToApi(
      { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Required, value: "" },
      STRING,
    );
    expect(reqDefault).toEqual({ required: true }); // untouched radio defaults to the positive flag
    const blockedFalse = gvkPolicyUtil.enforcementDraftToApi(
      { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Blocked, value: "false" },
      STRING,
    );
    expect(blockedFalse).toEqual({ blocked: false });
  });

  // FINDING #3 — quantityRange is built verbatim; min>max and invalid quantities are NOT caught here.
  it("FINDING #3: quantityRange passes min/max/step strings through unchecked (min>max, invalid)", () => {
    const out = gvkPolicyUtil.enforcementDraftToApi(
      { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.QuantityRange, min: "4Gi", max: "1Gi" },
      fieldInfo("quantity"),
    );
    expect(out).toEqual({ quantityRange: { min: "4Gi", max: "1Gi" } });
  });

  // FINDING #4 — range step is built verbatim; step=0 is NOT caught here.
  it("FINDING #4: range builds step verbatim, including step=0", () => {
    const out = gvkPolicyUtil.enforcementDraftToApi(
      { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Range, min: "1", max: "10", step: "0" },
      fieldInfo("number"),
    );
    expect(out).toEqual({ range: { min: 1, max: 10, step: 0 } });
  });

  it("builds a multi-param set, and an empty object when every param is blank", () => {
    const params: IMappingParam[] = [
      { name: "name", type: "string", required: true },
      { name: "key", type: "string" },
    ];
    const fi = fieldInfo("object", params);
    expect(
      gvkPolicyUtil.enforcementDraftToApi(
        { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Set, setParams: { name: "n", key: "k" } },
        fi,
      ),
    ).toEqual({ set: { name: "n", key: "k" } });
    // all-blank params → `{ set: {} }` (the util does not guard this; the per-input rule does)
    expect(
      gvkPolicyUtil.enforcementDraftToApi(
        { ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Set, setParams: {} },
        fi,
      ),
    ).toEqual({ set: {} });
  });

  it("returns null for an unselected enforcement type", () => {
    expect(
      gvkPolicyUtil.enforcementDraftToApi({ ...gvkPolicyUtil.createEmptyEnforcement(), type: null }, STRING),
    ).toBeNull();
  });
});

describe("gvk-policy.util — rule build & round-trip (ruleDraftToApi / apiRuleToDraft)", () => {
  it("omits itemSelector and selectorParams when empty, includes them when set", () => {
    const fi = fieldInfo("string", [], [{ name: "containerName", type: "string" }]);
    const bare = gvkPolicyUtil.ruleDraftToApi(
      {
        ...gvkPolicyUtil.createEmptyRuleDraft(),
        field: "image",
        specSelector: "mainContainer",
        enforcements: [{ ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Set, value: "nginx" }],
      },
      fi,
    );
    expect(bare.itemSelector).toBeUndefined();
    expect(bare.selectorParams).toBeUndefined();

    const withParams = gvkPolicyUtil.ruleDraftToApi(
      {
        ...gvkPolicyUtil.createEmptyRuleDraft(),
        field: "envVars",
        specSelector: "mainContainer",
        itemSelector: "byName",
        selectorParams: { containerName: "main" },
        enforcements: [{ ...gvkPolicyUtil.createEmptyEnforcement(), type: EEnforcementType.Required, value: "true" }],
      },
      fi,
    );
    expect(withParams.itemSelector).toBe("byName");
    expect(withParams.selectorParams).toEqual({ containerName: "main" });
  });

  it("round-trips a saved rule API → draft → API for typed values", () => {
    const rule: PolicyRule = {
      field: "image",
      specSelector: "mainContainer",
      enforcements: [{ set: { value: "nginx:1.27" } }, { required: false }],
    };
    const draft = gvkPolicyUtil.apiRuleToDraft(rule);
    expect(draft.field).toBe("image");
    expect(draft.enforcements[0].type).toBe(EEnforcementType.Set);
    expect(draft.enforcements[0].value).toBe("nginx:1.27");
    expect(draft.enforcements[1].type).toBe(EEnforcementType.Required);
    expect(draft.enforcements[1].value).toBe("false"); // required:false survives the round-trip
  });

  it("round-trips a typed array set value without losing element types", () => {
    const rule: PolicyRule = {
      field: "portNumbers",
      specSelector: "mainContainer",
      enforcements: [{ set: { value: [1, 2, 3] } }],
    };
    const draft = gvkPolicyUtil.apiRuleToDraft(rule);
    expect(draft.enforcements[0].value).toBe("1, 2, 3"); // arrays stringify to CSV in the draft
    const back = gvkPolicyUtil.ruleDraftToApi(draft, fieldInfo("integers"));
    expect(back.enforcements[0]).toEqual({ set: { value: [1, 2, 3] } });
  });
});

describe("gvk-policy.util — API mapping-type validation", () => {
  const isString = (result: boolean | string): boolean => typeof result === "string";

  it.each([
    "192.168.1.1",
    "0.0.0.0",
    "255.255.255.255",
    "2001:db8::1",
    "::1",
    "2001:0db8:0000:0000:0000:ff00:0042:8329",
    "::ffff:192.0.2.1",
  ])("accepts valid IP address %s", (value) => {
    expect(gvkPolicyUtil.isValidIpAddress(value)).toBe(true);
  });

  it.each([
    "asd",
    "256.1.1.1",
    "192.168.1",
    "192.168.001.1",
    "2001:::1",
    "1:2:3:4:5:6:7:8:9",
    "10.0.0.1/24",
    "10.0.0.1:8080",
    "[2001:db8::1]",
    "2001:db8::1%eth0",
    "192.0.2.1::",
  ])("rejects invalid IP address %s", (value) => {
    expect(gvkPolicyUtil.isValidIpAddress(value)).toBe(errorMessages.ENTER_VALID_IP_ADDRESS);
  });

  it("lets the required rule own the empty IP-address error", () => {
    expect(gvkPolicyUtil.isValidIpAddress("")).toBe(true);
    expect(gvkPolicyUtil.isValidIpAddress("   ")).toBe(true);
  });

  it("selects validators from the mapping type returned by field-info", () => {
    expect(gvkPolicyUtil.mappingTypeValidationRules("ipaddress")).toEqual([gvkPolicyUtil.isValidIpAddress]);
    expect(gvkPolicyUtil.mappingTypeValidationRules("quantity")).toEqual([gvkPolicyUtil.isValidK8sQuantity]);
    expect(gvkPolicyUtil.mappingTypeValidationRules("string")).toEqual([]);
  });

  it("accepts valid k8s quantities and an empty value (bounds are optional)", () => {
    for (const ok of ["1Gi", "500Mi", "2", "500m", "1.5G", ""]) {
      expect(gvkPolicyUtil.isValidK8sQuantity(ok)).toBe(true);
    }
  });

  it("rejects malformed quantities (lowercase suffix, GB, spaces, unknown unit)", () => {
    for (const bad of ["1gi", "1g", "100GB", "1 Gi", "1ki", "abc"]) {
      expect(isString(gvkPolicyUtil.isValidK8sQuantity(bad))).toBe(true); // returns an error message
    }
  });

  it("QuantityRange: min/max/step each validate quantity format", () => {
    const rules = gvkPolicyUtil.enforcementValidationRules({
      enforcementType: EEnforcementType.QuantityRange,
      fieldType: "quantity",
      min: "",
    });
    for (const slot of [rules.min, rules.max, rules.step]) {
      expect(slot[0]("1gi")).not.toBe(true);
      expect(slot[0]("1Gi")).toBe(true);
    }
  });

  it("QuantityRange: max must be >= min (compared as quantities), empty passes", () => {
    const rules = gvkPolicyUtil.enforcementValidationRules({
      enforcementType: EEnforcementType.QuantityRange,
      fieldType: "quantity",
      min: "4Gi",
    });
    const maxAboveMin = rules.max[rules.max.length - 1];
    expect(isString(maxAboveMin("1Gi"))).toBe(true); // 1Gi < 4Gi → error
    expect(maxAboveMin("8Gi")).toBe(true);
    expect(maxAboveMin("4Gi")).toBe(true); // equal is allowed
    expect(maxAboveMin("")).toBe(true); // optional
  });

  it("QuantityRange: step must be > 0 and <= max-min (mirrors backend ValidateQuantityRangeBounds)", () => {
    const rules = gvkPolicyUtil.enforcementValidationRules({
      enforcementType: EEnforcementType.QuantityRange,
      fieldType: "quantity",
      min: "1Gi",
      max: "3Gi",
    });
    const stepRule = rules.step[rules.step.length - 1];
    expect(stepRule("1Gi")).toBe(true); // within span (max-min = 2Gi)
    expect(stepRule("2Gi")).toBe(true); // equal to span allowed
    expect(isString(stepRule("4Gi"))).toBe(true); // exceeds span → error
    expect(stepRule("")).toBe(true); // optional

    // no max set → only the step>0 guard applies, span check skipped
    const noMax = gvkPolicyUtil.enforcementValidationRules({
      enforcementType: EEnforcementType.QuantityRange,
      fieldType: "quantity",
      min: "1Gi",
      max: "",
    });
    expect(noMax.step[noMax.step.length - 1]("9Gi")).toBe(true);
  });

  it("Set/Default on a quantity field validates the value; non-quantity fields add no rule", () => {
    const quantitySet = gvkPolicyUtil.enforcementValidationRules({
      enforcementType: EEnforcementType.Set,
      fieldType: "quantity",
    });
    expect(quantitySet.value[0]("1gi")).not.toBe(true);
    expect(quantitySet.value[0]("1Gi")).toBe(true);

    const stringSet = gvkPolicyUtil.enforcementValidationRules({
      enforcementType: EEnforcementType.Default,
      fieldType: "string",
    });
    expect(stringSet.value).toEqual([]);
  });

  it.each([EEnforcementType.Set, EEnforcementType.Default])(
    "%s on an ipaddress field uses the API-type validator",
    (enforcementType) => {
      const rules = gvkPolicyUtil.enforcementValidationRules({ enforcementType, fieldType: "ipaddress" });
      expect(rules.value).toEqual([gvkPolicyUtil.isValidIpAddress]);
      expect(rules.value[0]("asd")).toBe(errorMessages.ENTER_VALID_IP_ADDRESS);
      expect(rules.value[0]("192.168.1.1")).toBe(true);
    },
  );

  it("numeric Range adds no quantity rules (numeric input handles its own validation)", () => {
    const rules = gvkPolicyUtil.enforcementValidationRules({
      enforcementType: EEnforcementType.Range,
      fieldType: "number",
      min: "1",
    });
    expect(rules).toEqual({ value: [], min: [], max: [], step: [] });
  });
});

describe("gvk-policy.util — any-workload scope / save-error helpers", () => {
  it("isGvkTypeDisabledForScope: disabled for non-project scopes, enabled for project / unset", () => {
    expect(gvkPolicyUtil.isGvkTypeDisabledForScope(ScopeType.Project)).toBe(false);
    expect(gvkPolicyUtil.isGvkTypeDisabledForScope(undefined)).toBe(false);
    expect(gvkPolicyUtil.isGvkTypeDisabledForScope(ScopeType.Cluster)).toBe(true);
    expect(gvkPolicyUtil.isGvkTypeDisabledForScope(ScopeType.Department)).toBe(true);
    expect(gvkPolicyUtil.isGvkTypeDisabledForScope(ScopeType.Tenant)).toBe(true);
  });

  it("shouldResetGvkOnScopeChange: reset only when a GVK type is selected and the scope leaves project", () => {
    expect(gvkPolicyUtil.shouldResetGvkOnScopeChange({ isAnyWorkload: true, newScope: ScopeType.Cluster })).toBe(true);
    expect(gvkPolicyUtil.shouldResetGvkOnScopeChange({ isAnyWorkload: true, newScope: ScopeType.Project })).toBe(false);
    expect(gvkPolicyUtil.shouldResetGvkOnScopeChange({ isAnyWorkload: false, newScope: ScopeType.Cluster })).toBe(false);
  });

  it("isNetworkOrUnknownError: network (≤0) / 5xx / non-HttpErrorResponse → true; 4xx → false", () => {
    class FakeHttpError extends HttpErrorResponse {}
    expect(gvkPolicyUtil.isNetworkOrUnknownError(new FakeHttpError("network", -1, "no response"))).toBe(true);
    expect(gvkPolicyUtil.isNetworkOrUnknownError(new FakeHttpError("network", 0, "no response"))).toBe(true);
    expect(gvkPolicyUtil.isNetworkOrUnknownError(new FakeHttpError("server", 500, "boom"))).toBe(true);
    expect(gvkPolicyUtil.isNetworkOrUnknownError(new FakeHttpError("validation", 400, "bad field"))).toBe(false);
    expect(gvkPolicyUtil.isNetworkOrUnknownError(new Error("plain"))).toBe(true);
    expect(gvkPolicyUtil.isNetworkOrUnknownError("oops")).toBe(true);
  });
});

describe("enforcementValueError (cross-field value gaps shown inside the enforcement card)", () => {
  const base = () => gvkPolicyUtil.createEmptyEnforcement();
  it("flags empty Options, passes when a value exists", () => {
    expect(gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.Options, options: [] })).toBe(
      errorMessages.SELECT_A_VALUE,
    );
    expect(
      gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.Options, options: ["a"] }),
    ).toBeNull();
  });
  it("flags empty min/max for length/elements/items", () => {
    expect(gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.Length })).toBe(
      errorMessages.ENFORCEMENT_NEEDS_MIN_OR_MAX,
    );
    expect(gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.Items, min: "1" })).toBeNull();
  });
  it("flags empty min/max/step for range and quantityRange", () => {
    expect(gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.Range })).toBe(
      errorMessages.ENFORCEMENT_NEEDS_MIN_MAX_OR_STEP,
    );
    expect(
      gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.QuantityRange, max: "2Gi" }),
    ).toBeNull();
  });
  it("returns null for value-typed enforcements (Set / Required / Match)", () => {
    expect(gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.Set })).toBeNull();
    expect(gvkPolicyUtil.enforcementValueError({ ...base(), type: EEnforcementType.Required })).toBeNull();
  });
});

describe("ruleIdentityKey (builder duplicate detection)", () => {
  const rule = (over: Partial<PolicyRule> = {}): PolicyRule =>
    ({ field: "image", specSelector: "allAppContainers", enforcements: [], ...over }) as PolicyRule;

  it("is equal for the same field + selectors + params, regardless of param key order", () => {
    const a = rule({ specSelector: "byName", selectorParams: { containerName: "c", replicaType: "Chief" } });
    const b = rule({ specSelector: "byName", selectorParams: { replicaType: "Chief", containerName: "c" } });
    expect(gvkPolicyUtil.ruleIdentityKey(a)).toBe(gvkPolicyUtil.ruleIdentityKey(b));
  });

  it("differs when a selector param differs (e.g. containerName)", () => {
    const a = rule({ specSelector: "byName", selectorParams: { containerName: "a" } });
    const b = rule({ specSelector: "byName", selectorParams: { containerName: "b" } });
    expect(gvkPolicyUtil.ruleIdentityKey(a)).not.toBe(gvkPolicyUtil.ruleIdentityKey(b));
  });

  it("differs by field and by item selector", () => {
    expect(gvkPolicyUtil.ruleIdentityKey(rule({ field: "image" }))).not.toBe(
      gvkPolicyUtil.ruleIdentityKey(rule({ field: "command" })),
    );
    expect(gvkPolicyUtil.ruleIdentityKey(rule({ itemSelector: "x" }))).not.toBe(
      gvkPolicyUtil.ruleIdentityKey(rule({ itemSelector: "y" })),
    );
  });
});

describe("fillSingleOptionSelectorParams (single-option selector-param preselect)", () => {
  const enumParam = (name: string, enumValues: string[]): IMappingParam => ({
    name,
    type: "string" as MappingFieldType,
    enumValues,
  });
  const fieldInfoWith = (spec: IMappingParam[], item: IMappingParam[] = []): IFieldInfo => ({
    ...fieldInfo("string"),
    specSelectorParams: spec,
    itemSelectorParams: item,
  });

  it("preselects a spec param that has exactly one option", () => {
    const result = gvkPolicyUtil.fillSingleOptionSelectorParams(
      fieldInfoWith([enumParam("replicaType", ["Worker"])]),
      {},
    );
    expect(result.replicaType).toBe("Worker");
  });

  it("does not preselect a param with more than one option (the user must choose)", () => {
    const result = gvkPolicyUtil.fillSingleOptionSelectorParams(
      fieldInfoWith([enumParam("replicaType", ["Worker", "Launcher"])]),
      {},
    );
    expect(result.replicaType ?? "").toBe("");
  });

  it("does not preselect a free-text / numeric param (no options)", () => {
    const result = gvkPolicyUtil.fillSingleOptionSelectorParams(
      fieldInfoWith([{ name: "key", type: "string" as MappingFieldType }]),
      {},
    );
    expect(result.key ?? "").toBe("");
  });

  it("never overwrites a value the user already set", () => {
    const result = gvkPolicyUtil.fillSingleOptionSelectorParams(fieldInfoWith([enumParam("replicaType", ["Worker"])]), {
      replicaType: "Launcher",
    });
    expect(result.replicaType).toBe("Launcher");
  });

  it("preselects single-option item-selector params too", () => {
    const result = gvkPolicyUtil.fillSingleOptionSelectorParams(fieldInfoWith([], [enumParam("scope", ["All"])]), {});
    expect(result.scope).toBe("All");
  });

  it("does not mutate the passed-in params object", () => {
    const current = {};
    gvkPolicyUtil.fillSingleOptionSelectorParams(fieldInfoWith([enumParam("replicaType", ["Worker"])]), current);
    expect(current).toEqual({});
  });
});

describe("ruleSourceLabel (effective-rule source scope)", () => {
  const src = (over: Partial<SourceOfRule>): SourceOfRule => ({ scope: ScopeType.Project, ...over });

  it("maps each scope level to a label", () => {
    expect(gvkPolicyUtil.ruleSourceLabel(src({ scope: ScopeType.Project }))).toBe("Project");
    expect(gvkPolicyUtil.ruleSourceLabel(src({ scope: ScopeType.Department }))).toBe("Department");
    expect(gvkPolicyUtil.ruleSourceLabel(src({ scope: ScopeType.Cluster }))).toBe("Cluster");
    expect(gvkPolicyUtil.ruleSourceLabel(src({ scope: ScopeType.Tenant }))).toBe("Tenant");
  });

  it("labels a fallback source as 'Fallback' regardless of scope", () => {
    expect(gvkPolicyUtil.ruleSourceLabel(src({ scope: ScopeType.Cluster, isFallback: true }))).toBe("Fallback");
  });

  it("returns an empty string for a missing source", () => {
    expect(gvkPolicyUtil.ruleSourceLabel(null)).toBe("");
    expect(gvkPolicyUtil.ruleSourceLabel(undefined)).toBe("");
  });
});
