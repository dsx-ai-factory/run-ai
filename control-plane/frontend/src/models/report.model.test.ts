/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { isReportDeletionInProgress, reportBulkDeleteCopy } from "./report.model";
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
