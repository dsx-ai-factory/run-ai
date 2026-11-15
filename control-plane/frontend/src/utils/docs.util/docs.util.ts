/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import DOCS from "@/constants/docs";

export const docsUtil = {
  getSupportHTMLLink,
  getSupportAnchor,
  getTroubleshootingHTMLLink,
  getNumaAwareSchedulingHTMLLink,
  getPrerequisitesHTMLLink,
};

function getSupportAnchor(text = DOCS.SUPPORT.LABEL): string {
  if (!DOCS.SUPPORT.URL) return "";
  return `<a href="${escapeHTML(DOCS.SUPPORT.URL)}" target="_blank" rel="noopener noreferrer">${escapeHTML(text)}</a>`;
}

function getSupportHTMLLink(): string {
  if (!DOCS.SUPPORT.URL) return "";
  const supportURL = escapeHTML(DOCS.SUPPORT.URL);
  return `Contact <a href="${supportURL}" target="_blank" rel="noopener noreferrer">${escapeHTML(DOCS.SUPPORT.LABEL)}.</a>`;
}

function getTroubleshootingHTMLLink(href: string): string {
  return `See the <a href="${href}" target="_blank" rel="noopener noreferrer">Troubleshooting guide.</a>`;
}

function getNumaAwareSchedulingHTMLLink(): string {
  const numaGuideURL = escapeHTML(DOCS.NODE_POOLS.NUMA_AWARE_SCHEDULING);
  return `For more information see the <a href="${numaGuideURL}" target="_blank" rel="noopener noreferrer">Configuring NUMA-Aware Scheduling guide.</a>`;
}

function getPrerequisitesHTMLLink(): string {
  // The docs catalog in use may have no destination for this page, in which case it is rendered
  // without a link rather than pointing nowhere.
  if (!DOCS.INSTALLATION.CLUSTER_PREREQUISITES) return "";
  const prerequisitesURL = escapeHTML(DOCS.INSTALLATION.CLUSTER_PREREQUISITES);
  return `For more information, see the <a href="${prerequisitesURL}" target="_blank" rel="noopener noreferrer">Prerequisites guide.</a>`;
}

function escapeHTML(value: string): string {
  return value.replace(/[&<>"']/g, (character) => {
    switch (character) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      default:
        return "&#39;";
    }
  });
}
