<template>
  <runai-table
    :rows="rows"
    :columns="columns"
    :loading="false"
    hide-pagination
    :filter-by="{ rowsPerPage: 0 }"
    no-shadow
    sticky-header
    :style="{ maxHeight }"
    disable-selection
    :bordered="bordered"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiTable } from "@/components/common/runai-table";
import type { ITableColumn } from "@/models/table.model";
import {
  distributedPreviewPolicyColumns,
  distributedPreviewPolicyColumnsDataSourceTable,
  workloadPreviewPolicyColumns,
  workloadPreviewPolicyColumnsDataSourceTable,
} from "@/table-models/policy.table-model";
import { EPolicyGroup, type IPolicyViewProperty } from "@/utils/policy.util/policy-view.util";
import { AssetKind, type ComplianceInfo, type ComplianceInfoReason } from "@/swagger-models/assets-service-client";
import { stringUtil } from "@/utils/string.util";

export interface IComplianceData {
  complianceInfo: ComplianceInfo;
  kind?: AssetKind;
  distributedReplicaType?: string;
}

export default defineComponent({
  name: "policy-details-table",
  components: {
    RunaiTable,
  },
  props: {
    policyData: {
      type: Array as PropType<IPolicyViewProperty[]>,
      required: true,
    },
    policyTable: {
      type: String as PropType<EPolicyGroup>,
      required: true,
    },
    isDistributed: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    // when this field is passed the component will filter the policyData by the compliance fields
    complianceData: {
      type: Object as PropType<IComplianceData>,
      required: false,
    },
    distributedReplicaType: {
      type: String as PropType<string>,
      required: false,
    },
    maxHeight: {
      type: String,
      default: "500px",
    },
    bordered: {
      type: Boolean,
      default: true,
    },
  },
  computed: {
    rows(): Array<IPolicyViewProperty> {
      if (this.complianceData) return this.policyFilteredByComplianced;
      if (this.isDistributed && this.distributedReplicaType) return this.policyFilteredByReplica;
      return this.policyData;
    },
    policyFilteredByReplica(): Array<IPolicyViewProperty> {
      const replica = this.distributedReplicaType?.toLocaleLowerCase() || "";
      const replicaMatch = (row: IPolicyViewProperty) =>
        row.replica?.toLowerCase() === replica || (replica === "master" && row.replica === "leader");
      return this.policyData.filter(replicaMatch);
    },
    columns(): Array<ITableColumn> {
      switch (this.policyTable) {
        case EPolicyGroup.Storage:
          return this.dataSourceColumns;
        case EPolicyGroup.General:
        case EPolicyGroup.Compute:
        case EPolicyGroup.Environment:
          return this.defaultColumns;
        default:
          return [];
      }
    },
    defaultColumns(): Array<ITableColumn> {
      if (this.isDistributed) {
        return distributedPreviewPolicyColumns;
      }
      return workloadPreviewPolicyColumns;
    },
    dataSourceColumns(): Array<ITableColumn> {
      return this.isDistributed
        ? distributedPreviewPolicyColumnsDataSourceTable
        : workloadPreviewPolicyColumnsDataSourceTable;
    },
    policyFilteredByComplianced(): Array<IPolicyViewProperty> {
      if (!this.policyData.length || !this.complianceData) return [];

      const filteredRules: Array<IPolicyViewProperty> = this.policyData.filter((row: IPolicyViewProperty) => {
        const isRelevantField = this.isDataSourceField
          ? this.complianceKeys.has(`${this.complianceData?.kind}.${row.name}`)
          : this.complianceKeys.has(row.name);
        if (this.isDistributed) {
          return (
            row.replica?.includes(this.complianceData?.distributedReplicaType?.toLocaleLowerCase() || "") &&
            isRelevantField
          );
        }
        return isRelevantField;
      });

      return filteredRules;
    },
    isDataSourceField(): boolean {
      if (!this.complianceData) return false;
      return (
        this.complianceData?.kind === AssetKind.ConfigMap ||
        this.complianceData?.kind === AssetKind.HostPath ||
        this.complianceData?.kind === AssetKind.Git ||
        this.complianceData?.kind === AssetKind.Nfs ||
        this.complianceData?.kind === AssetKind.S3 ||
        this.complianceData?.kind === AssetKind.Pvc ||
        this.complianceData?.kind === AssetKind.SecretVolume
      );
    },
    complianceKeys(): Set<string> {
      if (!this.complianceData?.complianceInfo?.reason) return new Set<string>();

      return new Set<string>(
        this.complianceData.complianceInfo.reason
          .map((reason: ComplianceInfoReason) => {
            const field = this.removeAttributesFromField(reason.field || "");

            if (!field) return "";

            const fieldAfterFirstDot = stringUtil.getStringAfterFirstDot(field);

            if (fieldAfterFirstDot === "attributes" || fieldAfterFirstDot === "instances") return field;
            return fieldAfterFirstDot;
          })
          .filter((field: string) => field),
      );
    },
  },
  methods: {
    removeAttributesFromField(field: string): string {
      const attributesString = "attributes";
      const hasAttributesInField: boolean = field?.includes(attributesString);

      if (!hasAttributesInField) return field;

      return field
        .split(".")
        .filter((s: string) => s !== attributesString)
        .join(".");
    },
  },
});
</script>
