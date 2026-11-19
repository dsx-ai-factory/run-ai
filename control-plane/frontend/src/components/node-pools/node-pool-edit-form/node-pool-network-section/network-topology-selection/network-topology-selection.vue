<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section
    class="q-mb-md"
    title="Select the topology that represents this node pool's network"
    tooltip-text="This optimizes placement and accelerates distributed workloads by keeping pods on nodes that are as close to each other as possible in the network."
  >
    <div class="row">
      <runai-select
        :loading="loadingNetworkTopologies"
        :disable="disable || loadingNetworkTopologies || creating"
        emit-value
        class="col-5"
        placeholder="Select..."
        stack-label
        label="Network topology"
        :clear-icon="networkTopologyId ? 'ph-bold ph-x' : undefined"
        :custom-action="
          canCreate && !disable && !loadingNetworkTopologies && !creating ? { text: 'new network topology' } : undefined
        "
        @custom-action="setModalOpen(true)"
        @update:model-value="updateNetworkTopologyId"
        :model-value="networkTopologyId"
        :options="networkTopologyOptions"
        no-option-text="No network topologies found"
      />
    </div>

    <div v-if="loadError" class="row items-center q-gutter-sm text-negative" role="alert">
      <span>{{ loadError }}</span>
      <runai-btn
        aid="network-topologies-retry"
        label="Retry"
        color="primary"
        flat
        :disable="loadingNetworkTopologies"
        @click="loadNetworkTopologies"
      />
    </div>
    <div
      v-else-if="!loadingNetworkTopologies && !networkTopologies.length"
      aid="network-topologies-empty"
      class="text-grey"
      role="status"
    >
      No network topologies found
    </div>

    <runai-base-dialog
      aid="network-topologies-creation-modal"
      :model-value="showCreationModal"
      :is-modal-persistent="creating"
      :hide-x="creating"
      @close="setModalOpen(false)"
      size="width-lg"
      no-padding
    >
      <template #header> New Network Topology </template>

      <template #body>
        <section aid="network-topologies-modal-body">
          <fieldset class="topology-form-fields" :disabled="creating" :inert="creating || undefined">
            <div class="q-px-md q-my-md">Add a topology that represents this node pool's network</div>

            <network-topology-form
              ref="topologyFormComponent"
              :bordered="false"
              :show-action-buttons="false"
              show-distance-guide
              :cluster-id="clusterId"
              :external-submit-trigger="externalSubmitTrigger"
              @close="setModalOpen(false)"
              @submit="createNetworkTopology"
            />
          </fieldset>
        </section>
      </template>

      <template #footer>
        <runai-btn
          label="Cancel"
          color="primary"
          flat
          :disable="creating"
          @click="setModalOpen(false)"
          aid="network-topologies-modal-cancel-button"
        />
        <runai-btn
          :loading="creating"
          :disable="disable || !canCreate"
          label="Save network topology"
          color="primary"
          @click="triggerFormSubmit"
          aid="network-topologies-modal-save-button"
        />
      </template>
    </runai-base-dialog>
  </runai-section>
</template>
<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiBtn } from "@/components/common/runai-btn";
import { NetworkTopologyForm } from "@/components/cluster/cluster-network-topologies-modal/network-topology-form";
import type { NetworkTopology, NetworkTopologyCreateRequest } from "@/swagger-models/cluster-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { alertUtil } from "@/utils/alert.util";
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  name: "network-topology-selection",
  components: { RunaiSection, RunaiSelect, RunaiBaseDialog, RunaiBtn, NetworkTopologyForm },
  emits: ["select", "loaded", "creating"],
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    disable: {
      type: Boolean,
      default: false,
    },
    canCreate: {
      type: Boolean,
      default: true,
    },
    networkTopologyName: {
      type: String,
      default: "",
    },
    networkTopologyId: {
      type: [String, null] as PropType<string | null>,
      default: null,
    },
  },
  data() {
    return {
      showCreationModal: false,
      externalSubmitTrigger: 0,

      networkTopologies: [] as NetworkTopology[],
      loadingNetworkTopologies: false,
      loadError: "",
      loadRequestId: 0,
      createRequestId: 0,
      disposed: false,

      creating: false,
    };
  },
  computed: {
    networkTopologyOptions(): ISelectOption[] {
      const options = this.networkTopologies.map((topology) => ({
        label: topology.name,
        value: topology.id,
      }));
      if (this.networkTopologyId && !options.some((option) => option.value === this.networkTopologyId)) {
        options.push({ value: this.networkTopologyId, label: this.networkTopologyName || this.networkTopologyId });
      }
      return options;
    },
  },
  methods: {
    async loadNetworkTopologies(): Promise<void> {
      const requestId = ++this.loadRequestId;
      const clusterId = this.clusterId;
      this.loadError = "";
      this.loadingNetworkTopologies = !!clusterId;
      if (!clusterId || this.disposed) return;
      try {
        const topologies = await clusterService.getNetworkTopologies([
          filterUtil.getEqualsFilterString("clusterId", clusterId),
        ]);
        if (this.disposed || requestId !== this.loadRequestId) return;
        this.networkTopologies = topologies;
        this.$emit("loaded", topologies);
      } catch (error) {
        if (this.disposed || requestId !== this.loadRequestId) return;
        console.error("Failed to load network topologies", error);
        this.loadError = "Failed to load network topologies";
      } finally {
        if (!this.disposed && requestId === this.loadRequestId) this.loadingNetworkTopologies = false;
      }
    },
    async createNetworkTopology(request: NetworkTopologyCreateRequest): Promise<void> {
      if (this.disable || !this.canCreate || this.creating || this.disposed || !this.showCreationModal) return;
      if (request.clusterId !== this.clusterId) return;
      const requestId = ++this.createRequestId;
      this.creating = true;
      try {
        const newTopology = await clusterService.createNetworkTopology(request);
        if (this.disposed || requestId !== this.createRequestId) return;
        this.networkTopologies = [
          ...this.networkTopologies.filter((topology) => topology.id !== newTopology.id),
          newTopology,
        ];
        this.$emit("loaded", this.networkTopologies);
        this.$emit("select", newTopology.id);
        this.showCreationModal = false;
      } catch (error) {
        if (this.disposed || requestId !== this.createRequestId) return;
        console.error("Failed to create network topology", error);
        this.$q.notify(alertUtil.getError("Failed to create network topology"));
      } finally {
        if (!this.disposed && requestId === this.createRequestId) this.creating = false;
      }
    },
    updateNetworkTopologyId(value: string | null): void {
      if (this.disable || this.loadingNetworkTopologies || this.creating || this.disposed) return;
      this.$emit("select", value);
    },
    setModalOpen(value: boolean): void {
      if (this.creating) return;
      if (value && (this.disable || !this.canCreate || this.loadingNetworkTopologies || this.disposed)) return;
      this.showCreationModal = value;
    },
    triggerFormSubmit(): void {
      if (this.disable || !this.canCreate || this.creating || this.disposed) return;
      this.externalSubmitTrigger++;
    },
  },
  beforeUnmount() {
    this.creating = false;
    this.disposed = true;
    this.loadRequestId++;
    this.createRequestId++;
  },
  watch: {
    showCreationModal(value: boolean, previous: boolean) {
      if (value || !previous) return;
      void this.$nextTick(() => {
        if (this.disposed || this.disable || this.loadingNetworkTopologies || this.creating || this.showCreationModal)
          return;
        this.$el.querySelector('[role="combobox"]')?.focus();
      });
    },
    creating: {
      handler(value: boolean) {
        this.$emit("creating", value);
      },
      flush: "sync",
    },
    clusterId: {
      handler(newValue: string, oldValue: string) {
        if (newValue !== oldValue) {
          this.createRequestId++;
          this.creating = false;
          this.showCreationModal = false;
          this.networkTopologies = [];
          void this.loadNetworkTopologies();
        }
      },
      immediate: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.topology-form-fields {
  border: 0;
  margin: 0;
  min-width: 0;
  padding: 0;
}
</style>
