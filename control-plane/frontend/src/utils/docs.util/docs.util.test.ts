/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, vi } from "vitest";

const docsState = vi.hoisted(() => ({
  supportUrl: "",
  supportLabel: "Example support",
  numaAwareSchedulingUrl: "https://docs.example.test/numa-aware-scheduling",
  prerequisitesUrl: "https://docs.example.test/prerequisites-and-features",
}));

vi.mock("@/constants/docs", () => ({
  default: {
    SUPPORT: {
      get URL() {
        return docsState.supportUrl;
      },
      get LABEL() {
        return docsState.supportLabel;
      },
    },
    NODE_POOLS: {
      get NUMA_AWARE_SCHEDULING() {
        return docsState.numaAwareSchedulingUrl;
      },
    },
    INSTALLATION: {
      get CLUSTER_PREREQUISITES() {
        return docsState.prerequisitesUrl;
      },
    },
  },
}));

import { docsUtil } from "./docs.util";

describe("support links", () => {
  afterEach(() => {
    docsState.supportUrl = "";
    docsState.supportLabel = "Example support";
  });

  it("returns no support markup when the support URL is missing", () => {
    expect(docsUtil.getSupportAnchor()).toBe("");
    expect(docsUtil.getSupportHTMLLink()).toBe("");
  });

  it("returns clickable support markup when the support URL is configured", () => {
    docsState.supportUrl = "https://support.example.test/contact";

    expect(docsUtil.getSupportAnchor("contact support")).toBe(
      '<a href="https://support.example.test/contact" target="_blank" rel="noopener noreferrer">contact support</a>',
    );
    expect(docsUtil.getSupportHTMLLink()).toBe(
      'Contact <a href="https://support.example.test/contact" target="_blank" rel="noopener noreferrer">Example support.</a>',
    );
  });

  it("escapes the support URL, configured label, and custom anchor text", () => {
    docsState.supportUrl = 'https://support.example.test/contact" onmouseover="alert(1)';
    docsState.supportLabel = '<img src=x onerror="alert(1)">';

    expect(docsUtil.getSupportAnchor("<strong>contact support</strong>")).toBe(
      '<a href="https://support.example.test/contact&quot; onmouseover=&quot;alert(1)" target="_blank" rel="noopener noreferrer">&lt;strong&gt;contact support&lt;/strong&gt;</a>',
    );
    expect(docsUtil.getSupportHTMLLink()).toBe(
      'Contact <a href="https://support.example.test/contact&quot; onmouseover=&quot;alert(1)" target="_blank" rel="noopener noreferrer">&lt;img src=x onerror=&quot;alert(1)&quot;&gt;.</a>',
    );
  });
});

describe("getNumaAwareSchedulingHTMLLink", () => {
  it("returns clickable markup linking to the NUMA-aware scheduling guide", () => {
    expect(docsUtil.getNumaAwareSchedulingHTMLLink()).toBe(
      'For more information see the <a href="https://docs.example.test/numa-aware-scheduling" target="_blank" rel="noopener noreferrer">Configuring NUMA-Aware Scheduling guide.</a>',
    );
  });

  it("escapes the configured URL", () => {
    docsState.numaAwareSchedulingUrl = 'https://docs.example.test/numa" onmouseover="alert(1)';

    expect(docsUtil.getNumaAwareSchedulingHTMLLink()).toBe(
      'For more information see the <a href="https://docs.example.test/numa&quot; onmouseover=&quot;alert(1)" target="_blank" rel="noopener noreferrer">Configuring NUMA-Aware Scheduling guide.</a>',
    );

    docsState.numaAwareSchedulingUrl = "https://docs.example.test/numa-aware-scheduling";
  });
});

describe("prerequisites link", () => {
  afterEach(() => {
    docsState.prerequisitesUrl = "https://docs.example.test/prerequisites-and-features";
  });

  it("links to the prerequisites guide when the catalog has a destination", () => {
    expect(docsUtil.getPrerequisitesHTMLLink()).toBe(
      'For more information, see the <a href="https://docs.example.test/prerequisites-and-features" ' +
        'target="_blank" rel="noopener noreferrer">Prerequisites guide.</a>',
    );
  });

  it("returns no markup when the catalog has no destination for the page", () => {
    docsState.prerequisitesUrl = "";

    expect(docsUtil.getPrerequisitesHTMLLink()).toBe("");
  });
});
