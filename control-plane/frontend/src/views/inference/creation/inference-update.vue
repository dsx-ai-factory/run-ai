<template>
  <section class="inference-hybrid column items-center q-pt-md">
    <template v-if="isPageReady">
      <workload-form-wrapper
        :form-state="inference"
        :submitting="submitting"
        submit-btn-text="Update Inference"
        cancel-btn-text="Cancel"
        @canceled="onCancel"
        @submit="submitInference"
      >
        <title-and-back-section label="Cluster" :summary="clusterName" />

        <title-and-back-section label="Project" :summary="projectName" aid="project-section" />

        <title-and-back-section v-if="showInferenceType" label="Inference type" summary="Custom" />

        <title-and-back-section label="Inference name" :summary="inference.name" />

        <hybrid-environment-section
          :workload-spec="inference.spec"
          :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
          :policy="policy"
          :cluster-id="inference.clusterId"
          :support-config-map-as-env-var="supportConfigMapAsEnvVar"
          :supporting-serving-port-access="supportingServingPortAccess"
          :support-image-pull-secrets="supportImagePullSecrets"
          @update="updateModelSpec"
          :show-content-header="false"
          :disable-fields="disableEnvironmentFields"
          supporting-serving-port
        />

        <hybrid-compute-resource-section
          :workload-spec="inference.spec"
          :project-node-types="selectedProjectNodeTypes"
          :node-pools-resources="projectNodePoolsResources"
          :policy="policy"
          :support-multi-fraction-gpu="supportMultiFractionGpu"
          :cluster-id="inference.clusterId"
          :auto-scale-options="autoScaleOptions"
          :support-tolerations="supportTolerations"
          section-sub-title="Set the resources for each replica"
          @update="inference.spec = $event"
          :show-content-header="false"
          :disable-fields="disableComputeFields"
        />

        <hybrid-data-source-section
          :default-opened="false"
          :workload-storage="inference.spec.storage"
          :policy-rules="policy?.rules?.storage"
          :policy-defaults="policy?.defaults?.storage"
          :cluster-id="inference.clusterId"
          :secrets="secrets"
          :can-add-new="false"
          :show-content-header="false"
          section-disabled
          :show-volume-persistency-section="false"
          :disable-fields="disableDataSourceSectionFields"
        />

        <general-section
          aid="general-section"
          should-auto-validate
          :workload-form-type="EWorkloadFormType.Inference"
          :general-model="generalSectionModel"
          :show-backoff-limit="false"
          :policy-rules="policy?.rules"
          :itemize-defaults="generalSectionItemizeDefaults"
          @general-model-changed="updateGeneralModel"
          :disable-fields="disableGeneralSectionFields"
          :show-request-timeout="showRequestTimeout"
        />
      </workload-form-wrapper>
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";

// components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";
import { HybridEnvironmentSection } from "@/components/workload-creation/hybrid-environment-section";
import { HybridComputeResourceSection } from "@/components/workload-creation/hybrid-compute-resource-section";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { GeneralSection } from "@/components/section/general-section";

// models
import { EWorkloadFormType, type IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import { Scope, type StringOption } from "@/swagger-models/assets-service-client";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type {
  UnifiedPolicyInfoPerReplicaType,
  Inference as InferenceFlat,
  InferenceUpdateSpecSpec,
  UnifiedWorkload,
} from "@/swagger-models/workloads-client";
import { SettingKeys } from "@/models/setting.model";
import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list/runai-itemized-list.model";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import type { IHybridEnvironmentSectionDisableFields } from "@/components/workload-creation/hybrid-environment-section/hybrid-environment-section.model";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type { IGeneralSectionDisableFields } from "@/components/section/general-section/general-section.model";
import type { IHybridDataSourceSectionDisableFields } from "@/components/workload-creation/hybrid-data-source-section/hybrid-data-source-section.model";
import type { IHybridComputeSectionDisableFields } from "@/components/workload-creation/hybrid-compute-resource-section/hybrid-compute-resource-section.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { useAuthStore } from "@/stores/auth.store";
import { useDataSourceStore } from "@/stores/data-source.store";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";

// utils
import { alertUtil } from "@/utils/alert.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";

// routes
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// composables
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";

// consts
import {
  MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED,
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_REQUEST_TIMEOUT,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";

// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// others
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";

const disableMessage = "This can't be modified when updating an inference workload";

export default defineComponent({
  name: "inference-update",
  components: {
    TitleAndBackSection,
    WorkloadFormWrapper,
    HybridEnvironmentSection,
    HybridComputeResourceSection,
    HybridDataSourceSection,
    GeneralSection,
  },
  provide() {
    return {
      policy: computed(() => this.policy),
    };
  },
  data() {
    return {
      appStore: useAppStore(),
      inferenceStore: useInferenceStore(),
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      inferenceFlat: {} as InferenceFlat,
      inference: {} as IUISuperWorkloadCreationRequest,
      submitting: false as boolean,
      clusterName: "" as string,
      showInferenceType: false as boolean,
      selectedProject: null as Project | null,
      policy: undefined as UnifiedPolicyInfoPerReplicaType | undefined | null,
      projectNodePoolsResources: [] as StringOption[],
      selectedProjectNodeTypes: [] as string[],
      autoScaleOptions: {
        latencyEnabled: true,
        isAutoScaleEnabled: true,
      } as IAutoScaleOptions,
      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      EWorkloadFormType,
      supportImagePullSecrets: true as boolean,
    };
  },
  setup() {
    const { environmentVariablesSourceValueOptions } = useEnvironmentVariablesSection();
    const { getUnifiedHybridWorkload } = useHybridTemplates(EHybridFormTypes.Inference);

    return {
      environmentVariablesSourceValueOptions,
      getUnifiedHybridWorkload,
    };
  },
  async created() {
    const { updateWorkloadId } = this.$route.query;
    this.showInferenceType =
      this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog) ||
      this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);
    try {
      if (!updateWorkloadId) {
        this.$router.back();
        return;
      }
      this.inferenceFlat = await inferenceService.getFlatById(updateWorkloadId?.toString());
      this.selectedProject = await orgUnitService.getProject(this.inferenceFlat.projectId);
      await this.initModel(this.inferenceFlat);
      await Promise.all([this.loadAvailableDataSource()]);
      this.supportImagePullSecrets = this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS,
      );
      this.appStore.setPageLoading(false);
    } catch (error: unknown) {
      this.$q.notify(alertUtil.getError("Failed to load assets"));
      console.error(error);
      this.appStore.setFallback(true);
    }
  },
  computed: {
    showRequestTimeout(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_INFERENCE_REQUEST_TIMEOUT,
      );
    },
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    projectName(): string {
      return this.selectedProject?.name || "-";
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportsNfsAndHostPath(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
      );
    },
    generalSectionModel(): IUIGeneralSectionModel {
      const generalSectionModel: IUIGeneralSectionModel = {
        annotations: (this.inference.spec?.annotations as IItemizedListItem[]) || undefined,
        labels: (this.inference.spec?.labels as IItemizedListItem[]) || undefined,
      };

      if (this.showRequestTimeout) {
        generalSectionModel.servingConfiguration = this.inference.spec?.servingConfiguration;
      }

      return generalSectionModel;
    },
    generalSectionItemizeDefaults(): ILoadedItemizeDefaults {
      return {
        annotations: this.policy?.defaults?.annotations,
        labels: this.policy?.defaults?.labels,
      };
    },
    disableEnvironmentFields(): IHybridEnvironmentSectionDisableFields {
      return {
        servingPort: disableMessage,
        tools: disableMessage,
        toolsAccess: disableMessage,
        security: {
          uidGidSource: disableMessage,
          capabilities: disableMessage,
        },
      };
    },
    disableComputeFields(): IHybridComputeSectionDisableFields {
      return {
        tolerations: disableMessage,
      };
    },
    disableDataSourceSectionFields(): IHybridDataSourceSectionDisableFields {
      return {
        pvc: disableMessage,
        git: disableMessage,
        hostPath: disableMessage,
        nfs: disableMessage,
        volume: disableMessage,
        configMap: disableMessage,
      };
    },
    disableGeneralSectionFields(): IGeneralSectionDisableFields {
      return {
        annotations: disableMessage,
        labels: disableMessage,
      };
    },
    supportingServingPortAccess(): boolean {
      return this.inferenceStore.supportingServingPortAccess(this.inference.clusterId);
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.inference.clusterId, MIN_CLUSTER_VERSION_FOR_TOLERATIONS);
    },
  },
  methods: {
    async initModel(inference: InferenceFlat): Promise<void> {
      const unifiedWorkload: UnifiedWorkload = await this.getUnifiedHybridWorkload({
        workloadId: inference.workloadId,
        projectId: inference.projectId,
        clusterId: inference.clusterId,
      });

      this.policy = unifiedWorkload.policyInfo?.worker;
      this.projectNodePoolsResources = unifiedWorkload.policyInfo?.worker?.rules?.nodePools?.options || [];
      const uiModel = inferenceWorkloadUtil.getHybridWorkloadUIModel(inference);
      this.handleClusterDependencies(inference.clusterId);
      this.inference = inferenceWorkloadUtil.setInferenceUIDefaults(uiModel, this.authStore.isSSO, this.policy);
    },
    async loadSecrets(): Promise<void> {
      try {
        this.secrets = await credentialService.listExistingSecrets(
          Scope.Project,
          this.inference.projectId,
          undefined,
          "all",
          true,
        );
      } catch (error: unknown) {
        console.error("Failed to load secrets:\n", error);
        this.$q.notify(alertUtil.getError("Failed to load secrets"));
      }
    },
    updateModelSpec(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.inference.spec = spec;
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.inference.spec = {
        ...this.inference.spec,
        ...generalModel,
      };
    },
    async submitInference(): Promise<void> {
      try {
        this.submitting = true;
        const workloadUpdateRequest: InferenceUpdateSpecSpec = inferenceWorkloadUtil.prepareUpdateInferenceFlatRequest(
          this.inferenceFlat,
          this.inference,
        );

        await inferenceService.updateFlatWorkload(this.inferenceFlat.workloadId, workloadUpdateRequest);
        this.$q.notify(alertUtil.getSuccess(`Inference ${this.inference.name} updated`));
        this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query: { clusterId: this.inference.clusterId } });
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedUpdateMessage("inference"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error("Failed to update inference:\n", error);
      } finally {
        this.submitting = false;
      }
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.$router.back();
      }
    },
    getClusterById(clusterId: string): ClustersMinimalData | undefined {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === clusterId);
    },
    updateFeaturesByClusterVersion(cluster: ClustersMinimalData): void {
      this.autoScaleOptions.latencyEnabled = inferenceWorkloadUtil.isLatencyEnableOnCluster(cluster);
      if (this.clusterStore.isClusterVersionSufficient(cluster.uuid, MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED)) {
        this.autoScaleOptions.isAutoScaleEnabled = inferenceWorkloadUtil.isAutoScaleEnableOnCluster(cluster);
      }
    },
    handleClusterDependencies(clusterId: string): void {
      const cluster = this.getClusterById(clusterId);
      if (cluster) {
        this.clusterName = cluster.name;
        this.updateFeaturesByClusterVersion(cluster);
      }
    },
    async loadAvailableDataSource(): Promise<void> {
      this.allowedDatasourceOptions = await this.dataSourceStore.getAvailableDataSources();
      if (this.dataSourceStore.fallbackPolicyLoadingError) this.appStore.setFallback(true);
    },
  },
});
</script>
