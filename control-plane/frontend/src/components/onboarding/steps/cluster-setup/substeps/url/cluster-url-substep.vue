<template>
  <onboarding-content-wrapper>
    <template #title>Set your cluster URL</template>
    <template #subtitle>
      Enter the Kubernetes cluster's URL. It will only be accessible within the organization network. <br />
      For more information, see the
      <a :href="DOCS.INSTALLATION.FQDN_REQUIREMENTS" target="_blank">Fully Qualified Domain Name (FQDN)</a>
    </template>
    <template #content>
      <div class="row">
        <div class="col-7 cluster-url-wrapper">
          <q-input
            :readonly="readOnly"
            aid="cluster-url"
            :model-value="clusterUrl"
            @update:model-value="handleClusterUrlChange"
            class="cluster-url-input"
            label="Cluster URL"
            :rules="[isNotEmpty, isValidUrl]"
            no-error-icon
          />
        </div>
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent, computed, type PropType } from "vue";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { errorMessages } from "@/common/error-message.constant";
import { formValidators } from "@/common/form.validators";
import DOCS from "@/constants/docs";

export default defineComponent({
  name: "cluster-url-substep",
  components: { OnboardingContentWrapper },
  emits: {
    "update:cluster-url": (url: string) => url,
  },
  props: {
    clusterUrl: {
      type: String as PropType<string>,
      default: "",
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  setup(props, { emit }) {
    // Validation functions
    const isNotEmpty = (value: string) => {
      return !!value.trim() || errorMessages.URL_NOT_EMPTY;
    };

    const isValidUrl = (value: string) => {
      // Use global validator first
      if (formValidators.isValidUrl(value)) {
        return true;
      }

      // For development: also allow localhost URLs
      const localhostRegex = /^https?:\/\/(localhost|127\.0\.0\.1)(:\d+)?(\/.*)?$/;
      return localhostRegex.test(value) || errorMessages.INVALID_URL;
    };

    const isInvalid = computed(() => {
      return !props.clusterUrl || props.clusterUrl.trim().length === 0;
    });

    // Event handlers
    const handleClusterUrlChange = (url: string | number | null) => {
      emit("update:cluster-url", url ? String(url) : "");
    };

    return {
      isInvalid,
      isNotEmpty,
      isValidUrl,
      handleClusterUrlChange,
      DOCS,
    };
  },
});
</script>

<style scoped lang="scss">
.cluster-url-wrapper {
  max-width: 700px;
}
</style>
