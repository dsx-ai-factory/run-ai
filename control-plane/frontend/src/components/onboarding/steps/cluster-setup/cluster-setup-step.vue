<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="cluster-setup-step">
    <sub-step-transition>
      <cluster-name-substep
        v-if="subStep === 1"
        key="name-substep"
        :cluster-name="clusterData.clusterName"
        :installation-option="clusterData.installationOption"
        :read-only="clusterData.hasClusterValues"
        @update:cluster-name="(name: string) => handleClusterDataUpdate('clusterName', name)"
        @update:installation-option="(option: string) => handleClusterDataUpdate('installationOption', option)"
      />
      <cluster-location-substep
        v-else-if="subStep === 2 && showsClusterLocationStep"
        key="location-substep"
        :cluster-location="clusterData.clusterLocation"
        :read-only="clusterData.hasClusterValues"
        @update:cluster-location="(location: string) => handleClusterDataUpdate('clusterLocation', location)"
      />
      <cluster-url-substep
        v-else-if="(subStep === 2 && !showsClusterLocationStep) || (subStep === 3 && showsClusterLocationStep)"
        key="url-substep"
        :cluster-url="clusterData.clusterUrl"
        :read-only="clusterData.hasClusterValues || clusterData.clusterLocation === EClusterLocations.CONTROL_PLANE"
        @update:cluster-url="(url: string) => handleClusterDataUpdate('clusterUrl', url)"
      />
      <cluster-installation-substep
        v-else-if="(subStep === 3 && !showsClusterLocationStep) || (subStep === 4 && showsClusterLocationStep)"
        key="installation-substep"
        :cluster-data="clusterData"
        @update:is-cluster-connected="
          (isConnected: boolean) => handleClusterDataUpdate('isClusterConnected', isConnected)
        "
        @close="handleClose"
      />
    </sub-step-transition>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { ClusterNameSubstep } from "@/components/onboarding/steps/cluster-setup/substeps/name";
import { ClusterLocationSubstep } from "@/components/onboarding/steps/cluster-setup/substeps/location";
import { ClusterUrlSubstep } from "@/components/onboarding/steps/cluster-setup/substeps/url";
import { ClusterInstallationSubstep } from "@/components/onboarding/steps/cluster-setup/substeps/installation";
import { showsClusterLocationStep } from "@/services/infra/boot-config.service/boot-config.service";
import { SubStepTransition } from "@/components/onboarding/common/sub-step-transition";
import type { ClusterData, ClusterDataKey } from "@/components/onboarding/models/admin-onboarding.model";
import { EClusterLocations } from "@/models/cluster.model";

export default defineComponent({
  name: "cluster-setup-step",
  components: {
    SubStepTransition,
    ClusterNameSubstep,
    ClusterLocationSubstep,
    ClusterUrlSubstep,
    ClusterInstallationSubstep,
  },
  emits: {
    "update:cluster-data": (key: ClusterDataKey, value: string | boolean) => key && value !== undefined,
    close: () => true,
  },
  props: {
    subStep: {
      type: Number as PropType<number>,
      default: 1,
    },
    clusterData: {
      type: Object as PropType<ClusterData>,
      required: true,
    },
  },
  setup(_, { emit }) {
    const handleClusterDataUpdate = <K extends ClusterDataKey>(key: K, value: ClusterData[K]) => {
      emit("update:cluster-data", key, value);
    };

    const handleClose = () => {
      emit("close");
    };

    return {
      showsClusterLocationStep: showsClusterLocationStep(),
      handleClusterDataUpdate,
      handleClose,
    };
  },
  computed: {
    EClusterLocations() {
      return EClusterLocations;
    },
  },
});
</script>

<style scoped lang="scss">
.cluster-setup-step {
  width: 100%;
  height: 100%;
}
</style>
