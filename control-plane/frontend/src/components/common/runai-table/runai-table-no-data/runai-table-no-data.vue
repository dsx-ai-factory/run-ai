<template>
  <div :class="getClass()">
    <div class="col justify-center">
      <template v-if="showError">
        <section class="table-error-section text-center">
          <i class="ph-bold ph-warning text-body-color" style="font-size: 1.5em"></i>
          For some reason this page can't be displayed right now. Try refreshing the page or contact your admin.
        </section>
      </template>

      <template v-else>
        <template v-if="!isFiltered">
          <div class="row justify-center q-my-md empty-results-icon" v-if="iconName || phIconName">
            <q-icon v-if="phIconName" :name="phIconName" :size="iconSize + 'px'" class="empty-state-ph-icon" />
            <runai-svg-icon v-else :name="iconName || ''" :size="iconSize" class="" />
          </div>
          <div class="row justify-center empty-state-message text-b-subtitle2">{{ message }}</div>
          <div class="row justify-center empty-state-message text-b-subtitle2">
            <slot name="secondary-text"></slot>
          </div>
          <div class="row justify-center q-mt-md" v-if="showCreateBtn">
            <slot name="primary-btn" :disabled="disableCreateBtn">
              <q-btn
                v-permission="{ resourceType, action: Action.Create }"
                color="primary"
                flat
                class="btn-icon-sm runai-icon-btn--default"
                icon="ph-bold ph-plus"
                aid="empty-state-add-entity-btn"
                :label="'New ' + entityName"
                :disable="disableCreateBtn"
                @click="createNew"
              ></q-btn>
            </slot>
          </div>
        </template>
        <template v-else>
          <div class="row justify-center empty-state-message text-b-subtitle2">{{ noFilterResultsText }}</div>
          <section v-if="showFilterIconAndButton">
            <!-- lightIcon: minimal filter empty-state (plain bold funnel, no green circle, no clear-filters button) -->
            <div class="row justify-center q-my-md no-results-icon">
              <q-icon name="ph-bold ph-funnel" size="44px" :class="{ 'filter-no-results-icon': !lightIcon }" />
            </div>
            <div class="row justify-center" v-if="!lightIcon">
              <q-btn
                color="primary"
                flat
                class="btn-icon-sm"
                aid="empty-state-clear-filters-btn"
                label="Clear filters"
                @click="clearFilters"
              ></q-btn>
            </div>
          </section>
        </template>
      </template>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { CLUSTER_COLUMN_FILTER_NAME, type IFilterBy } from "@/models/filter.model";
import { Action, type ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  components: {
    RunaiSvgIcon,
  },
  emits: ["clear-filters", "create-new"],
  props: {
    entityName: {
      type: String as PropType<string>,
      required: true,
    },
    iconName: {
      type: String as PropType<string>,
      required: false,
    },
    phIconName: {
      type: String as PropType<string>,
      required: false,
    },
    customNoFilterResultsMessage: {
      type: String as PropType<string>,
      required: false,
    },
    lightIcon: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    iconSize: {
      type: String as PropType<string>,
      default: "83",
    },
    showCreateBtn: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disableCreateBtn: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
    showError: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    showFilterIconAndButton: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    customMessage: {
      type: String as PropType<string>,
      required: false,
    },
    ignoreClusterFilter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      message: "" as string,
    };
  },
  created() {
    this.message = this.customMessage ? this.customMessage : `You don't have any ${this.entityPlural} yet.`;
  },
  computed: {
    Action(): typeof Action {
      return Action;
    },
    resourceType(): ResourceType {
      return this.$route.meta.resourceType as ResourceType;
    },
    isFiltered(): boolean {
      const columnsFiltersLength: number = this.filterBy.columnFilters?.length || 0;
      if (columnsFiltersLength === 1) {
        const isClusterFilterExist = this.filterBy.columnFilters?.some(
          (filter) => filter.name === CLUSTER_COLUMN_FILTER_NAME,
        );
        if (isClusterFilterExist && this.ignoreClusterFilter) {
          return !!(this.filterBy.searchTerm || this.filterBy.advancedFilters?.length);
        }
      }
      return !!(
        this.filterBy.columnFilters?.length ||
        this.filterBy.searchTerm ||
        this.filterBy.advancedFilters?.length
      );
    },
    noFilterResultsText(): string {
      if (this.customNoFilterResultsMessage) {
        return this.customNoFilterResultsMessage;
      }
      return `No ${this.entityPlural.toLocaleLowerCase()} match your filters`;
    },
    entityPlural(): string {
      const lastCharacter: string = this.entityName.charAt(this.entityName.length - 1);
      switch (lastCharacter) {
        case "y":
          return this.entityName.slice(0, this.entityName.length - 1) + "ies";
        default:
          return this.entityName + "s";
      }
    },
  },
  methods: {
    clearFilters(): void {
      this.$emit("clear-filters");
    },
    createNew(): void {
      this.$emit("create-new");
    },
    getClass(): string {
      return `empty-state-container full-width row ${this.showError ? "q-py-sm" : "q-py-xl"}`;
    },
  },
});
</script>

<style lang="scss">
.empty-state-container {
  .table-error-section {
    i {
      margin-inline-end: var(--space-2xs);
    }
  }

  .empty-state-message {
    color: var(--b-text-main);
  }

  .empty-results-icon,
  .no-results-icon {
    svg {
      rect {
        fill: rgba(var(--secondary-color-rgb), var(--secondary-color-opacity)) !important;
      }

      path {
        fill: var(--secondary-color) !important;
      }
    }
  }

  .empty-state-ph-icon {
    color: var(--secondary-color);
    background: rgba(var(--secondary-color-rgb), var(--secondary-color-opacity));
    border-radius: 50%;
    padding: 16px;
  }

  .filter-no-results-icon {
    color: var(--secondary-color);
    background: rgba(var(--secondary-color-rgb), var(--secondary-color-opacity));
    border-radius: 50%;
    padding: 24px;
  }
}
</style>
