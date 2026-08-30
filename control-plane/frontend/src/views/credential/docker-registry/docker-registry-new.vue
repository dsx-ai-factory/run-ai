<template>
  <section class="docker-registry-new column items-center q-pt-md">
    <docker-registry-create-form
      v-if="!pageLoading && credentialModel"
      :credential-model="credentialModel"
      :submitting="submitting"
      @submit="onSubmit"
      @cancel="onCancel"
      :allowed-scopes="allowedCreateScopes"
    ></docker-registry-create-form>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// Stores
import { useAppStore } from "@/stores/app.store";
import { useCredentialStore } from "@/stores/credential.store";
// Components
import { DockerRegistryCreateForm } from "@/components/credential/docker-registry/docker-registry-create-form";
// Services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
// Models
import { HttpErrorResponse } from "@/models/http-response.model";
import {
  type DockerRegistryCreationRequest,
  type DockerRegistry,
  AssetKind,
} from "@/swagger-models/assets-service-client";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";

export default defineComponent({
  components: {
    DockerRegistryCreateForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      credentialStore: useCredentialStore(),
      credentialModel: null as null | DockerRegistryCreationRequest,
      submitting: false as boolean,
      pageLoading: false as boolean,
    };
  },
  async created() {
    this.pageLoading = true;
    this.credentialModel = credentialService.getEmptyDockerRegistryModel();
    this.appStore.setPageLoading(false);
    this.pageLoading = false;
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
    async onSubmit(submittedDockerRegistry: DockerRegistryCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const readyDockerRegistryRequest = this.prepareForServerRequest(submittedDockerRegistry);
        const savedDockerRegistry: DockerRegistry =
          await credentialService.createDockerRegistry(readyDockerRegistryRequest);
        this.$q.notify(alertUtil.getSuccess(`Credential ${savedDockerRegistry.meta.name} created`));

        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedDockerRegistry.meta.id, kind: AssetKind.DockerRegistry },
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
    prepareForServerRequest(dockerRegistryRequest: DockerRegistryCreationRequest): DockerRegistryCreationRequest {
      const dockerRegistry: DockerRegistryCreationRequest = { ...dockerRegistryRequest };
      if (dockerRegistry.spec.existingSecretName) {
        dockerRegistry.spec.user = null;
        dockerRegistry.spec.password = null;
        dockerRegistry.spec.url = null;
      }
      return dockerRegistry;
    },
  },
});
</script>
<style lang="scss" scoped></style>
