<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-yaml-tab column no-wrap">
    <header class="yaml-actions row justify-between items-center q-px-md q-py-sm">
      <runai-select
        v-model="selectedTargetKey"
        :options="targetOptions"
        emit-value
        map-options
        dense
        outlined
        hide-bottom-space
        options-dense
        selected-ellipsis
        :label="OBJECT_SELECT_LABEL"
        aid="workload-yaml-tab-target-select"
        class="target-select"
      />
      <div class="row items-center">
        <runai-copy-to-clipboard-btn
          v-if="yamlText"
          :text-to-copy="yamlText"
          aid="workload-yaml-tab-copy"
          icon="ph-bold ph-copy"
          size="12px"
          tooltip-anchor="center right"
          tooltip-self="center left"
        />
        <runai-btn
          icon-button
          size="12px"
          icon="ph-bold ph-download"
          aid="workload-yaml-tab-download"
          :disable="!yamlText"
          :tooltip="DOWNLOAD_TOOLTIP"
          tooltip-anchor="center right"
          tooltip-self="center left"
          :aria-label="DOWNLOAD_TOOLTIP"
          @click="onDownload"
        />
      </div>
    </header>
    <section class="yaml-body">
      <runai-default-loader v-if="loading" />
      <runai-drawer-no-data
        v-else-if="noData"
        :messages="noData.messages"
        :icon-name="noData.icon"
        wrapper-padding-class="q-pt-lg"
      />
      <runai-editor v-else read-only :code="yamlText" height="100%" />
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import yamlLib from "js-yaml";

import { type Phase, type Pod, type WorkloadElement, type Workload } from "@/swagger-models/workloads-client";
import { clusterApi } from "@/services/infra/client-apis/cluster-api/cluster-api";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { phaseColors } from "@/utils/workload.util/workload-chart.util";
import { createDownloadLink } from "@/utils/common.util";
import { NotFoundError } from "@/models/http-response.model";
import type { ISelectOption } from "@/models/global.model";
import {
  EYamlTabView,
  ETargetKind,
  KEY_SEPARATOR,
  OBJECT_SELECT_LABEL,
  DOWNLOAD_TOOLTIP,
  WORKLOAD_OPTION_LABEL,
  POD_OPTION_PREFIX,
  DEFAULT_ELEMENT_KIND,
  LABEL_SEPARATOR,
  DEFAULT_FILENAME,
  YAML_FILE_EXTENSION,
  YAML_MIME_TYPE,
} from "./workload-yaml-tab.model";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data/index";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader/index";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiEditor } from "@/components/common/runai-editor";

export interface IProps {
  workload?: Workload | null;
  view?: EYamlTabView;
  // Expanded-view only: the structure-map's selected element. The dropdown is
  // scoped to it and the active view jumps to it on change. null = workload root.
  scopeElementId?: string | null;
}

const props = withDefaults(defineProps<IProps>(), {
  view: EYamlTabView.Collapsed,
  scopeElementId: null,
});

interface ITargetOption extends Omit<ISelectOption, "value"> {
  value: string;
  filenameBase?: string;
}

interface ITargetSelection {
  kind: ETargetKind;
  id?: string;
}

const loading = ref(false);
const noData = ref<{ messages: string[]; icon: string } | null>(null);
const k8sObject = ref<Record<string, unknown> | null>(null);
const pods = ref<Pod[]>([]);
const elements = ref<WorkloadElement[]>([]);

let k8sObjectRequestId = 0;
let childrenRequestId = 0;

// Active YAML target. Dropdown picks change this freely; a scope change syncs it.
const selectedTargetKey = ref<string>(ETargetKind.Workload);

const elementById = computed<Map<string, WorkloadElement>>(
  () => new Map(elements.value.map((element) => [element.id, element])),
);

const buildTargetKey = (kind: ETargetKind, id?: string): string => (id ? `${kind}${KEY_SEPARATOR}${id}` : kind);

const buildPodOption = (pod: Pod): ITargetOption => ({
  value: buildTargetKey(ETargetKind.Pod, pod.id),
  label: `${POD_OPTION_PREFIX}${LABEL_SEPARATOR}${pod.name}`,
  filenameBase: pod.name,
  subLabel: pod.k8sPhase,
  subLabelColor: phaseColors[pod.k8sPhase as Phase] ?? "",
});

const buildElementOption = (element: WorkloadElement): ITargetOption => {
  const display = element.kubernetesRef?.name ?? element.name ?? element.id;
  const kind = element.kubernetesRef?.gvk?.kind ?? DEFAULT_ELEMENT_KIND;
  return {
    value: buildTargetKey(ETargetKind.Element, element.id),
    label: `${kind}${LABEL_SEPARATOR}${display}`,
    filenameBase: display,
  };
};

const workloadOption: ITargetOption = { value: ETargetKind.Workload, label: WORKLOAD_OPTION_LABEL };

const targetOptions = computed<ITargetOption[]>(() => {
  const concreteElements = elements.value.filter((e) => e.kubernetesRef?.gvk?.kind);

  if (props.view === EYamlTabView.Collapsed) {
    return [workloadOption, ...pods.value.map(buildPodOption)];
  }

  const scopeId = props.scopeElementId;
  if (scopeId === null) {
    return [workloadOption, ...concreteElements.map(buildElementOption), ...pods.value.map(buildPodOption)];
  }

  const scopeElement = elementById.value.get(scopeId);
  const options: ITargetOption[] = [];
  if (scopeElement?.kubernetesRef?.gvk?.kind) options.push(buildElementOption(scopeElement));
  options.push(
    ...concreteElements.filter((e) => e.id !== scopeId && isDescendantElement(e.id, scopeId)).map(buildElementOption),
  );
  options.push(
    ...pods.value.filter((p) => p.elements?.some((entry) => entry.elementId === scopeId)).map(buildPodOption),
  );
  return options;
});

const isDescendantElement = (candidateId: string, ancestorId: string): boolean => {
  let currentId: string | null | undefined = candidateId;
  for (let i = 0; i < elements.value.length && currentId; i++) {
    if (currentId === ancestorId) return true;
    currentId = elementById.value.get(currentId)?.parentId;
  }
  return false;
};

const selectedTarget = computed<ITargetOption | undefined>(() =>
  targetOptions.value.find((option) => option.value === selectedTargetKey.value),
);

const yamlText = computed<string>(() => (k8sObject.value ? yamlLib.dump(k8sObject.value) : ""));

const downloadFilename = computed<string>(() => {
  const baseName = selectedTarget.value?.filenameBase ?? props.workload?.name ?? DEFAULT_FILENAME;
  return `${baseName}${YAML_FILE_EXTENSION}`;
});

const fetchChildren = async (workloadId: string): Promise<void> => {
  const requestId = ++childrenRequestId;
  const [podsResult, elementsResult] = await Promise.allSettled([
    workloadService.getWorkloadPods(workloadId, false, undefined, undefined, undefined, true),
    workloadService.getWorkloadElements(workloadId),
  ]);
  if (requestId !== childrenRequestId) return;
  if (podsResult.status === "fulfilled") {
    pods.value = podsResult.value;
  } else {
    pods.value = [];
    console.warn("YAML selector: failed to load pods (non-fatal)", podsResult.reason);
  }
  if (elementsResult.status === "fulfilled") {
    elements.value = elementsResult.value;
  } else {
    elements.value = [];
    console.warn("YAML selector: failed to load elements (non-fatal)", elementsResult.reason);
  }
};

const parseTarget = (key: string): ITargetSelection => {
  const separatorIndex = key.indexOf(KEY_SEPARATOR);
  if (separatorIndex === -1) return { kind: key as ETargetKind };
  return { kind: key.slice(0, separatorIndex) as ETargetKind, id: key.slice(separatorIndex + 1) };
};

const NO_YAML_STATE = { messages: ["There is no YAML yet."], icon: "ph-bold ph-file" };
const YAML_UNAVAILABLE_STATE = {
  messages: ["YAML isn't available right now.", "If this issue persists, contact your administrator."],
  icon: "ph-bold ph-file-minus",
};

const fetchK8sObject = async (workloadId: string, target: ITargetSelection): Promise<void> => {
  const requestId = ++k8sObjectRequestId;
  loading.value = true;
  noData.value = null;
  k8sObject.value = null;
  try {
    let response;
    if (target.kind === ETargetKind.Pod && target.id) {
      response = await clusterApi.k8sObjectApi.getPodObject(workloadId, target.id);
    } else if (target.kind === ETargetKind.Element && target.id) {
      response = await clusterApi.k8sObjectApi.getElementObject(workloadId, target.id);
    } else {
      response = await clusterApi.k8sObjectApi.getWorkloadObject(workloadId);
    }
    if (requestId !== k8sObjectRequestId) return;
    k8sObject.value = (response.data?.object ?? null) as Record<string, unknown> | null;
    if (!k8sObject.value) {
      noData.value = NO_YAML_STATE;
    }
  } catch (err) {
    if (requestId !== k8sObjectRequestId) return;
    const status = (err as { response?: { status?: number } })?.response?.status;
    noData.value = status === NotFoundError.statusCode ? NO_YAML_STATE : YAML_UNAVAILABLE_STATE;
    console.warn(`YAML tab: failed to fetch ${target.kind} object`, (err as Error)?.message ?? err);
  } finally {
    if (requestId === k8sObjectRequestId) loading.value = false;
  }
};

const onDownload = (): void => {
  if (!yamlText.value) return;
  const blob = new Blob([yamlText.value], { type: YAML_MIME_TYPE });
  createDownloadLink(blob, downloadFilename.value);
};

watch(
  () => props.workload?.id,
  (newId) => {
    pods.value = [];
    elements.value = [];
    selectedTargetKey.value = ETargetKind.Workload;
    if (!newId) {
      k8sObject.value = null;
      noData.value = null;
      return;
    }
    void fetchChildren(newId);
    void fetchK8sObject(newId, { kind: ETargetKind.Workload });
  },
  { immediate: true },
);

watch(selectedTargetKey, (key) => {
  const id = props.workload?.id;
  if (!id) return;
  void fetchK8sObject(id, parseTarget(key));
});

const defaultTargetForScope = (scopeId: string | null): string => {
  if (!scopeId) return ETargetKind.Workload;
  const scopeElement = elementById.value.get(scopeId);
  if (scopeElement?.kubernetesRef?.gvk?.kind) return buildTargetKey(ETargetKind.Element, scopeId);
  return targetOptions.value[0]?.value ?? ETargetKind.Workload;
};

watch(
  () => props.scopeElementId,
  (scopeId) => {
    selectedTargetKey.value = defaultTargetForScope(scopeId);
  },
);

watch([elements, pods], () => {
  if (props.view !== EYamlTabView.Expanded) return;
  if (targetOptions.value.some((o) => o.value === selectedTargetKey.value)) return;
  selectedTargetKey.value = defaultTargetForScope(props.scopeElementId);
});
</script>

<style scoped lang="scss">
.workload-yaml-tab {
  height: 100%;
  width: 100%;
}

.yaml-actions {
  background-color: var(--bg-surface);
  border-bottom: 1px solid var(--border-default);
  gap: var(--space-xs);
}

.target-select {
  // Fixed width so the field doesn't resize with the selected name — matches
  // the logs-tab pod selector.
  width: 280px;
}

.yaml-body {
  flex: 1;
  overflow: hidden;
  background-color: var(--bg-code-surface);
  color: var(--text-inverse);
}

/* runai-editor dims read-only mode to 0.8 opacity (policy-form UX); the YAML
 * tab is purely a viewer — full opacity reads better. */
.yaml-body :deep(.readOnly) {
  opacity: 1;
}
</style>
