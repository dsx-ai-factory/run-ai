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

import { ROOT_URL, getUrlByEnvType, resolveDocsLinksInHtml } from "./docs";

describe("documentation URL resolution", () => {
  beforeEach(() => {
    bootConfigState.docsBaseUrl = "https://docs.example.test/product/2.26";
  });

  it("rebases a legacy topic onto the configured versioned base", () => {
    expect(getUrlByEnvType(`${ROOT_URL}/self-hosted/getting-started/overview#cluster`)).toBe(
      "https://docs.example.test/product/2.26/getting-started/overview#cluster",
    );
  });

  it("replaces an existing legacy version instead of duplicating it", () => {
    expect(getUrlByEnvType(`${ROOT_URL}/self-hosted/2.24/reference/cli`)).toBe(
      "https://docs.example.test/product/2.26/reference/cli",
    );
  });

  it("maps a legacy documentation URL onto a custom documentation base", () => {
    bootConfigState.docsBaseUrl = "/community-docs";

    expect(getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/reports`)).toBe(
      "/community-docs/platform-management/reports",
    );
  });

  it("avoids duplicate slashes when the configured base has a trailing slash", () => {
    bootConfigState.docsBaseUrl = "https://docs.example.test/product/2.26/";

    expect(getUrlByEnvType(`${ROOT_URL}/self-hosted/reference/cli`)).toBe(
      "https://docs.example.test/product/2.26/reference/cli",
    );
  });

  it("leaves unrelated external URLs unchanged", () => {
    expect(getUrlByEnvType("https://kubernetes.io/docs/reference/")).toBe("https://kubernetes.io/docs/reference/");
  });

  it("resolves documentation anchors embedded in HTML", () => {
    const description =
      `Read <a href='${ROOT_URL}/self-hosted/platform-management/reports' target='_blank'>the docs</a>` +
      ` or <a href="https://kubernetes.io/docs/">Kubernetes docs</a>.`;

    expect(resolveDocsLinksInHtml(description)).toBe(
      `Read <a href='https://docs.example.test/product/2.26/platform-management/reports' target='_blank'>the docs</a>` +
        ` or <a href="https://kubernetes.io/docs/">Kubernetes docs</a>.`,
    );
  });
});
