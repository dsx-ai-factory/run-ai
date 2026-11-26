/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface IRunaiPanelsLayoutProps {
  marginTop?: string;
  responsive?: boolean;
  /** Preferred content width; shrinks when the available space is smaller. */
  width?: string;
  leftWidth?: string;
  rightWidth?: string;
  collapsedLeftWidth?: string;
  hideRightOnSmallScreen?: boolean;
  collapseLeftOnSmallScreen?: boolean;
  /** Viewport width in px below which `hideRightOnSmallScreen` hides the right panel. */
  hideRightBreakpoint?: number;
  /** Viewport width in px below which `collapseLeftOnSmallScreen` collapses the left panel. */
  collapseLeftBreakpoint?: number;
}
