/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { afterEach, describe, expect, it, vi } from "vitest";
import { alertUtil } from "@/utils/alert.util";
import { RunaiCopyToClipboardBtn } from ".";

type TClipboardOutcome = "allowed" | "denied";

describe("runai-copy-to-clipboard-btn.vue", () => {
  const originalClipboard = Object.getOwnPropertyDescriptor(navigator, "clipboard");

  afterEach(() => {
    if (originalClipboard) Object.defineProperty(navigator, "clipboard", originalClipboard);
    else delete (navigator as { clipboard?: Clipboard }).clipboard;
  });

  it.each<{ name: string; clicks: TClipboardOutcome[]; expectedNotifications: string[] }>([
    {
      name: "copies and confirms when clipboard access is allowed",
      clicks: ["allowed"],
      expectedNotifications: ["Copied!"],
    },
    {
      name: "shows an error when clipboard access is denied",
      clicks: ["denied"],
      expectedNotifications: ["Failed to copy to clipboard"],
    },
    {
      name: "stays usable after a denied copy and copies on the next click",
      clicks: ["denied", "allowed"],
      expectedNotifications: ["Failed to copy to clipboard", "Copied!"],
    },
  ])("$name", async ({ clicks, expectedNotifications }) => {
    const writeText = vi.fn();
    clicks.forEach((outcome) =>
      writeText.mockImplementationOnce(() =>
        outcome === "allowed" ? Promise.resolve() : Promise.reject(new Error("Permission denied")),
      ),
    );
    Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true });

    const wrapper = mount(RunaiCopyToClipboardBtn, {
      props: { textToCopy: "kubectl get pods" },
      global: { plugins: [Quasar] },
    });
    const notify = vi.fn();
    wrapper.vm.$q.notify = notify;

    for (let click = 0; click < clicks.length; click++) {
      await wrapper.find("button").trigger("click");
      await flushPromises();
    }

    expect(writeText.mock.calls).toEqual(clicks.map(() => ["kubectl get pods"]));
    expect(notify.mock.calls).toEqual(
      expectedNotifications.map((message) => [
        message === "Copied!" ? alertUtil.getSuccess(message) : alertUtil.getError(message),
      ]),
    );
  });
});
