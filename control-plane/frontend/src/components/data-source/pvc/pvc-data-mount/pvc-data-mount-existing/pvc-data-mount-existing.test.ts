/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { PvcDataMountExisting } from "./index";
import { describe, expect, test } from "vitest";
import { Quasar } from "quasar";

describe("pvc-data-mount-existing.vue", () => {
  test("component loads properly", async () => {
    const props = {
      claimName: "test2" as string,
      pvcList: ["test1", "test2", "test3"] as string[],
      loading: false as boolean,
    };

    const wrapper = mount(PvcDataMountExisting, {
      global: {
        plugins: [Quasar],
      },
      props: props,
    });
    expect(wrapper.html()).toContain(props.claimName);
    const selectCmp = wrapper.find(".q-select");
    const selectOptions = selectCmp.getCurrentComponent()?.props.options;
    expect(selectOptions).toContain(props.pvcList[1]);
  });
});
