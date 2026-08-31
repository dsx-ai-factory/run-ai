<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    aid="network-topologies-modal"
    model-value
    @close="onClose"
    size="width-lg"
    no-padding
    :loading="loadingList"
  >
    <template #header> Network Topologies associated with {{ clusterName }} </template>

    <template #body>
      <section class="q-pa-md" aid="network-topologies-modal-body">
        <div class="q-mb-md">Add topologies that represent this cluster's network</div>
        <div class="q-mb-md italic">
          For more information, see the
          <a :href="DOCS.TOPOLOGIES.SCHEDULING" target="_blank">
            Accelerating Workloads with Network Topology-Aware Scheduling
          </a>
          guide
        </div>

        <card-item
          v-for="(item, index) in list"
          :aid="`network-topology-card-item-${index}`"
          :key="item.id"
          delete-button-tooltip="Delete network topology"
          @delete="deleteNetworkTopology(item.id)"
          :deleting="deletingMapById[item.id]"
          content-centered
          tight-spacing
        >
          <template #summary>
            <div>Network topology: {{ item.name }}</div>
          </template>
        </card-item>

        <template v-if="showForm">
          <q-separator v-if="list.length > 0" class="q-my-lg" />
          <network-topology-form
            :class="{ 'q-mt-lg': list.length === 0 }"
            :cluster-id="clusterUuid"
            @close="toggleFormOpen(false)"
            @submit="createNetworkTopology"
            :submitting="creating"
          />
        </template>
      </section>
      <q-card-actions class="q-px-md" v-if="!showForm">
        <q-btn
          aid="add-network-topology-btn"
          :disable="disableAddNetworkTopologiesBtn"
          v-permission="{ resourceType: ResourceType.NetworkTopologies, action: Action.Create }"
          @click="toggleFormOpen()"
          flat
          color="primary"
          class="btn-icon-sm"
          icon="ph-bold ph-plus"
          label="NETWORK TOPOLOGY"
        />
      </q-card-actions>
    </template>

    <template #footer>
      <runai-btn label="Close" color="primary" @click="onClose" aid="network-topologies-modal-cancel-button" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { DOCS } from "@/constants/docs";

// Components
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiBtn } from "@/components/common/runai-btn";
import { NetworkTopologyForm } from "@/components/cluster/cluster-network-topologies-modal/network-topology-form";
import { CardItem } from "@/components/common/card-item";
// constants
import { errorMessages } from "@/common/error-message.constant";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { NetworkTopology, NetworkTopologyCreateRequest } from "@/swagger-models/cluster-service-client";
// utils / services / stores
import { usePermissionStore } from "@/stores/permissions.store";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { alertUtil } from "@/utils/alert.util";
import { filterUtil } from "@/utils/filter.util/filter.util";

export enum Events {
  cancel = "cancel",
  remove = "remove",
}
export default defineComponent({
  name: "cluster-network-topologies-modal",
  components: {
    RunaiBaseDialog,
    RunaiBtn,
    NetworkTopologyForm,
    CardItem,
  },
  emits: [Events.cancel, Events.remove],
  props: {
    clusterUuid: {
      type: String as PropType<string>,
      required: true,
    },
    clusterName: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      DOCS,
      Action,
      ResourceType,

      permissionStore: usePermissionStore(),
      showForm: false as boolean,
      loadingList: false as boolean,
      deletingMapById: {} as Record<string, boolean>,
      creating: false as boolean,
      list: [] as Array<NetworkTopology>,
    };
  },
  created(): void {
    this.loadList();
  },
  computed: {
    disableAddNetworkTopologiesBtn(): boolean {
      return !this.hasCreatePermission;
    },
    hasCreatePermission(): boolean {
      return this.permissionStore.hasPermission(ResourceType.NetworkTopologies, Action.Create);
    },
  },
  methods: {
    toggleFormOpen(forceValue?: boolean): void {
      this.showForm = forceValue ?? !this.showForm;
    },
    async loadList(): Promise<void> {
      try {
        this.loadingList = true;
        this.list = await clusterService.getNetworkTopologies([
          filterUtil.getEqualsFilterString("clusterId", this.clusterUuid || ""),
        ]);
        this.loadingList = false;
      } catch (error) {
        this.loadingList = false;
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load network topologies"));
      }
    },
    async createNetworkTopology(request: NetworkTopologyCreateRequest): Promise<void> {
      try {
        this.creating = true;

        const res = await clusterService.createNetworkTopology({ ...request });

        this.list.push(res);

        this.creating = false;
        this.toggleFormOpen(false);
        this.$q.notify(alertUtil.getSuccess(`Network topology ${res.name} created`));
      } catch (error) {
        this.creating = false;
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to create network topology"));
      }
    },
    async deleteNetworkTopology(id: string): Promise<void> {
      try {
        this.deletingMapById[id] = true;
        await clusterService.deleteNetworkTopology(id);
        this.list = this.list.filter((item) => item.id !== id);
        this.deletingMapById[id] = false;
        this.$q.notify(alertUtil.getSuccess(`Network topology deleted`));
      } catch (error) {
        this.deletingMapById[id] = false;
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to delete network topology"));
      }
    },
    onRemove(): void {
      this.$emit(Events.remove);
    },
    onClose(): void {
      this.$emit(Events.cancel);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.FORM_INCOMPLETE;
    },
  },
});
</script>
