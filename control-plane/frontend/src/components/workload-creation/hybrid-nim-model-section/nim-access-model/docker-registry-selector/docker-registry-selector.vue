<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <image-pull-credential-layout
    class="docker-registry-selector"
    section-title="Set the credentials for pulling the image"
    :show-info-banner="showInfoBox"
  >
    <div class="row">
      <span class="col-8">
        <runai-select
          aid="docker-registry-selector"
          :model-value="modelValue.dockerRegistryName"
          :options="credentials"
          @update:model-value="onCredentialChanged"
          label="Credential name"
          no-option-text="No credentials found"
          stack-label
          placeholder="Select..."
        />
      </span>
    </div>
  </image-pull-credential-layout>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiSelect } from "@/components/common/runai-select";
import { ImagePullCredentialLayout } from "../image-pull-credential-layout";

// models
import type { IScopeModel, ISelectOption } from "@/models/global.model";
import type { IAssetsFilter } from "@/models/filter.model";
import { AssetSyncStatusEnum, type DockerRegistry } from "@/swagger-models/assets-service-client";
import { EDockerRegistrySelectorEvents, type IDockerRegistrySelectorModel } from "./docker-registry-selector.model";

// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";

// others
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";

// stores
import { useClusterStore } from "@/stores/cluster.store";

export default defineComponent({
  name: "docker-registry-selector",
  components: {
    RunaiSelect,
    ImagePullCredentialLayout,
  },
  emits: {
    [EDockerRegistrySelectorEvents.UpdateModelValue]: (modelValue: IDockerRegistrySelectorModel) => modelValue,
  },
  props: {
    modelValue: {
      type: Object as PropType<IDockerRegistrySelectorModel>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    isRequired: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      credentials: [] as Array<ISelectOption>,
      loading: false as boolean,
    };
  },
  async created() {
    await this.loadCredentials();
  },
  computed: {
    showInfoBox(): boolean {
      return !this.modelValue.dockerRegistryName;
    },
  },
  methods: {
    async getDockerRegistries(filterBy: IAssetsFilter): Promise<Array<ISelectOption>> {
      const credentials: DockerRegistry[] = await credentialService.listDockerRegistries(filterBy);

      return credentials.map((cred: DockerRegistry) => {
        const disableByStatus =
          !cred.status?.status || cred.status?.status === AssetSyncStatusEnum.Creating ? true : false;

        return {
          value: cred.meta.name,
          label: cred.meta.name,
          disable: disableByStatus,
          disabledTooltip: disableByStatus
            ? "This setup is being created. If it isn't created after a few seconds, contact your administrator."
            : undefined,
        };
      });
    },

    getFilterBy(): IAssetsFilter {
      return {
        projectId: this.scopeModel.projectId,
        clusterUuid: this.scopeModel.clusterId || undefined,
        departmentId: this.scopeModel.departmentId,
        privateOnly: true,
        statusInfo: true,
      };
    },
    async loadCredentials(): Promise<void> {
      try {
        this.loading = true;

        const filterBy = this.getFilterBy();
        this.credentials = await this.getDockerRegistries(filterBy);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("docker registries"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.loading = false;
      }
    },
    emitChanges(modelValue: IDockerRegistrySelectorModel): void {
      this.$emit(EDockerRegistrySelectorEvents.UpdateModelValue, modelValue);
    },
    onCredentialChanged(dockerRegistry: ISelectOption): void {
      this.emitChanges({
        dockerRegistryName: dockerRegistry.label,
      });
    },
  },
});
</script>
