<template>
  <label
    class="runai-general-search"
    :class="{ focus: isFocus, filled: searchTerm }"
    :for="inputId"
    v-click-outside="blur"
  >
    <div class="icon-container q-px-md">
      <i class="ph-bold ph-magnifying-glass"></i>
      <span v-if="showSearchText">SEARCH</span>
    </div>
    <input
      :id="inputId"
      type="search"
      :value="searchTerm"
      class="search-input"
      ref="search-input"
      @input="handleInput($event)"
      @focus="focus"
      aria-label="Search"
    />
    <i
      class="ph-bold ph-x search-clear"
      v-if="searchTerm"
      role="button"
      tabindex="0"
      aria-label="Clear search"
      @click.stop="clearSearch"
      @keydown.enter.prevent="clearSearch"
      @keydown.space.prevent="clearSearch"
    ></i>
    <q-tooltip v-if="tooltipText"> {{ tooltipText }}</q-tooltip>
  </label>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { debounce } from "quasar";

export default defineComponent({
  emits: ["update-search-term"],
  props: {
    searchTerm: {
      type: String,
      required: true,
    },
    showSearchText: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    tooltipText: {
      type: String as PropType<string>,
      default: "",
    },
  },
  data() {
    return {
      isFocus: false,
      elSearchInput: null as null | HTMLInputElement,
      inputId: `runai-general-search-input-${this.$.uid}`,
    };
  },
  created() {
    this.updateSearch = debounce(this.updateSearch, 300);
  },
  mounted() {
    this.elSearchInput = this.$refs["search-input"] as HTMLInputElement;
  },
  methods: {
    focus(): void {
      this.isFocus = true;
      this.elSearchInput?.focus();
    },
    setInputValue(val: string): void {
      (this.elSearchInput as HTMLInputElement).value = val;
    },
    blur(): void {
      this.isFocus = false;
    },
    updateSearch(val: string): void {
      this.$emit("update-search-term", val);
    },
    clearSearch(): void {
      this.isFocus = false;
      this.setInputValue("");
      this.$emit("update-search-term", "");
    },
    handleInput(ev: Event): void {
      const value: string = (ev.target as HTMLInputElement).value;
      this.setInputValue(value);
      this.updateSearch(value);
    },
  },
});
</script>

<style lang="scss">
.runai-general-search {
  position: relative;
  display: flex;
  align-items: center;
  outline: 1px solid transparent;
  border-radius: 3px;
  transition: all 0.3s;
  cursor: pointer;

  &:hover {
    background-color: $grey-3;
  }
  &.focus,
  &.filled {
    background-color: $grey-3;
    outline: 1px solid $grey-3;
    .search-input {
      width: 200px;
      padding-right: 35px;
    }
    .icon-container {
      span {
        display: none;
      }
    }
  }
  &.focus {
    outline-color: var(--primary-color);
  }

  .icon-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 500;
    height: 42px;
    i {
      font-size: var(--icon-size-default);
      color: var(--icon-default, #707070);
    }
  }

  .search-input {
    background: none;
    border: none;
    width: 0;
    transition: width 0.3s;
    &:focus-visible,
    &:active {
      outline: none;
    }
  }

  .search-clear {
    position: absolute;
    right: 5px;
    border-radius: 50%;
    height: 30px;
    width: 30px;
    text-align: center;
    line-height: 30px;
    font-size: var(--icon-size-default);
    color: var(--icon-default, #707070);
    transition: background-color 0.3s;
    &:hover {
      background-color: $grey-4;
    }
  }
}
</style>
