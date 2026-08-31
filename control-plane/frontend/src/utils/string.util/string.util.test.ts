/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import {
  shallowObjectToQueryString,
  spaceToDash,
  toLowerCase,
  getStringAfterFirstDot,
  formatCamelCaseToWords,
  parseCamelCaseAndCapitalize,
  replaceAll,
  isString,
  stringUtil,
  splitCamelCaseAndCapitalize,
} from "./string.util";

describe("string.util", () => {
  describe("spaceToDash", () => {
    it("Should convert space to dash", () => {
      expect(spaceToDash("camel Case")).toEqual("camel-Case");
      expect(spaceToDash("camelcase")).toEqual("camelcase");
      expect(spaceToDash("Camel Case")).toEqual("Camel-Case");
    });
  });

  describe("toLowerCase", () => {
    it("Should convert the characters to lowercase", () => {
      expect(toLowerCase("camelCase")).toEqual("camelcase");
      expect(toLowerCase("camelcase")).toEqual("camelcase");
      expect(toLowerCase("Camel Case")).toEqual("camel case");
    });
  });
  describe("shallowObjectToQueryString", () => {
    it("Should convert the an object to query string", () => {
      expect(shallowObjectToQueryString({ first: "test1", second: "test2" })).toEqual("?first=test1&second=test2");
      expect(shallowObjectToQueryString({ name: "person", age: "12", color: "red" })).toEqual(
        "?name=person&age=12&color=red",
      );
      expect(shallowObjectToQueryString({ test: "success" })).toEqual("?test=success");
    });
  });

  describe("getStringAfterFirstDot", () => {
    it("should return the string after the first dot", () => {
      expect(getStringAfterFirstDot("example.input.string")).toBe("input.string");
      expect(getStringAfterFirstDot("one.dot.only")).toBe("dot.only");
    });

    it("should return the original string if there are no dots", () => {
      expect(getStringAfterFirstDot("no_dots_here")).toBe("no_dots_here");
    });

    it("should handle an empty string", () => {
      expect(getStringAfterFirstDot("")).toBe("");
    });
  });

  describe("formatCamelCaseToWords", () => {
    it("should convert CamelCase to words with spaces", () => {
      expect(formatCamelCaseToWords("hostPath")).toBe("host path");
      expect(formatCamelCaseToWords("someLongCamelCaseString")).toBe("some long camel case string");
      expect(formatCamelCaseToWords("AnotherExample")).toBe("another example");
    });

    it("should handle input with no capital letters", () => {
      expect(formatCamelCaseToWords("noupper")).toBe("noupper");
    });
  });

  describe("parseCamelCase", () => {
    it("should parse camel case to separate words", () => {
      const camelCaseString = "TrainingWorkload";
      const expectedResult = "Training Workload";
      const result = parseCamelCaseAndCapitalize(camelCaseString);
      expect(result).toEqual(expectedResult);
    });

    it("should handle camel case with acronyms", () => {
      const input = "XMLHttpRequest";
      const expectedResult = "XML Http Request";
      const result = parseCamelCaseAndCapitalize(input);
      expect(result).toEqual(expectedResult);
    });

    it("should handle all uppercase camel case", () => {
      const input = "APIRequest";
      const expectedResult = "API Request";
      const result = parseCamelCaseAndCapitalize(input);
      expect(result).toEqual(expectedResult);
    });

    it("should handle single letter camel case", () => {
      const input = "aSimpleExample";
      const expectedResult = "A Simple Example";
      const result = parseCamelCaseAndCapitalize(input);
      expect(result).toEqual(expectedResult);
    });
  });
  describe("replaceAll", () => {
    it("should replace all occurrences of a substring within a string", () => {
      const source = "hello world hello";
      const match = "hello";
      const replace = "hi";
      const expectedResult = "hi world hi";

      expect(replaceAll(source, match, replace)).toEqual(expectedResult);
    });

    it("should handle empty source string", () => {
      const source = "";
      const match = "hello";
      const replace = "hi";
      const expectedResult = "";

      expect(replaceAll(source, match, replace)).toEqual(expectedResult);
    });

    it("should handle empty match and replace strings", () => {
      const source = "hello world";
      const match = "";
      const replace = "";
      const expectedResult = "hello world";

      expect(replaceAll(source, match, replace)).toEqual(expectedResult);
    });

    it("should handle empty replace string", () => {
      const source = "hello world";
      const match = "world";
      const replace = "";
      const expectedResult = "hello ";

      expect(replaceAll(source, match, replace)).toEqual(expectedResult);
    });

    it("should handle when match is not found", () => {
      const source = "hello world";
      const match = "foo";
      const replace = "bar";
      const expectedResult = "hello world";

      expect(replaceAll(source, match, replace)).toEqual(expectedResult);
    });

    it("should handle when source, match, and replace are not provided", () => {
      const expectedResult = "";

      expect(replaceAll()).toEqual(expectedResult);
    });
  });

  describe("isString", () => {
    it("check if variable is string - real string", () => {
      expect(isString("string")).true;
    });
    it("check if variable is string- empty string", () => {
      expect(isString("")).true;
    });
    it("check if variable is string - null", () => {
      expect(isString(null)).false;
    });
    it("check if variable is string - undefined", () => {
      expect(isString(undefined)).false;
    });
    it("check if variable is string - number", () => {
      expect(isString(7)).false;
    });
  });

  describe("slug", () => {
    it("should convert text to lowercase", () => {
      const result = stringUtil.slug("Hello");
      expect(result).toBe("hello");
    });

    it("should replace spaces with hyphens", () => {
      const result = stringUtil.slug("Hello World");
      expect(result).toBe("hello-world");
    });

    it("should trim leading and trailing spaces", () => {
      const result = stringUtil.slug(" Hello World ");
      expect(result).toBe("hello-world");
    });

    it("should replace spaces with a custom separator", () => {
      const result = stringUtil.slug("Hello World", "_");
      expect(result).toBe("hello_world");
    });

    it("should convert camelCase to kebab-case", () => {
      const result = stringUtil.slug("ProjectPriority");
      expect(result).toBe("project-priority");
    });

    it("should convert camelCase to kebab-case", () => {
      const result = stringUtil.slug("projectPriority");
      expect(result).toBe("project-priority");
    });

    it("should convert camelCase with acronyms to kebab-case", () => {
      const result = stringUtil.slug("XMLHttpRequest");
      expect(result).toBe("xmlhttp-request");
    });
  });
  describe("toBoolean", () => {
    it("should convert a string to a boolean", () => {
      expect(stringUtil.toBoolean("true")).toBe(true);
      expect(stringUtil.toBoolean("false")).toBe(false);
      expect(stringUtil.toBoolean("")).toBe(false);
      expect(stringUtil.toBoolean("foo")).toBe(false);
      expect(stringUtil.toBoolean(undefined)).toBe(false);
      expect(stringUtil.toBoolean(false)).toBe(false);
      expect(stringUtil.toBoolean(true)).toBe(true);
    });
  });

  describe("splitCamelCaseAndCapitalize", () => {
    it("should split camel case and capitalize only the first letter", () => {
      expect(splitCamelCaseAndCapitalize("WaitForConnect")).toBe("Wait for connect");
      expect(splitCamelCaseAndCapitalize("ServiceIssues")).toBe("Service issues");
      expect(splitCamelCaseAndCapitalize("Ready")).toBe("Ready");
      expect(splitCamelCaseAndCapitalize("aSimpleExample")).toBe("A simple example");
    });
    it("should handle empty string", () => {
      expect(splitCamelCaseAndCapitalize("")).toBe("");
    });
    it("should handle single word", () => {
      expect(splitCamelCaseAndCapitalize("Single")).toBe("Single");
    });
  });

  describe("truncateString", () => {
    it("should return the string as is when length is less than maxLength", () => {
      expect(stringUtil.truncateString("Hello", 10)).toBe("Hello");
      expect(stringUtil.truncateString("Short", 20)).toBe("Short");
    });

    it("should return the string as is when length equals maxLength", () => {
      expect(stringUtil.truncateString("Hello", 5)).toBe("Hello");
      expect(stringUtil.truncateString("Test String", 11)).toBe("Test String");
    });

    it("should truncate string and add default ellipsis when length exceeds maxLength", () => {
      expect(stringUtil.truncateString("Hello World", 8)).toBe("Hello...");
      expect(stringUtil.truncateString("This is a long string", 10)).toBe("This is...");
    });

    it("should handle edge case with maxLength of 3", () => {
      expect(stringUtil.truncateString("Hello", 3)).toBe("...");
    });

    it("should handle edge case with maxLength less than 3", () => {
      expect(stringUtil.truncateString("Hello", 2)).toBe("...");
      expect(stringUtil.truncateString("Hello", 1)).toBe("...");
    });

    it("should work with empty string", () => {
      expect(stringUtil.truncateString("", 10)).toBe("");
      expect(stringUtil.truncateString("", 0)).toBe("");
    });

    it("should work with special characters", () => {
      expect(stringUtil.truncateString("Hello! @#$ World", 10)).toBe("Hello! ...");
      expect(stringUtil.truncateString("Test-String_123", 10)).toBe("Test-St...");
    });

    it("should truncate string with custom suffix", () => {
      expect(stringUtil.truncateString("Hello World", 8, ">>")).toBe("Hello >>");
      expect(stringUtil.truncateString("This is a long string", 10, "…")).toBe("This is a…");
      expect(stringUtil.truncateString("Test String", 10, " ...")).toBe("Test S ...");
    });

    it("should handle custom suffix longer than maxLength", () => {
      expect(stringUtil.truncateString("Hello", 3, "-----")).toBe("-----");
      expect(stringUtil.truncateString("Testing", 4, "....")).toBe("....");
    });

    it("should work with empty custom suffix", () => {
      expect(stringUtil.truncateString("Hello World", 5, "")).toBe("Hello");
      expect(stringUtil.truncateString("Test String", 8, "")).toBe("Test Str");
    });

    it("should handle single character custom suffix", () => {
      expect(stringUtil.truncateString("Hello World", 8, ".")).toBe("Hello W.");
      expect(stringUtil.truncateString("Testing", 5, "!")).toBe("Test!");
    });
  });
});
