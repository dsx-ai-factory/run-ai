<template>
  <section class="notification-alert-container">
    <notification-alert-bar :variant="ENotificationBarVariant.Info">
      <template #default>
        <strong>You have access rules that use legacy roles.</strong> For continued compatibility and better permission
        control,
      </template>
      <template #inline-actions>
        <q-btn
          @click.stop="emit('transition-click')"
          class="btn-link no-padding transition-link link-clr"
          :ripple="false"
          flat
          color="primary"
          label="transition to new roles"
          no-caps
        />
        <span>. For more information, see the </span>
        <a target="_blank" rel="noopener noreferrer" :href="DOCS.AUTHENTICATION.ACCESS_RULES">Access rules guide</a>
        <span>.</span>
      </template>
    </notification-alert-bar>
  </section>
</template>

<script setup lang="ts">
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { DOCS } from "@/constants/docs";

defineOptions({ name: "legacy-access-rule-banner" });

const emit = defineEmits<{
  "transition-click": [];
}>();
</script>

<style lang="scss" scoped>
.transition-link {
  font: inherit;
}
</style>
