<template>
  <onboarding-content-wrapper :center-align="centerAlign">
    <template #title>{{ title }}</template>

    <template #subtitle>
      <!-- ACCESS RULES / USERS CREATION FAILED  -->
      <div v-if="completionState === IOnboardingCompletionState.AccessRulesCreationFailed">
        We ran into some trouble adding the following users to the team:
        <strong>{{ failedSubjectIds.join(", ") }}</strong>
        <br />
        Try adding them manually on the
        <a href="#" @click.prevent="redirectToUsers">users page</a>
        <template v-if="DOCS.SUPPORT.URL"> or <span v-dompurify-html="supportLink()"></span></template>.
      </div>

      <!-- ADMIN USER CREATION FAILED -->
      <div v-if="completionState === IOnboardingCompletionState.AdminUserCreationFailed">
        Try adding them manually on the
        <a href="#" @click.prevent="redirectToUsers">users page</a>
        <template v-if="DOCS.SUPPORT.URL"> or <span v-dompurify-html="supportLink()"></span></template>.
      </div>

      <!-- PROJECT CREATION FAILED -->
      <div v-else-if="completionState === IOnboardingCompletionState.ProjectCreationFailed">
        Try
        <a href="#" @click.prevent="retry">creating the team again</a>
        <template v-if="DOCS.SUPPORT.URL"> or <span v-dompurify-html="supportLink()"></span></template>.
      </div>

      <!--  OTHER COMPLETION STATES -->
      <div v-else v-dompurify-html="subtitle"></div>
    </template>

    <template #content>
      <div class="completion-content-body" :style="contentBodyStyles">
        <runai-svg-icon :name="completionSvg" size="100%" :key="completionState" class="completion-svg" />
        <div v-if="shouldShowDownloadButton" class="download-section">
          <div class="download-description">
            <span class="download-description-text">
              We recommend downloading your team members' NVIDIA Run:ai sign-in credentials and using them if the email
              invitations didn't go through.
            </span>
          </div>
          <q-btn
            class="download-button q-mt-md"
            color="grey-4"
            text-color="grey-8"
            @click="downloadCredentials"
            icon="ph-bold ph-download-simple"
            label="DOWNLOAD TEAM'S CREDENTIALS"
            no-caps
          />
        </div>
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { computed, defineComponent, type PropType } from "vue";
import { OnboardingContentWrapper } from "@/components/onboarding/common/onboarding-content-wrapper";
import {
  onboardingCompletionStateSvg,
  IOnboardingCompletionState,
} from "@/components/onboarding/models/admin-onboarding.model";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import DOCS from "@/constants/docs";

import { USER_ROUTES_NAMES } from "@/router/users.routes/users.routes.names";
import type { UserPostResponse } from "@/swagger-models/identity-manager-client";
import { useDownloadCredentials } from "@/composables/use-download-credentials.composable";
import { storageUtil } from "@/utils/storage.util/storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";
import { useAuthStore } from "@/stores/auth.store";

const createLink = (href: string, text: string, target = "_blank"): string => {
  return `<a href="${href}" target="${target}">${text}</a>`;
};

const documentationLink = (text = "documentation"): string => {
  return createLink(DOCS.ROOT, text);
};

const supportLink = (text = "contact support"): string => {
  return DOCS.SUPPORT.URL ? createLink(DOCS.SUPPORT.URL, text) : "";
};

export default defineComponent({
  name: "onboarding-completion-status",
  components: { RunaiSvgIcon, OnboardingContentWrapper },
  emits: {
    close: (routeName?: string) => typeof routeName === "string" || routeName === undefined,
    retry: () => true,
  },
  props: {
    completionState: {
      type: String as PropType<IOnboardingCompletionState>,
      required: true,
    },
    failedSubjectIds: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    centerAlign: {
      type: Boolean,
      default: false,
    },
    contentPaddingTop: {
      type: String as PropType<string>,
      default: "0",
    },
    createdUsers: {
      type: Array as PropType<UserPostResponse[]>,
      default: () => [],
    },
    teamName: {
      type: String as PropType<string>,
      default: "",
    },
  },
  setup(props, { emit }) {
    const { downloadCredentials: downloadCredentialsComposable } = useDownloadCredentials();

    const redirectToUsers = () => {
      emit("close", USER_ROUTES_NAMES.USER_INDEX);
    };

    const retry = () => {
      emit("retry");
    };

    const downloadCredentials = async () => {
      await downloadCredentialsComposable({
        createdUsers: props.createdUsers,
        teamName: props.teamName,
      });
    };

    const title = computed(() => {
      switch (props.completionState) {
        case IOnboardingCompletionState.ProjectCreationFailed:
          return "Yikes, we couldn't get your team up and running";
        case IOnboardingCompletionState.AccessRulesCreationFailed:
          return "Looks like a few lovely people didn't make it";
        case IOnboardingCompletionState.AdminUserCreationFailed:
          return `The invitation to ${props.failedSubjectIds[0]} didn't make it out the door`;
        default:
          return "You're all set!";
      }
    });

    const subtitle = computed(() => {
      switch (props.completionState) {
        case IOnboardingCompletionState.Completed:
          return [
            "Your first team is ready to start running workloads.",
            `Take a look around and check out ${documentationLink()} to learn how to get the most out of NVIDIA Run:ai.`,
          ].join("<br>");

        case IOnboardingCompletionState.CompletedWithoutProject:
          return `Take a look around and check out ${documentationLink()} to learn how to get the most out of NVIDIA Run:ai.`;

        // Special cases, handled in the template
        case IOnboardingCompletionState.AccessRulesCreationFailed:
        case IOnboardingCompletionState.ProjectCreationFailed:
        default:
          return "";
      }
    });

    const completionSvg = computed(() => onboardingCompletionStateSvg[props.completionState]);

    const contentBodyStyles = computed(() => ({
      paddingTop: props.contentPaddingTop,
    }));

    const isTeamSetupSkipped = storageUtil.get<boolean>(STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP);
    const isSSOEnabled = useAuthStore().isSSO;
    const shouldShowDownloadButton = computed(() => {
      return (
        !isTeamSetupSkipped &&
        (props.completionState === IOnboardingCompletionState.Completed ||
          props.completionState === IOnboardingCompletionState.CompletedWithoutProject ||
          (props.completionState === IOnboardingCompletionState.AccessRulesCreationFailed &&
            props.createdUsers.length > 0)) &&
        !isSSOEnabled
      );
    });

    return {
      completionSvg,
      title,
      subtitle,
      contentBodyStyles,
      redirectToUsers,
      retry,
      supportLink,
      downloadCredentials,
      shouldShowDownloadButton,
      IOnboardingCompletionState,
      DOCS,
    };
  },
});
</script>

<style scoped lang="scss">
.download-button {
  min-width: 280px;
  height: 42px;
  font-weight: 600;
  font-size: 14px;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.download-description-text {
  font-weight: 400;
  font-size: 16px;
  line-height: 26px;
}
</style>
