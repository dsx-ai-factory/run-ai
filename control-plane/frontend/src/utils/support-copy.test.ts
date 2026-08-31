/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it } from "vitest";
import DOCS from "@/constants/docs";
import { dashboardUtil } from "@/utils/dashboard.util/dashboard.util";
import { settingsUtil } from "@/utils/settings.util/settings.util";

const support = DOCS.SUPPORT as { URL: string };
const originalSupportUrl = support.URL;

describe("support-dependent notification copy", () => {
  afterEach(() => {
    support.URL = originalSupportUrl;
  });

  it("removes support clauses when the support URL is missing", () => {
    support.URL = "";

    expect(dashboardUtil.getCsvErrorMessage().message).toBe("CSV couldn't be downloaded. Try again.");
    expect(settingsUtil.getSettingErrorMessage().message).toBe("Settings couldn't be saved. Try again.");
  });

  it("retains clickable support clauses when the support URL is configured", () => {
    support.URL = "https://support.example.test/contact";

    expect(dashboardUtil.getCsvErrorMessage().message).toContain(
      '<a href="https://support.example.test/contact" target="_blank" rel="noopener noreferrer">contact Run:ai support</a>',
    );
    expect(settingsUtil.getSettingErrorMessage().message).toContain(
      '<a href="https://support.example.test/contact" target="_blank" rel="noopener noreferrer">contact Run:ai support</a>',
    );
  });
});
