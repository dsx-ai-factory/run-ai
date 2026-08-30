<template>
  <project-edit-form-v2
    v-if="!initialLoading && project"
    :key="formKey"
    :project="project"
    :department="department"
    :is-new-project="isNewProject"
    :is-project-submitting="isProjectSubmitting"
    :node-pools-minimal="nodePoolsMinimal"
    @update-scope="updateScope"
    @on-cancel="onCancel"
    @on-save="onSave"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";

// cmps
import { ProjectEditFormV2 } from "@/components/org-unit/v2/project/project-edit-form-v2";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";

// utils
import { alertUtil } from "@/utils/alert.util";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { deepCopy } from "@/utils/common.util";

// models
import type { INodePoolsNameAndId } from "@/models/node-pool.model";
import type { IEmptyProjectModelConfig } from "@/models/org-unit.model";
import type { Department, Project, ProjectCreationRequest, Resources } from "@/swagger-models/org-unit-service-client";
import { Units } from "@/swagger-models/org-unit-service-client";
import { type NodepoolMinimal, NodepoolSortFilterFields } from "@/swagger-models/cluster-service-client";
import { ErrorAlert } from "@/utils/error-alert.util";
import { DEFAULT_NODE_POOL_NAME } from "@/models/resource.model";
import { ESaveOrgUnitAction, type ISaveOrgUnitPayload } from "@/models/org-unit-forms.model";
import { filterUtil } from "@/utils/filter.util/filter.util";

const route = useRoute();
const router = useRouter();
const $q = useQuasar();
const appStore = useAppStore();
const authStore = useAuthStore();
const settingStore = useSettingStore();
const clusterStore = useClusterStore();

// null means "inherit from department"; the form always pre-fills a display value, so we
// track both the raw API value and what was displayed to detect whether the user changed it.
let rawProjectDefaultNodePools: string[] | null = null;
let displayedDefaultNodePools: string[] | null = null;

// State
const initialLoading = ref(true);
const loading = ref(false);
const isProjectSubmitting = ref(false);
const department = ref<Department | null>(null);
const project = ref<Project | ProjectCreationRequest | null>(null);
const clusterId = ref("");
const nodePoolsMinimal = ref<NodepoolMinimal[]>([]);

// Computed
const formKey = computed(() => route.query.t?.toString() || "default");
const isNewProject = computed(() => {
  if (project.value && (project.value as Project)?.id) {
    return false;
  }

  return route.name === PROJECT_ROUTE_NAMES.PROJECT_NEW;
});

const projectId = computed(() => route.params.id as string);
const isCpuEnabled = computed(() => settingStore.isCPUResourcesQuotaEnabled);

// Methods
const initClusterId = (): void => {
  if (route.query.clusterId) {
    clusterId.value = route.query.clusterId.toString();
  }
};

const getNodePoolsNamesAndIds = (): INodePoolsNameAndId[] => {
  return (
    department.value?.resources.map((resources) => {
      return {
        id: orgUnitUtil.getNodePoolIdByResource(resources),
        name: orgUnitUtil.getNodePoolNameByResource(resources),
      };
    }) || []
  );
};

const getEmptyProjectModelConfig = (clusterId: string): IEmptyProjectModelConfig => {
  const nodePoolsNamesAndIds = getNodePoolsNamesAndIds();

  // Inherit defaultNodePools from department, falling back to all available node pools
  // (when department.defaultNodePools is null, it means "use all available pools")
  const defaultNodePools = department.value?.defaultNodePools?.length
    ? [...department.value.defaultNodePools]
    : nodePoolsNamesAndIds.map((np) => np.name);

  return {
    nodePoolsNamesAndIds,
    clusterId: clusterId,
    isCpuEnabled: isCpuEnabled.value,
    parentId: department.value?.id || "",
    defaultNodePools: defaultNodePools.length ? defaultNodePools : [DEFAULT_NODE_POOL_NAME],
    memoryUnits: Units.Mb,
  };
};

const handleUpdateScopeError = (error: unknown): void => {
  const errorAlert = new ErrorAlert({
    generalMessage: `Failed to update project scope`,
  });
  $q.notify(errorAlert.getNotification(error));
};

/**
 * Initialize project with default values for UI display.
 * When defaultNodePools is null/empty (meaning "inherit"), we resolve it using priority:
 * 1. project.effective?.defaultNodePools (resolved inherited value from backend for existing projects)
 * 2. department.defaultNodePools (for new projects or fallback)
 * 3. [DEFAULT_NODE_POOL_NAME] (ultimate fallback)
 */
const initializeProjectNodePools = (projectData: Project | ProjectCreationRequest): Project | ProjectCreationRequest => {
  const initialized = deepCopy(projectData);
  if (!initialized.defaultNodePools?.length) {
    // For existing projects, use the backend's resolved effective value
    const effectiveNodePools = (projectData as Project).effective?.defaultNodePools;
    if (effectiveNodePools?.length) {
      initialized.defaultNodePools = [...effectiveNodePools];
    } else if (department.value?.defaultNodePools?.length) {
      // For new projects or when effective is not available, use department's setting
      initialized.defaultNodePools = [...department.value.defaultNodePools];
    } else {
      // Ultimate fallback
      initialized.defaultNodePools = [DEFAULT_NODE_POOL_NAME];
    }
  }
  displayedDefaultNodePools = initialized.defaultNodePools?.length ? [...initialized.defaultNodePools] : null;
  return initialized;
};

const updateScope = async (scope: { clusterId: string; departmentId?: string }): Promise<void> => {
  await loadNewProject(scope.clusterId, scope.departmentId);
  if (scope.clusterId) {
    await loadNodepoolsSchedulingCapabilities(scope.clusterId);
  }
};

const loadDefaultDepartment = async (clusterId: string): Promise<void> => {
  try {
    loading.value = true;
    department.value = await orgUnitService.getDefaultDepartment(clusterId);
  } catch (err) {
    handleUpdateScopeError(err);
  } finally {
    loading.value = false;
  }
};

const loadDepartment = async (departmentId: string): Promise<void> => {
  try {
    loading.value = true;
    department.value = await orgUnitService.getDepartment(departmentId);
  } catch (err) {
    handleUpdateScopeError(err);
  } finally {
    loading.value = false;
  }
};

const loadNewProject = async (clusterId: string, parentId?: string): Promise<void> => {
  if (parentId) {
    await loadDepartment(parentId);
  } else if (clusterId) {
    await loadDefaultDepartment(clusterId);
  }
  rawProjectDefaultNodePools = null;
  const emptyModelConfig = getEmptyProjectModelConfig(clusterId);
  const emptyProject = orgUnitUtil.getEmptyProjectModel(emptyModelConfig);
  project.value = initializeProjectNodePools(emptyProject);
};

const loadExistingProject = async (): Promise<void> => {
  const loadedProject: Project = await orgUnitService.getProject(projectId.value);
  rawProjectDefaultNodePools = loadedProject.defaultNodePools?.length ? [...loadedProject.defaultNodePools] : null;
  project.value = initializeProjectNodePools(loadedProject);
  appStore.setPageTitleOverride(`Edit project: ${loadedProject.name}`);
  if (loadedProject.parentId) {
    await loadDepartment(loadedProject.parentId);
  }
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

const loadProject = async (clusterIdParam: string, parentId?: string): Promise<void> => {
  if (isNewProject.value) {
    await loadNewProject(clusterIdParam, parentId);
    if (clusterIdParam) {
      await loadNodepoolsSchedulingCapabilities(clusterIdParam);
    }
  } else if (projectId.value) {
    await loadExistingProject();
    // For existing projects, use the project's clusterId
    if (project.value?.clusterId) {
      await loadNodepoolsSchedulingCapabilities(project.value.clusterId);
    }
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
 * When a department has a limited CPU/memory quota, null (unlimited) deserved must become 0.
 * The backend interprets null as unlimited (-1), which exceeds any finite department quota.
 */
const adjustUnlimitedCpuByDepartment = (resources: Resources[]): Resources[] => {
  if (!department.value || !isCpuEnabled.value) return resources;

  return resources.map((resource) => {
    const deptResource = department.value?.resources.find((r) => r.nodePool?.name === resource.nodePool?.name);
    if (!deptResource) return resource;

    const isDeptQuotaLimited = (deserved: number | null | undefined) => deserved != null;

    return {
      ...resource,
      cpu:
        resource.cpu && resource.cpu.deserved === null && isDeptQuotaLimited(deptResource.cpu?.deserved)
          ? { ...resource.cpu, deserved: 0 }
          : resource.cpu,
      memory:
        resource.memory && resource.memory.deserved === null && isDeptQuotaLimited(deptResource.memory?.deserved)
          ? { ...resource.memory, deserved: 0 }
          : resource.memory,
    };
  });
};

/**
 * Prepares project data for API submission by applying transformations
 */
const prepareProjectForSave = (projectData: Project | ProjectCreationRequest): Project | ProjectCreationRequest => {
  const prepared = deepCopy(projectData) as Project | ProjectCreationRequest;

  // Transform resources
  if (prepared.resources?.length) {
    // Default null (unlimited) CPU/memory to 0 when department has a finite quota
    prepared.resources = adjustUnlimitedCpuByDepartment(prepared.resources as Resources[]);
    // Align CPU/Memory limits with GPU over-quota state and reset over-quota weights to deserved
    prepared.resources = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(prepared.resources as Resources[]);
    // Ensure limits are not less than deserved to prevent API validation errors
    prepared.resources = orgUnitUtil.ensureLimitsNotBelowDeserved(prepared.resources as Resources[]);
    if (!settingStore.isOverQuotaPriorityEnabled) {
      // Reset over-quota weights to deserved if priority feature is disabled
      prepared.resources = orgUnitUtil.updateOverQuotaWeightToDeserved(prepared.resources as Resources[]);
    }
  }

  // Update scheduling rules
  if (department.value?.schedulingRules && prepared.schedulingRules) {
    prepared.schedulingRules = orgUnitUtil.updateSchedulingRules(
      prepared.schedulingRules,
      department.value.schedulingRules,
    );
  }

  // Inherit parent department's node types for new projects
  if (isNewProject.value && department.value?.nodeTypes) {
    const parentNT = department.value.nodeTypes;
    if (parentNT.training?.length || parentNT.workspace?.length) {
      prepared.nodeTypes = deepCopy(parentNT);
    }
  }

  // Send null (inherit) only when the user hasn't changed the pre-filled display value.
  const poolsUnchanged =
    prepared.defaultNodePools?.length === displayedDefaultNodePools?.length &&
    (prepared.defaultNodePools ?? []).every((v, i) => v === displayedDefaultNodePools?.[i]);
  prepared.defaultNodePools = poolsUnchanged ? rawProjectDefaultNodePools : prepared.defaultNodePools || null;

  return prepared;
};

/**
 * Persists project to API (create or update)
 */
const saveProject = async (projectData: Project | ProjectCreationRequest): Promise<Project> => {
  if (isNewProject.value) {
    const created = await orgUnitService.createProject(projectData as ProjectCreationRequest);
    await authStore.loadUserOrgUnits();
    return created;
  } else {
    const updateRequest = orgUnitUtil.getProjectUpdateRequest(projectData as Project);
    return await orgUnitService.updateProject((projectData as Project).id, updateRequest);
  }
};

/**
 * Handles post-save actions (notification, state update, navigation)
 */
const handleSaveSuccess = (savedProject: Project, action: ESaveOrgUnitAction): void => {
  const actionVerb = isNewProject.value ? "created" : "saved";
  $q.notify(alertUtil.getSuccess(`Project ${savedProject.name} ${actionVerb}`));

  rawProjectDefaultNodePools = savedProject.defaultNodePools?.length ? [...savedProject.defaultNodePools] : null;
  project.value = initializeProjectNodePools(savedProject);

  if (action === ESaveOrgUnitAction.Close || isNewProject.value) {
    const routeQuery = route.meta.backPageName;
    const routeName = routeQuery ? routeQuery.toString() : PROJECT_ROUTE_NAMES.PROJECT_INDEX;
    const shouldApplyClusterFilter = isNewProject.value && routeName === PROJECT_ROUTE_NAMES.PROJECT_INDEX;

    if (shouldApplyClusterFilter) {
      clusterStore.setSelectedClusterById(savedProject.clusterId);
    }

    router.push({
      name: routeName,
      query: { createdEntityId: savedProject.id },
    });
  }
};

/**
 * Handles save errors with appropriate notifications
 */
const handleSaveError = (error: unknown): void => {
  const errorAlert = new ErrorAlert({
    generalMessage: isNewProject.value
      ? ErrorAlert.failedCreateMessage("project")
      : ErrorAlert.failedUpdateMessage("project"),
  });
  $q.notify(errorAlert.getNotification(error));
};

// ==================== Main Save Handler ====================

const onSave = async (payload: ISaveOrgUnitPayload<Project | ProjectCreationRequest>) => {
  const { project: projectData, action, onComplete } = payload;

  if (!projectData) return;

  isProjectSubmitting.value = true;

  try {
    const preparedProject = prepareProjectForSave(projectData);
    const savedProject = await saveProject(preparedProject);

    // Refetch department to get updated projectsAggregatedResources (for excess quota banner)
    if (savedProject.parentId) {
      await loadDepartment(savedProject.parentId);
    }

    handleSaveSuccess(savedProject, action);
    onComplete?.();
  } catch (error: unknown) {
    handleSaveError(error);
  } finally {
    isProjectSubmitting.value = false;
  }
};

// Lifecycle
onMounted(async () => {
  initClusterId();
  try {
    await loadProject(clusterId.value);
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
  project.value = null;
  department.value = null;
  await nextTick(); // Ensure form unmounts before reloading
  initClusterId();
  await loadProject(clusterId.value);
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
