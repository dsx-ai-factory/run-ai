<template>
  <onboarding-content-wrapper class="protocol-substep">
    <template #title>Select the SSO protocol</template>
    <template #subtitle>
      Anyone given access to the NVIDIA Run:ai application via the identity provider system will be able to join the
      platform and authenticate. For more information, see the
      <a :href="DOCS.AUTHENTICATION.SSO" target="_blank">Single Sign-On guide</a>
    </template>
    <template #content>
      <div class="protocol-setup-content">
        <!-- Delegation Option -->
        <delegation-option
          :is-delegated="ssoData.isDelegated"
          :delegate-email="ssoData.delegateEmail"
          :read-only="ssoData.hasIdpCreated"
          description-text="Provide the email of the user responsible for SSO configuration, and we'll send them an invitation to continue."
          @update:is-delegated="handleDelegationChange"
          @update:delegate-email="handleDelegateEmailChange"
          class="q-mb-xl"
        />

        <!-- Protocol Selection (when not delegated) -->
        <transition name="protocol-content-transition" mode="out-in">
          <div v-if="!ssoData.isDelegated" key="protocol-selection" class="protocol-selection-section">
            <div class="onboarding-select-options">
              <runai-select-card
                v-for="option in protocolOptions"
                :key="option.protocol"
                :value="ssoData.protocol === option.protocol"
                :disabled="ssoData.hasIdpCreated"
                @clicked="selectProtocol(option.protocol)"
                class="onboarding-select-card"
              >
                <div class="option-icon">
                  <runai-svg-icon :name="option.icon" size="40" />
                </div>
                <div class="option-content">
                  <h3 class="option-title">{{ option.title }}</h3>
                  <p class="option-description">
                    {{ option.description }}
                  </p>
                </div>
              </runai-select-card>
            </div>

            <!-- HIDDEN FIELD FOR VALIDATION -->
            <q-field
              class="hidden-field onboarding-select-error"
              :rules="ssoData.isDelegated ? [] : [notEmpty]"
              no-error-icon
              :model-value="ssoData.protocol"
            />
          </div>
        </transition>
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { RunaiSelectCard } from "@/components/common/runai-select-card";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { DelegationOption } from "@/components/onboarding/common/delegation-option";
import { isNotEmpty } from "@/common/form.validators";
import type { SsoData } from "@/components/onboarding/models/admin-onboarding.model";
import { ssoDataUpdateEmitValidator } from "@/components/onboarding/models/admin-onboarding.model";
import DOCS from "@/constants/docs";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import { settingsUtil } from "@/utils/settings.util";
import { useAppStore } from "@/stores/app.store";

export default defineComponent({
  name: "protocol-substep",
  components: { OnboardingContentWrapper, RunaiSelectCard, RunaiSvgIcon, DelegationOption },
  emits: {
    "update:sso-data": ssoDataUpdateEmitValidator,
  },
  props: {
    ssoData: {
      type: Object as PropType<SsoData>,
      required: true,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      DOCS,
      protocolOptions: [
        {
          protocol: IdpTypeEnum.Saml,
          icon: "saml",
          title: "SAML 2.0",
          description: "Configure SSO to NVIDIA Run:ai using the SAML 2.0 protocol",
        },
        {
          protocol: IdpTypeEnum.Oidc,
          icon: "openid",
          title: "OpenID Connect",
          description: "Configure single sign-on to NVIDIA Run:ai using the OpenID Connect protocol",
        },
      ],
    };
  },
  methods: {
    selectProtocol(protocol: IdpTypeEnum) {
      this.$emit("update:sso-data", "protocol", protocol);
      const authHost = this.appStore.config.authHost;
      const idp = settingsUtil.getDefaultIdpConfig(protocol, authHost);
      if (idp) {
        this.$emit("update:sso-data", "idp", idp);
      }
    },
    handleDelegationChange(isDelegated: boolean) {
      this.$emit("update:sso-data", "isDelegated", isDelegated);
      // Clear protocol when delegating
      if (isDelegated) {
        this.$emit("update:sso-data", "protocol", "");
      }
    },

    handleDelegateEmailChange(email: string) {
      this.$emit("update:sso-data", "delegateEmail", email);
    },

    notEmpty(val: string) {
      return isNotEmpty(val) || "Please select an SSO protocol";
    },
  },
});
</script>
