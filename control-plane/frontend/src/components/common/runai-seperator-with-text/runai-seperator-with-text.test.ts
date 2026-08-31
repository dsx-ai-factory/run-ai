/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import RunaiSeperatorWithText from "./runai-seperator-with-text.vue";

const text = "Example Text";

describe("RunaiSeperatorWithText", () => {
  const mountComponent = () => {
    const wrapper = mount(RunaiSeperatorWithText, {
      global: {
        plugins: [Quasar],
      },
      props: {
        text,
      },
    });
    return wrapper;
  };

  it("renders wrapper class", async () => {
    const wrapper = mountComponent();
    expect(wrapper.find(".seperator-with-text").exists()).toBe(true);
  });

  it("check number of props", async () => {
    const wrapper = mountComponent();
    const props = wrapper.vm.$options.props;
    expect(Object.keys(props).length).toBe(1);
  });

  it("renders text correctly", async () => {
    const wrapper = mountComponent();
    expect(wrapper.find(".col-grow").text()).toBe(text);
  });

  it("renders two q-separators", async () => {
    const wrapper = mountComponent();
    expect(wrapper.findAllComponents({ name: "QSeparator" })).toHaveLength(2);
  });
});
