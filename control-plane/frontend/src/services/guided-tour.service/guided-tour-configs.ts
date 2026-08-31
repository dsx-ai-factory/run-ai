/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TourConfig } from "@/components/common/guided-tour-popup/guided-tour.types";

// Default configuration for all guided tours
export const DEFAULT_CONFIG: TourConfig = {
  steps: [],
  animate: true,
  overlayStyle: "spotlight",
  overlayColor: "#000",
  overlayOpacity: 0.85,
  overlayClickBehavior: "ignore",
  stagePadding: 10,
  stageRadius: 5,
  disableActiveInteraction: false,
  popoverClass: "",
  popoverOffset: [10, 10],
  showButtons: ["close"],
  showProgress: true,
  progressText: "{{current}}/{{total}}",
  nextBtnText: "Next",
  prevBtnText: "Previous",
  doneBtnText: "Done",
  allowClose: true,
  allowKeyboardControl: false,
  smoothScroll: true,
  scrollBehavior: "center",
  lockScrolling: true,
  debug: false,
};

export const OVERVIEW_TOUR_CONFIG: TourConfig = {
  ...DEFAULT_CONFIG,
  popoverClass: "guided-tour-popover-large",
  showButtons: ["next", "close"],
  disableActiveInteraction: true,
};

export const PROJECT_TOUR_CONFIG: TourConfig = {
  ...DEFAULT_CONFIG,
  popoverClass: "guided-tour-popover-medium",
  showButtons: ["next", "close"],
  disableActiveInteraction: true,
  scrollBehavior: "smart",
};

export const DEPARTMENT_TOUR_CONFIG: TourConfig = {
  ...DEFAULT_CONFIG,
  popoverClass: "guided-tour-popover-medium",
  showButtons: ["next", "close"],
  disableActiveInteraction: true,
  scrollBehavior: "smart",
};
