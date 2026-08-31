/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { computed, inject, nextTick, provide, ref, type ComputedRef, type Ref } from "vue";

/**
 * Unsaved changes guard composable
 *
 * Summary of flow:
 *
 *  * Code tries to run: await requestAction()
 *
 *  * Guard freezes time: Returns a Promise that doesn't resolve.
 *
 *  * Guard waits for UI: Opens a popup.
 *
 *  * UI resumes time: Finds the stored resolve function and calls it.
 *
 * Usage:
 * ```ts
 * // Parent component (e.g., wizard container)
 * const guard = provideUnsavedChangesGuard();
 *
 * await guard.requestAction({
 *   stepName: currentStep.value?.name ?? "",
 *   onProceed: () => navigateToStep(targetStep),
 * });
 *
 * // Child step component
 * const guard = useUnsavedChangesGuard();
 * guard.registerStepGuard("cpu-quota", {
 *   isDirty: () => isDirty.value,
 *   save: saveChanges,
 *   discard: resetForm,
 * });
 * ```
 */

type DirtyEvaluator = () => boolean | Promise<boolean>;
type StepAction = () => void | Promise<void>;

export interface IUnsavedStepGuardHandlers {
  isDirty: DirtyEvaluator;
  save?: StepAction;
  discard?: StepAction;
}

export type UnsavedGuardDecision = "save" | "discard";

export interface IUnsavedGuardRequest {
  stepName: string;
  onProceed?: () => void | Promise<void>;
}

interface IUnsavedGuardDialogState {
  isOpen: Ref<boolean>;
}

interface IPendingIntent {
  stepName: string;
  handler: IUnsavedStepGuardHandlers;
  onProceed?: () => void | Promise<void>;
  resolve: (shouldProceed: boolean) => void;
  reject: (error: unknown) => void;
}

export interface IUnsavedChangesGuard {
  registerStepGuard(stepName: string, handlers: IUnsavedStepGuardHandlers): () => void;
  requestAction(request: IUnsavedGuardRequest): Promise<boolean>;
  confirmAction(decision: UnsavedGuardDecision): Promise<void>;
  cancelPending(): void;
  isAnyStepDirty: ComputedRef<boolean>;
  dialog: IUnsavedGuardDialogState;
}

const UNSAVED_GUARD_SYMBOL = Symbol("unsaved-changes-guard");

export const provideUnsavedChangesGuard = (): IUnsavedChangesGuard => {
  const guard = createUnsavedChangesGuard();
  provide(UNSAVED_GUARD_SYMBOL, guard);
  return guard;
};

export const useUnsavedChangesGuard = (): IUnsavedChangesGuard => {
  const guard = inject<IUnsavedChangesGuard | null>(UNSAVED_GUARD_SYMBOL, null);
  if (!guard) {
    throw new Error("useUnsavedChangesGuard must be used within provideUnsavedChangesGuard");
  }
  return guard;
};

const createUnsavedChangesGuard = (): IUnsavedChangesGuard => {
  const stepHandlers = new Map<string, IUnsavedStepGuardHandlers>();
  const dialogVisible = ref(false);
  const registrationTrigger = ref(0);

  let pending: IPendingIntent | null = null;

  const closeDialog = () => {
    dialogVisible.value = false;
  };

  const registerStepGuard = (stepName: string, handlers: IUnsavedStepGuardHandlers): (() => void) => {
    stepHandlers.set(stepName, handlers);
    registrationTrigger.value++;
    return () => {
      stepHandlers.delete(stepName);
      registrationTrigger.value++;
    };
  };

  const requestAction = async (request: IUnsavedGuardRequest): Promise<boolean> => {
    const handler = stepHandlers.get(request.stepName);
    if (!handler) {
      await request.onProceed?.();
      return true;
    }

    const dirty = await handler.isDirty();
    if (!dirty) {
      await request.onProceed?.();
      return true;
    }

    dialogVisible.value = true;

    return new Promise<boolean>((resolve, reject) => {
      pending = {
        stepName: request.stepName,
        handler,
        onProceed: request.onProceed,
        resolve,
        reject,
      };
    });
  };

  const confirmAction = async (decision: UnsavedGuardDecision): Promise<void> => {
    if (!pending) {
      return;
    }

    try {
      if (decision === "save") {
        await pending.handler.save?.();
      } else {
        await pending.handler.discard?.();
      }
      // Flush reactive updates so downstream dirty-state checks
      // (e.g. requestToLeave → form-wrapper isDirty) see the saved state.
      await nextTick();
      await pending.onProceed?.();
      await nextTick();
      pending.resolve(true);
    } catch (error) {
      pending.reject(error);
    } finally {
      pending = null;
      closeDialog();
    }
  };

  const cancelPending = (): void => {
    if (!pending) {
      return;
    }
    pending.resolve(false);
    pending = null;
    closeDialog();
  };

  const isAnyStepDirty = computed(() => {
    // Access trigger so Vue tracks registration changes
    void registrationTrigger.value;
    for (const handler of stepHandlers.values()) {
      if (handler.isDirty()) {
        return true;
      }
    }
    return false;
  });

  return {
    registerStepGuard,
    requestAction,
    confirmAction,
    cancelPending,
    isAnyStepDirty,
    dialog: {
      isOpen: dialogVisible,
    },
  };
};
