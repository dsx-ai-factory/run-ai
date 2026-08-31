/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { guidedTourService } from "./guided-tour.service";
import type { TourConfig } from "@/components/common/guided-tour-popup/guided-tour.types";

// Mock DOM elements and methods
const createMockElement = (rect: Partial<DOMRect>) => {
  const element = document.createElement("div");
  element.getBoundingClientRect = vi.fn(
    () =>
      ({
        top: 0,
        left: 0,
        right: 100,
        bottom: 100,
        width: 100,
        height: 100,
        x: 0,
        y: 0,
        ...rect,
      }) as DOMRect,
  );
  element.scrollIntoView = vi.fn();
  return element;
};

// Mock window properties
const mockWindow = {
  innerHeight: 800,
  innerWidth: 1200,
  pageYOffset: 0,
  pageXOffset: 0,
  scrollTo: vi.fn(),
};

// Mock document properties
const mockDocument = {
  documentElement: {
    scrollTop: 0,
    scrollLeft: 0,
  },
};

describe("Guided Tour Service - Scroll Behavior", () => {
  beforeEach(() => {
    // Setup mocks
    Object.defineProperty(window, "innerHeight", { value: mockWindow.innerHeight, writable: true });
    Object.defineProperty(window, "innerWidth", { value: mockWindow.innerWidth, writable: true });
    Object.defineProperty(window, "pageYOffset", { value: mockWindow.pageYOffset, writable: true });
    Object.defineProperty(window, "pageXOffset", { value: mockWindow.pageXOffset, writable: true });
    Object.defineProperty(window, "scrollTo", { value: mockWindow.scrollTo, writable: true });
    Object.defineProperty(document, "documentElement", { value: mockDocument.documentElement, writable: true });

    // Reset mocks
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("bringIntoView", () => {
    it("should center element by default when scrollBehavior is not specified", () => {
      const element = createMockElement({ top: -200, bottom: -100 });
      const config: TourConfig = { smoothScroll: false };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "center",
        block: "center",
      });
    });

    it('should center element when scrollBehavior is "center"', () => {
      const element = createMockElement({ top: -200, bottom: -100 });
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "center" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "center",
        block: "center",
      });
    });

    it('should align element to start when scrollBehavior is "start"', () => {
      const element = createMockElement({ top: -200, bottom: -100 });
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "start" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "nearest",
        block: "start",
      });
    });

    it('should align element to end when scrollBehavior is "end"', () => {
      const element = createMockElement({ top: -200, bottom: -100 });
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "end" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "nearest",
        block: "end",
      });
    });

    it('should use nearest alignment when scrollBehavior is "nearest"', () => {
      const element = createMockElement({ top: -200, bottom: -100 });
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "nearest" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "nearest",
        block: "nearest",
      });
    });

    it('should use smart behavior when scrollBehavior is "smart"', () => {
      const element = createMockElement({ top: -200, bottom: -100, height: 900 }); // Taller than viewport
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "smart" };

      guidedTourService.bringIntoView(element, config);

      // Should use "start" for tall elements
      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "center",
        block: "start",
      });
    });

    it("should use smooth scrolling when smoothScroll is enabled", () => {
      const element = createMockElement({ top: -200, bottom: -100 });
      const config: TourConfig = { smoothScroll: true, scrollBehavior: "center" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "smooth",
        inline: "center",
        block: "center",
      });
    });

    it("should not scroll when element is already in view", () => {
      const element = createMockElement({ top: 100, bottom: 200 }); // Element is in view
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "center" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).not.toHaveBeenCalled();
    });

    it("should handle horizontal scrolling when element is off-screen horizontally", () => {
      const element = createMockElement({ left: -200, right: -100, top: 100, bottom: 200 });
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "center" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "center",
        block: "center",
      });
    });

    it("should handle both vertical and horizontal scrolling", () => {
      const element = createMockElement({
        top: -200,
        bottom: -100,
        left: -200,
        right: -100,
      });
      const config: TourConfig = { smoothScroll: false, scrollBehavior: "center" };

      guidedTourService.bringIntoView(element, config);

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: "auto",
        inline: "center",
        block: "center",
      });
    });
  });
});
