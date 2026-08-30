<template>
  <div></div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// stores
import { useAuthStore } from "@/stores/auth.store";

// services
import { authService } from "@/services/control-plane/auth.service/auth.service";

// models
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";
import { STORAGE_KEYS } from "@/models/storage.model";
import { HttpErrorResponse } from "@/models/http-response.model";

// utils
import { storageUtil } from "@/utils/storage.util";

export default defineComponent({
  data() {
    return {
      authStore: useAuthStore(),
    };
  },
  async created() {
    try {
      await this.authStore.postLogin();
      if (!this.authStore.accessToken) {
        // No OIDC code in URL (e.g. invite flow redirect without an active session).
        // Kick off a fresh login rather than trying to decode an empty token.
        this.authStore.login();
        return;
      }
      if (!authService.isValidSsoEmailMapperToken(this.authStore.accessToken)) {
        await this.$router.replace({ name: ERROR_ROUTE_NAMES.NO_SSO_EMAIL_MAPPER });
        return;
      }
      await this.redirectToApp();
    } catch (error: unknown) {
      if (error instanceof HttpErrorResponse && error.statusCode === 500) {
        await this.$router.replace({ name: ERROR_ROUTE_NAMES.INTERNAL_SERVER_ERROR });
      } else {
        await this.$router.replace({ name: ERROR_ROUTE_NAMES.BAD_GATEWAY_ERROR });
      }
    }
  },
  methods: {
    async redirectToApp(): Promise<void> {
      const location = storageUtil.get<string>(STORAGE_KEYS.LOCATION);
      if (location) {
        await this.$router.replace({ name: location });
        storageUtil.remove(STORAGE_KEYS.LOCATION);
      } else {
        await this.$router.replace("/");
      }
    },
  },
});
</script>
