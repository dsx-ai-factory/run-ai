/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it } from "vitest";
import DOCS from "@/constants/docs";
import { AssetSyncStatusEnum } from "@/swagger-models/assets-service-client";
import { assetsUtil } from "./assets.util";

const support = DOCS.SUPPORT as { URL: string };
const originalSupportUrl = support.URL;

describe("asset status support CTA", () => {
  afterEach(() => {
    support.URL = originalSupportUrl;
  });

  it("removes support copy and preserves punctuation when the support URL is missing", () => {
    support.URL = "";

    const result = assetsUtil.getStatusColOptions(
      {
        status: AssetSyncStatusEnum.IssuesFound,
        message: "The asset status couldn't be determined",
        url: "https://legacy-support.example.test/contact",
      },
      "2020-01-01T00:00:00Z",
    );

    expect(result.tooltipText).toBe("The asset status couldn't be determined.");
    expect(result.tooltipText).not.toContain("support");
  });

  it("uses the configured support URL for the existing clickable CTA", () => {
    support.URL = "https://support.example.test/contact";

    const result = assetsUtil.getStatusColOptions(
      {
        status: AssetSyncStatusEnum.IssuesFound,
        message: "The asset status couldn't be determined",
        url: "https://legacy-support.example.test/contact",
      },
      "2020-01-01T00:00:00Z",
    );

    expect(result.tooltipText).toContain('href="https://support.example.test/contact"');
    expect(result.tooltipText).not.toContain("legacy-support.example.test");
  });
});
