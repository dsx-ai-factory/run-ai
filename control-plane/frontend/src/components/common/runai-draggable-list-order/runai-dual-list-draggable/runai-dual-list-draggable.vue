<template>
  <section class="dual-list-draggable">
    <section class="list-container">
      <ul class="sortable-list">
        <vue-draggable
          class="dragArea list-group w-full"
          v-model="list1"
          :animation="150"
          ghost-class="ghost"
          group="people"
          @update="onUpdate"
          @add="onAdd"
          @remove="remove"
        >
          <transition-group type="transition" name="list">
            <slot name="sortable-list" />
          </transition-group>
        </vue-draggable>
      </ul>

      <slot name="separator" />

      <ul class="non-sortable-list">
        <vue-draggable
          class="dragArea list-group w-full"
          v-model="list2"
          :animation="150"
          group="people"
          ghost-class="ghost"
          :sort="false"
        >
          <transition-group type="transition" name="list">
            <slot name="non-sortable-list" />
          </transition-group>
        </vue-draggable>
      </ul>
    </section>
  </section>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";
import { VueDraggable } from "vue-draggable-plus";
import { stringUtil } from "@/utils/string.util";

export enum ERunaiDualListDraggableEmitName {
  SortableListChange = "sortable-list-change",
}

export default defineComponent({
  name: "RunaiDualListDraggable",
  components: {
    VueDraggable,
  },
  emits: {
    [ERunaiDualListDraggableEmitName.SortableListChange]: (sortableList: Array<any>) => sortableList, // eslint-disable-line @typescript-eslint/no-explicit-any
  },
  props: {
    sortableList: {
      type: Array as PropType<Array<any>>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    nonSortableList: {
      type: Array as PropType<Array<any>>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
  },
  data() {
    return {
      list1: [...this.sortableList],
      list2: [...this.nonSortableList],
    };
  },
  methods: {
    onUpdate() {
      this.$emit(ERunaiDualListDraggableEmitName.SortableListChange, this.list1);
    },
    onAdd() {
      this.$emit(ERunaiDualListDraggableEmitName.SortableListChange, this.list1);
    },
    remove() {
      this.$emit(ERunaiDualListDraggableEmitName.SortableListChange, this.list1);
    },
  },
  watch: {
    // Sync internal lists when props change to prevent mismatch
    // between visual representation and internal drag state
    sortableList: {
      // eslint-disable-next-line
      handler(newList: Array<any>) {
        // Only update if the lists are actually different (avoid infinite loops)
        if (stringUtil.toString(this.list1) !== stringUtil.toString(newList)) {
          this.list1 = [...newList];
        }
      },
      deep: true,
    },
    nonSortableList: {
      // eslint-disable-next-line
      handler(newList: Array<any>) {
        if (stringUtil.toString(this.list2) !== stringUtil.toString(newList)) {
          this.list2 = [...newList];
        }
      },
      deep: true,
    },
  },
});
</script>
<style lang="scss">
.dual-list-draggable {
  .ghost {
    background: transparent !important;
    border: $black-54 1px dashed;

    > div,
    > svg {
      opacity: 0;
    }
  }
}
</style>
