<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section
    class="q-mb-md"
    title="Select the topology that represents this node pool's network"
    tooltip-text="Optimizes placement and accelerates distributed workloads by keeping pods on nodes that are as close to each other as possible in the network"
  >
    <div class="row">
      <runai-select
        :loading="loadingNetworkTopologies"
        emit-value
        class="col-5"
        placeholder="Select..."
        stack-label
        label="Network topology"
        clear-icon="ph-bold ph-x"
        :custom-action="{ text: 'new network topology' }"
        @custom-action="setModalOpen(true)"
        @update:model-value="updateNetworkTopologyId"
        :model-value="networkTopologyId"
        :options="networkTopologyOptions"
        no-option-text="No network topologies found"
      />
    </div>

    <runai-base-dialog
      aid="network-topologies-creation-modal"
      :model-value="showCreationModal"
      @close="setModalOpen(false)"
      size="width-lg"
      no-padding
    >
      <template #header> New Network Topology </template>

      <template #body>
        <section aid="network-topologies-modal-body">
          <div class="q-px-md q-my-md">Add a topology that represents this node pool's network</div>

          <network-topology-form
            ref="topologyFormComponent"
            :bordered="false"
            :show-action-buttons="false"
            :cluster-id="clusterId"
            :external-submit-trigger="externalSubmitTrigger"
            @close="setModalOpen(false)"
            @submit="createNetworkTopology"
          />
        </section>
      </template>

      <template #footer>
        <runai-btn
          label="Cancel"
          color="primary"
          flat
          @click="setModalOpen(false)"
          aid="network-topologies-modal-cancel-button"
        />
        <runai-btn
          :loading="creating"
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
  emits: ["select"],
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
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
      loadingNetworkTopologies: false as boolean,

      creating: false,
    };
  },
  computed: {
    networkTopologyOptions(): ISelectOption[] {
      return this.networkTopologies.map((topology) => ({
        label: topology.name,
        value: topology.id,
      }));
    },
  },
  methods: {
    async loadNetworkTopologies(): Promise<void> {
      try {
        this.loadingNetworkTopologies = true;
        this.networkTopologies = await clusterService.getNetworkTopologies([
          filterUtil.getEqualsFilterString("clusterId", this.clusterId || ""),
        ]);
      } catch (error) {
        console.error("Failed to load network topologies", error);
        this.$q.notify(alertUtil.getError("Failed to load network topologies"));
      } finally {
        this.loadingNetworkTopologies = false;
      }
    },
    async createNetworkTopology(request: NetworkTopologyCreateRequest): Promise<void> {
      try {
        this.creating = true;
        const newTopology = await clusterService.createNetworkTopology(request);
        this.networkTopologies.push(newTopology);
        this.updateNetworkTopologyId(newTopology.id);
        this.setModalOpen(false);
      } catch (error) {
        console.error("Failed to create network topology", error);
        this.$q.notify(alertUtil.getError("Failed to create network topology"));
      } finally {
        this.creating = false;
      }
    },
    updateNetworkTopologyId(value: string | null): void {
      this.$emit("select", value);
    },
    setModalOpen(value: boolean): void {
      this.showCreationModal = value;
    },
    triggerFormSubmit(): void {
      // Increment the trigger to notify the form to validate and submit
      this.externalSubmitTrigger++;
    },
  },
  watch: {
    clusterId: {
      handler(newValue: string, oldValue: string) {
        if (newValue && newValue !== oldValue) {
          this.loadNetworkTopologies();
        }
      },
      immediate: true,
    },
  },
});
</script>
