<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <department-edit-form-v2
    v-if="!initialLoading && department"
    :key="formKey"
    :department="department"
    :is-new-department="isNewDepartment"
    :is-department-submitting="isDepartmentSubmitting"
    :node-pools-minimal="nodePoolsMinimal"
    :read-only="isReadOnly"
    :node-pool-totals="nodePoolTotals"
    @on-cancel="onCancel"
    @on-save="onSave"
    @update-scope="updateScope"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";

// cmps
import { DepartmentEditFormV2 } from "@/components/org-unit/v2/department/department-edit-form-v2";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";

// utils
import { alertUtil } from "@/utils/alert.util";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { deepCopy } from "@/utils/common.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { resourceUtil } from "@/utils/resource.util/resource.util";

// models
import type { Department, DepartmentCreationRequest, Resources } from "@/swagger-models/org-unit-service-client";
import {
  type NodepoolMinimal,
  NodepoolSortFilterFields,
  NodeTelemetryType,
  type TelemetryResponseValuesInner,
} from "@/swagger-models/cluster-service-client";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import { ErrorAlert } from "@/utils/error-alert.util";
import { ESaveOrgUnitAction, type ISaveOrgUnitPayload, type INodePoolTotals } from "@/models/org-unit-forms.model";
import type { INodePoolsNameAndId } from "@/models/node-pool.model";

const route = useRoute();
const router = useRouter();
const $q = useQuasar();
const appStore = useAppStore();
const authStore = useAuthStore();
const clusterStore = useClusterStore();
const settingStore = useSettingStore();

// State
const initialLoading = ref(true);
const isDepartmentSubmitting = ref(false);
const department = ref<Department | DepartmentCreationRequest | null>(null);
const clusterId = ref("");
const nodePoolsMinimal = ref<NodepoolMinimal[]>([]);
const nodePoolTotals = ref<INodePoolTotals>({});

// Computed
const formKey = computed(() => route.query.t?.toString() || "default");
const isNewDepartment = computed(() => {
  if (department.value && (department.value as Department)?.id) {
    return false;
  }

  return route.name === DEPARTMENT_ROUTE_NAMES.DEPARTMENT_NEW;
});

const departmentId = computed(() => route.params.id as string);
const isCpuEnabled = computed(() => settingStore.isCPUResourcesQuotaEnabled);
const isReadOnly = computed(() => route.name === DEPARTMENT_ROUTE_NAMES.DEPARTMENT_VIEW);

// Methods
const initClusterId = (): void => {
  if (route.query.clusterId) {
    clusterId.value = route.query.clusterId.toString();
  }
};

const getNodePoolsNamesAndIds = (): INodePoolsNameAndId[] => {
  return nodePoolsMinimal.value.map((nodePool) => ({
    id: nodePool.id,
    name: nodePool.name,
  }));
};

const loadNewDepartment = async (clusterIdParam: string): Promise<void> => {
  const emptyDepartment = orgUnitUtil.getEmptyDepartmentModel({
    clusterId: clusterIdParam,
    nodePoolsNamesAndIds: getNodePoolsNamesAndIds(),
    isCpuEnabled: isCpuEnabled.value,
  });
  department.value = emptyDepartment;
};

const loadExistingDepartment = async (): Promise<void> => {
  const loadedDepartment: Department = await orgUnitService.getDepartment(departmentId.value);
  department.value = deepCopy(loadedDepartment);
  appStore.setPageTitleOverride(`Edit department: ${loadedDepartment.name}`);
};

const loadNodepoolsSchedulingCapabilities = async (clusterIdToFilter: string): Promise<void> => {
  try {
    const filterBy = [filterUtil.getEqualsFilterString(NodepoolSortFilterFields.ClusterId, clusterIdToFilter)];
    nodePoolsMinimal.value = await clusterService.getNodepoolsMinimal(filterBy);
  } catch (err) {
    console.error("Failed to load node pools minimal:", err);
    nodePoolsMinimal.value = [];
  }
};

/**
 * Extracts node pool name from telemetry groups
 */
const getNodePoolNameFromGroups = (groups?: Array<{ key: string; value: string }>): string | null => {
  if (!groups) return null;
  const nodepoolGroup = groups.find((g) => g.key === "Nodepool");
  return nodepoolGroup?.value ?? null;
};

/**
 * Loads node pool total resources for validation.
 * GPU totals are always loaded; CPU and memory are loaded only if CPU quota is enabled.
 */
const loadNodePoolTotals = async (clusterIdToLoad: string): Promise<void> => {
  if (!clusterIdToLoad) {
    nodePoolTotals.value = {};
    return;
  }

  try {
    const groupBy = [ENodeTelemetryGroupBy.Nodepool];
    const isCpuEnabled = settingStore.isCPUResourcesQuotaEnabled;

    // Always fetch GPU totals; conditionally fetch CPU/memory
    const telemetryPromises: Promise<{ values?: TelemetryResponseValuesInner[] }>[] = [
      clusterService.getNodeTelemetry(NodeTelemetryType.TotalGpus, clusterIdToLoad, undefined, groupBy),
    ];

    if (isCpuEnabled) {
      telemetryPromises.push(
        clusterService.getNodeTelemetry(NodeTelemetryType.TotalCpuCores, clusterIdToLoad, undefined, groupBy),
        clusterService.getNodeTelemetry(NodeTelemetryType.TotalCpuMemoryBytes, clusterIdToLoad, undefined, groupBy),
      );
    }

    const responses = await Promise.all(telemetryPromises);
    const gpuResponse = responses[0];
    const cpuResponse = isCpuEnabled ? responses[1] : null;
    const memoryResponse = isCpuEnabled ? responses[2] : null;

    const totals: INodePoolTotals = {};

    const ensureNodePool = (nodePoolName: string) => {
      if (!totals[nodePoolName]) {
        totals[nodePoolName] = { cpuCores: 0, memoryBytes: 0, totalGpus: 0 };
      }
    };

    // Process total GPUs (always)
    gpuResponse.values?.forEach((item: TelemetryResponseValuesInner) => {
      const nodePoolName = getNodePoolNameFromGroups(item.groups);
      if (nodePoolName) {
        ensureNodePool(nodePoolName);
        totals[nodePoolName].totalGpus = parseFloat(item.value) || 0;
      }
    });

    // Process CPU cores (if enabled)
    cpuResponse?.values?.forEach((item: TelemetryResponseValuesInner) => {
      const nodePoolName = getNodePoolNameFromGroups(item.groups);
      if (nodePoolName) {
        ensureNodePool(nodePoolName);
        totals[nodePoolName].cpuCores = parseFloat(item.value) || 0;
      }
    });

    // Process memory bytes (if enabled)
    memoryResponse?.values?.forEach((item: TelemetryResponseValuesInner) => {
      const nodePoolName = getNodePoolNameFromGroups(item.groups);
      if (nodePoolName) {
        ensureNodePool(nodePoolName);
        totals[nodePoolName].memoryBytes = parseFloat(item.value) || 0;
      }
    });

    nodePoolTotals.value = totals;
  } catch (err) {
    console.warn("Failed to load node pool totals:", err);
    nodePoolTotals.value = {};
  }
};

/**
 * Initialize department with default values for UI display.
 * Ensures defaultNodePools is always populated - for departments, use ALL available node pools.
 * The "default" node pool is always placed first.
 */
const initializeDepartmentNodePools = (
  departmentData: Department | DepartmentCreationRequest,
): Department | DepartmentCreationRequest => {
  const initialized = deepCopy(departmentData);
  if (!initialized.defaultNodePools?.length) {
    // For departments: use ALL node pools from the cluster, with "default" first
    const allNodePoolNames = nodePoolsMinimal.value.map((np) => np.name).filter(Boolean) as string[];
    initialized.defaultNodePools = resourceUtil.sortNodePoolNames(allNodePoolNames);
  }
  return initialized;
};

const loadDepartment = async (clusterIdParam: string): Promise<void> => {
  if (isNewDepartment.value) {
    // Load node pools FIRST so they can be used to initialize resources
    if (clusterIdParam) {
      await Promise.all([loadNodepoolsSchedulingCapabilities(clusterIdParam), loadNodePoolTotals(clusterIdParam)]);
    }
    await loadNewDepartment(clusterIdParam);
    // Initialize defaultNodePools with all node pools
    if (department.value) {
      department.value = initializeDepartmentNodePools(department.value);
    }
  } else if (departmentId.value) {
    await loadExistingDepartment();
    // For existing departments, use the department's clusterId
    if (department.value?.clusterId) {
      await Promise.all([
        loadNodepoolsSchedulingCapabilities(department.value.clusterId),
        loadNodePoolTotals(department.value.clusterId),
      ]);
      // Initialize defaultNodePools with all node pools after they are loaded
      department.value = initializeDepartmentNodePools(department.value);
    }
  }
};

const updateScope = async (scope: { clusterId: string }): Promise<void> => {
  clusterId.value = scope.clusterId;
  // Load node pools FIRST so they can be used to initialize resources
  if (scope.clusterId) {
    await Promise.all([loadNodepoolsSchedulingCapabilities(scope.clusterId), loadNodePoolTotals(scope.clusterId)]);
  }
  await loadNewDepartment(scope.clusterId);
  // Initialize defaultNodePools with all node pools
  if (department.value) {
    department.value = initializeDepartmentNodePools(department.value);
  }
};

const onCancel = async (): Promise<void> => {
  const allowToLeave: boolean = await requestToLeave();
  if (allowToLeave) {
    router.back();
  }
};

// ==================== Save Logic Helpers ====================

/**
 * Prepares department data for API submission by applying transformations
 */
const prepareDepartmentForSave = (
  departmentData: Department | DepartmentCreationRequest,
): Department | DepartmentCreationRequest => {
  const prepared = deepCopy(departmentData) as Department | DepartmentCreationRequest;

  // Transform resources
  if (prepared.resources?.length) {
    // Align CPU/Memory limits with GPU over-quota state and reset over-quota weights to deserved
    prepared.resources = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(prepared.resources as Resources[]);
    // Ensure limits are not less than deserved to prevent API validation errors
    prepared.resources = orgUnitUtil.ensureLimitsNotBelowDeserved(prepared.resources as Resources[]);
    if (!settingStore.isOverQuotaPriorityEnabled) {
      // Reset over-quota weights to deserved if priority feature is disabled
      prepared.resources = orgUnitUtil.updateOverQuotaWeightToDeserved(prepared.resources as Resources[]);
    }
  }

  return prepared;
};

/**
 * Persists department to API (create or update)
 */
const saveDepartment = async (departmentData: Department | DepartmentCreationRequest): Promise<Department> => {
  if (isNewDepartment.value) {
    const created = await orgUnitService.createDepartment(departmentData as DepartmentCreationRequest);
    await authStore.loadUserOrgUnits();
    return created;
  } else {
    const updateRequest = orgUnitUtil.getDepartmentUpdateRequest(departmentData as Department);
    return await orgUnitService.updateDepartment((departmentData as Department).id, updateRequest);
  }
};

/**
 * Handles post-save actions (notification, state update, navigation)
 */
const handleSaveSuccess = (savedDepartment: Department, action: ESaveOrgUnitAction): void => {
  const actionVerb = isNewDepartment.value ? "created" : "saved";
  $q.notify(alertUtil.getSuccess(`Department ${savedDepartment.name} ${actionVerb}`));

  // The API response from update/create does not include projectsAggregatedResources, this cause subordinate quota validation fails,
  // remove next line after this is fixed
  const previousProjectsAggregatedResources = (department.value as Department)?.projectsAggregatedResources;

  const updatedDepartment = initializeDepartmentNodePools(savedDepartment);
  if (previousProjectsAggregatedResources) {
    (updatedDepartment as Department).projectsAggregatedResources = previousProjectsAggregatedResources;
  }

  department.value = updatedDepartment;

  if (action === ESaveOrgUnitAction.Close || isNewDepartment.value) {
    const routeQuery = route.meta.backPageName;
    const routeName = routeQuery ? routeQuery.toString() : DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX;
    const shouldApplyClusterFilter = isNewDepartment.value && routeName === DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX;

    if (shouldApplyClusterFilter) {
      clusterStore.setSelectedClusterById(savedDepartment.clusterId);
    }

    router.push({
      name: routeName,
      query: { createdEntityId: savedDepartment.id },
    });
  }
};

/**
 * Handles save errors with appropriate notifications
 */
const handleSaveError = (error: unknown): void => {
  const errorAlert = new ErrorAlert({
    generalMessage: isNewDepartment.value
      ? ErrorAlert.failedCreateMessage("department")
      : ErrorAlert.failedUpdateMessage("department"),
  });
  $q.notify(errorAlert.getNotification(error));
};

// ==================== Main Save Handler ====================

const onSave = async (payload: ISaveOrgUnitPayload<Department | DepartmentCreationRequest>) => {
  const { department: departmentData, action, onComplete } = payload;

  if (!departmentData) return;

  isDepartmentSubmitting.value = true;

  try {
    const preparedDepartment = prepareDepartmentForSave(departmentData);

    const savedDepartment = await saveDepartment(preparedDepartment);
    handleSaveSuccess(savedDepartment, action);
    onComplete?.();
  } catch (error: unknown) {
    handleSaveError(error);
  } finally {
    isDepartmentSubmitting.value = false;
  }
};

// Lifecycle
onMounted(async () => {
  initClusterId();
  try {
    await loadDepartment(clusterId.value);
  } catch (err) {
    console.log(err);
    appStore.setFallback(true);
  } finally {
    appStore.setPageLoading(false);
    initialLoading.value = false;
  }
});

onUnmounted(() => {
  appStore.setPageTitleOverride(null);
});

const clearFormState = async (): Promise<void> => {
  initialLoading.value = true;
  department.value = null;
  nodePoolTotals.value = {};
  await nextTick(); // Ensure form unmounts before reloading
  initClusterId();
  await loadDepartment(clusterId.value);
  initialLoading.value = false;
};

// Reset form when tour timestamp changes (for Help menu tour start)
watch(
  () => route.query.t,
  async (newVal, oldVal) => {
    if (newVal && newVal !== oldVal) {
      await clearFormState();
    }
  },
);
</script>
