/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

// Counts the page banners currently on screen, so lower-priority banners (e.g. the SAML signature one) can step aside.
export const usePageBannerStore = defineStore("page-banner", {
  state: () => ({
    activeCount: 0,
  }),
  getters: {
    hasActiveBanner(): boolean {
      return this.activeCount > 0;
    },
  },
  actions: {
    register(): void {
      this.activeCount++;
    },
    unregister(): void {
      this.activeCount = Math.max(0, this.activeCount - 1);
    },
  },
});
