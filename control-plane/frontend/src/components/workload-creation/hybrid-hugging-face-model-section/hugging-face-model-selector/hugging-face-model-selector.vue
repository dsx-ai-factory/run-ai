<template>
  <section class="hf-model-selector">
    <!--  -->
    <runai-section-title title="">
      <template #title-content>
        <div>Select or type a <b>language model</b> name as displayed in Hugging Face</div>
      </template>
    </runai-section-title>

    <div class="row no-wrap items-center">
      <runai-autocomplete-select
        class="hf-model-selector-input col-6"
        aid="hf-autocomplete-select"
        :ref="autocompleteSelectRef"
        search-hint="Search or type a model name..."
        label="Model name"
        no-result-label="No model name match your filter"
        :model-value="selectedModel"
        @update:model-value="onModelNameChanged"
        :async-options-promise="getHfModels"
        :minimum-characters="0"
        placeholder="e.g., meta/llama-3.1-8b-base"
        :show-add-new-values-option="false"
        :rules="[validateModel]"
        :disable="disable"
      >
        <template v-slot:option="scope">
          <q-item v-bind="scope.itemProps" @click="onModelNameChanged(scope.opt.value)">
            <q-item-section>
              <div class="row items-center justify-between full-width">
                <div class="row items-center">
                  <span class="hf-model-name">{{ scope.opt.label }}</span>
                  <span class="q-ml-sm hf-model-trending" v-if="modelsMap[scope.opt.value]?.trending">
                    <runai-svg-icon size="14px" name="fire" />
                  </span>
                  <div v-if="modelsMap[scope.opt.value]?.formattedDownloads" class="row items-center q-ml-md">
                    <q-icon name="ph-bold ph-arrow-line-down" size="16px" class="text-grey-6" />
                    <span class="text-grey-7">{{ modelsMap[scope.opt.value].formattedDownloads }}</span>
                  </div>
                </div>
                <q-badge
                  v-if="modelsMap[scope.opt.value]?.gated"
                  outline
                  color="grey-7"
                  label="Gated"
                  class="hf-gated-badge"
                />
              </div>
            </q-item-section>
          </q-item>
        </template>
      </runai-autocomplete-select>
      <span v-if="selectedModel" class="hf-model-selector__link-btn-wrapper q-ml-md">
        <runai-btn
          aid="hf-model-huggingface-link"
          round
          flat
          icon="ph-bold ph-arrow-square-out"
          color="grey-7"
          size="12px"
          tooltip="View on Hugging Face"
          tooltip-anchor="bottom middle"
          tooltip-self="top middle"
          aria-label="View on Hugging Face"
          @click="openHuggingFaceModel"
        />
      </span>
      <div class="col q-ml-md" style="min-width: 0">
        <notification-alert-bar v-if="warningMessage" :variant="ENotificationBarVariant.Warning">
          <span>
            This isn't identified as a language model.
            <br />
            Non-language models won't run.
          </span>
        </notification-alert-bar>
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiAutocompleteSelect } from "@/components/common/runai-autocomplete-select";
import { RunaiBtn } from "@/components/common/runai-btn";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

import DOCS from "@/constants/docs";
import { openNewWindow } from "@/utils/common.util/common.util";

// models
import type { HuggingFaceModelInfo } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import {
  BACKEND_MESSAGES,
  EHuggingFaceModelEvents,
  type IHuggingFaceModelSectionEvent,
  type IHuggingFaceModelValidationRes,
} from "./hugging-face-model-selector.model";
import type { HuggingFaceModelItem } from "@/swagger-models/catalog-service-client";

// services
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { catalogService } from "@/services/control-plane/catalog.service/catalog.service";

// constants
import { errorMessages } from "@/common/error-message.constant";

// others
import { isEmpty, huggingFaceModelNameValid } from "@/common/form.validators";
import type { QInput } from "quasar";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";
import { formatUtil } from "@/utils/format.util/format.util";

export default defineComponent({
  name: "hugging-face-model-selector",
  components: {
    RunaiSectionTitle,
    RunaiAutocompleteSelect,
    RunaiBtn,
    NotificationAlertBar,
    RunaiSvgIcon,
  },
  emits: {
    [EHuggingFaceModelEvents.UpdateModelValue]: (huggingFaceModelInfo: HuggingFaceModelInfo) => huggingFaceModelInfo,
    [EHuggingFaceModelEvents.IsValid]: (validationEvent: IHuggingFaceModelSectionEvent) => validationEvent,
  },
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      selectedModel: this.modelValue as string,
      autocompleteSelectRef: "hf-autocomplete-select-ref" as string,
      ENotificationBarVariant: ENotificationBarVariant,
      warningMessage: "" as string,
      modelsMap: {} as Record<string, HuggingFaceModelItem & { formattedDownloads?: string }>,
    };
  },
  mounted() {
    if (this.modelValue) {
      if (this.disable) {
        this.validateModel(this.modelValue);
      } else {
        (this.$refs[this.autocompleteSelectRef] as QInput)?.validate();
      }
    }
  },
  methods: {
    openHuggingFaceModel(): void {
      const path = this.selectedModel.split("/").map(encodeURIComponent).join("/");
      const url = `${DOCS.EXTERNAL.HUGGING_FACE.BASE_URL}/${path}`;
      openNewWindow(url);
    },
    onModelNameChanged(selectedModel: string | null | ISelectOption): void {
      if (selectedModel && typeof selectedModel === "object") {
        this.selectedModel = selectedModel.value as string;
      } else if (selectedModel && typeof selectedModel === "string") {
        this.selectedModel = selectedModel as string;
      } else {
        this.selectedModel = "";
      }
      this.warningMessage = "";
      this.invokeModelChanged({
        modelName: this.selectedModel,
      });
    },
    async getHfModels(modelName: string): Promise<ISelectOption[]> {
      try {
        const catalogModels = await catalogService.searchHuggingFaceCatalog(modelName);

        this.buildModelsMap(catalogModels);

        return this.createSelectOptions(catalogModels);
      } catch (error: unknown) {
        this.handleFetchError(error);
        return [];
      }
    },
    buildModelsMap(models: HuggingFaceModelItem[]): void {
      this.modelsMap = models.reduce(
        (acc, model) => {
          if (model.id) {
            acc[model.id] = {
              ...model,
              formattedDownloads: formatUtil.formatCompactNumber(model.downloads),
            };
          }
          return acc;
        },
        {} as Record<string, HuggingFaceModelItem & { formattedDownloads?: string }>,
      );
    },
    createSelectOptions(models: HuggingFaceModelItem[]): ISelectOption[] {
      return models.map((model) => ({
        label: model.id ?? "",
        value: model.id ?? "",
      }));
    },
    handleFetchError(error: unknown): void {
      this.modelsMap = {};
      const errorAlert = new ErrorAlert({
        generalMessage: "Hugging Face models couldn't be listed",
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
    invokeModelChanged(huggingFaceModelInfo: HuggingFaceModelInfo): void {
      this.$emit(EHuggingFaceModelEvents.UpdateModelValue, huggingFaceModelInfo);
    },
    invokeValidationChanged(valid: boolean, gated = false): void {
      this.$emit(EHuggingFaceModelEvents.IsValid, { valid, gated });
    },
    async validateModel(selectedModel: string): Promise<string | boolean> {
      const validationResults: IHuggingFaceModelValidationRes = await this.verifyModelName(selectedModel.trim());
      if (validationResults.errorMessage) {
        this.invokeValidationChanged(false, validationResults.gated);
      } else {
        this.invokeValidationChanged(true, validationResults.gated);
      }

      if (validationResults.warningMessage) {
        this.warningMessage = validationResults.warningMessage;
      }
      return validationResults.errorMessage || true;
    },
    async verifyModelName(modelName: string): Promise<IHuggingFaceModelValidationRes> {
      const modelValidationRes: IHuggingFaceModelValidationRes = {
        modelName,
      };
      try {
        if (isEmpty(modelName)) {
          modelValidationRes.errorMessage = errorMessages.SELECT_OR_TYPE_A_NAME;
        } else if (!huggingFaceModelNameValid(modelName)) {
          modelValidationRes.errorMessage = errorMessages.ENTER_A_VALID_NAME;
        } else {
          const huggingFaceModelInfo: HuggingFaceModelInfo = await inferenceService.getHuggingFaceModelInfo({
            modelName,
          });

          if (!huggingFaceModelInfo.language) {
            modelValidationRes.warningMessage = errorMessages.MODEL_IS_NOT_A_VLLM;
          }

          modelValidationRes.gated = huggingFaceModelInfo.gated;
        }
      } catch (error: unknown) {
        if (
          error instanceof HttpErrorResponse &&
          error.message === BACKEND_MESSAGES.MODEL_WAS_NOT_FOUND_ON_HUGGING_FACE
        ) {
          modelValidationRes.errorMessage = errorMessages.HUGGING_FACE_MODEL_NOT_FOUND;
        } else {
          modelValidationRes.errorMessage = errorMessages.FAILED_TO_GET_MODEL_INFO;
        }
      }

      return modelValidationRes;
    },
  },
});
</script>

<style lang="scss" scoped>
.hf-model-selector__link-btn-wrapper {
  flex-shrink: 0;
  display: inline-flex;
}
</style>
