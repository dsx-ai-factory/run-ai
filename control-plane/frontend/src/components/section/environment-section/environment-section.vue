<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="environment-section"
    label="Environment"
    default-opened
    :section-invalid="!selectedEnvironment?.id || showEnvVarValidation"
  >
    <template #subheader>
      <span>{{ environmentSummary }}</span>
    </template>

    <runai-form-card-section
      :loading="loading"
      empty-message="Looks like you don't have any environments yet..."
      :main-message="cardsMainTitle"
      entity-name="environment"
      :cards-list="environmentCards"
      :selected-cards-ids="selectedEnvironmentId"
      icon-name="ph-bold ph-x-square"
      search-name="environments"
      @selected-card-changed="onSelectedEnv"
      @create-new="createNew"
      :disabled="sectionDisabled"
      :disable-create-new="disableCreateButton"
      :sort-options="{ name: true, creation: true, recentUsage: true }"
      default-sort-option="recentUsage"
    />

    <div class="col-12 q-ml-sm">
      <q-field class="hidden-field" :model-value="selectedEnvironment" :rules="[isRequiredEnvironment]"></q-field>
    </div>

    <template v-if="!loading && specificEnv">
      <div v-if="supportingServingPort && servingEndpointModel" class="q-mb-md">
        <serving-endpoint-section
          title="Set an inference serving endpoint"
          title-tooltip="Define how applications or services should communicate inference requests to the workload."
          :model-value="servingEndpointModel"
          @update:model-value="onServingPortChanged"
          show-protocol-and-port
          protocol-disabled
          port-disabled
          :cluster-id="clusterId"
        />
      </div>

      <section v-if="uiConnections.length">
        <div class="q-mb-md">Set the connection for your tool(s)</div>
        <template v-for="tool in uiConnections" :key="tool.id">
          <tool-box
            :form-type="EToolBoxFormType.WorkloadAssetsBased"
            :tool="tool"
            :tools="uiConnections"
            :removeable="false"
            @update-tool="updateConnections"
            :is-required="isToolBoxRequired"
            :disable="sectionDisabled"
            :cluster-id="clusterId"
            :access-disabled="accessDisabled"
          ></tool-box>
        </template>
      </section>

      <section v-if="showUidGid" aid="uid-gid-section" class="q-mt-lg q-mb-md">
        <div class="q-mb-sm">
          Set either the UID, GID or the Supplementary groups that can run commands in the container
        </div>
        <div class="row justify-between">
          <div class="col-3">
            <policy-number-input
              aid="uid-input"
              :min-value="0"
              :model-value="uidModel"
              @update:model-value="updateUid"
              label="User ID (UID)"
              stack-label
              :disable="disableUidInput"
              :tooltip="uidInputTooltip"
              :policy-rules="policyRules?.rules?.specificRunParams?.runAsUid"
            />
          </div>

          <div class="col-3">
            <policy-number-input
              aid="gid-input"
              :min-value="0"
              :model-value="gidModel"
              @update:model-value="updateGid"
              label="Group ID (GID)"
              stack-label
              :disable="disableGidInput"
              :tooltip="gidInputTooltip"
              :policy-rules="policyRules?.rules?.specificRunParams?.runAsGid"
            />
          </div>

          <div class="col-5">
            <policy-string-field
              aid="supplemental-groups-input"
              label="Supplementary groups"
              hint="Add multiple groups separated by commas"
              stack-label
              no-error-icon
              :rules="[isSupplementalGroupsValid]"
              :model-value="securityValues.supplementalGroups"
              @update:model-value="updateSupplementalGroups(String($event))"
              :disable="disableSupplementalGroupsInput"
              :tooltip="supplementalGroupsInputTooltip"
              :policy-rules="policyRules?.rules?.specificRunParams?.supplementalGroups"
            />
          </div>
        </div>
      </section>

      <div class="row" v-if="environments.length > 0">
        <runai-sub-expansion-item
          label="Runtime settings"
          aid="env-section-more-settings-btn"
          validation-bar="Review and complete this section"
          :validation-bar-show="showEnvVarValidation"
        >
          <command-and-arguments
            :command-and-args="commandAndArgsModel"
            @update="commandAndArgsChanged"
            :disable="sectionDisabled"
            :args-policy-rules="policyRules?.rules?.specificRunParams?.args"
            :command-policy-rules="policyRules?.rules?.specificRunParams?.command"
          />

          <environment-variables-section
            :key="environmentId"
            :environment-variables="environmentVariableModel"
            :source-value-options="environmentVariableSourceValueOptions"
            @update-environment-variables="environmentVariablesChanged"
            :read-only="sectionDisabled"
            :policy-rules="policyRules?.rules?.specificRunParams?.environmentVariables"
            :policy-defaults="itemizeDefaults?.environmentVariables"
            validate-empty-env-vars
            :support-config-map="supportConfigMapAsEnvVar"
          />
        </runai-sub-expansion-item>
      </div>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// stores
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";

// components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSubExpansionItem } from "@/components/common/runai-sub-expansion-item";
import { CommandAndArguments } from "@/components/environment/command-and-arguments";
import { ToolBox, EToolBoxFormType } from "@/components/environment/form-tools-section/tool-box";
import { RunaiFormCardSection } from "@/components/common";
import { EnvironmentVariablesSection } from "@/components/environment/environment-variables-section";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { ServingEndpointSection } from "@/components/section/inference/serving-endpoint-section";

// models
import type {
  TEnvVarNoneCustomSourceOption,
  IUICommandAndArgs,
  IUIConnection,
  IUidGidSupplementalGroupsModel,
} from "@/models/environment.model";
import type { IUIWorkloadEnvSectionModel, IWorkloadEnvSectionOptions } from "./environment-section.model";
import type { ICardListItem } from "@/components/common/runai-card-list";
import { TCardCmpName } from "@/components/common/runai-card-list";
import {
  type EnvironmentAsset,
  type Connection,
  type SpecificRunConnectionInfo,
  UidGidSource,
  type ComplianceInfoReason,
  InternalConnectionType,
  ToolType,
  type PolicyInfo,
  type WorkspaceSpecificRunParams,
  type TrainingSpecificRunParams,
  type EnvironmentVariableOfAsset,
  type InternalToolInfo,
} from "@/swagger-models/assets-service-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";

// constants
import { errorMessages } from "@/common/error-message.constant";

// services
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import { isCommaSemicolonSeperatedNumber } from "@/common/form.validators";

// utils
import { arrayToObjectMap, fallbackDefaultIfNullOrUndefined } from "@/utils/common.util/common.util";
import { EWorkloadFormType } from "@/models/workload.model";

import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import { environmentSectionUtil } from "./environment-section.util";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSubExpansionItem,
    CommandAndArguments,
    ToolBox,
    RunaiFormCardSection,
    EnvironmentVariablesSection,
    PolicyStringField,
    PolicyNumberInput,
    ServingEndpointSection,
  },
  emits: ["environment-changed", "create-new"],
  props: {
    entityType: {
      type: String as PropType<EWorkloadFormType>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    environmentId: {
      type: String as PropType<string>,
      required: true,
    },
    specificEnv: {
      type: Object as PropType<IUIWorkloadEnvSectionModel>,
      required: false,
    },
    environments: {
      type: Array as PropType<Array<EnvironmentAsset>>,
      required: true,
    },
    sectionOptions: {
      type: Object as PropType<IWorkloadEnvSectionOptions>,
      required: false,
      default: () => ({}),
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    environmentVariableSourceValueOptions: {
      type: Array as PropType<Array<TEnvVarNoneCustomSourceOption>>,
      default: () => [],
    },
    policyRules: {
      type: [Object, null] as PropType<PolicyInfo | null>,
      required: false,
    },
    itemizeDefaults: {
      type: [Object, null] as PropType<ILoadedItemizeDefaults | null>,
      required: false,
    },
    loadedSpecificRunParamsWithDefaults: {
      type: [Object, null] as PropType<WorkspaceSpecificRunParams | TrainingSpecificRunParams | null>,
      required: false,
    },
    supportConfigMapAsEnvVar: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    accessDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    supportingServingPort: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      authStore: useAuthStore(),
      tokenGid: null as number | null,
      tokenUid: null as number | null,
      tokenSupplementalGroups: null as string | null,
      EToolBoxFormType,
    };
  },
  created() {
    this.tokenGid = this.authStore.getGID || null;
    this.tokenUid = this.authStore.getUID || null;
    this.tokenSupplementalGroups = this.authStore.getSupplementaryGroups;
  },
  computed: {
    showEnvVarValidation(): boolean {
      return this.environmentVariableModel.some((envVar: EnvironmentVariableOfAsset) => {
        if (!envVar.name) return true;
        if (envVar.credential) return !envVar.credential.assetId || !envVar.credential.key;
        if (envVar.configMap) return !envVar.configMap.name || !envVar.configMap.key;
        return !envVar.value;
      });
    },
    disableCreateButton(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.Environments, Action.Create);
    },
    uidModel(): number | null {
      if (this.securityValues.runAsUid === -1) return null;
      else return this.securityValues.runAsUid || null;
    },
    gidModel(): number | null {
      if (this.securityValues.runAsGid === -1) return null;
      else return this.securityValues.runAsGid || null;
    },
    selectedEnvironmentId(): Array<string> {
      return this.selectedEnvironment ? [String(this.selectedEnvironment.id)] : [];
    },
    entityText(): string {
      if (!this.entityType || this.entityType === EWorkloadFormType.Template) return "workload";
      return this.entityType;
    },
    isEnvironmentVariablesExist(): boolean {
      return this.environmentVariableModel.length > 0;
    },
    securityValues(): IUidGidSupplementalGroupsModel {
      const environment: EnvironmentAsset | undefined = this.selectedEnvironment?.data as EnvironmentAsset;
      if (!environment) return { runAsUid: null, runAsGid: null, supplementalGroups: null };

      if (environment.spec.overrideUidGidInWorkspace || environment.spec.uidGidSource === UidGidSource.FromIdpToken) {
        return {
          runAsUid: this.specificEnv?.runAsUid || null,
          runAsGid: this.specificEnv?.runAsGid || null,
          supplementalGroups: this.specificEnv?.supplementalGroups || null,
        };
      }

      return {
        runAsUid: environment.spec.runAsUid,
        runAsGid: environment.spec.runAsGid,
        supplementalGroups: environment.spec.supplementalGroups || null,
      };
    },
    uiConnections(): Array<IUIConnection> {
      if (!this.selectedEnvironment) {
        return [];
      } else {
        let envConnections: Array<IUIConnection> = [];
        const environment: EnvironmentAsset = this.selectedEnvironment.data as EnvironmentAsset;
        if (environment.spec.connections) {
          envConnections = environment.spec.connections
            .map((c: Connection) => environmentService.prepareConnectionForUI(c) as IUIConnection)
            .filter(
              (c: IUIConnection) =>
                c.isCustomPort || c.isCustomUrl || c.connectionType === InternalConnectionType.ExternalUrl,
            );
        }

        const specificEnvConnections: Array<SpecificRunConnectionInfo> = this.specificEnv?.connections || [];

        const specificConnectionsByToolName: Record<string, SpecificRunConnectionInfo> =
          arrayToObjectMap<SpecificRunConnectionInfo>(specificEnvConnections, "name");

        const envConnectionsByToolName: Record<string, IUIConnection> = arrayToObjectMap<IUIConnection>(
          envConnections,
          "name",
        );

        const toolBoxConnections: Array<IUIConnection> = [];
        for (const key in envConnectionsByToolName) {
          toolBoxConnections.push({
            ...envConnectionsByToolName[key],
            ...specificConnectionsByToolName[key],
            id: this.makeToolId(envConnectionsByToolName[key]),
          });
        }
        return toolBoxConnections;
      }
    },
    showUidGid(): boolean {
      const environment: EnvironmentAsset | undefined = this.selectedEnvironment?.data as EnvironmentAsset;
      if (!environment) return false;

      const { FromIdpToken, Custom } = UidGidSource;
      return (
        environment.spec.uidGidSource === FromIdpToken ||
        (environment.spec.uidGidSource === Custom && !!environment.spec.overrideUidGidInWorkspace)
      );
    },
    disableUidInput(): boolean {
      return (
        this.sectionDisabled ||
        ((this.selectedEnvironment?.data as EnvironmentAsset).spec.uidGidSource === UidGidSource.FromIdpToken &&
          !!this.tokenUid)
      );
    },
    uidInputTooltip(): string {
      if (this.sectionDisabled) return "";
      const environment: EnvironmentAsset = this.selectedEnvironment?.data as EnvironmentAsset;
      const { uidGidSource } = environment.spec;
      if (uidGidSource === UidGidSource.FromIdpToken) {
        return this.tokenUid
          ? "This value is taken automatically from the IdP."
          : "This value does not exist in the IdP, but can be modified. if left empty, it will be automatically taken from the image.";
      }
      return "The UID, GID, and supplementary groups can be modified. if left empty, they will be automatically taken from the image.";
    },
    disableGidInput(): boolean {
      return (
        this.sectionDisabled ||
        ((this.selectedEnvironment?.data as EnvironmentAsset).spec.uidGidSource === UidGidSource.FromIdpToken &&
          !!this.tokenGid)
      );
    },
    gidInputTooltip(): string {
      if (this.sectionDisabled) return "";
      const environment: EnvironmentAsset = this.selectedEnvironment?.data as EnvironmentAsset;
      const { uidGidSource } = environment.spec;
      if (uidGidSource === UidGidSource.FromIdpToken) {
        return this.tokenGid
          ? "This value is taken automatically from the IdP."
          : "This value does not exist in the IdP, but can be modified. if left empty, it will be automatically taken from the image.";
      }
      return "The UID, GID, and supplementary groups can be modified. if left empty, they will be automatically taken from the image.";
    },
    disableSupplementalGroupsInput(): boolean {
      if (this.sectionDisabled) return true;

      const environment: EnvironmentAsset = this.selectedEnvironment?.data as EnvironmentAsset;
      const { uidGidSource } = environment.spec;
      if (!this.specificEnv?.runAsGid || this.specificEnv?.runAsGid === -1) return true;
      return uidGidSource === UidGidSource.FromIdpToken && !!this.tokenSupplementalGroups;
    },
    supplementalGroupsInputTooltip(): string {
      if (this.sectionDisabled) return "";
      const environment: EnvironmentAsset = this.selectedEnvironment?.data as EnvironmentAsset;
      const { uidGidSource } = environment.spec;
      if (!this.specificEnv?.runAsGid || this.specificEnv?.runAsGid === -1) {
        return "To modify this value , enter a GID first";
      }
      if (uidGidSource === UidGidSource.FromIdpToken) {
        return this.tokenSupplementalGroups
          ? "This value is taken automatically from the IdP."
          : "This value does not exist in the IdP, but can be modified. if left empty, it will be automatically taken from the image.";
      }
      return "The UID, GID, and supplementary groups can be modified. if left empty, they will be automatically taken from the image.";
    },
    environmentSummary(): string {
      return !this.selectedEnvironment ? "None" : (this.selectedEnvironment.data as EnvironmentAsset).meta.name;
    },
    selectedEnvironmentAsset(): EnvironmentAsset | undefined {
      return this.environments.find((env: EnvironmentAsset) => env.meta.id === this.environmentId);
    },
    selectedEnvironment(): ICardListItem | null {
      if (!this.selectedEnvironmentAsset) return null;

      return {
        id: this.selectedEnvironmentAsset.meta.id,
        cardName: TCardCmpName.ENVIRONMENT,
        data: this.selectedEnvironmentAsset,
      };
    },
    commandAndArgsModel(): IUICommandAndArgs {
      return {
        args: this.specificEnv?.args || "",
        command: this.specificEnv?.command || "",
      };
    },
    environmentVariableModel(): Array<EnvironmentVariableOfAsset> {
      return this.specificEnv?.environmentVariables || [];
    },
    environmentCards(): Array<ICardListItem> {
      return this.environments.map((env: EnvironmentAsset) => {
        let disabled = false;
        let tooltip = "";
        let showDisabledInfo = false;
        if (env.compliance && !env.compliance?.compliance) {
          disabled = true;
          // if it does not comply due to policy vs if it does not comply for some other reason
          if (env.compliance?.reason?.some((reason: ComplianceInfoReason) => reason.field)) {
            showDisabledInfo = true;
            tooltip = "This environment can't be used because it doesn't comply with the policy your administrator set.";
          } else if (env.compliance?.reason?.length) {
            tooltip = env.compliance.reason[0].details;
          }
        }

        // Setting searchable values
        const connections = env.spec.connections?.map((connection) => connection.name) || [];
        const name = env.meta.name || "";
        const image = env.spec.image || "";
        return {
          id: env.meta.id,
          cardName: TCardCmpName.ENVIRONMENT,
          data: env,
          searchValues: [name, image, ...connections].filter((val) => !!val),
          disabled,
          tooltip,
          showDisabledInfo,
          sortInfo: {
            name: env.meta.name,
            createdAt: env.meta.createdAt,
            recentUsage: env.usageTimes?.lastUsedByWorkload,
          },
        };
      });
    },
    cardsMainTitle(): string {
      return this.sectionOptions.cardsTitle || "Select the environment for your workload";
    },
    canAddEnvVariable(): boolean {
      return fallbackDefaultIfNullOrUndefined(this.sectionOptions.canAddEnvVariable, true);
    },
    isToolBoxRequired(): boolean {
      return fallbackDefaultIfNullOrUndefined(this.sectionOptions.isToolBoxRequired, true);
    },
    servingPortInternalToolInfo(): InternalToolInfo | undefined {
      if (!this.selectedEnvironmentAsset?.spec.connections) return undefined;

      const servingPortConnection: Connection | undefined = this.selectedEnvironmentAsset.spec.connections.find(
        (c: Connection) => c.internalToolInfo?.connectionType === InternalConnectionType.ServingPort,
      );
      return servingPortConnection?.internalToolInfo || undefined;
    },
    servingEndpointModel(): IServingEndpointAccessData | undefined {
      if (!this.servingPortInternalToolInfo) return;
      const servingPortDetails = {
        protocol: this.servingPortInternalToolInfo?.servingPortInfo?.protocol,
        port: this.servingPortInternalToolInfo?.containerPort,
      };

      if (!this.specificEnv?.servingPortAccess) return servingPortDetails;
      return {
        ...servingPortDetails,
        access: this.specificEnv.servingPortAccess.access,
      };
    },
  },
  methods: {
    createNew(): void {
      this.$emit("create-new");
    },
    makeToolId(tool: IUIConnection): string {
      if (tool.toolType === ToolType.Comet || tool.toolType === ToolType.Wandb)
        return `${this.selectedEnvironment?.id}-${tool.toolType}`;
      return `${this.selectedEnvironment?.id}-${tool.toolType}-${tool.connectionType}`;
    },
    isSupplementalGroupsValid(val: string): boolean | string {
      return isCommaSemicolonSeperatedNumber(val) || errorMessages.SUPPLEMENTAL_GROUPS_NOT_VALID;
    },
    isRequiredEnvironment(val: ICardListItem): boolean | string {
      return val && val.id ? true : errorMessages.SELECT_ENVIRONMENT;
    },
    commandAndArgsChanged(commandAndArgs: IUICommandAndArgs) {
      this.$emit("environment-changed", {
        environmentId: this.environmentId,
        specificEnv: { ...this.specificEnv, ...commandAndArgs },
      });
    },
    environmentVariablesChanged(environmentVariables: Array<EnvironmentVariableOfAsset>): void {
      this.$emit("environment-changed", {
        environmentId: this.environmentId,
        specificEnv: {
          ...this.specificEnv,
          environmentVariables,
        },
      });
    },
    onSelectedEnv(selectedEnvironment: Array<string>): void {
      let environment: EnvironmentAsset | null = null;
      if (selectedEnvironment.length > 0) {
        const environmentCard: ICardListItem | undefined = this.environmentCards.find(
          (item: ICardListItem) => item.id === selectedEnvironment[0],
        );
        if (environmentCard) environment = environmentCard.data as EnvironmentAsset;
      }

      const securitySettings: IUidGidSupplementalGroupsModel = {
        runAsGid: undefined,
        runAsUid: undefined,
        supplementalGroups: undefined,
      };

      if (environment?.spec.uidGidSource === UidGidSource.FromIdpToken) {
        securitySettings.runAsUid = this.tokenUid;
        securitySettings.runAsGid = this.tokenGid;
        securitySettings.supplementalGroups = this.tokenSupplementalGroups;
      } else if (environment?.spec.overrideUidGidInWorkspace) {
        securitySettings.runAsUid =
          environment?.spec.runAsUid || this.loadedSpecificRunParamsWithDefaults?.runAsUid || null;
        securitySettings.runAsGid =
          environment?.spec.runAsGid || this.loadedSpecificRunParamsWithDefaults?.runAsGid || null;
        securitySettings.supplementalGroups =
          environment?.spec.supplementalGroups || this.loadedSpecificRunParamsWithDefaults?.supplementalGroups || null;
      }

      const connections: SpecificRunConnectionInfo[] = environment?.spec.connections
        ? environment.spec.connections
            .filter((c: Connection) => {
              return (
                c.internalToolInfo?.nodePortInfo?.isCustomPort ||
                c.internalToolInfo?.externalUrlInfo?.isCustomUrl ||
                c.internalToolInfo?.connectionType === InternalConnectionType.ExternalUrl
              );
            })
            .map((c: Connection) => {
              const uiConnection = environmentService.prepareConnectionForUI(c) as IUIConnection;
              const { name, externalUrl, nodePort } = uiConnection;
              return { name, externalUrl, nodePort };
            })
        : [];

      const environmentVariables = environmentSectionUtil.mergeEnvironmentVariables(
        environment?.spec.environmentVariables,
        this.itemizeDefaults?.environmentVariables?.instances,
      );

      this.$emit("environment-changed", {
        environmentId: environment?.meta?.id || null,
        specificEnv: {
          args: environment?.spec.args || this.loadedSpecificRunParamsWithDefaults?.args || "",
          command: environment?.spec.command || this.loadedSpecificRunParamsWithDefaults?.command || "",
          environmentVariables,
          connections: connections,
          ...securitySettings,
        },
      });
    },
    updateUid(runAsUidInputValue: number | string | null): void {
      // The input returns empty string when its empty. We need to convert it to null
      let runAsUid: number | null;
      if (runAsUidInputValue === "" || runAsUidInputValue === null) {
        runAsUid = -1;
      } else {
        runAsUid = +runAsUidInputValue;
      }
      this.$emit("environment-changed", {
        environmentId: this.environmentId,
        specificEnv: { ...this.specificEnv, runAsUid },
      });
    },
    updateGid(runAsGidInputValue: number | string | null): void {
      // The input returns empty string when its empty. We need to convert it to null
      let runAsGid: number | null;
      if (runAsGidInputValue === "" || runAsGidInputValue === null) {
        runAsGid = -1;
      } else {
        runAsGid = +runAsGidInputValue;
      }
      // When gid is deleted we reset the supplementalGroups
      const supplementalGroups = runAsGidInputValue === "" ? "" : this.specificEnv?.supplementalGroups;
      this.$emit("environment-changed", {
        environmentId: this.environmentId,
        specificEnv: { ...this.specificEnv, runAsGid, supplementalGroups },
      });
    },
    updateSupplementalGroups(supplementalGroups: string | null): void {
      this.$emit("environment-changed", {
        environmentId: this.environmentId,
        specificEnv: { ...this.specificEnv, supplementalGroups },
      });
    },
    updateConnections(newConnection: IUIConnection): void {
      const connections: Array<SpecificRunConnectionInfo> = this.uiConnections.map((uiConnection: IUIConnection) => {
        const { name, externalUrl, nodePort, authorizedUsers, authorizedGroups } = uiConnection;
        if (newConnection.name === name) {
          if (uiConnection.connectionType === InternalConnectionType.ExternalUrl) {
            return {
              name: newConnection.name,
              externalUrl: newConnection.externalUrl,
              nodePort: newConnection.nodePort,
              authorizedUsers: newConnection.authorizedUsers || null,
              authorizedGroups: newConnection.authorizedGroups || null,
            };
          }
          return {
            name: newConnection.name,
            externalUrl: newConnection.externalUrl,
            nodePort: newConnection.nodePort,
          };
        } else {
          if (uiConnection.connectionType === InternalConnectionType.ExternalUrl) {
            return {
              name,
              externalUrl,
              nodePort,
              authorizedUsers: authorizedUsers || null,
              authorizedGroups: authorizedGroups || null,
            };
          }
          return { name, externalUrl, nodePort };
        }
      });
      this.$emit("environment-changed", {
        environmentId: this.environmentId,
        specificEnv: { ...this.specificEnv, connections },
      });
    },
    onServingPortChanged(servingPort: IServingEndpointAccessData): void {
      this.$emit("environment-changed", {
        environmentId: this.environmentId,
        specificEnv: {
          ...this.specificEnv,
          servingPortAccess: servingPort,
        },
      });
    },
  },
});
</script>
<style lang="scss">
.capitalized {
  text-transform: capitalize;
}
</style>
