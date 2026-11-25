/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { afterEach, describe, expect, it } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { defineComponent } from "vue";
import RunaiBaseDialog from "./runai-base-dialog.vue";

let wrapper: VueWrapper;
afterEach(() => wrapper?.unmount());

async function render(attrs: Record<string, string> = {}) {
  wrapper = mount(RunaiBaseDialog, {
    attachTo: document.body,
    props: { modelValue: true },
    attrs,
    slots: { header: "New Network Topology", body: "Topology fields" },
    global: { plugins: [Quasar] },
  });
  await flushPromises();
  return document.body.querySelector('[role="dialog"]') as HTMLElement;
}

describe("base dialog accessible naming", () => {
  it("names the dialog using its visible heading", async () => {
    const dialog = await render();
    const headingId = dialog.getAttribute("aria-labelledby");
    expect(headingId).toBeTruthy();
    expect(document.getElementById(headingId as string)?.textContent).toContain("New Network Topology");
    expect(dialog.querySelector("h6")?.id).toBe(headingId);
  });

  it("gives each dialog a unique heading within the same app", async () => {
    wrapper = mount(
      defineComponent({
        components: { RunaiBaseDialog },
        template: `<div>
          <runai-base-dialog :model-value="true"><template #header>First dialog</template></runai-base-dialog>
          <runai-base-dialog :model-value="true"><template #header>Second dialog</template></runai-base-dialog>
        </div>`,
      }),
      { attachTo: document.body, global: { plugins: [Quasar] } },
    );
    await flushPromises();
    const headings = [...document.querySelectorAll<HTMLElement>('[role="dialog"] h6')];
    expect(headings).toHaveLength(2);
    expect(headings[0].id).toBeTruthy();
    expect(headings[1].id).toBeTruthy();
    expect(headings[0].id).not.toBe(headings[1].id);
  });

  it("preserves an explicit accessible label without overriding it with the heading", async () => {
    const dialog = await render({ "aria-label": "Create a topology" });
    expect(dialog.getAttribute("aria-label")).toBe("Create a topology");
    expect(dialog.hasAttribute("aria-labelledby")).toBe(false);
  });

  it("preserves an explicit labelledby reference", async () => {
    const dialog = await render({ "aria-labelledby": "custom-heading" });
    expect(dialog.getAttribute("aria-labelledby")).toBe("custom-heading");
  });
});
