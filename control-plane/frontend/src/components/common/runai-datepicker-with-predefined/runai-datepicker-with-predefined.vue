<template>
  <div class="predefined-datepicker">
    <q-btn-dropdown no-caps outline @hide="onPopupHide" class="date-picker-dropdown" :aid="aid">
      <template #label>
        <q-item-label class="q-pr-md date-picker-label-wrapper">
          <q-icon
            name="ph-bold ph-calendar-dot q-mr-sm text-icon-default"
            size="20px"
            class="dash-icon"
            aid="selected-label"
          />
          <div>
            <div class="text-black-30 selected-option-label">{{ selectedOptionLabel }}</div>
            <div class="date-picker-label">{{ title }}</div>
          </div>
        </q-item-label>
      </template>
      <q-item class="dates-info" role="presentation" v-if="infoBarMessage">
        <q-item-section>
          <q-item-label>{{ infoBarMessage }}</q-item-label>
        </q-item-section>
      </q-item>
      <!-- Custom dates -->
      <q-item
        clickable
        v-if="hasCustom"
        role="menuitem"
        :class="{ selected: 'Custom' === selectedPredefinedOption.label }"
        class="custom-label"
      >
        <q-item-section avatar>
          <q-icon
            size="16px"
            name="ph-bold ph-check"
            color="primary"
            v-if="'Custom' === selectedPredefinedOption.label"
          />
        </q-item-section>
        <q-item-section @click="onCustomDateChange(startDate, true)" aid="custom-select">
          <q-item-label> Custom</q-item-label>
        </q-item-section>
        <q-item-section class="q-pr-md">
          <runai-datepicker
            aid="custom-select-start"
            :default-date="startDate"
            @changed="onCustomDateChange($event, true)"
            :range-of-dates="(date) => innerRangeOfDatesCallback(date, true)"
          ></runai-datepicker>
        </q-item-section>
        <q-icon name="ph-bold ph-minus" size="16px" class="q-px-md" />
        <q-item-section class="q-pr-md">
          <runai-datepicker
            aid="custom-select-end"
            :default-date="endDate"
            @changed="onCustomDateChange($event, false)"
            :range-of-dates="(date) => innerRangeOfDatesCallback(date, false)"
          ></runai-datepicker>
        </q-item-section>
      </q-item>
      <!-- Predefined dates -->
      <q-list role="group" aria-label="Predefined date ranges">
        <q-item
          v-for="option in filteredPredefinedOptions"
          :key="option.label"
          clickable
          v-close-popup
          role="menuitemradio"
          :aria-checked="option.label === selectedPredefinedOption.label"
          @click="onSelectPredefinedDate(option)"
          :class="{
            selected: option.label === selectedPredefinedOption.label,
          }"
          :disable="option.disabled"
          :aid="option.label"
        >
          <q-item-section avatar>
            <q-icon
              size="16px"
              name="ph-bold ph-check"
              color="primary"
              v-if="option.label === selectedPredefinedOption.label"
            />
          </q-item-section>
          <q-item-section>
            {{ option.label }}
          </q-item-section>
        </q-item>
      </q-list>
    </q-btn-dropdown>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import type { PredefinedOption } from "@/models/date.model";
import { PREDEFINED_DATES } from "@/models/date.model";
import { RunaiDatepicker } from "@/components/common/runai-datepicker";
import { defaultPredefinedOptions, dateUtil, TimeUnit } from "@/utils/date.util";
import type { DateFormatType } from "@/constants/date-formats";

export default defineComponent({
  components: {
    RunaiDatepicker,
  },
  emits: ["changed"],
  props: {
    defaultEndDate: {
      type: Date as PropType<Date>,
      required: false,
    },
    workloadLifespanEndDate: {
      type: Date as PropType<Date>,
      required: false,
    },
    workloadFirstRunDate: {
      type: Date as PropType<Date>,
      required: false,
    },
    defaultOptionLabel: {
      type: String as PropType<string>,
      required: false,
    },
    defaultStartDate: {
      type: Date as PropType<Date>,
      required: false,
    },
    dateFormat: {
      type: String as PropType<DateFormatType>,
      required: false,
      default: "MMM dd",
    },
    predefinedOptions: {
      type: Array as PropType<PredefinedOption[]>,
      required: false,
      default: defaultPredefinedOptions,
    },
    defaultOptionIndex: {
      type: Number as PropType<number>,
      default: 2,
    },
    hasCustom: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    rangeOfDates: {
      type: Function as PropType<(date: string) => boolean>,
      required: false,
    },
    infoBarMessage: {
      type: String as PropType<string>,
      required: false,
      default: null,
    },
    disableBeforeDate: {
      type: Date as PropType<Date>,
      required: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      startDate: this.getDefaultStartDate() as Date,
      endDate: new Date() as Date,
      selectedPredefinedOption: this.predefinedOptions[this.defaultOptionIndex] as PredefinedOption,
    };
  },
  created() {
    if (this.predefinedOptions.find((option) => option.label === PREDEFINED_DATES.WORKLOADS_LIFESPAN)) {
      if (!this.defaultStartDate) {
        console.error("defaultStartDate is required when using WORKLOADS_LIFESPAN predefined option");
      }
    }
    this.setDefaultDate();
  },
  computed: {
    filteredPredefinedOptions(): PredefinedOption[] {
      let options = this.predefinedOptions;
      if (this.disableBeforeDate) {
        options = dateUtil.disablePredefinedOptionsBeforeDate(options, this.disableBeforeDate);
      }
      return options.map((option) => {
        if (option.label === PREDEFINED_DATES.SINCE_FIRST_RUN && !this.workloadFirstRunDate) {
          return { ...option, disabled: true };
        }
        return option;
      });
    },
    selectedOptionLabel(): string {
      return this.selectedPredefinedOption.label;
    },
    title(): string {
      const startYear = this.startDate.getFullYear();
      const endYear = this.endDate.getFullYear();
      return startYear === endYear
        ? `${dateUtil.dateFormat(this.startDate, this.dateFormat)} - ${dateUtil.dateFormat(
            this.endDate,
            this.dateFormat,
          )}, ${startYear}`
        : `${dateUtil.dateFormat(this.startDate, this.dateFormat)}, ${startYear} - ${dateUtil.dateFormat(
            this.endDate,
            this.dateFormat,
          )}, ${endYear}`;
    },
  },
  methods: {
    setDefaultDate(): void {
      this.setDefaultEndDate();
      this.setDefaultStartDate();
      this.setDefaultOptionLabel();
      this.onSelectPredefinedDate(this.selectedPredefinedOption);
    },
    setDefaultEndDate(): void {
      if (this.defaultEndDate) {
        this.endDate = this.defaultEndDate;
      } else {
        this.endDate = new Date();
      }
    },
    setDefaultStartDate(): void {
      if (this.defaultStartDate) {
        this.startDate = this.defaultStartDate;
      } else {
        this.startDate = dateUtil.dateSevenDaysAgo();
      }
    },
    setDefaultOptionLabel(): void {
      if (this.defaultOptionLabel) {
        const defaultOption = this.predefinedOptions.find((option) => option.label === this.defaultOptionLabel);
        if (defaultOption) {
          this.selectedPredefinedOption = defaultOption;
        }
      }
    },
    innerRangeOfDatesCallback(date: string, isStartDate: boolean): boolean {
      if (!this.rangeOfDates) return true;

      if (this.disableBeforeDate) {
        const currentTime = new Date(date).getTime();
        if (dateUtil.isSameDay(this.disableBeforeDate.getTime(), currentTime)) {
          return true;
        }
        return currentTime >= this.disableBeforeDate.getTime() && currentTime <= new Date().getTime();
      }
      const dateToCheck: number = Date.parse(date);
      if (isStartDate) {
        return this.rangeOfDates(date);
      }

      if (dateUtil.isSameDay(this.startDate.getTime(), dateToCheck)) {
        return true;
      }

      return this.rangeOfDates(date) && dateToCheck >= this.startDate.getTime();
    },
    onPopupHide(): void {
      if (this.selectedPredefinedOption.daysCount || this.selectedPredefinedOption.label === "Today") {
        this.adjustDates();
      }

      let dateEnd: Date;
      if (dateUtil.isDateIsToday(this.endDate)) {
        dateEnd = new Date();
      } else {
        dateEnd = dateUtil.getEndOfDay(this.endDate);
      }
      const startDate = dateUtil.setSecondsToDate(this.startDate, 0);
      this.$emit("changed", [startDate, dateEnd], this.selectedOptionLabel);
    },
    onSelectPredefinedDate(selectedOption: PredefinedOption) {
      this.selectedPredefinedOption = selectedOption;

      if (selectedOption.label === PREDEFINED_DATES.WORKLOADS_LIFESPAN && this.defaultStartDate) {
        this.startDate = this.defaultStartDate;
        this.endDate = this.workloadLifespanEndDate || this.defaultEndDate || new Date();
        return;
      }

      if (selectedOption.label === PREDEFINED_DATES.SINCE_FIRST_RUN && this.workloadFirstRunDate) {
        this.startDate = this.workloadFirstRunDate;
        this.endDate = this.workloadLifespanEndDate || this.defaultEndDate || new Date();
        return;
      }
      this.endDate = new Date();

      if (selectedOption.daysCount != undefined) {
        this.startDate = dateUtil.adjustDateBy(TimeUnit.day, new Date(), 0 - selectedOption.daysCount);
      } else if (selectedOption.hoursCount != undefined) {
        this.startDate = dateUtil.adjustDateBy(TimeUnit.hour, new Date(), 0 - selectedOption.hoursCount);
      } else if (selectedOption.minutesCount != undefined) {
        this.startDate = dateUtil.adjustDateBy(TimeUnit.minute, new Date(), 0 - selectedOption.minutesCount);
      }
    },
    onCustomDateChange(date: Date, isStartDate: boolean): void {
      this.selectedPredefinedOption = {
        label: "Custom",
        daysCount: 0,
        hoursCount: 0,
      };
      isStartDate ? (this.startDate = date) : (this.endDate = date);
      const days: number = dateUtil.differenceBy(TimeUnit.day, this.endDate, this.startDate);
      if (days > 30) {
        if (isStartDate) {
          this.endDate = dateUtil.adjustDateBy(TimeUnit.day, date, 30);
        } else {
          this.startDate = dateUtil.adjustDateBy(TimeUnit.day, date, -30);
        }
      }
    },
    adjustDates(): void {
      dateUtil.adjustToStartOfTheDay(this.startDate);
      if (this.selectedPredefinedOption.endDate) {
        this.endDate = this.selectedPredefinedOption.endDate;
      }
      dateUtil.adjustToEndOfTheDay(this.endDate);
    },
    getDefaultStartDate(): Date {
      const defaultPredefined = this.predefinedOptions[this.defaultOptionIndex];

      if (defaultPredefined.label === PREDEFINED_DATES.WORKLOADS_LIFESPAN && this.defaultStartDate) {
        return this.defaultStartDate;
      }

      if (defaultPredefined.label === PREDEFINED_DATES.SINCE_FIRST_RUN && this.workloadFirstRunDate) {
        return this.workloadFirstRunDate;
      }

      if (defaultPredefined.daysCount != undefined) {
        return dateUtil.adjustDateBy(TimeUnit.day, new Date(), 0 - defaultPredefined.daysCount);
      } else if (defaultPredefined.hoursCount != undefined) {
        return dateUtil.adjustDateBy(TimeUnit.hour, new Date(), 0 - defaultPredefined.hoursCount);
      } else if (defaultPredefined.minutesCount != undefined) {
        return dateUtil.adjustDateBy(TimeUnit.minute, new Date(), 0 - defaultPredefined.minutesCount);
      } else {
        throw new Error("Invalid default option, daysCount, hoursCount or minutesCount must be defined.");
      }
    },
  },
});
</script>
<style scoped lang="scss">
.predefined-datepicker {
  display: flex;
  justify-content: right;
  align-items: center;
  .date-picker-dropdown {
    border-radius: 4px;
    height: 40px;
  }
  .date-picker-label-wrapper {
    display: flex;
    align-items: center;
    text-align: start;
    .date-picker-label {
      font-size: 14px;
      color: $black-87;
      font-weight: 400;
      line-height: 16.41px;
    }
    .selected-option-label {
      font-size: 10px;
      color: $black-54;
      font-weight: 400;
      line-height: 11.72px;
    }
  }
}
.selected {
  background-color: $option-selected;
}
.custom-label {
  align-items: center;
}
.q-item__section--avatar {
  min-width: 30px;
}
.q-icon {
  font-size: 10px;
}
.q-btn {
  height: 30px;
  font-size: 10px;
}
.q-btn--outline:before {
  border: 1px solid rgba(0, 0, 0, 0.24);
}

.q-field {
  cursor: pointer;
}

.dates-info {
  background-color: $body-background-color;
  font-size: 12px;
  text-align: center;
}
</style>
<style lang="scss">
.predefined-datepicker {
  .block {
    font-size: 14px;
    font-weight: 400;
  }
}
</style>
