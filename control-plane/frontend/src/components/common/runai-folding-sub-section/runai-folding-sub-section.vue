<template>
  <section class="row no-wrap items-center">
    <q-expansion-item
      :label="label"
      class="runai-folding-sub-section"
      expand-icon-class="runai-expansion-icon"
      expand-separator
      expand-icon-toggle
      :model-value="isOpen"
      :header-class="headerClass"
      hide-expand-icon
    >
      <template #header>
        <div
          class="runai-folding-sub-section-header"
          :class="{ 'is-open': isOpen }"
          @click="isOpen = !isOpen"
          @keydown.enter.prevent="isOpen = !isOpen"
          @keydown.space.prevent="isOpen = !isOpen"
          :aid="aid"
          role="button"
          tabindex="0"
          :aria-expanded="isOpen"
          :aria-label="label"
        >
          <span class="q-mr-md">{{ label }}</span>
          <q-icon class="q-mr-md folding-icon" :class="isOpen ? 'open' : ''" name="ph-bold ph-caret-down" size="20px" />
          <q-icon v-if="hasError && !isOpen" name="ph-bold ph-warning-circle" color="red" size="20px" />
        </div>
      </template>
      <q-form class="q-mb-md" greedy ref="formRef" @vue:updated="updateRenderKey" @validation-error="updateRenderKey">
        <slot v-if="!lazyMount || hasBeenOpened" />
      </q-form>
    </q-expansion-item>
  </section>
</template>

<script lang="ts">
import type { PropType, Ref } from "vue";
import { defineComponent, ref, nextTick } from "vue";
import type { QField, QForm, QFormChildComponent } from "quasar";
import { useFormChild } from "quasar";

export default defineComponent({
  name: "runai-folding-sub-section",
  emits: ["validation-state-changed"],
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
    headerClass: {
      type: String as PropType<string>,
      default: "no-padding",
    },
    /**
     * When true, slot content is only rendered after the section has been opened at least once.
     * Useful for preventing unnecessary API calls or heavy component initialization on initial load.
     * Once opened, content stays mounted even when section is closed.
     */
    lazyMount: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      isOpen: !!this.defaultOpened,
      hasBeenOpened: !!this.defaultOpened, // Track if section was ever opened
      renderKey: 1000,
    };
  },
  setup() {
    const formRef: Ref<QForm | null> = ref(null);

    const validate = async () => {
      return await (formRef.value as QForm)?.validate();
    };

    useFormChild({
      validate,
    });

    return { formRef };
  },
  mounted() {
    // Quasar's q-expansion-item renders a q-item with role="listitem" internally.
    // Since this component is not inside a role="list" container, that creates an
    // AXE-ARIA-REQUIRED-PARENT violation. Remove the invalid role.
    const qItem = this.$el.querySelector('.q-item[role="listitem"]');
    if (qItem) {
      qItem.removeAttribute("role");
    }
  },
  computed: {
    hasError(): boolean {
      this.renderKey; // DO NOT REMOVE THIS LINE - it's critical to trigger reactivity when new components are dynamically added/removed to the QForm
      return !!(this.$refs.formRef as QForm)?.getValidationComponents().some((c: QFormChildComponent) => {
        return (c as QField).hasError;
      });
    },
  },
  methods: {
    updateRenderKey() {
      this.renderKey++;
    },
    /**
     * Programmatically opens the section and waits for content to render.
     * Useful for guided tours that need to wait for lazy-mounted content.
     */
    async open(): Promise<void> {
      if (this.isOpen) return;
      this.hasBeenOpened = true;
      this.isOpen = true;
      await nextTick();
    },
  },
  watch: {
    isOpen(newVal: boolean) {
      if (newVal && !this.hasBeenOpened) {
        this.hasBeenOpened = true;
      }
    },
    hasError: {
      handler(hasError: boolean) {
        this.$emit("validation-state-changed", hasError);
      },
      immediate: true,
    },
  },
});
</script>

<style lang="scss">
.runai-folding-sub-section {
  width: 100%;

  .runai-folding-sub-section-header {
    font-weight: 500;
    color: $primary;
    display: flex;
    align-items: center;
    cursor: pointer;

    .folding-icon {
      color: var(--interactive-secondary-default-icon, #65a800);
      transform: rotate(0deg);
      transition: all 0.3s ease-in-out;

      &.open {
        transform: rotate(180deg);
      }
    }
  }
}
</style>
