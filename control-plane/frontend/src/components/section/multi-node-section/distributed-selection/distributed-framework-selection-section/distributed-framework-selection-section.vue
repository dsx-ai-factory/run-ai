<template>
  <runai-section class="distributed-framework-select-section" title="Set the framework for the distributed workload">
    <div class="row">
      <runai-select
        class="col-3 q-mr-md"
        emit-value
        aid="dist-framework-select"
        :model-value="distFramework"
        @update:model-value="$emit('update', $event)"
        :options="distFrameworkOptions"
        :rules="[validFramework]"
        placeholder="Select..."
        unclearable
        :disable="readOnly || disableSelection"
      />
    </div>
    <div v-if="showDocsLink" class="italic q-mt-md q-mb-lg">
      To learn how to enable all frameworks, see the
      <a target="_blank" :href="DOCS.INSTALLATION.DISTRIBUTED_TRAINING"> System Requirements guide</a>
    </div>
    <q-tooltip v-if="disableSelection" anchor="center middle" self="center middle" max-width="370px">
      {{ disabledTooltip }}
    </q-tooltip>
  </runai-section>
</template>

<script lang="ts">
import { DistributedFramework } from "@/swagger-models/assets-service-client";
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { errorMessages } from "@/common/error-message.constant";
import DOCS from "@/constants/docs";
import type { ISelectOption } from "@/models/global.model";
import { useClusterStore } from "@/stores/cluster.store";
import { MIN_CLUSTER_VERSION_JAX } from "@/common/version.constant";
import { RunaiSection } from "@/components/common/runai-section";

const clusterDependenciesKeys: Record<DistributedFramework, string> = {
  [DistributedFramework.PyTorch]: "pytorch",
  [DistributedFramework.Tf]: "tensorflow",
  [DistributedFramework.XgBoost]: "xgboost",
  [DistributedFramework.Mpi]: "mpi",
  [DistributedFramework.Jax]: "jax",
};

export default defineComponent({
  name: "distributed-framework-selection-section",
  components: {
    RunaiSelect,
    RunaiSection,
  },
  props: {
    distFramework: {
      type: [String, null] as PropType<DistributedFramework | null>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    disableSelection: {
      type: Boolean as PropType<boolean>,
    },
    disabledTooltip: {
      type: String as PropType<string>,
      default: "The framework is defined by the distributed workload and cannot be modified",
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      distFrameworkOptions: [
        { value: DistributedFramework.PyTorch, label: "PyTorch", icon: "pytorch", disable: false },
        { value: DistributedFramework.Tf, label: "TensorFlow", icon: "tensorboard", disable: false },
        { value: DistributedFramework.XgBoost, label: "XG Boost", icon: "xgboost", disable: false },
        { value: DistributedFramework.Mpi, label: "MPI", icon: "mpi", disable: false },
        { value: DistributedFramework.Jax, label: "JAX", icon: "jax", disable: false },
      ] as Array<ISelectOption>,
      DOCS,
    };
  },
  created() {
    this.initOptionsAvailability();
  },
  computed: {
    showDocsLink(): boolean {
      return this.distFrameworkOptions.some((opt: ISelectOption) => opt.disable);
    },
  },
  methods: {
    initOptionsAvailability(): void {
      const availableDependencies: Set<string> = new Set(
        this.clusterStore.availableDependenciesByClusterUuid(this.clusterId),
      );

      this.distFrameworkOptions.forEach((option: ISelectOption) => {
        const dependencyKey = clusterDependenciesKeys[option.value as DistributedFramework];
        option.disable = !availableDependencies.has(String(dependencyKey));
        if (option.disable) {
          option.disabledTooltip = "To select this framework, first install it on your cluster";
        }
      });

      const supportsJAX = this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_JAX);
      if (!supportsJAX)
        this.distFrameworkOptions = this.distFrameworkOptions.filter(
          (option: ISelectOption) => option.value !== DistributedFramework.Jax,
        );
    },
    validFramework(value: string | null): boolean | string {
      return !!value || errorMessages.SELECT_A_FRAMEWORK;
    },
  },
});
</script>
