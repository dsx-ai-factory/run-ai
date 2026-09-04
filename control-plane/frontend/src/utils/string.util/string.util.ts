/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const stringUtil = {
  parse,
  toString,
  spaceToDash,
  toLowerCase,
  generateTempPassword,
  getStringAfterFirstDot,
  formatCamelCaseToWords,
  parseCamelCaseAndCapitalize,
  stringToBase64,
  camelCaseToHyphenated,
  slug,
  toBoolean,
  replaceNewLineWithBreaks,
  capitalizeString,
  truncateString,
  getHighlightedTextSegments,
};

export interface HighlightedTextSegment {
  text: string;
  highlighted: boolean;
}

export function getHighlightedTextSegments(text: string, searchTerm = ""): HighlightedTextSegment[] {
  const normalizedSearchTerm = searchTerm.trim().toLocaleLowerCase();
  if (!normalizedSearchTerm) return [{ text, highlighted: false }];

  const normalizedText = text.toLocaleLowerCase();
  const segments: HighlightedTextSegment[] = [];
  let currentIndex = 0;
  let matchIndex = normalizedText.indexOf(normalizedSearchTerm);

  while (matchIndex !== -1) {
    if (matchIndex > currentIndex) {
      segments.push({ text: text.slice(currentIndex, matchIndex), highlighted: false });
    }
    const matchEnd = matchIndex + normalizedSearchTerm.length;
    segments.push({ text: text.slice(matchIndex, matchEnd), highlighted: true });
    currentIndex = matchEnd;
    matchIndex = normalizedText.indexOf(normalizedSearchTerm, currentIndex);
  }

  if (currentIndex < text.length) {
    segments.push({ text: text.slice(currentIndex), highlighted: false });
  }

  return segments.length > 0 ? segments : [{ text, highlighted: false }];
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function toString(data: any) {
  if (!data) return "";
  return JSON.stringify(data);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function parse(data: any) {
  if (!data) return data;
  return JSON.parse(data);
}

export function spaceToDash(word: string): string {
  return word.replace(/[ ]/g, "-");
}

export function toLowerCase(word: string): string {
  return word.toLowerCase();
}

export function generateTempPassword(): string {
  return Math.random().toString(36).slice(-8).concat("Ii2!");
}

export function capitalizeString(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function shallowObjectToQueryString(obj: Record<string, string>): string {
  return "?" + new URLSearchParams(obj).toString();
}

export function getStringAfterFirstDot(input: string): string {
  const parts = input.split(".");
  return parts.length > 1 ? parts.slice(1).join(".") : input;
}

export function formatCamelCaseToWords(input: string): string {
  return input
    .replace(/([A-Z])/g, " $1")
    .toLowerCase()
    .trim();
}

export function parseCamelCaseAndCapitalize(input: string): string {
  return capitalizeString(input.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/([A-Z])([A-Z][a-z])/g, "$1 $2"));
}

export function getEntityWithArticle(entity: string): string {
  const vowels = ["a", "e", "i", "o", "u"];
  const article = vowels.includes(entity.toLocaleLowerCase()[0]) ? "an" : "a";

  return `${article} ${entity}`;
}

export function replaceAll(source = "", match = "", replace = ""): string {
  return source.split(match).join(replace);
}

export function stringToBase64(input: string): string {
  return btoa(input);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function isString(input: any): boolean {
  return input instanceof String || typeof input === "string";
}

/**
 * Converts a camelCase string to a hyphenated string.
 *
 * @param {string} str - The camelCase string to be converted.
 * @returns {string} The hyphenated version of the input string.
 *
 * @example
 * // returns "hello-world"
 * camelCaseToHyphenated("helloWorld");
 */
function camelCaseToHyphenated(str: string): string {
  return str.replace(/[A-Z]/g, function (match, offset) {
    return offset > 0 ? "-" + match.toLowerCase() : match.toLowerCase();
  });
}

/**
 * Converts a string to a slug.
 *
 * This function takes a string as input, converts all the characters in the string to lowercase, trims any leading or trailing spaces, and replaces all spaces in the string with the specified separator (default is a hyphen).
 *
 * @param {string} text - The string to be converted to a slug.
 * @param {string} [separator="-"] - The separator to replace spaces with. Default is "-".
 * @returns {string} The slug version of the input string.
 *
 * @example
 * // returns "hello-world"
 * slug("Hello World");
 */
function slug(text: string, separator = "-"): string {
  return text
    .replace(/([a-z])([A-Z])/g, `$1${separator}$2`) // Insert separator before each uppercase letter
    .toLowerCase()
    .trim()
    .replace(/ /g, separator); // Replace spaces with the separator
}

export function toBoolean(value: string | boolean | undefined): boolean {
  return value === "true" || value === true;
}

/**
 * Replaces all newline characters in a string with HTML line break tags.
 *
 * This function takes a string as input and replaces every occurrence of the newline character (`\n`)
 * with the HTML line break tag (`<br>`). This is useful for rendering multi-line text in HTML contexts
 * where newline characters need to be converted to line breaks.
 *
 * @param {string} text - The input string containing newline characters.
 * @returns {string} The modified string with newline characters replaced by `<br>` tags.
 *
 * @example
 * // returns "Hello<br>World"
 * replaceNewLineWithBreaks("Hello\nWorld");
 */
function replaceNewLineWithBreaks(text: string): string {
  return text.replace(/\n/g, "<br>");
}

export function splitCamelCaseAndCapitalize(str: string): string {
  if (!str) return "";
  const spaced = str.replace(/([a-z])([A-Z])/g, "$1 $2").toLowerCase();
  return spaced.charAt(0).toUpperCase() + spaced.slice(1);
}

/**
 * Truncates a string to a specified maximum length.
 *
 * If the string length exceeds the maximum length, it is cut at the specified index
 * and the suffix is appended. If the string length is less than or equal to the maximum
 * length, the string is returned as is.
 *
 * @param {string} str - The string to be truncated.
 * @param {number} maxLength - The maximum length of the string (including the suffix).
 * @param {string} [suffix="..."] - The suffix to append when truncating. Default is "..." (ellipsis).
 * @returns {string} The truncated string with suffix appended if it exceeds maxLength, otherwise the original string.
 *
 * @example
 * // returns "Hello World"
 * truncateString("Hello World", 15);
 *
 * @example
 * // returns "Hello..."
 * truncateString("Hello World", 8);
 *
 * @example
 * // returns "Hello>>"
 * truncateString("Hello World", 8, ">>");
 *
 * @example
 * // returns "Hello Wor"
 * truncateString("Hello World", 9, "");
 */
function truncateString(str: string, maxLength: number, suffix = "..."): string {
  if (str.length <= maxLength) {
    return str;
  }
  const cutLength = Math.max(0, maxLength - suffix.length);
  return str.slice(0, cutLength) + suffix;
}
