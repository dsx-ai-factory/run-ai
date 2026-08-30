<template>
  <span class="help-dropdown">
    <runai-btn
      class="help-dropdown runai-icon-btn--default"
      flat
      round
      icon="ph-bold ph-question"
      color="icon-default"
      aid="help-dropdown"
      aria-label="Help"
    >
      <q-menu class="help-dropdown-menu" :anchor="anchor" :self="self">
        <q-list>
          <q-item v-if="!isWhiteLabel" v-close-popup clickable @click="openUrl(runaiDocs)">
            <q-item-section>
              <q-item-label class="text-b-subtitle2 help-menu-item-label">
                {{ EHelpMenuTexts.DOCS }}
              </q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-icon class="help-menu-external-icon" name="ph-bold ph-arrow-square-out" size="20px" />
            </q-item-section>
          </q-item>

          <q-item v-if="!isWhiteLabel" v-close-popup clickable @click="openUrl(apiDocsLink)">
            <q-item-section>
              <q-item-label class="text-b-subtitle2 help-menu-item-label">
                {{ EHelpMenuTexts.API }}
              </q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-icon class="help-menu-external-icon" name="ph-bold ph-arrow-square-out" size="20px" />
            </q-item-section>
          </q-item>

          <q-item clickable v-close-popup @click="openCliInstallationModal" :disable="isDownloadCLILinkDisabled">
            <q-item-section>
              <runai-tooltip-wrapper
                :display-tooltip="isDownloadCLILinkDisabled"
                :tooltip-text="disabledLinkToolTip"
                max-width="400px"
              >
                <q-item-label class="text-b-subtitle2 help-menu-item-label" aid="researcher-cli-option">
                  {{ EHelpMenuTexts.RESEARCHER }}
                </q-item-label>
              </runai-tooltip-wrapper>
            </q-item-section>
          </q-item>

          <q-item v-if="!isWhiteLabel && supportUrl" v-close-popup clickable @click="openUrl(supportUrl)">
            <q-item-section>
              <q-item-label class="text-b-subtitle2 help-menu-item-label">
                {{ EHelpMenuTexts.SUPPORT }}
              </q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-icon class="help-menu-external-icon" name="ph-bold ph-arrow-square-out" size="20px" />
            </q-item-section>
          </q-item>

          <q-item v-if="!isWhiteLabel" clickable v-close-popup @click="openWhatsNewUrl">
            <q-item-section>
              <q-item-label class="text-b-subtitle2 help-menu-item-label">
                {{ whatNewText }}
              </q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-icon class="help-menu-external-icon" name="ph-bold ph-arrow-square-out" size="20px" />
            </q-item-section>
          </q-item>

          <guided-steps-menu />
        </q-list>
      </q-menu>
    </runai-btn>
    <q-tooltip class="q-pa-sm">
      <div>Help</div>
    </q-tooltip>

    <researcher-download-cli-modal v-if="openCliModal" :open="openCliModal" @close="openCliModal = false" />
  </span>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { DOCS } from "@/constants/docs";

// cmps
import { ResearcherDownloadCliModal } from "../researcher-download-cli-modal";
import { GuidedStepsMenu } from "./guided-steps-menu";
// stores
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { openNewWindow } from "@/utils/common.util/common.util";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { RunaiBtn } from "@/components/common/runai-btn";

export enum EHelpMenuTexts {
  DOCS = "Documentation Center",
  API = "Administrator API",
  RESEARCHER = "Researcher Command Line Interface",
  SUPPORT = "Contact Support",
  WHATS_NEW = "See what's new",
}

export default defineComponent({
  name: "help-dropdown",
  components: {
    RunaiBtn,
    RunaiTooltipWrapper,
    ResearcherDownloadCliModal,
    GuidedStepsMenu,
  },
  data() {
    return {
      authStore: useAuthStore(),
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      permissionStore: usePermissionStore(),
      openCliModal: false as boolean,
      EHelpMenuTexts: EHelpMenuTexts,
      DOCS,
    };
  },
  computed: {
    supportUrl(): string {
      return DOCS.SUPPORT.URL;
    },
    apiDocsLink(): string {
      return `${window.location.origin}/api/docs`;
    },
    runaiDocs(): string {
      return DOCS.ROOT;
    },
    isWhiteLabel(): boolean {
      return this.appStore.isWhiteLabel;
    },
    anchor(): "bottom right" | "bottom middle" {
      return this.isWhiteLabel ? "bottom right" : "bottom middle";
    },
    self(): "top end" | "top middle" {
      return this.isWhiteLabel ? "top end" : "top middle";
    },
    whatNewText(): string {
      return DOCS.WHATS_NEW.LABEL;
    },
    isDownloadCLILinkDisabled(): boolean {
      return this.clusterStore?.clusters.length === 0;
    },
    disabledLinkToolTip() {
      const isAllowedToCreateCluster = this.permissionStore.hasPermission(ResourceType.Cluster, Action.Create);
      return isAllowedToCreateCluster
        ? "To install the Researcher CLI first add at least one cluster"
        : "To install the Researcher CLI you need to be assigned to a cluster. Contact your administrator to be assigned.";
    },
    docLink(): string {
      return DOCS.WHATS_NEW.URL;
    },
  },
  methods: {
    openCliInstallationModal(): void {
      this.openCliModal = true;
    },
    openUrl(url: string): void {
      openNewWindow(url);
    },
    openWhatsNewUrl(): void {
      openNewWindow(this.docLink);
    },
  },
});
</script>

<style lang="scss">
.help-dropdown-menu.q-menu {
  background: var(--bg-primary-default);
  box-shadow: var(--shadow-m);
  border: none;
  border-radius: var(--space-2xs);
  width: 270px;

  .q-item {
    padding: var(--space-m);
  }

  .help-menu-item-label {
    line-height: 19.5px !important; // overrides Quasar's .q-item__label { line-height: 1.2em !important }
  }

  .help-menu-external-icon {
    color: var(--icon-default, #707070);
  }
}
</style>
