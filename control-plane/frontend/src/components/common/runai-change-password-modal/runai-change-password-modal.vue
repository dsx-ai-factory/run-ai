<template>
  <runai-base-dialog :model-value="true" @close="$emit('close')" size="width-sm" :no-padding="true">
    <template #header>Change Password</template>

    <template #body>
      <q-form ref="formRef" class="change-password-form" greedy>
        <notification-alert-bar
          class="change-password-info"
          :variant="ENotificationBarVariant.Info"
          message="This will sign you out of all active sessions."
        />

        <div class="change-password-fields">
          <q-input
            v-model="currentPassword"
            dense
            :type="isCurrentPasswordHidden ? 'password' : 'text'"
            label="Current password"
            :rules="[currentPasswordRequired]"
            lazy-rules
            aid="current-password"
            no-error-icon
          >
            <template #append>
              <q-icon
                aid="toggle-current-password-visibility"
                :name="isCurrentPasswordHidden ? 'ph-fill ph-eye-slash' : 'ph-fill ph-eye'"
                class="eye-icon text-black-54"
                @click="isCurrentPasswordHidden = !isCurrentPasswordHidden"
              >
                <q-tooltip>{{ isCurrentPasswordHidden ? "show password" : "hide password" }}</q-tooltip>
              </q-icon>
            </template>
          </q-input>

          <q-input
            v-model="password"
            dense
            :type="isPasswordHidden ? 'password' : 'text'"
            label="New password"
            :hint="passwordFormatHint"
            :rules="[validNewPassword]"
            lazy-rules
            aid="new-password"
            no-error-icon
          >
            <template #append>
              <q-icon
                aid="toggle-new-password-visibility"
                :name="isPasswordHidden ? 'ph-fill ph-eye-slash' : 'ph-fill ph-eye'"
                class="eye-icon text-black-54"
                @click="isPasswordHidden = !isPasswordHidden"
              >
                <q-tooltip>{{ isPasswordHidden ? "show password" : "hide password" }}</q-tooltip>
              </q-icon>
            </template>
          </q-input>

          <q-input
            v-model="verifyPassword"
            dense
            :type="isVerifyPasswordHidden ? 'password' : 'text'"
            label="Confirm new password"
            :rules="[verifyPasswordValidation]"
            lazy-rules
            aid="verify-password"
            no-error-icon
          >
            <template #append>
              <q-icon
                aid="toggle-verify-password-visibility"
                :name="isVerifyPasswordHidden ? 'ph-fill ph-eye-slash' : 'ph-fill ph-eye'"
                class="eye-icon text-black-54"
                @click="isVerifyPasswordHidden = !isVerifyPasswordHidden"
              >
                <q-tooltip>{{ isVerifyPasswordHidden ? "show password" : "hide password" }}</q-tooltip>
              </q-icon>
            </template>
          </q-input>
        </div>
      </q-form>
    </template>

    <template #footer>
      <runai-btn label="Cancel" color="primary" flat @click="$emit('close')" aid="cancel-button" />
      <runai-btn
        label="Change Password"
        color="primary"
        aid="reset-password-button"
        :loading="processing"
        @click="reset"
      />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { QForm } from "quasar";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiBtn } from "@/components/common/runai-btn";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";

// stores
import { useAuthStore } from "@/stores/auth.store";

// services
import { authService } from "@/services/control-plane/auth.service/auth.service";

// utils
import { alertUtil } from "@/utils/alert.util";

// others
import { errorMessages } from "@/common/error-message.constant";
import { isValidPassword } from "@/common/form.validators";
import type { HttpError } from "@/models/global.model";

export default defineComponent({
  components: {
    RunaiBtn,
    RunaiBaseDialog,
    NotificationAlertBar,
  },
  emits: ["close"],
  props: {
    reloadUser: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      currentPassword: "" as string,
      password: "" as string,
      verifyPassword: "" as string,
      processing: false as boolean,
      isCurrentPasswordHidden: true as boolean,
      isPasswordHidden: true as boolean,
      isVerifyPasswordHidden: true as boolean,
      ENotificationBarVariant,
    };
  },
  computed: {
    passwordFormatHint(): string {
      return errorMessages.INVALID_PASSWORD_FORMAT;
    },
  },
  methods: {
    currentPasswordRequired(val: string): boolean | string {
      return !!val || errorMessages.PASSWORD_NOT_EMPTY;
    },
    validNewPassword(val: string): boolean | string {
      return isValidPassword(val) || errorMessages.INVALID_PASSWORD_FORMAT;
    },
    verifyPasswordValidation(val: string): boolean | string {
      return this.password === val || errorMessages.PASSWORD_DOES_NOT_MATCH;
    },
    async reset(): Promise<void> {
      // Guard against re-entry: validate() can be async (lazy rules), and
      // without this a rapid second click would fire a second changePassword.
      if (this.processing) return;
      this.processing = true;

      const form = this.$refs.formRef as QForm | undefined;
      const isValid = (await form?.validate()) ?? false;
      if (!isValid) {
        this.processing = false;
        return;
      }

      try {
        await authService.changePassword(this.password, this.currentPassword);
        if (this.reloadUser) {
          await this.authStore.loadUserInfo();
        }
        this.$q.notify(alertUtil.getSuccess("Change password succeeded!"));
        this.$emit("close");
        this.authStore.logout();
      } catch (error: unknown) {
        if ((error as HttpError)?.response?.status === 401) {
          this.$q.notify(alertUtil.getError("Current password is incorrect!"));
        } else {
          this.$q.notify(alertUtil.getError("Change password failed!"));
        }
      } finally {
        this.processing = false;
      }
    },
  },
});
</script>

<style lang="scss" scoped>
.change-password-form {
  display: flex;
  flex-direction: column;
  padding: var(--space-m);
}

.change-password-info {
  margin-bottom: var(--space-m);
}

.change-password-fields {
  display: flex;
  flex-direction: column;
  // Figma wraps each field in a 16px-padded slot → ~32px between fields.
  // 2xl (40px) absorbs q-input's intrinsic .q-field__bottom space and lands
  // visually around 32px between the visible elements.
  gap: var(--space-2xl);
}

.eye-icon {
  font-size: 20px;
  cursor: pointer;
}
</style>
