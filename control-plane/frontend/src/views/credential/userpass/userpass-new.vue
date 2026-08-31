<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="userpass-new column items-center q-pt-md">
    <userpass-creation-form
      v-if="!isPageLoading && credentialModel"
      :credential-model="credentialModel"
      :submitting="submitting"
      @cancel="onCancel"
      @submit="onSubmit"
      :allowed-scopes="allowedCreateScopes"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { alertUtil } from "@/utils/alert.util";

// cmps
import { UserpassCreationForm } from "@/components/credential/userpass/userpass-creation-form";

// models
import { type Password, type PasswordCreationRequest, AssetKind } from "@/swagger-models/assets-service-client";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useCredentialStore } from "@/stores/credential.store";

export default defineComponent({
  components: {
    UserpassCreationForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      credentialStore: useCredentialStore(),
      credentialModel: null as null | PasswordCreationRequest,
      submitting: false as boolean,
    };
  },
  async created() {
    this.credentialModel = credentialService.getEmptyUserpassModel();
    this.appStore.setPageLoading(false);
  },
  computed: {
    isPageLoading(): boolean {
      return this.appStore.isPageLoading;
    },
    allowedCreateScopes(): PermittedScopes {
      return this.credentialStore.actionPermissionsByScopes[Action.Create];
    },
  },
  methods: {
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      allowToLeave && this.$router.push({ name: this.$route.meta.backPageName as string });
    },
    async onSubmit(submittedUserpass: PasswordCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const readyUserpassRequest = this.prepareForServerRequest(submittedUserpass);
        const savedUserpass: Password = await credentialService.createUserpass(readyUserpassRequest);
        this.$q.notify(alertUtil.getSuccess(`Credential ${savedUserpass.meta.name} created`));

        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedUserpass.meta.id, kind: AssetKind.Password },
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create credential"));
        }
      } finally {
        this.submitting = false;
      }
    },
    prepareForServerRequest(userpassRequest: PasswordCreationRequest): PasswordCreationRequest {
      const userpass: PasswordCreationRequest = { ...userpassRequest };
      if (userpass.spec.existingSecretName) {
        userpass.spec.user = null;
        userpass.spec.password = null;
      }
      return userpass;
    },
  },
});
</script>

<style lang="scss" scoped></style>
