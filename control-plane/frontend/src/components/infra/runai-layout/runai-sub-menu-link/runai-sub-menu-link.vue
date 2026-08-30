<template>
  <router-link :to="{ name: route.name }" custom v-slot="{ href, navigate, isExactActive }">
    <a
      class="side-nav-menu-item-link-anchor"
      role="menuitem"
      :aria-current="isExactActive ? 'page' : undefined"
      :aid="route.aid"
      :href="href"
      @click="navigate"
    >
      <q-item
        role="none"
        class="side-nav-menu-item q-px-sm"
        clickable
        v-ripple
        :class="{ 'is-active-link': isExactActive }"
      >
        <div class="menu-label-container">
          <div class="menu-label-main">{{ route.label }}</div>
          <div v-if="route.subtitle" class="menu-label-subtitle">{{ route.subtitle }}</div>
        </div>

        <runai-stage-badge class="q-ml-md" v-if="route.stageBadge" :badge="route.stageBadge" />
      </q-item>
    </a>
  </router-link>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { type ISubjectRoutes } from "../routes-model/routes-model";
import { RunaiStageBadge } from "@/components/common/runai-stage-badge";
export default defineComponent({
  name: "side-nav-menu-item",
  components: {
    RunaiStageBadge,
  },
  props: {
    route: {
      type: Object as PropType<ISubjectRoutes>,
      required: true,
    },
  },
});
</script>
<style lang="scss">
.side-nav-menu-item-link-anchor {
  width: 100%;

  &:hover {
    text-decoration: none;
    background-color: $light-navy-0;
  }

  .side-nav-menu-item {
    border-radius: 5px;
    min-height: 36px;
    color: $main-font-color;
    padding-left: 10px;
    padding-right: 10px;

    &.is-active-link {
      @include m-bg-secondary-opacity-clr;
    }

    .menu-label-container {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .menu-label-main {
      font-size: 14px;
      line-height: 1.2;
    }

    .menu-label-subtitle {
      font-size: 12px;
      color: rgba(0, 0, 0, 0.54);
      font-style: italic;
      line-height: 1;
    }
  }
}
</style>
