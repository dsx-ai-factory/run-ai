<template>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { usePolicyStore } from "@/stores/policy.store";

export default defineComponent({
  data() {
    return {
      loading: true,
    };
  },
  async created() {
    await usePolicyStore().loadActionPermissions();
  },
});
</script>
