<template>
  <q-menu
    ref="popupMenu"
    v-model="showPopover"
    :anchor="anchor"
    :self="self"
    :offset="offset"
    no-parent-event
    :persistent="true"
    :target="target"
    class="guided-tour-popover"
    :class="[customClass, actualArrowClass]"
    transition-show="scale"
    transition-hide="scale"
    @show="onShow"
    @hide="onHide"
    @after-show="onAfterShow"
  >
    <div id="guided-tour-popover-content" class="guided-tour-content">
      <div class="guided-tour-header">
        <h5 class="guided-tour-title q-ma-none" v-dompurify-html="title"></h5>

        <q-btn
          v-if="showCloseButton"
          class="guided-tour-close-btn text-black-54"
          icon="ph-bold ph-x"
          flat
          round
          @click="handleClose"
        >
          <q-tooltip class="guided-tour-tooltip" anchor="top middle" :offset="[0, 25]"> Skip guide </q-tooltip>
        </q-btn>

        <!-- Debug button for testing -->
        <q-btn
          v-if="showDebugButton"
          class="guided-tour-debug-btn q-ml-sm"
          icon="ph-bold ph-bug"
          flat
          round
          size="sm"
          @click="debugArrowPosition"
        >
          <q-tooltip class="guided-tour-tooltip" anchor="top middle" :offset="[0, 25]"> Debug arrow </q-tooltip>
        </q-btn>
      </div>

      <div v-if="description" class="guided-tour-description q-mt-sm">
        <div class="text-body2" v-dompurify-html="description"></div>
      </div>

      <div v-if="showFooter" class="guided-tour-footer q-mt-md">
        <div class="guided-tour-progress-section">
          <div v-if="showProgress" class="guided-tour-progress text-caption text-black-54">
            {{ progressText }}
          </div>
        </div>

        <div class="guided-tour-actions q-gutter-sm">
          <q-btn
            v-if="showPreviousButton"
            :label="prevBtnText"
            outline
            color="primary"
            size="md"
            flat
            class="guided-tour-prev-btn"
            :disable="!hasPrev"
            @click="handlePrevious"
          />

          <q-btn
            v-if="showNextButton"
            :label="nextBtnText"
            color="primary"
            size="md"
            flat
            class="guided-tour-next-btn"
            @click="handleNext"
          />
        </div>
      </div>
    </div>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import type { TourButton } from "@/components/common/guided-tour-popup/guided-tour.types";
import type { QMenu } from "quasar";

export type MenuAnchor = QMenu["anchor"];

export default defineComponent({
  name: "guided-tour-popup",
  emits: {
    "update:modelValue": (value: boolean) => typeof value === "boolean",
    show: () => true,
    hide: () => true,
    next: () => true,
    previous: () => true,
    close: () => true,
  },
  props: {
    modelValue: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    anchor: {
      type: String as PropType<MenuAnchor>,
      default: "bottom middle" as MenuAnchor,
    },
    self: {
      type: String as PropType<MenuAnchor>,
      default: "top middle" as MenuAnchor,
    },
    offset: {
      type: Array as unknown as PropType<[number, number]>,
      default: () => [10, 10] as [number, number],
    },
    target: {
      type: [Boolean, String] as PropType<boolean | string>,
      default: false,
    },
    title: {
      type: String as PropType<string>,
      default: "",
    },
    description: {
      type: String as PropType<string>,
      default: "",
    },
    showProgress: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    progressText: {
      type: String as PropType<string>,
      default: "",
    },
    showButtons: {
      type: Array as PropType<TourButton[]>,
      default: () => ["next", "previous", "close"],
    },
    hasNext: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    hasPrev: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    nextBtnText: {
      type: String as PropType<string>,
      default: "Next",
    },
    prevBtnText: {
      type: String as PropType<string>,
      default: "Previous",
    },
    customClass: {
      type: String as PropType<string>,
      default: "",
    },
  },
  data() {
    return {
      actualPosition: null as string | null,
      resizeTimeout: null as ReturnType<typeof setTimeout> | null,
      positionDetectionTimeout: null as ReturnType<typeof setTimeout> | null,
      isClosing: false,
    };
  },
  mounted() {
    // Add resize listener to handle position changes
    window.addEventListener("resize", this.handleResize);

    // Add keydown listener for ESC key to close popup
    document.addEventListener("keydown", this.handleKeydown);
  },
  computed: {
    showPopover: {
      get(): boolean {
        return this.modelValue;
      },
      set(value: boolean): void {
        this.$emit("update:modelValue", value);
      },
    },
    showButtonsList(): TourButton[] {
      return this.showButtons || [];
    },
    showCloseButton(): boolean {
      return this.showButtonsList.includes("close");
    },
    showNextButton(): boolean {
      return this.showButtonsList.includes("next");
    },
    showPreviousButton(): boolean {
      return this.showButtonsList.includes("previous");
    },
    showFooter(): boolean {
      return this.showNextButton || this.showPreviousButton || this.showProgress;
    },
    arrowClass(): string {
      const anchor = this.anchor || "bottom middle";
      const parts = anchor.split(" ");
      const vertical = parts[0];
      const horizontal = parts[1];

      let arrowSide = "";

      if (vertical === "top") {
        arrowSide = "down";
      } else if (vertical === "bottom") {
        arrowSide = "up";
      } else if (vertical === "center") {
        if (horizontal === "left") {
          arrowSide = "right";
        } else if (horizontal === "right") {
          arrowSide = "left";
        }
      }

      return arrowSide ? `q-menu--arrow-${arrowSide}` : "";
    },
    actualArrowClass(): string {
      // Use detected position if available, otherwise fall back to intended position
      const arrowClass = this.actualPosition || this.arrowClass;

      return arrowClass;
    },
    showDebugButton(): boolean {
      return false;
    },
  },
  methods: {
    handleNext(): void {
      this.$emit("next");
    },
    handlePrevious(): void {
      this.$emit("previous");
    },
    handleClose(): void {
      // Set closing flag to prevent position detection
      this.isClosing = true;

      // Clear any pending position detection timeouts
      if (this.positionDetectionTimeout) {
        clearTimeout(this.positionDetectionTimeout);
        this.positionDetectionTimeout = null;
      }

      this.showPopover = false;
      this.$emit("close");
    },
    handleKeydown(event: KeyboardEvent): void {
      // Close popup when ESC key is pressed
      if (event.key === "Escape" && this.showPopover) {
        event.preventDefault();
        event.stopPropagation();
        this.handleClose();
      }
    },
    // Note to dev: the following methods and watchers are used to detect the actual position of the popup and to show the directional arrow correctly
    // Due to quasar's implementation, the popup is not always positioned correctly, so we need to detect the actual position of the popup and show the directional arrow correctly
    // Even though this is working most of the time, there are scenarios where it wont work especially when the target element is a proxy like a context-menu option which is located outside of the App main container.
    // Helper to get the actual DOM element from the QMenu component
    getPopupDomEl(): HTMLElement | null {
      const popupComponent = this.$refs.popupMenu as unknown as
        | { contentEl?: HTMLElement; $el?: HTMLElement }
        | undefined;
      if (!popupComponent) return null;
      return popupComponent.contentEl || popupComponent.$el || null;
    },
    // Helper method to check if DOM is ready
    isDOMReady(): boolean {
      const domElement = this.getPopupDomEl();
      if (!domElement) return false;

      // Simple check: just verify the element exists and has getBoundingClientRect
      return typeof domElement.getBoundingClientRect === "function";
    },
    resetPositionDetection(): void {
      // Reset retry count and position
      this.actualPosition = null;
    },
    handleResize(): void {
      // Debounce resize events to avoid excessive calls
      if (this.resizeTimeout) {
        clearTimeout(this.resizeTimeout);
      }
      this.resizeTimeout = setTimeout(() => {
        if (this.showPopover && !this.isClosing) {
          this.detectActualPosition();
        }
      }, 100);
    },
    onShow(): void {
      // Use nextTick to ensure DOM is updated, then add a small delay for positioning
      this.$nextTick(() => {
        this.positionDetectionTimeout = setTimeout(() => {
          if (this.showPopover && !this.isClosing) {
            // Only detect position if we haven't already
            if (!this.actualPosition) {
              this.detectActualPosition();
            }
          }
        }, 300); // Increased delay to ensure Quasar has finished positioning
      });
    },
    onHide(): void {
      // Reset position when hiding
      this.actualPosition = null;
    },
    onAfterShow(): void {
      // Detect position after Quasar's after-show event
      if (this.showPopover && !this.isClosing) {
        // Only detect position if we haven't already
        if (!this.actualPosition) {
          this.detectActualPosition();
        }
      }
    },
    detectActualPosition(): void {
      try {
        // Don't detect position if we're closing or the popup is hidden
        if (this.isClosing || !this.showPopover) {
          return;
        }

        // Check if DOM is ready before proceeding
        if (!this.isDOMReady()) {
          console.warn("[Guided Tour] DOM not ready for position detection, retry...");
          // Retry after a short delay
          this.positionDetectionTimeout = setTimeout(() => {
            if (this.showPopover && !this.isClosing) {
              this.detectActualPosition();
            }
          }, 100);
          return;
        }

        // Get the popup element directly from refs
        const popupElement = this.getPopupDomEl();
        if (!popupElement) {
          console.warn("[Guided Tour] Popup DOM element not accessible");
          return;
        }

        // Quasar components store the actual DOM element in contentEl
        // const popupElement = popupComponent.contentEl || popupComponent.$el || popupComponent;
        if (!popupElement || typeof popupElement.getBoundingClientRect !== "function") {
          console.warn("[Guided Tour] Popup DOM element not accessible:", popupElement);
          return;
        }

        // Get the target element if available
        let targetElement: Element | null = null;
        if (this.target && typeof this.target === "string") {
          targetElement = document.querySelector(this.target);
        }

        if (!targetElement) {
          console.warn("[Guided Tour] Target element not found:", this.target);
          return;
        }

        // Get the popup's bounding rect
        const popupRect = popupElement.getBoundingClientRect();
        const targetRect = targetElement.getBoundingClientRect();

        // Determine which side the popup is actually on relative to the target
        let arrowSide = "";

        // Vertical positioning
        if (popupRect.bottom < targetRect.top) {
          // Popup is above target
          arrowSide = "down";
        } else if (popupRect.top > targetRect.bottom) {
          // Popup is below target
          arrowSide = "up";
        } else {
          // Popup is at same vertical level, check horizontal
          if (popupRect.right < targetRect.left) {
            // Popup is to the left of target
            arrowSide = "right";
          } else if (popupRect.left > targetRect.right) {
            // Popup is to the right of target
            arrowSide = "left";
          }
        }

        // Update the actual arrow class
        this.actualPosition = arrowSide ? `q-menu--arrow-${arrowSide}` : "";
      } catch (error) {
        console.error("[Guided Tour] Error detecting position:", error);
      }
    },
    debugArrowPosition(): void {
      console.log("[Guided Tour] Debugging arrow position...");
      console.log("[Guided Tour] Current state:", {
        showPopover: this.showPopover,
        actualPosition: this.actualPosition,
        intendedPosition: this.arrowClass,
        isDOMReady: this.isDOMReady(),
        popupRef: this.$refs.popupMenu,
        target: this.target,
      });

      // Reset and try again
      this.resetPositionDetection();
      this.detectActualPosition();

      console.log("[Guided Tour] After detection - actualPosition:", this.actualPosition);
    },
  },
  watch: {
    showPopover(newValue: boolean) {
      if (newValue) {
        // Reset closing flag when popup becomes visible
        this.isClosing = false;

        // When popup becomes visible, detect position after a delay
        this.$nextTick(() => {
          this.positionDetectionTimeout = setTimeout(() => {
            if (this.showPopover && !this.isClosing && !this.actualPosition) {
              this.detectActualPosition();
            }
          }, 300); // Longer delay to ensure positioning is complete
        });
      } else {
        // Reset position when hiding
        this.actualPosition = null;
      }
    },
  },
  beforeUnmount() {
    // Clean up resize listener
    window.removeEventListener("resize", this.handleResize);

    // Clean up keydown listener
    document.removeEventListener("keydown", this.handleKeydown);

    // Clean up timeouts
    if (this.resizeTimeout) {
      clearTimeout(this.resizeTimeout);
    }
    if (this.positionDetectionTimeout) {
      clearTimeout(this.positionDetectionTimeout);
    }
  },
});
</script>
