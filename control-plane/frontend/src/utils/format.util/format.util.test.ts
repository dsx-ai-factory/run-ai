/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import {
  memoryFormat,
  numberFormat,
  numberWithPercentFormat,
  onlyPercentFormat,
  toPercent,
  roundToDecimal,
  convertToBytes,
  parseSizeString,
  formatToFixedOrZero,
  cpuParser,
  memoryParser,
  formatK8sResourceMemory,
  formatCompactNumber,
} from "./format.util";

describe("format.util", () => {
  describe("format.util::memoryFormat", () => {
    it("should return an empty string when gets one of the options: -1, '-1'", () => {
      const negative = -1;
      const negativeStr = "-1";

      expect(memoryFormat(negative)).toBe("");
      expect(memoryFormat(negativeStr)).toBe("");
    });

    it("should return '0 Bytes' when gets one of the options: '0', 0, NaN, empty string and a string", () => {
      const zero = 0;
      const zeroStr = "0";
      const nan = NaN;
      const emptyStr = "";
      const str = "asdsa";
      expect(memoryFormat(zero)).toBe("0 Bytes");
      expect(memoryFormat(zeroStr)).toBe("0 Bytes");
      expect(memoryFormat(nan)).toBe("0 Bytes");
      expect(memoryFormat(emptyStr)).toBe("0 Bytes");
      expect(memoryFormat(str)).toBe("0 Bytes");
    });

    it("should return '-' when gets one of the options: '-', undefined.", () => {
      const dashStr = "-";
      expect(memoryFormat(undefined)).toBe("-");
      expect(memoryFormat(dashStr)).toBe("-");
    });

    it("memoryFormat should return the correct memory unit converted from bytes", () => {
      const bytes = 999;
      const kb = 1001;
      const mb = 1000120;
      const gb = 10001200000;
      const tb = 100012000000343;
      const pb = 1200120000003431;
      const eb = 120040000000000000000;
      const zb = 1500000000000000000000;
      const yb = 15410000000000000000000000;

      expect(memoryFormat(bytes)).toBe("999 Bytes");
      expect(memoryFormat(kb)).toBe("1 KB");
      expect(memoryFormat(mb)).toBe("1 MB");
      expect(memoryFormat(gb)).toBe("10 GB");
      expect(memoryFormat(tb)).toBe("100.01 TB");
      expect(memoryFormat(pb)).toBe("1.2 PB");
      expect(memoryFormat(eb)).toBe("120.04 EB");
      expect(memoryFormat(zb)).toBe("1.5 ZB");
      expect(memoryFormat(yb)).toBe("15.41 YB");
    });
  });

  describe("format.util::numberFormat", () => {
    it("should return 'N/A' when gets one of the options: -1, '-1', 'NA'", () => {
      const negative = -1;
      const negativeStr = "-1";
      const noValue = "NA";

      expect(numberFormat(negative)).toBe("N/A");
      expect(numberFormat(negativeStr)).toBe("N/A");
      expect(numberFormat(noValue)).toBe("N/A");
    });

    it("should return '-' when gets one of the options: NaN, or string that is not number", () => {
      const str = "dsada";
      const nan = NaN;

      expect(numberFormat(str)).toBe("-");
      expect(numberFormat(nan)).toBe("-");
    });

    it("should return the value as a string with exactly 2 digits after the dots", () => {
      const val1 = 23;
      const val2 = 34.3453;
      const val3 = "23";
      const val4 = "0";

      expect(numberFormat(val1)).toBe("23.00");
      expect(numberFormat(val2)).toBe("34.35");
      expect(numberFormat(val3)).toBe("23.00");
      expect(numberFormat(val4)).toBe("0.00");
    });

    it("should return a string of the value with a suffix", () => {
      const val1 = 34.356;
      const val2 = 65;

      expect(numberFormat(val1, false, "%")).toBe("34.36%");
      expect(numberFormat(val2, false, "%")).toBe("65.00%");
    });

    it("should return the value as a string without formatting if it is an integer and formatOnlyIfFloat is true", () => {
      const intVal = 23;
      const intStrVal = "23";
      expect(numberFormat(intVal, true)).toBe("23");
      expect(numberFormat(intStrVal, true)).toBe("23");
    });

    it("should return the formatted value if it is a float and formatOnlyIfFloat is true", () => {
      const floatVal = 23.456;
      const floatStrVal = "23.456";
      expect(numberFormat(floatVal, true)).toBe("23.46");
      expect(numberFormat(floatStrVal, true)).toBe("23.46");
    });

    it("should return the formatted value if formatOnlyIfFloat is false, regardless of whether the value is an integer or float", () => {
      const intVal = 23;
      const floatVal = 23.456;
      expect(numberFormat(intVal, false)).toBe("23.00");
      expect(numberFormat(floatVal, false)).toBe("23.46");
    });
  });

  describe("format.util::numberWithPercentFormat", () => {
    it("should return the formatted value and percentage", () => {
      expect(numberWithPercentFormat(0, 2)).toBe("0.00");
      expect(numberWithPercentFormat(1, 2)).toBe("1.00 (50.00%)");
    });

    it("should return the formatted memory value with percentage", () => {
      expect(numberWithPercentFormat(0, 4194304000, true)).toBe("0 Bytes");
      expect(numberWithPercentFormat(2097152000, 4194304000, true)).toBe("2.1 GB (50.00%)");
    });
  });

  describe("format.util::onlyPercentFormat", () => {
    it("should return the formatted percentage", () => {
      expect(onlyPercentFormat(0, 2)).toBe("0.00%");
      expect(onlyPercentFormat(1, 2)).toBe("50.00%");
    });
    it("should return - for null val", () => {
      expect(onlyPercentFormat(null, 3)).toBe("-");
    });
  });

  describe("format.util::toPercent", () => {
    it("should return the calculated percentage of two values", () => {
      expect(toPercent(40, 80)).toBe("50.00%");
      expect(toPercent(1, 3)).toBe("33.33%");
      expect(toPercent(0, 3)).toBe("0.00%");
      expect(toPercent(3, 3)).toBe("100.00%");
    });
  });

  describe("format.util::roundToDecimal", () => {
    it("should round a number to the default number of decimal places", () => {
      expect(roundToDecimal(1.2345)).toBe(1.23);
      expect(roundToDecimal(1.2399)).toBe(1.24);
      expect(roundToDecimal(1.2)).toBe(1.2);
      expect(roundToDecimal(1.0)).toBe(1.0);
    });

    it("should round a number to the maximum number of decimal places specified", () => {
      expect(roundToDecimal(1.2345, 3)).toBe(1.234);
      expect(roundToDecimal(1.2345, 0)).toBe(1);
      expect(roundToDecimal(1.2399, 1)).toBe(1.2);
    });

    it("should return the same number if it is already an integer", () => {
      expect(roundToDecimal(5)).toBe(5);
      expect(roundToDecimal(10, 3)).toBe(10);
      expect(roundToDecimal(0, 0)).toBe(0);
    });

    it("should handle negative numbers correctly", () => {
      expect(roundToDecimal(-1.2345)).toBe(-1.23);
      expect(roundToDecimal(-1.2399)).toBe(-1.24);
      expect(roundToDecimal(-1.2)).toBe(-1.2);
    });
  });

  describe("format.util::convertToBytes", () => {
    it("should convert a number to bytes", () => {
      expect(convertToBytes(1, "B")).toBe(1);
      expect(convertToBytes(1, "K")).toBe(1000);
      expect(convertToBytes(1, "KB")).toBe(1000);
      expect(convertToBytes(1, "M")).toBe(1000000);
      expect(convertToBytes(1, "MB")).toBe(1000000);
      expect(convertToBytes(1, "G")).toBe(1000000000);
      expect(convertToBytes(1, "GB")).toBe(1000000000);
      expect(convertToBytes(1, "T")).toBe(1000000000000);
      expect(convertToBytes(1, "TB")).toBe(1000000000000);
      expect(convertToBytes(1, "P")).toBe(1000000000000000);
      expect(convertToBytes(1, "PB")).toBe(1000000000000000);
      expect(convertToBytes(1, "EB")).toBe(1000000000000000000);
      expect(convertToBytes(1, "ZB")).toBe(1000000000000000000000);
      expect(convertToBytes(1, "YB")).toBe(1000000000000000000000000);
    });

    it("should handle lowercase unit strings", () => {
      expect(convertToBytes(1, "b")).toBe(1);
      expect(convertToBytes(1, "k")).toBe(1000);
      expect(convertToBytes(1, "m")).toBe(1000000);
      expect(convertToBytes(1, "g")).toBe(1000000000);
      expect(convertToBytes(1, "t")).toBe(1000000000000);
      expect(convertToBytes(1, "p")).toBe(1000000000000000);
      expect(convertToBytes(1, "eb")).toBe(1000000000000000000);
      expect(convertToBytes(1, "zb")).toBe(1000000000000000000000);
      expect(convertToBytes(1, "yb")).toBe(1000000000000000000000000);
    });

    it("should throw an error for invalid units", () => {
      expect(() => convertToBytes(1, "X")).toThrowError("Invalid unit: X");
      expect(() => convertToBytes(1, "ZK")).toThrowError("Invalid unit: ZK");
      expect(() => convertToBytes(1, "")).toThrowError("Invalid unit: ");
    });
  });

  describe("format.util::parseSizeString", () => {
    it("should parse size strings with units correctly", () => {
      expect(parseSizeString("10GB")).toEqual([10, "GB"]);
      expect(parseSizeString("10MB")).toEqual([10, "MB"]);
      expect(parseSizeString("10KB")).toEqual([10, "KB"]);
      expect(parseSizeString("1.5TB")).toEqual([1.5, "TB"]);
      expect(parseSizeString("0.75EB")).toEqual([0.75, "EB"]);
    });

    it("should handle lowercase units", () => {
      expect(parseSizeString("10gb")).toEqual([10, "GB"]);
      expect(parseSizeString("10mb")).toEqual([10, "MB"]);
      expect(parseSizeString("10kb")).toEqual([10, "KB"]);
    });

    it("should handle decimals in size", () => {
      expect(parseSizeString("1.2345GB")).toEqual([1.2345, "GB"]);
      expect(parseSizeString("0.75EB")).toEqual([0.75, "EB"]);
    });

    it("should return null for invalid input", () => {
      expect(parseSizeString("")).toBeNull(); // Empty string
      expect(parseSizeString("invalid")).toBeNull(); // No size or unit
      expect(parseSizeString("10")).toBeNull(); // No unit
      expect(parseSizeString("GB")).toBeNull(); // No size
      expect(parseSizeString("10 G")).toBeNull(); // Space between size and unit
      expect(parseSizeString("10.5X")).toBeNull(); // Invalid unit
    });
  });

  describe("format.util::formatToFixedOrZero", () => {
    it("should format a number with two decimal places", () => {
      expect(formatToFixedOrZero(10)).toBe(10);
      expect(formatToFixedOrZero(10.1234)).toBeCloseTo(10.12);
      expect(formatToFixedOrZero(0.1234)).toBeCloseTo(0.12);
      expect(formatToFixedOrZero(0)).toBe(0);
      expect(formatToFixedOrZero(0.005)).toBe(0.01);
    });

    it('should return zero for numbers rounded to "0.00"', () => {
      expect(formatToFixedOrZero(0.001)).toBe(0);
      expect(formatToFixedOrZero(0.0)).toBe(0);
      expect(formatToFixedOrZero(-0.001)).toBe(0);
      expect(formatToFixedOrZero(-0.0)).toBe(0);
    });

    it("should handle negative numbers", () => {
      expect(formatToFixedOrZero(-0.005)).toBe(-0.01);
      expect(formatToFixedOrZero(-10.1234)).toBeCloseTo(-10.12);
      expect(formatToFixedOrZero(-0.1234)).toBeCloseTo(-0.12);
    });
  });
  describe("a cpu parser", () => {
    const tests = [
      ["parses full numbers", "1", 1],
      ["parses floats (< 1)", "1.5", 1.5],
      ["parses floats (> 1)", "0.5", 0.5],
      ["parses strings with milli (m) unit (whole number)", "1000m", 1],
      ["parses strings with milli (m) unit (decimal number)", "1300m", 1.3],
      ["parses strings with milli (m) unit (< 1)", "300m", 0.3],
    ];

    tests.map((t) =>
      it(`${t[0]} (${t[1]} to ${t[2]})`, () => {
        expect(cpuParser(t[1] as string)).toEqual(t[2]);
      }),
    );
  });

  describe("a memory parser", () => {
    const tests = [
      ["parses full numbers", "1", 1],
      ["parses kilo strings", "1k", 1000],
      ["parses Mega strings", "2M", 2 * 1000 ** 2],
      ["parses Giga strings", "3G", 3 * 1000 ** 3],
      ["parses Tera strings", "4T", 4 * 1000 ** 4],
      ["parses Peta strings", "5P", 5 * 1000 ** 5],
      ["parses Exa strings", "6E", 6 * 1000 ** 6],
      ["parses kibi strings", "1Ki", 1024],
      ["parses Mebi strings", "2Mi", 2 * 1024 ** 2],
      ["parses Gibi strings", "3Gi", 3 * 1024 ** 3],
      ["parses Tebi strings", "4Ti", 4 * 1024 ** 4],
      ["parses Pebi strings", "5Pi", 5 * 1024 ** 5],
      ["parses Exbi strings", "6Ei", 6 * 1024 ** 6],
    ];

    tests.map((t) =>
      it(`${t[0]} (${t[1]} to ${t[2]})`, () => {
        expect(memoryParser(t[1] as string)).toEqual(t[2]);
      }),
    );
  });
  describe("a formatK8sResourceMemory parser", () => {
    const tests = [
      ["parses Kilo strings", "10000", "10 KB"],
      ["parses Mega strings", "200M", "200 MB"],
      ["parses Mega strings", "2M", "2 MB"],
      ["parses Mebi strings", "123Mi", "128.97 MB"],
      ["parses Mega strings", "129M", "129 MB"],
      ["parses Mega strings", "128974848", "128.97 MB"],
    ];

    tests.map((t: string[]) =>
      it(`${t[0]} (${t[1]} to ${t[2]})`, () => {
        expect(formatK8sResourceMemory(t[1] as string)).toEqual(t[2]);
      }),
    );
  });

  describe("format.util::formatCompactNumber", () => {
    it("should format numbers in millions with 'm' suffix", () => {
      expect(formatCompactNumber(1000000)).toBe("1.0m");
      expect(formatCompactNumber(1835100)).toBe("1.8m");
      expect(formatCompactNumber(7745400)).toBe("7.7m");
      expect(formatCompactNumber(2300000)).toBe("2.3m");
    });

    it("should format numbers in thousands with 'k' suffix", () => {
      expect(formatCompactNumber(1000)).toBe("1.0k");
      expect(formatCompactNumber(89400)).toBe("89.4k");
      expect(formatCompactNumber(500000)).toBe("500.0k");
      expect(formatCompactNumber(1500)).toBe("1.5k");
    });

    it("should format small numbers without suffix", () => {
      expect(formatCompactNumber(100)).toBe("100");
      expect(formatCompactNumber(999)).toBe("999");
      expect(formatCompactNumber(1)).toBe("1");
    });

    it("should handle zero and undefined values", () => {
      expect(formatCompactNumber(0)).toBe("");
      expect(formatCompactNumber(undefined)).toBe("");
      expect(formatCompactNumber(null as unknown as number)).toBe("");
    });
  });
});
