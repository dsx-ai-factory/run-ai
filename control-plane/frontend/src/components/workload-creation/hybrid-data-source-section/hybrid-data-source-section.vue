<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <hybrid-section-wrapper
    aid="hybrid-data-source-section"
    class="hybrid-data-source-section"
    :label="sectionLabel"
    @load-section-templates="$emit('load-section-templates')"
    :default-opened="defaultOpened"
    :section-disabled="sectionDisabled"
    :show-content-header="showContentHeader"
  >
    <template #subheader>
      <span>{{ dataSourceSummary }}</span>
    </template>

    <slot name="datasource-title">
      <div>Set the {{ dataSourceTitle }} your workload needs to access</div>
    </slot>
    <q-field
      ref="validationField"
      class="q-my-md form-hint no-padding"
      :model-value="isValidationError"
      :rules="validationRules"
      lazy-rules="ondemand"
    />
    <section class="row items-end no-wrap q-mt-md">
      <runai-tab variant="underline" :model-value="defaultTabsValue" :tabs="addTabItem" class="data-source-tabs q-mb-sm">
        <template #tab>
          <div @click.stop>
            <q-icon name="ph-bold ph-plus" size="20px" color="icon-default" />
            <q-tooltip anchor="top middle" self="bottom middle">
              {{
                cannotAddOneOffDataSources
                  ? "Your administrator's policy does not allow adding one-time data or storage."
                  : "Add a one-time setup for this workload only"
              }}
            </q-tooltip>
            <data-source-dropdown
              :allowed-options="allowedOneOffDataSources"
              @data-source-selected="addOneOffStorage"
              avoid-permissions-check
              show-workload-volume-option
            />
          </div>
        </template>
      </runai-tab>
      <runai-tab v-model="selectedId" outside-arrows :tabs="dataSourceTabItems" class="data-source-tabs q-mb-sm flex-1">
        <template #tab="{ tab }">
          <div class="row items-center no-wrap">
            <runai-svg-icon
              v-if="dataSourceTabsById[tab.name]?.icon"
              :name="dataSourceTabsById[tab.name]?.icon ?? ''"
              size="24px"
              class="q-mr-sm"
            />
            <span class="q-pr-sm text-weight-medium">{{ tab.label }}</span>
            <q-icon v-if="tab.errorIcon" class="q-mr-sm" name="ph-bold ph-warning-circle" color="red" size="16px" />
            <runai-btn
              v-if="!sectionDisabled"
              aid="remove-storage-btn"
              round
              color="grey-4"
              text-color="b-main"
              unelevated
              size="sm"
              icon="ph-bold ph-x"
              class="close-btn"
              @click.stop="removeStorage(tab.name, dataSourceTabsById[tab.name]?.label)"
              :disable="!dataSourceTabsById[tab.name]?.removeable || sectionDisabled"
            />
            <q-tooltip v-if="!dataSourceTabsById[tab.name]?.removeable">
              This setup must stay selected based on the policy your administrator set.
            </q-tooltip>
          </div>
        </template>
      </runai-tab>
    </section>

    <runai-tab-panels v-model="selectedId" keep-alive>
      <runai-tab-panel v-for="tab in tabs" :name="tab.id" :key="tab.id">
        <q-tooltip v-if="!tab.removeable" anchor="center middle">
          This can't be modified based on the policy your administrator set.
        </q-tooltip>
        <host-path-instance-section
          v-if="tab.label === AssetKind.HostPath && hostPathInstancesMapById[tab.id]"
          :key="hostPathInstancesMapById[tab.id]?.id"
          :model-value="hostPathInstancesMapById[tab.id]"
          @update:model-value="updateHostPath"
          :policy-rules="policyRules?.hostPath?.attributes"
          :disable="!tab.removeable || sectionDisabled"
          :disable-info="disableFields?.hostPath"
        />
        <nfs-instance-section
          v-else-if="tab.label === AssetKind.Nfs && nfsInstancesMapById[tab.id]"
          :key="nfsInstancesMapById[tab.id]?.id"
          :model-value="nfsInstancesMapById[tab.id]"
          @update:model-value="updateNfs"
          :policy-rules="policyRules?.nfs?.attributes"
          :disable="!tab.removeable || sectionDisabled"
          :disable-info="disableFields?.nfs"
        />
        <s3-instance-section
          v-else-if="tab.label === AssetKind.S3 && s3InstancesMapById[tab.id]"
          :key="s3InstancesMapById[tab.id]?.id"
          :model-value="s3InstancesMapById[tab.id]"
          :secrets="s3Secrets"
          @update:model-value="updateS3"
          :policy-rules="policyRules?.s3?.attributes"
          :disable="!tab.removeable || sectionDisabled"
        />

        <git-instance-section
          v-else-if="tab.label === AssetKind.Git && gitInstancesMapById[tab.id]"
          :key="gitInstancesMapById[tab.id]?.id"
          :model-value="gitInstancesMapById[tab.id]"
          :secrets="gitSecrets"
          :support-git-sync-one-parameter="true"
          @update:model-value="updateGit"
          :policy-rules="policyRules?.git?.attributes"
          :disable="!tab.removeable || sectionDisabled"
          :disable-info="disableFields?.git"
        />

        <secret-volume-instance-section
          v-else-if="tab.label === AssetKind.SecretVolume && secretVolumeInstancesMapById[tab.id]"
          :key="secretVolumeInstancesMapById[tab.id]?.id"
          :model-value="secretVolumeInstancesMapById[tab.id]"
          :secrets="[secretVolumeInstancesMapById[tab.id]?.secret || '']"
          allow-only-container-path
          :loading-secrets="false"
          @update:model-value="updateSecretVolume"
          :disable="!tab.removeable || sectionDisabled"
        />

        <config-map-instance-section
          v-else-if="tab.label === AssetKind.ConfigMap && configMapInstancesMapById[tab.id]"
          :key="configMapInstancesMapById[tab.id]?.id"
          :model-value="configMapInstancesMapById[tab.id]"
          :config-map-names="[configMapInstancesMapById[tab.id]?.configMap || '']"
          allow-only-container-path
          :loading-config-map-names="false"
          :supports-sub-path="supportsSubPath"
          @update:model-value="updateConfigMap"
          :disable="!tab.removeable || sectionDisabled"
          :disable-info="disableFields?.configMap"
        />

        <pvc-instance-section
          v-else-if="tab.label === AssetKind.Pvc && pvcInstancesMapById[tab.id]"
          :key="pvcInstancesMapById[tab.id]?.id"
          :model-value="pvcInstancesMapById[tab.id]"
          :loading-pvc-claim-names="false"
          :pvc-claim-names="[pvcInstancesMapById[tab.id]?.claimName || '']"
          allow-only-container-path
          @update:model-value="updatePvc"
          :disable="!tab.removeable || sectionDisabled"
          :disable-info="disableFields?.pvc"
        />

        <data-volume-instance-section
          v-else-if="tab.label === AssetKind.DataVolume && dataVolumeInstancesMapById[tab.id]"
          :key="dataVolumeInstancesMapById[tab.id]?.id"
          :model-value="dataVolumeInstancesMapById[tab.id]"
          @update:model-value="updateDataVolume"
          :policy-rules="policyRules?.dataVolume?.attributes"
          :disable="!tab.removeable || sectionDisabled"
        />

        <volume-instance-section
          v-else-if="tab.label === UI_VOLUME_KIND && volumeInstancesMapById[tab.id]"
          :disable="!tab.removeable || sectionDisabled"
          :model-value="volumeInstancesMapById[tab.id]"
          :policy-rules="policyRules?.pvc?.attributes"
          :storage-classes="storageClasses"
          :loading-storage-class="loadingStorageClass"
          @update:model-value="updateVolume"
          :show-volume-persistency-section="showVolumePersistencySection"
          :disable-info="disableFields?.volume"
        />

        <empty-dir-instance-section
          v-else-if="tab.label === UI_EMPTY_DIR_KIND && emptyDirInstancesMapById[tab.id]"
          :key="emptyDirInstancesMapById[tab.id].id"
          :model-value="emptyDirInstancesMapById[tab.id]"
          @update:model-value="updateEmptyDir"
          :policy-rules="policyRules?.emptyDirVolume?.attributes"
          :disable="!tab.removeable || sectionDisabled"
          :disable-info="disableFields?.emptyDir"
        />
      </runai-tab-panel>
    </runai-tab-panels>

    <div v-if="!tabs?.length" class="q-my-lg column items-center text-italic text-b-main">
      <q-icon size="55px" name="ph-bold ph-database" class="q-mb-md" color="black-54"></q-icon>
      <slot name="empty-state-title">
        <div class="text-black-54">No {{ dataSourceTitle }} yet.</div>
        <div v-if="!sectionDisabled" class="text-black-54">
          Click the buttons above to load existing setups or create a one-time configuration
        </div>
      </slot>
    </div>
  </hybrid-section-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { HybridSectionWrapper } from "../hybrid-section-wrapper";
import { NfsInstanceSection } from "./nfs-instance-section";
import { HostPathInstanceSection } from "./host-path-instance-section";
import { S3InstanceSection } from "./s3-instance-section";
import { PvcInstanceSection } from "./pvc-instance-section";
import { GitInstanceSection } from "@/components/workload-creation/hybrid-data-source-section/git-instance-section";
import { VolumeInstanceSection } from "./volume-instance-section";
import { SecretVolumeInstanceSection } from "./secret-volume-instance-section";
import { ConfigMapInstanceSection } from "./config-map-instance-section";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiTab, RunaiTabPanels, RunaiTabPanel, type ITab } from "@/components/common/runai-tab";
import { DataSourceDropdown } from "@/components/data-source/data-source-dropdown";
import { DataVolumeInstanceSection } from "./data-volume-instance-section";
import { EmptyDirInstanceSection } from "./empty-dir-instance-section";
// Models
import { AssetKind } from "@/swagger-models/assets-service-client";
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";
import {
  type IUIDataVolumeInstance,
  type IUIConfigMapInstance,
  type IUIGitInstance,
  type IUIHostPathInstance,
  type IUINfsInstance,
  type IUIPvcInstance,
  type IUIS3Instance,
  type IUISecretVolumeInstance,
  type IUIStorageFields,
  type IUIEmptyDirInstance,
} from "@/models/workload.model";
import type { ITabProps } from "@/components/section/compute-resource-section/container-path-override/container-path-override.model";
import {
  isNotEmpty,
  isValidDirectoryPath,
  isValidPrivateUrl,
  isValidSubPath,
  isValidUrl,
} from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import type { TDataSourceKinds, TUIDataSourceInstancesKinds } from "@/models/data-source.model";
import {
  DATA_SOURCE_KIND_TO_CONTAINER_KEY,
  DATA_SOURCE_KIND_TO_STORAGE_FIELD_KEY,
  DS_KIND_TO_STORAGE_FIELDS_KEY,
  UI_EMPTY_DIR_KIND,
  UI_VOLUME_KIND,
} from "@/models/data-source.model";
import { makeId } from "@/utils/common.util";
import { SecretType, type SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type {
  EmptyDirInstance,
  PvcInstance,
  SupersetDefaultsAllOfStorage,
  SupersetRulesAllOfStorage,
} from "@/swagger-models/policy-service-client";
import { useStorageClassConfiguration } from "@/composables/use-storage-class-configuration.composable";
import { useSettingStore } from "@/stores/setting.store";
import type { IHybridDataSourceSectionDisableFields } from "./hybrid-data-source-section.model";
import {
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_SUB_PATH,
  MIN_CLUSTER_VERSION_FOR_EMPTY_DIR,
  MIN_CLUSTER_VERSION_FOR_GIT_SSH,
} from "@/common/version.constant";
import { useClusterStore } from "@/stores/cluster.store";
import { RunaiBtn } from "@/components/common/runai-btn";
import type { ValidationRule, QField } from "quasar";

export default defineComponent({
  name: "hybrid-data-source-section",
  components: {
    HybridSectionWrapper,
    NfsInstanceSection,
    S3InstanceSection,
    PvcInstanceSection,
    HostPathInstanceSection,
    GitInstanceSection,
    SecretVolumeInstanceSection,
    ConfigMapInstanceSection,
    DataVolumeInstanceSection,
    VolumeInstanceSection,
    EmptyDirInstanceSection,
    RunaiSvgIcon,
    RunaiTab,
    RunaiTabPanels,
    RunaiTabPanel,
    DataSourceDropdown,
    RunaiBtn,
  },
  emits: ["load-section-templates", "update"],
  props: {
    workloadStorage: {
      type: Object as PropType<IUIStorageFields>,
      default: () => ({}),
    },
    secrets: {
      type: Array as PropType<SecretListInfo[]>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfStorage | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<SupersetDefaultsAllOfStorage | null>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    allowedDataSourceTypes: {
      type: Object as PropType<Set<TUIDataSourceInstancesKinds>>,
      required: false,
      default: () => new Set([AssetKind.Nfs, AssetKind.HostPath, AssetKind.S3, AssetKind.Git, UI_VOLUME_KIND]),
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    canAddNew: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showContentHeader: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showVolumePersistencySection: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    isVolumesEnabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    disableFields: {
      type: Object as PropType<IHybridDataSourceSectionDisableFields>,
      required: false,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    required: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    customSectionLabel: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      assetKindToCmpMap: {
        [AssetKind.Nfs]: NfsInstanceSection,
        [AssetKind.S3]: S3InstanceSection,
        [AssetKind.HostPath]: HostPathInstanceSection,
        [AssetKind.Git]: GitInstanceSection,
        [AssetKind.SecretVolume]: SecretVolumeInstanceSection,
        [AssetKind.ConfigMap]: ConfigMapInstanceSection,
        [AssetKind.Pvc]: PvcInstanceSection,
        [UI_VOLUME_KIND]: VolumeInstanceSection,
        [UI_EMPTY_DIR_KIND]: EmptyDirInstanceSection,
      },
      assetKindToLabelMap: {
        [AssetKind.Nfs]: "NFS",
        [AssetKind.S3]: "S3",
        [AssetKind.HostPath]: "Host path",
        [AssetKind.Git]: "Git",
        [AssetKind.SecretVolume]: "Secret",
        [AssetKind.ConfigMap]: "Config map",
        [AssetKind.Pvc]: "PVC",
        [AssetKind.DataVolume]: "Data volume",
        [UI_VOLUME_KIND]: "Volume",
        [UI_EMPTY_DIR_KIND]: "EmptyDir",
      } as Record<TUIDataSourceInstancesKinds, string>,
      selectedId: null as string | null,
      dataSourceIconNames: DATA_SOURCE_ICONS_NAMES,
      defaultTabsValue: null,
      settingStore: useSettingStore(),
      clusterStore: useClusterStore(),
      AssetKind,
      UI_VOLUME_KIND,
      UI_EMPTY_DIR_KIND,
    };
  },
  setup() {
    const { storageClasses, loadingStorageClass, loadStorageClasses, loadStorageClassConfiguration } =
      useStorageClassConfiguration("workload");

    return {
      storageClasses,
      loadingStorageClass,
      loadStorageClasses,
      loadStorageClassConfiguration,
    };
  },
  async created() {
    await this.loadStorageClasses(
      this.clusterId,
      this.workloadStorage.volumes?.map((ds) => ds.claimInfo?.storageClass) || [],
    );
  },
  computed: {
    validationRules(): ValidationRule[] {
      if (this.required) {
        return [this.isRequired, this.isIncomplete];
      }
      return [this.isIncomplete];
    },
    nfsInstancesMapById(): Record<string, IUINfsInstance> {
      return (
        this.workloadStorage.nfs?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUINfsInstance>,
        ) || {}
      );
    },
    hostPathInstancesMapById(): Record<string, IUIHostPathInstance> {
      return (
        this.workloadStorage.hostPath?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIHostPathInstance>,
        ) || {}
      );
    },
    s3InstancesMapById(): Record<string, IUIS3Instance> {
      return (
        this.workloadStorage.s3?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIS3Instance>,
        ) || {}
      );
    },
    gitInstancesMapById(): Record<string, IUIGitInstance> {
      return (
        this.workloadStorage.git?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIGitInstance>,
        ) || {}
      );
    },
    secretVolumeInstancesMapById(): Record<string, IUISecretVolumeInstance> {
      return (
        this.workloadStorage.secretVolume?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUISecretVolumeInstance>,
        ) || {}
      );
    },
    emptyDirInstancesMapById(): Record<string, IUIEmptyDirInstance> {
      return (
        this.workloadStorage.emptyDirVolume?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIEmptyDirInstance>,
        ) || {}
      );
    },
    configMapInstancesMapById(): Record<string, IUIConfigMapInstance> {
      return (
        this.workloadStorage.configMapVolume?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIConfigMapInstance>,
        ) || {}
      );
    },
    pvcInstancesMapById(): Record<string, IUIPvcInstance> {
      return (
        this.workloadStorage.pvc?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIPvcInstance>,
        ) || {}
      );
    },
    dataVolumeInstancesMapById(): Record<string, IUIDataVolumeInstance> {
      return (
        this.workloadStorage.dataVolume?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIDataVolumeInstance>,
        ) || {}
      );
    },
    volumeInstancesMapById(): Record<string, IUIPvcInstance> {
      return (
        this.workloadStorage.volumes?.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, IUIPvcInstance>,
        ) || {}
      );
    },
    sectionLabel(): string {
      if (this.customSectionLabel) return this.customSectionLabel;
      return this.settingStore.isDataVolumeEnabled ? "Data & storage" : "Data Sources";
    },
    dataSourceTitle(): string {
      return this.settingStore.isDataVolumeEnabled ? "data and storage" : "data sources";
    },
    s3Secrets(): string[] {
      return this.secrets.filter((secret) => secret.type === SecretType.AccessKey).map((secret) => secret.name);
    },
    gitSecrets(): SecretListInfo[] {
      const filterCB = (s: SecretListInfo) =>
        s.type === SecretType.Password || (this.supportsGenericSecretsForGit && s.type === SecretType.Generic);
      return this.secrets.filter(filterCB);
    },
    supportsSubPath(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_SUB_PATH);
    },
    supportsGenericSecretsForGit(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_GIT_SSH);
    },
    supportsEmptyDir(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_EMPTY_DIR);
    },
    allowedOneOffDataSources(): Set<TUIDataSourceInstancesKinds> {
      const oneOffSupported: TUIDataSourceInstancesKinds[] = [
        AssetKind.Nfs,
        AssetKind.HostPath,
        AssetKind.S3,
        AssetKind.Git,
      ];

      const allowed: TUIDataSourceInstancesKinds[] = [...this.allowedDataSourceTypes];

      if (this.isVolumesEnabled) {
        oneOffSupported.push(UI_VOLUME_KIND);
        allowed.push(UI_VOLUME_KIND);
      }

      if (this.supportsEmptyDir) {
        oneOffSupported.push(UI_EMPTY_DIR_KIND);
        allowed.push(UI_EMPTY_DIR_KIND);
      }

      const options = allowed
        .filter((ds) => oneOffSupported.includes(ds))
        .filter(
          (t) =>
            this.policyRules?.[DS_KIND_TO_STORAGE_FIELDS_KEY[t] as keyof SupersetRulesAllOfStorage]?.instances
              ?.canAdd !== false,
        );

      return new Set(options);
    },
    addTabItem(): ITab[] {
      return [
        {
          name: "add-one-off",
          label: "",
          hidden: !this.canAddNew,
          disabled: this.sectionDisabled || this.cannotAddOneOffDataSources,
          tabClass: "add-one-off-tab",
          aid: "add-one-off-tab",
        },
      ];
    },
    dataSourceTabsById(): Record<string, ITabProps> {
      return this.tabs.reduce(
        (acc, tab) => {
          acc[tab.id] = tab;
          return acc;
        },
        {} as Record<string, ITabProps>,
      );
    },
    dataSourceTabItems(): ITab[] {
      return this.tabs.map((tab) => ({
        name: tab.id,
        label: this.assetKindToLabelMap[tab.label as TUIDataSourceInstancesKinds],
        errorIcon: tab.isInvalid,
        tabClass: { tab: true, row: true, "items-center": true, disabled: !tab.removeable },
        aid: `data-source-tab-${tab.id}`,
      }));
    },
    cannotAddOneOffDataSources(): boolean {
      return this.allowedOneOffDataSources.size === 0;
    },
    dataSourceSummary(): string {
      let summary = "";
      if (this.tabs.length > 0) {
        summary = `${this.assetKindToLabelMap[this.tabs[0].label as TUIDataSourceInstancesKinds]}`;
      }
      if (this.tabs.length > 1) {
        return summary + ` + ${this.tabs.length - 1} more`;
      }
      if (this.tabs.length === 0) {
        return "None";
      }
      return summary;
    },
    tabs(): ITabProps[] {
      return Object.values(this.workloadStorage)
        .flat()
        .filter((ds) => !!ds && !ds.exclude)
        .sort((a, b) => ((a.createdAt || 0) > (b.createdAt || 0) ? -1 : 1))
        .map((ds) => {
          return {
            key: ds.id,
            id: ds.id,
            label: ds.type,
            isInvalid: this.isInvalidStorage(ds),
            removeable: this.isRemovableStorage(ds),
            icon: DATA_SOURCE_ICONS_NAMES[ds.type as keyof typeof DATA_SOURCE_ICONS_NAMES],
          };
        });
    },
    isValidationError(): boolean {
      return this.tabs.some((tab) => tab.isInvalid) || (this.required && this.tabs.length === 0);
    },
  },
  methods: {
    addOneOffStorage(type: TUIDataSourceInstancesKinds): void {
      const storageKey = DS_KIND_TO_STORAGE_FIELDS_KEY[type];

      const dsAttributesDefaults =
        this.policyDefaults?.[storageKey as keyof SupersetDefaultsAllOfStorage]?.attributes || {};

      if (type && storageKey) {
        let storageToAdd = { id: makeId(), type, createdAt: Date.now(), ...dsAttributesDefaults };

        let key = storageKey as keyof IUIStorageFields;

        if (type === UI_VOLUME_KIND) {
          key = "volumes";

          storageToAdd = {
            ...storageToAdd,
            claimInfo: {
              ...((storageToAdd as IUIPvcInstance).claimInfo || {}),
              size: (dsAttributesDefaults as PvcInstance).claimInfo?.size || "1G",
            },
          };
        }

        if (type === UI_EMPTY_DIR_KIND) {
          key = "emptyDirVolume";
          storageToAdd = {
            ...storageToAdd,
            sizeLimit: (dsAttributesDefaults as EmptyDirInstance).sizeLimit || "1G",
          };
        }

        this.$emit("update", {
          ...this.workloadStorage,
          [key]: [storageToAdd, ...(this.workloadStorage[key] || [])],
        });

        this.selectedId = storageToAdd.id;

        // if its a volume and there is only one storage class, autoselect that storage class for that specific volume
        if (type === UI_VOLUME_KIND && this.storageClasses.length === 1) {
          this.updateVolume({
            ...(storageToAdd as IUIPvcInstance),
            claimInfo: {
              ...(storageToAdd as IUIPvcInstance).claimInfo,
              storageClass: this.storageClasses[0].storageClassName,
            },
          } as IUIPvcInstance);
        }
      }
    },
    isRemovableStorage(
      ds:
        | IUINfsInstance
        | IUIHostPathInstance
        | IUIS3Instance
        | IUIGitInstance
        | IUISecretVolumeInstance
        | IUIConfigMapInstance
        | IUIPvcInstance,
    ): boolean {
      switch (ds.type) {
        case AssetKind.Nfs:
          return !this.policyRules?.nfs?.instances?.locked?.includes(ds.name || "");
        case AssetKind.HostPath:
          return !this.policyRules?.hostPath?.instances?.locked?.includes(ds.name || "");
        case AssetKind.S3:
          return !this.policyRules?.s3?.instances?.locked?.includes(ds.name || "");
        case AssetKind.Git:
          return !this.policyRules?.git?.instances?.locked?.includes(ds.name || "");
        case AssetKind.SecretVolume:
          return !this.policyRules?.secretVolume?.instances?.locked?.includes(ds.name || "");
        case AssetKind.ConfigMap:
          return !this.policyRules?.configMapVolume?.instances?.locked?.includes(ds.name || "");
        case AssetKind.Pvc:
          return !this.policyRules?.pvc?.instances?.locked?.includes(ds.name || "");
        case AssetKind.DataVolume:
          return !this.policyRules?.dataVolume?.instances?.locked?.includes(ds.name || "");
        case UI_VOLUME_KIND:
          return !this.policyRules?.pvc?.instances?.locked?.includes(ds.name || "");
        case UI_EMPTY_DIR_KIND:
          return !this.policyRules?.emptyDirVolume?.instances?.locked?.includes(ds.name || "");
        default:
          return true;
      }
    },
    isInvalidStorage(
      ds:
        | IUINfsInstance
        | IUIS3Instance
        | IUIPvcInstance
        | IUIHostPathInstance
        | IUIGitInstance
        | IUISecretVolumeInstance
        | IUIConfigMapInstance
        | IUIDataVolumeInstance,
    ): boolean {
      switch (ds.type) {
        case AssetKind.Nfs:
          return this.isInvalidNfs(ds);
        case AssetKind.HostPath:
          return this.isInvalidHostPath(ds);
        case AssetKind.S3:
          return this.isInvalidS3(ds);
        case AssetKind.Git:
          return this.isInvalidGit(ds);
        case UI_VOLUME_KIND:
          return this.isInvalidVolume(ds);
        case AssetKind.ConfigMap:
          return this.isInvalidConfigMap(ds);
        case UI_EMPTY_DIR_KIND:
          return this.isInvalidEmptyDir(ds);
        default:
          return this.isInvalidContainerPath(ds);
      }
    },
    isInvalidEmptyDir(ds: IUIEmptyDirInstance): boolean {
      const validSize: boolean = isNotEmpty(ds.sizeLimit || "");
      const validPath: boolean = isNotEmpty(ds.path || "") && isValidDirectoryPath(ds.path || "");
      return !validSize || !validPath;
    },
    isInvalidNfs(ds: IUINfsInstance): boolean {
      const validServer: boolean = isNotEmpty(ds.server || "");
      const validPath: boolean = isNotEmpty(ds.path || "") && isValidDirectoryPath(ds.path || "");
      const validMountPath: boolean = isNotEmpty(ds.mountPath || "") && isValidDirectoryPath(ds.mountPath || "");
      return !validServer || !validPath || !validMountPath;
    },
    isInvalidHostPath(ds: IUIHostPathInstance): boolean {
      const validPath: boolean = isNotEmpty(ds.path || "") && isValidDirectoryPath(ds.path || "");
      const validMountPath: boolean = isNotEmpty(ds.mountPath || "") && isValidDirectoryPath(ds.mountPath || "");
      return !validPath || !validMountPath;
    },
    isInvalidS3(ds: IUIS3Instance): boolean {
      const validBucket: boolean = isNotEmpty(ds.bucket || "");
      const validPath: boolean = isNotEmpty(ds.path || "") && isValidDirectoryPath(ds.path || "");
      const validUrl: boolean = isNotEmpty(ds.url || "");
      return !validBucket || !validPath || !validUrl;
    },
    isInvalidGit(ds: IUIGitInstance): boolean {
      const isPrivate: boolean = !!ds.passwordSecret || !!ds.secretRef?.name;
      const urlNotEmpty = isNotEmpty(ds.repository || "");
      const validUrl: boolean =
        urlNotEmpty && (isPrivate ? isValidPrivateUrl(ds.repository || "") : isValidUrl(ds.repository || ""));
      const validPath: boolean = isNotEmpty(ds.path || "") && isValidDirectoryPath(ds.path || "");
      return !validUrl || !validPath;
    },
    isInvalidVolume(ds: IUIPvcInstance): boolean {
      const validSC: boolean = isNotEmpty(ds.claimInfo?.storageClass || "");
      const sc = this.storageClasses.find((sc) => sc.storageClassName === ds.claimInfo?.storageClass);
      const validACM: boolean = sc?.customization?.accessMode?.required
        ? !!ds.claimInfo?.accessModes?.readOnlyMany ||
          !!ds.claimInfo?.accessModes?.readWriteMany ||
          !!ds.claimInfo?.accessModes?.readWriteOnce
        : true;
      const validSize: boolean = isNotEmpty(ds.claimInfo?.size || "");
      const validPath: boolean = isNotEmpty(ds.path || "") && isValidDirectoryPath(ds.path || "");

      return !validSC || !validACM || !validSize || !validPath;
    },
    isInvalidConfigMap(ds: IUIConfigMapInstance): boolean {
      const validConfigMap: boolean = isNotEmpty(ds.configMap || "");
      const validMountPath: boolean = isNotEmpty(ds.mountPath || "") && isValidDirectoryPath(ds.mountPath || "");
      const validSubPath: boolean = ds.subPath ? isValidSubPath(ds.subPath) : true;

      return !validConfigMap || !validMountPath || !validSubPath;
    },
    isInvalidContainerPath(
      ds:
        | IUINfsInstance
        | IUIS3Instance
        | IUIPvcInstance
        | IUIHostPathInstance
        | IUIGitInstance
        | IUISecretVolumeInstance
        | IUIConfigMapInstance
        | IUIDataVolumeInstance,
    ) {
      if (!ds) {
        return false;
      }
      const containerPathKey = DATA_SOURCE_KIND_TO_CONTAINER_KEY[ds.type as TDataSourceKinds] as keyof typeof ds;
      const containerPath = ds[containerPathKey];

      return (
        this.notEmpty(containerPath?.toString() || "") !== true ||
        this.isDirectoryPath(containerPath?.toString() || "") !== true
      );
    },
    updateNfs(data: IUINfsInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        nfs: this.workloadStorage.nfs?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updateHostPath(data: IUIHostPathInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        hostPath: this.workloadStorage.hostPath?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updateS3(data: IUIS3Instance) {
      this.$emit("update", {
        ...this.workloadStorage,
        s3: this.workloadStorage.s3?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updateGit(data: IUIGitInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        git: this.workloadStorage.git?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updateSecretVolume(data: IUISecretVolumeInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        secretVolume: this.workloadStorage.secretVolume?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updateConfigMap(data: IUIConfigMapInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        configMapVolume: this.workloadStorage.configMapVolume?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updatePvc(data: IUIPvcInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        pvc: this.workloadStorage.pvc?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updateEmptyDir(data: IUIEmptyDirInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        emptyDirVolume: this.workloadStorage.emptyDirVolume?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    updateDataVolume(data: IUIDataVolumeInstance) {
      this.$emit("update", {
        ...this.workloadStorage,
        dataVolume: this.workloadStorage.dataVolume?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    async updateVolume(data: IUIPvcInstance) {
      const newStorageClass = data?.claimInfo?.storageClass;
      const oldStorageClass = this.workloadStorage?.volumes?.find((v) => v.id === data.id)?.claimInfo?.storageClass;

      if (newStorageClass && newStorageClass !== oldStorageClass) {
        data.claimInfo = await this.loadStorageClassConfiguration(this.clusterId, newStorageClass);
      }

      this.$emit("update", {
        ...this.workloadStorage,
        volumes: this.workloadStorage.volumes?.map((ds) => (ds.id === data.id ? data : ds)),
      });
    },
    removeStorage(id: string, type: string): void {
      let storageKey = DATA_SOURCE_KIND_TO_STORAGE_FIELD_KEY[type as TDataSourceKinds] as keyof IUIStorageFields;

      if (type === UI_VOLUME_KIND) storageKey = "volumes";
      if (type === UI_EMPTY_DIR_KIND) storageKey = "emptyDirVolume";

      const storage = this.workloadStorage[storageKey];
      if (!storage) {
        return;
      }

      const itemToRemove = storage.find((ds) => ds.id === id);
      const hasBackendAssignedName =
        itemToRemove && "name" in itemToRemove && !!(itemToRemove as { name?: string | null }).name;

      if (hasBackendAssignedName) {
        this.$emit("update", {
          ...this.workloadStorage,
          [storageKey]: storage.map((ds) => (ds.id === id ? { ...ds, exclude: true } : ds)),
        });
      } else {
        this.$emit("update", {
          ...this.workloadStorage,
          [storageKey]: storage.filter((ds) => ds.id !== id),
        });
      }
      (this.$refs.validationField as QField)?.resetValidation();
    },
    isDirectoryPath(val: string): boolean | string {
      return isValidDirectoryPath(val) || errorMessages.INVALID_DIRECTORY_PATH;
    },
    notEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.PATH_NOT_EMPTY;
    },
    isIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.AT_LEAST_ONE_TARGET_LOCATION_ISSUE;
    },
    isRequired(): boolean | string {
      return this.tabs.length > 0 ? true : errorMessages.ADD_A_DATA_SOURCE;
    },
    clearInvalidStorageClasses(): void {
      const updatedVolumes: IUIPvcInstance[] | undefined = this.workloadStorage.volumes?.map((ds) => {
        const scFound = (this.storageClasses || []).find(
          (sc) => sc.storageClassName === ds.claimInfo?.storageClass,
        )?.storageClassName;
        if (!!ds.claimInfo?.storageClass && !scFound) {
          ds.claimInfo = {
            ...(ds.claimInfo || {}),
            storageClass: undefined,
          };
        }
        return ds;
      });

      this.$emit("update", {
        ...this.workloadStorage,
        volumes: updatedVolumes,
      });
    },
    clearInvalidSecrets(): void {
      this.workloadStorage.configMapVolume;
      this.workloadStorage.secretVolume;

      const clearedS3 = this.workloadStorage.s3?.map((s3) => {
        if (s3.accessKeySecret && !this.s3Secrets.includes(s3.accessKeySecret)) {
          return {
            ...s3,
            accessKeySecret: undefined,
          };
        }
        return s3;
      });
      const clearedGit = this.workloadStorage.git?.map((git) => {
        if (git.secretRef?.name && !this.gitSecrets.find((s) => s.name === git.secretRef?.name)) {
          return {
            ...git,
            secretRef: undefined,
          };
        }
        return git;
      });

      this.$emit("update", {
        ...this.workloadStorage,
        s3: clearedS3,
        git: clearedGit,
      });
    },
  },
  watch: {
    workloadStorage: {
      handler() {
        const selectedTab = this.tabs.find((tab) => tab.id === this.selectedId);
        // Selecting the first tab if there is no selected tab
        if (!selectedTab && this.tabs.length > 0) {
          this.selectedId = this.tabs[0]?.id;
        }
      },
      immediate: true,
    },
    clusterId: {
      async handler(newValue, oldValue) {
        if (oldValue !== newValue) {
          await this.loadStorageClasses(
            newValue,
            this.workloadStorage.volumes?.map((ds) => ds.claimInfo?.storageClass) || [],
          );
          this.clearInvalidStorageClasses();
        }
      },
    },
    secrets: {
      handler() {
        this.clearInvalidSecrets();
      },
    },
  },
});
</script>

<style lang="scss">
.hybrid-data-source-section {
  .data-source-tabs {
    overflow: hidden;
    .q-tab {
      position: relative;
      .close-btn {
        color: $black-54;
        font-size: 12px;
        display: none;
      }

      &.add-one-off-tab {
        padding: 0 8px;
      }

      &:hover {
        .close-btn {
          display: block;
          position: absolute;
          right: -12px;
        }
      }
    }
  }
}
</style>
