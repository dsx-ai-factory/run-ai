/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface OnboardingSubStep {
  id: number;
  title: string;
  isClickable?: boolean;
}

export interface OnboardingSidebarStep {
  id: number;
  title: string;
  subSteps?: OnboardingSubStep[];
  isClickable?: boolean;
}

export interface OnboardingSidebarConfig {
  steps: OnboardingSidebarStep[];
}
