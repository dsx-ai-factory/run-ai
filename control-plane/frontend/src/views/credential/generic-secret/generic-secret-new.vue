<template>
  <section class="column items-center q-pt-md">
    <generic-secret-create-form
      v-if="!isPageLoading && credentialModel"
      :credential-model="credentialModel"
      :submitting="submitting"
      @cancel="onCancel"
      @submit="onSubmit"
      :allowed-scopes="allowedCreateScopes"
      :is-scope-selectable="!$route.query.scope"
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
import { GenericSecretCreateForm } from "@/components/credential/generic-secret/generic-secret-create-form";

// models
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import { AssetKind, type GenericSecretCreationRequest } from "@/swagger-models/assets-service-client";
import type { IScopeModel } from "@/models/global.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useCredentialStore } from "@/stores/credential.store";

// utils
import { assetsUtil } from "@/utils/assets.util/assets.util";

export default defineComponent({
  components: {
    GenericSecretCreateForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      credentialStore: useCredentialStore(),
      credentialModel: null as null | GenericSecretCreationRequest,
      submitting: false as boolean,
    };
  },
  async created() {
    this.credentialModel = credentialService.getEmptyGenericSecretCredentialModel();
    if (this.$route.query.scope) {
      this.credentialModel.meta = assetsUtil.setAssetMetaFromQuery(
        this.credentialModel.meta,
        this.$route.query as unknown as IScopeModel,
        this.authStore.orgUnitList,
      ) as GenericSecretCreationRequest["meta"];
    }

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

    async onSubmit(genericSecretCreationRequest: GenericSecretCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const preparedGenericSecret =
          credentialService.prepareGenericSecretForServerRequest(genericSecretCreationRequest);
        const savedGenericSecret = await credentialService.createGeneralCredential(preparedGenericSecret);
        this.$q.notify(alertUtil.getSuccess(`Credential ${savedGenericSecret.meta.name} created`));

        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedGenericSecret.meta.id, kind: AssetKind.GenericSecret },
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
  },
});
</script>

<style lang="scss" scoped></style>
