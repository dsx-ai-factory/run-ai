<template>
  <runai-expansion-item
    :default-opened="defaultOpen"
    :label="label"
    :disable-closing="disableClosing"
    :section-invalid="sectionInvalid"
    should-validate
    class="workload-name-section"
    aid="workload-name-section"
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>
    <section class="row">
      <runai-name-validation
        ref="workload-name-input"
        :model-value="name"
        @update:model-value="$emit('update:name', $event)"
        :rules="[isEmpty, isValidFormat, isNameUnique]"
        @invalid="nameAlreadyExist = $event"
        placeholder="Enter a name"
        class="col-6 workload-name-input"
        debounce="300"
        :aid="entityType + '-name-input'"
        :auto-focus="autoFocus"
        :reactive-rules="reactiveRules"
      />
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { errorMessages } from "@/common/error-message.constant";
import { isValidRunaiEntityNameFormat, isNotEmpty, isEmpty } from "@/common/form.validators";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiNameValidation,
  },
  emits: ["update:name", "is-section-invalid"],
  props: {
    entityType: {
      type: String as PropType<string>,
      required: true,
    },
    name: {
      type: String as PropType<string>,
      required: true,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    projectId: {
      type: [Number, null] as PropType<number | null>,
      required: true,
    },
    defaultOpen: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    reactiveRules: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return {
      nameAlreadyExist: false as boolean,
      autoFocus: false as boolean,
    };
  },
  created() {
    this.autoFocus = !!this.$route.query.fromCopyId;
  },
  computed: {
    label(): string {
      return `${this.entityType.charAt(0).toUpperCase() + this.entityType.slice(1)} name`;
    },
    sectionInvalid(): boolean {
      return isEmpty(this.name) || !isValidRunaiEntityNameFormat(this.name) || this.nameAlreadyExist;
    },
    summary(): string {
      return this.name || "None";
    },
  },
  methods: {
    isValidFormat(val: string): boolean | string {
      return isValidRunaiEntityNameFormat(val) || errorMessages.VALID_FORMAT;
    },
    isEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    async isNameUnique(name: string): Promise<boolean | string> {
      const numOfEqualNames: number = await workloadService.countWorkloadsByName(name, this.clusterId, this.projectId);
      this.nameAlreadyExist = numOfEqualNames > 0;
      return this.nameAlreadyExist ? errorMessages.NAME_ALREADY_EXIST : true;
    },
  },
  watch: {
    sectionInvalid: {
      handler(newVal: boolean): void {
        this.$emit("is-section-invalid", newVal);
      },
      immediate: true,
    },
  },
});
</script>
