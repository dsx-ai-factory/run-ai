<template>
  <div class="idp-scope">
    <div class="row items-center">
      <span class="q-mr-md">Set the OIDC scopes <span class="optional">(optional)</span></span>
      <runai-tooltip
        aid="tooltip-scope-settings"
        :tooltip-text="scopeTooltipText"
        width="400px"
        tooltip-position="right"
      />
    </div>
    <setting-readonly-input :resource-type="securityResourceType">
      <template v-slot:default="{ readonly }">
        <runai-addable-multi-select
          label="Scopes"
          new-option-label="scope"
          @options-select="updateSelectedScope"
          @add-option="(name: string, callback: () => void) => addScope(name, callback)"
          :options="scopeOptions"
          :selected-options="selectedScopeOptions"
          :disable="readonly || disabled"
        />
      </template>
    </setting-readonly-input>
    <div v-if="scopeInputError" class="scopeError">Enter scopes containing only letters, numbers, or - _ : .</div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { DEFAULT_SCOPE_OPTIONS, OPENID_SCOPE, SCOPE_REGEX } from "../idp-common/idp-common";
import type { ISelectOption } from "@/models/global.model";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "idp-scope",
  components: { SettingReadonlyInput, RunaiAddableMultiSelect, RunaiTooltip },
  emits: ["update-scope"],
  props: {
    idpName: {
      type: String as PropType<string>,
      required: true,
    },
    scopes: {
      type: Array as PropType<Array<string>>,
      default: [] as ISelectOption[],
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      selectedScopeOptions: [] as ISelectOption[],
      scopeInputError: false,
    };
  },
  created() {
    this.setSelectedScopeOptions();
  },
  computed: {
    scopeOptions(): ISelectOption[] {
      const scopeOptions: ISelectOption[] = [];
      this.scopes.forEach((name: string) => {
        if (this.isOidc && name !== OPENID_SCOPE) {
          scopeOptions.push({ label: name, value: name });
        }
      });
      return scopeOptions;
    },
    scopeTooltipText(): string {
      return `The ${this.idpName} scopes to be used during authentication to authorize access to a user's details. Each scope returns a set of user attributes. Make sure they match the names in your identity provider.`;
    },
    isOidc(): boolean {
      return this.idpName === IdpTypeEnum.Oidc.toUpperCase();
    },
    securityResourceType(): keyof typeof ResourceType {
      return "SecuritySettings";
    },
  },
  methods: {
    updateSelectedScope(updatedScope: ISelectOption[]): void {
      this.scopeInputError = false;
      if (this.isOidc) {
        this.handleOidcScope(updatedScope);
      } else {
        this.selectedScopeOptions = updatedScope;
      }
      this.emitScopeUpdate();
    },
    handleOidcScope(updatedScope: ISelectOption[]): void {
      if (updatedScope.length > 0) {
        if (!updatedScope.find((option) => option.value === OPENID_SCOPE)) return;
        this.selectedScopeOptions = updatedScope;
      } else {
        this.selectedScopeOptions = [DEFAULT_SCOPE_OPTIONS];
      }
      this.emitScopeUpdate();
    },
    isScopeValid(scopeName: string): boolean {
      return SCOPE_REGEX.test(scopeName);
    },
    async addScope(scopeName: string, callback: () => void): Promise<void> {
      if (this.isScopeExist(scopeName)) return;
      if (this.isScopeValid(scopeName)) {
        this.scopeInputError = false;
        const option = { label: scopeName, value: scopeName };
        this.scopeOptions.push(option);
        await this.$nextTick();
        this.selectedScopeOptions.push(option);
      } else {
        this.scopeInputError = true;
      }
      this.emitScopeUpdate();
      callback();
    },
    isScopeExist(scopeName: string): boolean {
      return (
        this.selectedScopeOptions.find(
          (option: ISelectOption) => option.value === scopeName || (this.isOidc && scopeName === OPENID_SCOPE),
        ) !== undefined
      );
    },
    setSelectedScopeOptions(): void {
      this.selectedScopeOptions = this.scopes.map((name) => ({
        label: name,
        value: name,
        disable: name === OPENID_SCOPE,
      })) as ISelectOption[];
    },
    emitScopeUpdate(): void {
      this.$emit(
        "update-scope",
        this.selectedScopeOptions.map((option: ISelectOption) => option.value),
      );
    },
  },
});
</script>
<style lang="scss" scoped>
.idp-scope {
  .scopeError {
    color: $negative;
    font-size: 12px;
  }
  .optional {
    font-style: italic;
  }
}
</style>
