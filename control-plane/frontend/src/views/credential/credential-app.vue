<template>
  <router-view />
</template>

<script lang="ts">
import { useCredentialStore } from "@/stores/credential.store";
import { defineComponent } from "vue";

export default defineComponent({
  async created() {
    await useCredentialStore().loadActionPermissions();
  },
});
</script>
