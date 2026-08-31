/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import RunaiCopyableCodeBlock from "./runai-copyable-code-block.vue";
import { describe, expect, test, vi } from "vitest";
import { Quasar, type QNotifyCreateOptions } from "quasar";
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn";
import { alertUtil } from "@/utils/alert.util";

describe("runai-copyable-code-block.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiCopyableCodeBlock>>;

  function mountComponent(props = {}) {
    const defaultProps = {
      code: "console.log('Hello world');",
      codeWithHtml: "",
      useColors: false,
      maxHeight: undefined,
    };

    wrapper = mount(RunaiCopyableCodeBlock, {
      global: {
        plugins: [Quasar],
        components: { RunaiCopyToClipboardBtn },
      },
      props: {
        ...defaultProps,
        ...props,
      },
    });
    wrapper.vm.$q.notify = vi.fn();
  }

  test("component renders correctly", () => {
    mountComponent();
    expect(wrapper.html()).toContain("console.log('Hello world');");
  });

  test("renders safe HTML through sanitization", () => {
    const codeWithHtml = `<span style="color:red;">console.log('Hello, world!');</span>`;
    mountComponent({ codeWithHtml, code: "", useColors: false });
    // DOMPurify keeps safe markup (span + inline style)
    expect(wrapper.html()).toContain(`style="color:red;"`);
    expect(wrapper.html()).toContain("console.log('Hello, world!');");
  });

  test("sanitizes malicious HTML and strips event handlers", () => {
    const codeWithHtml = `<img src=x onerror=alert(1)>`;
    mountComponent({ codeWithHtml, code: "", useColors: false });
    const html = wrapper.html();
    // DOMPurify neutralizes the dangerous onerror handler / script payload
    expect(html).not.toContain("onerror");
    expect(html).not.toContain("alert(1)");
  });

  test("displays code without syntax highlighting", () => {
    mountComponent({ useColors: false });
    expect(wrapper.classes()).not.toContain("code-with-colors");
  });

  test("displays code with syntax highlighting", async () => {
    mountComponent({ useColors: true, maxHeight: 200 });
    await wrapper.vm.$nextTick();
    const coloredElement = wrapper.find(".code-with-colors");
    expect(coloredElement.exists()).toBe(true);
  });

  test("applies max height correctly", () => {
    mountComponent({ maxHeight: 200 });
    const preElement = wrapper.find("pre");
    expect(preElement.attributes("style")).toContain("max-height: 200px;");
  });

  test("copies code to clipboard", async () => {
    const successMessage: QNotifyCreateOptions = {
      message: "Copied!",
      color: "positive",
      icon: "done",
    };

    const getSuccessSpy = vi.spyOn(alertUtil, "getSuccess").mockReturnValue(successMessage);
    mountComponent();
    const execCommandSpy = vi.fn().mockReturnValue(true);
    document.execCommand = execCommandSpy;
    const copyButton = wrapper.findComponent(RunaiCopyToClipboardBtn);
    // Click the underlying button element (the component renders runai-btn,
    // whose multi-root template means the wrapper root isn't the button).
    await copyButton.get("button").trigger("click");
    await wrapper.vm.$nextTick();
    expect(execCommandSpy).toHaveBeenCalledWith("copy");
    expect(getSuccessSpy).toHaveBeenCalledWith("Copied!");
  });
});
