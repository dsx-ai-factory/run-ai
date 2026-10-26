/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import type { BootConfig } from "@/swagger-models/bff-service-client";

const getBootConfigResponse = vi.hoisted(() => vi.fn());

vi.mock("@/swagger-models/bff-service-client", () => ({
  Configuration: class Configuration {},
  BootConfigApi: class BootConfigApi {
    getBootConfig = getBootConfigResponse;
  },
}));

describe("V2 deployment documentation", () => {
  beforeEach(() => {
    vi.resetModules();
    getBootConfigResponse.mockReset();
  });

  it.each(["cc", "ce"] as const)("uses the configured V2 domain and mapping with profile %s", async (profile) => {
    const baseUrl = "https://docs.nvidia.com/run-ai/self-hosted/v3.0";
    const config: BootConfig = {
      distributionAssets: { profile },
      version: { display: "v3.0.0" },
      docs: {
        catalogVersion: "v2",
        baseUrl,
        whatsNew: {
          label: "What's New in 3.0",
          url: `${baseUrl}/documentation/getting-started/whats-new/whats-new-3-0`,
        },
      },
      support: { label: "Get help" },
      capabilities: { showsClusterLocationStep: true },
      unleashConfig: null,
    };
    getBootConfigResponse.mockResolvedValue({ data: config });

    const { loadBootConfig } = await import("@/services/infra/boot-config.service/boot-config.service");
    await expect(loadBootConfig()).resolves.toEqual(config);
    const { DOCS, resolveDocsLinksInHtml } = await import("./docs");

    expect(DOCS.ROOT).toBe(baseUrl);
    expect(DOCS.AUTHENTICATION.OVERVIEW).toBe(`${baseUrl}/documentation/platform-management/authentication/overview`);
    expect(DOCS.CLI.INSTALL).toBe(`${baseUrl}/reference/cli/install-and-configure-cli`);
    expect(DOCS.WHATS_NEW).toEqual({ URL: config.docs.whatsNew.url, LABEL: config.docs.whatsNew.label });

    const template = document.createElement("template");
    template.innerHTML = resolveDocsLinksInHtml('<a data-doc-key="workloadTemplates">Guide</a>');
    expect(template.content.querySelector("a")?.getAttribute("href")).toBe(
      `${baseUrl}/documentation/workloads-in-nvidia-run-ai/workload-templates`,
    );
  });
});
