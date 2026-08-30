<template>
  <onboarding-content-wrapper class="cluster-location-substep">
    <template #title>Set your cluster location</template>
    <template #subtitle>
      The <a :href="DOCS.INSTALLATION.CLUSTER" target="_blank">control plane</a> is where resource and workload
      management and monitoring happens on NVIDIA Run:ai. <br />
      Choose whether to run your cluster on the control plane's Kubernetes cluster or a different one.
    </template>
    <template #content>
      <div class="onboarding-select-options">
        <!-- SAME AS CONTROL PLANE -->
        <runai-select-card
          :value="clusterLocation === EClusterLocations.CONTROL_PLANE"
          @clicked="selectLocation(EClusterLocations.CONTROL_PLANE)"
          class="onboarding-select-card"
          :disabled="readOnly"
        >
          <div class="option-icon">
            <q-icon name="ph-bold ph-plug" size="40px" />
          </div>
          <div class="option-content">
            <h3 class="option-title">Same as the control plane</h3>
            <p class="option-description">
              The NVIDIA Run:ai cluster will be installed on the same Kubernetes cluster as the NVIDIA Run:ai control
              plane
            </p>
          </div>
        </runai-select-card>

        <!-- REMOTE -->
        <runai-select-card
          :value="clusterLocation === EClusterLocations.REMOTE"
          @clicked="selectLocation(EClusterLocations.REMOTE)"
          class="onboarding-select-card"
          :disabled="readOnly"
        >
          <div class="option-icon">
            <q-icon name="ph-bold ph-wifi-high" size="40px" />
          </div>
          <div class="option-content">
            <h3 class="option-title">Remote control plane</h3>
            <p class="option-description">
              The NVIDIA Run:ai cluster will be installed on a different Kubernetes cluster than the NVIDIA Run:ai
              control plane
            </p>
          </div>
        </runai-select-card>
      </div>

      <!--  HIDDEN FIELD FOR VALIDATION -->
      <q-field
        class="hidden-field onboarding-select-error"
        :rules="[notEmpty]"
        no-error-icon
        :model-value="clusterLocation"
      />
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import DOCS from "@/constants/docs";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { RunaiSelectCard } from "@/components/common/runai-select-card";
import { EClusterLocations } from "@/models/cluster.model";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "cluster-location-substep",
  components: { OnboardingContentWrapper, RunaiSelectCard },
  emits: {
    "update:cluster-location": (location: string) => location,
  },
  props: {
    clusterLocation: {
      type: String as PropType<string>,
      required: true,
    },
    readOnly: {
      type: Boolean,
      default: false,
    },
  },
  setup(_, { emit }) {
    const selectLocation = (location: EClusterLocations) => {
      emit("update:cluster-location", location);
    };
    const notEmpty = (val: string) => {
      return isNotEmpty(val) || errorMessages.SELECT_A_CLUSTER_LOCATION;
    };

    return { notEmpty, selectLocation, DOCS, EClusterLocations };
  },
});
</script>
