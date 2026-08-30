<template>
  <form>
    <div class="row col-12">
      <setting-readonly-input class="col-9" :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
        <template v-slot:default="{ readonly }">
          <!-- eslint-disable vuejs-accessibility/no-autofocus -- this form only renders once the user clicks Edit on the email server settings box, never on a bare page load -->
          <q-input
            aid="smpt-host-input"
            class="col-9 q-mb-lg"
            :model-value="emailServerSettings.smtpHost"
            placeholder="e.g., smtp.mail.com"
            input-class="placeholder-italic"
            :rules="[isValidUrl]"
            no-error-icon
            @update:model-value="updateSmtpHost"
            stack-label
            label="SMTP host"
            ref="smtpHostInput"
            autofocus
            :disable="readonly"
          />
          <!-- eslint-enable vuejs-accessibility/no-autofocus -->
        </template>
      </setting-readonly-input>
      <setting-readonly-input class="col-2" :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
        <template v-slot:default="{ readonly }">
          <q-input
            aid="smpt-port-input"
            class="col-2 q-ml-lg q-mb-lg"
            :model-value="emailServerSettings.smtpPort || DEFAULT_SMTP_PORT"
            :rules="[isValidPort]"
            no-error-icon
            @update:model-value="updatePort"
            stack-label
            label="SMTP port"
            ref="smtpPortInput"
            type="number"
            :disable="readonly"
          />
        </template>
      </setting-readonly-input>
      <div class="row col-12">
        <q-item-label :class="{ 'text-grey-6': readOnly }">Set the authentication type for the SMTP host</q-item-label>
      </div>
      <div class="row col-12 q-my-md auth-type-radio">
        <setting-readonly-input :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
          <template v-slot:default="{ readonly }">
            <div>
              <q-radio
                :disable="readonly"
                :model-value="emailServerSettings.authType"
                aid="auth-type-radio-plain"
                label="Plain"
                :val="EmailAuthType.AuthPlain"
                @update:model-value="updateAuthType"
              />
            </div>
            <div>
              <q-radio
                :disable="readonly"
                :model-value="emailServerSettings.authType"
                aid="auth-type-radio-login"
                label="Login"
                :val="EmailAuthType.AuthLogin"
                @update:model-value="updateAuthType"
              />
            </div>
            <div>
              <q-radio
                :disable="readonly"
                :model-value="emailServerSettings.authType"
                aid="auth-type-radio-none"
                label="None"
                :val="EmailAuthType.AuthNone"
                @update:model-value="updateAuthType"
              />
            </div>
          </template>
        </setting-readonly-input>
      </div>
      <div class="row col-12" v-if="emailServerSettings.authType !== EmailAuthType.AuthNone">
        <setting-readonly-input class="col-5" :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
          <template v-slot:default="{ readonly }">
            <q-input
              aid="smpt-username-input"
              class="col-5 q-mr-lg q-mb-lg"
              :model-value="emailServerSettings.username"
              :rules="[isNotEmptyUsername]"
              no-error-icon
              @update:model-value="updateUsername"
              label="Username"
              ref="usernameInput"
              :disable="readonly"
            />
          </template>
        </setting-readonly-input>
        <setting-readonly-input class="col-5" :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
          <template v-slot:default="{ readonly }">
            <q-input
              aid="smpt-password-input"
              class="col-5 q-mb-lg"
              :model-value="emailServerSettings.password"
              :rules="[isNotEmptyPassword]"
              no-error-icon
              @update:model-value="updatePassword"
              autocomplete="on"
              label="Password"
              ref="passwordInput"
              type="password"
              :disable="readonly"
            />
          </template>
        </setting-readonly-input>
      </div>
      <div class="row col-12">
        <setting-readonly-input class="col-5" :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
          <template v-slot:default="{ readonly }">
            <q-input
              aid="smpt-from-email-input"
              class="col-5 q-mr-lg q-mb-sm"
              :model-value="emailServerSettings.fromEmail"
              placeholder="e.g., nvidia_run:ai@company.com"
              input-class="placeholder-italic"
              :rules="[isValidEmail]"
              no-error-icon
              @update:model-value="updateFromEmail"
              stack-label
              label="From email address"
              ref="fromEmailInput"
              :disable="readonly"
            />
          </template>
        </setting-readonly-input>
        <setting-readonly-input class="col-5" :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
          <template v-slot:default="{ readonly }">
            <q-input
              aid="smpt-from-name-input"
              class="col-5 q-mb-sm"
              :model-value="emailServerSettings.fromName"
              placeholder="e.g., NVIDIA Run:ai notification"
              input-class="placeholder-italic"
              :rules="[isNotEmptyFromName]"
              no-error-icon
              @update:model-value="updateFromName"
              stack-label
              label="From display name"
              ref="fromNameInput"
              :disable="readonly"
            />
          </template>
        </setting-readonly-input>
      </div>
    </div>
  </form>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { DEFAULT_SMTP_PORT, type ISettingEmailServer } from "@/models/setting.model";
import { isNotEmpty, isNumberInRange, isValidDomain, isValidEmail } from "@/common/form.validators";
import type { QInput } from "quasar";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import { EmailAuthType } from "@/swagger-models/notifications-proxy-client";

const MAX_PORT_NUMBER = 65535;
export default defineComponent({
  name: "email-server-form",
  components: { SettingReadonlyInput },
  emits: ["update"],
  props: {
    emailServerSettings: {
      type: Object as PropType<ISettingEmailServer>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: undefined,
    },
    displayReadOnlyTooltip: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  computed: {
    EmailAuthType() {
      return EmailAuthType;
    },
    DEFAULT_SMTP_PORT() {
      return DEFAULT_SMTP_PORT;
    },
  },
  methods: {
    isValidUrl(url: string): boolean | string {
      return isValidDomain(url) || "Enter a valid URL";
    },
    isValidPort(port: number): boolean | string {
      if (String(port) === "" || port == undefined) {
        return "Enter port";
      }
      return isNumberInRange(port, 0, MAX_PORT_NUMBER) || "Enter a valid port number";
    },
    isNotEmptyUsername(value: string): boolean | string {
      return isNotEmpty(value) || "Enter username";
    },
    isNotEmptyPassword(value: string): boolean | string {
      return isNotEmpty(value) || "Enter password";
    },
    isValidEmail(value: string): boolean | string {
      return isValidEmail(value) || "Enter email address";
    },
    isNotEmptyFromName(value: string): boolean | string {
      return isNotEmpty(value) || "Enter display name";
    },
    updateSmtpHost(smtpHost: string | number | null): void {
      this.$emit("update", { ...this.emailServerSettings, smtpHost }, true);
    },
    updatePort(smtpPort: string | number | null): void {
      smtpPort = Number(smtpPort);
      this.$emit("update", { ...this.emailServerSettings, smtpPort }, true);
    },
    updateUsername(username: string | number | null): void {
      this.$emit("update", { ...this.emailServerSettings, username }, true);
    },
    updatePassword(password: string | number | null): void {
      this.$emit("update", { ...this.emailServerSettings, password }, true);
    },
    updateFromEmail(fromEmail: string | number | null): void {
      this.$emit("update", { ...this.emailServerSettings, fromEmail }, false);
    },
    updateFromName(fromName: string | number | null): void {
      this.$emit("update", { ...this.emailServerSettings, fromName }, false);
    },
    updateAuthType(authType: string): void {
      this.$emit("update", { ...this.emailServerSettings, authType }, true);
    },
    async validate(): Promise<boolean> {
      const fieldsToValidate: Array<Promise<boolean>> = [
        (this.$refs.smtpHostInput as typeof QInput).validate(),
        (this.$refs.smtpPortInput as typeof QInput).validate(),
        (this.$refs.fromEmailInput as typeof QInput).validate(),
        (this.$refs.fromNameInput as typeof QInput).validate(),
      ];
      if (this.emailServerSettings.authType !== EmailAuthType.AuthNone) {
        fieldsToValidate.push((this.$refs.usernameInput as typeof QInput).validate());
        fieldsToValidate.push((this.$refs.passwordInput as typeof QInput).validate());
      }
      const validFields: boolean[] = await Promise.all(fieldsToValidate);
      return validFields.every((valid) => valid);
    },
  },
});
</script>
<style scoped lang="scss">
.auth-type-radio {
  margin-left: -10px;
}
</style>
