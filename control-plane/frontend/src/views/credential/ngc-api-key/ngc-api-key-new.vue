<template>
  <section class="ngc-api-key-new column items-center q-pt-md">
    <ngc-api-key-create-form
      v-if="credentialModel"
      :credential-model="credentialModel"
      :submitting="submitting"
      @submit="onSubmit"
      @cancel="onCancel"
      :allowed-scopes="allowedCreateScopes"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// Stores
import { useAppStore } from "@/stores/app.store";
import { useCredentialStore } from "@/stores/credential.store";
// Components
import { NgcApiKeyCreateForm } from "@/components/credential/ngc-api-key/ngc-api-key-create-form";
// Services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
// Models
import { HttpErrorResponse } from "@/models/http-response.model";
import { AssetKind, type NgcApiKey, type NgcApiKeyCreationRequest } from "@/swagger-models/assets-service-client";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "ngc-api-key-new",
  components: {
    NgcApiKeyCreateForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      credentialStore: useCredentialStore(),
      credentialModel: credentialService.getEmptyNgcApiKeyModel() as NgcApiKeyCreationRequest,
      submitting: false as boolean,
    };
  },
  async created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    allowedCreateScopes(): PermittedScopes {
      return this.credentialStore.actionPermissionsByScopes[Action.Create];
    },
  },
  methods: {
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      allowToLeave && this.$router.push({ name: this.$route.meta.backPageName as string });
    },
    async onSubmit(submittedNgcApiKey: NgcApiKeyCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const readyNgcApiKeyRequest = { ...submittedNgcApiKey };
        const savedNgcApiKey: NgcApiKey = await credentialService.createNgcApiKey(readyNgcApiKeyRequest);
        this.$q.notify(alertUtil.getSuccess(`Credential ${savedNgcApiKey.meta.name} created`));

        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedNgcApiKey.meta.id, kind: AssetKind.NgcApiKey },
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
