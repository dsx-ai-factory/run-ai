/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";

const bootConfigState = vi.hoisted(() => ({
  docsBaseUrl: "https://docs.example.test/product/2.26",
}));

vi.mock("@/services/infra/boot-config.service/boot-config.service", () => ({
  getBootConfig: () => ({
    distributionAssets: { profile: "cc" },
    version: { display: "release-42" },
    docs: {
      baseUrl: bootConfigState.docsBaseUrl,
      whatsNew: { label: "Product updates", url: `${bootConfigState.docsBaseUrl}/updates` },
    },
    support: { url: "", label: "Get help" },
    capabilities: { showsClusterLocationStep: true },
    unleashConfig: null,
  }),
}));

import { getDocsUrl, resolveDocsLinksInHtml } from "./docs";

describe("documentation URL resolution", () => {
  beforeEach(() => {
    bootConfigState.docsBaseUrl = "https://docs.example.test/product/2.26";
  });

  it("builds a documentation URL from the configured base and topic path", () => {
    expect(getDocsUrl("getting-started/overview#cluster")).toBe(
      "https://docs.example.test/product/2.26/getting-started/overview#cluster",
    );
  });

  it("supports a relative documentation base", () => {
    bootConfigState.docsBaseUrl = "/community-docs";

    expect(getDocsUrl("platform-management/reports")).toBe("/community-docs/platform-management/reports");
  });

  it("avoids duplicate slashes", () => {
    bootConfigState.docsBaseUrl = "https://docs.example.test/product/2.26/";

    expect(getDocsUrl("/reference/cli")).toBe("https://docs.example.test/product/2.26/reference/cli");
  });

  it("replaces the existing product and version path with the configured base", () => {
    const description =
      'Read <a href="https://run-ai-docs.nvidia.com/self-hosted/2.24/platform-management/reports?tab=all#usage">the docs</a>.';

    expect(resolveDocsLinksInHtml(description)).toBe(
      'Read <a href="https://docs.example.test/product/2.26/platform-management/reports?tab=all#usage">the docs</a>.',
    );
  });

  it("resolves existing links without a version path", () => {
    const description =
      "Read <a href='https://run-ai-docs.nvidia.com/self-hosted/platform-management/reports'>the docs</a>.";

    expect(resolveDocsLinksInHtml(description)).toBe(
      "Read <a href='https://docs.example.test/product/2.26/platform-management/reports'>the docs</a>.",
    );
  });

  it("leaves unrelated and malformed links unchanged", () => {
    const description =
      'Read <a href="https://kubernetes.io/docs/">Kubernetes docs</a> or <a href="not a url">this text</a>.';

    expect(resolveDocsLinksInHtml(description)).toBe(description);
  });
});
