<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    :model-value="open"
    @close="$emit('close')"
    size="width-md"
    dialog-body-size="extra-large"
    no-padding
    scroll
  >
    <template #header>Researcher Command Line Interface</template>

    <template #body>
      <section class="researcher-download-cli q-pa-md">
        <template v-if="clusterOptions.length > 1">
          <div class="q-mb-md">Select the cluster you want the CLI to communicate with</div>
          <div class="row">
            <q-select
              label="Cluster"
              class="col-6 cluster-select q-mb-lg"
              :options="clusterOptions"
              :model-value="selectedCluster"
              placeholder="Select a cluster"
              aid="cluster-select-cli-modal"
              @update:model-value="onSelectCluster"
              no-error-icon
            />
          </div>
        </template>
        <section class="os-selection-box">
          <p>{{ radioOptionsTitle }}</p>
          <runai-radio-options
            class="q-mb-md"
            :options="allOSOptions"
            :model-value="selectedOS"
            @update:model-value="onSelectOS"
            horizontal
          />
        </section>

        <template v-if="selectedOS && selectedCluster">
          <q-separator />
          <section>
            <section class="installation-instructions">
              <!-- UNIX -->
              <template v-if="showUnixInstaller">
                <div class="q-py-md">Copy the following and run it in the terminal to install Run:ai CLI:</div>
                <q-skeleton
                  v-if="loadingInstallerAndCommands && !cliCommandForDisplay"
                  class="q-mb-md"
                  type="rect"
                  height="60px"
                />
                <div v-else-if="!cliCommandForDisplay" class="q-mb-md text-b-text-sub" aid="cli-command-load-error">
                  {{ cliCommandLoadErrorMessage }}
                </div>
                <runai-copyable-code-block v-else class="q-mb-md" :code="cliCommandForDisplay" :max-height="60" />
                <runai-sub-expansion-text-item class="installers-sections">
                  <template #header>
                    <div>
                      <span>Installer file <span class="optional">(Optional)</span></span>
                    </div>
                  </template>
                  <template #body>
                    <div>
                      <div class="q-mb-sm">Download the installer file for</div>
                      <div class="download-buttons">
                        <download-button-with-progress
                          :download-progress="downloadProgress.amd64.value"
                          :loading="downloadingNewAmd64Cli"
                          aid="download-amd64"
                          label="AMD64 (Intel)"
                          @download="downloadUnixInstallerFile(EArchitectureType.AMD)"
                        />
                        <span class="q-mx-md">OR</span>
                        <download-button-with-progress
                          :download-progress="downloadProgress.arm64.value"
                          :loading="downloadingNewArm64Cli"
                          aid="download-arm64"
                          label="ARM64"
                          @download="downloadUnixInstallerFile(EArchitectureType.ARM)"
                        />
                      </div>
                    </div>
                  </template>
                </runai-sub-expansion-text-item>
              </template>
              <!-- WINDOWS -->
              <template v-else>
                <div class="q-my-md">
                  <div class="q-mr-sm q-mb-sm">1. Download the binary file</div>
                  <download-button-with-progress
                    :download-progress="downloadProgress.windows.amd64.value"
                    :loading="downloadingNewWindowsAmd64Cli"
                    aid="download-windows-amd64"
                    label="AMD64 (Intel)"
                    @download="downloadNewCliWindowsAmd64"
                  />
                  <span class="q-mx-md">OR</span>
                  <download-button-with-progress
                    :download-progress="downloadProgress.windows.arm64.value"
                    :loading="downloadingNewWindowsArm64Cli"
                    aid="download-windows-arm64"
                    label="ARM64"
                    @download="downloadNewCliWindowsArm64"
                  />
                </div>
                <div class="q-mb-sm">2. Rename the downloaded file to have a .exe extension</div>
                <div class="q-mb-sm">3. Move the file to a folder that is a part of the PATH</div>
              </template>
            </section>
          </section>
        </template>

        <section class="more-info q-mt-md" v-if="!isWhiteLabel">
          For more information, see the
          <a :href="DOCS.CLI.INSTALL" target="_blank" aid="cli-install-doc-link"> Install and Configure CLI guide </a>
          and the
          <a :href="getDocumentationLink" target="_blank" aid="cli-documentation-link">CLI documentation</a>
        </section>
      </section>
    </template>

    <template #footer>
      <runai-btn label="Close" color="primary" @click="$emit('close')" aid="researcher-cluster-cli-modal-close-button" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// stores
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiSubExpansionTextItem } from "@/components/common/runai-sub-expansion-text-item";
import RunaiCopyableCodeBlock from "@/components/common/runai-copyable-code-block/runai-copyable-code-block.vue";

// utils
import {
  EArchitectureType,
  EOperatingSystems,
  cliService,
} from "@/services/control-plane/cli-exposer.service/cli-exposer.service";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { clusterUtil } from "@/utils/cluster.util";
import { DownloadButtonWithProgress } from "@/components/infra/top-header/researcher-download-cli-modal/download-button-with-progress";
import { DOCS } from "@/constants/docs";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";

export interface IClusterSelectOption {
  value: string;
  label: string;
  version: string | null | undefined;
  domain: string | null | undefined;
}

export default defineComponent({
  components: {
    RunaiCopyableCodeBlock,
    RunaiBaseDialog,
    RunaiBtn,
    RunaiRadioOptions,
    RunaiSubExpansionTextItem,
    DownloadButtonWithProgress,
  },
  props: {
    open: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      selectedOS: "" as EOperatingSystems,
      authStore: useAuthStore(),
      appStore: useAppStore(),
      selectedCluster: null as IClusterSelectOption | null,
      loadingInstallerAndCommands: false,
      newCliWindowsCommand: "" as string,
      newCliLinuxCommand: "" as string,
      newCliMacCommand: "" as string,
      EArchitectureType,
      downloadingNewAmd64Cli: false,
      downloadingNewArm64Cli: false,
      downloadingNewWindowsAmd64Cli: false,
      downloadingNewWindowsArm64Cli: false,
      allOSOptions: [
        { label: "Linux", value: EOperatingSystems.LINUX },
        { label: "Mac", value: EOperatingSystems.MAC },
        { label: "Windows", value: EOperatingSystems.WINDOWS },
      ],
      downloadProgress: {
        [EArchitectureType.AMD]: { value: 0 },
        [EArchitectureType.ARM]: { value: 0 },
        [EOperatingSystems.WINDOWS]: {
          [EArchitectureType.AMD]: { value: 0 },
          [EArchitectureType.ARM]: { value: 0 },
        },
      },
      DOCS,
    };
  },
  async created() {
    // Auto-select first cluster, regardless of whether dropdown is shown
    !!this.clusterOptions[0] && this.onSelectCluster(this.clusterOptions[0]);
    await this.loadInstallerAndCommands();
  },
  computed: {
    cliCommandLoadErrorMessage(): string {
      return `${ErrorAlert.failedLoadingMessage("install command")}. Please refresh the page and try again.`;
    },
    showUnixInstaller(): boolean {
      return this.selectedOS === EOperatingSystems.LINUX || this.selectedOS === EOperatingSystems.MAC;
    },
    getDocumentationLink(): string {
      return `${this.controlPlaneUrl}/api/v1/cli/docs/runai.md`;
    },
    accessToken(): string {
      return this.authStore.accessToken;
    },
    controlPlaneUrl(): string {
      return window.location.origin;
    },
    radioOptionsTitle(): string {
      return "Select the operating system of your computer ";
    },
    clusterOptions(): IClusterSelectOption[] {
      return this.clusterStore.clusterList
        .filter((c) => {
          return c.version && c.domain && clusterUtil.isClusterReachable(c.state);
        })
        .map((c: ClustersMinimalData) => {
          return {
            label: c.name,
            value: c.uuid,
            version: c.version,
            domain: c.domain,
          };
        });
    },
    cliCommandForDisplay(): string {
      return this.selectedOS === EOperatingSystems.LINUX ? this.newCliLinuxCommand : this.newCliMacCommand;
    },
    isWhiteLabel(): boolean {
      return this.appStore.isWhiteLabel;
    },
  },
  methods: {
    onSelectCluster(newSelection: IClusterSelectOption): void {
      this.selectedCluster = newSelection;
      this.selectedOS = "" as EOperatingSystems;
    },
    async onSelectOS(newSelection: EOperatingSystems): Promise<void> {
      this.selectedOS = newSelection;
      if (this.showUnixInstaller && !this.cliCommandForDisplay && !this.loadingInstallerAndCommands) {
        await this.loadInstallerAndCommands();
      }
    },
    async loadInstallerAndCommands(): Promise<void> {
      try {
        this.loadingInstallerAndCommands = true;
        [this.newCliLinuxCommand, this.newCliMacCommand, this.newCliWindowsCommand] = await Promise.all([
          cliService.getUILinuxCommand(),
          cliService.getUIMacCommand(),
          cliService.getUIWindowsCommand(),
        ]);
      } catch (err: unknown) {
        console.error(err);
      } finally {
        this.loadingInstallerAndCommands = false;
      }
    },
    async downloadNewCliWindowsAmd64(): Promise<void> {
      if (this.downloadingNewWindowsAmd64Cli) return;
      try {
        this.downloadingNewWindowsAmd64Cli = true;
        await cliService.getBinaryWithProgress(
          EOperatingSystems.WINDOWS,
          EArchitectureType.AMD,
          this.downloadProgress[EOperatingSystems.WINDOWS][EArchitectureType.AMD],
        );
      } catch (err: unknown) {
        console.error(err);
      } finally {
        this.downloadingNewWindowsAmd64Cli = false;
        this.downloadProgress[EOperatingSystems.WINDOWS][EArchitectureType.AMD].value = 0;
      }
    },
    async downloadNewCliWindowsArm64(): Promise<void> {
      if (this.downloadingNewWindowsArm64Cli) return;
      try {
        this.downloadingNewWindowsArm64Cli = true;
        await cliService.getBinaryWithProgress(
          EOperatingSystems.WINDOWS,
          EArchitectureType.ARM,
          this.downloadProgress[EOperatingSystems.WINDOWS][EArchitectureType.ARM],
        );
      } catch (err: unknown) {
        console.error(err);
      } finally {
        this.downloadingNewWindowsArm64Cli = false;
        this.downloadProgress[EOperatingSystems.WINDOWS][EArchitectureType.ARM].value = 0;
      }
    },
    async downloadUnixInstallerFile(archi: EArchitectureType): Promise<void> {
      if (
        (archi === EArchitectureType.AMD && this.downloadingNewAmd64Cli) ||
        (archi === EArchitectureType.ARM && this.downloadingNewArm64Cli)
      )
        return;
      try {
        archi === EArchitectureType.AMD ? (this.downloadingNewAmd64Cli = true) : (this.downloadingNewArm64Cli = true);
        await cliService.getBinaryWithProgress(this.selectedOS, archi, this.downloadProgress[archi]);
      } catch (err: unknown) {
        console.error(err);
      } finally {
        archi === EArchitectureType.AMD ? (this.downloadingNewAmd64Cli = false) : (this.downloadingNewArm64Cli = false);
        this.downloadProgress[archi].value = 0;
      }
    },
  },
});
</script>

<style lang="scss" scoped>
.help-section-italic,
.more-info {
  font-style: italic;
}
.seperator-with-text {
  max-width: 50%;
}
.optional {
  color: $black-54;
  font-style: italic;
}
</style>
