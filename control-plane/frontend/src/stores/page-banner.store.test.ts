/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { usePageBannerStore } from "./page-banner.store";

describe("page-banner.store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("has no active banner by default", () => {
    expect(usePageBannerStore().hasActiveBanner).toBe(false);
  });

  it("stays active until every registered banner is gone", () => {
    const store = usePageBannerStore();

    store.register();
    store.register();
    store.unregister();
    expect(store.hasActiveBanner).toBe(true);

    store.unregister();
    expect(store.hasActiveBanner).toBe(false);
  });

  it("never goes below zero", () => {
    const store = usePageBannerStore();

    store.unregister();
    store.register();

    expect(store.activeCount).toBe(1);
  });
});
