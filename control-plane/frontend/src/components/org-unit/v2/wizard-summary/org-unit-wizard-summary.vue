<template>
  <div class="org-unit-wizard-summary column full-height q-pa-md">
    <div class="section-title q-mb-xs">Resources</div>

    <div class="column q-gutter-md">
      <section>
        <div class="sub-section-title fw-800 q-mb-xs">{{ resourcesSubtitle }}</div>
        <div class="content-text fw-400">
          {{ resourcesStateSummary }}
        </div>
      </section>

      <template v-if="showQuotaSection">
        <section>
          <div class="sub-section-title fw-500 q-mb-xs">{{ gpuQuotaLabel }}</div>
          <div class="content-text fw-400" v-if="!singleNodePoolResourceAvailable">
            <div v-for="item in gpuQuotaByNodePool" :key="item.name">
              <span>{{ item.name }}</span
              >: <span>{{ item.value }}</span>
            </div>
          </div>
        </section>

        <section>
          <div class="sub-section-title fw-500 q-mb-xs">{{ gpuOverQuotaLabel }}</div>
          <div class="content-text fw-400" v-if="!singleNodePoolResourceAvailable">
            <div v-for="item in gpuOverQuotaByNodePool" :key="item.name">
              <span>{{ item.name }}</span
              >: <span>{{ item.value ? "Enabled" : "Disabled" }}</span>
            </div>
          </div>
        </section>
      </template>
      <q-separator />

      <section>
        <div class="row items-center">
          <div class="section-title">Scheduling rules</div>
          <runai-tooltip
            class="q-ml-sm"
            :tooltip-text="`Apply resource and runtime restrictions to all workloads associated with the ${props.entityType}`"
          />
        </div>
        <div class="column q-mt-xs">
          <div class="row items-baseline fw-500">
            <div class="fs-16">{{ schedulingRulesCount }}</div>
            <div class="q-ml-xs">{{ schedulingRulesCount === 1 ? "rule" : "rules" }}</div>
          </div>
        </div>
      </section>

      <q-separator />

      <section>
        <div class="row items-center">
          <div class="section-title">Access rules</div>
          <runai-tooltip class="q-ml-sm" :tooltip-text="`Authorize access to this ${props.entityType}`" />
        </div>
        <div class="column q-mt-xs">
          <div class="row items-baseline fw-500">
            <div class="fs-16">{{ props.accessRulesCount }}</div>
            <div class="q-ml-xs">{{ props.accessRulesCount === 1 ? "rule" : "rules" }}</div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {
  Resources,
  ResourcesNullable,
  ResourcesNullableGpu,
  SchedulingRules,
} from "@/swagger-models/org-unit-service-client";
import { computed } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { useWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { resourceUtil } from "@/utils/resource.util";
import { EGpuResourceState } from "@/models/org-unit-forms.model";
import { EQuotaEntity } from "@/models/resource.model";

const props = defineProps<{
  entityType: EQuotaEntity;
  resources: Array<Resources | ResourcesNullable>;
  defaultNodePools: string[];
  schedulingRules?: SchedulingRules | null;
  scopeId?: string;
  accessRulesCount: number;
}>();

// Inject wizard preview state for real-time updates
const wizardPreview = useWizardPreview();

/**
 * =============================================================================
 * RESOURCES SUMMARY
 * =============================================================================
 *
 * The summary panel displays GPU resource allocation status based on two key values
 * per node pool: "deserved" (guaranteed quota) and "limit" (max allowed).
 *
 * GPU STATE CLASSIFICATION (per node pool):
 * ─────────────────────────────────────────
 * Each node pool is classified into one of four states based on its GPU config:
 *
 * 1. NoGPUResourcesAreSet (limit === 0)
 *    → Workloads cannot run on GPUs at all
 *
 * 2. OpportunisticOnly (deserved === 0, limit > 0 or unlimited)
 *    → No guaranteed resources, but can use spare GPUs when available
 *    → Workloads are "best effort" and may be preempted
 *
 * 3. DedicatedGPUs (deserved > 0, limit === deserved)
 *    → Guaranteed resources with no over-quota allowed
 *    → Workloads get exactly what's promised, no more
 *
 * 4. DedicatedAndOpportunisticGPUs (deserved > 0, limit > deserved or unlimited)
 *    → Guaranteed resources PLUS ability to use spare GPUs
 *    → Best of both worlds: baseline guarantee + opportunistic usage
 *
 * AGGREGATION ACROSS NODE POOLS:
 * ──────────────────────────────
 * When an org unit has multiple node pools, states are aggregated with this logic:
 *
 * - If ANY node pool has DedicatedAndOpportunistic → show that (highest priority)
 * - If MIXED: some Opportunistic + some Dedicated → show DedicatedAndOpportunistic
 *   (because overall the org unit has both dedicated and opportunistic resources)
 * - Otherwise, use the highest priority state found
 *
 * Priority order: DedicatedAndOpportunistic > Dedicated > Opportunistic > None
 *
 * See: orgUnitUtil.getNodePoolGpuState() and orgUnitUtil.aggregateGpuStates()
 * =============================================================================
 */

export interface IResourceStateTexts {
  subtitle: string;
  summary: string;
}

/**
 * Maps each GPU state to its display texts (subtitle + summary).
 * Using Record<EGpuResourceState, ...> ensures TypeScript will error if any state is missing.
 */
const RESOURCE_STATE_TEXT_MAP: Record<EGpuResourceState, IResourceStateTexts> = {
  [EGpuResourceState.OpportunisticOnly]: {
    subtitle: "Opportunistic only",
    summary: "No dedicated GPU resources are set. Workloads can only run when spare GPUs are available.",
  },
  [EGpuResourceState.DedicatedAndOpportunisticGPUs]: {
    subtitle: "Dedicated and opportunistic GPUs",
    summary: "Guaranteed GPU resources are set. Workloads can also use spare GPUs when available.",
  },
  [EGpuResourceState.DedicatedGPUs]: {
    subtitle: "Dedicated GPUs",
    summary: "Guaranteed GPU resources are set. Workloads cannot exceed the dedicated GPU quota.",
  },
  [EGpuResourceState.NoGPUResourcesAreSet]: {
    subtitle: "No GPU resources are set",
    summary: "Workloads cannot run on GPUs.",
  },
};

/**
 * Get the effective defaultNodePools (preview or saved)
 */
const effectiveDefaultNodePools = computed(() => {
  const previewNodePools = wizardPreview.state.value.gpuQuota?.defaultNodePools;
  return previewNodePools ?? props.defaultNodePools ?? [];
});

/**
 * All node pools sorted (default first) with preview GPU data applied.
 * Used for summary display and as the base for relevantNodePools.
 */
const allNodePoolsWithPreview = computed(() => {
  const gpuQuotaPreview = wizardPreview.state.value.gpuQuota;
  const gpuOverQuotaPreview = wizardPreview.state.value.gpuOverQuota;
  const defaultNodePoolsOrder = effectiveDefaultNodePools.value;

  const sortedNodePools = resourceUtil.sortResourcesByNodePoolOrder(props.resources ?? [], defaultNodePoolsOrder);

  return sortedNodePools.map((resource) => {
    const nodePoolName = resource.nodePool?.name ?? "";
    const hasQuotaPreview = gpuQuotaPreview?.nodePoolName === nodePoolName;
    const hasOverQuotaPreview = gpuOverQuotaPreview?.nodePoolName === nodePoolName;

    if (!hasQuotaPreview && !hasOverQuotaPreview) {
      return resource;
    }

    const gpu: ResourcesNullableGpu = {
      ...resource.gpu,
      ...(hasQuotaPreview ? { deserved: gpuQuotaPreview.gpuQuota } : {}),
      ...(hasOverQuotaPreview ? { limit: gpuOverQuotaPreview.limit === null ? -1 : gpuOverQuotaPreview.limit } : {}),
    };

    return { ...resource, gpu };
  });
});

/**
 * Only default node pools — used for GPU state aggregation (determines workload scheduling behavior).
 */
const relevantNodePools = computed(() => {
  const defaultNodePoolsOrder = effectiveDefaultNodePools.value;
  return allNodePoolsWithPreview.value.filter((resource) =>
    defaultNodePoolsOrder.includes(resource.nodePool?.name ?? ""),
  );
});

/**
 * Aggregates GPU states from all relevant nodepools into a single state
 * that represents the overall GPU resource allocation for this org unit.
 * See business logic documentation above for aggregation rules.
 */
const aggregatedGpuState = computed((): EGpuResourceState => {
  const states = relevantNodePools.value.map((resource) => orgUnitUtil.getNodePoolGpuState(resource.gpu));
  return orgUnitUtil.aggregateGpuStates(states);
});

/**
 * Gets the display texts (subtitle + summary) for the current aggregated GPU state.
 */
const resourceStateTexts = computed((): IResourceStateTexts => RESOURCE_STATE_TEXT_MAP[aggregatedGpuState.value]);

/** Bold heading shown in the Resources section */
const resourcesSubtitle = computed(() => resourceStateTexts.value.subtitle);

/** Descriptive text explaining what the configuration means for workload scheduling */
const resourcesStateSummary = computed(() => resourceStateTexts.value.summary);

const singleNodePoolResourceAvailable = computed(() => props.resources?.length === 1);

const gpuQuotaLabel = computed(() => {
  // if there is only a single node pool available, return "GPU quota: {value}"
  if (singleNodePoolResourceAvailable.value) {
    const value = gpuQuotaByNodePool.value[0]?.value;
    return `GPU quota: ${value ?? "0"}`;
  }
  return "GPU quota by node pool";
});

const gpuOverQuotaLabel = computed(() => {
  // if there is only a single node pool available, return "GPU over quota: Enabled/Disabled"
  if (singleNodePoolResourceAvailable.value) {
    return `GPU over quota: ${gpuOverQuotaByNodePool.value[0].value ? "Enabled" : "Disabled"}`;
  }
  return "GPU over quota by node pool";
});

/**
 * Formats GPU quota value for display.
 * For departments, -1 means "Unlimited".
 * Negative values (except department's -1) are invalid.
 */
const formatGpuQuotaValue = (value: number | null | undefined): string => {
  const numValue = value ?? 0;

  // For departments, -1 means "Unlimited"
  if (props.entityType === EQuotaEntity.department && numValue === -1) {
    return "Unlimited";
  }

  if (numValue < 0) {
    return "-";
  }

  return String(numValue);
};

const gpuQuotaByNodePool = computed(() => {
  // Preview values are already applied in allNodePoolsWithPreview
  // Shows ALL node pools with quota, not just default ones
  return allNodePoolsWithPreview.value.map((resource) => ({
    name: resource.nodePool?.name ?? "",
    value: formatGpuQuotaValue(resource.gpu?.deserved),
  }));
});

/**
 * Determines if over-quota (opportunistic usage) is enabled per node pool.
 * Over-quota is enabled when:
 * - limit > deserved (can use more than guaranteed)
 * - limit === -1 (unlimited)
 */
const gpuOverQuotaByNodePool = computed(() => {
  const gpuOverQuotaPreview = wizardPreview.state.value.gpuOverQuota;
  // Shows ALL node pools, not just default ones
  return allNodePoolsWithPreview.value.map((resource) => {
    const isOverQuotaEnabled = IsOverQuotaEnabled(resource.gpu?.limit ?? null, resource.gpu?.deserved ?? null);

    if (gpuOverQuotaPreview?.nodePoolName === resource.nodePool?.name) {
      return {
        name: resource.nodePool?.name ?? "",
        value: gpuOverQuotaPreview?.overQuotaEnabled ?? false,
      };
    }

    return {
      name: resource.nodePool?.name ?? "",
      value: isOverQuotaEnabled,
    };
  });
});

const IsOverQuotaEnabled = (limitValue: number | null, deservedValue: number | null): boolean => {
  if (limitValue === null || limitValue === -1) return true;
  if (deservedValue === null) return true;
  return limitValue > deservedValue;
};

const schedulingRulesCount = computed(() => {
  const previewSchedulingRules = wizardPreview.state.value.schedulingRules;
  // Use preview data if available, otherwise use saved prop data
  const schedulingRules = previewSchedulingRules?.schedulingRules ?? props.schedulingRules;

  // count only the keys that has value
  return Object.keys(schedulingRules ?? {}).filter(
    (key) =>
      schedulingRules?.[key as keyof typeof schedulingRules] !== undefined &&
      schedulingRules?.[key as keyof typeof schedulingRules] !== null,
  ).length;
});

const showQuotaSection = computed(() => {
  // check if id exists, if not hide the section
  if (!props.scopeId) return false;

  return gpuQuotaByNodePool.value.length > 0;
});
</script>

<style scoped lang="scss">
.org-unit-wizard-summary {
  font-size: 12px;
  // Ensure container constrains width for word-break to work
  overflow: hidden;

  .section-title {
    font-weight: 500;
    color: $black-54;
    line-height: 140%;
  }

  .items-baseline {
    margin-top: -6px;
  }

  .fw-400 {
    font-weight: 400;
  }

  .fw-500 {
    font-weight: 500;
  }

  .fw-800 {
    font-weight: 800;
  }

  .content-text {
    word-break: break-word;
    overflow-wrap: anywhere;
  }
}
</style>
