<template>
  <section class="runai-drawer-content main-drawer-layout">
    <section class="drawer-layout-header" :class="{ 'dense-header': denseHeader }">
      <slot name="header" />
    </section>

    <section class="drawer-layout-subheader" v-if="$slots.subheader">
      <slot name="subheader" />
    </section>

    <section class="drawer-layout-body">
      <slot name="body" />
    </section>
  </section>
</template>
<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
export default defineComponent({
  props: {
    // dense header should be passed where tables are not multiple selection since their header is 48px.
    // in multiple selection tables, the header is 55px.
    denseHeader: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>
<style lang="scss">
$resizer-width: 14px;
.main-drawer-layout {
  display: flex;
  height: 100%;
  flex-direction: column;
  flex: 1;
  .drawer-layout-header {
    display: flex;
    gap: 10px;
    align-items: center;
    min-height: 55px;
    max-height: 55px;
    padding-right: 50px;
    display: flex;
    border-bottom: solid 1px $black-10;
    align-items: center;
    padding-left: $resizer-width;
    font-size: 16px;
    font-weight: 500;
  }
  .dense-header {
    min-height: 48px;
    max-height: 48px;
  }
  .drawer-layout-subheader {
    height: 64px;
    display: flex;
    border-bottom: solid 1px $black-10;
    align-items: center;
    font-size: 16px;
    font-weight: 500;
    .q-tabs {
      height: 100%;
    }
  }

  .drawer-layout-body {
    overflow: hidden;
    height: 100%;
    .q-tab-panels.q-panel-parent {
      height: 100%;
    }
  }
}
</style>
