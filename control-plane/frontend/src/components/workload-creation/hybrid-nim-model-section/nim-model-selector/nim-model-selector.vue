<template>
  <section class="nim-model-selector">
    <runai-section-title title="Select or type the model name" />

    <div class="row no-wrap items-center">
      <runai-autocomplete-select
        class="nim-model-selector-input col-6"
        aid="nim-autocomplete-select"
        popup-content-class="nim-model-selector-dropdown"
        search-hint="Search or type a model name..."
        label="Model"
        no-result-label=""
        :model-value="modelName"
        @update:model-value="updateModelName"
        :async-options-promise="getNimModels"
        :minimum-characters="0"
        :debounce="300"
        :rules="[isNameEmpty, nimFormatValidator]"
        placeholder="e.g., meta/llama-3.1-8b-base"
        can-add-new-values
        new-values-label-postfix="(+model)"
        :show-add-new-values-option="false"
      />
      <span v-if="showCatalogLink" class="nim-model-selector__link-btn-wrapper q-ml-md">
        <runai-btn
          aid="nim-model-catalog-link"
          round
          flat
          icon="ph-bold ph-arrow-square-out"
          color="grey-7"
          size="12px"
          tooltip="View on NIM model catalog"
          tooltip-anchor="bottom middle"
          tooltip-self="top middle"
          aria-label="View on NIM model catalog"
          @click="openNimModelCatalog"
        />
      </span>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiAutocompleteSelect } from "@/components/common/runai-autocomplete-select";
import { RunaiBtn } from "@/components/common/runai-btn";

import DOCS from "@/constants/docs";
import { openNewWindow } from "@/utils/common.util/common.util";
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty, nimModelNameValidator } from "@/common/form.validators";
import type { ISelectOption } from "@/models/global.model";

import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";

export default defineComponent({
  components: {
    RunaiSectionTitle,
    RunaiAutocompleteSelect,
    RunaiBtn,
  },
  emits: ["nim-model-name-changed", "is-valid"],
  props: {
    modelName: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      staticModelsList: null as ISelectOption[] | null,
      apiModelsList: null as ISelectOption[] | null,
      page: 0 as number,
      pageSize: 500 as number,
    };
  },
  async created() {
    await this.listModelsFromApi();
  },
  computed: {
    showCatalogLink(): boolean {
      return isNotEmpty(this.modelName) && this.modelName.includes("/");
    },
  },
  methods: {
    openNimModelCatalog(): void {
      const parts = this.modelName.split("/");
      const team = parts[0];
      const model = parts[1];
      const url = `${DOCS.EXTERNAL.NIM.CATALOG_BASE_URL}/${team}/containers/${model}`;
      openNewWindow(url);
    },
    updateModelName(selectedModel: string) {
      this.$emit("nim-model-name-changed", selectedModel);
    },
    isNameEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.SELECT_OR_TYPE_A_NAME;
    },
    nimFormatValidator(val: string): boolean | string {
      return nimModelNameValidator(val) || errorMessages.ENTER_A_VALID_NAME;
    },
    async getNimModels(modelName: string): Promise<ISelectOption[]> {
      let nimModelsList = [] as ISelectOption[];

      if (this.staticModelsList) {
        nimModelsList = this.getFromStaticData(this.staticModelsList, modelName);
      } else {
        await this.listModelsFromApi(modelName);
        nimModelsList = this.apiModelsList ?? [];
      }
      return Promise.resolve(nimModelsList);
    },
    getFromStaticData(staticModelsList: ISelectOption[], modelName: string): ISelectOption[] {
      let modelsOptions = [...staticModelsList];
      const needle = modelName.toLowerCase().trim();
      if (needle.length > 0) {
        modelsOptions = staticModelsList.filter((model: ISelectOption) =>
          (model.value as string).toLowerCase().includes(needle),
        );
      }
      return modelsOptions;
    },
    async listModelsFromApi(name = ""): Promise<void> {
      try {
        const catalogRes = await inferenceService.getNimCatalogModels(name, this.page, this.pageSize);
        const models =
          catalogRes.resources?.map((model) => ({
            label: model.resourceId ?? "",
            value: model.resourceId ?? "",
            subLabel: model.description ?? "",
          })) ?? [];

        if (catalogRes.isStaticData) {
          this.staticModelsList = models;
          this.apiModelsList = null;
        } else {
          this.apiModelsList = models;
          this.staticModelsList = null;
        }
      } catch (error: unknown) {
        this.staticModelsList = null;
        this.apiModelsList = null;
        const errorAlert = new ErrorAlert({
          generalMessage: "NIM models couldn't be listed",
        });
        this.$q.notify(errorAlert.getNotification(error));
      }
    },
  },
  watch: {
    modelName: {
      handler(val: string): void {
        const isValid = isNotEmpty(val) && nimModelNameValidator(val);
        this.$emit("is-valid", isValid);
      },
      immediate: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.nim-model-selector__link-btn-wrapper {
  flex-shrink: 0;
  display: inline-flex;
}
</style>
