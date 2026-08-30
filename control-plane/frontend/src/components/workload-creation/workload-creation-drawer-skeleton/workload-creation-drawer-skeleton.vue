<template>
  <overlay-drawer-section-wrapper @close="$emit('close')">
    <template #header-title>
      <div>
        <q-skeleton width="260px" height="30px" />
      </div>
    </template>

    <template #header-main>
      <q-item class="full-width q-pa-md row items-center no-wrap">
        <q-skeleton width="240px" height="40px" class="q-mr-lg" />
        <q-skeleton size="36px" class="q-mr-xs" />
        <q-skeleton size="36px" />
      </q-item>
    </template>

    <template #main>
      <q-list>
        <q-item-section class="text-center q-pa-md">
          <q-skeleton v-for="(_, index) in 3" width="340px" height="50px" class="q-mb-md" :key="index" />
        </q-item-section>
      </q-list>
    </template>
  </overlay-drawer-section-wrapper>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { OverlayDrawerSectionWrapper } from "@/components/workload-creation/overlay-drawer-section-wrapper";

export default defineComponent({
  name: "workload-creation-drawer-skeleton",
  components: { OverlayDrawerSectionWrapper },
});
</script>
