import { ref } from "vue";
import { gvkPolicyService } from "@/services/control-plane/gvk-policy.service/gvk-policy.service";
import { EEnforcementType, type IFieldInfo, type IMappingParam } from "@/models/gvk-policy.model";
import type {
  BreakdownFieldVariation,
  BreakdownFormatter,
  Breakdown,
  GetMappingMinimalResp,
  GVK,
  MappingFieldType,
} from "@/swagger-models/policy-service-client";

const JQ_NAMED_ARG = /\$ARGS\.named\.([A-Za-z_][\w]*)/g;

// Module-scope (not instance) so the cache survives the rule-builder's unmount/remount on builder<->YAML switches.
const minimalCache = new Map<string, GetMappingMinimalResp>();
function gvkCacheKey(gvk: GVK): string {
  return `${gvk.group ?? ""}/${gvk.version}/${gvk.kind}`;
}

// Filter out values a newer backend might add but this client doesn't know, so they can't break the enforcement dropdown.
const ENFORCEMENT_TYPE_VALUES = new Set<string>(Object.values(EEnforcementType));
function toAllowedEnforcements(possible: readonly string[] | undefined): EEnforcementType[] {
  return (possible ?? []).filter((value): value is EEnforcementType => ENFORCEMENT_TYPE_VALUES.has(value));
}

function extractJqArgs(expr: string | null | undefined): string[] {
  if (!expr) return [];
  const names = new Set<string>();
  let match: RegExpExecArray | null;
  JQ_NAMED_ARG.lastIndex = 0;
  while ((match = JQ_NAMED_ARG.exec(expr)) !== null) {
    names.add(match[1]);
  }
  return [...names];
}

function paramsFromArgs(
  argsInfo: { [key: string]: MappingFieldType } | undefined,
  argsEnum: { [key: string]: Array<string> } | undefined,
  required: boolean,
): IMappingParam[] {
  return Object.keys(argsInfo ?? {}).map((name) => ({
    name,
    type: argsInfo?.[name] ?? "string",
    enumValues: argsEnum?.[name],
    required,
  }));
}

function computeSetParams(formatter: BreakdownFormatter | undefined): IMappingParam[] {
  if (!formatter) return [];
  if (formatter.builtin === "KeyValueMap") {
    return [
      { name: "key", type: "string", required: false },
      { name: "value", type: "string", required: true },
    ];
  }
  if (formatter.setExpr) {
    const arrayKeyArg = formatter.arrayKey?.argName;
    return extractJqArgs(formatter.setExpr).map((name) => ({
      name,
      type: formatter.argsInfo?.[name] ?? "string",
      enumValues: formatter.argsEnum?.[name],
      required: name === arrayKeyArg,
    }));
  }
  return [];
}

function setDefaultGroupingArg(formatter: BreakdownFormatter | undefined): string | null {
  if (formatter?.arrayKey?.argName) return formatter.arrayKey.argName;
  // KeyValueMap has an implicit key argument in the backend formatter contract.
  return formatter?.builtin === "KeyValueMap" ? "key" : null;
}

export function usePolicyMapping() {
  const minimal = ref<GetMappingMinimalResp | null>(null);
  const loadingMinimal = ref<boolean>(false);
  const fieldInfoCache = new Map<string, IFieldInfo>();
  const fieldBreakdownCache = new Map<string, Breakdown>();
  const fieldBreakdownRequests = new Map<string, Promise<Breakdown>>();
  let currentGvk: GVK | null = null;

  async function loadMinimal(gvk: GVK): Promise<void> {
    currentGvk = gvk;
    fieldInfoCache.clear();
    fieldBreakdownCache.clear();
    fieldBreakdownRequests.clear();
    const cached = minimalCache.get(gvkCacheKey(gvk));
    if (cached) {
      minimal.value = cached;
      return;
    }
    minimal.value = null;
    loadingMinimal.value = true;
    try {
      const resp = await gvkPolicyService.getMappingMinimal(gvk);
      minimalCache.set(gvkCacheKey(gvk), resp);
      minimal.value = resp;
    } finally {
      loadingMinimal.value = false;
    }
  }

  function fields(): string[] {
    return Object.keys(minimal.value?.fields ?? {}).sort((a, b) => a.localeCompare(b));
  }

  function specSelectorsForField(field: string | null): string[] {
    if (!field) return [];
    return Object.keys(minimal.value?.fields?.[field] ?? {}).sort((a, b) => a.localeCompare(b));
  }

  function itemSelectorsFor(field: string | null, specSelector: string | null): string[] {
    if (!field || !specSelector) return [];
    return Object.keys(minimal.value?.fields?.[field]?.[specSelector] ?? {})
      .filter((name) => name !== "")
      .sort((a, b) => a.localeCompare(b));
  }

  function specSelectorDescription(name: string | null): string {
    if (!name) return "";
    return minimal.value?.specSelectors?.[name] ?? "";
  }

  async function getFieldBreakdown(field: string): Promise<Breakdown> {
    if (!currentGvk) return {};
    const cacheKey = `${gvkCacheKey(currentGvk)}|${field}`;
    const cached = fieldBreakdownCache.get(cacheKey);
    if (cached) return cached;
    const pending = fieldBreakdownRequests.get(cacheKey);
    if (pending) return pending;

    const request = gvkPolicyService.getMappingFieldInfo({
      workloadType: currentGvk,
      fieldsRef: [{ fieldName: field }],
    });
    fieldBreakdownRequests.set(cacheKey, request);
    try {
      const response = await request;
      fieldBreakdownCache.set(cacheKey, response);
      return response;
    } finally {
      fieldBreakdownRequests.delete(cacheKey);
    }
  }

  async function loadFieldInfo(field: string): Promise<void> {
    await getFieldBreakdown(field);
  }

  async function resolveFieldInfo(
    field: string,
    specSelector: string,
    itemSelector: string | null,
  ): Promise<IFieldInfo | null> {
    if (!currentGvk) return null;
    const cacheKey = `${field}|${specSelector}|${itemSelector ?? ""}`;
    const cached = fieldInfoCache.get(cacheKey);
    if (cached) return cached;

    const resp = await getFieldBreakdown(field);

    const breakdown = resp.breakdown;
    const variations: BreakdownFieldVariation[] = breakdown?.fields?.[field] ?? [];
    const variation =
      variations.find((candidate) => (candidate.specSelectors ?? []).includes(specSelector)) ?? variations[0];
    if (!variation) return null;

    const type: MappingFieldType = variation.type ?? "string";
    const formatter = variation.formatter;
    const specSelectorMeta = breakdown?.specSelectors?.[specSelector];
    const itemSelectorMeta = itemSelector ? variation.itemSelectors?.[itemSelector] : undefined;

    const info: IFieldInfo = {
      type,
      builtin: formatter?.builtin ?? null,
      setDefaultGroupingArg: setDefaultGroupingArg(formatter),
      itemSelectors: Object.keys(variation.itemSelectors ?? {}),
      allowedEnforcements: toAllowedEnforcements(variation.possibleEnforcements),
      setParams: computeSetParams(formatter),
      // Backend-verified: validateSelectorParams validates only provided params, never rejects a missing one, so selector params must stay optional here.
      specSelectorParams: paramsFromArgs(specSelectorMeta?.argsInfo, specSelectorMeta?.argsEnum, false),
      itemSelectorParams: paramsFromArgs(itemSelectorMeta?.argsInfo, itemSelectorMeta?.argsEnum, false),
    };

    fieldInfoCache.set(cacheKey, info);
    return info;
  }

  return {
    minimal,
    loadingMinimal,
    loadMinimal,
    fields,
    specSelectorsForField,
    itemSelectorsFor,
    specSelectorDescription,
    loadFieldInfo,
    resolveFieldInfo,
  };
}

export type PolicyMapping = ReturnType<typeof usePolicyMapping>;
