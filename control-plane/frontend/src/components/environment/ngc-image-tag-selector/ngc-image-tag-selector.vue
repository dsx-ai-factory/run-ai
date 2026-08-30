<template>
  <div class="ngc-image-tag-selector row items-center">
    <q-select
      aid="ngc-image-select"
      class="col-6 q-mr-md"
      no-error-icon
      use-input
      :input-debounce="300"
      label="Image name"
      :options="imageOptions"
      v-model="selectedImage"
      :loading="loadingImages"
      @filter="onImageFilter"
      @keydown.delete="
        selectedImage = null;
        onImageSelected(null);
      "
      @update:model-value="onImageSelected"
      no-option-text="No images found"
      :rules="[requiredRule]"
    />
    <div class="row items-center col">
      <div class="text-h5 q-mr-md">:</div>
      <q-select
        aid="ngc-tag-select"
        class="col"
        no-error-icon
        v-model="selectedTag"
        :options="filteredTagOptions"
        label="Image tag"
        use-input
        :loading="loadingTags"
        @filter="onTagFilter"
        @keydown.delete="
          selectedTag = null;
          onTagSelected();
        "
        @update:model-value="onTagSelected"
        no-option-text="No tags found"
        :rules="[requiredRule]"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { ISelectOption } from "@/models/global.model";
import { catalogService } from "@/services/control-plane/catalog.service/catalog.service";
import { alertUtil } from "@/utils/alert.util";

export default defineComponent({
  name: "ngc-image-tag-selector",
  emits: ["update:model-value"],
  props: {
    registryId: {
      type: String as PropType<string>,
      required: true,
    },
    credentialId: {
      type: String as PropType<string | undefined>,
      required: false,
      default: undefined,
    },
  },
  data() {
    return {
      selectedImage: null as ISelectOption | null,
      imageOptions: [] as ISelectOption[],
      selectedTag: null as ISelectOption | null,
      tagsByImage: {} as Record<string, ISelectOption[]>,
      loadingTags: false,
      loadingImages: false,
      imageRequestId: 0,
      tagFilter: "",
      registryPrefix: "",
    };
  },
  computed: {
    tagOptions(): ISelectOption[] {
      return this.tagsByImage[(this.selectedImage?.value as string) || ""] || [];
    },
    filteredTagOptions(): ISelectOption[] {
      return this.tagOptions.filter((option) => option.label.toLowerCase().includes(this.tagFilter.toLowerCase()));
    },
  },
  methods: {
    reset(): void {
      this.selectedImage = null;
      this.selectedTag = null;
      this.imageOptions = [];
      this.tagsByImage = {};
      this.registryPrefix = "";
      this.imageRequestId++;
      this.emitValue();
    },
    onImageFilter(val: string, update: (func: { (): void }) => void): void {
      update(() => {
        this.loadImages(val);
      });
    },
    onTagFilter(val: string, update: (func: { (): void }) => void): void {
      update(() => {
        this.tagFilter = val;
      });
    },
    onImageSelected(image: ISelectOption | null): void {
      this.selectedTag = null;
      if (image) {
        this.loadTags(image.value as string);
      }
      this.emitValue();
    },
    onTagSelected(): void {
      this.emitValue();
    },
    async loadImages(search?: string): Promise<void> {
      if (!this.registryId) return;
      const requestId = ++this.imageRequestId;
      try {
        this.loadingImages = true;
        const images = await catalogService.getImages(this.registryId, search, this.credentialId);
        if (requestId !== this.imageRequestId) return; // Prevent stale responses from overwriting the latest results.
        this.imageOptions = images.images.map((image) => {
          return {
            label: image,
            value: image,
          };
        });
        this.registryPrefix = images.prefix || "";
      } catch (error: unknown) {
        if (requestId !== this.imageRequestId) return;
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load images"));
      } finally {
        if (requestId === this.imageRequestId) {
          this.loadingImages = false;
        }
      }
    },
    async loadTags(image: string): Promise<void> {
      if (this.tagsByImage[image]) return; // Prevent duplicate requests.
      try {
        this.loadingTags = true;
        const tags = await catalogService.getImageTags(this.registryId, image, this.credentialId);
        this.tagsByImage[image] = tags.map((tag: string) => {
          return {
            label: tag,
            value: tag,
          };
        });
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load tags"));
      } finally {
        this.loadingTags = false;
      }
    },
    emitValue(): void {
      if (this.selectedImage && this.selectedTag) {
        this.$emit("update:model-value", `${this.registryPrefix}${this.selectedImage.value}:${this.selectedTag.value}`);
      } else {
        this.$emit("update:model-value", null);
      }
    },
    requiredRule(val: string | null): string | boolean | Promise<string | boolean> {
      return !!val || "Select";
    },
  },
  watch: {
    registryId: {
      handler(newVal: string | undefined, oldVal: string | undefined): void {
        if (newVal && newVal !== oldVal) {
          this.reset();
          this.loadImages();
        }
      },
      immediate: true,
    },
    credentialId: {
      handler(newVal: string | undefined, oldVal: string | undefined): void {
        if (newVal !== oldVal) {
          this.reset();
          if (this.registryId) {
            this.loadImages();
          }
        }
      },
    },
  },
});
</script>
