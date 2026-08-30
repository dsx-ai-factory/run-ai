<template>
  <section class="access-key-new column items-center q-pt-md">
    <access-key-creation-form
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
import { AccessKeyCreationForm } from "@/components/credential/access-key/access-key-creation-form";

// models
import { type AccessKey, type AccessKeyCreationRequest, AssetKind } from "@/swagger-models/assets-service-client";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useCredentialStore } from "@/stores/credential.store";

export default defineComponent({
  components: {
    AccessKeyCreationForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      credentialStore: useCredentialStore(),
      credentialModel: null as null | AccessKeyCreationRequest,
      submitting: false as boolean,
    };
  },
  async created() {
    this.credentialModel = credentialService.getEmptyAccessKeyModel();
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
    async onSubmit(submittedAccessKey: AccessKeyCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const readyAccessKeyRequest = this.prepareForServerRequest(submittedAccessKey);
        const savedAccessKey: AccessKey = await credentialService.createAccessKey(readyAccessKeyRequest);
        this.$q.notify(alertUtil.getSuccess(`Credential ${savedAccessKey.meta.name} created`));

        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedAccessKey.meta.id, kind: AssetKind.AccessKey },
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
    prepareForServerRequest(accessKeyRequest: AccessKeyCreationRequest): AccessKeyCreationRequest {
      const accessKey: AccessKeyCreationRequest = { ...accessKeyRequest };
      if (accessKey.spec.existingSecretName) {
        accessKey.spec.accessKeyId = null;
        accessKey.spec.secretAccessKey = null;
      }

      return accessKey;
    },
  },
});
</script>

<style lang="scss" scoped></style>
