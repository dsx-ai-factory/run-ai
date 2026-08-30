<template>
  <section class="serving-endpoint-access">
    <runai-section :title="title" :tooltip-text="titleTooltip" tooltip-width="520px" tooltip-position="right">
      <section class="row q-mb-md">
        <span class="row col-8 gap-20" v-if="showProtocolAndPort">
          <span class="col-4 connection-protocol">
            <template v-if="protocolShouldBeDisabled">
              <policy-string-field
                aid="connection-protocol"
                label="Protocol"
                :model-value="protocolView"
                disable
                stack-label
              >
              </policy-string-field>
              <q-tooltip v-if="protocolDisabledInfo" anchor="bottom middle" :offset="[0, -15]">{{
                protocolDisabledInfo
              }}</q-tooltip>
            </template>
            <runai-select
              v-else
              :model-value="modelValue.protocol || ''"
              @update:model-value="updateSelectedProtocol"
              :options="protocolOptions"
              label="Protocol"
              stack-label
              aid="connection-protocol"
            />
          </span>
          <span class="col-3 container-port">
            <policy-number-input
              aid="container-port"
              label="Container port"
              stack-label
              :model-value="modelValue.port || null"
              @update:model-value="updatePort"
              :rules="portRules"
              :disable="portDisabled || disabled"
              :min="minValidPort"
              :max="maxValidPort"
              :policy-rules="policyRules?.container"
            />
            <q-tooltip v-if="portDisabledInfo">{{ portDisabledInfo }}</q-tooltip>
          </span>
        </span>
        <span class="col-4 connection-access">
          <connection-access
            v-if="modelValue.access"
            :access="modelValue.access"
            :tool-name="'Endpoint'"
            @save="onAccessChanged"
            enable-public-access
            connection-type="endpoint"
            :disabled="disabled"
            :disabled-tooltip="disabledInfo"
            :policy-rules="policyRules"
          />
        </span>
      </section>
    </runai-section>

    <runai-section
      v-if="supportsClusterLocalAccess"
      title="Set whether the serving endpoint can be accessed externally or only by internal cluster traffic"
      class="q-mb-md"
    >
      <div class="endpoint-type-section">
        <runai-radio-options
          :model-value="endpointTypeValue"
          :options="endpointTypeOptions"
          @update:model-value="onEndpointTypeChanged"
        />
      </div>
    </runai-section>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { ConnectionAccess } from "@/components/environment/connection-access";
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { useClusterStore } from "@/stores/cluster.store";

// models
import { type IConnectionAccess } from "@/components/environment/connection-access";
import type { ISelectOption } from "@/models/global.model";
import { errorMessages } from "@/common/error-message.constant";
import {
  EServingEndpointAccessEvents,
  maxValidPort,
  minValidPort,
  ProtocolOptions,
  type IServingEndpointAccessData,
  ENDPOINT_TYPE_EXTERNAL,
  ENDPOINT_TYPE_INTERNAL,
} from "./serving-endpoint-access.model";
import type { SupersetServingPortRules } from "@/swagger-models/policy-service-client";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import { PolicyStringField } from "@/components/common/policy-string-field";
import type { ValidationRule } from "quasar";
import { isNullOrUndefined } from "@/utils/common.util";
import { MIN_CLUSTER_VERSION_FOR_INFERENCE_CLUSTER_EXTERNAL_ACCESS_STATUS } from "@/common/version.constant";

export default defineComponent({
  components: {
    RunaiSection,
    RunaiRadioOptions,
    ConnectionAccess,
    RunaiSelect,
    PolicyNumberInput,
    PolicyStringField,
  },
  emits: {
    [EServingEndpointAccessEvents.UpdateModelValue]: (modelValue: IServingEndpointAccessData) => modelValue,
  },
  props: {
    modelValue: {
      type: Object as PropType<IServingEndpointAccessData>,
      required: true,
    },
    title: {
      type: String as PropType<string>,
      required: true,
    },
    titleTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    protocolDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    portDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showProtocolAndPort: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disabledInfo: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetServingPortRules | null>,
      required: false,
    },
    isPortRequired: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      protocolOptions: ProtocolOptions,
      maxValidPort,
      minValidPort,
      portRules: [() => true] as ValidationRule[],
      ENDPOINT_TYPE_EXTERNAL,
      ENDPOINT_TYPE_INTERNAL,
    };
  },
  created() {
    if (this.isPortRequired) {
      this.portRules = [this.portNotEmpty, this.validPort];
    }
    if (this.supportsClusterLocalAccess) {
      if (isNullOrUndefined(this.modelValue.clusterLocalAccessOnly)) {
        this.initializeEndpointType();
      }
    }
  },
  computed: {
    supportsClusterLocalAccess(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.clusterId,
        MIN_CLUSTER_VERSION_FOR_INFERENCE_CLUSTER_EXTERNAL_ACCESS_STATUS,
      );
    },
    isExternalAccessAvailable(): boolean {
      return this.clusterStore.isConfigDomainEnabled(this.clusterId);
    },
    endpointTypeValue(): string {
      if (!this.isExternalAccessAvailable) {
        return ENDPOINT_TYPE_INTERNAL;
      }
      return this.modelValue.clusterLocalAccessOnly ? ENDPOINT_TYPE_INTERNAL : ENDPOINT_TYPE_EXTERNAL;
    },
    endpointTypeOptions(): ISelectOption[] {
      return [
        {
          label: "External (Public access)",
          value: ENDPOINT_TYPE_EXTERNAL,
          disable: !this.isExternalAccessAvailable,
          disabledTooltip: !this.isExternalAccessAvailable
            ? "To enable external access, contact your administrator about\nconfiguring Knative for the cluster"
            : undefined,
          svg: !this.isExternalAccessAvailable ? "ban" : undefined,
        },
        {
          label: "Internal only",
          value: ENDPOINT_TYPE_INTERNAL,
        },
      ];
    },
    protocolView(): string {
      const protocol = ProtocolOptions.find((option) => option.value === this.modelValue.protocol);
      return protocol?.label || "";
    },
    portDisabledInfo(): string | undefined {
      if ((this.disabled || this.portDisabled) && this.disabledInfo) return this.disabledInfo;
      if (this.portDisabled) return "The connection protocol and the container port are defined within the environment";

      return undefined;
    },
    isDisabledByPolicy(): boolean {
      return this.policyRules?.protocol?.canEdit === false;
    },
    protocolDisabledInfo(): string | undefined {
      if ((this.disabled || this.portDisabled) && this.disabledInfo) return this.disabledInfo;
      if (this.isDisabledByPolicy) {
        return errorMessages.CANT_BE_MODIFIED_POLICY;
      }
      if (this.portDisabled) return "The connection protocol and the container port are defined within the environment";

      return undefined;
    },
    protocolRules(): IGenericSelectPolicyRules | undefined {
      return this.policyRules?.protocol ? { ...this.policyRules?.protocol, isClosedList: true } : undefined;
    },
    protocolShouldBeDisabled(): boolean {
      return this.protocolDisabled || this.disabled || this.isDisabledByPolicy;
    },
  },
  methods: {
    initializeEndpointType(): void {
      const clusterLocalAccessOnly = !this.isExternalAccessAvailable;
      this.emitChanges({
        ...this.modelValue,
        clusterLocalAccessOnly,
      });
    },
    onEndpointTypeChanged(value: string): void {
      if (!this.supportsClusterLocalAccess) {
        return;
      }
      if (value === ENDPOINT_TYPE_EXTERNAL && !this.isExternalAccessAvailable) {
        return;
      }
      const clusterLocalAccessOnly = value === ENDPOINT_TYPE_INTERNAL;
      this.emitChanges({
        ...this.modelValue,
        clusterLocalAccessOnly,
      });
    },
    onAccessChanged(access: IConnectionAccess): void {
      this.emitChanges({
        ...this.modelValue,
        access,
      });
    },
    updateSelectedProtocol(protocol: ISelectOption): void {
      this.emitChanges({
        ...this.modelValue,
        protocol: protocol.value?.toString(),
      });
    },
    updatePort(containerPort: number | string | null): void {
      let port: number | undefined = undefined;

      if (containerPort !== null && containerPort !== undefined && containerPort !== "") {
        port = Number(containerPort);
      }

      this.emitChanges({
        ...this.modelValue,
        port,
      });
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
    emitChanges(servingEndpointAccessData: IServingEndpointAccessData): void {
      this.$emit(EServingEndpointAccessEvents.UpdateModelValue, servingEndpointAccessData);
    },
  },
});
</script>
<style lang="scss" scoped></style>
