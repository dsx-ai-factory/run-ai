<template>
  <section class="distributed-hybrid-master column items-center q-pt-md">
    <workload-form-wrapper
      :form-state="hybridContext.hybridWorkloadModel.value"
      :submitting="submitting"
      submit-btn-text="Continue"
      cancel-btn-text="Back"
      :blurred="hybridContext.isDrawerOpen.value"
      init-dirty
      @unblur="hybridContext.closeDrawer"
      @canceled="onBack"
      @submit="onContinue"
      @leave-page="onLeave"
    >
      <hybrid-environment-section
        prevent-top-padding-when-open
        :key="`he-${hybridContext.selectedTemplates.value.environment.sectionKey}`"
        :render-key="hybridContext.selectedTemplates.value.environment.renderKey.toString()"
        :selected-name="hybridContext.selectedEnvironmentName.value"
        :workload-spec="hybridContext.hybridWorkloadModel.value.spec"
        :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
        :policy="policy"
        :cluster-id="formMetaData.clusterId"
        :required-fields="isTemplate ? {} : { image: true }"
        :support-config-map-as-env-var="supportConfigMapAsEnvVar"
        :support-user-credentials="supportUserCredentials"
        support-authenticated-ngc
        :scope-id="scopeModel.scopeId"
        :scope-type="scopeModel.scopeType"
        :support-image-pull-secrets="!hideSpecificSections?.environment?.imagePullSecrets && supportImagePullSecrets"
        :support-linking="isTemplate"
        :is-linked="!!hybridContext.selectedTemplates.value.environment.isLinked"
        :validate-empty-env-vars="!isTemplate"
        @load-section-templates="hybridContext.displayEnvironmentsList"
        @update="updateModelSpec"
        @reset="resetEnvironment"
        @clear="clearEnvironment"
        @credential-created="onCredentialCreated"
        @toggle-linked="hybridContext.toggleEnvironmentLinked"
      />

      <hybrid-compute-resource-section
        :key="`hc-${hybridContext.selectedTemplates.value['compute-resource'].sectionKey}`"
        :render-key="hybridContext.selectedTemplates.value['compute-resource'].renderKey.toString()"
        :selected-name="hybridContext.selectedComputeName.value"
        :workload-spec="hybridContext.hybridWorkloadModel.value.spec"
        :project-node-types="projectData.selectedProjectNodeTypes"
        :node-pools-resources="projectData.projectNodePoolsResources"
        :policy="policy"
        :support-multi-fraction-gpu="supportMultiFractionGpu"
        :support-tolerations="supportTolerations"
        :hide-specific-sections="hideSpecificSections?.computeResource"
        :support-linking="isTemplate"
        :is-linked="!!hybridContext.selectedTemplates.value['compute-resource'].isLinked"
        :hide-gpu-resources="hideGpuResources"
        section-sub-title="Set the resources for the master"
        @load-section-templates="hybridContext.displayComputesList"
        @update="updateModelSpec"
        @reset="resetCompute"
        @clear="clearCompute"
        @toggle-linked="hybridContext.toggleComputeLinked"
      />

      <hybrid-data-source-section
        :default-opened="shouldOpenDataSourceSection"
        :workload-storage="hybridContext.hybridWorkloadModel.value.spec.storage"
        @load-section-templates="hybridContext.displayDataSourcesList"
        @update="updateModelStorage"
        :secrets="secrets"
        :policy-rules="policy?.rules?.storage"
        :policy-defaults="policy?.defaults?.storage"
        :allowed-data-source-types="allowedDataSources"
        :cluster-id="formMetaData.clusterId"
      />

      <general-section
        should-auto-validate
        :workload-form-type="EWorkloadFormType.Distributed"
        :distributed-replica-type="EDistributedReplicaType.Master"
        :general-model="generalSectionModel"
        :policy-rules="policy?.rules"
        :itemize-defaults="generalSectionItemizeDefaults"
        :show-mpi-launcher-creation-policy="showMpiLauncherCreationPolicy"
        @general-model-changed="updateGeneralModel"
      />
    </workload-form-wrapper>

    <workload-creation-drawer
      :is-open="hybridContext.isDrawerOpen.value"
      :comply-to="drawerFilters"
      :entity="hybridContext.drawerDataEntity.value"
      :policy-details="policyViewDetails"
      :created-entity-id="hybridContext.recentlyCreatedEntityId.value"
      :allowed-data-source-options="allowedDataSources"
      @close="hybridContext.closeDrawer"
      @select="hybridContext.handleSelection"
      @select-multiple="hybridContext.handleMultiSelection"
      @create-new="onCreateNew"
    />
  </section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, type DefineComponent } from "vue";
// Components
import { WorkloadCreationDrawer } from "@/components/workload-creation/workload-creation-drawer";
import { HybridEnvironmentSection } from "@/components/workload-creation/hybrid-environment-section";
import { HybridComputeResourceSection } from "@/components/workload-creation/hybrid-compute-resource-section";
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { GeneralSection } from "@/components/section/general-section";
// Stores & Services
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { deepCopy } from "@/utils/common.util";
// Models
import { DistributedFramework } from "@/swagger-models/assets-service-client";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { EDistributedReplicaType, EWorkloadTypeSupportedInAssetsService } from "@/models/workload.model";
import { type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { EWorkloadFormType } from "@/models/workload.model";
import {
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { IUISuperWorkloadSpec, IUIStorageFields } from "@/models/workload.model";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import { type UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
import type { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import type { TProjectData } from "./distributed-hybrid-worker.vue";
import type { IHideSpecificSections } from "@/models/workload-hybrid.model";
import type { TDistributedFormMetaInfo } from "./distributed-hybrid.model";

const DistributedHybridMaster = defineComponent({
  name: "distributed-hybrid-master",
  components: {
    WorkloadCreationDrawer,
    HybridEnvironmentSection,
    HybridComputeResourceSection,
    WorkloadFormWrapper,
    HybridDataSourceSection,
    GeneralSection,
  },
  emits: ["submit", "continue", "update", "save-form-state", "back", "leave-page", "credential-created"],
  props: {
    hybridContext: {
      type: Object as PropType<ReturnType<typeof useHybridTemplates>>,
      required: true,
    },
    formMetaData: {
      type: Object as PropType<TDistributedFormMetaInfo>,
      required: true,
    },
    projectData: {
      type: Object as PropType<TProjectData>,
      required: true,
    },
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    policy: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaType | null>,
      required: false,
    },
    policyViewDetails: {
      type: Object as PropType<IPolicyViewDetails>,
      required: false,
    },
    environmentVariablesSourceValueOptions: {
      type: Array as PropType<TEnvVarNoneCustomSourceOption[]>,
      required: false,
    },
    secrets: {
      type: Array as PropType<SecretListInfo[]>,
      required: false,
      default: () => [],
    },
    allowedDataSources: {
      type: Object as PropType<Set<TDataSourceKinds>>,
      required: false,
    },
    isTemplate: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hideSpecificSections: {
      type: Object as PropType<IHideSpecificSections>,
      required: false,
    },
    supportImagePullSecrets: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    supportAuthenticatedNgc: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),

      EWorkloadFormType,
      EDistributedReplicaType,
    };
  },
  async created() {
    const { kind, createdEntityId, workloadFormName } = this.$route.query;

    if (workloadFormName === EDistributedReplicaType.Master) {
      const query = deepCopy(this.$route.query);
      delete query.workloadFormName;
      delete query.kind;
      delete query.createdEntityId;
      await this.$router.replace({
        query: { ...query },
      });
      this.hybridContext.handleCreatedAssets(kind?.toString(), createdEntityId?.toString());
    }

    this.hybridContext.createSectionsClearedStates();
  },
  computed: {
    scopeModel() {
      return {
        scopeType: this.hybridContext.hybridWorkloadModel.value.scopeType || "project",
        scopeId:
          this.hybridContext.hybridWorkloadModel.value.projectId?.toString() ||
          (this.hybridContext.hybridWorkloadModel.value.scopeId as string) ||
          "",
      };
    },
    shouldOpenDataSourceSection(): boolean {
      return (
        !!this.$route.query.createdEntityId &&
        [
          AssetKind.HostPath,
          AssetKind.Nfs,
          AssetKind.Pvc,
          AssetKind.Git,
          AssetKind.S3,
          AssetKind.ConfigMap,
          AssetKind.SecretVolume,
          AssetKind.DataVolume,
        ].includes(this.$route.query.kind as TDataSourceKinds)
      );
    },
    drawerFilters(): IAssetsFilter {
      if (this.isTemplate) {
        return {
          ...this.hybridContext.baseDrawerFilters.value,
          complyToReplicaType: "Master",
          distributedFramework: this.formMetaData.framework,
        };
      }

      return {
        projectId: +this.formMetaData.projectId,
        complyToProject: this.isTemplate ? undefined : +this.formMetaData.projectId,
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Distributed,
        complyToReplicaType: "Master",
        distributedFramework: this.formMetaData.framework,
        isDistributed: true,
      };
    },
    generalSectionModel(): IUIGeneralSectionModel {
      return {
        annotations:
          (this.hybridContext.hybridWorkloadModel.value.spec?.annotations as IItemizedListItem[]) || undefined,
        labels: (this.hybridContext.hybridWorkloadModel.value.spec?.labels as IItemizedListItem[]) || undefined,
        mpiLauncherCreationPolicy:
          this.hybridContext.hybridWorkloadModel.value.spec?.mpiLauncherCreationPolicy || undefined,
      };
    },
    generalSectionItemizeDefaults(): ILoadedItemizeDefaults {
      return {
        environmentVariables: this.policy?.defaults?.environmentVariables,
        annotations: this.policy?.defaults?.annotations,
        labels: this.policy?.defaults?.labels,
        tolerations: this.policy?.defaults?.tolerations,
      };
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formMetaData.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formMetaData.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
    supportUserCredentials(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formMetaData.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
      );
    },
    showMpiLauncherCreationPolicy(): boolean {
      return this.formMetaData.framework === DistributedFramework.Mpi;
    },
    hideGpuResources(): boolean {
      // The MPI master (launcher) replica cannot request GPUs, so hide the GPU fields and summary
      return this.formMetaData.framework === DistributedFramework.Mpi;
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formMetaData.clusterId,
        MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
      );
    },
  },
  methods: {
    onCredentialCreated(): void {
      this.$emit("credential-created");
    },
    updateParent(spec: IUISuperWorkloadSpec): void {
      this.$emit("update", spec);
    },
    updateModelSpec(spec: IUISuperWorkloadSpec): void {
      this.updateParent(spec);
    },
    updateModelStorage(storage: IUIStorageFields): void {
      const updatedSpec = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        storage,
      };
      this.updateParent(updatedSpec);
    },
    saveFormState(): void {
      this.$emit("save-form-state");
    },
    onLeave(): void {
      this.$emit("leave-page");
    },
    onSubmit(): void {
      this.$emit("submit");
    },
    onContinue(): void {
      this.$emit("continue");
    },
    onBack(): void {
      this.$emit("back");
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      const updatedSpec = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        annotations: generalModel.annotations,
        labels: generalModel.labels,
        mpiLauncherCreationPolicy: generalModel.mpiLauncherCreationPolicy,
      };

      this.updateParent(updatedSpec);
    },
    clearCompute(): void {
      this.hybridContext.clearComputeTemplate();
    },
    clearEnvironment(): void {
      this.hybridContext.clearEnvironmentTemplate();
    },
    resetEnvironment(): void {
      this.hybridContext.resetEnvironmentToDefault();
    },
    resetCompute(): void {
      this.hybridContext.resetComputeToDefault();
    },
    onCreateNew(type: string): void {
      this.saveFormState();
      this.hybridContext.onCreateNewSetup(type, {
        distFramework: this.formMetaData.framework,
        workloadType: EWorkloadFormType.Training,
        workloadFormName: EDistributedReplicaType.Master,
      });
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default DistributedHybridMaster as typeof DistributedHybridMaster & DefineComponent;
</script>
