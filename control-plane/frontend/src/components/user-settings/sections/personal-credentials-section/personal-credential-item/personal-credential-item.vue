<template>
  <section class="personal-credential-item q-pa-md row justify-between q-mb-md">
    <div class="flex-1 q-mr-md">
      <component :is="component" :credential="credential" />
    </div>
    <div class="flex items-center">
      <q-btn
        class="delete-credential-btn runai-icon-btn--default"
        :aid="`delete-credential-${credential.meta.name}`"
        icon="ph-fill ph-trash"
        flat
        round
        color="black-54"
        @click="
          $emit('delete-credential', { kind: credential.meta.kind, id: credential.meta.id, name: credential.meta.name })
        "
      >
        <q-tooltip> Delete credential </q-tooltip>
      </q-btn>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { CredentialNameAndDescriptionPreviewBox } from "../credential-name-and-description-preview-box";
import { DockerRegistryBox } from "../docker-registry-box";
import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";
import { AssetKind } from "@/swagger-models/assets-service-client";

export default defineComponent({
  name: "personal-credential-item",
  components: {
    CredentialNameAndDescriptionPreviewBox,
    DockerRegistryBox,
  },
  props: {
    credential: {
      type: Object as PropType<CredentialsListResponseEntry>,
      required: true,
    },
  },
  data() {
    return {
      AssetKind,
    };
  },
  computed: {
    component(): typeof CredentialNameAndDescriptionPreviewBox | typeof DockerRegistryBox | null {
      switch (this.credential.meta.kind) {
        case AssetKind.DockerRegistry:
          return DockerRegistryBox;
        case AssetKind.GenericSecret:
        case AssetKind.NgcApiKey:
          return CredentialNameAndDescriptionPreviewBox;
        default:
          return null;
      }
    },
  },
});
</script>
<style lang="scss" scoped>
.personal-credential-item {
  background-color: $body-background-color;
}
</style>
