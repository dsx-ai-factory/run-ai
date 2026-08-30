<template>
  <section class="view-policy column items-center q-pt-lg">
    <runai-expansion-item class="scope-section" label="Scope" hide-expend-icon>
      <template #subheader>
        <div class="row items-center">
          <span aid="policy-view-scope">{{ scopeName }}</span>
        </div>
      </template>
    </runai-expansion-item>

    <runai-expansion-item class="workload-type" label="Workload type" hide-expend-icon>
      <template #subheader>
        <div class="row items-center">
          <span aid="policy-view-workload-type">{{ workloadTypeText }}</span>
        </div>
      </template>
    </runai-expansion-item>

    <runai-expansion-item v-if="isGvk" label="Rules" default-opened :subheader="gvkRules.length + ' rules'">
      <gvk-policy-view :rules="gvkRules" />
    </runai-expansion-item>

    <template v-else-if="policy">
      <template v-for="table in tablesData">
        <runai-expansion-item
          :key="table.label"
          v-if="table.data.length"
          :label="table.label"
          :subheader="table.data.length + ' rules'"
        >
          <policy-details-table :policy-data="table.data" :policy-table="table.table" :is-distributed="isDistributed" />
        </runai-expansion-item>
      </template>
    </template>
    <section class="row justify-end q-mt-md">
      <q-btn label="Close" color="primary" @click="onClose" />
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useAppStore } from "@/stores/app.store";
import { EWorkloadType } from "@/models/workload.model";
import { policyUtil } from "@/utils/policy.util/policy.util";

import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import { gvkPolicyService } from "@/services/control-plane/gvk-policy.service/gvk-policy.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { PolicyDetailsTable } from "@/components/policy/policy-details-table";
import GvkPolicyView from "@/components/policy/policy-rule-builder/gvk-policy-view.vue";
import {
  PolicyType,
  ScopeType,
  type GetGvkPolicyResp,
  type EffectiveRulesEffectiveRulesInner,
  type DistributedPolicyDefaultsAndRulesV2,
  type TrainingPolicyDefaultsAndRulesV2,
  type WorkspacePolicyDefaultsAndRulesV2,
} from "@/swagger-models/policy-service-client";
import {
  getDistributedPolicyViewDetails,
  getPolicyViewDetails,
  EPolicyGroup,
  type IPolicyViewDetails,
  type IPolicyViewProperty,
} from "@/utils/policy.util/policy-view.util";
import { useAuthStore } from "@/stores/auth.store";

// Private types
export interface ITableData {
  label: string;
  table: EPolicyGroup;
  data: Array<IPolicyViewProperty>;
}

export default defineComponent({
  name: "view-policy",
  components: {
    RunaiExpansionItem,
    PolicyDetailsTable,
    GvkPolicyView,
  },
  data() {
    return {
      authStore: useAuthStore(),
      policyId: (this.$route.params.id?.toString() || "") as string,
      policy: null as IPolicyViewDetails | null,
      gvkPolicy: null as GetGvkPolicyResp | null,
      gvkEffectiveRules: [] as EffectiveRulesEffectiveRulesInner[],
    };
  },
  async created() {
    await this.loadPolicyByPolicyId();
    useAppStore().setPageLoading(false);
  },
  computed: {
    isGvk(): boolean {
      return policyUtil.isGvkPolicyId(this.policyId);
    },
    gvkRules(): EffectiveRulesEffectiveRulesInner[] {
      return this.gvkEffectiveRules;
    },
    scopeName(): string {
      if (this.isGvk) {
        if (!this.gvkPolicy) return "";
        if (this.gvkPolicy.scope === ScopeType.Project) {
          const name =
            this.authStore.getOrgUnitByScopeAndId(ScopeType.Project, String(this.gvkPolicy.projectId ?? ""))?.name || "";
          return `Project: ${name}`;
        }
        return this.gvkPolicy.scope === ScopeType.Cluster ? "Cluster" : "";
      }
      const policyScopeDetails = policyUtil.getPolicyTypeScopeAndScopeId(this.policyId);
      const orgUnit = this.authStore.getOrgUnitByScopeAndId(
        policyScopeDetails.scope,
        policyScopeDetails.scopeId?.toString() || "",
      );
      const scopeName = orgUnit?.name || "";

      switch (policyScopeDetails.scope) {
        case ScopeType.Tenant:
          return scopeName;
        case ScopeType.Department:
          return `Department: ${scopeName}`;
        case ScopeType.Project:
          return `Project: ${scopeName}`;
        case ScopeType.Cluster:
          return `Cluster: ${scopeName}`;
        default:
          return "";
      }
    },
    workloadTypeText(): string {
      if (this.isGvk) return this.gvkPolicy?.workloadType.kind ?? "";
      switch (this.policyType) {
        case PolicyType.Distributed:
          return "Training (Distributed)";
        case PolicyType.Training:
          return "Training (Standard)";
        case PolicyType.Workspace:
          return EWorkloadType.Workspace;
        case PolicyType.Inference:
          return EWorkloadType.Inference;
        default:
          return "";
      }
    },
    tablesData(): ITableData[] {
      return [
        {
          label: "Compute resource",
          table: EPolicyGroup.Compute,
          data: Object.values(this.policy?.compute || {}),
        },
        {
          label: "Environment",
          table: EPolicyGroup.Environment,
          data: Object.values(this.policy?.environment || {}),
        },
        {
          label: "Data source",
          table: EPolicyGroup.Storage,
          data: Object.values(this.policy?.storage || {}),
        },
        {
          label: "General",
          table: EPolicyGroup.General,
          data: Object.values(this.policy?.general || {}),
        },
      ];
    },
    policyType(): PolicyType {
      return policyUtil.getPolicyTypeScopeAndScopeId(this.policyId).type;
    },
    isDistributed(): boolean {
      return this.policyType === PolicyType.Distributed;
    },
  },
  methods: {
    async loadPolicyByPolicyId(): Promise<void> {
      if (!this.policyId) {
        this.policy = {} as IPolicyViewDetails;
        return;
      }
      if (this.isGvk) {
        return await this.loadGvkPolicyById();
      }
      try {
        const policy = await policyService.loadPolicyById(this.policyId);
        if (!policy?.effective) {
          this.policy = {} as IPolicyViewDetails;
          return;
        }

        if (this.isDistributed) {
          this.policy = getDistributedPolicyViewDetails(policy.effective as DistributedPolicyDefaultsAndRulesV2);
        } else {
          this.policy = getPolicyViewDetails(
            policy.effective as WorkspacePolicyDefaultsAndRulesV2 | TrainingPolicyDefaultsAndRulesV2,
          );
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("policy"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        useAppStore().setFallback(true);
      }
    },
    async loadGvkPolicyById(): Promise<void> {
      try {
        this.gvkPolicy = await gvkPolicyService.getGvkPolicy(this.policyId);
        // View mode shows the effective (merged across scopes) rules, like native policies do, via the
        // dedicated endpoint — each rule carries the scope it was defined at.
        this.gvkEffectiveRules = await gvkPolicyService.getEffectivePolicy({
          workloadType: this.gvkPolicy.workloadType,
          scope: this.gvkPolicy.scope,
          projectId: this.gvkPolicy.projectId,
        });
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage("policy") });
        this.$q.notify(errorAlert.getNotification(error));
        useAppStore().setFallback(true);
      }
    },
    onClose(): void {
      if (this.$route?.query?.previousRoute) this.$router.push({ name: this.$route.query.previousRoute as string });
      else this.$router.push({ name: this.$route?.meta?.backPageName as string });
    },
  },
});
</script>
<style lang="scss" scoped>
.view-policy {
  > * {
    width: 800px;
  }
}
</style>
