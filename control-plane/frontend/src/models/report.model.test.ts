/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import { isReportDeletionInProgress, remainingDeletingIds, reportBulkDeleteCopy } from "./report.model";
import DOCS from "@/constants/docs";

vi.mock("@/constants/docs", () => ({
  default: { SUPPORT: { URL: "", LABEL: "Run:ai support" } },
}));

// DOCS is declared `as const`, so the mock has to be reached through a mutable view.
const docsSupport = DOCS.SUPPORT as { URL: string; LABEL: string };
import { ReportStatus } from "@/swagger-models/org-unit-service-client";

describe("reportBulkDeleteCopy", () => {
  describe("confirmation dialog", () => {
    it("states how many reports will be deleted", () => {
      expect(reportBulkDeleteCopy.title(2)).toBe("Permanently delete 2 reports?");
    });

    it("puts the count on the confirm button", () => {
      expect(reportBulkDeleteCopy.confirmLabel(2)).toBe("Delete 2 reports");
    });
  });

  describe("failure message", () => {
    it("names the report when a single one failed", () => {
      expect(reportBulkDeleteCopy.failure("my-report", 1)).toBe("Report my-report couldn't be deleted");
    });

    it("appends the remaining count when more than one failed", () => {
      expect(reportBulkDeleteCopy.failure("my-report", 4)).toBe("Report my-report +3 couldn't be deleted");
    });

    it("counts the named report itself, so two failures render as +1", () => {
      expect(reportBulkDeleteCopy.failure("my-report", 2)).toBe("Report my-report +1 couldn't be deleted");
    });
  });
});

describe("isReportDeletionInProgress", () => {
  it("is true only while the deletion is still running", () => {
    expect(isReportDeletionInProgress(ReportStatus.PendingDelete)).toBe(true);
    expect(isReportDeletionInProgress(ReportStatus.Deleting)).toBe(true);
  });

  it("is false for DeletionFailed, whose row stays in the grid", () => {
    expect(isReportDeletionInProgress(ReportStatus.DeletionFailed)).toBe(false);
  });

  it("is false for every non-deletion status", () => {
    [ReportStatus.Pending, ReportStatus.Processing, ReportStatus.Ready, ReportStatus.Failed].forEach(
      (status: ReportStatus) => expect(isReportDeletionInProgress(status)).toBe(false),
    );
  });
});

describe("reportBulkDeleteCopy.requestFailure", () => {
  afterEach(() => {
    docsSupport.URL = "";
  });

  it("points at support when a support URL is configured", () => {
    docsSupport.URL = "https://support.example.com";
    expect(reportBulkDeleteCopy.requestFailure()).toBe(
      "Reports couldn't be deleted. Try again. If the issue persists, contact Run:ai support.",
    );
  });

  it("omits the support sentence when no URL is configured", () => {
    docsSupport.URL = "";
    expect(reportBulkDeleteCopy.requestFailure()).toBe("Reports couldn't be deleted. Try again.");
  });
});

describe("remainingDeletingIds", () => {
  const statuses = (entries: Record<string, ReportStatus>) => {
    const byId = new Map<string, ReportStatus>(Object.entries(entries));
    return async (id: string) => byId.get(id) ?? null;
  };

  it("keeps only the ids whose deletion is still running", async () => {
    const remaining = await remainingDeletingIds(
      ["a", "b", "c"],
      statuses({ a: ReportStatus.Deleting, b: ReportStatus.PendingDelete, c: ReportStatus.Ready }),
    );
    expect(remaining).toEqual(["a", "b"]);
  });

  it("drops ids that no longer exist, which is deletion having completed", async () => {
    const remaining = await remainingDeletingIds(["a", "gone"], statuses({ a: ReportStatus.Deleting }));
    expect(remaining).toEqual(["a"]);
  });

  it("drops ids that ended in DeletionFailed so polling can stop", async () => {
    const remaining = await remainingDeletingIds(["a"], statuses({ a: ReportStatus.DeletionFailed }));
    expect(remaining).toEqual([]);
  });

  it("drops ids whose lookup rejects rather than polling forever", async () => {
    const remaining = await remainingDeletingIds(["a"], () => Promise.reject(new Error("boom")));
    expect(remaining).toEqual([]);
  });

  // The grid is filtered, so a report being deleted can vanish from it while the
  // deletion is still running. The status lookup must not depend on the grid.
  it("still reports a deletion in progress when the grid no longer lists it", async () => {
    const remaining = await remainingDeletingIds(["filtered-out"], statuses({ "filtered-out": ReportStatus.Deleting }));
    expect(remaining).toEqual(["filtered-out"]);
  });
});
