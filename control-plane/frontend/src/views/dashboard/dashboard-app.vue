<template>
  <section class="dashboard-app">
    <section class="dashboard-view">
      <router-view />
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({});
</script>

<style lang="scss" scoped>
.dashboard-app {
  display: flex;
  flex-direction: column;
  flex: 1;

  .dashboard-view {
    flex: 1;
    display: flex;
    flex-direction: column;

    > :first-child {
      flex: 1;
    }
  }
}
</style>
