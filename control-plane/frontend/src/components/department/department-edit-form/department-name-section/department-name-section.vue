<template>
  <runai-expansion-item label="Department name & description" default-opened :section-invalid="sectionInvalid">
    <template #subheader>
      <span>
        {{ departmentName || "&lt;Department name&gt;" }}
      </span>
    </template>
    <div class="row">
      <div class="col-7">
        <runai-tooltip-wrapper :display-tooltip="isInputDisabled" tooltip-text="This can't be modified">
          <runai-name-validation
            aid="department-name-input"
            :disable="isInputDisabled"
            :model-value="departmentName"
            @update:model-value="$emit('update:departmentName', $event)"
            :rules="[isValidEntityName, isNameUnique]"
          />
        </runai-tooltip-wrapper>
      </div>
    </div>

    <div class="row q-mt-md">
      <div class="col-11">
        <q-input
          aid="department-description-input"
          :model-value="description"
          :disable="readOnly"
          @update:model-value="$emit('update:description', $event)"
          :maxlength="250"
          counter
          label="Description"
          autogrow
        />
      </div>
    </div>
  </runai-expansion-item>
</template>
<script lang="ts">
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { defineComponent, type PropType } from "vue";
import { isValidEntityName } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

export default defineComponent({
  name: "department-name-section",
  components: {
    RunaiNameValidation,
    RunaiExpansionItem,
    RunaiTooltipWrapper,
  },
  emits: ["update:departmentName", "is-section-invalid", "update:description"],
  props: {
    departmentName: {
      type: String as PropType<string>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    description: {
      type: String as PropType<string>,
      required: false,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      isDepartmentNameUnique: true as boolean,
    };
  },
  computed: {
    sectionInvalid(): boolean {
      return !isValidEntityName(this.departmentName) || !this.isDepartmentNameUnique;
    },
    isInputDisabled(): boolean {
      return this.readOnly || this.isEditMode;
    },
  },
  methods: {
    isValidEntityName(name: string): boolean | string {
      return isValidEntityName(name) || errorMessages.VALID_FORMAT;
    },
    async isNameUnique(name: string): Promise<boolean | string> {
      try {
        const isNameUnique = await orgUnitService.isDepartmentNameUnique(this.clusterId, name);
        this.isDepartmentNameUnique = isNameUnique;
        if (!isNameUnique) {
          return errorMessages.NAME_ALREADY_EXIST;
        }
        return true;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedToValidateMessage("department name"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error(error);
        return false;
      }
    },
  },
  watch: {
    sectionInvalid: {
      handler(isSectionInvalid: boolean): void {
        this.$emit("is-section-invalid", isSectionInvalid);
      },
      immediate: true,
    },
  },
});
</script>

<style scoped lang="scss"></style>
