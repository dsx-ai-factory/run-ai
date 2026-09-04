<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="template-section" label="Template" disable-closing>
    <runai-form-card-section
      :loading="loading"
      :main-message="`Select a preconfigured template to launch a new workload quickly`"
      :cards-list="templateCards"
      :selected-cards-ids="selectedTemplateIds"
      :sort-options="{ name: true, creation: true }"
      default-sort-option="creation"
      search-name="templates"
      @selected-card-changed="onSelectedTemplate"
    >
      <template #empty-data-display>
        <q-icon name="ph-bold ph-certificate" size="44px" class="text-icon-default" />
        <p class="empty-data-display-paragraph">
          Looks like you don't have any templates yet... <br />
          Start from scratch and then you can save it as a template.
        </p>
      </template>
    </runai-form-card-section>
    <div class="col-12">
      <q-field class="hidden-field" :model-value="templateId" :rules="[isRequiredTemplate]"></q-field>
    </div>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiFormCardSection } from "@/components/common";

// Models
import type { WorkloadTemplate } from "@/swagger-models/assets-service-client";
import { TCardCmpName, type ICardListItem } from "@/components/common/runai-card-list";

// Utils
import { required } from "@/common/form.validators";

// Constants
import { errorMessages } from "@/common/error-message.constant";
import { START_FROM_SCRATCH_ID } from "@/components/common/runai-form-card-section/runai-form-card-section.vue";

const emptyCard: ICardListItem = {
  id: START_FROM_SCRATCH_ID,
  cardName: TCardCmpName.CREATE_NEW,
  data: {
    meta: {
      id: START_FROM_SCRATCH_ID,
    },
  },
};

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiFormCardSection,
  },
  props: {
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    templates: {
      type: Array as PropType<Array<WorkloadTemplate>>,
      required: true,
    },
    templateId: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
  },
  computed: {
    templateCards(): Array<ICardListItem> {
      const workloadTemplate: Array<ICardListItem> = this.templates.map((workloadTemplate: WorkloadTemplate) => {
        const name = workloadTemplate.meta.name;
        const environment = workloadTemplate.spec.assets.environment?.name || "";
        const compute = workloadTemplate.spec.assets.compute?.name || "";
        const dataSources = workloadTemplate.spec.assets.datasources?.map((ds) => ds.name) || [];
        return {
          id: workloadTemplate.meta.id,
          cardName: TCardCmpName.TEMPLATE,
          searchValues: [name, environment, compute, ...dataSources].filter((val) => !!val),
          data: workloadTemplate,
          sortInfo: {
            name: workloadTemplate.meta.name,
            createdAt: workloadTemplate.meta.createdAt,
          },
        };
      });

      return [emptyCard, ...workloadTemplate];
    },
    selectedTemplateIds(): Array<string> {
      const card: ICardListItem | undefined = this.templateCards.find(
        (card: ICardListItem) => card.id === this.templateId,
      );
      return card ? [card.id] : [];
    },
  },
  methods: {
    onSelectedTemplate(selectedItems: Array<string>): void {
      selectedItems.length === 0
        ? this.$emit("template-changed", null)
        : this.$emit("template-changed", selectedItems[0]);
    },
    isRequiredTemplate(val: string): boolean | string {
      return required(val) || errorMessages.SELECT_TEMPLATE;
    },
  },
});
</script>
<style lang="scss" scoped></style>
