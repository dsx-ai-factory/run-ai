<template>
  <section class="ngc-private-registry-box">
    <div class="q-px-md q-py-sm">
      <div v-if="!isEditMode && !createMode" class="row no-wrap justify-between items-center">
        <div class="row no-wrap q-gutter-x-lg items-center">
          <span class="registry-name ellipsis">{{ name }}</span>
          <span class="registry-url ellipsis">{{ url }}</span>
        </div>
        <setting-readonly-input :read-only="readonly">
          <template v-slot:default="{ readonly: ro }">
            <runai-btn
              round
              flat
              size="sm"
              color="black-54"
              class="q-mr-md collapsed-state-icon"
              :aid="`edit-registry-${name}`"
              icon="ph-bold ph-pencil-simple"
              tooltip="Edit registry"
              :disable="ro"
              @click="editClicked"
            />
            <runai-btn
              round
              flat
              size="sm"
              color="black-54"
              class="collapsed-state-icon"
              :aid="`delete-registry-${name}`"
              icon="ph-bold ph-x"
              tooltip="Remove registry"
              :disable="ro"
              @click="removedClicked"
            />
          </template>
        </setting-readonly-input>
      </div>
      <q-form v-else class="row justify-between items-center q-gutter-x-lg q-pb-md" @submit="saveClicked" ref="formRef">
        <q-input class="col-2" v-model="localName" no-error-icon label="Display name" :rules="[isNotEmpty]" />
        <q-input
          class="col"
          no-error-icon
          hint="Enter the unique identifier for your org"
          v-model="orgName"
          label="Organization’s unique identifier"
          :rules="[isNotEmptyOrgName]"
          :prefix="ngcUrlPrefix"
          :suffix="ngcUrlSuffix"
        />
      </q-form>
    </div>
    <div v-if="isEditMode || createMode" class="footer row justify-end items-center q-px-md q-py-sm">
      <q-btn label="Cancel" flat color="primary" @click="cancelClicked" />
      <q-btn label="Save" flat color="primary" @click="saveClicked" :loading="loading" />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Utils
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
// Models
import type { QForm } from "quasar";
// Components
import { RunaiBtn } from "@/components/common/runai-btn";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";

export default defineComponent({
  name: "ngc-private-registry-box",
  components: { RunaiBtn, SettingReadonlyInput },
  emits: ["remove-registry", "update-registry", "cancel"],
  props: {
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    readonly: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    name: {
      type: String as PropType<string>,
      required: true,
    },
    url: {
      type: String as PropType<string>,
      required: true,
    },
    createMode: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      isEditMode: false,
      localName: "",
      localUrl: "",
      ngcUrlPrefix: "https://api.ngc.nvidia.com/v2/search/org/",
      ngcUrlSuffix: "/resources/CONTAINER",
    };
  },
  created() {
    this.localName = this.name;
    this.localUrl = this.url;
  },
  computed: {
    orgName: {
      get() {
        if (this.localUrl && this.localUrl.startsWith(this.ngcUrlPrefix) && this.localUrl.includes(this.ngcUrlSuffix)) {
          return this.localUrl.substring(this.ngcUrlPrefix.length, this.localUrl.indexOf(this.ngcUrlSuffix));
        }
        return "";
      },
      set(val: string): void {
        this.localUrl = `${this.ngcUrlPrefix}${val}${this.ngcUrlSuffix}`;
      },
    },
  },
  methods: {
    editClicked(): void {
      this.isEditMode = true;
      this.localName = this.name;
      this.localUrl = this.url;
    },
    removedClicked(): void {
      this.$emit("remove-registry");
    },
    async saveClicked(): Promise<void> {
      const isValid = await (this.$refs.formRef as QForm)?.validate();
      if (!isValid) return;

      this.$emit("update-registry", { name: this.localName, url: this.localUrl });
      this.isEditMode = false;
    },
    cancelClicked(): void {
      this.isEditMode = false;
      this.$emit("cancel");
    },
    isNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    isNotEmptyOrgName(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.ENTRER_IDENTIFIER;
    },
  },
});
</script>

<style lang="scss" scoped>
.ngc-private-registry-box {
  border: 1px solid $black-12;

  .footer {
    border-top: 1px solid $black-12;
  }

  .registry-name {
    width: 100px;
  }

  .registry-url {
    width: 530px;
  }

  .collapsed-state-icon {
    :deep(.q-icon) {
      font-size: 20px !important;
    }
  }
}
</style>
