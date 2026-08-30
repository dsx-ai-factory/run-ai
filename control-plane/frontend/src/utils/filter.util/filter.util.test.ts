import { describe, test, expect } from "vitest";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";

describe("filter.util", () => {
  test("parseEnumToSelectOptions", () => {
    enum ReportStatus {
      Pending = "pending",
      Processing = "processing",
    }
    const result = filterUtil.parseEnumToSelectOptions(ReportStatus);

    expect(result).toEqual([
      { label: "Pending", value: "Pending" },
      { label: "Processing", value: "Processing" },
    ]);
  });

  test("getFreeTextFilterQuery", () => {
    const name = "workloadName";
    const term = "test";
    const label = "Workload Name";
    const operator = EFilterOperator.Contains;

    const result = filterUtil.getFreeTextFilterQuery(name, term, label, operator);

    expect(result).toBe(
      `{ "name": "${name}", "term": "${term}", "label": "${label}", "value": "${operator}", "type": "${EColumnFilterType.FreeText}" }`,
    );
  });
});
