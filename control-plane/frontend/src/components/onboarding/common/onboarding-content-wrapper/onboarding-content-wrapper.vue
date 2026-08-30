<template>
  <div class="onboarding-content-wrapper" :class="{ 'center-aligned': centerAlign }">
    <div class="content-header">
      <div class="content-title">
        <slot name="title" />
      </div>
      <div class="content-subtitle" v-if="$slots.subtitle">
        <slot name="subtitle" />
      </div>
    </div>

    <div class="content-body">
      <slot name="content" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "onboarding-content-wrapper",
  props: {
    centerAlign: {
      type: Boolean,
      default: false,
    },
  },
});
</script>

<style lang="scss">
.onboarding-content-wrapper {
  width: 100%;
  height: 100%;
  margin: 0 auto;
  text-align: left;
  animation: fadeInContent 0.3s ease-out;
  display: flex;
  flex-direction: column;
  justify-content: center;

  &.center-aligned {
    text-align: center;
    align-items: center;

    .content-header {
      text-align: center;
      display: flex;
      flex-direction: column;
      align-items: center;
      text-wrap: nowrap;
    }

    .content-body {
      display: flex;
      justify-content: center;
      align-items: center;
      margin-right: 140px;
    }
  }

  .content-header {
    text-align: left;
    color: $black-87;
    .content-title {
      font-weight: 500;
      font-size: 24px;
      line-height: 30px;
      animation: slideInTitle 0.3s ease-out 0.1s both;
      margin-bottom: 20px;
    }

    .content-subtitle {
      font-weight: 400;
      font-size: 16px;
      line-height: 26px;
      animation: slideInSubtitle 0.3s ease-out 0.15s both;
      margin-bottom: 20px;
    }
  }

  .content-body {
    padding-top: 40px;
    animation: slideInBody 0.3s ease-out 0.2s both;
  }

  // Content animations
  @keyframes fadeInContent {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes slideInTitle {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes slideInSubtitle {
    from {
      opacity: 0;
      transform: translateY(-8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes slideInBody {
    from {
      opacity: 0;
      transform: translateY(-6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  // Onboarding select cards styles (used in protocol selection, cluster location, etc.)
  .onboarding-select-options {
    display: flex;
    gap: 39px;
  }

  .onboarding-select-card {
    width: 239px;
    height: 201px;

    .card-content {
      padding: 9px 23px;
      height: 100%;
      display: flex;
      flex-direction: column;
    }
  }

  .option-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 64px;
    height: 64px;
    margin-bottom: 16px;

    .runai-select-card .card-content.is-checked & {
      color: var(--primary-color);
    }
  }

  .option-content {
    flex: 1;

    .option-title {
      font-weight: 700;
      font-size: 16px;
      line-height: 100%;
      margin-bottom: 13px;
    }

    .option-description {
      font-weight: 400;
      font-size: 12px;
      line-height: 16px;
    }
  }

  .onboarding-select-error {
    margin-top: 30px;
    padding-left: 10px;
  }
}
</style>
