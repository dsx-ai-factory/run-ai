/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, vi, it } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiStageBadge } from "@/components/common/runai-stage-badge/index";
import type { TestingPinia } from "@pinia/testing";
import { createTestingPinia } from "@pinia/testing";
import { EBadgeLabel } from "@/models/setting.model";

let wrapper: VueWrapper<InstanceType<typeof RunaiStageBadge>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });

describe("RunaiStageBadge component", () => {
  const mountComponent = (props: {
    badge: EBadgeLabel;
    displayTooltip: boolean;
    color?: string;
    useDefaultCursor?: boolean;
  }) => {
    wrapper = mount(RunaiStageBadge, {
      global: {
        plugins: [Quasar, pinia],
      },
      props,
    });
  };

  it("should render RunaiStageBadge component", () => {
    [EBadgeLabel.Legacy, EBadgeLabel.Beta, EBadgeLabel.Experimental].forEach((badge) => {
      mountComponent({ badge, displayTooltip: true });
      expect(wrapper.vm).toBeTruthy();
    });
  });

  it("should override color", () => {
    mountComponent({ badge: EBadgeLabel.Legacy, displayTooltip: true, color: "secondary" });
    expect(wrapper.vm.badgeStyle).eql({ padding: "1px 24px", color: "secondary", width: "90px" });
  });

  it("should apply 'badge-default-cursor' class when useDefaultCursor is true", () => {
    mountComponent({ badge: EBadgeLabel.Legacy, displayTooltip: false, useDefaultCursor: true });
    const badge = wrapper.find('[aid="badge-Legacy"]');
    expect(badge.classes()).toContain("badge-default-cursor");

    mountComponent({ badge: EBadgeLabel.Legacy, displayTooltip: false, useDefaultCursor: false });
    const badgeNoCursor = wrapper.find('[aid="badge-Legacy"]');
    expect(badgeNoCursor.classes()).not.toContain("badge-default-cursor");
  });
});
