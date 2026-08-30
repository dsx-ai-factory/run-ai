<template>
  <onboarding-content-wrapper>
    <template #title>Name your cluster</template>
    <template #subtitle>
      A cluster is your organization's compute infrastructure, where AI workloads are executed.<br />
      NVIDIA Run:ai connects to your <a :href="DOCS.INSTALLATION.CLUSTER" target="_blank">cluster</a> to run, manage, and
      monitor those workloads efficiently.
    </template>
    <template #content>
      <div class="row">
        <div class="col-5 cluster-name-wrapper">
          <runai-name-validation
            aid="cluster-name"
            :model-value="clusterName"
            @update:model-value="handleClusterNameChange"
            :is-invalid="isInvalid"
            class="cluster-name-input"
            label="Cluster name"
            @input="handleClusterNameChange"
            :rules="[isEmpty, isValidFormat, isNameUnique]"
            :read-only="readOnly"
          />
        </div>
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, ref, computed, watch } from "vue";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import DOCS from "@/constants/docs";
import { isNotEmpty, isValidRunaiEntityNameFormat } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { useClusterStore } from "@/stores/cluster.store";
export default defineComponent({
  name: "cluster-name-substep",
  components: { OnboardingContentWrapper, RunaiNameValidation },
  emits: {
    "update:cluster-name": (name: string) => name,
    "update:installation-option": (option: string) => option,
  },
  props: {
    clusterName: {
      type: String as PropType<string>,
      default: "",
    },
    installationOption: {
      type: String as PropType<string>,
      default: "",
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  setup(props, { emit }) {
    const selectedInstallationOption = ref(props.installationOption);

    // Validation functions
    const isEmpty = (value: string) => {
      return isNotEmpty(value) || errorMessages.NAME_NOT_EMPTY;
    };

    const isValidFormat = (value: string) => {
      return isValidRunaiEntityNameFormat(value) || errorMessages.VALID_FORMAT;
    };

    const isNameUnique = (value: string) => {
      if (props.readOnly) {
        return true; //Step already completed
      }
      return !useClusterStore().isClusterNameExists(value) || errorMessages.NAME_ALREADY_EXIST;
    };

    const isInvalid = computed(() => {
      return props.clusterName.length === 0 || props.clusterName.length > 63;
    });

    // Event handlers
    const handleClusterNameChange = (name: string) => {
      emit("update:cluster-name", name);
    };

    const selectInstallationOption = (option: string) => {
      selectedInstallationOption.value = option;
      emit("update:installation-option", option);
    };
    watch(
      () => props.installationOption,
      (newValue) => {
        selectedInstallationOption.value = newValue;
      },
    );

    return {
      selectedInstallationOption,

      // Validation
      isInvalid,
      isEmpty,
      isValidFormat,
      isNameUnique,
      // Handlers
      handleClusterNameChange,
      selectInstallationOption,

      DOCS,
    };
  },
});
</script>

<style scoped lang="scss">
.cluster-name-wrapper {
  max-width: 500px;
}
</style>
