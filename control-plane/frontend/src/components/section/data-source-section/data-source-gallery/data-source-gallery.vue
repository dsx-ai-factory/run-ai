<template>
  <runai-form-card-section
    :loading="loading"
    empty-message="Looks like you don't have any setups yet..."
    :main-message="cardsMainTitle"
    entity-name="data source"
    :cards-list="dataSourceCards"
    :selected-cards-ids="selectedDataSourceCards"
    icon-name="ph-bold ph-database"
    search-name="data sources"
    :multi-select="multiSelect"
    @selected-card-changed="onSelectedDataSource"
    @create-new="$emit('create-new', $event)"
    :disabled="sectionDisabled"
    :disable-create-new="disableCreateButton"
    :sort-options="{ name: true, creation: true, recentUsage: true }"
    default-sort-option="recentUsage"
  >
    <template #add-new>
      <q-btn
        flat
        color="primary"
        :disabled="sectionDisabled || disableCreateButton"
        aid="data-source-gallery-add-new-btn"
      >
        <q-icon left name="ph-bold ph-plus" size="20px"></q-icon>
        <div>{{ isDataVolumeEnabled ? "new data or storage" : "new data source" }}</div>
        <data-source-dropdown
          @data-source-selected="$emit('create-new', $event)"
          :allowed-options="allowedCreatedTypes"
          :include-data-volume="isDataVolumeEnabled"
        />
      </q-btn>
    </template>
  </runai-form-card-section>
  <container-path-override
    v-if="showContainerPathOverride"
    :policy-rules="policyRules?.rules?.assets?.datasources"
    :disabled="sectionDisabled"
    :data-sources-container-path-data="dataSourcesContainerPathData"
    @update-container-path-override="updateContainerPathOverride"
    @is-section-invalid="$emit('is-section-invalid', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type {
  AssetIdAndKind,
  DatasourceListResponseEntry,
  ComplianceInfoReason,
  DatasourceListResponseAssetSpec,
  PolicyInfo,
} from "@/swagger-models/assets-service-client";
import type {
  IContainerPathOverride,
  ISelectedDatasourcesContainerPathData,
} from "@/components/section/compute-resource-section/container-path-override/container-path-override.model";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { TCardCmpName, type ICardListItem } from "@/components/common/runai-card-list";
import type { ICardState } from "./data-source-gallery.model";
import {
  DATA_SOURCE_KIND_TO_CONTAINER_KEY,
  DATA_SOURCE_KIND_TO_SPEC_KEY,
  DATA_SOURCE_TYPE,
  type TDataSourceKinds,
} from "@/models/data-source.model";
import { EWorkloadFormType } from "@/models/workload.model";

// stores
import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";

// components
import { DataSourceDropdown } from "@/components/data-source/data-source-dropdown";
import { RunaiFormCardSection } from "@/components/common";
import { ContainerPathOverride } from "@/components/section/compute-resource-section/container-path-override";

// constants
import { MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH } from "@/common/version.constant";
import { useSettingStore } from "@/stores/setting.store";

// utils
import { inferenceDataSourceRestrictionsUtil } from "@/utils/inference-data-source-restrictions.util/inference-data-source-restrictions.util";
import { dataSourceUtil } from "@/utils/data-source.util";

export default defineComponent({
  name: "data-source-gallery",
  components: {
    DataSourceDropdown,
    RunaiFormCardSection,
    ContainerPathOverride,
  },
  emits: ["datasource-changed", "create-new", "is-section-invalid"],
  props: {
    entityType: {
      type: String as PropType<EWorkloadFormType>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    /**
     * The data sources that are shown in the gallery.
     */
    dataSources: {
      type: Array as PropType<Array<DatasourceListResponseEntry>>,
      required: true,
    },
    /**
     * The data sources that are allowed to create for the workload type.
     * This is used to filter the data sources that are shown in create new data source button.
     * If you want to filter the data sources cards from the gallery, you need to filter the dataSources prop.
     */
    allowedCreatedTypes: {
      type: Set as PropType<Set<TDataSourceKinds>>,
      required: true,
    },
    selectedDataSources: {
      type: Array as PropType<Array<AssetIdAndKind>>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    cardsMainTitle: {
      type: String as PropType<string>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    createdDataSourceId: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PolicyInfo | null>,
      required: false,
    },
    containerPathOverrideAllowed: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    multiSelect: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
    };
  },
  computed: {
    isDataVolumeEnabled(): boolean {
      return this.settingStore.isDataVolumeEnabled;
    },
    disableCreateButton(): boolean {
      return !this.permissionStore.hasPermission(
        [
          ResourceType.HostPathAssets,
          ResourceType.PvcAssets,
          ResourceType.CmVolumeAssets,
          ResourceType.S3Assets,
          ResourceType.NfsAssets,
          ResourceType.GitAssets,
          ResourceType.SecretVolumeAssets,
        ],
        Action.Create,
      );
    },
    lockedCards() {
      return this.dataSourceCards.filter((card) => card.locked);
    },
    dataSourceCards(): Array<ICardListItem> {
      return this.dataSources.map((dataSource: DatasourceListResponseEntry) => {
        let cardState: ICardState = {
          locked: false,
          disabled: false,
          tooltip: "",
          showDisabledInfo: false,
        };

        const name = dataSource.meta.name;
        const type = dataSourceUtil.isTDataSourceKinds(dataSource.meta.kind)
          ? DATA_SOURCE_TYPE[dataSource.meta.kind]
          : "";
        cardState = this.getCardStateFromImposed(dataSource, cardState);
        cardState = this.getCardStateFromCompliance(dataSource, cardState);
        cardState = this.getCardStateForInferenceSupport(dataSource, cardState);

        return {
          id: dataSource.meta.id,
          cardName: TCardCmpName.DATA_SOURCE,
          data: dataSource,
          searchValues: [name, type].filter((val) => !!val),
          locked: cardState.locked,
          disabled: cardState.disabled,
          tooltip: cardState.tooltip,
          showDisabledInfo: cardState.showDisabledInfo,
          sortInfo: {
            name: dataSource.meta.name,
            createdAt: dataSource.meta.createdAt,
            recentUsage: dataSource.usageTimes?.lastUsedByWorkload,
          },
        };
      });
    },
    dataSourceMap(): Record<string, DatasourceListResponseEntry> {
      return this.dataSources.reduce(
        (acc: Record<string, DatasourceListResponseEntry>, dataSource: DatasourceListResponseEntry) => {
          acc[dataSource.meta.id] = dataSource;
          return acc;
        },
        {},
      );
    },
    selectedDataSourceCards(): Array<string> {
      return this.selectedDataSources.map((dataSource: AssetIdAndKind) => dataSource.id);
    },
    isClusterSupportedNfsAndHostPath(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH);
    },
    dataSourcesContainerPathData(): ISelectedDatasourcesContainerPathData[] {
      return (
        this.selectedDataSources.map((ds) => {
          const selected = this.dataSourceMap[ds.id];

          return {
            id: ds.id,
            name: selected.meta.name,
            kind: selected.meta.kind,
            overrides: ds.overrides || undefined,
            origin: this.getDataSourceContainerPath(selected),
          };
        }) || []
      );
    },
    showContainerPathOverride(): boolean {
      return !!this.dataSourcesContainerPathData.length && this.containerPathOverrideAllowed;
    },
  },
  methods: {
    getCardStateForInferenceSupport(dataSource: DatasourceListResponseEntry, cardState: ICardState): ICardState {
      if (this.entityType !== EWorkloadFormType.Inference) return cardState;

      const restrictionInfo = inferenceDataSourceRestrictionsUtil.getDataSourceRestrictionInfo(
        dataSource.meta.kind,
        this.clusterId,
        false,
      );

      if (!restrictionInfo.disabled || !restrictionInfo.disabledInfo) return cardState;

      return { ...cardState, locked: false, disabled: true, tooltip: restrictionInfo.disabledInfo };
    },
    getCardStateFromCompliance(dataSource: DatasourceListResponseEntry, cardState: ICardState): ICardState {
      const currCardState: ICardState = { ...cardState };
      if (dataSource.compliance && !dataSource.compliance?.compliance) {
        currCardState.disabled = true;
        currCardState.locked = false;
        // if it does not comply due to policy vs if it does not comply for some other reason
        if (dataSource.compliance?.reason?.some((reason: ComplianceInfoReason) => reason.field)) {
          currCardState.showDisabledInfo = true;
          currCardState.tooltip =
            "This setup can't be used because it doesn't comply with the policy your administrator set.";
        } else if (dataSource.compliance?.reason?.length) {
          if (dataSource.meta.id !== this.createdDataSourceId) {
            currCardState.tooltip = dataSource.compliance.reason[0].details;
          } else {
            currCardState.tooltip =
              "This setup is being created. If it isn't created after a few seconds, contact your administrator.";
          }
        }
      }

      return currCardState;
    },
    getCardStateFromImposed(dataSource: DatasourceListResponseEntry, cardState: ICardState): ICardState {
      const currCardState: ICardState = { ...cardState };
      if (dataSource.compliance && dataSource.compliance?.imposed) {
        currCardState.locked = true;
        currCardState.disabled = true;
        currCardState.tooltip = "This setup must stay selected based on the policy your administrator set.";
      }

      return currCardState;
    },
    onSelectedDataSource(selectedDataSourceIds: Array<string>): void {
      const selectedMap = new Map<string, AssetIdAndKind>(
        this.selectedDataSources.map((selected) => [selected.id, selected]),
      );

      const selectedDataSourceModel: Array<AssetIdAndKind> = selectedDataSourceIds.map((dataSourceId: string) => {
        return {
          id: dataSourceId,
          kind: this.dataSourceMap[dataSourceId].meta.kind,
          overrides: selectedMap.get(dataSourceId)?.overrides || undefined,
        };
      });
      this.$emit("datasource-changed", selectedDataSourceModel);
    },
    updateContainerPathOverride(cpOverride: IContainerPathOverride) {
      const updatedSelected: AssetIdAndKind[] = this.selectedDataSources.map((ds) => {
        if (ds.id !== cpOverride.dataSourceId) return ds;
        if (typeof cpOverride.containerPathOverride !== "string") {
          return {
            ...ds,
            overrides: undefined,
          };
        }
        return {
          ...ds,
          overrides: {
            containerPath: cpOverride.containerPathOverride,
          },
        };
      });

      this.$emit("datasource-changed", updatedSelected);
    },
    getDataSourceContainerPath(dataSource: DatasourceListResponseEntry): string {
      const specKey = DATA_SOURCE_KIND_TO_SPEC_KEY[
        dataSource.meta.kind as TDataSourceKinds
      ] as keyof DatasourceListResponseAssetSpec;

      const spec = dataSource.spec[specKey];

      const containerPathKey = DATA_SOURCE_KIND_TO_CONTAINER_KEY[
        dataSource.meta.kind as TDataSourceKinds
      ] as keyof (typeof dataSource.spec)[typeof specKey];

      const originalContainerPath = spec?.[containerPathKey] || "";
      return originalContainerPath;
    },
  },
});
</script>

<style lang="scss" scoped></style>
