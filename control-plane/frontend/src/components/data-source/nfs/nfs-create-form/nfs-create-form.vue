<template>
  <section class="nfs-create-form column items-center q-pt-md">
    <runai-form-wrapper ref="elNfsForm" class="nfs-form" :form-state="nfsModel" :enter-callback="save">
      <runai-expansion-wrapper>
        <data-source-type-section data-source-kind="nfs" />
        <runai-scope-section
          kind-of-asset="data source"
          :read-only="!isScopeSelectable"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
          :allowed-scopes="allowedScopes"
        />
        <runai-name-section
          label="Data source"
          asset-kind="nfs"
          v-model="nfsModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="nfsModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <nfs-data-mount-section v-if="isScopeSelected" v-model="nfsModel.spec" />
        </transition>
      </runai-expansion-wrapper>

      <section class="row items-center q-mt-md">
        <q-field class="col-4 form-hint no-padding" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
        <div class="buttons q-ml-auto">
          <runai-btn aid="cancel-btn" class="q-mr-md" @click="onCancel" flat color="primary" label="cancel" />
          <runai-btn
            aid="submit-btn"
            color="primary"
            @click="save"
            :loading="validating || submitting"
            label="create data source"
            type="submit"
          />
        </div>
      </section>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { NfsDataMountSection } from "@/components/data-source/nfs/nfs-data-mount-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// models
import type { IUINFSCreationRequest } from "@/models/data-source.model";
import type { IScopeModel } from "@/models/global.model";
// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { DataSourceTypeSection } from "../../sections/data-source-type-section";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    RunaiNameSection,
    NfsDataMountSection,
    RunaiScopeSection,
    DataSourceTypeSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    dataSourceModel: {
      type: Object as PropType<IUINFSCreationRequest>,
      required: true,
    },
    isScopeSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
  },
  data() {
    return {
      nfsForm: null as HTMLFormElement | null,
      nfsModel: deepCopy(this.dataSourceModel) as IUINFSCreationRequest,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      selectedCluster: "" as string,
      validating: false as boolean,
    };
  },
  mounted() {
    this.nfsForm = this.$refs["elNfsForm"] as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.nfsModel.meta.scope,
        projectId: this.nfsModel.meta.projectId,
        departmentId: this.nfsModel.meta.departmentId,
        clusterId: this.nfsModel.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.nfsModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.nfsModel.meta = scopeUtil.updateScope(this.nfsModel.meta, updatedScope);
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      this.displayFormHint = false;
      try {
        this.validating = true;
        const success = await this.validate();
        if (!success) {
          this.showHint();
          return;
        }
        this.$emit("submit", this.nfsModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.nfsForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.FORM_INCOMPLETE;
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>

<style lang="scss" scoped></style>
