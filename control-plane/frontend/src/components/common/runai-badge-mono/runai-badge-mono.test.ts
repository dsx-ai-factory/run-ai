/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { QIcon, Quasar } from "quasar";
import { describe, expect, it } from "vitest";
import { RunaiBadgeMono } from "@/components/common/runai-badge-mono";

describe("RunaiBadgeMono component", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiBadgeMono>>;

  function mountComponent(
    props: Partial<{
      aid: string;
      content: string;
      showBadgeGroup: boolean;
      outlinedLabel: string;
      filledLabel: string;
      status: string;
    }> = {},
  ) {
    wrapper = mount(RunaiBadgeMono, {
      global: {
        plugins: [Quasar],
      },
      props: {
        content: "network.topology.nvidia.com/zone",
        ...props,
      },
    });
  }

  it("renders the mono badge content", () => {
    mountComponent();

    expect(wrapper.find('[aid="runai-badge-mono"]').exists()).toBe(true);
    expect(wrapper.find('[aid="runai-badge-mono-content"]').text()).toBe("network.topology.nvidia.com/zone");
    expect(wrapper.find('[aid="runai-badge-mono-status-group"]').exists()).toBe(false);
  });

  it("renders the default status group", () => {
    mountComponent({ showBadgeGroup: true });

    expect(wrapper.find('[aid="runai-badge-mono-outlined-label"]').text()).toBe("Required");
    expect(wrapper.find('[aid="runai-badge-mono-filled-label"]').text()).toBe("Actual");

    const icons = wrapper.findAllComponents(QIcon);
    const successIcon = icons.find((icon) => icon.props("name") === "ph-bold ph-check-circle");

    expect(successIcon).toBeTruthy();
    expect(successIcon?.classes()).toContain("runai-badge-mono-status-icon--success");
  });

  it("renders the warning variation without a filled badge", () => {
    mountComponent({
      showBadgeGroup: true,
      outlinedLabel: "Preferred",
      filledLabel: "",
      status: "warning",
    });

    expect(wrapper.find('[aid="runai-badge-mono-outlined-label"]').text()).toBe("Preferred");
    expect(wrapper.find('[aid="runai-badge-mono-filled-label"]').exists()).toBe(false);

    const icons = wrapper.findAllComponents(QIcon);
    const warningIcon = icons.find((icon) => icon.props("name") === "ph-bold ph-warning-circle");

    expect(warningIcon).toBeTruthy();
    expect(warningIcon?.classes()).toContain("runai-badge-mono-status-icon--warning");
  });
});
