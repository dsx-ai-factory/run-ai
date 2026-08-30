<template>
  <runai-system-page>
    <template #header-end>
      <span v-if="errorNum" class="error text-h3">{{ errorNum }}</span>
    </template>
    <router-view class="flex-1 q-mx-xl" />
  </runai-system-page>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiSystemPage } from "@/components/common/runai-system-page";

export default defineComponent({
  components: {
    RunaiSystemPage,
  },
  data() {
    return {
      errorNum: this.$route.meta?.pageTitle,
    };
  },
});
</script>

<style lang="scss" scoped>
.error {
  color: $black-12;
  font-weight: 700;
}
</style>
