/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { afterEach, describe, expect, it, vi } from "vitest";
import RunaiFormWrapper from "./runai-form-wrapper.vue";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { eventEmitter } from "@/services/infra/event-emitter.service/event-emitter.service";
import { NAVIGATION_EVENTS } from "@/common/event.constant";

const wrappers: VueWrapper[] = [];
function render(preventLeave = false) {
  const wrapper = mount(RunaiFormWrapper, {
    props: { formState: { name: "Initial" }, preventLeave },
    global: { plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })] },
  });
  wrappers.push(wrapper);
  return wrapper;
}
afterEach(() => {
  eventEmitter.emit(NAVIGATION_EVENTS.LEAVE_PAGE_STATUS, { allowToLeave: false });
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount());
});

describe("form navigation while persistence is pending", () => {
  it.each([false, true])("immediately refuses header leave without a dialog when dirty=%s", async (dirty) => {
    const wrapper = render(true);
    if (dirty) await wrapper.setProps({ formState: { name: "Changed" } });
    let allowed: boolean | undefined;
    void requestToLeave().then((value) => {
      allowed = value;
    });
    await flushPromises();
    expect(allowed).toBe(false);
    expect(document.querySelector('[aid="leave-page-modal-cancel-btn"]')).toBeNull();
    expect(wrapper.emitted("leave-page")).toBeUndefined();
  });

  it("restores the existing dirty confirmation once persistence finishes", async () => {
    const wrapper = render(true);
    await wrapper.setProps({ formState: { name: "Changed" }, preventLeave: false });
    let allowed: boolean | undefined;
    const decision = requestToLeave().then((value) => {
      allowed = value;
    });
    await flushPromises();
    expect(allowed).toBeUndefined();
    (document.querySelector('[aid="leave-page-modal-cancel-btn"]') as HTMLButtonElement).click();
    await decision;
    expect(allowed).toBe(true);
    expect(wrapper.emitted("leave-page")).toHaveLength(1);
  });

  it("warns before unloading a clean form while persistence is pending", () => {
    render(true);
    const event = new Event("beforeunload", { cancelable: true });
    window.dispatchEvent(event);
    expect(event.defaultPrevented).toBe(true);
  });
});
