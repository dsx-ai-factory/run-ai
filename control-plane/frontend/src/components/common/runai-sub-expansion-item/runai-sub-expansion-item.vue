<template>
  <div role="list" class="runai-sub-expansion-list">
    <q-expansion-item
      :label="label"
      class="runai-sub-expansion-item"
      expand-icon-class="runai-expansion-icon"
      expand-icon="ph-bold ph-caret-down"
      expand-separator
      expand-icon-toggle
      v-model="isOpen"
      :header-class="headerClass"
    >
      <template #header>
        <div
          class="runai-sub-expansion-header runai-sub-expansion-header--label"
          role="button"
          tabindex="0"
          :aria-expanded="isOpen"
          @click="isOpen = !isOpen"
          @keydown.enter.prevent="isOpen = !isOpen"
          @keydown.space.prevent="isOpen = !isOpen"
          :aid="aid"
        >
          <span class="runai-sub-expansion-label">{{ label }}</span>
        </div>
        <div
          class="runai-sub-expansion-header runai-sub-expansion-header--validation row items-center"
          v-if="validationBarShow && validationBar && !isOpen"
          @click.stop
        >
          <q-icon name="ph-bold ph-warning-circle" color="red" size="20px" class="q-mr-xs" />
          <span class="validation-bar-inline-text">{{ validationBar }}</span>
        </div>
        <div class="information-bar-container">
          <notification-alert-bar class="flex-1" v-if="informationBar && !isOpen">
            <div class="information-bar-container-text">
              {{ informationBar }}
            </div>
          </notification-alert-bar>
        </div>
      </template>
      <slot />
    </q-expansion-item>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { NotificationAlertBar } from "@/components/common/notification-alert-bar";

export default defineComponent({
  components: {
    NotificationAlertBar,
  },
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    validationBarShow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    informationBar: {
      type: String as PropType<string>,
      required: false,
    },
    validationBar: {
      type: String as PropType<string>,
      required: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
    headerClass: {
      type: String as PropType<string>,
      default: "",
    },
  },
  data() {
    return {
      isOpen: !!this.defaultOpened,
    };
  },
});
</script>

<style lang="scss">
.runai-sub-expansion-item {
  width: 100%;

  .q-item {
    padding-left: 0;
  }

  hr {
    display: none;
  }
  .runai-expansion-icon {
    color: var(--icon-default, #707070);
    .q-icon {
      font-size: var(--icon-size-default) !important;
    }
  }

  .q-item.q-hoverable {
    margin-top: 20px;
    padding: 10px 0;

    &:hover {
      .q-focus-helper {
        opacity: 0;
      }
    }
  }

  .q-item > .q-item__section--side:last-of-type {
    order: 2;
  }

  .runai-sub-expansion-header {
    font-weight: 500;
    color: $primary;
    display: flex;
    align-items: center;
    padding-right: 20px;
    cursor: pointer;

    &--label {
      order: 1;
    }

    &--validation {
      order: 3;
      font-size: 12px;
      gap: 4px;
      cursor: default;

      .validation-bar-inline-text {
        color: $negative;
      }
    }
  }

  .information-bar-container {
    position: absolute;
    width: 580px;
    right: 5px;
    top: -4px;

    .information-bar-container-text {
      width: 430px;
    }
  }
}
</style>
