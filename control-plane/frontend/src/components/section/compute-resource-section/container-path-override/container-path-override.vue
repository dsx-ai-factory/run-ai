<template>
  <section class="container-path-override q-mt-xl">
    <div class="q-mb-lg">
      Modify the data target location for the selected
      {{ settingStore.isDataVolumeEnabled ? "data and storage" : "data sources" }} <span class="italic">(optional)</span>
      <q-field class="q-mt-md form-hint no-padding" :model-value="isValidationError" :rules="[isIncomplete]"></q-field>
    </div>
    <div class="q-gutter-y-md q-mb-md">
      <runai-tab v-model="selectedDatasourceId" outside-arrows class="container-path-tabs" :tabs="containerPathTabs" />
    </div>

    <policy-string-field
      aid="container-path-input"
      :model-value="containerPathValue"
      no-error-icon
      label="Container path"
      stack-label
      placeholder="e.g., /home/localdata"
      input-class="placeholder-italic"
      :debounce="300"
      :disable="disabled"
      :rules="[notEmpty, isDirectoryPath]"
      :policy-rules="relevantPolicyRule"
      @update:model-value="updateContainerPath"
    />
  </section>
</template>

<script lang="ts">
import { isNotEmpty, isValidDirectoryPath } from "@/common/form.validators";
import { defineComponent } from "vue";
import type { PropType } from "vue";
// Models
import { errorMessages } from "@/common/error-message.constant";
import type { StringRules, SupersetRulesAllOfStorage } from "@/swagger-models/policy-service-client";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiTab, type ITab } from "@/components/common/runai-tab";

import type {
  ISelectedDatasourcesContainerPathData,
  IContainerPathOverride,
  ITabProps,
} from "./container-path-override.model";
import { useSettingStore } from "@/stores/setting.store";

export default defineComponent({
  name: "container-path-override",
  components: {
    PolicyStringField,
    RunaiTab,
  },
  emits: ["update-container-path-override", "is-section-invalid"],
  props: {
    dataSourcesContainerPathData: {
      type: Object as PropType<ISelectedDatasourcesContainerPathData[]>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfStorage | null>,
      required: false,
    },
  },
  data() {
    return {
      selectedDatasourceId: "",
      dsMapByIds: {} as Record<string, ISelectedDatasourcesContainerPathData>,
      settingStore: useSettingStore(),
    };
  },
  created() {
    this.selectedDatasourceId = this.dataSourceTabs?.[0]?.id || "";
  },
  computed: {
    containerPathTabs(): ITab[] {
      return this.dataSourceTabs.map((ds) => ({
        name: ds.id,
        label: ds.label,
        errorIcon: ds.isInvalid,
        aid: `data-source-cp-tab-${ds.id}`,
      }));
    },
    dataSourceTabs(): ITabProps[] {
      return this.dataSourcesContainerPathData.map((ds) => {
        return {
          key: ds.id,
          id: ds.id,
          label: ds.name,
          isInvalid: this.shouldShowIcon(ds.id),
        };
      });
    },
    containerPathValue(): string {
      // if there is an override, show the override value, otherwise show the default value
      if (typeof this.selectedDataSourceOverride === "string") return this.selectedDataSourceOverride;
      return this.selectedDataSourceContainerPath;
    },
    selectedDataSourceContainerPath(): string {
      return this.dsMapByIds[this.selectedDatasourceId]?.origin || "";
    },
    selectedDataSourceOverride(): string | undefined {
      const val = this.dsMapByIds[this.selectedDatasourceId]?.overrides?.containerPath;
      return typeof val === "string" ? val : undefined;
    },
    isValidationError(): boolean {
      return this.dataSourceTabs.some((tab) => tab.isInvalid);
    },
    relevantPolicyRule(): StringRules | null {
      const kind = this.dsMapByIds[this.selectedDatasourceId].kind;
      if (!kind) return null;

      switch (kind) {
        case "git":
          return this.policyRules?.git?.attributes?.path || null;
        case "pvc":
          return this.policyRules?.pvc?.attributes?.path || null;
        case "s3":
          return this.policyRules?.s3?.attributes?.path || null;
        case "config-map":
          return this.policyRules?.configMapVolume?.attributes?.mountPath || null;
        case "nfs":
          return this.policyRules?.nfs?.attributes?.mountPath || null;
        case "hostPath":
          return this.policyRules?.hostPath?.attributes?.mountPath || null;
        case "secret-volume":
          return this.policyRules?.secretVolume?.attributes?.mountPath || null;
        default:
          return null;
      }
    },
  },
  methods: {
    updateContainerPath(val: string | number | null) {
      const cpOverrride: IContainerPathOverride = {
        dataSourceId: this.selectedDatasourceId,
        containerPathOverride: val?.toString() || "",
      };
      this.$emit("update-container-path-override", cpOverrride);
    },
    notEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.PATH_NOT_EMPTY;
    },
    isDirectoryPath(val: string): boolean | string {
      return isValidDirectoryPath(val) || errorMessages.INVALID_DIRECTORY_PATH;
    },
    shouldShowIcon(dataSourceId: string): boolean {
      const dsOverride = this.dsMapByIds[dataSourceId]?.overrides;
      if (typeof dsOverride?.containerPath !== "string") return false;
      return !(isNotEmpty(dsOverride.containerPath) && isValidDirectoryPath(dsOverride.containerPath));
    },
    isIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.AT_LEAST_ONE_TARGET_LOCATION_ISSUE;
    },
  },
  watch: {
    dataSourcesContainerPathData: {
      handler(
        newVal: ISelectedDatasourcesContainerPathData[],
        oldVal: ISelectedDatasourcesContainerPathData[] = [],
      ): void {
        if (newVal.length < oldVal.length) {
          const selectedStillExists = !!newVal.find((ds) => ds.id === this.selectedDatasourceId);
          if (!selectedStillExists) {
            // if the selected data source was removed, update selected data source to the first one
            this.selectedDatasourceId = newVal[0].id;
          }
        } else if (newVal.length > oldVal.length) {
          // if a new data source was added, select it
          this.selectedDatasourceId = newVal[newVal.length - 1].id;
        }
        this.dsMapByIds = newVal.reduce(
          (acc, ds) => {
            acc[ds.id] = ds;
            return acc;
          },
          {} as Record<string, ISelectedDatasourcesContainerPathData>,
        );
      },
      immediate: true,
    },
    isValidationError: {
      handler(val: boolean): void {
        this.$emit("is-section-invalid", val);
      },
      immediate: true,
    },
  },
});
</script>
<style lang="scss">
.container-path-override {
  .container-path-tabs {
    .q-tabs__arrow {
      font-size: 18px;
    }
  }
}
</style>
