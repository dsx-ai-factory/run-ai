<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-form-wrapper ref="formRef" class="ai-application-new column items-center q-pt-md" :form-state="formState">
    <runai-expansion-wrapper>
      <runai-cluster-section
        prevent-top-padding-when-open
        entity-name="AI application"
        :cluster-id="formState.clusterId"
        :disabled-cluster-options="disabledClusterOptions"
        @update:cluster-id="updateClusterId"
      />
      <transition-group name="fade">
        <template v-if="formState.clusterId && isClusterSectionValid">
          <project-section
            aid="project-section"
            entity-type="AI application"
            :loading="loadingProject"
            :selected-project-id="numericProjectId"
            :projects="loadedProjects"
            :cluster-id="formState.clusterId"
            @project-changed="handleSelectedProject"
          />

          <workload-name-section
            entity-type="AI application"
            v-model:name="formState.name"
            :cluster-id="formState.clusterId"
            :project-id="numericProjectId"
            default-opened
          />

          <ai-application-source-section
            v-model:chart-ref="formState.spec.chartRef"
            v-model:auth="formState.spec.auth"
            v-model:yaml-content="yamlContent"
            :cluster-id="formState.clusterId"
            :project-id="numericProjectId"
          />
        </template>
      </transition-group>
    </runai-expansion-wrapper>
    <div class="row items-center">
      <q-field class="col-4 form-hint no-padding" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
      <div class="buttons q-ml-auto">
        <runai-btn
          class="q-mr-md"
          aid="form-cancel-ai-application-btn"
          @click="onCancel"
          flat
          color="primary"
          label="cancel"
        />
        <runai-btn
          type="submit"
          color="primary"
          aid="form-create-ai-application-btn"
          @click="save"
          :loading="submitting"
          label="create AI application"
        />
      </div>
    </div>
  </runai-form-wrapper>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";
import type { QForm } from "quasar";
import { useQuasar } from "quasar";

import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";

import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { ProjectSection } from "@/components/section/project-section";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { AiApplicationSourceSection } from "@/components/ai-application/ai-application-source-section";
import { RunaiBtn } from "@/components/common/runai-btn";

import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { aiApplicationService } from "@/services/cluster/ai-application.service/ai-application.service";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { urlService } from "@/services/url.service/url.service";

import { alertUtil } from "@/utils/alert.util";
import { yamlUtil } from "@/utils/yaml.util/yaml.util";

import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { AIApplicationCreationRequest } from "@/swagger-models/workloads-client";
import router from "@/router";
import { AI_APPLICATION_ROUTE_NAMES } from "@/router/ai-application.routes/ai-application.routes.names";
import { deepCopy } from "@/utils/common.util";
import { MIN_CLUSTER_VERSION_FOR_CREATING_AI_APPLICATIONS } from "@/common/version.constant";

const route = useRoute();
const $q = useQuasar();
const appStore = useAppStore();
const clusterStore = useClusterStore();
const authStore = useAuthStore();

const formRef = ref<QForm | null>(null);
const formState = ref<AIApplicationCreationRequest>({
  clusterId: "",
  projectId: "",
  name: "",
  spec: {
    chartRef: {
      repo: "",
      chartName: "",
      chartVersion: "",
      chartURL: "",
    },
  },
});
const yamlContent = ref("");
const numericProjectId = computed(() => (formState.value.projectId ? Number(formState.value.projectId) : null));

const loadedProjects = ref<Array<ProjectMetadata>>([]);
const loadingProject = ref(false);
const submitting = ref(false);
const displayFormHint = ref(false);
const timeOutId = ref<ReturnType<typeof setTimeout> | null>(null);
const isClusterSectionValid = ref(false);
const actionPermissionsByScopes = ref<Record<Action, PermittedScopes>>({} as Record<Action, PermittedScopes>);

function getDefaultClusterId(): string {
  if (clusterStore.clusterList.length === 1) {
    return clusterStore.clusterList[0]?.uuid ?? "";
  }
  const queryClusterId = route.query.clusterId;
  if (queryClusterId) return String(queryClusterId);
  return "";
}

function findClusterById(id: string): ClustersMinimalData | undefined {
  return clusterStore.clusterList.find((c) => c.uuid === id);
}

const disabledClusterOptions = computed(() => {
  const disabled: Record<string, string> = {};
  for (const cluster of clusterStore.clusterList) {
    const isSupported = clusterStore.isClusterVersionSufficient(
      cluster.uuid,
      MIN_CLUSTER_VERSION_FOR_CREATING_AI_APPLICATIONS,
    );
    if (!isSupported) {
      disabled[cluster.uuid] =
        `This cluster can’t be selected because AI applications are supported from cluster version 2.25 or higher`;
    }
  }
  return disabled;
});

function canSubmitWorkload(clusterUuid: string): boolean {
  if (!clusterUuid) return false;
  const cluster = findClusterById(clusterUuid);
  if (!cluster) return false;
  return workloadService.canSubmitWorkload(cluster) === null;
}

async function loadActionPermissions(): Promise<void> {
  if (Object.keys(actionPermissionsByScopes.value).length > 0) return;
  actionPermissionsByScopes.value = await permissionService.getPermittedScopesForAction(ResourceType.Workloads);
}

async function loadProjects(): Promise<void> {
  try {
    loadingProject.value = true;
    loadedProjects.value = await orgUnitService.getProjectsMetadata(
      formState.value.clusterId,
      actionPermissionsByScopes.value[Action.Create],
      authStore.orgUnitList,
    );

    if (loadedProjects.value.length === 1) {
      await handleSelectedProject(+loadedProjects.value[0].id);
    }
  } catch (error: unknown) {
    $q.notify(alertUtil.getError("Failed to load projects"));
    console.error(error);
    appStore.setFallback(true);
  } finally {
    loadingProject.value = false;
  }
}

async function handleSelectedProject(projectId: number | null): Promise<void> {
  formState.value.projectId = projectId != null && projectId !== -1 ? String(projectId) : "";
}

async function updateClusterId(id: string, payload: { disabled: boolean }): Promise<void> {
  formState.value.clusterId = id;
  isClusterSectionValid.value = !payload.disabled && canSubmitWorkload(id);
  formState.value.projectId = "";
  urlService.updateQueryParams({ clusterId: id, createdEntityId: null });
  if (isClusterSectionValid.value) {
    await loadProjects();
  }
}

async function handleClusterDependencies(selectedDefaultClusterId: string): Promise<void> {
  if (!selectedDefaultClusterId) return;
  if (canSubmitWorkload(selectedDefaultClusterId)) {
    await updateClusterId(selectedDefaultClusterId, { disabled: false });
  }
}

async function loadInitialState(): Promise<void> {
  await loadActionPermissions();
  const defaultClusterId = getDefaultClusterId();
  await handleClusterDependencies(defaultClusterId);

  const createdEntityId = route.query.createdEntityId?.toString();
  if (createdEntityId) {
    await handleSelectedProject(Number(createdEntityId));
  }
  appStore.setPageLoading(false);
}
loadInitialState();

function isFormIncomplete(val: boolean): boolean | string {
  return !val ? true : "Please review and fix the issues in the form";
}

function showHint(): void {
  displayFormHint.value = true;
  timeOutId.value && clearTimeout(timeOutId.value);
  timeOutId.value = setTimeout(() => (displayFormHint.value = false), 15000);
}

async function validate(): Promise<boolean> {
  const isValid: boolean = await (formRef.value as QForm)?.validate();
  return isValid;
}

function onCancel(): void {
  router.push({ name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX });
}

function buildCreationRequest(): AIApplicationCreationRequest {
  const request: AIApplicationCreationRequest = deepCopy(formState.value);

  if (yamlContent.value) {
    const values = yamlUtil.yamlToObject(yamlContent.value) as { [key: string]: unknown } | null;
    if (values) {
      request.spec.values = values;
    }
  }

  return request;
}

async function save(): Promise<void> {
  submitting.value = true;
  displayFormHint.value = false;
  try {
    const isValid = await validate();
    if (!isValid) {
      showHint();
      return;
    }

    const request = buildCreationRequest();
    await aiApplicationService.createApplication(request);
    $q.notify(alertUtil.getSuccess(`AI application ${formState.value.name} created`));
    router.push({ name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX });
  } catch (error: unknown) {
    console.error(error);
    $q.notify(alertUtil.getError("Failed to create AI application"));
  } finally {
    submitting.value = false;
  }
}

onUnmounted(() => {
  timeOutId.value && clearTimeout(timeOutId.value);
});
</script>

<style scoped lang="scss">
.ai-application-new {
  width: 100%;

  .runai-expansion-item {
    width: 800px;
  }
}
</style>
