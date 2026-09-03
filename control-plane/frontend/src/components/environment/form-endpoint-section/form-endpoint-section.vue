<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    aid="endpoint"
    label="Endpoint"
    class="form-endpoint-section"
    default-opened
    :section-invalid="!sectionValid"
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>

    <section>
      <div class="form-endpoint-section-connection-title q-mb-sm">
        Set an inference serving endpoint
        <runai-tooltip
          tooltip-text="Define how applications or services should communicate inference requests to the workload. The endpoint will be ignored when this environment is used by non-inference workloads."
          tooltip-position="right"
          width="490px"
        />
      </div>

      <div class="row gap-20">
        <div class="col-3">
          <span class="form-endpoint-section-protocol-wrap">
            <runai-select
              :model-value="connection.protocol"
              @update:model-value="updateSelectedProtocol"
              :options="protocolOptions"
              label="Protocol"
              stack-label
              no-error-icon
              aid="connection-protocol"
              :disable="protocolLockedToHttp"
            />
            <q-tooltip v-if="protocolLockedToHttp" anchor="top middle" self="bottom middle">
              For distributed inference only HTTP is supported
            </q-tooltip>
          </span>
        </div>
        <div class="col-2">
          <runai-number-input
            aid="container-port"
            label="Container port"
            stack-label
            :model-value="connection.containerPort"
            @update:model-value="updatePort"
            :rules="[portNotEmpty, validPort]"
            :min="minValidPort"
            :max="maxValidPort"
          />
        </div>
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiNumberInput } from "@/components/common/form/runai-number-input";
// models
import type { ISelectOption } from "@/models/global.model";
import { ServingPortInfoProtocolEnum } from "@/swagger-models/assets-service-client";
import type { IUIConnection } from "@/models/environment.model";

import { errorMessages } from "@/common/error-message.constant";
import {
  maxValidPort,
  minValidPort,
} from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";

export const protocolOptions: Array<ISelectOption> = [
  {
    label: "HTTP",
    value: ServingPortInfoProtocolEnum.Http,
  },
  {
    label: "gRPC",
    value: ServingPortInfoProtocolEnum.Grpc,
  },
];

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSelect,
    RunaiTooltip,
    RunaiNumberInput,
  },
  emits: ["update"],
  props: {
    connection: {
      type: Object as PropType<IUIConnection>,
      required: true,
    },
    protocolLockedToHttp: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      protocolOptions,
      maxValidPort,
      minValidPort,
    };
  },
  computed: {
    summary(): string {
      return "None";
    },
    sectionValid(): boolean {
      return this.connection.containerPort !== undefined;
    },
  },
  methods: {
    updateSelectedProtocol(protocol: ISelectOption): void {
      this.$emit("update", { ...this.connection, protocol: protocol.value });
    },
    updatePort(containerPort: number | string | null): void {
      let port: number | undefined = undefined;

      if (containerPort !== null && containerPort !== undefined && containerPort !== "") {
        port = Number(containerPort);
      }

      this.$emit("update", { ...this.connection, containerPort: port });
    },
    portNotEmpty(val: number | undefined): boolean | string {
      return val !== undefined || errorMessages.ENTER_A_PORT;
    },
    validPort(val: number): boolean | string {
      const message = errorMessages.ENTER_PORT_IN_RANGE.replace("${min}", `${minValidPort}`).replace(
        "${max}",
        `${maxValidPort}`,
      );

      return (val >= minValidPort && val <= maxValidPort) || message;
    },
  },
});
</script>

<style lang="scss" scoped>
.form-endpoint-section-protocol-wrap {
  display: block;
  width: 100%;
}
</style>
