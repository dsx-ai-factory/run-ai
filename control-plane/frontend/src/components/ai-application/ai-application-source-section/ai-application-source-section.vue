<template>
  <runai-expansion-item class="ai-application-source-section" label="Source" default-opened>
    <runai-radio-options
      title="Select AI application source"
      :options="aiApplicationSourceOptions"
      :model-value="selectedSource"
      @update:model-value="handleSourceChange"
    />

    <template v-if="selectedSource === 'ngc-catalog'">
      <runai-section class="q-mt-lg" title="Set the Helm chart source">
        <div class="row items-center">
          <runai-select
            class="col-4 q-pr-sm"
            dense
            label="Repository"
            stack-label
            :options="repoOptions"
            emit-value
            :loading="loadingRepos"
            :model-value="props.chartRef.repo"
            @update:model-value="handleRepoChange"
            placeholder="Select a repository"
            :rules="[selectRepositoryRule]"
          />
        </div>
      </runai-section>

      <runai-section v-if="selectedRepo?.isPrivate" class="q-mt-md" title="Set repository credentials">
        <runai-credential-selector-by-source
          :key="credentialSelectorKey"
          v-model="credentialSelectorModel"
          :cluster-id="props.clusterId"
          :scope-model="scopeModel"
          :source-select-config="sharedSecretSourceSelectConfig"
          :type-select-config="ngcApiKeyTypeSelectConfig"
          :credential-select-config="{ showCreateNewBtn: true }"
          @update:model-value="handleCredentialChange"
        />
      </runai-section>

      <runai-section v-if="canBrowseSelectedRepo" class="q-mt-md">
        <div class="row items-center q-gutter-x-md">
          <runai-select
            dense
            class="col q-pr-sm"
            label="Chart name"
            stack-label
            :options="chartNameOptions"
            no-option-text="No charts found"
            emit-value
            :loading="loadingCharts"
            :model-value="props.chartRef.chartName"
            @update:model-value="handleChartNameChange"
            placeholder="Select a chart"
            :rules="[validationRules.validateNameSelected]"
          />
          <runai-select
            dense
            class="col-4"
            label="Chart version"
            stack-label
            :options="chartVersionOptions"
            no-option-text="No versions found"
            emit-value
            :loading="loadingCharts"
            :disable="!props.chartRef.chartName"
            :model-value="props.chartRef.chartVersion"
            @update:model-value="(val: string) => updateChartRef({ chartVersion: val })"
            placeholder="Select a version"
            :rules="[selectVersionRule]"
          />
        </div>
      </runai-section>
    </template>

    <runai-section
      v-else-if="selectedSource === 'custom-url'"
      class="q-mt-lg"
      title="Set the URL of your application's Helm chart"
    >
      <q-input
        dense
        label="Chart URL"
        :rules="[validationRules.validateChartUrlFormat]"
        no-error-icon
        :model-value="props.chartRef.chartURL"
        @update:model-value="updateChartRef({ chartURL: String($event ?? '') })"
        placeholder="e.g., https://helm.ngc.nvidia.com/nvidia/blueprint/charts/nvidia-blueprint-rag-v2.3.0.tgz "
      />
    </runai-section>

    <runai-section
      v-else-if="selectedSource === 'upload-file'"
      class="q-mt-lg"
      title="Upload a packaged Helm chart (.tgz)"
    >
      <ai-application-chart-upload-field
        :selected-file="selectedFile"
        :chart-id="uploadedChartId"
        :uploading="uploadingChart"
        :upload-error="chartUploadError"
        @update:selected-file="handleChartFileSelected"
        @rejected="handleChartFileRejected"
      />
    </runai-section>

    <runai-section
      class="q-mt-md"
      title="Set application overrides"
      tooltip-text="To override specific parameters in your Helm chart, enter the corresponding key-value pairs in YAML format. These will replace the defaults in your values.yaml."
    >
      <runai-editor
        ref="yamlEditorRef"
        :code="props.yamlContent"
        :height="editorHeight"
        show-expand-button
        :is-expanded="isExpanded"
        show-placeholder
        placeholder="Paste your YAML content here"
        @changed="(value: string) => emit('update:yamlContent', value)"
        @validation-status="onValidationStatusChanged"
        @toggle-expand="toggleExpand"
        @pasted="(value: string) => emit('update:yamlContent', value)"
      />
    </runai-section>
  </runai-expansion-item>
</template>

<script setup lang="ts">
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiEditor } from "@/components/common/runai-editor";
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
import AiApplicationChartUploadField from "./ai-application-chart-upload-field.vue";
import { computed, onMounted, ref, watch } from "vue";
import { useQuasar } from "quasar";
import type { IScopeModel, ISelectOption } from "@/models/global.model";
import type { AIApplicationAuth, AIApplicationSpecChartRef } from "@/swagger-models/workloads-client";
import type { HelmRepo, HelmChart } from "@/swagger-models/catalog-service-client";
import { validationRules } from "@/common/validation-rules";
import { errorMessages } from "@/common/error-message.constant";
import { catalogService } from "@/services/control-plane/catalog.service/catalog.service";
import { aiApplicationService } from "@/services/cluster/ai-application.service/ai-application.service";
import { alertUtil } from "@/utils/alert.util/alert.util";
import { HttpErrorResponse } from "@/models/http-response.model";
import {
  ERunaiCredentialSource,
  ERunaiCredentialType,
  type ICredentialSelectorModel,
  type ISourceSelectConfig,
  type ITypeSelectConfig,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

const props = withDefaults(
  defineProps<{
    chartRef: AIApplicationSpecChartRef;
    auth?: AIApplicationAuth;
    yamlContent: string;
    clusterId: string;
    projectId: number | null;
  }>(),
  {
    chartRef: () => ({ repo: "", chartName: "", chartVersion: "", chartURL: "" }),
    yamlContent: "",
    clusterId: "",
    projectId: null,
  },
);

const $q = useQuasar();

const emit = defineEmits<{
  "update:chartRef": [value: AIApplicationSpecChartRef];
  "update:auth": [value: AIApplicationAuth | undefined];
  "update:yamlContent": [value: string];
}>();

function updateChartRef(partial: Partial<AIApplicationSpecChartRef>): void {
  emit("update:chartRef", { ...props.chartRef, ...partial });
}

function updateAuth(assetId?: string): void {
  emit("update:auth", assetId ? { assetId } : undefined);
}

const selectedSource = ref("ngc-catalog");

const aiApplicationSourceOptions = [
  {
    label: "NGC catalog",
    value: "ngc-catalog",
    toolTip:
      "GPU-optimized containers, pre-trained models, and Helm charts will be pulled from the NVIDIA GPU Cloud (NGC). This is recommend for standard NVIDIA-certified AI stacks and enterprise-ready models.",
  },
  {
    label: "Custom URL",
    value: "custom-url",
    toolTip:
      "Provide a link to a Helm chart or container hosted on a private or public repository (e.g., GitHub, S3, or a private registry). Use this for proprietary applications or customized versions of existing charts.",
  },
  {
    label: "Upload file",
    value: "upload-file",
    toolTip:
      "Upload a packaged Helm chart (.tgz) from your computer. Use this to deploy a chart you built locally without hosting it in a repository or registry.",
  },
];

// Repo/chart state: repo select uses URL (for AIApplicationSpecChartRef); charts are keyed by repo URL.
const loadingRepos = ref(false);
const loadingCharts = ref(false);
const chartLoadError = ref("");
const repoOptions = ref<ISelectOption[]>([]);
const helmRepos = ref<HelmRepo[]>([]);
const repoChartsMap = ref(new Map<string, HelmChart[]>());
const loadingChartKeys = new Set<string>();
const credentialSelectorModel = ref<ICredentialSelectorModel>({
  source: ERunaiCredentialSource.SharedSecret,
  type: ERunaiCredentialType.NgcApiKey,
  credentialName: "",
  id: undefined,
});

const sharedSecretSourceSelectConfig: ISourceSelectConfig = {
  disable: true,
  showPersonalCredentialsOption: false,
};

const ngcApiKeyTypeSelectConfig: ITypeSelectConfig = {
  hidden: false,
  allowedTypes: [ERunaiCredentialType.NgcApiKey],
  disable: true,
};

const selectedRepo = computed(() => helmRepos.value.find((r) => r.url === props.chartRef.repo));
const selectedCredentialId = computed(() => credentialSelectorModel.value.id);
// Remount the selector when scope or repo changes so it reloads scoped credentials and drops internal selection state.
const credentialSelectorKey = computed(() => `${props.clusterId}:${props.projectId ?? ""}:${props.chartRef.repo ?? ""}`);
const canBrowseSelectedRepo = computed(() => {
  if (!props.chartRef.repo) return false;
  if (!selectedRepo.value?.isPrivate) return true;
  return !!selectedCredentialId.value;
});

const scopeModel = computed<IScopeModel>(() => ({
  scope: null,
  projectId: props.projectId,
  clusterId: props.clusterId || undefined,
}));

function getEmptyCredentialSelectorModel(): ICredentialSelectorModel {
  return {
    source: ERunaiCredentialSource.SharedSecret,
    type: ERunaiCredentialType.NgcApiKey,
    credentialName: "",
    id: undefined,
  };
}

function chartsCacheKey(repoUrl: string, assetId?: string): string {
  return assetId ? `${repoUrl}:${assetId}` : repoUrl;
}

function getChartLoadErrorMessage(error: unknown): string {
  if (error instanceof HttpErrorResponse) {
    return error.message;
  }
  return "Failed to load charts from the selected repository";
}

async function loadRepos(): Promise<void> {
  loadingRepos.value = true;
  try {
    const repos = await catalogService.getHelmRepos();
    helmRepos.value = repos;
    repoOptions.value = repos.map((r) => ({ label: r.name ?? r.url, value: r.url }));
    const firstRepo = repos[0];
    if (firstRepo?.url) {
      updateChartRef({ repo: firstRepo.url, chartName: "", chartVersion: "" });
      if (!firstRepo.isPrivate) {
        await loadRepoCharts(firstRepo.url);
      }
    }
  } finally {
    loadingRepos.value = false;
  }
}

async function loadRepoCharts(repoUrl: string, assetId?: string): Promise<void> {
  const cacheKey = chartsCacheKey(repoUrl, assetId);
  if (repoChartsMap.value.has(cacheKey) || loadingChartKeys.has(cacheKey)) return;
  const repo = helmRepos.value.find((r) => r.url === repoUrl);
  if (!repo?.id) return;
  loadingChartKeys.add(cacheKey);
  loadingCharts.value = true;
  chartLoadError.value = "";
  try {
    const { charts } = await catalogService.browseHelmRepoCharts(repo.id, assetId);
    repoChartsMap.value.set(cacheKey, charts);
  } catch (error: unknown) {
    chartLoadError.value = getChartLoadErrorMessage(error);
    $q.notify(alertUtil.getError(chartLoadError.value));
  } finally {
    loadingChartKeys.delete(cacheKey);
    loadingCharts.value = false;
  }
}

function selectRepositoryRule(value: string): string | boolean {
  return validationRules.validateValueSelected(value) || errorMessages.SELECT_A_REPOSITORY;
}
function selectVersionRule(value: string): string | boolean {
  return validationRules.validateValueSelected(value) || errorMessages.SELECT_VERSION;
}

// Upload-file state and the upload call live here so switching source away and back does not force a
// re-upload. The child field is presentational: it emits the picked file and renders the state below.
const selectedFile = ref<File | null>(null);
const uploadedChartId = ref("");
const uploadingChart = ref(false);
const chartUploadError = ref("");

async function handleChartFileSelected(file: File | null): Promise<void> {
  chartUploadError.value = "";
  selectedFile.value = file;
  syncUploadedChartId("");

  if (!file) {
    return;
  }

  uploadingChart.value = true;
  try {
    syncUploadedChartId(await aiApplicationService.uploadChart(file));
  } catch (error: unknown) {
    chartUploadError.value = error instanceof HttpErrorResponse ? error.message : errorMessages.CHART_UPLOAD_FAILED;
  } finally {
    uploadingChart.value = false;
  }
}

function handleChartFileRejected(rejections: { failedPropValidation: string }[]): void {
  if (rejections.some((rejection) => rejection.failedPropValidation === "max-file-size")) {
    chartUploadError.value = errorMessages.CHART_FILE_TOO_LARGE;
  }
}

function syncUploadedChartId(chartId: string): void {
  uploadedChartId.value = chartId;
  if (selectedSource.value === "upload-file") {
    updateChartRef({ chartId });
  }
}

const chartNameOptions = computed<ISelectOption[]>(() => {
  const charts = repoChartsMap.value.get(chartsCacheKey(props.chartRef.repo ?? "", selectedCredentialId.value));
  if (!charts) return [];
  return charts.map((c) => ({ label: c.name, value: c.name }));
});

const chartVersionOptions = computed<ISelectOption[]>(() => {
  const charts = repoChartsMap.value.get(chartsCacheKey(props.chartRef.repo ?? "", selectedCredentialId.value));
  const chartName = props.chartRef.chartName;
  if (!charts || !chartName) return [];
  const chart = charts.find((c) => c.name === chartName);
  if (!chart) return [];
  return chart.versions.map((v) => ({ label: v.version, value: v.version }));
});

function handleSourceChange(source: string): void {
  selectedSource.value = source;
  credentialSelectorModel.value = getEmptyCredentialSelectorModel();
  chartLoadError.value = "";
  chartUploadError.value = "";
  updateAuth();
  // Project only the active source into chartRef. The uploaded file/chartId stay in local state so
  // returning to "Upload file" restores the selection without re-uploading.
  const restoredChartId = source === "upload-file" ? uploadedChartId.value : "";
  updateChartRef({ repo: "", chartName: "", chartVersion: "", chartURL: "", chartId: restoredChartId });
}

async function handleRepoChange(repo: string): Promise<void> {
  credentialSelectorModel.value = getEmptyCredentialSelectorModel();
  chartLoadError.value = "";
  updateAuth();
  updateChartRef({ repo, chartName: "", chartVersion: "" });
  const selectedHelmRepo = helmRepos.value.find((r) => r.url === repo);
  if (repo && !selectedHelmRepo?.isPrivate) {
    await loadRepoCharts(repo);
  }
}

async function handleCredentialChange(model: ICredentialSelectorModel): Promise<void> {
  credentialSelectorModel.value = model;
  chartLoadError.value = "";
  updateAuth(model.id);
  updateChartRef({ chartName: "", chartVersion: "" });
  if (props.chartRef.repo && model.id) {
    await loadRepoCharts(props.chartRef.repo, model.id);
  }
}

function handleChartNameChange(chartName: string): void {
  updateChartRef({ chartName, chartVersion: "" });
}

watch([() => props.clusterId, () => props.projectId], () => {
  credentialSelectorModel.value = getEmptyCredentialSelectorModel();
  chartLoadError.value = "";
  updateAuth();
  if (selectedRepo.value?.isPrivate) {
    updateChartRef({ chartName: "", chartVersion: "" });
  }
});

onMounted(() => {
  loadRepos();
});

const isExpanded = ref(false);
const isYamlValid = ref(false);
const MIN_HEIGHT = "300px";
const MAX_HEIGHT = "700px";

const onValidationStatusChanged = (value: { isValid: boolean; error: string }) => {
  isYamlValid.value = value.isValid;
};
const toggleExpand = () => {
  isExpanded.value = !isExpanded.value;
};
const editorHeight = computed(() => (isExpanded.value ? MAX_HEIGHT : MIN_HEIGHT));
</script>

<style scoped lang="scss"></style>
