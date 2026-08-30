<template>
  <section class="runai-table-wrapper">
    <section :class="{ 'runai-sticky': sticky }">
      <slot name="banner" />
      <slot name="actions" />
      <section ref="table"><slot name="table" /></section>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import type { IFilterBy } from "@/models/filter.model";
import { useClusterStore } from "@/stores/cluster.store";

const defaultDistanceToTop = 188;
const headerHeight = 120;
export default defineComponent({
  props: {
    filtersObject: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
    sticky: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
  },
  data() {
    return {
      tableHeight: `calc(100vh - ${defaultDistanceToTop}px)`,
      clusterStore: useClusterStore(),
    };
  },
  mounted() {
    if (this.sticky) {
      addEventListener("resize", this.setTableHeight);
    }
  },
  methods: {
    setTableHeight(): void {
      this.$nextTick(() => {
        const stickyElement = this.$refs.table as HTMLElement;
        if (stickyElement) {
          const distToTop: number = stickyElement.getBoundingClientRect().top;
          const tableMiddleElement: HTMLElement = this.$el?.getElementsByClassName("q-table__middle")[0];
          if (tableMiddleElement) {
            tableMiddleElement.style.maxHeight = `calc(100vh - ${distToTop + headerHeight}px)`;
          }
        }
      });
    },
  },
  watch: {
    filtersObject: {
      handler() {
        this.setTableHeight();
      },
      deep: true,
      immediate: true,
    },
    "clusterStore.clusterAlert": {
      handler() {
        this.setTableHeight();
      },
      deep: true,
      immediate: true,
    },
  },
  unmounted() {
    if (this.sticky) {
      removeEventListener("resize", this.setTableHeight);
    }
  },
});
</script>
