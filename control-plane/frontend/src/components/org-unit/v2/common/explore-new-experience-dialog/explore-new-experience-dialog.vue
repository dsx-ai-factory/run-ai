<template>
  <runai-base-dialog
    :model-value="modelValue"
    :hide-x="true"
    :is-modal-persistent="true"
    custom-class="explore-new-experience-dialog"
    :aid="dialogAid"
    @close="showLeaveConfirmation"
  >
    <template #header>
      <div class="row items-center justify-between full-width">
        <span>Explore the New Experience</span>
        <q-btn
          class="explore-dialog-close-btn text-black-54"
          icon="ph-bold ph-x close-icon"
          flat
          round
          :aid="closeBtnAid"
          @click="showLeaveConfirmation"
        >
          <q-tooltip>Leave the guide</q-tooltip>
        </q-btn>
      </div>
    </template>

    <template #body>
      <div>
        Start a guided tour to explore the new {{ entityLabel }} creation experience, which gives you more visibility and
        control over resource distribution.
      </div>
      <div class="q-mt-xl">To view the guide later, go to Help > Guided steps</div>
    </template>

    <template #footer>
      <q-btn flat color="primary" label="Skip" :aid="skipBtnAid" @click="showLeaveConfirmation" />
      <q-btn color="primary" label="Start guided tour" :aid="startBtnAid" @click="handleStartTour" />
    </template>
  </runai-base-dialog>

  <!-- Leave Confirmation Modal -->
  <guided-tour-leave-confirm-modal v-model="showConfirmModal" @go-back="hideLeaveConfirmation" @leave="confirmLeave" />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { GuidedTourLeaveConfirmModal } from "@/components/common/guided-tour-popup";

export type EntityType = "project" | "department";

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    entityType?: EntityType;
  }>(),
  {
    entityType: "project",
  },
);

const emit = defineEmits<{
  (e: "update:model-value", value: boolean): void;
  (e: "skip"): void;
  (e: "start-tour"): void;
}>();

const showConfirmModal = ref(false);
const isLeaving = ref(false);

const entityLabel = computed(() => props.entityType);

const dialogAid = computed(() => `explore-new-experience-dialog-${props.entityType}`);
const skipBtnAid = computed(() => `explore-dialog-skip-btn-${props.entityType}`);
const startBtnAid = computed(() => `explore-dialog-start-btn-${props.entityType}`);
const closeBtnAid = computed(() => `explore-dialog-close-btn-${props.entityType}`);

const closeDialog = (): void => {
  emit("update:model-value", false);
};

const showLeaveConfirmation = (): void => {
  if (isLeaving.value) return;
  showConfirmModal.value = true;
};

const hideLeaveConfirmation = (): void => {
  showConfirmModal.value = false;
};

const confirmLeave = (): void => {
  isLeaving.value = true;
  showConfirmModal.value = false;
  emit("skip");
  closeDialog();
};

const handleStartTour = (): void => {
  isLeaving.value = true;
  emit("start-tour");
  closeDialog();
};
</script>

<style lang="scss">
.explore-new-experience-dialog {
  width: 490px;

  .runai-base-dialog-header {
    overflow: visible;

    h6 {
      width: 100%;
      overflow: visible;
    }
  }

  .explore-dialog-close-btn {
    flex-shrink: 0;
  }

  .close-icon {
    &.q-icon {
      font-size: 20px;
    }
  }
  .runai-base-dialog-footer {
    padding: 16px;
  }
}
</style>
