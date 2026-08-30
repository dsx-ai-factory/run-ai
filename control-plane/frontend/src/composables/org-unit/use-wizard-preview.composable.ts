import { inject, provide, ref, type Ref } from "vue";
import type { NodeTypesPerWorkload, SchedulingRules } from "@/swagger-models/org-unit-service-client";

/**
 * Wizard Preview State Composable
 *
 * Allows step components to publish uncommitted changes that the summary
 * component can display as a live preview before saving.
 *
 * Usage:
 * ```ts
 * // Parent component (e.g., project-edit-form-v2)
 * const preview = provideWizardPreview();
 *
 * // Child step component
 * const preview = useWizardPreview();
 * watch(localState, () => {
 *   preview.updatePreview('gpuQuota', { ... });
 * });
 *
 * onBeforeUnmount(() => {
 *   preview.clearPreview('gpuQuota');
 * });
 *
 * // Summary component
 * const preview = useWizardPreview();
 * const rules = computed(() => preview.state.value.schedulingRules?.schedulingRules ?? props.schedulingRules);
 * ```
 */

export interface IGpuQuotaPreview {
  nodePoolName: string;
  gpuQuota: number;
  defaultNodePools?: string[];
}

export interface IGpuOverQuotaPreview {
  nodePoolName: string;
  overQuotaEnabled: boolean;
  limit: number | null;
  defaultNodePools?: string[];
}

export interface ISchedulingRulesPreview {
  schedulingRules: SchedulingRules | null;
  nodeTypes?: NodeTypesPerWorkload;
}

export interface IWizardPreviewState {
  gpuQuota?: IGpuQuotaPreview;
  gpuOverQuota?: IGpuOverQuotaPreview;
  schedulingRules?: ISchedulingRulesPreview;
}

export interface IWizardPreview {
  /** Reactive state containing all preview data from active steps */
  state: Ref<IWizardPreviewState>;
  /** Update a specific preview key with new data */
  updatePreview<K extends keyof IWizardPreviewState>(key: K, value: IWizardPreviewState[K]): void;
  /** Clear preview data for a specific key, or all keys if none specified */
  clearPreview(key?: keyof IWizardPreviewState): void;
}

const WIZARD_PREVIEW_SYMBOL = Symbol("wizard-preview");

/**
 * Null-object pattern: provides a no-op implementation when used outside provider context.
 * This eliminates the need for optional chaining in consumers.
 */
const NULL_PREVIEW: IWizardPreview = {
  state: ref({}),
  updatePreview: () => {},
  clearPreview: () => {},
};

/**
 * Creates and provides the wizard preview state.
 * Call this in the parent wizard/form component.
 */
export const provideWizardPreview = (): IWizardPreview => {
  const preview = createWizardPreview();
  provide(WIZARD_PREVIEW_SYMBOL, preview);
  return preview;
};

/**
 * Injects the wizard preview state.
 * Call this in step components to update preview, or in summary to read preview.
 * Returns a no-op implementation if used outside of a provider context.
 */
export const useWizardPreview = (): IWizardPreview => {
  return inject<IWizardPreview>(WIZARD_PREVIEW_SYMBOL, NULL_PREVIEW);
};

/**
 * Injects the wizard preview state, throwing if not available.
 * Use this when the preview is required and you want to fail fast.
 */
export const useWizardPreviewRequired = (): IWizardPreview => {
  const preview = inject<IWizardPreview>(WIZARD_PREVIEW_SYMBOL, NULL_PREVIEW);
  if (preview === NULL_PREVIEW) {
    throw new Error("useWizardPreviewRequired must be used within provideWizardPreview");
  }
  return preview;
};

const createWizardPreview = (): IWizardPreview => {
  const state = ref<IWizardPreviewState>({});

  const updatePreview = <K extends keyof IWizardPreviewState>(key: K, value: IWizardPreviewState[K]): void => {
    state.value = {
      ...state.value,
      [key]: value,
    };
  };

  const clearPreview = (key?: keyof IWizardPreviewState): void => {
    if (key) {
      const newState = { ...state.value };
      delete newState[key];
      state.value = newState;
    } else {
      state.value = {};
    }
  };

  return {
    state,
    updatePreview,
    clearPreview,
  };
};
