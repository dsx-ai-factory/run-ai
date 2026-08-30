<template>
  <div class="runai-sub-menu">
    <template v-if="headerText">
      <header class="sub-menu-header text-subtitle1 text-bold">
        <template v-if="hasSubtitle">
          <div>{{ mainHeaderText }}</div>
          <div class="header-subtitle">{{ headerSubtitle }}</div>
        </template>
        <template v-else>
          {{ headerText }}
        </template>
      </header>
      <q-separator class="title-separator" />
    </template>
    <div class="runai-sub-menu-items" role="menu">
      <template v-for="item in items" :key="item.name">
        <template v-if="item.subItems">
          <q-item
            class="sub-items-title"
            clickable
            role="menuitem"
            :aria-expanded="subItemOpen"
            @click="subItemOpen = !subItemOpen"
            :aid="item.aid"
          >
            <span>{{ item.label }}</span>
            <q-space />
            <runai-svg-icon class="sub-items-icon" size="12" :class="{ open: !subItemOpen }" name="sort-down" />
          </q-item>
          <div class="sub-items-section row gap-5 q-px-md" role="group" :class="{ open: subItemOpen }">
            <template v-for="subItem in item.subItems">
              <runai-sub-menu-link v-if="!subItem.subItems" :key="subItem.name" :route="subItem" />
              <template v-else>
                <div :key="subItem.name" class="column gap-5">
                  <q-item
                    class="sub-items-title"
                    clickable
                    role="menuitem"
                    :aria-expanded="subSubItemOpen"
                    @click="subSubItemOpen = !subSubItemOpen"
                    :aid="subItem.aid"
                  >
                    <span>{{ subItem.label }}</span>
                    <q-space />
                    <runai-svg-icon
                      class="sub-items-icon"
                      size="12"
                      :class="{ open: !subSubItemOpen }"
                      name="sort-down"
                    />
                  </q-item>
                  <div class="sub-items-section row gap-5 q-px-md" role="group" :class="{ open: subSubItemOpen }">
                    <template v-for="subSubItem in subItem.subItems">
                      <runai-sub-menu-link v-if="!subSubItem.subItems" :key="subSubItem.name" :route="subSubItem" />
                    </template>
                  </div>
                </div>
              </template>
            </template>
          </div>
        </template>
        <runai-sub-menu-link v-else :route="item" />
      </template>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiSubMenuLink } from "@/components/infra/runai-layout/runai-sub-menu-link";

import type { ISubjectRoutes } from "../routes-model/routes-model";

export default defineComponent({
  name: "runai-sub-menu",
  components: {
    RunaiSubMenuLink,
    RunaiSvgIcon,
  },
  props: {
    headerText: {
      type: String as PropType<string>,
      required: false,
    },
    items: {
      type: Array as PropType<ISubjectRoutes[]>,
      required: true,
    },
  },
  data() {
    return {
      subItemOpen: true,
      subSubItemOpen: true,
    };
  },
  computed: {
    hasSubtitle(): boolean {
      return this.headerText?.includes("\n") || false;
    },
    mainHeaderText(): string {
      if (!this.hasSubtitle || !this.headerText) return this.headerText || "";
      return this.headerText.split("\n")[0];
    },
    headerSubtitle(): string {
      if (!this.hasSubtitle || !this.headerText) return "";
      return this.headerText.split("\n")[1];
    },
  },
});
</script>
<style scoped lang="scss">
.runai-sub-menu {
  width: 215px;
  background-color: $light-grey;
  display: flex;
  flex-direction: column;
  padding: 10px 0;

  .sub-menu-header {
    padding-left: 20px;
    padding-right: 20px;

    .header-subtitle {
      font-size: 12px;
      font-weight: 400;
      color: rgba(0, 0, 0, 0.54);
      font-style: italic;
      margin-top: 2px;
    }
  }

  .title-separator {
    margin: 10px 0;
  }

  .sub-items-section {
    overflow: hidden;
    max-height: 0px;
    transition: all 0.3s ease-in-out;

    &.open {
      max-height: 500px;
    }
  }

  .sub-items-title {
    min-height: 36px;
    color: $main-font-color;
    border-radius: 5px;
    display: flex;
    align-items: center;
    padding: 8px 10px;

    .sub-items-icon {
      transform: rotate(0deg);
      transition: all 0.3s ease-in-out;

      &.open {
        transform: rotate(-180deg);
      }
    }
  }

  .runai-sub-menu-items {
    padding-right: 10px;
    padding-left: 10px;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
}
</style>
