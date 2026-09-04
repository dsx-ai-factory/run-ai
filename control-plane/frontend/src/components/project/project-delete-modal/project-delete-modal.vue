<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="isModalOpen" @close="$emit('close')" size="width-md" no-padding :loading="loading">
    <template #header>Delete</template>

    <template #body>
      <q-card>
        <q-card-section class="text-wrapper">
          <div class="q-px-lg" v-if="!hasWorkloads">
            <span class="text-subtitle1">Permanently delete project: {{ projectName }}?</span>
            <p class="text-body2 q-py-xs">
              Once a project is deleted, it cannot be restored. <br />
              The associated namespace <span class="text-weight-bold">{{ namespace }}</span> will remain in the cluster.
              <br />
              To delete the namespace run the following command as the kube admin in the cluster:
            </p>
            <div class="row">
              <div class="col">
                <runai-copyable-code-block :code="displayedNamespaceText" />
              </div>
            </div>
          </div>

          <div class="q-px-lg" v-else>
            <span class="text-subtitle1"
              >Project {{ projectName }} has
              <q-btn flat color="primary" class="btn-link no-padding" @click="redirectToWorkloads"
                >{{ projectWorkloadsCount }} associated workloads.</q-btn
              ></span
            >
            <br />
            <span class="text-subtitle1">Permanently delete this project and its associated workloads and assets?</span>
            <q-separator class="q-my-md" />
            <p class="text-body2 q-py-xs">
              The associated namespace <span class="text-weight-bold">{{ namespace }}</span> will remain in the cluster.
              <br />
              To delete the namespace run the following command as the kube admin in the cluster:
            </p>
            <div class="row">
              <div class="col">
                <runai-copyable-code-block :code="displayedNamespaceText" />
              </div>
            </div>
          </div>
        </q-card-section>
      </q-card>
    </template>

    <template #footer>
      <q-btn label="Cancel" color="primary" flat @click="$emit('close')" aid="delete-modal-cancel-button" />
      <q-btn
        :loading="deleting"
        :disable="deleting || loading"
        label="Delete"
        color="primary"
        @click="handleDeleteProject"
        aid="delete-modal-delete-button"
      />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiCopyableCodeBlock } from "@/components/common/runai-copyable-code-block/";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { DeletionTypes, type Project } from "@/swagger-models/org-unit-service-client";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { copyToClipboard } from "quasar";
import { defineComponent } from "vue";
import type { PropType } from "vue";

export default defineComponent({
  components: { RunaiCopyableCodeBlock, RunaiBaseDialog },
  emits: ["close", "delete"],
  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
    deleting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      isModalOpen: true as boolean,
      projectWorkloadsCount: 0 as number,
      loading: true as boolean,
    };
  },
  created() {
    this.loadWorkloadsCount();
  },
  computed: {
    projectName(): string {
      return this.project.name;
    },
    namespace(): string {
      return this.project.status?.namespace || `runai-${this.projectName}`;
    },
    displayedNamespaceText() {
      return `kubectl delete ns ${this.namespace}`;
    },
    hasWorkloads(): boolean {
      return this.projectWorkloadsCount > 0;
    },
  },
  methods: {
    async loadWorkloadsCount(): Promise<void> {
      try {
        this.projectWorkloadsCount = await workloadService.getWorkloadsCount([
          filterUtil.getEqualsFilterString(WorkloadSortFilterFields.ProjectId, this.project.id),
        ]);
      } catch (error: unknown) {
        this.handleWorkloadsCountError(error);
      } finally {
        this.loading = false;
      }
    },
    handleWorkloadsCountError(error: unknown): void {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedDeleteMessage("project"),
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
    async copyText(text: string): Promise<void> {
      await copyToClipboard(text);
      this.$q.notify(alertUtil.getSuccess("Copied to clipboard"));
    },
    redirectToWorkloads(): void {
      const query = {
        clusterUuid: this.project.clusterId,
        advancedFilters: JSON.stringify([
          {
            term: this.project.name,
            value: EFilterOperator.Equals,
            label: "Project",
            name: WorkloadSortFilterFields.ProjectName,
            type: EColumnFilterType.FreeText,
          },
        ]),
      };
      this.$emit("close");
      this.$router.push({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
        query,
      });
    },
    handleDeleteProject(): void {
      this.$emit("delete", DeletionTypes.Cascade);
    },
  },
});
</script>
<style lang="scss">
.text-wrapper {
  padding: 10px 2px 1px;
}
</style>
