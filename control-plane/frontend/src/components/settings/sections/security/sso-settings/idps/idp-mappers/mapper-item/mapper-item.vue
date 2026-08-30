<template>
  <div class="mapper-item q-mb-md">
    <div class="row">
      <div class="col-3 label">
        <q-input :placeholder="label" no-error-icon stack-label disable :model-value="null" />
      </div>
      <div class="col-1 align-content-center">
        <span><q-icon name="ph-bold ph-arrow-right" size="16px" class="q-mr-md arrow text-icon-default" /></span>
      </div>
      <div class="col-4">
        <setting-readonly-input :resource-type="securitySettingsResourceType">
          <template v-slot:default="{ readonly }">
            <q-input
              :model-value="value"
              @update:model-value="updateValue"
              no-error-icon
              stack-label
              aid="mapper-value"
              :disable="readonly || disabled"
            />
          </template>
        </setting-readonly-input>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "mapper-item",
  components: { SettingReadonlyInput },
  emits: ["update-value"],
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
    value: {
      type: String as PropType<string>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
  },
  computed: {
    securitySettingsResourceType(): keyof typeof ResourceType {
      return "SecuritySettings";
    },
  },
  methods: {
    updateValue(value: string | null | number): void {
      this.$emit("update-value", value);
    },
  },
});
</script>
<style lang="scss" scoped>
.mapper-item {
  .label {
    cursor: default !important;
  }
  .arrow {
    padding: 20px 0 0 20px;
  }
}
</style>
<style lang="scss">
.mapper-item {
  .q-field__inner {
    cursor: default !important;
  }
}
</style>
