<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    class="report-creation-modal"
    model-value
    @close="$emit('close')"
    size="width-xl"
    no-padding
    dialog-body-size="extra-large"
  >
    <template #header>Consumption Report </template>
    <template #body>
      <q-card>
        <q-card-section>
          <report-creation-form v-model:report="formData" ref="reportFormWrapper" />
        </q-card-section>
      </q-card>
    </template>
    <template #footer>
      <runai-form-footer
        class="q-mt-md"
        save-btn-label="create report"
        @cancel="$emit('close')"
        @save="createReport"
        :loading="loading"
        :is-form-valid="isFormValid"
      />
    </template>
  </runai-base-dialog>
</template>
<script lang="ts">
import { defineComponent } from "vue";
//svc
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
//utils
import { ErrorAlert } from "@/utils/error-alert.util";
import { alertUtil } from "@/utils/alert.util";
//cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog/";
//model
import { ReportGroupByOptions, type ReportRequestFields } from "@/swagger-models/org-unit-service-client";

export default defineComponent({
  name: "report-creation-modal",
  components: { RunaiBaseDialog },
  emits: ["report-created", "close"],
  data: function () {
    return {
      loading: false as boolean,
      isFormValid: true as boolean,
      formData: {
        name: "",
        start: new Date(new Date().setDate(new Date().getDate() - 30)).toISOString(),
        end: new Date().toISOString(),
        groupBy: ReportGroupByOptions.Nodepool,
        filterBy: [],
        description: "",
      } as ReportRequestFields,
    };
  },
  methods: {
    async validate(): Promise<boolean> {
      return await ((this.$refs.reportFormWrapper as HTMLFormElement).$refs.reportForm as HTMLFormElement).validate();
    },
    async createReport(): Promise<void> {
      const isValid = await this.validate();
      this.isFormValid = isValid;
      if (!isValid) return;
      this.loading = true;
      try {
        const createdReport = await orgUnitService.createReport(this.formData);
        this.$emit("report-created", createdReport);
        this.$q.notify(
          alertUtil.getSuccess(
            `Report ${createdReport.name} will be available for download once its status change to "Ready"`,
          ),
        );
      } catch (e: unknown) {
        this.handleError(e);
      } finally {
        this.loading = false;
      }
    },
    handleError(error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedCreateMessage("report"),
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
  },
});
</script>

<script setup lang="ts">
import ReportCreationForm from "@/components/report/report-creation-modal/report-creation-form/report-creation-form.vue";
import RunaiFormFooter from "@/components/common/runai-form-footer/runai-form-footer.vue";
</script>
<style scoped lang="scss"></style>
