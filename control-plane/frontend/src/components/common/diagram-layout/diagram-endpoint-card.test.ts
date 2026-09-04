/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";

import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import DiagramEndpointCard from "./diagram-endpoint-card.vue";

const mountCard = (props: { url: string; variant?: "workload" | "application"; highlightText?: string }) =>
  mount(DiagramEndpointCard, {
    global: {
      plugins: [Quasar],
      stubs: { Handle: true },
    },
    props,
  });

describe("diagram-endpoint-card.vue", () => {
  it("should render the endpoint URL", () => {
    const url = "https://inference-api.example.com/v1/completions";
    const wrapper = mountCard({ url });

    expect(wrapper.find(".endpoint-url").text()).toBe(url);
  });

  it("should expose the full URL via the title attribute for truncation tooltips", () => {
    const url = "https://very-long-host.example.com/v1/some/long/path";
    const wrapper = mountCard({ url });

    expect(wrapper.find(".endpoint-url").attributes("title")).toBe(url);
  });

  it("should render the serving-endpoint icon inside the colored badge", () => {
    const wrapper = mountCard({ url: "https://a.example.com" });
    const iconBox = wrapper.find(".endpoint-icon-box");
    expect(iconBox.exists()).toBe(true);
    const svgIcon = wrapper.findComponent(RunaiSvgIcon);
    expect(svgIcon.exists()).toBe(true);
    expect(svgIcon.props("name")).toBe("serving-endpoint");
  });

  it("should render the copy button wired with the URL as textToCopy", () => {
    const url = "https://a.example.com/v1";
    const wrapper = mountCard({ url });

    const copyBtn = wrapper.findComponent({ name: "runai-copy-to-clipboard-btn" });
    expect(copyBtn.exists()).toBe(true);
    expect(copyBtn.props("textToCopy")).toBe(url);
    expect(copyBtn.props("customTooltip")).toBe("Copy");
  });

  it("should align with application resource card styling", () => {
    const wrapper = mountCard({ url: "https://a.example.com", variant: "application" });

    expect(wrapper.find(".endpoint-card").classes()).toContain("application-resource-card");
  });

  it("should highlight matching text in the endpoint URL", () => {
    const wrapper = mountCard({ url: "https://inference.example.com", highlightText: "Inference" });

    expect(wrapper.find("mark.search-highlight").text()).toBe("inference");
    expect(wrapper.find(".endpoint-url").text()).toBe("https://inference.example.com");
  });
});
