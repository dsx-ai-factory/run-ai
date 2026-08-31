/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, test, expect, vi, beforeEach } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import OidcIndex from "@/views/oidc/oidc-index.vue";
import { RunaiCopyableCodeBlock } from "@/components/common/runai-copyable-code-block";

describe("oidc-index.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof OidcIndex>>;

  function mountComponent({
    hash = "",
    query = {} as Record<string, string | string[]>,
  }: { hash?: string; query?: Record<string, string | string[]> } = {}) {
    wrapper = mount(OidcIndex, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        stubs: {
          RunaiBrandLogo: true,
          RunaiCopyableCodeBlock: true,
          // Stub RunaiSystemPage with a slot-passing wrapper so the component's
          // own implementation (which reads from localStorage) does not run,
          // while still rendering its default slot for template assertions.
          RunaiSystemPage: { template: "<div><slot /></div>" },
        },
        mocks: {
          $route: { hash, query },
        },
      },
    });
  }

  beforeEach(() => {
    vi.clearAllMocks();
  });

  // --- fragment delivery (browser login flow: response_mode=fragment) ---

  test("reads code from hash fragment (browser login flow)", () => {
    mountComponent({ hash: "#code=10982&state=test-state" });

    expect(wrapper.find(".title").text()).toBe("Login Successful!");
    const codeBlock = wrapper.findComponent(RunaiCopyableCodeBlock);
    expect(codeBlock.exists()).toBe(true);
    expect(codeBlock.props("code")).toBe("10982");
  });

  // --- query string delivery (CLI remote-browser flow: response_mode=query) ---

  test("reads code from query string when hash is absent (CLI remote-browser flow)", () => {
    mountComponent({ query: { code: "cli-auth-code" } });

    expect(wrapper.find(".title").text()).toBe("Login Successful!");
    const codeBlock = wrapper.findComponent(RunaiCopyableCodeBlock);
    expect(codeBlock.exists()).toBe(true);
    expect(codeBlock.props("code")).toBe("cli-auth-code");
  });

  test("hash takes priority over query string when both are present", () => {
    mountComponent({ hash: "#code=hash-code", query: { code: "query-code" } });

    const codeBlock = wrapper.findComponent(RunaiCopyableCodeBlock);
    expect(codeBlock.props("code")).toBe("hash-code");
  });

  // --- shared validation (both delivery paths) ---

  test("shows the fallback message and no code block when no auth code is present", () => {
    mountComponent();

    expect(wrapper.find(".title").exists()).toBe(false);
    expect(wrapper.find(".subtitle").text()).toBe(
      "No OIDC auth code detected, either you reached this page by mistake or you should redo the auth flow.",
    );
    expect(wrapper.findComponent(RunaiCopyableCodeBlock).exists()).toBe(false);
  });

  test("does not pass the raw code through an HTML-rendering path", () => {
    // Spaceless payload — passes VSCHAR (%x21-7E) so it reaches the component,
    // where it must be bound via v-text (plain text), never codeWithHtml (v-html path).
    mountComponent({ hash: "#code=<script>alert(1)</script>&state=s" });

    const codeBlock = wrapper.findComponent(RunaiCopyableCodeBlock);
    expect(codeBlock.props("code")).toBe("<script>alert(1)</script>");
    expect(codeBlock.props("codeWithHtml")).toBeUndefined();
  });

  test("rejects code containing spaces (below VSCHAR %x21 lower bound)", () => {
    mountComponent({ hash: "#code=<img src=x onerror=alert(1)>&state=s" });

    expect(wrapper.find(".title").exists()).toBe(false);
    expect(wrapper.findComponent(RunaiCopyableCodeBlock).exists()).toBe(false);
  });

  test("rejects code containing non-VSCHAR bytes (RFC 6749 §4.1.2 VSCHAR = %x21-7E)", () => {
    mountComponent({ hash: "#code=valid-prefix\x00injection&state=s" });

    expect(wrapper.find(".title").exists()).toBe(false);
    expect(wrapper.findComponent(RunaiCopyableCodeBlock).exists()).toBe(false);
  });

  test("treats duplicate code params in query string as missing (CLI path) — prevents toString() array coercion", () => {
    // Vue Router produces string[] for duplicate query keys. We must not call
    // .toString() on the array; requiring typeof === "string" rejects it outright.
    mountComponent({ query: { code: ["<img src=x onerror=alert(1)>", "REAL-AUTH-CODE"] } });

    expect(wrapper.find(".title").exists()).toBe(false);
    expect(wrapper.find(".subtitle").text()).toBe(
      "No OIDC auth code detected, either you reached this page by mistake or you should redo the auth flow.",
    );
  });
});
