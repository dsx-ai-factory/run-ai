import type {
  GVK,
  MappingFieldType,
  PolicyRule,
  PolicyRuleEnforcementsEnforcementsInner,
  SourceOfRule,
} from "@/swagger-models/policy-service-client";
import { EEnforcementType, type IEnforcementDraft, type IFieldInfo, type IRuleDraft } from "@/models/gvk-policy.model";
import { ScopeType } from "@/swagger-models/policy-service-client";
import { resourceUtil } from "@/utils/resource.util";
import { errorMessages } from "@/common/error-message.constant";
import { HttpErrorResponse } from "@/models/http-response.model";

const GVK_KEY_PREFIX = "gvk:";

let draftIdCounter = 0;
function nextDraftId(): string {
  draftIdCounter += 1;
  return `gvk-draft-${draftIdCounter}`;
}

function gvkKey(gvk: GVK): string {
  // core resources (e.g. Pod) have an empty group; normalize null to "" to avoid a literal "null" key segment
  return `${GVK_KEY_PREFIX}${gvk.group ?? ""}/${gvk.version}/${gvk.kind}`;
}

function isGvkKey(value: string | null | undefined): boolean {
  return typeof value === "string" && value.startsWith(GVK_KEY_PREFIX);
}

function parseGvkKey(key: string | null | undefined): GVK | null {
  if (!isGvkKey(key)) return null;
  const [group, version, kind] = (key as string).slice(GVK_KEY_PREFIX.length).split("/");
  if (!version || !kind) return null;
  return { group: group ?? "", version, kind };
}

// Set/Default uniqueness is target-aware and is checked separately by setDefaultCombinationError.
function enforcementMayRepeat(type: EEnforcementType): boolean {
  return type === EEnforcementType.Set || type === EEnforcementType.Default;
}

function isSetOrDefault(type: EEnforcementType | null): boolean {
  return type === EEnforcementType.Set || type === EEnforcementType.Default;
}

function setDefaultCombinationError(enforcements: IEnforcementDraft[], fieldInfo: IFieldInfo): string | null {
  const candidates = enforcements.filter((enforcement) => isSetOrDefault(enforcement.type));
  if (candidates.length < 2) return null;

  const groupingArg = fieldInfo.setDefaultGroupingArg;
  if (!groupingArg) return errorMessages.DUPLICATE_SET_DEFAULT_TARGET;

  const groupingParam = fieldInfo.setParams.find((param) => param.name === groupingArg);
  const seenTargets = new Set<string>();
  for (const enforcement of candidates) {
    const rawTarget = Object.entries(enforcement.setParams).find(([name]) => name === groupingArg)?.[1];
    // Omitting the array key is the backend's fan-out form, which overlaps every concrete target.
    if (rawTarget === undefined || isBlank(rawTarget)) return errorMessages.DUPLICATE_SET_DEFAULT_TARGET;
    const convertedTarget = convertValueByType(rawTarget, groupingParam?.type ?? "string");
    const target = String(convertedTarget);
    if (seenTargets.has(target)) return errorMessages.DUPLICATE_SET_DEFAULT_TARGET;
    seenTargets.add(target);
  }
  return null;
}

function isActiveBooleanEnforcement(enforcement: IEnforcementDraft, type: EEnforcementType): boolean {
  return enforcement.type === type && enforcement.value !== "false";
}

function enforcementsConflict(enforcement: IEnforcementDraft, other: IEnforcementDraft): boolean {
  const blocked = isActiveBooleanEnforcement(enforcement, EEnforcementType.Blocked)
    ? enforcement
    : isActiveBooleanEnforcement(other, EEnforcementType.Blocked)
      ? other
      : null;
  if (!blocked) return false;

  const peer = blocked === enforcement ? other : enforcement;
  if (!peer.type || peer.type === EEnforcementType.Blocked) return false;
  return !(peer.type === EEnforcementType.Required && !isActiveBooleanEnforcement(peer, EEnforcementType.Required));
}

function enforcementTypeUnavailableWith(type: EEnforcementType, other: IEnforcementDraft): boolean {
  if (!isActiveBooleanEnforcement(other, EEnforcementType.Blocked)) return false;
  // Required can still be selected as `required: false`, which the backend permits with blocked=true.
  return type !== EEnforcementType.Blocked && type !== EEnforcementType.Required;
}

function enforcementCombinationError(enforcements: IEnforcementDraft[]): string | null {
  const hasConflict = enforcements.some((enforcement, index) =>
    enforcements.slice(index + 1).some((other) => enforcementsConflict(enforcement, other)),
  );
  return hasConflict ? errorMessages.INCOMPATIBLE_ENFORCEMENTS : null;
}

const MIN_MAX_ENFORCEMENTS = new Set<EEnforcementType>([
  EEnforcementType.Length,
  EEnforcementType.Elements,
  EEnforcementType.Items,
]);
const RANGE_ENFORCEMENTS = new Set<EEnforcementType>([EEnforcementType.Range, EEnforcementType.QuantityRange]);

function isMinMaxEnforcement(type: EEnforcementType | null): boolean {
  return type != null && MIN_MAX_ENFORCEMENTS.has(type);
}
function isRangeEnforcement(type: EEnforcementType | null): boolean {
  return type != null && RANGE_ENFORCEMENTS.has(type);
}

function enforcementValueError(enforcement: IEnforcementDraft): string | null {
  if (enforcement.type === EEnforcementType.Options && enforcement.options.length === 0) {
    return errorMessages.SELECT_A_VALUE;
  }
  if (isMinMaxEnforcement(enforcement.type) && isBlank(enforcement.min) && isBlank(enforcement.max)) {
    return errorMessages.ENFORCEMENT_NEEDS_MIN_OR_MAX;
  }
  if (
    isRangeEnforcement(enforcement.type) &&
    isBlank(enforcement.min) &&
    isBlank(enforcement.max) &&
    isBlank(enforcement.step)
  ) {
    return errorMessages.ENFORCEMENT_NEEDS_MIN_MAX_OR_STEP;
  }
  return null;
}

export type TEnforcementValueRule = (value: string) => boolean | string;

export interface IEnforcementValueRules {
  value: TEnforcementValueRule[];
  min: TEnforcementValueRule[];
  max: TEnforcementValueRule[];
  step: TEnforcementValueRule[];
}

function isValidK8sQuantity(value: string): boolean | string {
  if (isBlank(value)) return true;
  try {
    resourceUtil.parseK8sResourceSizeToBytes(value);
    return true;
  } catch {
    return errorMessages.INVALID_QUANTITY;
  }
}

function isValidIpv4(value: string): boolean {
  const parts = value.split(".");
  if (parts.length !== 4) return false;
  return parts.every((part) => {
    if (!/^(0|[1-9]\d{0,2})$/.test(part)) return false;
    return Number(part) <= 255;
  });
}

function isValidIpv6(value: string): boolean {
  if (!value.includes(":")) return false;
  const compressedParts = value.split("::");
  if (compressedParts.length > 2) return false;

  const segments = compressedParts.flatMap((part) => (part === "" ? [] : part.split(":")));
  let segmentCount = segments.length;
  const ipv4Segment = segments.find((segment) => segment.includes("."));

  if (ipv4Segment !== undefined) {
    if (segments.at(-1) !== ipv4Segment || !value.endsWith(ipv4Segment) || !isValidIpv4(ipv4Segment)) return false;
    segmentCount += 1; // An embedded IPv4 address occupies two IPv6 hextets.
  }

  const hexSegments = ipv4Segment !== undefined ? segments.slice(0, -1) : segments;
  if (!hexSegments.every((segment) => /^[0-9a-fA-F]{1,4}$/.test(segment))) return false;

  const isCompressed = compressedParts.length === 2;
  return isCompressed ? segmentCount < 8 : segmentCount === 8;
}

function isValidIpAddress(value: string): boolean | string {
  if (isBlank(value)) return true;
  return isValidIpv4(value) || isValidIpv6(value) || errorMessages.ENTER_VALID_IP_ADDRESS;
}

function mappingTypeValidationRules(fieldType: MappingFieldType): TEnforcementValueRule[] {
  switch (fieldType) {
    case "quantity":
      return [isValidK8sQuantity];
    case "ipaddress":
      return [isValidIpAddress];
    default:
      return [];
  }
}

function quantityMaxAboveMin(min: string): TEnforcementValueRule {
  return (max: string): boolean | string => {
    if (isBlank(min) || isBlank(max)) return true;
    try {
      if (resourceUtil.parseK8sResourceSizeToBytes(max) >= resourceUtil.parseK8sResourceSizeToBytes(min)) return true;
    } catch {
      return true;
    }
    return errorMessages.EQUAL_TO_OR_HIGHER_THAN_THE_MINIMUM;
  };
}

// Backend ValidateQuantityRangeBounds: step > 0 and, when max is set, step <= max-min (min defaults to 0).
function quantityStepWithinRange(min: string, max: string): TEnforcementValueRule {
  return (step: string): boolean | string => {
    if (isBlank(step)) return true;
    try {
      const stepBytes = resourceUtil.parseK8sResourceSizeToBytes(step);
      if (stepBytes <= 0) return errorMessages.STEP_MUST_BE_ABOVE_ZERO;
      if (isBlank(max)) return true;
      const minBytes = isBlank(min) ? 0 : resourceUtil.parseK8sResourceSizeToBytes(min);
      if (stepBytes <= resourceUtil.parseK8sResourceSizeToBytes(max) - minBytes) return true;
    } catch {
      return true; // malformed values are flagged by isValidK8sQuantity — don't double-report
    }
    return errorMessages.STEP_EXCEEDS_RANGE;
  };
}

function enforcementValidationRules(opts: {
  enforcementType: EEnforcementType | null;
  fieldType: MappingFieldType;
  min?: string;
  max?: string;
}): IEnforcementValueRules {
  const empty: IEnforcementValueRules = { value: [], min: [], max: [], step: [] };
  const { enforcementType, fieldType, min = "", max = "" } = opts;

  if (enforcementType === EEnforcementType.QuantityRange) {
    return {
      value: [],
      min: [isValidK8sQuantity],
      step: [isValidK8sQuantity, quantityStepWithinRange(min, max)],
      max: [isValidK8sQuantity, quantityMaxAboveMin(min)],
    };
  }

  const isSetOrDefault = enforcementType === EEnforcementType.Set || enforcementType === EEnforcementType.Default;
  if (isSetOrDefault) {
    return { ...empty, value: mappingTypeValidationRules(fieldType) };
  }

  return empty;
}

function splitCsv(value: string): string[] {
  return value
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part.length > 0);
}

function convertValueByType(raw: string, type: MappingFieldType): unknown {
  if (raw === "" || raw == null) return raw;
  switch (type) {
    case "boolean":
      return raw === "true";
    case "integer": {
      const n = parseInt(raw, 10);
      return Number.isNaN(n) ? raw : n;
    }
    case "number": {
      const n = parseFloat(raw);
      return Number.isNaN(n) ? raw : n;
    }
    case "strings":
    case "namesArray":
      return splitCsv(raw);
    case "integers":
      return splitCsv(raw).map((s) => {
        const n = parseInt(s, 10);
        return Number.isNaN(n) ? s : n;
      });
    case "numbers":
      return splitCsv(raw).map((s) => {
        const n = parseFloat(s);
        return Number.isNaN(n) ? s : n;
      });
    case "object":
      try {
        return JSON.parse(raw);
      } catch {
        return raw;
      }
    default:
      return String(raw);
  }
}

function convertScalarParam(raw: string, type?: MappingFieldType): unknown {
  if (raw === "" || raw == null) return raw;
  switch (type) {
    case "boolean":
      return raw === "true";
    case "integer": {
      const n = parseInt(raw, 10);
      return Number.isNaN(n) ? raw : n;
    }
    case "number": {
      const n = parseFloat(raw);
      return Number.isNaN(n) ? raw : n;
    }
    default:
      return String(raw);
  }
}

function isBlank(value: string | undefined | null): boolean {
  return value === undefined || value === null || value.trim() === "";
}

function createEmptyEnforcement(): IEnforcementDraft {
  return {
    id: nextDraftId(),
    type: null,
    value: "",
    setParams: {},
    options: [],
    min: "",
    max: "",
    step: "",
  };
}

function createEmptyRuleDraft(): IRuleDraft {
  return {
    id: nextDraftId(),
    field: null,
    specSelector: null,
    itemSelector: null,
    selectorParams: {},
    enforcements: [createEmptyEnforcement()],
  };
}

function compactIntRange(draft: IEnforcementDraft): { min?: number; max?: number } {
  const out: { min?: number; max?: number } = {};
  if (!isBlank(draft.min)) out.min = Number(draft.min);
  if (!isBlank(draft.max)) out.max = Number(draft.max);
  return out;
}

function compactNumberRange(draft: IEnforcementDraft): { min?: number; max?: number; step?: number } {
  const out: { min?: number; max?: number; step?: number } = {};
  if (!isBlank(draft.min)) out.min = Number(draft.min);
  if (!isBlank(draft.max)) out.max = Number(draft.max);
  if (!isBlank(draft.step)) out.step = Number(draft.step);
  return out;
}

function compactStringRange(draft: IEnforcementDraft): { min?: string; max?: string; step?: string } {
  const out: { min?: string; max?: string; step?: string } = {};
  if (!isBlank(draft.min)) out.min = draft.min.trim();
  if (!isBlank(draft.max)) out.max = draft.max.trim();
  if (!isBlank(draft.step)) out.step = draft.step.trim();
  return out;
}

function buildSetValue(draft: IEnforcementDraft, fieldInfo: IFieldInfo): { [key: string]: unknown } {
  if (fieldInfo.setParams.length > 0) {
    const obj: { [key: string]: unknown } = {};
    for (const param of fieldInfo.setParams) {
      const raw = draft.setParams[param.name];
      if (!isBlank(raw)) obj[param.name] = convertValueByType(raw, param.type);
    }
    return obj;
  }
  return { value: convertValueByType(draft.value, fieldInfo.type) };
}

function enforcementDraftToApi(
  draft: IEnforcementDraft,
  fieldInfo: IFieldInfo,
): PolicyRuleEnforcementsEnforcementsInner | null {
  switch (draft.type) {
    case EEnforcementType.Set:
      return { set: buildSetValue(draft, fieldInfo) };
    case EEnforcementType.Default:
      return { default: buildSetValue(draft, fieldInfo) };
    case EEnforcementType.Required:
      return { required: draft.value !== "false" };
    case EEnforcementType.Blocked:
      return { blocked: draft.value !== "false" };
    case EEnforcementType.Match:
      return { match: draft.value };
    case EEnforcementType.Options:
      return { options: [...draft.options] };
    case EEnforcementType.Length:
      return { length: compactIntRange(draft) };
    case EEnforcementType.Elements:
      return { elements: compactIntRange(draft) };
    case EEnforcementType.Items:
      return { items: compactIntRange(draft) };
    case EEnforcementType.Range:
      return { range: compactNumberRange(draft) };
    case EEnforcementType.QuantityRange:
      return { quantityRange: compactStringRange(draft) };
    default:
      return null;
  }
}

function ruleDraftToApi(draft: IRuleDraft, fieldInfo: IFieldInfo): PolicyRule {
  const rule: PolicyRule = {
    field: draft.field as string,
    specSelector: draft.specSelector as string,
    enforcements: draft.enforcements
      .map((enforcement) => enforcementDraftToApi(enforcement, fieldInfo))
      .filter((enforcement): enforcement is PolicyRuleEnforcementsEnforcementsInner => enforcement !== null),
  };

  if (draft.itemSelector) rule.itemSelector = draft.itemSelector;

  const selectorParams: { [key: string]: unknown } = {};
  const allSelectorParams = [...fieldInfo.specSelectorParams, ...fieldInfo.itemSelectorParams];
  for (const param of allSelectorParams) {
    const raw = draft.selectorParams[param.name];
    if (!isBlank(raw)) selectorParams[param.name] = convertScalarParam(raw, param.type);
  }
  if (Object.keys(selectorParams).length > 0) rule.selectorParams = selectorParams;

  return rule;
}

function detectEnforcementType(enforcement: PolicyRuleEnforcementsEnforcementsInner): EEnforcementType | null {
  for (const type of Object.values(EEnforcementType)) {
    if ((enforcement as Record<string, unknown>)[type] != null) return type;
  }
  return null;
}

function apiEnforcementToDraft(enforcement: PolicyRuleEnforcementsEnforcementsInner): IEnforcementDraft {
  const draft = createEmptyEnforcement();
  const type = detectEnforcementType(enforcement);
  draft.type = type;
  switch (type) {
    case EEnforcementType.Set:
    case EEnforcementType.Default: {
      const value = (type === EEnforcementType.Set ? enforcement.set : enforcement.default) ?? {};
      const keys = Object.keys(value);
      if (keys.length === 1 && keys[0] === "value") {
        draft.value = stringifyScalar(value.value);
      } else {
        for (const key of keys) draft.setParams[key] = stringifyScalar(value[key]);
      }
      break;
    }
    case EEnforcementType.Required:
      draft.value = enforcement.required === false ? "false" : "true";
      break;
    case EEnforcementType.Blocked:
      draft.value = enforcement.blocked === false ? "false" : "true";
      break;
    case EEnforcementType.Match:
      draft.value = enforcement.match ?? "";
      break;
    case EEnforcementType.Options:
      draft.options = [...(enforcement.options ?? [])];
      break;
    case EEnforcementType.Length:
    case EEnforcementType.Elements:
    case EEnforcementType.Items: {
      const range = (enforcement[type as "length" | "elements" | "items"] ?? {}) as { min?: number; max?: number };
      draft.min = range.min != null ? String(range.min) : "";
      draft.max = range.max != null ? String(range.max) : "";
      break;
    }
    case EEnforcementType.Range: {
      const range = enforcement.range ?? {};
      draft.min = range.min != null ? String(range.min) : "";
      draft.max = range.max != null ? String(range.max) : "";
      draft.step = range.step != null ? String(range.step) : "";
      break;
    }
    case EEnforcementType.QuantityRange: {
      const range = enforcement.quantityRange ?? {};
      draft.min = range.min ?? "";
      draft.max = range.max ?? "";
      draft.step = range.step ?? "";
      break;
    }
    default:
      break;
  }
  return draft;
}

function stringifyScalar(value: unknown): string {
  if (value == null) return "";
  if (Array.isArray(value)) return value.join(", ");
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

function apiRuleToDraft(rule: PolicyRule): IRuleDraft {
  const draft = createEmptyRuleDraft();
  draft.field = rule.field;
  draft.specSelector = rule.specSelector;
  draft.itemSelector = rule.itemSelector ?? null;
  draft.selectorParams = {};
  for (const [key, value] of Object.entries(rule.selectorParams ?? {})) {
    draft.selectorParams[key] = stringifyScalar(value);
  }
  draft.enforcements = (rule.enforcements ?? []).map((enforcement) => apiEnforcementToDraft(enforcement));
  if (draft.enforcements.length === 0) draft.enforcements = [createEmptyEnforcement()];
  return draft;
}

function ruleIdentityKey(rule: PolicyRule): string {
  const params = rule.selectorParams
    ? JSON.stringify(
        Object.keys(rule.selectorParams)
          .sort()
          .reduce<Record<string, unknown>>((acc, key) => {
            acc[key] = rule.selectorParams?.[key];
            return acc;
          }, {}),
      )
    : "";
  return `${rule.field}|${rule.specSelector}|${rule.itemSelector ?? ""}|${params}`;
}

// Preselect a selector param that has exactly one option (fills blanks only) so the user needn't pick the sole value; the field stays visible so the scope it sets is still shown.
function fillSingleOptionSelectorParams(fieldInfo: IFieldInfo, current: Record<string, string>): Record<string, string> {
  const filled: Record<string, string> = { ...current };
  for (const param of [...fieldInfo.specSelectorParams, ...fieldInfo.itemSelectorParams]) {
    if ((filled[param.name] ?? "") === "" && param.enumValues?.length === 1) {
      filled[param.name] = param.enumValues[0];
    }
  }
  return filled;
}

export const GVK_NON_PROJECT_SCOPE_TOOLTIP = "To select this, first select a project scope above";

function isGvkTypeDisabledForScope(scope: ScopeType | undefined | null): boolean {
  return !!scope && scope !== ScopeType.Project;
}

function shouldResetGvkOnScopeChange(opts: { isAnyWorkload: boolean; newScope: ScopeType | undefined | null }): boolean {
  return opts.isAnyWorkload && opts.newScope !== ScopeType.Project;
}

// Scope level an effective rule was defined at (its sourceOfRule), so the view can mark inherited
// rules. Returns the scope level (or "Fallback"); "" when the source is unknown.
function ruleSourceLabel(source: SourceOfRule | null | undefined): string {
  if (!source) return "";
  if (source.isFallback) return "Fallback";
  switch (source.scope) {
    case ScopeType.Project:
      return "Project";
    case ScopeType.Department:
      return "Department";
    case ScopeType.Cluster:
      return "Cluster";
    case ScopeType.Tenant:
      return "Tenant";
    default:
      return source.scope ?? "";
  }
}

function isNetworkOrUnknownError(error: unknown): boolean {
  if (!(error instanceof HttpErrorResponse)) return true;
  return error.statusCode <= 0 || error.statusCode >= 500;
}

export const gvkPolicyUtil = {
  gvkKey,
  isGvkKey,
  parseGvkKey,
  enforcementMayRepeat,
  isSetOrDefault,
  setDefaultCombinationError,
  enforcementsConflict,
  enforcementTypeUnavailableWith,
  enforcementCombinationError,
  isMinMaxEnforcement,
  isRangeEnforcement,
  enforcementValueError,
  isValidK8sQuantity,
  isValidIpAddress,
  mappingTypeValidationRules,
  enforcementValidationRules,
  isGvkTypeDisabledForScope,
  shouldResetGvkOnScopeChange,
  isNetworkOrUnknownError,
  convertValueByType,
  convertScalarParam,
  isBlank,
  createEmptyEnforcement,
  createEmptyRuleDraft,
  ruleDraftToApi,
  ruleIdentityKey,
  fillSingleOptionSelectorParams,
  ruleSourceLabel,
  enforcementDraftToApi,
  apiRuleToDraft,
};
