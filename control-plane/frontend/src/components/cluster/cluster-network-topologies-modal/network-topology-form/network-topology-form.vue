<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card flat :bordered="bordered">
    <q-card-section>
      <q-form ref="topologyForm" aid="network-topology-form">
        <div class="row q-mb-md">
          <runai-name-validation
            class="col-6"
            aid="network-topology-name-input"
            label="Network topology name"
            :model-value="name"
            @update:model-value="name = $event"
            :rules="[requiredName, isValidFormat, isNameUnique]"
          />
        </div>

        <div class="q-mb-md">
          Set node labels representing the network topology and
          {{ showLevelOrderControls ? "use the move buttons or drag" : "drag" }} to order by their distance from the node
          (far to close)
          <runai-tooltip
            tooltip-text="Order levels from farthest (top) to closest (bottom). Use the Kubernetes label key on the nodes for each level."
            width="542px"
          />
        </div>

        <div class="topology-levels q-mb-md">
          <div v-if="showLevelOrderControls" class="topology-distance-guide text-b-body1" aria-hidden="true">
            <span class="topology-distance-label">Far</span>
            <span class="topology-distance-line"></span>
            <span class="topology-distance-label">Close</span>
          </div>
          <runai-draggable-list-order
            class="topology-level-list"
            :list="labels"
            row-key="id"
            main-list-title="Label"
            @change="labelsChanged"
            :show-border="false"
            :show-header="false"
            item-size="md"
          >
            <template #list-item="item">
              <div class="row items-center justify-between" :class="{ 'no-wrap': showLevelOrderControls }">
                <policy-string-field
                  class="q-mb-md q-mt-sm"
                  :class="showLevelOrderControls ? 'col q-pr-sm' : 'col-11'"
                  aid="network-topology-label-key-input"
                  label="Label key"
                  input-class="placeholder-italic"
                  stack-label
                  placeholder="e.g., topology.kubernetes.io/zone"
                  :model-value="item.name"
                  :rules="[validateLevelKey]"
                  @update:model-value="onUpdateLabel(item.id, $event)"
                />
                <div v-if="showLevelOrderControls" class="row items-center no-wrap q-gutter-xs">
                  <runai-btn
                    aid="network-topology-move-up"
                    class="topology-move-button"
                    icon-button
                    dense
                    icon="ph-bold ph-arrow-up"
                    tooltip="Move up"
                    :aria-label="`Move up level ${labels.findIndex((label) => label.id === item.id) + 1}`"
                    :aria-disabled="item.id === labels[0]?.id"
                    @click="moveLabel(item.id, -1, $event)"
                  />
                  <runai-btn
                    aid="network-topology-move-down"
                    class="topology-move-button"
                    icon-button
                    dense
                    icon="ph-bold ph-arrow-down"
                    tooltip="Move down"
                    :aria-label="`Move down level ${labels.findIndex((label) => label.id === item.id) + 1}`"
                    :aria-disabled="item.id === labels.at(-1)?.id"
                    @click="moveLabel(item.id, 1, $event)"
                  />
                </div>
                <runai-tooltip
                  icon="ph-bold ph-x"
                  @click="onRemoveLabel(item.id)"
                  tooltip-text="Remove"
                  class="remove-button"
                  v-show="labels.length > 1"
                />
              </div>
            </template>
          </runai-draggable-list-order>
        </div>
        <p v-if="showLevelOrderControls" role="status" class="topology-order-status">{{ orderAnnouncement }}</p>
        <runai-btn
          flat
          aid="network-topology-add-label-btn"
          color="primary"
          class="btn-icon-sm"
          icon="ph-bold ph-plus"
          @click="onAddLabel"
          label="label"
        />
      </q-form>
    </q-card-section>
    <template v-if="showActionButtons">
      <q-separator />
      <q-card-actions>
        <runai-form-footer
          @cancel="$emit('close')"
          @save="onSave"
          :is-form-valid="isFormValid"
          save-btn-label="save network topology"
          save-btn-flat
          :loading="submitting"
        />
      </q-card-actions>
    </template>
  </q-card>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { RunaiFormFooter } from "@/components/common/runai-form-footer";
import { RunaiDraggableListOrder } from "@/components/common/runai-draggable-list-order";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiBtn } from "@/components/common/runai-btn";
// utils & services
import { ErrorAlert } from "@/utils/error-alert.util";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { isNotEmpty, isValidRunaiEntityNameFormat, isValidNetworkTopologyLevel } from "@/common/form.validators";
import { makeId } from "@/utils/common.util";

// constants
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "network-topology-form",
  components: {
    RunaiFormFooter,
    RunaiNameValidation,
    RunaiDraggableListOrder,
    RunaiTooltip,
    PolicyStringField,
    RunaiBtn,
  },
  emits: ["close", "submit"],
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    submitting: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    bordered: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showActionButtons: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showLevelOrderControls: Boolean,
    externalSubmitTrigger: {
      type: Number as PropType<number>,
      default: 0,
    },
  },
  data() {
    return {
      labels: [{ name: "", id: makeId() }] as { name: string; id: string }[],
      name: "" as string,
      orderAnnouncement: "",

      isFormValid: true as boolean,
    };
  },
  methods: {
    async validate(): Promise<boolean> {
      const form = this.$refs["topologyForm"] as HTMLFormElement;
      return await (form as HTMLFormElement).validate();
    },
    notifyCreateError(error: unknown): void {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedCreateMessage("network topologies"),
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
    async onSave(): Promise<void> {
      this.isFormValid = await this.validate();

      if (!this.isFormValid) return;

      this.$emit("submit", {
        name: this.name,
        levels: this.labels.map((label) => label.name),
        clusterId: this.clusterId,
      });
    },
    onRemoveLabel(id: string): void {
      this.labels = this.labels.filter((label) => label.id !== id);
    },
    onAddLabel(): void {
      this.labels.push({ name: "", id: makeId() });
    },
    onUpdateLabel(id: string, name: string): void {
      this.labels = this.labels.map((label) => (label.id === id ? { ...label, name } : label));
    },
    async moveLabel(id: string, offset: -1 | 1, event: MouseEvent): Promise<void> {
      const from = this.labels.findIndex((label) => label.id === id);
      const to = from + offset;
      if (from < 0 || to < 0 || to >= this.labels.length) return;

      const button = event.currentTarget as HTMLButtonElement;
      const labels = [...this.labels];
      const [level] = labels.splice(from, 1);
      labels.splice(to, 0, level);
      this.labelsChanged(labels);
      this.orderAnnouncement = `Moved level from position ${from + 1} to position ${to + 1} of ${labels.length}.`;
      await this.$nextTick();
      button.focus();
    },
    labelsChanged(labels: { name: string; id: string }[]): void {
      this.labels = labels;
    },
    requiredName(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    isValidFormat(val: string): boolean | string {
      return isValidRunaiEntityNameFormat(val) || errorMessages.VALID_FORMAT;
    },
    validateLevelKey(val: string): boolean | string {
      return (
        isValidNetworkTopologyLevel(val) ||
        "Enter a Kubernetes label key of up to 63 characters, including its optional lowercase DNS prefix."
      );
    },
    async isNameUnique(name: string): Promise<boolean | string> {
      try {
        const amount = await clusterService.getNetworkTopologyCount([
          filterUtil.getEqualsFilterString("name", name),
          filterUtil.getEqualsFilterString("clusterId", this.clusterId),
        ]);
        if (amount > 0) {
          return errorMessages.NAME_ALREADY_EXIST;
        }
        return true;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedToValidateMessage("network topology name"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        return false;
      }
    },
  },
  watch: {
    externalSubmitTrigger: {
      handler(newValue: number, oldValue: number) {
        if (newValue !== oldValue) {
          this.onSave();
        }
      },
    },
  },
});
</script>

<style scoped lang="scss">
.topology-levels {
  display: flex;
  gap: 10px;
}

.topology-level-list {
  flex: 1;
  min-width: 0;
}

.topology-distance-guide {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.topology-distance-label {
  writing-mode: vertical-rl;
  color: var(--text-tertiary);
}

.topology-distance-line {
  flex: 1;
  width: 1px;
  background-color: var(--border-default);
}

.topology-levels :deep(.topology-move-button[aria-disabled="true"]) {
  color: var(--icon-disabled);
  cursor: default;
}

.topology-order-status {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
</style>
