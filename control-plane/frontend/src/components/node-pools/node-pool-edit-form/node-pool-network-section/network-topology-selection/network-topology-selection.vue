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
        :disable="selectDisabled"
        emit-value
        class="col-5"
        placeholder="Select..."
        stack-label
        label="Network topology"
        popup-content-class="network-topology-menu"
        :clear-icon="networkTopologyId ? 'ph-bold ph-x' : undefined"
        :custom-action="canStartCreation ? { text: 'new network topology' } : undefined"
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
        :disable="loadingNetworkTopologies || creating"
        @click="loadNetworkTopologies"
      />
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
              :bordered="false"
              :show-action-buttons="false"
              show-level-order-controls
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
          :disable="!canStartCreation"
          label="Save network topology"
          color="primary"
          @click="externalSubmitTrigger++"
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
      requestId: 0,

      creating: false,
    };
  },
  computed: {
    selectDisabled(): boolean {
      return this.disable || this.loadingNetworkTopologies || this.creating;
    },
    canStartCreation(): boolean {
      return !!this.clusterId && this.canCreate && !this.selectDisabled;
    },
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
      if (this.creating) return;
      const requestId = ++this.requestId;
      this.loadError = "";
      this.loadingNetworkTopologies = !!this.clusterId;
      if (!this.clusterId) return;
      try {
        const topologies = await clusterService.getNetworkTopologies([
          filterUtil.getEqualsFilterString("clusterId", this.clusterId),
        ]);
        if (requestId !== this.requestId) return;
        this.networkTopologies = topologies;
        this.loadingNetworkTopologies = false;
        this.$emit("loaded", topologies);
      } catch (error) {
        if (requestId !== this.requestId) return;
        console.error("Failed to load network topologies", error);
        this.loadError = "Failed to load network topologies";
        this.loadingNetworkTopologies = false;
      }
    },
    async createNetworkTopology(request: NetworkTopologyCreateRequest): Promise<void> {
      if (!this.showCreationModal || !this.canStartCreation) return;
      if (request.clusterId !== this.clusterId) {
        this.$q.notify(alertUtil.getWarning("The selected cluster changed. Review the topology and save it again."));
        return;
      }
      // Loading a different cluster invalidates its predecessor's creation as well.
      const requestId = this.requestId;
      this.creating = true;
      try {
        const newTopology = await clusterService.createNetworkTopology(request);
        if (requestId !== this.requestId) return;
        this.networkTopologies = [
          ...this.networkTopologies.filter((topology) => topology.id !== newTopology.id),
          newTopology,
        ];
        this.$emit("loaded", this.networkTopologies);
        this.$emit("select", newTopology.id);
        this.showCreationModal = false;
      } catch (error) {
        if (requestId !== this.requestId) return;
        console.error("Failed to create network topology", error);
        this.$q.notify(alertUtil.getError("Failed to create network topology"));
      }
      this.creating = false;
    },
    updateNetworkTopologyId(value: string | null): void {
      if (!this.selectDisabled) this.$emit("select", value);
    },
    setModalOpen(value: boolean): void {
      if (value ? this.canStartCreation : !this.creating) this.showCreationModal = value;
    },
  },
  beforeUnmount() {
    this.requestId++;
    this.creating = false;
  },
  watch: {
    showCreationModal: {
      flush: "post",
      handler(open: boolean) {
        if (!open && !this.selectDisabled) this.$el.querySelector('[role="combobox"]')?.focus();
      },
    },
    creating: {
      handler(value: boolean) {
        this.$emit("creating", value);
      },
      flush: "sync",
    },
    clusterId: {
      handler() {
        this.creating = false;
        this.showCreationModal = false;
        this.networkTopologies = [];
        void this.loadNetworkTopologies();
      },
      immediate: true,
    },
  },
});
</script>

<style lang="scss" scoped>
:global(.network-topology-menu .custom-action-item-text) {
  font-weight: 500;
}

.topology-form-fields {
  border: 0;
  margin: 0;
  min-width: 0;
  padding: 0;
}
</style>
