<template>
  <div class="sso-step">
    <sub-step-transition>
      <protocol-substep
        v-if="subStep === SsoSubstep.Protocol"
        key="protocol-substep"
        ref="protocolSubstepRef"
        :sso-data="ssoData"
        @update:sso-data="handleSsoDataUpdate"
      />
      <idp-metadata-substep
        v-else-if="subStep === SsoSubstep.IdpMetadata"
        key="idp-metadata-substep"
        ref="idpMetadataSubstepRef"
        :idp-metadata="ssoData.idpMetadata"
        :sso-data="ssoData"
        @update:idp-metadata="handleSsoDataUpdate('idpMetadata', $event)"
        @update:sso-data="handleSsoDataUpdate"
      />
      <runai-metadata-substep
        v-else-if="subStep === SsoSubstep.RunaiMetadata"
        key="runai-metadata-substep"
        ref="runaiMetadataSubstepRef"
        :runai-metadata="ssoData.runaiMetadata"
        :sso-data="ssoData"
        @update:runai-metadata="handleSsoDataUpdate('runaiMetadata', $event)"
      />
    </sub-step-transition>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import ProtocolSubstep from "@/components/onboarding/steps/sso/substeps/protocol/protocol-substep.vue";
import IdpMetadataSubstep from "@/components/onboarding/steps/sso/substeps/idp-metadata/idp-metadata-substep.vue";
import RunaiMetadataSubstep from "@/components/onboarding/steps/sso/substeps/runai-metadata/runai-metadata-substep.vue";
import { SubStepTransition } from "@/components/onboarding/common/sub-step-transition";
import type { SsoData, SsoDataKey } from "@/components/onboarding/models/admin-onboarding.model";
import { SsoSubstep } from "@/components/onboarding/models/admin-onboarding.model";
import { ssoDataUpdateEmitValidator } from "@/components/onboarding/models/admin-onboarding.model";

export default defineComponent({
  name: "sso-step",
  components: {
    ProtocolSubstep,
    IdpMetadataSubstep,
    RunaiMetadataSubstep,
    SubStepTransition,
  },
  emits: {
    "update:sso-data": ssoDataUpdateEmitValidator,
  },
  props: {
    subStep: {
      type: Number as PropType<number>,
      required: true,
    },
    ssoData: {
      type: Object as PropType<SsoData>,
      required: true,
    },
  },
  data() {
    return {
      SsoSubstep,
    };
  },
  methods: {
    handleSsoDataUpdate<K extends SsoDataKey>(key: K, value: SsoData[K]) {
      this.$emit("update:sso-data", key, value);
    },
  },
});
</script>

<style scoped lang="scss">
.sso-step {
  width: 100%;
  height: 100%;
}
</style>
