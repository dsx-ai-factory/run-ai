<template>
  <section class="distributed-hybrid-setup column items-center q-pt-md">
    <workload-form-wrapper
      :form-state="hybridContext.hybridWorkloadModel.value"
      :submitting="submitting"
      :submit-btn-text="submitBtnText"
      cancel-btn-text="Back"
      :blurred="hybridContext.isDrawerOpen.value"
      init-dirty
      @unblur="hybridContext.closeDrawer"
      @canceled="onBack"
      @submit="onSubmit"
      @leave-page="onLeave"
    >
      <general-section
        default-opened
        should-auto-validate
        prevent-top-padding-when-open
        :show-annotations-and-labels="false"
        :workload-form-type="EWorkloadFormType.Distributed"
        :general-model="generalSectionModel"
        :show-backoff-limit="supportBackoffLimit"
        :policy-rules="policy?.rules"
        :itemize-defaults="generalSectionItemizeDefaults"
        :show-termination-grace-period="supportTerminationGracePeriod"
        :support-clean-pod-policy="supportCleanPodPolicy"
        :show-ssh-auth-mount-path="showSshAuthMountPath"
        :workload-priority-data="workloadPriorityData"
        :support-priority-class="supportPriorityClass"
        :support-preemptibility="supportPreemptibility"
        @general-model-changed="updateGeneralModel"
      />
    </workload-form-wrapper>
  </section>
</template>

<script setup lang="ts">
import { MIN_CLUSTER_VERSION_CLEAN_POD_POLICY } from "@/common/version.constant";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type {
  IUIGeneralSectionModel,
  IWorkloadPriorityData,
} from "@/components/section/general-section/general-section.model";
import type { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { EWorkloadFormType, type IUISuperWorkloadSpec } from "@/models/workload.model";
import { useClusterStore } from "@/stores/cluster.store";
import { DistributedFramework } from "@/swagger-models/assets-service-client";
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import { type ComputedRef } from "vue";
import { computed } from "vue";
import { GeneralSection } from "@/components/section/general-section";
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";

export interface Props {
  hybridContext: ReturnType<typeof useHybridTemplates>;
  submitting: boolean;
  policy?: UnifiedPolicyInfoPerReplicaType | null;
  workloadPriorityData?: IWorkloadPriorityData;
  supportPreemptibility?: boolean;
  supportPriorityClass?: boolean;
  supportTerminationGracePeriod?: boolean;
  isEditMode?: boolean;
  isTemplate?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  submitting: false,
  policy: null,
  workloadPriorityData: () => ({}),
});

const clusterStore = useClusterStore();

const emit = defineEmits<{
  (e: "update", payload: IUISuperWorkloadSpec): void;
  (e: "back"): void;
  (e: "leave-page"): void;
  (e: "close"): void;
  (e: "submit"): void;
}>();

const submitBtnText = computed(
  () => `${props.isEditMode ? "Save" : "Create"} ${props.isTemplate ? "Template" : "Training"}`,
);

const generalSectionModel: ComputedRef<IUIGeneralSectionModel> = computed(() => {
  return {
    autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
      props.hybridContext.hybridWorkloadModel.value.spec?.autoDeletionTimeAfterCompletionSeconds,
    )
      ? props.hybridContext.hybridWorkloadModel.value.spec?.autoDeletionTimeAfterCompletionSeconds
      : null,
    annotations: (props.hybridContext.hybridWorkloadModel.value.spec?.annotations as IItemizedListItem[]) || undefined,
    labels: (props.hybridContext.hybridWorkloadModel.value.spec?.labels as IItemizedListItem[]) || undefined,
    backoffLimit: Number.isInteger(props.hybridContext.hybridWorkloadModel.value.spec?.backoffLimit)
      ? props.hybridContext.hybridWorkloadModel.value.spec?.backoffLimit
      : null,
    terminationGracePeriodSeconds:
      props.hybridContext.hybridWorkloadModel.value.spec?.terminationGracePeriodSeconds || null,
    cleanPodPolicy: props.hybridContext.hybridWorkloadModel.value.spec?.cleanPodPolicy || undefined,
    sshAuthMountPath: props.hybridContext.hybridWorkloadModel.value.spec?.sshAuthMountPath || undefined,
    priorityClass: props.hybridContext.hybridWorkloadModel.value.spec.priorityClass || null,
    preemptibility: props.hybridContext.hybridWorkloadModel.value.spec?.preemptibility || null,
  };
});

const generalSectionItemizeDefaults = computed(() => {
  return {
    environmentVariables: props.policy?.defaults?.environmentVariables,
    annotations: props.policy?.defaults?.annotations,
    labels: props.policy?.defaults?.labels,
    tolerations: props.policy?.defaults?.tolerations,
  };
});

const supportCleanPodPolicy = computed(() => {
  return clusterStore.isClusterVersionSufficient(
    props.hybridContext.hybridWorkloadModel.value.clusterId,
    MIN_CLUSTER_VERSION_CLEAN_POD_POLICY,
  );
});

const showSshAuthMountPath = computed(() => {
  return props.hybridContext.hybridWorkloadModel.value.spec.distributedFramework === DistributedFramework.Mpi;
});

const supportBackoffLimit = computed(() => {
  return clusterStore.isClusterVersionSupportBackoffLimit(props.hybridContext.hybridWorkloadModel.value.clusterId);
});

const updateGeneralModel = (generalModel: IUIGeneralSectionModel): void => {
  const updatedSpec = {
    ...props.hybridContext.hybridWorkloadModel.value.spec,
    backoffLimit: generalModel.backoffLimit,
    terminationGracePeriodSeconds: generalModel.terminationGracePeriodSeconds,
    autoDeletionTimeAfterCompletionSeconds: generalModel.autoDeletionTimeAfterCompletionSeconds,
    cleanPodPolicy: generalModel.cleanPodPolicy,
    sshAuthMountPath: generalModel.sshAuthMountPath,
    priorityClass: generalModel.priorityClass,
    preemptibility: props.supportPreemptibility ? generalModel.preemptibility : undefined,
  };
  emit("update", updatedSpec);
};

const onLeave = (): void => {
  emit("leave-page");
};
const onSubmit = (): void => {
  emit("submit");
};
const onBack = (): void => {
  emit("back");
};
</script>
