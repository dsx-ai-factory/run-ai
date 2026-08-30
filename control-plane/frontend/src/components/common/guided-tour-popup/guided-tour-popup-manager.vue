<template>
  <div class="guided-tour-wrapper">
    <!-- Guided Tour Popup using target positioning -->
    <guided-tour-popup
      v-if="currentElement && showPopover"
      v-bind="popoverProps"
      @next="onNext"
      @previous="onPrevious"
      @close="onClose"
    />

    <!-- Exit Confirmation Modal -->
    <guided-tour-leave-confirm-modal
      :model-value="showExitConfirmModal"
      @go-back="hideExitConfirmation"
      @leave="confirmExit"
    />

    <!-- Completion Modal -->
    <guided-tour-completion-modal :open="showCompletionModal" @close="closeCompletionModal" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, onUnmounted } from "vue";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
import GuidedTourPopup from "./guided-tour-popup.vue";
import GuidedTourLeaveConfirmModal from "./guided-tour-leave-confirm-modal.vue";
import { GuidedTourCompletionModal } from "@/components/common/guided-tour-completion-modal";

export default defineComponent({
  name: "guided-tour-manager",
  components: {
    GuidedTourPopup,
    GuidedTourLeaveConfirmModal,
    GuidedTourCompletionModal,
  },
  emits: {
    close: () => true,
  },
  setup() {
    const store = useGuidedTourStore();

    store.setWorkloadTourAllowed();
    onUnmounted(() => {
      store.destroy();
    });

    return {
      // State
      showPopover: computed(() => store.showPopover),
      popoverProps: computed(() => store.popoverProps),
      currentElement: computed(() => store.currentElement),
      showExitConfirmModal: computed(() => store.showExitConfirmModal),
      showCompletionModal: computed(() => store.showCompletionModal),

      // Handlers
      closeCompletionModal: () => store.closeCompletionModal(),
      hideExitConfirmation: () => store.hideExitConfirmation(),
      confirmExit: () => store.confirmExit(),
      onNext: () => store.onNext(),
      onPrevious: () => store.onPrevious(),
      onClose: () => store.onClose(),
    };
  },
});
</script>
