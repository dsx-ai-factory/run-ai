<template>
  <div class="runai-table-no-projects">
    <span v-if="hasEntityCreatePermission && hasProjectCreatePermission">
      First <a href="#" class="link" @click.prevent="createProject">create at least one project</a> so
      {{ entityType }} can be created.
    </span>
    <div class="text-center" v-else-if="hasEntityCreatePermission && !hasProjectCreatePermission">
      <div>{{ entityTextPluralCapitalized }} can only be created under a project.</div>
      <div>Contact your administrator to add a project.</div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// stores
import { useAuthStore } from "@/stores/auth.store";
import type { PropType } from "vue";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { capitalizeString } from "@/utils/string.util/string.util";
export default defineComponent({
  props: {
    entityType: {
      type: String as PropType<string>,
      required: true,
    },
    hasEntityCreatePermission: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    hasProjectCreatePermission: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
    };
  },
  computed: {
    entityTextPluralCapitalized(): string {
      return capitalizeString(this.entityType + "s");
    },
  },
  methods: {
    createProject(): void {
      this.$router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
        query: { clusterId: this.clusterId },
      });
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-table-no-projects {
  .link {
    @include m-link-clr;
    cursor: pointer;
    &:hover {
      text-decoration: underline;
      cursor: pointer;
    }
  }
}
</style>
