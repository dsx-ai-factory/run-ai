/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";

import NotificationAlertBar from "./notification-alert-bar.vue";
import { RunaiHtmlWrapper } from "@/components/common/runai-html-wrapper";
import { ENotificationBarVariant, variantConfig } from "./notification-alert-bar.model";
import type { INotificationBarCta } from "./notification-alert-bar.model";

describe("notification-alert-bar.vue", () => {
  const getMountedComponent = (props = {}, slots = {}): VueWrapper => {
    return shallowMount(NotificationAlertBar, {
      global: {
        plugins: [Quasar],
      },
      props,
      slots,
    });
  };

  describe("variant rendering", () => {
    it.each(Object.values(ENotificationBarVariant))("renders %s variant with correct background class", (variant) => {
      const wrapper = getMountedComponent({ variant });
      const config = variantConfig[variant as ENotificationBarVariant];
      expect(wrapper.find(".notification-alert-bar").classes()).toContain(config.background);
    });

    it.each(Object.values(ENotificationBarVariant))("renders %s variant with correct icon class", (variant) => {
      const wrapper = getMountedComponent({ variant });
      const config = variantConfig[variant as ENotificationBarVariant];
      const variantIcon = wrapper.find("q-icon-stub.variant-icon");
      expect(variantIcon.exists()).toBe(true);
      expect(variantIcon.classes()).toContain(config.iconColor);
    });

    it("defaults to info variant when no variant is provided", () => {
      const wrapper = getMountedComponent();
      const infoConfig = variantConfig[ENotificationBarVariant.Info];
      expect(wrapper.find(".notification-alert-bar").classes()).toContain(infoConfig.background);
      expect(wrapper.find("q-icon-stub.variant-icon").classes()).toContain(infoConfig.iconColor);
    });
  });

  describe("message prop", () => {
    it("renders html wrapper with message prop", () => {
      const wrapper = getMountedComponent({ message: "Test message" });
      const htmlWrapper = wrapper.findComponent(RunaiHtmlWrapper);
      expect(htmlWrapper.exists()).toBe(true);
      expect(htmlWrapper.props("html")).toBe("Test message");
      expect(htmlWrapper.props("tag")).toBe("span");
    });

    it("does not render html wrapper when message is not provided and no slot", () => {
      const wrapper = getMountedComponent();
      const htmlWrapper = wrapper.findComponent(RunaiHtmlWrapper);
      expect(htmlWrapper.exists()).toBe(false);
    });
  });

  describe("default slot", () => {
    it("renders default slot content instead of message prop", () => {
      const wrapper = getMountedComponent({}, { default: "<span>Slot content here</span>" });
      expect(wrapper.find(".content").text()).toContain("Slot content here");
    });

    it("prefers slot over message prop", () => {
      const wrapper = getMountedComponent({ message: "Prop message" }, { default: "<span>Slot wins</span>" });
      expect(wrapper.find(".content").text()).toContain("Slot wins");
      expect(wrapper.find(".content").text()).not.toContain("Prop message");
    });
  });

  describe("title and showH1Content", () => {
    it("does not render title when showH1Content is false", () => {
      const wrapper = getMountedComponent({ title: "My title", showH1Content: false });
      expect(wrapper.find(".notification-alert-bar__title").exists()).toBe(false);
    });

    it("does not render title when title prop is not provided", () => {
      const wrapper = getMountedComponent({ showH1Content: true });
      expect(wrapper.find(".notification-alert-bar__title").exists()).toBe(false);
    });

    it("renders title when both showH1Content and title are set", () => {
      const wrapper = getMountedComponent({ showH1Content: true, title: "Important title" });
      const titleEl = wrapper.find(".notification-alert-bar__title");
      expect(titleEl.exists()).toBe(true);
      expect(titleEl.text()).toBe("Important title");
      expect(titleEl.classes()).toContain("text-b-body1-md");
    });
  });

  describe("orientation", () => {
    it("defaults to horizontal orientation", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.find(".notification-alert-bar").classes()).toContain("notification-alert-bar--horizontal");
    });

    it("applies vertical orientation class", () => {
      const wrapper = getMountedComponent({ orientation: "vertical" });
      expect(wrapper.find(".notification-alert-bar").classes()).toContain("notification-alert-bar--vertical");
    });

    it("applies horizontal orientation class", () => {
      const wrapper = getMountedComponent({ orientation: "horizontal" });
      expect(wrapper.find(".notification-alert-bar").classes()).toContain("notification-alert-bar--horizontal");
    });
  });

  describe("close action", () => {
    const closeSelector = "q-icon-stub.notification-alert-bar__close";

    it("does not render close button by default", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.find(closeSelector).exists()).toBe(false);
    });

    it("renders close button when showCloseAction is true", () => {
      const wrapper = getMountedComponent({ showCloseAction: true });
      expect(wrapper.find(closeSelector).exists()).toBe(true);
    });

    it("emits close event when close button is clicked", async () => {
      const wrapper = getMountedComponent({ showCloseAction: true });
      await wrapper.find(closeSelector).trigger("click");
      expect(wrapper.emitted("close")).toHaveLength(1);
    });

    it("has correct accessibility attributes on close button", () => {
      const wrapper = getMountedComponent({ showCloseAction: true });
      const closeBtn = wrapper.find(closeSelector);
      expect(closeBtn.attributes("role")).toBe("button");
      expect(closeBtn.attributes("tabindex")).toBe("0");
    });
  });

  describe("inline-actions slot", () => {
    it("renders inline-actions slot content inside content area", () => {
      const wrapper = getMountedComponent(
        {},
        {
          default: "<span>Follow the</span>",
          "inline-actions": '<a href="#">link text</a>',
        },
      );
      const content = wrapper.find(".notification-alert-bar__content");
      expect(content.text()).toContain("Follow the");
      expect(content.text()).toContain("link text");
    });
  });

  describe("CTAs (warning only)", () => {
    const createCtas = (): INotificationBarCta[] => [
      { label: "Action 1", onClick: vi.fn() },
      { label: "Action 2", onClick: vi.fn() },
    ];

    it("does not render CTAs when ctAs is not provided", () => {
      const wrapper = getMountedComponent({ variant: ENotificationBarVariant.Warning });
      expect(wrapper.find(".notification-alert-bar__ctas").exists()).toBe(false);
    });

    it("renders CTAs for warning variant", () => {
      const wrapper = getMountedComponent({
        variant: ENotificationBarVariant.Warning,
        ctAs: createCtas(),
      });
      expect(wrapper.find(".notification-alert-bar__ctas").exists()).toBe(true);
      const buttons = wrapper.findAll("q-btn-stub.notification-alert-bar__cta-btn");
      expect(buttons).toHaveLength(2);
    });

    it("does not render CTAs for non-warning variants", () => {
      const nonWarningVariants = [
        ENotificationBarVariant.Info,
        ENotificationBarVariant.Error,
        ENotificationBarVariant.Success,
        ENotificationBarVariant.Default,
      ];

      nonWarningVariants.forEach((variant) => {
        const wrapper = getMountedComponent({ variant, ctAs: createCtas() });
        expect(wrapper.find(".notification-alert-bar__ctas").exists()).toBe(false);
      });
    });

    it("renders CTA button with correct label", () => {
      const ctas: INotificationBarCta[] = [{ label: "Click me", onClick: vi.fn() }];
      const wrapper = getMountedComponent({
        variant: ENotificationBarVariant.Warning,
        ctAs: ctas,
      });
      const btn = wrapper.find("q-btn-stub.notification-alert-bar__cta-btn");
      expect(btn.attributes("label")).toBe("Click me");
    });

    it("does not render CTAs when array is empty", () => {
      const wrapper = getMountedComponent({
        variant: ENotificationBarVariant.Warning,
        ctAs: [],
      });
      expect(wrapper.find(".notification-alert-bar__ctas").exists()).toBe(false);
    });
  });

  describe("typography classes", () => {
    it("applies text-b-body1 class to root element", () => {
      const wrapper = getMountedComponent();
      expect(wrapper.find(".notification-alert-bar").classes()).toContain("text-b-body1");
    });

    it("applies text-b-body1-md class to title", () => {
      const wrapper = getMountedComponent({ showH1Content: true, title: "Title" });
      expect(wrapper.find(".notification-alert-bar__title").classes()).toContain("text-b-body1-md");
    });
  });
});
