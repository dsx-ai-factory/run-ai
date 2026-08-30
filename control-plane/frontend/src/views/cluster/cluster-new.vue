<template>
  <div class="runai-stepper-form q-pa-md row no-wrap justify-center">
    <div class="steps-map q-mt-sm">
      <q-stepper
        v-model="currentStep"
        active-icon="ph-fill ph-pencil-simple"
        vertical
        ref="stepper"
        color="primary"
        animated
        flat
        style="background-color: transparent"
      >
        <q-step
          v-for="(step, idx) in steps"
          :key="step.name"
          :name="step.name"
          :done="currentStep > idx + 1"
          :title="step.title"
          :icon="step.icon"
        />
      </q-stepper>
    </div>
    <div class="main-content cluster-creation" v-if="currentStep === 1">
      <div class="justify-center">
        <runai-form-wrapper class="cluster-form" :form-state="clusterFormModel" ref="elClusterForm">
          <cluster-creation-step
            v-if="clusterCreationStepLoading"
            :is-cluster-edit-mode="isClusterEditMode"
            :is-multi-tenant="isMultiTenant"
            :cluster-form-model="clusterFormModel"
            :versions="sortedVersions"
            :is-higher-cluster-version="isHigherClusterVersion"
            @update-form-model="onUpdateCreationFormModel"
          ></cluster-creation-step>
        </runai-form-wrapper>
      </div>
      <div class="stepper-controls row justify-between q-mt-md">
        <q-field class="col-4 form-hint no-padding" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
        <q-stepper-navigation class="q-pt-none">
          <runai-btn aid="continue-btn" @click="goNext" :loading="loading" color="primary" label="Continue" />
        </q-stepper-navigation>
      </div>
    </div>
    <div class="main-content cluster-installation" v-else>
      <div class="row justify-center">
        <cluster-installation-step
          v-if="createdCluster?.uuid && operatorVersionsMap"
          :cluster-form-model="clusterFormModel"
          :is-multi-tenant="isMultiTenant"
          :created-cluster="createdCluster"
          :versions-map="operatorVersionsMap"
          :is-higher-cluster-version="isHigherClusterVersion"
        ></cluster-installation-step>
      </div>
      <div class="stepper-controls row justify-end">
        <q-stepper-navigation>
          <runai-btn flat @click="goBack" aid="back-btn" color="primary" label="Back" class="q-mr-md" />
          <runai-btn aid="leave-btn" @click="leave" color="primary" label="Done" />
        </q-stepper-navigation>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { clusterNewInstallerThreshold, EClusterLocations, type IClusterInstalltionModel } from "@/models/cluster.model";
// stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import type { QStepper } from "quasar";
// cmps
import { ClusterCreationStep } from "@/components/cluster/cluster-creation-step";
import { ClusterInstallationStep } from "@/components/cluster/cluster-installation-step";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiBtn } from "@/components/common/runai-btn";
// services
import { isNewerVersion } from "@/utils/version.util";
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { clusterUtil } from "@/utils/cluster.util";
// models
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { HttpErrorResponse } from "@/models/http-response.model";
import { defineComponent } from "vue";
import { getVersionWithoutPatch } from "@/utils/version.util/version.util";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import DOCS from "@/constants/docs";

export default defineComponent({
  components: { ClusterCreationStep, ClusterInstallationStep, RunaiFormWrapper, RunaiBtn },
  data() {
    return {
      loading: false,
      clusterCreationStepLoading: false,
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      currentStep: 1,
      steps: [
        { title: "Setup", name: 1, icon: "ph-bold ph-clipboard-text" },
        { title: "Installation instructions", name: 2, icon: "ph-fill ph-download" },
      ],
      timeOutId: null as number | null,
      displayFormHint: false,
      createdCluster: null as null | ClustersMinimalData,
      isClusterEditMode: false as boolean,
      clusterFormModel: {
        name: "",
        version: "",
        prerequisites: {
          location: "",
          clusterUrl: "",
          k8sDistribution: "",
        },
      } as IClusterInstalltionModel,
      operatorVersionsMap: null as Record<string, string> | null,
    };
  },
  async created() {
    const clusterUuid = this.$route.params.id;
    await this.setOperatorVersionMap();
    if (clusterUuid) {
      this.setClusterDataEditMode(clusterUuid as string);
    } else {
      this.setClusterDataCreateMode();
    }
    this.appStore.setPageLoading(false);
    this.clusterCreationStepLoading = true;
  },
  computed: {
    versions(): string[] {
      if (!this.operatorVersionsMap) return [];
      return Object.keys(this.operatorVersionsMap);
    },
    sortedVersions(): string[] {
      return [...this.versions].sort((versionA, versionB) => (isNewerVersion(versionA, versionB) ? -1 : 1));
    },
    isMultiTenant(): boolean {
      return this.appStore.isMultiTenant;
    },
    isHigherClusterVersion(): boolean {
      if (!this.clusterFormModel.version) return false;
      return isNewerVersion(this.clusterFormModel.version, clusterNewInstallerThreshold);
    },
  },
  methods: {
    onUpdateCreationFormModel(newModel: IClusterInstalltionModel): void {
      const isUrlChanged = this.clusterFormModel.prerequisites.clusterUrl != newModel.prerequisites.clusterUrl;
      this.clusterFormModel = newModel;
      if (isUrlChanged) return;
      const selectedCluster: ClustersMinimalData | null = this.clusterStore.getClusterById(
        this.createdCluster?.uuid as string,
      );
      if (selectedCluster?.domain && this.clusterFormModel.prerequisites.location === EClusterLocations.REMOTE) {
        this.clusterFormModel.prerequisites.clusterUrl = selectedCluster?.domain;
      } else if (this.clusterFormModel.prerequisites.location !== EClusterLocations.REMOTE) {
        this.clusterFormModel.prerequisites.clusterUrl = "";
      }
    },
    setClusterDataEditMode(clusterId: string): void {
      const selectedCluster: ClustersMinimalData | null = this.clusterStore.getClusterById(clusterId);
      if (!selectedCluster) return;
      this.createdCluster = selectedCluster;
      this.clusterFormModel.name = selectedCluster.name;

      // Get the cluster's version without patch
      const clusterVersion = getVersionWithoutPatch(selectedCluster.version || "");

      // If cluster version exists in allowed versions, use it; otherwise use the latest version
      if (clusterVersion && this.versions.includes(clusterVersion)) {
        this.clusterFormModel.version = clusterVersion;
      } else {
        this.clusterFormModel.version = this.sortedVersions[0]; // sortedVersions[0] is the latest version
      }

      this.clusterFormModel.prerequisites.clusterUrl = selectedCluster.domain || "";
      this.clusterFormModel.prerequisites.location =
        selectedCluster.domain === window.location.hostname ? EClusterLocations.CONTROL_PLANE : EClusterLocations.REMOTE;
      this.isClusterEditMode = true;
    },
    setClusterDataCreateMode(): void {
      this.clusterFormModel.version = this.sortedVersions[0];
      this.clusterFormModel.prerequisites.location = this.isMultiTenant
        ? EClusterLocations.REMOTE
        : EClusterLocations.CONTROL_PLANE;
    },
    async setOperatorVersionMap(): Promise<void> {
      try {
        this.operatorVersionsMap = await clusterService.getOperatorVersions();
      } catch (error: unknown) {
        console.error("Failed to load cluster versions:", error);
        const failedLoadingMessage = `${ErrorAlert.failedLoadingMessage("cluster versions")}. Please try again`;
        const message = DOCS.SUPPORT.URL ? `${failedLoadingMessage} or contact support.` : `${failedLoadingMessage}.`;
        this.$q.notify(alertUtil.getError(message));
      }
    },
    async validate(): Promise<boolean> {
      return await (this.$refs["elClusterForm"] as HTMLFormElement).validate();
    },
    async validateAndNotify(): Promise<boolean> {
      this.displayFormHint = false;
      const success = await this.validate();
      if (!success) {
        this.showHint();
      }
      return success;
    },
    async createCluster(): Promise<void> {
      this.displayFormHint = false;
      const isValid: boolean = await this.validateAndNotify();
      if (!isValid) return;
      try {
        this.loading = true;
        if (this.createdCluster?.uuid) await this.deleteLastCreatedCluster();
        this.createdCluster = await this.clusterStore.create(this.clusterFormModel.name);
        this.createdCluster?.uuid && clusterUtil.addSessionCreatedCluster(this.createdCluster.uuid);
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create cluster"));
        }
      } finally {
        this.loading = false;
        (this.$refs.stepper as QStepper).next();
        window.scrollTo({ top: 0, behavior: "smooth" });
      }
    },
    async deleteLastCreatedCluster(): Promise<void> {
      if (!this.createdCluster?.uuid) return;
      try {
        this.loading = true;
        this.createdCluster?.uuid && clusterUtil.removeSessionCreatedCluster(this.createdCluster.uuid);
        await this.clusterStore.deleteCluster(this.createdCluster.uuid);
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to delete cluster"));
        }
      } finally {
        this.loading = false;
      }
    },
    async goNext(): Promise<void> {
      if (this.isClusterEditMode) {
        const isValid: boolean = await this.validateAndNotify();
        if (!isValid) return;
        (this.$refs.stepper as QStepper).next();
        return;
      }
      await this.createCluster();
    },
    goBack(): void {
      (this.$refs.stepper as QStepper).previous();
      window.scrollTo({ top: 0, behavior: "smooth" });
    },
    leave(): void {
      if (!this.isClusterEditMode && this.createdCluster?.name) {
        this.$q.notify(alertUtil.getSuccess(`Cluster ${this.createdCluster.name} added`));
      }
      this.$router.push({
        name: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
        query: { createdEntityId: this.createdCluster?.uuid },
      });
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = window.setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
  },
});
</script>
<style lang="scss">
.runai-stepper-form {
  position: relative;
  .steps-map {
    top: 80px;
    left: 0;
    position: fixed;

    .q-stepper__dot .q-icon {
      font-size: 16px;
    }
  }
  a:hover {
    text-decoration: underline;
  }
  .stepper-controls {
    .q-stepper__nav {
      margin-top: 0;
      padding-top: 0;
    }
  }
}
</style>
