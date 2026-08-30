<template>
  <div class="runai-card-list">
    <ul class="list-container" aid="skeleton-list" v-if="loading">
      <li v-for="num in list.length > 3 ? 6 : 3" :key="num">
        <skeleton-card />
      </li>
    </ul>
    <template v-else>
      <q-carousel :model-value="slide" animated control-color="purple" height="fit-content" aid="list-content">
        <q-carousel-slide v-for="(row, index) in listForSlides" :name="index" class="no-padding" :key="index">
          <ul class="list-container">
            <template v-for="item in row" :key="item.id">
              <li v-if="item">
                <runai-tooltip-wrapper
                  anchor="bottom middle"
                  self="top middle"
                  :display-tooltip="!!item.tooltip"
                  :tooltip-text="item.tooltip"
                  max-width="350px"
                >
                  <component
                    :is="item.cardName"
                    :data="item.data"
                    :disabled="item.disabled || disabled"
                    :show-disabled-info="item.showDisabledInfo"
                    @card-clicked="cardClicked(item)"
                    :is-selected="isSelected(item.id)"
                  ></component>
                </runai-tooltip-wrapper>
              </li>
            </template>
            <section class="empty-data-section q-ml-md" v-if="showEmptyDisplay">
              <slot name="empty-data-display" />
            </section>
          </ul>
        </q-carousel-slide>
      </q-carousel>
    </template>
    <footer class="row q-mt-md" v-if="isMultiplePages || displaySelectCounter">
      <div class="selected-container col">
        <template v-if="displaySelectCounter">Selected {{ selectedCount }} / {{ list.length }}</template>
      </div>

      <div class="pagination-container col row justify-center">
        <template v-if="isMultiplePages">
          <q-pagination
            :disable="loading"
            aid="pagination-btns-container"
            :model-value="page"
            @update:model-value="updateSlide"
            color="grey-6"
            text-color="grey-10"
            active-text-color="white"
            :min="1"
            :max="maxAmountOfTotalPages"
            :max-pages="maxAmountOfPageButtonsToRender"
            boundary-numbers
          />
        </template>
      </div>
      <div class="col"></div>
    </footer>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiSelectCard } from "@/components/common/runai-select-card";
import { ProjectCard } from "./cards/project-card";
import { TemplateCard } from "./cards/template-card";
import { EnvironmentCard } from "./cards/environment-card";
import { ComputeResourceCard } from "./cards/compute-resource-card";
import { DataSourceCard } from "./cards/data-source-card";
import { SkeletonCard } from "./cards/skeleton-card";
import { CreateNewCard } from "./cards/create-new-card";
import { CredentialCard } from "./cards/credential-card";
import { K8sDistributionCard } from "./cards/k8s-distribution-card";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { TemplateCardV2 } from "./cards/template-card-v2";

// models
import type { ICardQuery } from "./runai-card-list.model";
import { TCardCmpName, type ICardListItem } from "./runai-card-list.model";

// utils
import { getEmptyArray, isNullOrUndefined } from "@/utils/common.util";
import { ComplianceInfoModal } from "@/components/policy/compliance-info-modal";
import type { ComplianceInfo } from "@/swagger-models/assets-service-client";

export default defineComponent({
  name: "runai-card-list",
  components: {
    ProjectCard,
    TemplateCard,
    RunaiSelectCard,
    SkeletonCard,
    CreateNewCard,
    EnvironmentCard,
    ComputeResourceCard,
    DataSourceCard,
    K8sDistributionCard,
    RunaiTooltipWrapper,
    ComplianceInfoModal,
    CredentialCard,
    TemplateCardV2,
  },
  emits: ["changed", "get-next-batch"],
  props: {
    selectedItems: {
      type: Array as PropType<Array<string>>,
      required: true,
    },
    list: {
      type: Array as PropType<Array<ICardListItem>>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    multiSelect: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    // also used as indicator for server side pagination
    listQuery: {
      type: Object as PropType<ICardQuery>,
      required: false,
    },
    totalCount: {
      type: Number as PropType<number>,
      required: false,
    },
  },
  data() {
    return {
      slide: 0 as number,
      page: 1 as number,
      slideItemsMap: {} as Record<string, number>, // <id, slide number>
      listForSlides: [] as Array<Array<ICardListItem>>,
      policyComplianceModalOpen: false as boolean,
      selectedComplianceInfo: null as null | ComplianceInfo,
    };
  },
  computed: {
    selectedCount(): number {
      return this.selectedItems.length;
    },
    showEmptyDisplay(): boolean {
      return this.list.length === 0 || (this.list.length === 1 && this.list[0].cardName === TCardCmpName.CREATE_NEW);
    },
    displaySelectCounter(): boolean {
      return this.multiSelect;
    },
    maxAmountOfPageButtonsToRender(): number {
      const relevantLength =
        this.totalCount && !isNullOrUndefined(this.totalCount) ? this.totalCount : this.listForSlides.length;
      return relevantLength > 100 ? 5 : 6;
    },
    maxAmountOfTotalPages(): number {
      if (this.totalCount && !isNullOrUndefined(this.totalCount)) {
        const countModifier = this.isEmptyCardPresent ? 1 : 0;
        return Math.ceil((this.totalCount + countModifier) / 6);
      }

      return this.listForSlides.length;
    },
    isEmptyCardPresent(): boolean {
      return !!this.list.find((i) => i.id === "-1");
    },
    isMultiplePages(): boolean {
      // if emptycard exists we need to add 1 to the total count
      if (this.totalCount && !isNullOrUndefined(this.totalCount)) {
        const countModifier = this.isEmptyCardPresent ? 1 : 0;
        return this.totalCount + countModifier > 6;
      }

      return this.listForSlides.length > 1;
    },
    queryPropertiesForWatcher(): (string | number | null | undefined)[] {
      return [this.listQuery?.filterBy, this.listQuery?.search, this.listQuery?.sortBy];
    },
  },
  methods: {
    initListForSlides(): void {
      const cardsAmountInSlide = 6;
      const countModifierForTotalCount = this.isEmptyCardPresent ? 1 : 0;
      const amountOfCardsToRender =
        this.totalCount && !isNullOrUndefined(this.totalCount)
          ? this.totalCount + countModifierForTotalCount
          : this.list.length;

      const slidesAmount: number = Math.ceil(amountOfCardsToRender / cardsAmountInSlide);

      const slides: Array<undefined> = getEmptyArray(slidesAmount); // getting the slides array
      this.listForSlides = slides.map(
        (
          currentSlide: undefined,
          idx: number, // on each slide need to fill the cards
        ) => {
          const cards: Array<ICardListItem> = this.getCards(idx * cardsAmountInSlide, cardsAmountInSlide);
          cards.forEach((card: ICardListItem) => (this.slideItemsMap[card.id] = idx)); // <id, slide number>
          return cards;
        },
      ); // fill the cards in current slide (each row is an array of cards)
    },
    updateDisplaySlide(): void {
      // for model with type array we should find the fist slide
      // that contains selected card

      if (this.listQuery) return; // no clue in which batch the once selected card is, so we skip this logic if its a server side pagination

      if (this.listForSlides.length !== 0 && this.selectedCount !== 0 && this.page === 1) {
        let tempSlide: number = this.list.length + 1; // slide can't be more then the full array count
        this.selectedItems.forEach((itemId: string) => {
          const currentSlide: number | undefined = this.slideItemsMap[itemId]; // <id, slide number>
          if (currentSlide === undefined) return;

          tempSlide = Math.min(tempSlide, currentSlide);
        });

        this.slide = tempSlide;
        this.page = this.slide + 1;
      }
    },
    isSelected(id: string): boolean {
      return this.selectedItems.some((itemId: string) => itemId === id);
    },
    getCards(startIdx: number, cardsCount: number): Array<ICardListItem> {
      return this.list.slice(startIdx, startIdx + cardsCount);
    },
    cardClicked(item: ICardListItem): void {
      if (this.multiSelect) {
        this.handleMultiSelect(item);
      } else {
        this.handleSingleSelect(item);
      }
    },
    handleSingleSelect(item: ICardListItem): void {
      let newModel: Array<string> = [];
      if (this.selectedItems.length === 0) {
        newModel = [item.id];
      } else if (this.selectedItems[0] === item.id) {
        newModel = [];
      } else {
        newModel = [item.id];
      }

      this.$emit("changed", newModel);
    },
    handleMultiSelect(item: ICardListItem): void {
      const idx = this.selectedItems.findIndex((id: string) => id === item.id);
      if (idx === -1) {
        this.$emit("changed", [...this.selectedItems, item.id]);
      } else {
        this.$emit(
          "changed",
          this.selectedItems.filter((id: string) => id !== item.id),
        );
      }
    },
    updateSlide(page: number): void {
      // if its server side pagination, we prevent updating the slide during loading
      if (this.loading && this.listQuery) return;

      const requiredSlidesAmount = page * 6;
      const relevantLength = this.isEmptyCardPresent ? this.list.length - 1 : this.list.length;
      if (
        this.totalCount &&
        !isNullOrUndefined(this.totalCount) &&
        requiredSlidesAmount >= relevantLength &&
        relevantLength < this.totalCount
      ) {
        const nextNaturalBatchAmount = relevantLength + 6;
        const isWithinNextBatch = nextNaturalBatchAmount >= requiredSlidesAmount;
        this.$emit("get-next-batch", isWithinNextBatch ? undefined : requiredSlidesAmount);
      }

      this.slide = page - 1;
      this.page = page;
    },
  },
  watch: {
    list: {
      handler(): void {
        this.initListForSlides();
        this.updateDisplaySlide();
      },
      immediate: true,
    },
    loading(newVal: boolean): void {
      !newVal && this.updateDisplaySlide();
    },
    queryPropertiesForWatcher: {
      handler(newVal: (string | number | null | undefined)[], oldVal: (string | number | null | undefined)[]): void {
        if (!newVal || !oldVal) return;
        if (newVal.some((val, idx) => val !== oldVal[idx])) {
          this.slide = 0;
          this.page = 1;
          this.initListForSlides();
          this.updateDisplaySlide();
        }
      },
      immediate: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-card-list {
  .list-container {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 5px;

    .empty-data-section {
      display: flex;
      flex-direction: column;
      justify-content: center;
      gap: 15px;
    }
  }
}
</style>
