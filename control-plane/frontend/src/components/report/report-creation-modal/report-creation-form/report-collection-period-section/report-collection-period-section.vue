<template>
  <section aid="report-collection-period-section" class="q-pt-lg">
    <runai-section-title title="Set the report's data collection period" tooltip-text="Set a period of up to 30 days" />

    <div class="row items-center q-mt-lg">
      <!-- START DATE -->
      <runai-datepicker
        aid="custom-select-start"
        @changed="onStartDateChanged"
        label="Start date"
        :dense="false"
        :date-format="DATE_FORMATS.FULL_DATE_MONTH_NAME"
        :default-date="new Date(startDate)"
        :range-of-dates="(date: string | null) => rangeOfDatesCallback(date, true)"
      />
      <q-icon name="ph-bold ph-minus" class="q-px-md" />
      <!-- END DATE -->
      <runai-datepicker
        aid="custom-select-end"
        @changed="onEndDateChanged"
        label="End date"
        :dense="false"
        :default-date="new Date(endDate)"
        :date-format="DATE_FORMATS.FULL_DATE_MONTH_NAME"
        :range-of-dates="(date: string | null) => rangeOfDatesCallback(date, false)"
      />
    </div>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiDatepicker } from "@/components/common/runai-datepicker";
import RunaiSectionTitle from "@/components/common/runai-section-title/runai-section-title.vue";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
//cmps

//model

export default defineComponent({
  name: "report-collection-period-section",
  components: { RunaiSectionTitle, RunaiDatepicker },
  emits: ["update:start-date", "update:end-date"],
  props: {
    startDate: {
      type: String as PropType<string>,
      required: true,
    },
    endDate: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      DATE_FORMATS,
    };
  },
  methods: {
    onStartDateChanged(date: Date): void {
      const startOfDay = dateUtil.getStartOfDay(date).toISOString();
      this.$emit("update:start-date", startOfDay);
    },
    onEndDateChanged(date: Date): void {
      const endOfDay = dateUtil.getEndOfDay(date).toISOString();
      this.$emit("update:end-date", endOfDay);
    },
    rangeOfDatesCallback(date: string | null, isStart: boolean): boolean {
      if (!date) return false;
      const THIRTY_DAYS_IN_MS = 30 * 24 * 60 * 60 * 1000;
      return dateUtil.rangeOfDatesCallback(date, this.startDate, this.endDate, isStart, THIRTY_DAYS_IN_MS);
    },
  },
});
</script>

<style scoped lang="scss"></style>
