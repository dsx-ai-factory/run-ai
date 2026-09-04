<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-item v-if="!isWhiteLabel" clickable aid="guided-steps-menu" class="guided-steps-trigger">
    <q-item-section>
      <q-item-label class="text-b-subtitle2 guided-steps-trigger-label">Guided steps</q-item-label>
    </q-item-section>
    <q-item-section side>
      <q-icon name="ph-bold ph-caret-right" size="16px" class="guided-steps-chevron" />
    </q-item-section>

    <q-menu class="guided-steps-submenu" anchor="top start" self="top end">
      <q-list>
        <q-item
          v-for="option in visibleTourOptions"
          :key="option.id"
          :disable="option.disabled"
          clickable
          v-close-popup
          :aid="option.aid"
          @click="option.action"
        >
          <q-item-section>
            <q-item-label
              :class="['text-b-subtitle2 guided-step-title', { 'guided-step-title--disabled': option.disabled }]"
              :aid="option.labelAid"
            >
              {{ option.label }}
            </q-item-label>
            <q-item-label
              :class="['text-b-body1 guided-step-description', { 'guided-step-description--disabled': option.disabled }]"
            >
              {{ option.tooltip }}
            </q-item-label>
          </q-item-section>
          <q-tooltip v-if="option.disabledTooltip && option.disabled" class="guided-step-tooltip" max-width="300px">
            {{ option.disabledTooltip }}
          </q-tooltip>
        </q-item>
      </q-list>
    </q-menu>
  </q-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { DASHBOARD_ROUTE_NAMES } from "@/router/dashboard.routes/dashboard.routes.names";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

export interface TourOption {
  id: string;
  label: string;
  tooltip: string;
  aid: string;
  labelAid: string;
  visible: boolean;
  disabled: boolean;
  disabledTooltip?: string;
  action: () => void | Promise<void>;
}

export default defineComponent({
  name: "guided-steps-menu",
  data() {
    return {
      guidedTourStore: useGuidedTourStore(),
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      permissionStore: usePermissionStore(),
    };
  },
  computed: {
    isWhiteLabel(): boolean {
      return this.appStore.isWhiteLabel;
    },
    tourOptions(): TourOption[] {
      return [
        {
          id: "jupyter-notebook",
          label: "Jupyter Notebook",
          tooltip: "Learn how to submit your first Jupyter notebook",
          aid: "jupyter-notebook-guided-steps-option",
          labelAid: "guided-steps-option-jupyter-notebook",
          visible: true,
          disabled: !this.guidedTourStore.workloadTourAllowed,
          disabledTooltip: `This feature requires certain missing prerequisites, which are described in the Jupyter Notebook guide.`,
          action: this.openWorkloadsGuidedTour,
        },
        {
          id: "project-creation",
          label: "Project creation",
          tooltip: "Learn how to create a project and set its resources",
          aid: "project-creation-guided-steps-option",
          labelAid: "guided-steps-option-project-creation",
          visible: this.permissionStore.hasPermission(ResourceType.Project, Action.Create),
          disabled: false,
          action: this.openProjectCreationTour,
        },
        {
          id: "department-creation",
          label: "Department creation",
          tooltip: "Learn how to create a department and set its resources",
          aid: "department-creation-guided-steps-option",
          labelAid: "guided-steps-option-department-creation",
          visible: this.permissionStore.hasPermission(ResourceType.Department, Action.Create),
          disabled: false,
          action: this.openDepartmentCreationTour,
        },
        {
          id: "overview-dashboard",
          label: "Overview dashboard",
          tooltip: "Learn how to get the most out of the Overview dashboard",
          aid: "overview-guided-steps-option",
          labelAid: "guided-steps-option-overview-dashboard",
          visible: this.permissionStore.hasOverviewPermission(),
          disabled: false,
          action: this.openOverviewGuidedTour,
        },
      ];
    },
    visibleTourOptions(): TourOption[] {
      return this.tourOptions.filter((option) => option.visible);
    },
  },
  methods: {
    async openWorkloadsGuidedTour(): Promise<void> {
      if (!this.guidedTourStore.workloadTourAllowed) return;

      if (this.$route.name !== WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX) {
        await this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
        await this.$nextTick();
      }

      this.guidedTourStore.openWorkloadsGuidedTour();
    },
    async openOverviewGuidedTour(): Promise<void> {
      if (this.$route.name !== DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX) {
        await this.$router.push({ name: DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX });
        await this.$nextTick();
      }

      this.guidedTourStore.openOverviewGuidedTour();
    },
    openProjectCreationTour(): void {
      this.$router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
        query: {
          startTour: "true",
          clusterId: this.clusterStore.selectedClusterId,
          t: Date.now().toString(),
        },
      });
    },
    openDepartmentCreationTour(): void {
      this.$router.push({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_NEW,
        query: {
          startTour: "true",
          clusterId: this.clusterStore.selectedClusterId,
          t: Date.now().toString(),
        },
      });
    },
  },
});
</script>

<style lang="scss">
.guided-steps-trigger {
  .guided-steps-trigger-label {
    line-height: 19.5px !important; // overrides Quasar's .q-item__label { line-height: 1.2em !important }
  }

  .guided-steps-chevron {
    color: var(--icon-interactive);
  }
}

.guided-steps-submenu.q-menu {
  background: var(--bg-primary-default);
  box-shadow: var(--shadow-s);
  border: none;
  border-radius: var(--space-2xs);
  width: 340px;

  .q-item {
    padding: var(--space-m);
  }

  .guided-step-title {
    line-height: 19.5px !important; // overrides Quasar's .q-item__label { line-height: 1.2em !important }
  }

  .guided-step-title--disabled {
    color: var(--text-disabled);
  }

  .guided-step-description {
    line-height: 16.8px !important; // overrides Quasar's .q-item__label { line-height: 1.2em !important }
    color: var(--b-text-sub);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .guided-step-description--disabled {
    color: var(--text-disabled);
  }
}

.guided-step-tooltip {
  background: var(--bg-primary-default);
  box-shadow: var(--shadow-s);
  color: var(--b-text-main);
  font-size: 12px;
  font-weight: 400;
  line-height: 16.8px;
  letter-spacing: 0;
  padding: var(--space-s);
}
</style>
