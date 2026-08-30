<template>
  <q-popup-proxy
    class="runai-small-popover"
    @hide="$emit('hide')"
    @show="$emit('show')"
    ref="popover"
    :cover="cover"
    :breakpoint="breakpoint"
  >
    <q-card class="popover-content-container" :class="[customClass, size]" :style="width">
      <q-card-section class="runai-small-popover-header row justify-between items-center">
        <h6 v-if="showSeparator">
          <b><slot name="header"></slot></b>
        </h6>
        <div class="text-black-54" v-else style="font-size: 14px; font-weight: 400">
          <slot name="header"></slot>
        </div>
        <q-btn
          class="runai-small-popover-close-btn runai-icon-btn--default"
          icon="ph-bold ph-x"
          flat
          round
          @click="hide"
        />
      </q-card-section>

      <q-separator v-if="showSeparator" />

      <q-card-section class="runai-small-popover-body" :class="[noPadding ? 'no-padding' : '', scroll ? 'scroll' : '']">
        <slot name="body"></slot>
      </q-card-section>

      <q-separator v-if="showSeparator" />

      <q-card-actions class="runai-small-popover-footer row">
        <div class="footer-left">
          <slot name="footer-left"></slot>
        </div>
        <div class="footer">
          <slot name="footer"></slot>
        </div>
      </q-card-actions>
    </q-card>
  </q-popup-proxy>
</template>

<script lang="ts">
import type { QPopupProxy } from "quasar";
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  emits: ["hide", "show"],
  props: {
    showSeparator: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    customClass: {
      type: String as PropType<string>,
      required: false,
    },
    size: {
      type: String as PropType<"width-xs" | "width-sm" | "width-md" | "width-lg">,
      required: false,
      default: "",
    },
    customWidth: {
      type: String as PropType<string>,
      required: false,
    },
    noPadding: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    scroll: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    cover: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    breakpoint: {
      type: Number as PropType<number>,
      default: 800,
    },
  },
  computed: {
    width(): { width: string } | string {
      return this.customWidth ? { width: this.customWidth } : "";
    },
  },
  methods: {
    hide() {
      const popover = this.$refs.popover as QPopupProxy;
      popover.hide();
      this.$emit("hide");
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-small-popover {
  .popover-content-container {
    &.width-xs {
      width: 300px;
    }

    &.width-sm {
      width: 540px;
    }

    &.width-md {
      width: 700px;
      max-width: 80vw;
    }

    &.width-lg {
      width: 1000px;
      max-width: 80vw;
    }
  }

  .runai-small-popover-body {
    &.scroll {
      max-height: 50vh;
    }
  }

  .runai-small-popover-footer {
    .footer-left {
      &:empty {
        display: none;
      }
    }
    .footer {
      display: flex;
      justify-content: flex-end;
      flex: 1;
    }
  }
}
</style>

<style lang="scss">
.runai-small-popover {
  .runai-small-popover-footer .q-btn + .q-btn {
    margin-left: 8px;
  }
}
</style>
