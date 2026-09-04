/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable  @typescript-eslint/no-explicit-any */
export const policyModelUtil = {
  policyRulesToPreviewString,
  policyDefaultToPreviewString,
  formatObjectAdvanced,
};

function policyRulesToPreviewString(val: any): string {
  if (!val) return "";
  return Object.entries(val).reduce((acc, [key, value]) => {
    if (_isObject(value)) {
      acc += `• ${key}:<br />&nbsp;`;
      acc += _objectToPreviewString(value as object, undefined, 2);
    } else if (Array.isArray(value)) {
      acc += `• ${key} = `;
      acc += `${value.map((item) => item.value || item).join(", ")}<br />`;
    } else {
      acc += `• ${key} = ${value} <br />`;
    }
    return acc;
  }, "");
}

function policyDefaultToPreviewString(val: any[] | string | number | boolean): string {
  if (!val && typeof val !== "boolean") return "";
  if (!Array.isArray(val)) return val.toString();
  return val.reduce((acc, v, idx) => {
    acc += `•`;
    if (_isObject(v)) {
      acc += _objectToPreviewString(v);
      if (idx !== val.length - 1) acc += "<br />";
    } else {
      acc += ` ${v} <br />`;
    }
    return acc;
  }, "");
}

function _objectToPreviewString(val: object, outterKey?: string, forcedSpacing = 0): string {
  return Object.entries(val).reduce((acc, [key, value], idx) => {
    const combinedKey = outterKey ? `${outterKey}.${key}` : key;
    if (_isObject(value)) {
      acc += _objectToPreviewString(value, combinedKey);
    } else {
      const isFirstRow = !outterKey && idx === 0;
      const spacing = "&nbsp;".repeat(isFirstRow ? forcedSpacing : forcedSpacing + 1);
      acc += `${spacing} ${combinedKey} = ${value} <br />`;
    }
    return acc;
  }, "");
}

function _isObject(value: any): boolean {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

// this function needs to be considered with product for future use
function formatObjectAdvanced(data: Record<string, any>, indent = 0): string {
  let result = "";

  const space = "&nbsp;";
  const newLine = "<br />";
  const prefix = space.repeat(indent);

  for (const key in data) {
    if (data[key]) {
      if (Array.isArray(data[key])) {
        // Handle arrays
        result += `${prefix}• ${key} (array)${newLine}`;
        data[key].forEach((item: any, index: number) => {
          result += `${prefix}${space.repeat(4)}• [${index}]${newLine}`; // Display array index
          result +=
            typeof item === "object"
              ? formatObjectAdvanced(item, indent + 8) // Process nested object in array with increased indent
              : `${prefix}${space.repeat(8)}${item}${newLine}`; // Primitive values
        });
      } else if (typeof data[key] === "object") {
        // Handle nested objects
        result += `${prefix}• ${key}${newLine}`;
        result += formatObjectAdvanced(data[key], indent + 4);
      } else {
        // Handle primitive values
        result += `${prefix}• ${key} = ${data[key]}${newLine}`;
      }
    }
  }

  return result;
}
