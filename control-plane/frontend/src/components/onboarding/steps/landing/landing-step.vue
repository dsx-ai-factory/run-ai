<template>
  <onboarding-content-wrapper>
    <template #title> Welcome, {{ userName }}! </template>

    <template #subtitle>
      <p>Let's install NVIDIA Run:ai and set up your first research team so they can run workloads.</p>
    </template>

    <template #content>
      <div class="illustration-section">
        <div class="gpu-illustration">
          <runai-svg-icon name="onboarding-hello" class="gpu-icon" />
        </div>
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  name: "landing-step",
  components: {
    RunaiSvgIcon,
    OnboardingContentWrapper,
  },
  setup() {
    const userName = computed(() => useAuthStore().currentUser.username);

    return {
      userName,
    };
  },
});
</script>

<style scoped lang="scss">
.illustration-section {
  display: flex;
  justify-content: center;
  margin-top: 150px;
  .gpu-illustration {
    .gpu-icon {
      width: 200px;
      height: 200px;
      opacity: 0.9;
    }
  }
}
</style>
