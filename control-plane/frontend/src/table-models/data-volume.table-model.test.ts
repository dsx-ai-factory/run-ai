/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it } from "vitest";
import DOCS from "@/constants/docs";
import type { IStatusColOptions } from "@/models/table.model";
import { DatavolumePhase, type DatavolumeNoSharedScopes } from "@/swagger-models/datavolumes-service-client";
import { allDataVolumeColumnsMap } from "./data-volume.table-model";

const support = DOCS.SUPPORT as { URL: string };
const originalSupportUrl = support.URL;
const SUPPORT_LINK =
  '<a href="https://support.example.test/contact" target="_blank" rel="noopener noreferrer">Contact NVIDIA Run:ai support.</a>';

function formatStatus(phase: DatavolumePhase, phaseMessage?: string): IStatusColOptions {
  const row = { createdAt: "2020-01-01T00:00:00Z", status: { phase, phaseMessage } } as DatavolumeNoSharedScopes;
  const format = allDataVolumeColumnsMap.status.format as (status: unknown, row: unknown) => IStatusColOptions;
  return format(row.status, row);
}

describe("data volume status support link", () => {
  afterEach(() => {
    support.URL = originalSupportUrl;
  });

  it("appends the support link to the red Issues found tooltip", () => {
    support.URL = "https://support.example.test/contact";

    const result = formatStatus(DatavolumePhase.NotReady, "OriginalPvcNotBound");

    expect(result.status).toBe("Issues found");
    expect(result.tooltipText).toBe(`Original PVC not bound<br/>${SUPPORT_LINK}`);
  });

  it("leaves the tooltip unchanged when the support URL is not configured", () => {
    support.URL = "";

    const result = formatStatus(DatavolumePhase.NotReady, "OriginalPvcNotBound");

    expect(result.tooltipText).toBe("Original PVC not bound");
  });

  it.each([DatavolumePhase.Ready, DatavolumePhase.Creating, DatavolumePhase.Deleting, DatavolumePhase.Unknown])(
    "does not add the support link for %s",
    (phase) => {
      support.URL = "https://support.example.test/contact";

      expect(formatStatus(phase, "Some message").tooltipText ?? "").not.toContain("support");
    },
  );
});
