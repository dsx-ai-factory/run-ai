/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { RunaiStatus } from "@/components/common/runai-status";
import { Quasar } from "quasar";
import { describe, test, expect } from "vitest";
import type { IStatusColOptions } from "@/models/table.model";
import { dateUtil } from "@/utils/date.util";
import { Phase } from "@/swagger-models/workloads-client";

describe("RunaiStatus.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiStatus>>;
  const createComponent = (options: IStatusColOptions) => {
    wrapper = mount(RunaiStatus, {
      global: {
        plugins: [Quasar],
      },
      props: {
        options,
      },
    });
  };
  test("renders status text", () => {
    const options: IStatusColOptions = {
      status: Phase.Running,
      displayAnimation: false,
    };
    createComponent(options);

    expect(wrapper.text()).toBe("Running");
  });

  test("renders animation when displayAnimation is true", async () => {
    const options = {
      status: Phase.Creating,
      displayAnimation: true,
    };
    createComponent(options);

    expect(wrapper.find(".q-spinner").exists()).toBe(true);
    expect(wrapper.text()).toBe("Creating");
  });

  test("applies correct status color class", () => {
    const options = {
      status: Phase.Failed,
      displayAnimation: false,
      color: "negative",
    };

    createComponent(options);
    expect(wrapper.find(".negative-color").exists()).toBe(true);
  });
  test("displays time correctly", () => {
    const options = {
      status: Phase.Pending,
      displayAnimation: false,
      statusUpdatedTimeInMs: 1234567890,
    };

    const wrapper = mount(RunaiStatus, {
      props: {
        options,
      },
    });
    const statusUpdateTimeText = `${Phase.Pending} (${dateUtil.timeAgo(1234567890)})`;

    expect(wrapper.text()).toBe(statusUpdateTimeText);
  });
});
