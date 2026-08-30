<template>
  <div class="idp-mappers">
    <div class="row items-center q-mb-md">
      <span class="q-mr-xs">Map the following user attributes as defined by the identity provider</span
      ><span class="text-italic">(optional)</span>
      <runai-tooltip
        aid="tool-tip-idp-mappers"
        tooltip-text="Mapping enriches Run:ai users with attributes from the identity provider."
        width="400px"
        tooltip-position="right"
      />
    </div>
    <div class="row title">
      <div class="col-4">Attribute</div>
      <div class="col-8">Value in the IDP</div>
    </div>
    <mapper-item
      :value="idpMapper.value"
      v-for="idpMapper in idpMappers"
      :label="idpMapper.label"
      @update-value="idpMapper.value = $event"
      :key="idpMapper.label"
      :disabled="disabled"
    ></mapper-item>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import MapperItem from "@/components/settings/sections/security/sso-settings/idps/idp-mappers/mapper-item/mapper-item.vue";
import type { IIdpMapperItem } from "@/models/setting.model";

export default defineComponent({
  name: "idp-mappers",
  components: { MapperItem, RunaiTooltip },
  props: {
    idpMappers: {
      type: Object as PropType<Record<string, IIdpMapperItem>>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
  },
  data() {
    return {};
  },
});
</script>
<style lang="scss" scoped>
.idp-mappers {
  .title {
    font-weight: 500;
    font-size: 16px;
  }
  .label {
    cursor: default;
  }
  .arrow {
    padding: 20px 0 0 20px;
  }
}
</style>
