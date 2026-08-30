import { ref, computed, watch, type Ref, type ComputedRef } from "vue";
import type { Resources } from "@/swagger-models/org-unit-service-client";
import { DEFAULT_NODE_POOL_NAME } from "@/models/resource.model";

export interface OrgUnitNodePoolOption {
  id: string;
  name: string;
}

export interface UseOrgUnitNodePoolsOptions {
  /** Resources array containing node pool info */
  resources: () => Resources[];
  /** Current default node pools (from the entity being edited) */
  defaultNodePools: () => string[] | undefined;
  /** Fallback from parent (e.g., department's defaults for project) */
  parentDefaultNodePools?: () => string[] | undefined;
  /** Whether node pool list can be modified (default: false) */
  editable?: boolean;
}

export interface UseOrgUnitNodePoolsReturn {
  /** All node pools extracted from resources */
  allNodePools: ComputedRef<OrgUnitNodePoolOption[]>;
  /** Selected node pools (mapped from defaultNodePools to NodePoolOption) */
  selectedNodePools: ComputedRef<OrgUnitNodePoolOption[]>;
  /** Currently selected node pool ID (for tab selection) */
  selectedNodePoolId: Ref<string>;
  /** Currently selected node pool object */
  selectedNodePool: ComputedRef<OrgUnitNodePoolOption | null>;
  /** Currently selected node pool name (for resource lookups) */
  selectedNodePoolName: ComputedRef<string>;
  /** Local state for node pool list (only meaningful when editable) */
  localDefaultNodePools: Ref<string[]>;
  /** Whether the node pool list has unsaved changes (only meaningful when editable) */
  isNodePoolListDirty: ComputedRef<boolean>;
  /** Update the local node pool list (only meaningful when editable) */
  updateNodePools: (pools: string[]) => void;
  /** Reset node pool list to baseline (only meaningful when editable) */
  resetNodePoolsToBaseline: () => void;
}

export function useOrgUnitNodePools(options: UseOrgUnitNodePoolsOptions): UseOrgUnitNodePoolsReturn {
  const { resources, defaultNodePools, parentDefaultNodePools, editable = false } = options;

  // Get initial default node pools with fallback logic
  const getInitialDefaultNodePools = (): string[] => {
    const current = defaultNodePools();
    if (current?.length) {
      return [...current];
    }
    const parent = parentDefaultNodePools?.();
    if (parent?.length) {
      return [...parent];
    }
    return [DEFAULT_NODE_POOL_NAME];
  };

  // Baseline represents the "clean/saved" state (with fallback applied)
  // Only used for editable mode to track dirty state
  // Must be reactive so isNodePoolListDirty re-computes when baseline changes
  const baselineDefaultNodePools = ref<string[]>(getInitialDefaultNodePools());

  // Local state for node pool list
  // In editable mode: user can modify this
  // In read-only mode: syncs with props
  const localDefaultNodePools = ref<string[]>([...baselineDefaultNodePools.value]);

  // Currently selected node pool (by ID)
  const selectedNodePoolId = ref<string>("");

  // Extract all node pools from resources
  const allNodePools = computed<OrgUnitNodePoolOption[]>(() =>
    resources()
      .map((resource) => ({
        id: resource.nodePool?.id ?? "",
        name: resource.nodePool?.name ?? "",
      }))
      .filter((pool) => pool.name),
  );

  // Selected node pools mapped to NodePoolOption format
  const selectedNodePools = computed<OrgUnitNodePoolOption[]>(() => {
    const poolNames = editable ? localDefaultNodePools.value : getInitialDefaultNodePools();
    return poolNames
      .map((poolName) => allNodePools.value.find((np) => np.name === poolName))
      .filter((np): np is OrgUnitNodePoolOption => np !== undefined);
  });

  // Check if node pool list has unsaved changes (compares against baseline)
  const isNodePoolListDirty = computed(() => {
    if (!editable) return false;
    const current = localDefaultNodePools.value;
    const baseline = baselineDefaultNodePools.value;
    if (baseline.length !== current.length) return true;
    return baseline.some((name, index) => name !== current[index]);
  });

  // Sync baseline and local state when source props change
  watch(
    () => defaultNodePools(),
    () => {
      // Update baseline to new saved state (with fallback applied)
      baselineDefaultNodePools.value = getInitialDefaultNodePools();

      if (editable) {
        // In editable mode: sync local to match new baseline (after save)
        // This ensures dirty state clears after save
        localDefaultNodePools.value = [...baselineDefaultNodePools.value];
      } else {
        // In read-only mode: always sync local with source
        localDefaultNodePools.value = [...baselineDefaultNodePools.value];
      }
    },
  );

  // Get the selected node pool object
  const selectedNodePool = computed<OrgUnitNodePoolOption | null>(() => {
    return selectedNodePools.value.find((np) => np.id === selectedNodePoolId.value) || null;
  });

  // Selected node pool name (for resource lookups)
  const selectedNodePoolName = computed(() => selectedNodePool.value?.name || "");

  // Initialize selected node pool to first in selectedNodePools
  watch(
    () => selectedNodePools.value,
    (pools) => {
      const currentSelectionExists = pools.some((np) => np.id === selectedNodePoolId.value);
      if (pools.length > 0 && (!selectedNodePoolId.value || !currentSelectionExists)) {
        selectedNodePoolId.value = pools[0].id;
      }
    },
    { immediate: true },
  );

  // Update local node pool list (only meaningful when editable)
  const updateNodePools = (pools: string[]) => {
    localDefaultNodePools.value = pools;
  };

  // Reset node pool list to baseline
  const resetNodePoolsToBaseline = () => {
    localDefaultNodePools.value = [...baselineDefaultNodePools.value];
  };

  return {
    allNodePools,
    selectedNodePools,
    selectedNodePoolId,
    selectedNodePool,
    selectedNodePoolName,
    localDefaultNodePools,
    isNodePoolListDirty,
    updateNodePools,
    resetNodePoolsToBaseline,
  };
}
