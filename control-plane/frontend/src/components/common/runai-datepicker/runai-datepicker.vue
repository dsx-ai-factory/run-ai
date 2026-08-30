<template>
  <q-field
    outlined
    class="date-field clickable-field"
    :class="{ 'q-field--highlighted': isDatePickerVisible }"
    :label="label"
    @focus="isDatePickerVisible = true"
    :model-value="date"
    :dense="dense"
  >
    <template v-slot:control>
      <div class="self-center full-width no-outline">{{ date }}</div>
      <q-popup-proxy @before-hide="isDatePickerVisible = false">
        <q-date
          v-if="isDatePickerVisible"
          :model-value="date"
          :mask="dateFormat.toUpperCase()"
          @update:model-value="dateSelected"
          minimal
          :options="rangeOfDates"
        />
      </q-popup-proxy>
    </template>
  </q-field>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { dateUtil } from "@/utils/date.util";
import type { DateFormatType } from "@/constants/date-formats";
import { DATE_FORMATS } from "@/constants/date-formats";

export default defineComponent({
  emits: ["changed"],
  props: {
    defaultDate: {
      type: Date as PropType<Date>,
      required: false,
      default: new Date(),
    },
    dateFormat: {
      type: String as PropType<DateFormatType>,
      required: false,
      default: DATE_FORMATS.STANDARD_DATE,
    },
    rangeOfDates: {
      type: Function as unknown as PropType<string[] | ((date: string) => boolean) | undefined>,
      required: false,
      default: (date: string) => {
        return date >= "01/01/1900";
      },
    },
    dense: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      isDatePickerVisible: false,
    };
  },
  computed: {
    date(): string {
      return dateUtil.dateFormat(this.defaultDate, this.dateFormat);
    },
  },
  methods: {
    dateSelected(val: string | null): void {
      this.isDatePickerVisible = false;
      if (!val) return;
      this.$emit("changed", dateUtil.parseToDate(val, this.dateFormat));
    },
  },
});
</script>
