/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import type { THybridSelectedTemplates } from "@/models/workload-hybrid.model";
import {
  type IUIEnvironmentFields,
  type IUISuperWorkloadSpec,
  type IUISuperWorkloadCreationRequest,
  EWorkloadTypeSupportedInAssetsService,
} from "@/models/workload.model";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import type { ScopeType, TemplateAssetsIds, UnifiedWorkloadRequest } from "@/swagger-models/workloads-client";
import { alertUtil } from "@/utils/alert.util";
import { deepCopy } from "@/utils/common.util";
import { workloadHybridUtil as whUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { useQuasar } from "quasar";
import { computed, type Ref, ref } from "vue";
import { EHybridFormTypes } from "./helpers/use-hybrid-templates.mapper";
import { useHybridTemplatesMapper } from "./helpers/use-hybrid-templates.mapper";
import { useRoute, useRouter } from "vue-router";
import { scopeUtil } from "@/utils/scope.util";
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import { computeResourceService } from "@/services/control-plane/compute-resource.service/compute-resource.service";
import type { IAssetsFilter } from "@/models/filter.model";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";

type TStorageType = WorkloadCreationDrawerEntities.ModelStorage | WorkloadCreationDrawerEntities.DataSource;

export function useHybridTemplates(
  workloadType: EHybridFormTypes,
  initialModel: IUISuperWorkloadCreationRequest = whUtil.getEmptyUISuperWorkloadModel(),
  isTemplate = false,
) {
  const $q = useQuasar();
  const $router = useRouter();
  const $route = useRoute();

  const hybridWorkloadModel: Ref<IUISuperWorkloadCreationRequest> = ref(initialModel);
  const selectedTemplates: Ref<THybridSelectedTemplates> = ref(whUtil.getWorkloadHybridTemplatesSetup());
  const drawerDataEntity: Ref<WorkloadCreationDrawerEntities | null> = ref(null);
  const isDrawerOpen: Ref<boolean> = ref(false);
  const applyingSelection: Ref<boolean> = ref(false);
  const recentlyCreatedEntityId: Ref<string | null> = ref(null);

  const closeDrawer = (): void => {
    isDrawerOpen.value = false;
    drawerDataEntity.value = null;
  };

  const openDrawer = (entity: WorkloadCreationDrawerEntities): void => {
    drawerDataEntity.value = entity;
    isDrawerOpen.value = true;
  };

  const displayEnvironmentsList = (): void => {
    openDrawer(WorkloadCreationDrawerEntities.Environment);
  };
  const displayComputesList = (): void => {
    openDrawer(WorkloadCreationDrawerEntities.Compute);
  };
  const displayDataSourcesList = (): void => {
    openDrawer(WorkloadCreationDrawerEntities.DataSource);
  };

  const displayModelStorageList = (): void => {
    openDrawer(WorkloadCreationDrawerEntities.ModelStorage);
  };

  const updateEnvironmentRenderKey = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].renderKey++;
  };

  const updateComputeRenderKey = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].renderKey++;
  };

  const reRenderSectionEntirely = (section: WorkloadCreationDrawerEntities): void => {
    selectedTemplates.value[section].sectionKey++;
  };

  const updateWorkloadName = (name: string): void => {
    hybridWorkloadModel.value.name = name;
  };

  const getUnifiedHybridWorkload = useHybridTemplatesMapper[workloadType].getUnified;

  const combineModels = async (
    assetId: string | string[],
    dataVolumeId?: string | string[],
    forcedEntity?: WorkloadCreationDrawerEntities, // used for setup purposes
  ): Promise<void> => {
    const requestOptions: UnifiedWorkloadRequest = {
      clusterId: hybridWorkloadModel.value.clusterId,
    };
    if (!isTemplate) requestOptions.projectId = hybridWorkloadModel.value.projectId;

    Array.isArray(assetId) ? (requestOptions.assetIds = assetId) : (requestOptions.assetId = assetId);
    if (dataVolumeId) {
      Array.isArray(dataVolumeId)
        ? (requestOptions.dataVolumeIds = dataVolumeId)
        : (requestOptions.dataVolumeIds = [dataVolumeId]);
    }

    const modelWithItemDetailsFlattened = await getUnifiedHybridWorkload(requestOptions);

    const uiModel = useHybridTemplatesMapper[workloadType].getUIModel(modelWithItemDetailsFlattened);

    switch (forcedEntity || drawerDataEntity.value) {
      case WorkloadCreationDrawerEntities.Environment: {
        const envFields = useHybridTemplatesMapper[workloadType].getRelevantEnvironmentFields(uiModel);

        const combineWorkloadModelWithEnvironmentDetails =
          useHybridTemplatesMapper[workloadType].combineWorkloadModelWithEnvironmentDetails ||
          whUtil.combineWorkloadModelWithEnvironmentDetails;
        hybridWorkloadModel.value = combineWorkloadModelWithEnvironmentDetails(hybridWorkloadModel.value, envFields);
        selectedTemplates.value.environment.defaultData = envFields;
        updateEnvironmentRenderKey();
        break;
      }
      case WorkloadCreationDrawerEntities.Compute: {
        const getRelevantComputeFields =
          useHybridTemplatesMapper[workloadType].getRelevantComputeFields || whUtil.getRelevantComputeFields;
        const computeFields = getRelevantComputeFields(uiModel);
        const combineWorkloadModelWithComputeDetails =
          useHybridTemplatesMapper[workloadType].combineWorkloadModelWithComputeDetails ||
          whUtil.combineWorkloadModelWithComputeDetails;
        hybridWorkloadModel.value = combineWorkloadModelWithComputeDetails(hybridWorkloadModel.value, computeFields);
        selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].defaultData = computeFields;
        updateComputeRenderKey();
        break;
      }
      case WorkloadCreationDrawerEntities.DataSource: {
        const getRelevantStorageFields =
          useHybridTemplatesMapper[workloadType].getRelevantStorageFields || whUtil.getRelevantStorageFields;
        const storageFields = getRelevantStorageFields(uiModel) || {};

        selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].defaultData = storageFields;
        const combineWorkloadModelWithStorageDetails =
          useHybridTemplatesMapper[workloadType].combineWorkloadModelWithStorageDetails ||
          whUtil.combineWorkloadModelWithStorageDetails;
        hybridWorkloadModel.value = combineWorkloadModelWithStorageDetails(hybridWorkloadModel.value, storageFields);
        break;
      }
      case WorkloadCreationDrawerEntities.ModelStorage: {
        clearModelStorageTemplate();
        const storageFields = whUtil.getRelevantStorageFields(uiModel);
        selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].defaultData = storageFields;
        hybridWorkloadModel.value.spec.storage = {}; // single storage only
        hybridWorkloadModel.value = whUtil.combineWorkloadModelWithStorageDetails(
          hybridWorkloadModel.value,
          storageFields,
        );
        break;
      }
    }
  };

  const handleSelection = async (itemIdAndName: { id: string; name: string }): Promise<void> => {
    if (!drawerDataEntity.value || applyingSelection.value) return;
    applyingSelection.value = true;
    try {
      await combineModels(itemIdAndName.id);
      selectedTemplates.value[drawerDataEntity.value].name = itemIdAndName.name;
      selectedTemplates.value[drawerDataEntity.value].id = itemIdAndName.id;
      closeDrawer();
    } catch (e) {
      $q.notify(alertUtil.getError("Failed to select asset template"));
      console.error("Failed to select asset template:\n", e);
    } finally {
      applyingSelection.value = false;
    }
  };

  const handleMultiSelection = async (selectedItems: { id: string; name: string; kind: AssetKind }[]): Promise<void> => {
    if (!drawerDataEntity.value || applyingSelection.value) return;
    applyingSelection.value = true;
    try {
      await combineModels(
        selectedItems.filter((item) => item.kind !== AssetKind.DataVolume).map((item) => item.id),
        selectedItems.filter((item) => item.kind === AssetKind.DataVolume).map((item) => item.id),
      );

      closeDrawer();
    } catch (e) {
      $q.notify(alertUtil.getError("Failed to select asset template"));
      console.error("Failed to select asset template:\n", e);
    } finally {
      applyingSelection.value = false;
    }
  };

  const resetEnvironmentToDefault = (): void => {
    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].defaultData) return;

    const combineWorkloadModelWithEnvironmentDetails =
      useHybridTemplatesMapper[workloadType].combineWorkloadModelWithEnvironmentDetails ||
      whUtil.combineWorkloadModelWithEnvironmentDetails;
    hybridWorkloadModel.value = combineWorkloadModelWithEnvironmentDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].defaultData,
    );
    updateEnvironmentRenderKey();
  };

  const resetComputeToDefault = (): void => {
    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].defaultData) return;

    const combineWorkloadModelWithComputeDetails =
      useHybridTemplatesMapper[workloadType].combineWorkloadModelWithComputeDetails ||
      whUtil.combineWorkloadModelWithComputeDetails;
    hybridWorkloadModel.value = combineWorkloadModelWithComputeDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].defaultData,
    );
    updateComputeRenderKey();
  };

  const resetDataSourcesToDefault = (): void => {
    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData) return;

    hybridWorkloadModel.value = whUtil.combineWorkloadModelWithStorageDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData,
    );
  };

  const resetModelStorageToDefault = (): void => {
    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].defaultData) return;

    hybridWorkloadModel.value = whUtil.combineWorkloadModelWithStorageDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].defaultData,
    );
  };

  const clearEnvironmentTemplate = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].name = "";
    selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].id = "";
    selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].defaultData = null;

    if (isTemplate && selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked) {
      selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked = false;
    }

    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].clearedState) return;

    const combineWorkloadModelWithEnvironmentDetails =
      useHybridTemplatesMapper[workloadType].combineWorkloadModelWithEnvironmentDetails ||
      whUtil.combineWorkloadModelWithEnvironmentDetails;

    hybridWorkloadModel.value = combineWorkloadModelWithEnvironmentDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].clearedState,
    );

    updateEnvironmentRenderKey();
    reRenderSectionEntirely(WorkloadCreationDrawerEntities.Environment);
  };

  const clearComputeTemplate = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].name = "";
    selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id = "";
    selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].defaultData = null;

    if (isTemplate && selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked) {
      selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked = false;
    }

    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].clearedState) return;

    const combineWorkloadModelWithComputeDetails =
      useHybridTemplatesMapper[workloadType].combineWorkloadModelWithComputeDetails ||
      whUtil.combineWorkloadModelWithComputeDetails;

    hybridWorkloadModel.value = combineWorkloadModelWithComputeDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].clearedState,
    );
    updateComputeRenderKey();
    reRenderSectionEntirely(WorkloadCreationDrawerEntities.Compute);
  };

  const clearDataSourceTemplate = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].name = "";
    selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData = null;

    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].clearedState) return;

    hybridWorkloadModel.value = whUtil.combineWorkloadModelWithStorageDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].clearedState,
    );
  };

  const clearModelStorageTemplate = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].name = "";
    selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].defaultData = null;

    if (!selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].clearedState) return;

    hybridWorkloadModel.value = whUtil.combineWorkloadModelWithStorageDetails(
      hybridWorkloadModel.value,
      selectedTemplates.value[WorkloadCreationDrawerEntities.ModelStorage].clearedState,
    );
  };

  const createSectionsClearedStates = (): void => {
    const base = deepCopy(hybridWorkloadModel.value);
    selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].clearedState =
      useHybridTemplatesMapper[workloadType].getRelevantEnvironmentFields(base);
    selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].clearedState = whUtil.getRelevantComputeFields(base);
    selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].clearedState =
      whUtil.getRelevantStorageFields(base);
  };

  const handleCreatedAssets = (kind?: string, createdEntityId?: string, storageKind?: TStorageType): void => {
    if (!kind || !createdEntityId) return;

    recentlyCreatedEntityId.value = createdEntityId;
    switch (kind) {
      case AssetKind.Environment:
        displayEnvironmentsList();
        break;
      case AssetKind.Compute:
        displayComputesList();
        break;
      case AssetKind.HostPath:
      case AssetKind.Nfs:
      case AssetKind.Git:
      case AssetKind.S3:
      case AssetKind.Pvc:
      case AssetKind.ConfigMap:
      case AssetKind.SecretVolume:
        if (storageKind === WorkloadCreationDrawerEntities.ModelStorage) {
          displayModelStorageList();
        } else {
          displayDataSourcesList();
        }
        break;
      default:
        recentlyCreatedEntityId.value = null;
    }
  };

  const onCreateNewSetup = (type: string, additionalQuery = {}): void => {
    switch (drawerDataEntity.value) {
      case WorkloadCreationDrawerEntities.Environment:
        onCreateNewEnvironment(additionalQuery);
        break;
      case WorkloadCreationDrawerEntities.Compute:
        onCreateNewComputeResource(additionalQuery);
        break;
      case WorkloadCreationDrawerEntities.DataSource:
        onCreateNewDataSource(type, additionalQuery);
        break;
      case WorkloadCreationDrawerEntities.ModelStorage:
        onCreateNewDataSource(type, additionalQuery);
        break;
    }
  };

  const _getCreationRouteName = (assetType: "environment" | "compute" | string): string | undefined => {
    const routes = isTemplate
      ? useHybridTemplatesMapper[workloadType].templateCreationRoutes
      : useHybridTemplatesMapper[workloadType].creationRoutes;

    if (!routes) return undefined;

    if (assetType !== "environment" && assetType !== "compute") {
      return routes.datasource.pages[assetType];
    }
    return routes[assetType]?.name;
  };

  const _getCreationScopeQuery = (): Record<string, string> | undefined => {
    if (isTemplate) {
      const scopeModel = scopeUtil.getScopeModelFromTypeAndId(
        hybridWorkloadModel.value.scopeType as ScopeType,
        hybridWorkloadModel.value.scopeId as string,
      );

      return {
        ...Object.fromEntries(
          Object.entries(scopeModel)
            .filter(([_, value]) => value != null)
            .map(([key, value]) => [key, String(value)]),
        ),
      };
    } else {
      return {
        projectId: hybridWorkloadModel.value.projectId,
        scope: Scope.Project,
      };
    }
  };

  const onCreateNewEnvironment = (additionalQuery = {}): void => {
    const routeName = _getCreationRouteName("environment");
    let query = _getCreationScopeQuery();

    const queryFromMapper = isTemplate
      ? useHybridTemplatesMapper[workloadType].templateCreationRoutes?.environment?.query
      : useHybridTemplatesMapper[workloadType].creationRoutes?.environment?.query;

    if (queryFromMapper) {
      query = {
        ...queryFromMapper,
        ...query,
      };
    }

    if (!routeName || !query) return;

    const paramId = $route.params.id;

    $router.push({
      name: routeName,
      query: {
        ...query,
        ...additionalQuery,
      },
      // configure state ONLY if id exists in params
      state: paramId ? { appData: { paramId } } : undefined,
    });
  };

  const onCreateNewComputeResource = (additionalQuery = {}): void => {
    const routeName = _getCreationRouteName("compute");
    let query = _getCreationScopeQuery();

    const queryFromMapper = isTemplate
      ? useHybridTemplatesMapper[workloadType].templateCreationRoutes?.compute?.query
      : useHybridTemplatesMapper[workloadType].creationRoutes?.compute?.query;

    const paramId = $route.params.id;
    if (queryFromMapper) {
      query = {
        ...queryFromMapper,
        ...query,
      };
    }

    if (!routeName || !query) return;

    $router.push({
      name: routeName,
      query: {
        ...query,
        ...additionalQuery,
      },
      // configure state ONLY if id exists in params
      state: paramId ? { appData: { paramId } } : undefined,
    });
  };

  const onCreateNewDataSource = (selectedDataSourceType: string, additionalQuery = {}): void => {
    const routeName = _getCreationRouteName(selectedDataSourceType);
    if (!routeName) return;
    const query = _getCreationScopeQuery();

    const paramId = $route.params.id;

    if (!routeName || !query) return;

    $router.push({
      name: routeName,
      query: {
        fromPage: routeName,
        ...query,
        ...additionalQuery,
      },
      // configure state ONLY if id exists in params
      state: paramId ? { appData: { paramId } } : undefined,
    });
  };

  const toggleEnvironmentLinked = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked =
      !selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked;

    if (selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked) resetEnvironmentToDefault();
  };

  const toggleComputeLinked = (): void => {
    selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked =
      !selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked;

    if (selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked) resetComputeToDefault();
  };

  // NOTE: for setup purposes only (when we load a template with linked assets)
  const _linkedAssetSetup = async (
    entity: WorkloadCreationDrawerEntities.Compute | WorkloadCreationDrawerEntities.Environment,
    id?: string | null,
    name?: string | null,
  ) => {
    if (!id || !name || !entity) return;

    selectedTemplates.value[entity].isLinked = true;
    selectedTemplates.value[entity].id = id;
    selectedTemplates.value[entity].name = name;

    let defaultData: IUISuperWorkloadSpec["compute"] | IUIEnvironmentFields | null = null;
    if (entity === WorkloadCreationDrawerEntities.Environment) {
      defaultData = whUtil.getRelevantEnvironmentFields(hybridWorkloadModel.value);
    } else if (entity === WorkloadCreationDrawerEntities.Compute) {
      defaultData = whUtil.getRelevantComputeFields(hybridWorkloadModel.value);
    }

    selectedTemplates.value[entity].defaultData = defaultData;
  };

  // NOTE: only to be used for the initial setup of a template model which is based on an existing template (edit/clone)
  const setupLinkedEntities = async (entities: TemplateAssetsIds): Promise<void> => {
    if (!entities) return;

    try {
      const loadedAssets = await Promise.all([
        !!entities.environment && environmentService.getById(entities.environment),
        !!entities.compute && computeResourceService.getById(entities.compute),
      ]);
      if (loadedAssets) {
        const envSetupPromise = loadedAssets[0]
          ? _linkedAssetSetup(
              WorkloadCreationDrawerEntities.Environment,
              entities.environment,
              loadedAssets[0]?.meta?.name,
            )
          : Promise.resolve();

        const computeSetupPromise = loadedAssets[1]
          ? _linkedAssetSetup(WorkloadCreationDrawerEntities.Compute, entities.compute, loadedAssets[1]?.meta?.name)
          : Promise.resolve();

        await Promise.all([envSetupPromise, computeSetupPromise]);
      }
    } catch (error: unknown) {
      console.error("Failed to load linked assets:\n", error);
      $q.notify(alertUtil.getError("Failed to load linked assets"));
      // if for some reason we failed to load the assets,
      // we will ignore the linked assets and reset the asset templates state
      selectedTemplates.value.environment.isLinked = false;
      selectedTemplates.value.environment.id = "";
      selectedTemplates.value.environment.name = "";
      selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked = false;
      selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id = "";
      selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].name = "";
    }
  };

  // computed
  const selectedEnvironmentName = computed(() => {
    return selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].name;
  });

  const selectedComputeName = computed(() => {
    return selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].name;
  });

  const selectedDataSourceName = computed(() => {
    return selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].name;
  });

  const linkedEntitiesIds = computed((): TemplateAssetsIds => {
    return {
      environment: selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked
        ? selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].id
        : undefined,
      compute: selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked
        ? selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id
        : undefined,
    };
  });

  const baseDrawerFilters = computed((): IAssetsFilter => {
    const baseMapper = {
      [EHybridFormTypes.Workspace]: {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        isWorkspace: true,
      },
      [EHybridFormTypes.Training]: {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Training,
        isTraining: true,
      },
      [EHybridFormTypes.DistributedMaster]: {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Distributed,
        isDistributed: true,
      },
      [EHybridFormTypes.DistributedWorker]: {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Distributed,
        isDistributed: true,
      },
      [EHybridFormTypes.Inference]: {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
        isInference: true,
      },
      [EHybridFormTypes.DistributedInferenceLeader]: {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
        isInference: true,
      },
      [EHybridFormTypes.DistributedInferenceWorker]: {
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
        isInference: true,
      },
    };

    const baseFilter = {
      ...baseMapper[workloadType],
    };

    switch (hybridWorkloadModel.value.scopeType) {
      case Scope.Project:
        return {
          ...baseFilter,
          projectId: hybridWorkloadModel.value.scopeId ? +hybridWorkloadModel.value.scopeId : undefined,
          complyToProject: isTemplate ? undefined : +hybridWorkloadModel.value.projectId,
        };
      case Scope.Cluster:
        return {
          ...baseFilter,
          clusterUuid: hybridWorkloadModel.value.scopeId || hybridWorkloadModel.value.clusterId,
        };
      case Scope.Department:
        return {
          ...baseFilter,
          departmentId: hybridWorkloadModel.value.scopeId,
        };
      default:
        return {
          ...baseFilter,
          projectId: +hybridWorkloadModel.value.projectId,
          complyToProject: isTemplate ? undefined : +hybridWorkloadModel.value.projectId,
        };
    }
  });

  // meant to be used for scope change for workload-hybrid-templates forms (not workloads)
  const validateDrawerEntitiesForScopeChange = async (
    existingSecrets: string[],
    additionalFilters: IAssetsFilter = {},
  ): Promise<void> => {
    try {
      const listFilter = {
        ...baseDrawerFilters.value,
        ...additionalFilters,
      };

      // verify environment in use
      if (selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].id) {
        const envList = await environmentService.list(
          listFilter,
          selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].name,
        );
        const envExistsInScope =
          envList.findIndex(
            (env) => env.meta.id === selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].id,
          ) !== -1;
        if (!envExistsInScope) {
          selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].name = "";
          selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].id = "";
          selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].defaultData = null;
          selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked = false;
        }
      }
      // verify compute resource in use
      if (selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id) {
        const computeList = await computeResourceService.list(
          listFilter,
          selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].name,
        );
        const computeExistsInScope =
          computeList.findIndex(
            (compute) => compute.meta.id === selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id,
          ) !== -1;
        if (!computeExistsInScope) {
          selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].name = "";
          selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id = "";
          selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].defaultData = null;
          selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked = false;
        }
      }
      // verify data source in use
      // data sources such as existing PVCs, ConfigMaps, SecretsVolumes are unusable if their existing claim/configMap/Secrets do not exist in the new scope

      if (hybridWorkloadModel.value.spec.storage) {
        if (hybridWorkloadModel.value.spec.storage.pvc?.length) {
          const pvcList = await dataSourceService.listExistingPVCs(
            hybridWorkloadModel.value.scopeType as string,
            hybridWorkloadModel.value.scopeId as string,
            "all",
            true,
          );
          const existingPvcs = pvcList.map((pvc) => pvc.name);

          // check each pvc in the model to see if its claimName exists in the pvcList if not remove this PVC from model
          hybridWorkloadModel.value.spec.storage.pvc = hybridWorkloadModel.value.spec.storage.pvc.filter((pvc) => {
            if (pvc.claimName) {
              return existingPvcs.includes(pvc.claimName);
            }
            return true;
          });
        }
        if (hybridWorkloadModel.value.spec.storage.configMapVolume?.length) {
          const configMapList = await dataSourceService.listExistingConfigMaps(
            hybridWorkloadModel.value.scopeType as string,
            hybridWorkloadModel.value.scopeId as string,
            true,
          );
          const existingConfigMaps = configMapList.map((configMap) => configMap.name);

          // check each configMap in the model to see if its name exists in the configMapList if not remove this configMap from model
          hybridWorkloadModel.value.spec.storage.configMapVolume =
            hybridWorkloadModel.value.spec.storage.configMapVolume.filter((configMap) => {
              if (!configMap.configMap) return false;
              return existingConfigMaps.includes(configMap.configMap);
            });
        }
        if (hybridWorkloadModel.value.spec.storage.secretVolume?.length) {
          // check each secret in the model to see if its name exists in the secretList if not remove this secret from model
          hybridWorkloadModel.value.spec.storage.secretVolume =
            hybridWorkloadModel.value.spec.storage.secretVolume.filter((secret) => {
              if (!secret.secret) return false;
              return existingSecrets.includes(secret.secret);
            });
        }

        // based on the model lets clear the default data for the data source section
        if (selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData?.pvc?.length) {
          selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData.pvc = selectedTemplates.value[
            WorkloadCreationDrawerEntities.DataSource
          ].defaultData.pvc.filter((pvc) => {
            return hybridWorkloadModel.value.spec.storage?.pvc?.some((modelPvc) => modelPvc.claimName === pvc.claimName);
          });
        }
        if (selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData?.configMapVolume?.length) {
          selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData.configMapVolume =
            selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData.configMapVolume.filter(
              (configMap) => {
                return hybridWorkloadModel.value.spec.storage?.configMapVolume?.some(
                  (modelConfigMap) => modelConfigMap.configMap === configMap.configMap,
                );
              },
            );
        }
        if (selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData?.secretVolume?.length) {
          selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData.secretVolume =
            selectedTemplates.value[WorkloadCreationDrawerEntities.DataSource].defaultData.secretVolume.filter(
              (secretVol) => {
                return hybridWorkloadModel.value.spec.storage?.secretVolume?.some(
                  (modelSecret) => modelSecret.secret === secretVol.secret,
                );
              },
            );
        }
      }
    } catch (error: unknown) {
      console.error("Failed to check assets existence in scope:", error);
    }
  };

  return {
    // state
    isDrawerOpen,
    drawerDataEntity,
    applyingSelection,
    hybridWorkloadModel,
    selectedTemplates,
    recentlyCreatedEntityId,
    // methods
    closeDrawer,
    openDrawer,
    displayEnvironmentsList,
    displayComputesList,
    displayDataSourcesList,
    displayModelStorageList,
    handleSelection,
    handleMultiSelection,
    createSectionsClearedStates,
    clearComputeTemplate,
    clearDataSourceTemplate,
    clearEnvironmentTemplate,
    clearModelStorageTemplate,
    resetEnvironmentToDefault,
    resetComputeToDefault,
    resetDataSourcesToDefault,
    resetModelStorageToDefault,
    handleCreatedAssets,
    getUnifiedHybridWorkload,
    reRenderSectionEntirely,
    onCreateNewSetup,
    toggleEnvironmentLinked,
    toggleComputeLinked,
    setupLinkedEntities,
    updateWorkloadName,
    validateDrawerEntitiesForScopeChange,
    // computed
    selectedEnvironmentName,
    selectedComputeName,
    selectedDataSourceName,
    linkedEntitiesIds,
    baseDrawerFilters,
  };
}
