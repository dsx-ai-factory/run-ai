/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface ITab {
  name: string;
  label: string;
  disabled?: boolean;
  hidden?: boolean;
  errorIcon?: boolean;
  numberingBadge?: string;
  aid: string;
  tabClass?: string | Record<string, boolean>;
}
