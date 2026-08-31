<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="s3-create-form column items-center q-pt-md">
    <runai-form-wrapper ref="elS3Form" class="s3-form" :form-state="s3Model" :enter-callback="save">
      <runai-expansion-wrapper>
        <data-source-type-section data-source-kind="s3" />

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
          asset-kind="s3"
          v-model="s3Model.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="s3Model.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <s3-data-mount-section
            v-if="isScopeSelected"
            v-model="s3Model.spec"
            :credentials="credentials"
            :loading-credentials="loadingCredentials"
          />
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
import { S3DataMountSection } from "@/components/data-source/s3/s3-data-mount-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// models
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { AccessKey } from "@/swagger-models/assets-service-client";
import type { IUIS3CreationRequest } from "@/models/data-source.model";
import type { IScopeCredentialsList } from "@/models/credential.model";

// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { alertUtil } from "@/utils/alert.util";

// stores
import { useDataSourceStore } from "@/stores/data-source.store";
import type { IScopeModel } from "@/models/global.model";
import { DataSourceTypeSection } from "../../sections/data-source-type-section";
import { errorMessages } from "@/common/error-message.constant";
import { dataSourceUtil } from "@/utils/data-source.util";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    S3DataMountSection,
    RunaiScopeSection,
    DataSourceTypeSection,
    RunaiNameSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    dataSourceModel: {
      type: Object as PropType<IUIS3CreationRequest>,
      required: true,
    },
    isScopeSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: true,
    },
  },
  data() {
    return {
      dataSourceStore: useDataSourceStore(),
      s3Form: null as HTMLFormElement | null,
      s3Model: deepCopy(this.dataSourceModel) as IUIS3CreationRequest,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      credentials: [] as AccessKey[],
      loadingCredentials: false as boolean,
      selectedCluster: "" as string,
      validating: false as boolean,
    };
  },
  async created() {
    this.loadCredentials();
  },
  mounted() {
    this.s3Form = this.$refs["elS3Form"] as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.s3Model.meta.scope,
        projectId: this.s3Model.meta.projectId,
        departmentId: this.s3Model.meta.departmentId,
        clusterId: this.s3Model.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.s3Model.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.s3Model.meta = scopeUtil.updateScope(this.s3Model.meta, updatedScope);
      this.loadCredentials();
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      try {
        this.validating = true;
        this.displayFormHint = false;
        const success = await this.validate();
        if (!success) {
          this.showHint();
          return;
        }
        this.$emit("submit", this.s3Model);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.s3Form as HTMLFormElement).validate();
      return isValid;
    },
    async loadCredentials(): Promise<void> {
      if (!this.scopeModel.scope) return;
      try {
        this.loadingCredentials = true;
        const listScope: IScopeCredentialsList = dataSourceUtil.getListScopeByMeta(this.s3Model.meta);
        this.credentials = await dataSourceService.listAccessKeysByScope(listScope);
        const selectedSecret = this.s3Model.spec.accessKeyAssetId;
        if (!!selectedSecret && !this.credentials.map((credential) => credential.meta.id).includes(selectedSecret))
          this.s3Model.spec.accessKeyAssetId = "";
      } catch (err: unknown) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load credentials"));
      } finally {
        this.loadingCredentials = false;
      }
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
