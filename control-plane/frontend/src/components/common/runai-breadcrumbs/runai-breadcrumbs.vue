<template>
  <section class="breadcrumbs" aid="page-title">
    <div class="row items-center justify-start q-gutter-x-sm">
      <template v-if="fullPath">
        <template v-for="(item, index) in items" :key="item">
          <div
            class="breadcrumbs-item row items-center"
            :class="index === items.length - 1 ? 'breadcrumbs-last text-b-h6-bold' : 'text-b-h6'"
          >
            <template v-if="index === items.length - 1 && hasSubtitleInItem(item, index)">
              <div class="breadcrumb-with-subtitle">
                <span class="breadcrumb-main">{{ getMainTitle(item) }}</span>
                <span class="breadcrumb-subtitle text-b-body1">{{ getSubtitle() }}</span>
              </div>
            </template>
            <template v-else>
              {{ item }}
            </template>
          </div>
          <div v-if="index < items.length - 1" class="breadcrumbs-separator row items-center text-b-h6">
            <i class="ph-bold ph-caret-right"></i>
          </div>
        </template>
      </template>

      <template v-else>
        <div class="breadcrumbs-item row items-center breadcrumbs-last text-b-h6-bold">
          <template v-if="hasSubtitle">
            <div class="breadcrumb-with-subtitle">
              <span class="breadcrumb-main">{{ mainTitle }}</span>
              <span class="breadcrumb-subtitle text-b-body1">{{ subtitle }}</span>
            </div>
          </template>
          <template v-else>
            {{ items[items.length - 1] }}
          </template>
        </div>
      </template>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  components: {},
  props: {
    items: {
      type: Array as PropType<string[]>,
      required: true,
    },
    fullPath: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    subtitle: {
      type: String as PropType<string>,
      default: undefined,
    },
  },
  computed: {
    lastItem(): string {
      return this.items.at(-1) || "";
    },
    hasSubtitle(): boolean {
      return !!this.subtitle;
    },
    mainTitle(): string {
      return this.lastItem;
    },
  },
  methods: {
    hasSubtitleInItem(item: string, index: number): boolean {
      return index === this.items.length - 1 && this.hasSubtitle;
    },
    getMainTitle(item: string): string {
      return item;
    },
    getSubtitle(): string {
      return this.subtitle || "";
    },
  },
});
</script>
<style lang="scss" scoped>
.breadcrumbs {
  .breadcrumbs-item {
    color: var(--nav-primary-breadcrumbs);

    &:first-child {
      margin-left: var(--space-m);
    }
  }

  .breadcrumbs-separator {
    color: var(--nav-primary-breadcrumbs);
  }

  .breadcrumbs-last {
    color: var(--b-text-main);

    .breadcrumb-with-subtitle {
      display: flex;
      align-items: baseline;
      gap: 6px;

      .breadcrumb-main {
        color: var(--b-text-main);
      }

      .breadcrumb-subtitle {
        color: var(--b-text-sub);
        font-style: italic;
      }
    }
  }
}
</style>
