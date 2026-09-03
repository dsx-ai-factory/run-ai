/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import { defineComponent, type ComponentPublicInstance } from "vue";

import {
  provideUnsavedChangesGuard,
  type IUnsavedChangesGuard,
} from "@/composables/org-unit/use-unsaved-changes-guard.composable";

const GuardHost = defineComponent({
  setup() {
    const guard = provideUnsavedChangesGuard();
    return { guard };
  },
  template: "<div />",
});

const mountGuard = () => {
  const wrapper = mount(GuardHost);
  const guard = (wrapper.vm as ComponentPublicInstance<{ guard: IUnsavedChangesGuard }>).guard;
  return { guard, wrapper };
};

const flushGuardQueue = async () => {
  await Promise.resolve();
};

describe("use-unsaved-changes-guard", () => {
  let guard: IUnsavedChangesGuard;
  let wrapper: VueWrapper<ComponentPublicInstance<{ guard: IUnsavedChangesGuard }>>;

  beforeEach(() => {
    const mounted = mountGuard();
    guard = mounted.guard;
    wrapper = mounted.wrapper as VueWrapper<ComponentPublicInstance<{ guard: IUnsavedChangesGuard }>>;
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("resolves immediately for unregistered steps and executes onProceed", async () => {
    const onProceed = vi.fn();
    const result = await guard.requestAction({ stepName: "cpu", onProceed });
    expect(result).toBe(true);
    expect(onProceed).toHaveBeenCalledTimes(1);
    expect(guard.dialog.isOpen.value).toBe(false);
  });

  it("bypasses the dialog when the registered step is clean", async () => {
    const onProceed = vi.fn();
    guard.registerStepGuard("cpu", {
      isDirty: () => false,
    });

    const result = await guard.requestAction({ stepName: "cpu", onProceed });
    expect(result).toBe(true);
    expect(onProceed).toHaveBeenCalledTimes(1);
    expect(guard.dialog.isOpen.value).toBe(false);
  });

  it("opens the dialog and resolves after saving dirty state", async () => {
    const save = vi.fn();
    const onProceed = vi.fn();
    guard.registerStepGuard("cpu", {
      isDirty: () => true,
      save,
    });

    const pending = guard.requestAction({ stepName: "cpu", onProceed });
    await flushGuardQueue();
    expect(guard.dialog.isOpen.value).toBe(true);

    await guard.confirmAction("save");
    const result = await pending;

    expect(result).toBe(true);
    expect(save).toHaveBeenCalledTimes(1);
    expect(onProceed).toHaveBeenCalledTimes(1);
    expect(guard.dialog.isOpen.value).toBe(false);
  });

  it("runs discard handler when choosing discard", async () => {
    const discard = vi.fn();
    guard.registerStepGuard("cpu", {
      isDirty: () => true,
      discard,
    });

    const pending = guard.requestAction({ stepName: "cpu" });
    await flushGuardQueue();
    await guard.confirmAction("discard");
    const result = await pending;

    expect(result).toBe(true);
    expect(discard).toHaveBeenCalledTimes(1);
  });

  it("resolves false when cancelPending is called", async () => {
    guard.registerStepGuard("cpu", {
      isDirty: () => true,
    });

    const pending = guard.requestAction({ stepName: "cpu" });
    await flushGuardQueue();
    guard.cancelPending();
    const result = await pending;

    expect(result).toBe(false);
    expect(guard.dialog.isOpen.value).toBe(false);
  });

  it("unregisters the handler correctly", async () => {
    const unregister = guard.registerStepGuard("cpu", {
      isDirty: () => true,
    });

    unregister();

    const result = await guard.requestAction({ stepName: "cpu" });
    expect(result).toBe(true);
    expect(guard.dialog.isOpen.value).toBe(false);
  });

  it("rejects the pending promise if save throws", async () => {
    const error = new Error("Save failed");
    guard.registerStepGuard("cpu", {
      isDirty: () => true,
      save: () => Promise.reject(error),
    });

    const pending = guard.requestAction({ stepName: "cpu" });
    await flushGuardQueue();

    // We must catch the rejection to avoid test failure
    const promise = guard.confirmAction("save");

    await expect(pending).rejects.toThrow("Save failed");
    await promise; // cleanup

    expect(guard.dialog.isOpen.value).toBe(false);
  });

  it("handles async isDirty check", async () => {
    guard.registerStepGuard("cpu", {
      isDirty: async () => true,
    });

    const pending = guard.requestAction({ stepName: "cpu" });

    await flushGuardQueue();

    expect(guard.dialog.isOpen.value).toBe(true);

    guard.cancelPending();
    await pending;
  });

  describe("isAnyStepDirty", () => {
    it("returns false when no steps are registered", () => {
      expect(guard.isAnyStepDirty.value).toBe(false);
    });

    it("returns false when all registered steps are clean", () => {
      guard.registerStepGuard("cpu", { isDirty: () => false });
      guard.registerStepGuard("gpu", { isDirty: () => false });
      expect(guard.isAnyStepDirty.value).toBe(false);
    });

    it("returns true when any registered step is dirty", () => {
      guard.registerStepGuard("cpu", { isDirty: () => false });
      guard.registerStepGuard("gpu", { isDirty: () => true });
      expect(guard.isAnyStepDirty.value).toBe(true);
    });

    it("updates when a step is unregistered", () => {
      guard.registerStepGuard("cpu", { isDirty: () => false });
      const unregister = guard.registerStepGuard("gpu", { isDirty: () => true });
      expect(guard.isAnyStepDirty.value).toBe(true);

      unregister();
      expect(guard.isAnyStepDirty.value).toBe(false);
    });
  });
});
