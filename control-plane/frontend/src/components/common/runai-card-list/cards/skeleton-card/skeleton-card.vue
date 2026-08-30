<template>
  <q-card class="skelaton-card">
    <q-card-section class="workload-card-content">
      <q-skeleton size="50px" class="q-mb-md" />
      <q-skeleton height="100px" />
    </q-card-section>
  </q-card>
</template>

<style lang="scss" scoped>
.skelaton-card {
  @include runaiSelectCardSize;

  .q-card__section--vert {
    padding: 0;
  }
}
</style>
