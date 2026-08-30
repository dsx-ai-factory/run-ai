<template>
  <section class="policy-edit" :class="submitting ? 'form-disabled' : ''">
    <runai-default-loader v-if="initialLoading" />
    <policy-creation-form
      v-if="!initialLoading"
      :loading-policy="loadingPolicy"
      :initial-policy="policyModel"
      :is-edit="isEdit"
      :allowed-scopes="allowedUpdateScopes"
      :existing-policy-type="existingPolicyType"
      :initial-gvk-policy="loadedGvkPolicy"
      :submitting="submitting"
      @cancel="onCancel"
      @load-policy="loadPolicyByTypeAndMeta"
      @load-gvk-policy="loadGvkPolicyForSelection"
      @submit="onSubmit"
      @submit-any-workload="onSubmitAnyWorkload"
    ></policy-creation-form>
    <policy-sync-failure-modal
      v-if="policySyncFailure"
      @close="onSyncFailureClose"
      @download="downloadPolicyYaml"
    ></policy-sync-failure-modal>
  </section>
</template>

<script lang="ts">
import { useAppStore } from "@/stores/app.store";
import { defineComponent } from "vue";
import { PolicyCreationForm } from "@/components/policy/policy-creation-form";
import {
  PolicyType,
  type PolicyMeta,
  type TrainingPolicyV2,
  type WorkspacePolicyV2,
  type DistributedPolicyV2,
} from "@/swagger-models/policy-service-client";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import { gvkPolicyService } from "@/services/control-plane/gvk-policy.service/gvk-policy.service";
import type { CreateGvkPolicyReq, GetGvkPolicyResp, GVK } from "@/swagger-models/policy-service-client";
import { usePolicyStore, type IPolicy, type ISyncedPolicyStatus } from "@/stores/policy.store";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { POLICIES_ROUTE_NAMES } from "@/router/policy.routes/policy.routes.names";
import { alertUtil } from "@/utils/alert.util";
import { policyUtil } from "@/utils/policy.util";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";
import { intervalUtil } from "@/utils/interval.util";
import { PolicySyncFailureModal } from "@/components/policy/policy-sync-failure-modal";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import fileDownload from "js-file-download";
import { yamlUtil } from "@/utils/yaml.util";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import type { TPolicyChangeRequest, TPolicyDefaultAndRules, TWorkloadPolicy } from "@/models/policy.model";
import { EIntervalLabels } from "@/models/interval.model";

import DOMPurify from "dompurify";
import { ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { DOCS } from "@/constants/docs";
import { docsUtil } from "@/utils/docs.util/docs.util";

export interface IPolicyMetaAndType {
  policyType: PolicyType;
  policyMeta: PolicyMeta;
}

export interface IPolicyModelAndType {
  policyModel: TWorkloadPolicy;
  policyType: PolicyType;
}

export default defineComponent({
  components: {
    PolicyCreationForm,
    PolicySyncFailureModal,
    RunaiDefaultLoader,
  },
  data() {
    return {
      policyStore: usePolicyStore(),
      appStore: useAppStore(),
      policyId: (this.$route.params.id?.toString() || "") as string,
      policyModel: policyService.getEmptyPolicyModel() as TWorkloadPolicy,
      originalPolicy: null as TWorkloadPolicy | null,
      loadedGvkPolicy: null as GetGvkPolicyResp | null,
      gvkPolicyId: "" as string,
      loadingPolicy: false as boolean,
      initialLoading: true as boolean,
      isEdit: false as boolean,
      submitting: false as boolean,
      policySyncFailure: false as boolean,
      failedGvkReq: null as CreateGvkPolicyReq | null,
      ENotificationBarVariant,
      DOCS,
    };
  },
  async created() {
    await this.loadPolicies();
    if (this.policyId) {
      const nativePolicy = this.policyStore.getPolicyByPolicyId(this.policyId);
      // Store list holds both native and GVK rows, so id shape (not membership) discriminates them.
      if (!policyUtil.isGvkPolicyId(this.policyId) && nativePolicy?.meta && nativePolicy.type) {
        await this.loadExistingPolicy();
        if (nativePolicy.meta.name) {
          this.appStore.setPageTitleOverride(`Edit policy: ${nativePolicy.meta.name}`);
        }
      } else {
        await this.loadExistingGvkPolicy();
      }
      this.isEdit = true;
    }
    this.initialLoading = false;
    useAppStore().setPageLoading(false);
  },
  computed: {
    allowedUpdateScopes(): PermittedScopes {
      return this.policyStore.actionPermissionsByScopes[Action.Update];
    },
    existingPolicyType(): PolicyType | null {
      if (policyUtil.isGvkPolicyId(this.policyId as string)) return null;
      return this.policyStore.getPolicyByPolicyId(this.policyId as string)?.type || null;
    },
  },
  methods: {
    async loadPolicies(): Promise<void> {
      try {
        await this.policyStore.loadPolicies();
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("policies"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        useAppStore().setFallback(true);
      }
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      allowToLeave && this.$router.push({ name: this.$route.meta.backPageName as string });
    },
    async loadExistingPolicy(): Promise<void> {
      if (!this.policyId) return;
      const policy: IPolicy | null = this.policyStore.getPolicyByPolicyId(this.policyId as string);
      if (!policy || !policy.meta || !policy.type) return;
      await this.loadPolicyByTypeAndMeta({ policyType: policy.type, policyMeta: policy.meta });
    },
    async loadExistingGvkPolicy(): Promise<void> {
      if (!this.policyId) return;
      try {
        this.loadingPolicy = true;
        const policy = await gvkPolicyService.getGvkPolicy(this.policyId);
        this.loadedGvkPolicy = policy;
        this.gvkPolicyId = this.policyId;
        this.appStore.setPageTitleOverride(`Edit policy: ${policy.workloadType.kind}`);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage("policy") });
        this.$q.notify(errorAlert.getNotification(error));
        useAppStore().setFallback(true);
      } finally {
        this.loadingPolicy = false;
      }
    },
    // Create mode: seed the form with any existing policy for the scope+GVK pair to avoid a duplicate error.
    async loadGvkPolicyForSelection(req: { workloadType: GVK; projectId: string } | null): Promise<void> {
      if (this.isEdit) return;
      if (!req) {
        this.loadedGvkPolicy = null;
        this.gvkPolicyId = "";
        return;
      }
      try {
        this.loadingPolicy = true;
        const { items } = await gvkPolicyService.listGvkPolicies();
        const targetKey = gvkPolicyUtil.gvkKey(req.workloadType);
        const match = items.find(
          (policy) =>
            gvkPolicyUtil.gvkKey(policy.workloadType) === targetKey && String(policy.projectId) === req.projectId,
        );
        if (!match?.id) {
          this.loadedGvkPolicy = null;
          this.gvkPolicyId = "";
          return;
        }
        this.loadedGvkPolicy = await gvkPolicyService.getGvkPolicy(match.id);
        this.gvkPolicyId = match.id;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage("policy") });
        this.$q.notify(errorAlert.getNotification(error));
        this.loadedGvkPolicy = null;
        this.gvkPolicyId = "";
      } finally {
        this.loadingPolicy = false;
      }
    },
    async loadPolicyByTypeAndMeta(policyMetaAndType: IPolicyMetaAndType): Promise<void> {
      try {
        this.loadingPolicy = true;
        const loadedPolicy = await policyService.loadPolicyByMetaAndType(
          policyMetaAndType.policyType,
          policyMetaAndType.policyMeta,
        );
        this.policyModel = loadedPolicy;
        this.originalPolicy = loadedPolicy;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("policy"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        useAppStore().setFallback(true);
      } finally {
        this.loadingPolicy = false;
      }
    },
    async onSubmit(updatedPolicy: IPolicyModelAndType): Promise<void> {
      if (!updatedPolicy.policyType || !updatedPolicy.policyModel) return;
      try {
        this.submitting = true;
        this.appStore.setIsAppHeaderDisabled(true);
        const policyToUpdate = this.getPreparedPolicyForRequest(updatedPolicy);
        // This is the policy created in the control plane, but it does not mean it was synced to the cluster
        const modifiedControlPlanePolicy = await policyService.updatePolicy(updatedPolicy.policyType, policyToUpdate);

        // Checking that the policy is synced to the cluster, if it fails we will have to revert back to the previous policy. (i.e: this.originalPolicy)
        const policyId = policyUtil.createPolicyIdFromPolicyTypeAndMeta(
          updatedPolicy.policyType,
          modifiedControlPlanePolicy.meta,
        );

        if (!modifiedControlPlanePolicy.meta?.clusterId) throw new Error("Cluster ID is missing in the policy meta");

        const syncStatus = await this.checkForPolicySync(policyId, modifiedControlPlanePolicy.meta.clusterId);
        // If sync was successful we proceed as planned, if it failed we revert changes in the BE only, so the user can still edit his work
        if (syncStatus.isSycned && !syncStatus.errorMessage) {
          this.$q.notify(alertUtil.getSuccess(`Policy saved`));
          this.$router.push({
            name: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
            query: {
              createdEntityId: policyId,
            },
          });
        } else {
          if (syncStatus.errorMessage) {
            const safeError = DOMPurify.sanitize(syncStatus.errorMessage, { ALLOWED_TAGS: [], ALLOWED_ATTR: [] });
            const message = DOCS.SUPPORT.URL
              ? `Policy couldn't be saved because of the following: ${safeError}. Try again. If the issue persists, contact ${docsUtil.getSupportAnchor()}.`
              : `Policy couldn't be saved because of the following: ${safeError}. Try again.`;
            this.$q.notify({ ...alertUtil.getError(message), timeout: 0, html: true });
          }

          // Updating the model so the user will be able to download its current draft
          this.policyModel = updatedPolicy.policyModel;
          this.policySyncFailure = true;
          await this.revertPolicy(updatedPolicy.policyType);
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedUpdateMessage("policy"),
        });
        const removePattern = /line.*?:/g;
        this.$q.notify({ ...errorAlert.getNotification(error, removePattern), timeout: 0 });
      } finally {
        this.appStore.setIsAppHeaderDisabled(false);
        this.submitting = false;
      }
    },
    async onSubmitAnyWorkload(req: CreateGvkPolicyReq): Promise<void> {
      try {
        this.submitting = true;
        this.appStore.setIsAppHeaderDisabled(true);
        let savedPolicy;
        if (this.gvkPolicyId) {
          // PUT replaces rules; the form can't edit description, so keep the loaded one rather than wiping it.
          savedPolicy = await gvkPolicyService.applyGvkPolicy(this.gvkPolicyId, {
            rules: req.rules,
            description: req.description ?? this.loadedGvkPolicy?.description,
          });
        } else {
          savedPolicy = await gvkPolicyService.createGvkPolicy(req);
        }
        this.$q.notify(alertUtil.getSuccess("Policy saved"));
        this.$router.push({
          name: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
          query: savedPolicy.id ? { createdEntityId: savedPolicy.id } : undefined,
        });
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: this.gvkPolicyId
            ? ErrorAlert.failedUpdateMessage("policy")
            : ErrorAlert.failedCreateMessage("policy"),
        });
        const removePattern = /line.*?:/g;
        this.$q.notify({ ...errorAlert.getNotification(error, removePattern), timeout: 0 });
        // Only network/unknown failures get the "couldn't be saved" dialog + YAML draft; API errors stay toast-only.
        if (gvkPolicyUtil.isNetworkOrUnknownError(error)) {
          this.failedGvkReq = req;
          this.policySyncFailure = true;
        }
      } finally {
        this.appStore.setIsAppHeaderDisabled(false);
        this.submitting = false;
      }
    },
    async checkForPolicySync(policyId: string, clusterId: string): Promise<ISyncedPolicyStatus> {
      // Here we continually check whether the policy was synced to the current cluster (polling of 2.5 seconds)
      // If it is synced we resolve and stop the interval,
      // if after 60 seconds it didnt sync we abort
      let syncTimedOut = false;
      const timeoutId = setTimeout(() => (syncTimedOut = true), 60000);
      return new Promise((resolve) => {
        intervalUtil.startInterval(EIntervalLabels.ApplyPolicySync, async () => {
          const syncStatus: ISyncedPolicyStatus = await this.policyStore.verifyPolicySync(policyId, clusterId);
          const isSynced: boolean = syncStatus.isSycned;
          if (isSynced || syncTimedOut) {
            intervalUtil.stopInterval(EIntervalLabels.ApplyPolicySync);
            clearTimeout(timeoutId);
            resolve(syncStatus); // If it is synced or timed-out we resolve. (if syncTimedOut is true, it would resolve({ syncStatus: false }))
          }
        });
      });
    },

    async revertPolicy(policyType: PolicyType) {
      // this is a silent revert (without announcements), we are reverting the change in the control-plane back to the previous policy
      if (!this.originalPolicy?.policy) return;
      try {
        // If this is a brand new policy we need to delete it as the backend does not allow us to send an empty policy
        if (!this.originalPolicy.policy.defaults && !this.originalPolicy.policy.rules) {
          if (!this.originalPolicy.meta) return;
          await policyService.deletePolicyByType(policyType, this.originalPolicy.meta);
          return;
        }
        const policyToUpdate = this.getPreparedPolicyForRequest({ policyType, policyModel: this.originalPolicy });

        await policyService.updatePolicy(policyType, policyToUpdate);
      } catch (error) {
        console.error("policy revert failed", error);
      }
    },

    onSyncFailureClose(): void {
      this.policySyncFailure = false;
      this.failedGvkReq = null;
    },
    async downloadPolicyYaml() {
      const draft: object = this.failedGvkReq
        ? { rules: this.failedGvkReq.rules }
        : ((this.policyModel.policy || {}) as TPolicyDefaultAndRules);
      const policyYaml: string = yamlUtil.objectToYaml(draft) || "";
      const policyName: string = dateUtil.dateFormat(new Date(), DATE_FORMATS.STANDARD_DATE_TIME);

      const blob = new Blob([policyYaml], { type: "application/yaml" });

      fileDownload(blob, `${policyName}.yaml`);
    },
    getPreparedPolicyForRequest(updatedPolicy: IPolicyModelAndType): TPolicyChangeRequest {
      return updatedPolicy.policyType === PolicyType.Distributed
        ? policyService.prepareDistributedDataForUpdateRequest(updatedPolicy.policyModel as DistributedPolicyV2)
        : policyService.prepareDataForUpdateRequest(updatedPolicy.policyModel as WorkspacePolicyV2 | TrainingPolicyV2);
    },
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});
</script>
<style lang="scss" scoped>
.form-disabled {
  opacity: 0.85;
  pointer-events: none;
}
</style>
