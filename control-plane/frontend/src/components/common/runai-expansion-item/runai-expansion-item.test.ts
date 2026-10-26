/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiExpansionItem } from "./index";

describe("runai-expansion-item", () => {
  it("should not display the content when there is not default slot", async () => {
    const slotContent = "the content inside the card";
    const subheader = "<div>this is the subheader</div>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        subheader,
      },
      props: {
        label: "Default",
      },
    });

    expect(cmp.html()).not.toContain(slotContent);
    const expIconEl = cmp.find(".runai-expansion-icon");
    expect(expIconEl.classes("display-none")).toBe(true);
    const headerEl = cmp.find(".runai-expansion-header");
    await headerEl.trigger("click");
    expect(cmp.html()).not.toContain(slotContent);
  });

  it("should display subheader when it is closed", async () => {
    const slotContent = "the content inside the card";
    const subheader = "<div>this is the subheader</div>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: slotContent,
        subheader,
      },
      props: {
        label: "Default",
      },
    });

    const headerEl = cmp.find(".runai-expansion-header");
    expect(headerEl.html()).toContain(subheader);
  });

  it("should display the sub header when expansion is close", async () => {
    const slotContent = "the content inside the card";
    const subheader = "<div>this is the subheader</div>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: slotContent,
        subheader,
      },
      props: {
        label: "Default",
        showSubHeader: true,
      },
    });

    const headerEl = cmp.find(".runai-expansion-header");
    expect(headerEl.html()).toContain(subheader);
    await headerEl.trigger("click");
    expect(cmp.html()).toContain(slotContent);
    expect(headerEl.html()).toContain(subheader);
  });

  it("should not have subheader when not using the subheader slot", async () => {
    const slotContent = "the content inside the card";
    const subheader = "<div>this is the subheader</div>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: slotContent,
      },
      props: {
        label: "Default",
        defaultOpened: true,
      },
    });

    const headerEl = cmp.find(".runai-expansion-header");
    expect(headerEl.html()).not.toContain(subheader);
  });

  it("should display content when default-opened is true", async () => {
    const slotContent = "the content inside the card";
    const subheader = "<div>this is the subheader</div>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: slotContent,
        subheader,
      },
      props: {
        label: "Default",
        defaultOpened: true,
      },
    });

    const contentEl = cmp.find(".runai-expansion-content");
    expect(contentEl.html()).toContain(slotContent);

    const headerEl = cmp.find(".runai-expansion-header");
    expect(headerEl.html()).not.toContain(subheader);
  });

  it("should display hidding the content when disable-closing is true", async () => {
    const slotContent = "the content inside the card";
    const subheader = "<div>this is the subheader</div>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: slotContent,
        subheader,
      },
      props: {
        label: "Default",
        disableClosing: true,
      },
    });

    const contentEl = cmp.find(".runai-expansion-content");
    expect(contentEl.html()).toContain(slotContent);
    const expIconEl = cmp.find(".runai-expansion-icon");
    expect(expIconEl.classes("display-none")).toBe(true);

    const headerEl = cmp.find(".runai-expansion-header");
    await headerEl.trigger("click");
    expect(contentEl.html()).toContain(slotContent);
  });

  it("should display content when disable-closing is true", async () => {
    const slotContent = "the content inside the card";
    const subheader = "<div>this is the subheader</div>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: slotContent,
        subheader,
      },
      props: {
        label: "Default",
        hideExpendIcon: true,
      },
    });

    const expIconEl = cmp.find(".runai-expansion-icon");
    expect(expIconEl.classes("display-none")).toBe(true);
  });

  it("should not render the footer when the footer slot is not provided", async () => {
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: "the content inside the card",
      },
      props: {
        label: "Default",
        defaultOpened: true,
      },
    });

    expect(cmp.find(".runai-expansion-footer").exists()).toBe(false);
  });

  it("should render the footer slot content when provided", async () => {
    const footerContent = "<button>save changes</button>";
    const cmp = mount(RunaiExpansionItem, {
      global: {
        plugins: [Quasar],
      },
      slots: {
        default: "the content inside the card",
        footer: footerContent,
      },
      props: {
        label: "Default",
        defaultOpened: true,
      },
    });

    const footerEl = cmp.find(".runai-expansion-footer");
    expect(footerEl.exists()).toBe(true);
    expect(footerEl.html()).toContain(footerContent);
  });

  describe("status icon while open", () => {
    const mountOpen = (props: Record<string, unknown>) =>
      mount(RunaiExpansionItem, {
        global: { plugins: [Quasar] },
        slots: { default: "<div>content</div>" },
        props: { label: "Default", defaultOpened: true, ...props },
      });

    it("should not show the invalid icon in an open section by default", () => {
      const cmp = mountOpen({ sectionInvalid: true });
      expect(cmp.find(".ph-warning-circle").exists()).toBe(false);
    });

    it("should show the invalid icon in an open section when showStatusIconWhenOpen is true", () => {
      const cmp = mountOpen({ sectionInvalid: true, showStatusIconWhenOpen: true });
      expect(cmp.find(".ph-warning-circle").exists()).toBe(true);
    });

    it("should show the warning icon in an open section when showStatusIconWhenOpen is true", () => {
      const cmp = mountOpen({ sectionWarning: true, showStatusIconWhenOpen: true });
      expect(cmp.find(".ph-warning-circle").exists()).toBe(true);
    });
  });
});
