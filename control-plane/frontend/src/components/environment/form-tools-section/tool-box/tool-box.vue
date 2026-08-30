<template>
  <section class="environment-tool-box q-mb-md" v-if="tool.id">
    <q-btn
      aid="remove-tool-button"
      round
      flat
      icon="ph-bold ph-x"
      class="close-btn"
      @click="$emit('removed')"
      v-if="!isAssetBasedWorkloadForm && removeable"
      :disable="disable"
    />

    <section class="row gap q-pa-md top-section">
      <span>
        <tool-type-select
          :disable="isAssetBasedWorkloadForm || disable"
          :tool-type="tool.toolType"
          :options="availableToolTypeOptions"
          @value-changed="updateToolType"
          :not-required="defaultToolBox"
          aid="tool-type-select"
        />
        <q-tooltip v-if="disabledTooltipText" max-width="350px"> {{ disabledTooltipText }} </q-tooltip>
      </span>
      <span :class="showAuthorizedSection ? 'col-5 flex-1' : 'col-6'">
        <q-input
          v-if="tool.toolType === 'custom'"
          :model-value="tool.name"
          @update:model-value="updateToolName($event?.toString() || '')"
          label="Tool name"
          no-error-icon
          stack-label
          placeholder="Enter a name"
          :disable="isAssetBasedWorkloadForm || disable"
          :rules="[nameNotEmpty]"
          aid="tool-type-custom-input"
        />
        <q-input
          aid="tool-type-wandb-input"
          v-else-if="isExternalTool"
          label="URL"
          stack-label
          placeholder="Enter a URL"
          :rules="[urlNotEmpty]"
          no-error-icon
          :model-value="tool.externalUrl"
          :disable="disable"
          @update:model-value="updateExternalUrl($event?.toString() || '')"
        />
        <q-tooltip v-if="disabledTooltipText" max-width="350px"> {{ disabledTooltipText }} </q-tooltip>
      </span>
      <div v-if="showAuthorizedSection && clusterId" class="private-connection-toggle row items-center">
        <connection-access
          :access="toolAccessModel"
          :tool-name="tool.name"
          @save="onToolAccessChanged"
          is-multi-users
          connection-type="tool"
          :disabled="accessDisabled"
          :disabled-tooltip="toolDisabledInfo"
          :policy-rules="connectionAccessPolicyRules"
        />
      </div>
    </section>

    <transition name="slide-fade">
      <section v-if="tool.toolType && !isExternalTool" class="bottom-section row gap justify-between q-pa-md">
        <section class="row gap">
          <connection-type-select
            v-if="tool.connectionType"
            :disable="isAssetBasedWorkloadForm || disable"
            :connection-type="tool.connectionType"
            :options="availableConnectionTypeOptions"
            @value-changed="updateConnectionType"
            :tool-tip-text="disabledTooltipText"
          />
          <template v-if="tool.connectionType === 'ExternalUrl' && tool.isCustomUrl !== null">
            <span>
              <autogen-select
                :disable="isAssetBasedWorkloadForm || disable"
                :init-value="tool.isCustomUrl"
                custom-label="Custom URL"
                @value-changed="updateIsCustomUrl"
              />
              <q-tooltip v-if="disabledTooltipText" max-width="350px"> {{ disabledTooltipText }} </q-tooltip>
            </span>
            <span v-if="tool.isCustomUrl">
              <q-input
                aid="custom-url-input"
                :disable="isEnvironmentForm || disable"
                stack-label
                :input-style="{ minWidth: '220px' }"
                input-class="placeholder-italic"
                placeholder="e.g., https://runai.domain.hpc.org"
                :debounce="300"
                :model-value="tool.externalUrl"
                @update:model-value="updateTool({ ...tool, externalUrl: $event as string })"
                label="URL"
                :rules="isRequired ? [urlNotEmpty, isUrl] : [isUrl]"
                no-error-icon
              />
              <q-tooltip max-width="350px" v-if="customUrlInputText"> {{ customUrlInputText }} </q-tooltip>
            </span>
          </template>
          <template v-if="tool.connectionType === 'NodePort' && tool.isCustomPort !== null">
            <span>
              <autogen-select
                :init-value="tool.isCustomPort"
                :disable="isAssetBasedWorkloadForm || disable"
                custom-label="Custom port"
                @value-changed="updateIsCustomPort"
              />
              <q-tooltip v-if="disabledTooltipText" max-width="350px"> {{ disabledTooltipText }} </q-tooltip>
            </span>
            <span v-if="tool.isCustomPort">
              <q-input
                aid="custom-port-input"
                type="number"
                :debounce="300"
                stack-label
                :input-style="{ minWidth: '220px' }"
                input-class="placeholder-italic"
                placeholder="e.g., 30065"
                :disable="isEnvironmentForm || disable"
                :model-value="tool.nodePort"
                @update:model-value="updateTool({ ...tool, nodePort: Number($event) })"
                label="Node port"
                :rules="isRequired ? [validPort, validPortRange] : [validPortRange]"
                no-error-icon
              />
              <q-tooltip max-width="350px" v-if="nodePortTooltip">{{ nodePortTooltip }}</q-tooltip>
            </span>
          </template>
        </section>
        <span>
          <q-input
            :disable="isAssetBasedWorkloadForm || disable"
            label="Container port"
            :input-style="{ maxWidth: '100px' }"
            type="number"
            :model-value="tool.containerPort"
            @update:model-value="updateTool({ ...tool, containerPort: Number($event) })"
            :rules="containerPortValidationRules"
            no-error-icon
          />
          <q-tooltip v-if="disabledTooltipText" max-width="350px"> {{ disabledTooltipText }} </q-tooltip>
        </span>
        <p v-if="tool.connectionType === InternalConnectionType.LoadBalancer" class="doc-ref-text q-mb-none text-italic">
          Connection handling is managed by the load balancer. For more information, see the
          <a :href="DOCS.WORKLOADS.EXTERNAL_ACCESS_TO_CONTAINERS" target="_blank">External Access to Containers</a> guide
        </p>
      </section>
    </transition>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Utils
import { isNotEmpty, isValidUrl } from "@/common/form.validators";

// Models
import { InternalConnectionType, ToolType } from "@/swagger-models/assets-service-client";
import { errorMessages } from "@/common/error-message.constant";
import type { ISelectOption } from "@/models/global.model";
import {
  connectionTypeOptions,
  toolTypeOptions,
  type IUIConnection,
  type IConnectionTypeSelectOption,
  type IToolTypeSelectOption,
} from "@/models/environment.model";
import {
  type IConnectionAccess,
  EAccessOptions,
} from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import {
  AuthorizationType,
  type ExposedUrlsRules,
  type ExposedUrlsDefaults,
  type SupersetServingPortRules,
} from "@/swagger-models/policy-service-client";

// Cmps
import ConnectionTypeSelect from "../connection-type-select.vue";
import AutogenSelect from "../autogen-select.vue";
import ToolTypeSelect from "../tool-type-select.vue";
import { ConnectionAccess } from "@/components/environment/connection-access";

// Constants
import { DOCS } from "@/constants/docs";

export enum EToolBoxFormType {
  Environment = "environment",
  WorkloadAssetsBased = "workload-assets",
}

export default defineComponent({
  components: {
    ToolTypeSelect,
    ConnectionTypeSelect,
    AutogenSelect,
    ConnectionAccess,
  },
  emits: ["removed", "update-tool"],
  props: {
    defaultToolBox: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    tools: {
      type: Array as PropType<Array<IUIConnection>>,
      required: true,
    },
    tool: {
      type: Object as PropType<IUIConnection>,
      required: true,
    },
    removeable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    formType: {
      type: String as PropType<EToolBoxFormType>,
      required: false,
    },
    toolNames: {
      type: Array as PropType<Array<string>>,
      required: false,
      default: () => [],
    },
    isRequired: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    accessDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    toolDisabledInfo: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<ExposedUrlsRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<ExposedUrlsDefaults | null>,
      required: false,
    },
  },
  data() {
    return {
      externalTools: [ToolType.Wandb, ToolType.Comet] as Array<ToolType>,
      InternalConnectionType: InternalConnectionType,
      DOCS: DOCS,
    };
  },
  created() {
    this.applyPolicyDefaults();
  },
  computed: {
    defaultToolTypeAndConnectionForName(): { toolType: string; connectionType: string } {
      if (!this.tool.name) return { toolType: "", connectionType: "" };
      const [toolType, connectionType] = this.tool.name.split(" - ");
      return { toolType: toolType || "", connectionType: connectionType || "" };
    },
    isEnvironmentForm(): boolean {
      return this.formType === EToolBoxFormType.Environment;
    },
    isAssetBasedWorkloadForm(): boolean {
      return this.formType === EToolBoxFormType.WorkloadAssetsBased;
    },
    isExternalTool(): boolean {
      if (!this.tool.toolType) return false;
      return this.externalTools.includes(this.tool.toolType);
    },
    disabledTooltipText(): string {
      if (this.toolDisabledInfo) return this.toolDisabledInfo;

      if (!this.isAssetBasedWorkloadForm) return "";
      const urlText = "The tool and its connection are defined within the environment. Set the URL here.";
      const portText = "The tool and its connection are defined within the environment. Set the node port here.";
      return this.tool?.isCustomPort ? portText : urlText;
    },
    availableToolTypeOptions(): Array<IToolTypeSelectOption> {
      return toolTypeOptions.map((typeOption: IToolTypeSelectOption) => {
        const usedConnectionTypes = this.getUsedConnectionTypes(typeOption.value as ToolType);
        const allConnectionTypesUsed = this.supportedConnectionTypeOptions.every(
          (connectionTypeOption) => usedConnectionTypes[connectionTypeOption.value],
        );
        if (typeOption.value !== ToolType.Custom && allConnectionTypesUsed)
          return {
            ...typeOption,
            disable: true,
            disabledTooltip: "This tool is already used with all possible connection types",
          };
        return typeOption;
      });
    },
    supportedConnectionTypeOptions(): Array<IConnectionTypeSelectOption> {
      return connectionTypeOptions.filter((typeOption: IConnectionTypeSelectOption) => {
        const loadBalancerUnavailable = this.isAssetBasedWorkloadForm || this.isEnvironmentForm;
        return !(loadBalancerUnavailable && typeOption.value === InternalConnectionType.LoadBalancer);
      });
    },
    availableConnectionTypeOptions(): Array<IConnectionTypeSelectOption> {
      const usedConnectionTypes = this.getUsedConnectionTypes(this.tool.toolType);
      return this.supportedConnectionTypeOptions.map((typeOption: IConnectionTypeSelectOption) => {
        const unavailable = !!usedConnectionTypes[typeOption.value];
        const updatedOption: IConnectionTypeSelectOption = { ...typeOption, disable: unavailable };
        if (unavailable) updatedOption.disabledTooltip = "This connection type is already in use by this tool";
        return updatedOption;
      });
    },
    showAuthorizedSection(): boolean {
      if (this.isEnvironmentForm) return false;
      return (
        this.tool.connectionType === InternalConnectionType.ExternalUrl ||
        this.tool.connectionType === InternalConnectionType.LoadBalancer
      );
    },
    toolAccessModel(): IConnectionAccess {
      let accessOption = EAccessOptions.EVERYONE;
      let authorizedUsers = this.tool.authorizedUsers;
      let authorizedGroups = this.tool.authorizedGroups;
      const isToolNeverConfigured = !this.tool.authorizedUsers && !this.tool.authorizedGroups && !this.tool.accessOption;

      // Get policy defaults to use as fallback values
      const policyDefaults = this.getPolicyDefaultsForTool();
      const policyDefaultUsers = this.policyDefaults?.attributes?.authorizedUsers;
      const policyDefaultGroups = this.policyDefaults?.attributes?.authorizedGroups;

      // If user has explicitly set an access option, use that
      if (this.tool.accessOption) {
        accessOption = this.tool.accessOption as EAccessOptions;
        // Provide both lists to the modal so user can switch between them
        // Use actual tool data if available, otherwise use policy defaults
        authorizedUsers = this.tool.authorizedUsers || policyDefaultUsers || undefined;
        authorizedGroups = this.tool.authorizedGroups || policyDefaultGroups || undefined;
      } else if (isToolNeverConfigured) {
        // Tool has never been configured, use policy defaults if toolType matches
        if (policyDefaults) {
          authorizedUsers = policyDefaults.authorizedUsers;
          authorizedGroups = policyDefaults.authorizedGroups;
          accessOption = policyDefaults.accessOption;
        }
      } else {
        // Tool has access configured but no explicit option, determine from data
        if (this.tool.authorizedUsers) {
          accessOption = EAccessOptions.SPECIFIC_USERS;
          // Also provide groups from policy defaults if available
          authorizedGroups = policyDefaultGroups || undefined;
        } else if (this.tool.authorizedGroups) {
          accessOption = EAccessOptions.GROUPS;
          // Also provide users from policy defaults if available
          authorizedUsers = policyDefaultUsers || undefined;
        }
      }

      return {
        authorizedUsers,
        authorizedGroups,
        accessOption,
      };
    },
    nodePortTooltip(): string {
      if (this.isEnvironmentForm) {
        return "The Node port is configured within the workload using this environment by the workload creator";
      }
      if (this.disable && this.toolDisabledInfo) return this.toolDisabledInfo;

      return "";
    },
    customUrlInputText(): string {
      if (this.isEnvironmentForm) {
        return "The URL is configured by the workload creator within the workload using this environment";
      }
      if (this.disable && this.toolDisabledInfo) return this.toolDisabledInfo;

      return "";
    },
    isToolTypeInPolicyDefaults(): boolean {
      return this.policyDefaults?.attributes?.toolType === this.tool.toolType;
    },
    shouldApplyPolicyDefaults(): boolean {
      // If no policy defaults available, don't apply
      if (!this.policyDefaults?.attributes) return false;

      // If policy doesn't specify a toolType, apply to all tools (general policy)
      if (!this.policyDefaults.attributes.toolType) return true;

      // If current tool has no toolType set, don't apply policy defaults
      if (!this.tool.toolType) return false;

      // Apply policy defaults only if tool types match (specific tool policy)
      return this.isToolTypeInPolicyDefaults;
    },
    connectionAccessPolicyRules(): SupersetServingPortRules | null {
      // No policy rules or defaults available
      if (!this.policyRules?.attributes && !this.policyDefaults?.attributes) return null;

      // Only apply policy rules if toolType matches (if specified in policy)
      if (!this.shouldApplyPolicyDefaults) return null;

      // Transform ExposedUrl policy to ServingPortRules format for connection-access compatibility
      return {
        // Use rules for constraints (canEdit, required, etc.)
        authorizedUsers: this.policyRules?.attributes?.authorizedUsers || null,
        authorizedGroups: this.policyRules?.attributes?.authorizedGroups || null,
        authorizationType: null, // ExposedUrlRules doesn't have authorizationType constraints
        container: this.policyRules?.attributes?.container || null,
        protocol: null,
      };
    },
    containerPortValidationRules(): Array<(val: number) => boolean | string> {
      if (this.tool.connectionType === InternalConnectionType.LoadBalancer) {
        return [this.validLoadBalancerPort];
      }
      return [this.validPort];
    },
  },
  methods: {
    getPolicyDefaultsForTool(): {
      authorizedUsers?: string[];
      authorizedGroups?: string[];
      accessOption: EAccessOptions;
    } | null {
      if (!this.shouldApplyPolicyDefaults) {
        return null;
      }

      const policyDefaultUsers = this.policyDefaults?.attributes?.authorizedUsers;
      const policyDefaultGroups = this.policyDefaults?.attributes?.authorizedGroups;

      if (policyDefaultUsers?.length) {
        return {
          authorizedUsers: policyDefaultUsers,
          accessOption: EAccessOptions.SPECIFIC_USERS,
        };
      } else if (policyDefaultGroups?.length) {
        return {
          authorizedGroups: policyDefaultGroups,
          accessOption: EAccessOptions.GROUPS,
        };
      }

      return null;
    },
    getUsedConnectionTypes(toolType: ToolType | null): Record<InternalConnectionType, number> {
      return this.tools.reduce(
        (acc, { id, toolType: existingToolType, connectionType }) => {
          if (
            id === this.tool.id ||
            existingToolType !== toolType ||
            !connectionType ||
            existingToolType === ToolType.Custom
          )
            return acc;
          acc[connectionType] = 1;
          return acc;
        },
        {} as Record<InternalConnectionType, number>,
      );
    },
    updateIsCustomUrl(val: boolean): void {
      this.updateTool({ ...this.tool, isCustomUrl: val, externalUrl: null, nodePort: undefined });
    },
    updateIsCustomPort(val: boolean): void {
      this.updateTool({ ...this.tool, isCustomPort: val, nodePort: null, externalUrl: undefined });
    },
    updateTool(updatedTool: IUIConnection): void {
      this.$emit("update-tool", updatedTool);
    },
    nameNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    urlNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.URL_NOT_EMPTY;
    },
    isUrl(val: string): boolean | string {
      if (!val) return true;
      return isValidUrl(val) || errorMessages.INVALID_URL;
    },
    validPort(val: number): boolean | string {
      return !!val || errorMessages.ENTER_A_PORT;
    },
    validPortRange(val: number): boolean | string {
      if (!val) return true;
      const message = errorMessages.ENTER_PORT_IN_RANGE.replace("${min}", "30000").replace("${max}", "32767");
      return (val >= 30000 && val <= 32767) || message;
    },
    validLoadBalancerPort(val: number): boolean | string {
      if (!val) return errorMessages.ENTER_A_PORT;
      const message = errorMessages.ENTER_PORT_IN_RANGE.replace("${min}", "1").replace("${max}", "65535");
      return (val >= 1 && val <= 65535) || message;
    },
    updateConnectionType(connectionType: ISelectOption): void {
      const updatedTool = { ...this.tool };
      switch (connectionType.value) {
        case InternalConnectionType.ExternalUrl:
          updatedTool.isCustomUrl = false;
          updatedTool.isCustomPort = undefined;
          break;
        case InternalConnectionType.LoadBalancer:
          updatedTool.isCustomUrl = undefined;
          updatedTool.isCustomPort = undefined;
          // Keep authorization settings for LoadBalancer (like ExternalUrl)
          break;
        case InternalConnectionType.NodePort:
          updatedTool.isCustomPort = false;
          updatedTool.isCustomUrl = undefined;
          updatedTool.authorizedGroups = undefined;
          updatedTool.authorizedUsers = undefined;
          updatedTool.accessOption = undefined;
          break;
        default:
          updatedTool.isCustomPort = undefined;
          updatedTool.isCustomUrl = undefined;
          updatedTool.authorizedGroups = undefined;
          updatedTool.authorizedUsers = undefined;
          updatedTool.accessOption = undefined;
      }
      updatedTool.connectionType = connectionType.value as InternalConnectionType;

      if (this.tool.toolType !== ToolType.Custom) {
        updatedTool.name = this.getToolName({ connectionTypeForName: connectionType.label });
      }
      this.updateTool(updatedTool);
    },
    updateToolType(toolType: ISelectOption): void {
      const toolTypePortMap: Record<string, number> = {
        tensorboard: 6006,
        "jupyter-notebook": 8888,
        "visual-studio-code": 8080,
        mlflow: 5000,
        rstudio: 8787,
        custom: 8080,
        matlab: 8888,
        "chatbot-ui": 3000,
      };

      const updatedTool = {
        ...this.tool,
        containerPort: toolTypePortMap[toolType.value as ToolType] || 0,
        toolType: toolType.value as ToolType,
      };
      if (toolType.value !== ToolType.Wandb && toolType.value !== ToolType.Comet) {
        // Calculate available connection types for the new tool type
        const newToolType = toolType.value as ToolType;
        const usedConnectionTypes = this.getUsedConnectionTypes(newToolType);

        const availableOptionsForNewToolType = connectionTypeOptions.map((typeOption: IConnectionTypeSelectOption) => {
          const unavailable = !!usedConnectionTypes[typeOption.value];
          return { ...typeOption, disable: unavailable };
        });

        // Check if the current connection type is still available for the new tool type
        const currentConnectionTypeOption = availableOptionsForNewToolType.find(
          (option) => option.value === this.tool.connectionType,
        );

        // If current connection type is disabled (unavailable) or doesn't exist, set to first available
        if (!currentConnectionTypeOption || currentConnectionTypeOption.disable) {
          const firstAvailableOption = availableOptionsForNewToolType.find((option) => !option.disable);
          if (firstAvailableOption) {
            updatedTool.connectionType = firstAvailableOption.value as InternalConnectionType;
          }
        }
      }
      // Get the connection type name for the updated connection type
      const connectionTypeName = updatedTool.connectionType
        ? connectionTypeOptions.find((opt) => opt.value === updatedTool.connectionType)?.label
        : undefined;

      updatedTool.name = this.getToolName({
        toolTypeForName: toolType.label,
        connectionTypeForName: connectionTypeName,
      });
      this.updateTool(updatedTool);
    },
    getToolName(opt?: { toolTypeForName?: string; connectionTypeForName?: string }): string {
      const toolType = opt?.toolTypeForName ? this.getToolTypeFromLabel(opt.toolTypeForName) : this.tool.toolType;

      if (toolType === ToolType.Custom) {
        return "";
      } else {
        let toolName = opt?.toolTypeForName || this.defaultToolTypeAndConnectionForName.toolType;

        // For external tools (Wandb, Comet), add a number suffix if there are multiple
        if (toolType === ToolType.Wandb || toolType === ToolType.Comet) {
          const existingToolsOfSameType = this.tools.filter((t) => t.id !== this.tool.id && t.toolType === toolType);
          if (existingToolsOfSameType.length > 0) {
            toolName += ` ${existingToolsOfSameType.length + 1}`;
          }
        }
        // For non-external tools, append connection type unless it's ExternalUrl
        else if (
          opt?.connectionTypeForName &&
          opt.connectionTypeForName !== connectionTypeOptions[0].label // External URL
        ) {
          toolName += ` - ${opt.connectionTypeForName}`;
        }
        return toolName;
      }
    },
    getToolTypeFromLabel(label: string): ToolType | null {
      const option = toolTypeOptions.find((opt) => opt.label === label);
      return option ? option.value : null;
    },
    onToolAccessChanged(toolAccess: IConnectionAccess): void {
      // Explicitly clear the unused field based on the selected access option
      let authorizedUsers = toolAccess.authorizedUsers;
      let authorizedGroups = toolAccess.authorizedGroups;

      switch (toolAccess.accessOption) {
        case EAccessOptions.GROUPS:
          // User selected groups, clear users
          authorizedUsers = null;
          break;
        case EAccessOptions.SPECIFIC_USERS:
        case EAccessOptions.PRIVATE:
          // User selected users, clear groups
          authorizedGroups = null;
          break;
        case EAccessOptions.EVERYONE:
        case EAccessOptions.PUBLIC:
          // User selected everyone/public, clear both
          authorizedUsers = null;
          authorizedGroups = null;
          break;
      }

      this.updateTool({
        ...this.tool,
        authorizedUsers,
        authorizedGroups,
        accessOption: toolAccess.accessOption, // Store the user's explicit choice
      });
    },
    updateToolName(name: string): void {
      this.updateTool({ ...this.tool, name });
    },
    updateExternalUrl(url: string): void {
      this.updateTool({ ...this.tool, externalUrl: url });
    },
    applyAuthorizationTypeDefaults(
      authorizationType: AuthorizationType,
      policyDefaultUsers: string[] | null | undefined,
      policyDefaultGroups: string[] | null | undefined,
    ): {
      accessOption: EAccessOptions;
      authorizedUsers: string[] | null;
      authorizedGroups: string[] | null;
    } {
      switch (authorizationType) {
        case AuthorizationType.AuthenticatedUsers:
          // Set to "All authenticated users", clear both users and groups
          return {
            accessOption: EAccessOptions.EVERYONE,
            authorizedUsers: null,
            authorizedGroups: null,
          };

        case AuthorizationType.AuthorizedUsers:
          // Set to "Specific users", use users defaults or empty array, clear groups
          return {
            accessOption: EAccessOptions.SPECIFIC_USERS,
            authorizedUsers: policyDefaultUsers?.length ? policyDefaultUsers : [],
            authorizedGroups: null,
          };

        case AuthorizationType.AuthorizedGroups:
          // Set to "Groups", use groups defaults or empty array, clear users
          return {
            accessOption: EAccessOptions.GROUPS,
            authorizedUsers: null,
            authorizedGroups: policyDefaultGroups?.length ? policyDefaultGroups : [],
          };

        default:
          // Should never reach here, but return safe defaults
          return {
            accessOption: EAccessOptions.EVERYONE,
            authorizedUsers: null,
            authorizedGroups: null,
          };
      }
    },
    applyPolicyDefaults(): void {
      // Only apply defaults if tool doesn't have access configured and policy allows editing
      if (this.tool.authorizedUsers || this.tool.authorizedGroups || this.tool.accessOption) {
        return; // Tool already has access configured or user made explicit choice
      }

      // Only apply if policy defaults should apply to this tool type
      if (!this.shouldApplyPolicyDefaults) {
        return; // Policy defaults don't apply to this tool type
      }

      // Check if policy rules allow editing (only apply defaults if we can edit)
      const canEditUsers = this.policyRules?.attributes?.authorizedUsers?.canEdit !== false;
      // Note: ExposedUrlRules doesn't support authorizedGroups constraints, so groups are always editable

      const policyDefaultUsers = this.policyDefaults?.attributes?.authorizedUsers;
      const policyDefaultGroups = this.policyDefaults?.attributes?.authorizedGroups;
      const authorizationType = this.policyDefaults?.attributes?.authorizationType;

      const updatedTool = { ...this.tool };
      let hasUpdates = false;

      // If authorizationType is specified, use it to determine which authorization to apply
      if (authorizationType) {
        const authDefaults = this.applyAuthorizationTypeDefaults(
          authorizationType,
          policyDefaultUsers,
          policyDefaultGroups,
        );
        updatedTool.accessOption = authDefaults.accessOption;
        updatedTool.authorizedUsers = authDefaults.authorizedUsers;
        updatedTool.authorizedGroups = authDefaults.authorizedGroups;
        hasUpdates = true;
      } else {
        // Fallback to original logic if authorizationType is not specified
        // Apply default users if available and editable
        if (policyDefaultUsers?.length && canEditUsers) {
          updatedTool.authorizedUsers = policyDefaultUsers;
          updatedTool.accessOption = EAccessOptions.SPECIFIC_USERS;
          hasUpdates = true;
        }

        // Apply default groups if available (and no users were set)
        // Note: Groups are always editable for tools since ExposedUrlRules doesn't constrain them
        if (policyDefaultGroups?.length && !policyDefaultUsers?.length) {
          updatedTool.authorizedGroups = policyDefaultGroups;
          updatedTool.accessOption = EAccessOptions.GROUPS;
          hasUpdates = true;
        }
      }

      // Update the tool if we have changes
      if (hasUpdates) {
        this.updateTool(updatedTool);
      }
    },
  },
  watch: {
    // Apply policy defaults when policy changes
    policyDefaults: {
      handler() {
        this.applyPolicyDefaults();
      },
      deep: true,
      immediate: false,
    },
  },
});
</script>

<style lang="scss" scoped>
.environment-tool-box {
  position: relative;

  .close-btn {
    position: absolute;
    right: 10px;
    top: 10px;
    color: $black-54;
    font-size: 12px;
  }

  .gap {
    gap: 35px;
  }

  &:has(.doc-ref-text) {
    .gap {
      gap: 25px;
    }
  }

  .bottom-section,
  .top-section {
    border: 1px solid $black-12;
  }

  .bottom-section {
    background-color: $body-background-color;
    border-top: none;
  }

  .doc-ref-text {
    font-size: 13px;
    margin-bottom: 0;
  }
}
</style>
