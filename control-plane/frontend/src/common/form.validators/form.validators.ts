import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { errorMessages } from "../error-message.constant";
import type { ValidationRule } from "quasar";
import type { IAssetNameValidationOptions } from "@/models/asset.model";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import type { IValidationRule, ScopeTypeSystemExcluded } from "@/models/global.model";

export const formValidators = {
  isValidUrl,
  isValidEmail,
  isValidEntityName,
  isNumberGreaterOrEqualToZero,
};

const MAX_K8S_NAME_CHARACTERS = 63;
export function isEmptyObject(val: object | undefined | null): boolean {
  return !val || Object.keys(val).length === 0;
}

export function isEmpty(val: string | undefined | null): boolean {
  return !val || val.length === 0;
}

export function isNotEmpty(val: string): boolean {
  return !!val && val.length > 0 && val.trim().length > 0;
}

export function isNotEmptyArray<T>(val: T[]): boolean {
  return !!val && val.length > 0;
}

export function required(val: string): boolean {
  return !!val;
}

export function isLowerCaseAndNumbers(val: string): boolean {
  return !!val && /^[a-z][-a-z0-9]*[a-z0-9]$/.test(val);
}

export function isLowerCaseAndNumbersAndUnderscore(val: string): boolean {
  return !!val && /^[a-z][-_a-z0-9]*[a-z0-9]$/.test(val);
}

export function isValidEntityName(str: string) {
  return !!str && !!str.match(/^[a-z][-a-z0-9]*[a-z0-9]?$/) && !str.endsWith("-");
}

export function isValidKubernetesNamespaceName(name: string): boolean {
  if (name.length > MAX_K8S_NAME_CHARACTERS) {
    return false;
  }
  return isValidEntityName(name);
}

export function isFirstLetterIsLowerCase(val: string): boolean {
  return !!val && val[0] === val[0].toLowerCase();
}

export function isValidRunaiEntityNameFormat(val: string): boolean {
  return isLowerCaseAndNumbers(val) && isFirstLetterIsLowerCase(val);
}

export function isValidNumber(val: string): boolean {
  return parseFloat(val) >= 0;
}

export function isNumberGreaterThanZero(val: number): boolean {
  return typeof val === "number" && val > 0;
}

function isNumberGreaterOrEqualToZero(val: number): boolean {
  return typeof val === "number" && val >= 0;
}

export function isNumberInRange(val: number, ge: number, le: number): boolean {
  return val > ge && val < le;
}

export function isMaxEqualOrHigherThenMin(max: number, min: number): boolean {
  return max >= min;
}

export function isValueLowerOrEqualToMax(value: number, max: number): boolean {
  return value <= max;
}

export function isValidDirectoryPath(val: string): boolean {
  return /^(\/(?:[a-zA-Z\d._\-{} ]+\/?)*)$/.test(val);
}

export function isValidSubPath(val: string): boolean {
  const regex = /^(?!\/)(?!.*(?:^|\/)\.\.(?:\/|$))[^\0]+$/;
  return regex.test(val);
}

export function isValidUrl(val: string): boolean {
  const regex =
    /^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()!@:%_\+.~#?&\/\/=]*)$/; //eslint-disable-line
  return regex.test(val);
}

/** Valid for Helm chart URLs: https/http or oci:// (Helm OCI registry). */
export function isValidChartUrl(val: string): boolean {
  if (isValidUrl(val)) return true;
  const ociRegex = /^oci:\/\/[a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})?(:[0-9]+)?(\/[a-zA-Z0-9._\-/:]+)*\/?$/;
  return ociRegex.test(val);
}

export function isValidPrivateUrl(val: string): boolean {
  // Handle Git SSH URLs (e.g., git@github.com:username/repo.git)
  if (val.includes("@") && val.includes(":")) {
    // Support both formats: git@host:path and git@host:port:path
    // Must have at least one character in the path part after the colon
    // Path cannot end with a slash (incomplete path)
    const sshRegex =
      /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(:\d+)?:[a-zA-Z0-9._/-]+[a-zA-Z0-9._/-]*[a-zA-Z0-9._-]$/;
    if (sshRegex.test(val)) {
      return true;
    }
  }

  // Handle standard URLs with more permissive path validation
  const urlRegex =
    /^(https?:\/\/)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}(:\d+)?(\/[a-zA-Z0-9._\-~:/?#[\]@!$&'()*+,;=%]*)?\/?$/;
  return urlRegex.test(val);
}

export function isValidDockerServerUrl(val: string): boolean {
  const regex = /^(https?:\/\/)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(:\d+)?(\/[a-zA-Z0-9._\-/]*)?\/?$/;
  return regex.test(val);
}

export function isValidDomain(val: string): boolean {
  return /^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(val);
}

export function isValidTopologyKey(val: string): boolean {
  return /^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?(?:\/[a-z0-9._-]*[a-z0-9])?$/.test(val);
}

export function isValidRegExp(val: string, regexp: string): boolean {
  return new RegExp(regexp).test(val);
}

export function isValidPassword(val: string): boolean {
  const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#$%^&*])(?=.{8,})/; //eslint-disable-line
  return regex.test(val);
}

export function isAssetNameUnique(kind: string, options: IAssetNameValidationOptions): ValidationRule {
  return async (val: string) => {
    const isExist: boolean = await assetService.isExist(kind, val, options);
    return isExist ? errorMessages.NAME_ALREADY_EXIST : true;
  };
}

export function isFlatTemplateNameUnique(scopeType: ScopeTypeSystemExcluded, scopeId: string): IValidationRule {
  return async (val: string) => {
    try {
      const res = await workloadTemplateService.templateNameAvailability(val, scopeType, scopeId);
      return res;
    } catch (error: unknown) {
      return errorMessages.NAME_ALREADY_EXIST;
    }
  };
}

// This function checks that the string contains only numbers seperated by comma or semicolon
// Example: '1,2,3;4;5'
// Invalid example 'a;2;3,b'
export function isCommaSemicolonSeperatedNumber(val: string): boolean {
  return !val || !!val.match(/^[0-9]+([,;][0-9]+)*$/);
}

export function isValidEmail(email: string): boolean {
  return (
    !!email &&
    /^[a-zA-Z0-9_%+-](?:[a-zA-Z0-9._%+-]*[a-zA-Z0-9_%+-])?@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(email) &&
    !/(\.\.)/.test(email)
  );
}

export function haggingFaceTokenValid(token: string): boolean {
  return /^hf_[a-zA-Z0-9]+$/.test(token);
}

export function huggingFaceModelNameValid(name: string): boolean {
  return /^[a-zA-Z0-9-_]+(\x2F[a-zA-Z0-9-_.]+)?$/.test(name);
}

/**
 *
 * @param name the name of the model
 * @returns true if the name is valid, false otherwise
 */
export function nimModelNameValidator(name: string): boolean {
  return /^[A-Za-z0-9_\-./]+$/.test(name);
}

/**
 *
 * @param name the profile name or hash
 * @returns true if the name/ hash is valid, false otherwise
 */
export function nimProfileValidator(name: string): boolean {
  return /^[A-Za-z0-9_-]+$/.test(name);
}

/**
 *
 * @param str the string to check
 * @returns true if the string contains only alphanumeric characters, false otherwise
 */
export function isAlphaNumericOnly(str: string): boolean {
  return /^[A-Za-z0-9]+$/.test(str);
}

enum EStateEnum {
  StateStart,
  StateArg,
  StateQuotes,
}

export function validCommandAndArgsLine(command: string | null): boolean {
  let state = EStateEnum.StateStart;
  let quote = "";
  let escapeNext = false;

  if (command === null) return true;
  if (command.length === 0) return false;

  for (let i = 0; i < command.length; i++) {
    const currChar: string = command[i];

    if (currChar === "\\") {
      escapeNext = true;
      continue;
    }

    if (escapeNext) {
      escapeNext = false;
      continue;
    }

    if (state === EStateEnum.StateQuotes) {
      if (currChar === quote) {
        state = EStateEnum.StateStart;
      }
      continue;
    }

    if (currChar === '"' || currChar == "'") {
      state = EStateEnum.StateQuotes;
      quote = currChar;
      continue;
    }

    if (state == EStateEnum.StateArg) {
      if (currChar === " " || currChar === "\t") {
        state = EStateEnum.StateStart;
      }
      continue;
    }

    if (currChar !== " " && currChar !== "\t") {
      state = EStateEnum.StateArg;
    }
  }
  if (state === EStateEnum.StateQuotes) {
    return false;
  }
  return true;
}

export function isValidNgcApiKey(val: string): boolean {
  return /^[-._A-Za-z0-9]+$/.test(val);
}
