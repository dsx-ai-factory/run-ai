/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import RunaiWrapperWithHeader from "./runai-wrapper-with-header.vue";

const header = "Test Header";
const subheader = "Test Subheader";

describe("RunaiWrapperWithHeader", () => {
  const mountComponent = (header?: string, subheader?: string) => {
    const wrapper = mount(RunaiWrapperWithHeader, {
      global: {
        plugins: [Quasar],
      },
      props: {
        header: header,
        subheader: subheader,
      },
    });
    return wrapper;
  };

  it("renders correctly with default props", async () => {
    const wrapper = mountComponent();
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.find(".runai-wrapper-with-header").exists()).toBe(true);
  });

  it("renders correctly with props", async () => {
    const wrapper = mountComponent(header, subheader);
    expect(wrapper.find(".text-subtitle2").text()).toBe(header);
    expect(wrapper.find(".text-body2").text()).toBe(subheader);
  });

  it("renders correctly with empty props", async () => {
    const wrapper = mountComponent();
    expect(wrapper.find(".text-subtitle2").text()).toBe("");
    expect(wrapper.find(".text-body2").text()).toBe("");
  });

  it('emits "close" event when close button is clicked', async () => {
    const wrapper = mountComponent(header, subheader);
    await wrapper.find('[aid="wrapper-x-btn"]').trigger("click");
    expect(wrapper.emitted("close")).toBeTruthy();
  });
});
