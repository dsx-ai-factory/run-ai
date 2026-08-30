<template>
  <div class="notification-alert-bar text-b-body1" :class="[containerClasses, orientationClass]">
    <div class="notification-alert-bar__main">
      <q-icon class="variant-icon" :class="iconClasses" :name="icon" size="20px" />

      <div class="notification-alert-bar__content">
        <span v-if="showH1Content && title" class="notification-alert-bar__title text-b-body1-md">
          {{ title }}
        </span>

        <span class="content">
          <slot>
            <runai-html-wrapper v-if="message" tag="span" :html="message" />
          </slot> </span
        >{{ " " }}<slot v-if="$slots['inline-actions']" name="inline-actions" />
      </div>

      <q-icon
        v-if="showCloseAction"
        name="ph-bold ph-x"
        size="20px"
        class="notification-alert-bar__close"
        role="button"
        tabindex="0"
        @click="$emit('close')"
      />
    </div>

    <div
      v-if="ctAs && ctAs.length > 0 && variant === ENotificationBarVariant.Warning"
      class="notification-alert-bar__ctas"
    >
      <q-btn
        v-for="(cta, index) in ctAs"
        :key="index"
        flat
        dense
        class="notification-alert-bar__cta-btn"
        :label="cta.label"
        @click="cta.onClick"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiHtmlWrapper } from "@/components/common/runai-html-wrapper";
import { ENotificationBarVariant, variantConfig } from "./notification-alert-bar.model";
import type { INotificationBarCta } from "./notification-alert-bar.model";

defineOptions({ name: "NotificationAlertBar" });

const props = withDefaults(
  defineProps<{
    variant?: ENotificationBarVariant;
    message?: string;
    textAlign?: "left" | "center" | "right";
    showCloseAction?: boolean;
    showH1Content?: boolean;
    title?: string;
    orientation?: "horizontal" | "vertical";
    ctAs?: INotificationBarCta[];
  }>(),
  {
    variant: ENotificationBarVariant.Info,
    message: undefined,
    textAlign: "left",
    showCloseAction: false,
    showH1Content: false,
    title: undefined,
    orientation: "horizontal",
    ctAs: undefined,
  },
);

defineEmits<{
  close: [];
}>();

const config = computed(() => variantConfig[props.variant] ?? variantConfig[ENotificationBarVariant.Info]);

const containerClasses = computed(() => config.value.background);
const icon = computed(() => config.value.icon);
const iconClasses = computed(() => config.value.iconColor);
const orientationClass = computed(() =>
  props.orientation === "vertical" ? "notification-alert-bar--vertical" : "notification-alert-bar--horizontal",
);
</script>

<style lang="scss" scoped>
.notification-alert-bar {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 50px;
  border-radius: 3px;
  padding: var(--space-s) var(--space-m);
  color: var(--b-text-main);
  animation: fade-in 0.3s ease-in-out forwards;

  &__main {
    display: flex;
    align-items: center;
    gap: var(--space-m);
    text-align: v-bind(textAlign);
  }

  &__content {
    flex: 1;
    min-width: 0;
  }

  &__title {
    display: block;
  }

  &__close {
    margin-left: auto;
    flex-shrink: 0;
    color: var(--icon-default);
    cursor: pointer;
  }

  &__ctas {
    display: flex;
    gap: var(--space-m);
    margin-top: var(--space-m);
    padding-left: calc(17px + var(--space-m));
  }

  &__cta-btn {
    color: var(--interactive-secondary-default-text);
    text-transform: uppercase;
    font-weight: 500;
    font-size: 14px;
  }

  &--horizontal {
    .notification-alert-bar__title {
      display: inline;
      margin-right: var(--space-2xs);
    }

    .notification-alert-bar__content {
      display: inline;
    }
  }

  &--vertical {
    .notification-alert-bar__title {
      display: block;
      margin-bottom: var(--space-2xs);
    }
  }
}

// --- Variant backgrounds ---
.variant-info {
  background-color: var(--bg-primary-info);
}

.variant-warning {
  background-color: var(--bg-primary-warning);
}

.variant-error {
  background-color: var(--bg-primary-error);
}

.variant-success {
  background-color: var(--bg-primary-success);
}

.variant-default {
  background-color: var(--bg-secondary-default);
}

// --- Variant icon colors ---
.variant-icon-info {
  color: var(--icon-info);
}

.variant-icon-warning {
  color: var(--text-warning);
}

.variant-icon-error {
  color: var(--text-error);
}

.variant-icon-success {
  color: var(--text-success);
}

.variant-icon-default {
  color: var(--icon-default);
}

@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
