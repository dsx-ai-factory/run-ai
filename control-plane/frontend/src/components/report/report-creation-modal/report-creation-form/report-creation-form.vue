<template>
  <runai-form-wrapper :form-state="formData" ref="reportForm" full-width>
    <report-name-section v-model:report-name="formData.name" v-model:description="formData.description" />
    <report-collection-period-section v-model:start-date="formData.start" v-model:end-date="formData.end" />
    <report-filters-section v-model:group-by="formData.groupBy" v-model:filter-by="formData.filterBy" />
  </runai-form-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//store
//svc

//cmps

//model
import type { ReportRequestFields } from "@/swagger-models/org-unit-service-client";
import ReportNameSection from "@/components/report/report-creation-modal/report-creation-form/report-name-section/report-name-section.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import ReportCollectionPeriodSection from "@/components/report/report-creation-modal/report-creation-form/report-collection-period-section/report-collection-period-section.vue";
import { deepCopy } from "@/utils/common.util";
import ReportFiltersSection from "@/components/report/report-creation-modal/report-creation-form/report-filters-section/report-filters-section.vue";

export default defineComponent({
  name: "report-creation-form",
  components: { ReportFiltersSection, ReportCollectionPeriodSection, RunaiFormWrapper, ReportNameSection },
  props: {
    report: {
      type: Object as PropType<ReportRequestFields>,
      required: true,
    },
  },
  data: function () {
    return {
      formData: deepCopy(this.report),
    };
  },
  watch: {
    formData: {
      handler() {
        this.$emit("update:report", this.formData);
      },
      deep: true,
    },
  },
});
</script>

<style scoped lang="scss"></style>
