import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import { INFERENCE_TEMPLATE_ROUTE_NAMES } from "./inference-template.routes.names";
import WorkloadTemplateInferenceNew from "@/views/workload-template/workload-template-hybrid/inference/template-inference-new.vue";
import WorkloadTemplateInferenceHuggingFace from "@/views/workload-template/workload-template-hybrid/inference/template-inference-hugging-face.vue";
import WorkloadTemplateInferenceNim from "@/views/workload-template/workload-template-hybrid/inference/template-inference-nim.vue";
import WorkloadTemplateInferenceCustom from "@/views/workload-template/workload-template-hybrid/inference/template-inference-custom.vue";

export const inferenceTemplateRoutes: RunaiRouteRecord[] = [
  {
    path: "inference",
    name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW,
    component: WorkloadTemplateInferenceNew,
    meta: {
      pageTitle: "New inference template",
      fullWidth: true,
      requestToLeave: true,
      backPageName: "workload-template-index",
      resourceType: ResourceType.Templates,
      minPermittedActions: [Action.Create, Action.Update],
      requiresAuth: true,
      parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
    },
  },
  {
    path: "inference/tgi/new",
    name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_TGI,
    component: WorkloadTemplateInferenceHuggingFace,
    meta: {
      pageTitle: "New inference template",
      fullWidth: true,
      requestToLeave: true,
      backPageName: "workload-template-index",
      resourceType: ResourceType.Templates,
      minPermittedActions: [Action.Create, Action.Update],
      requiresAuth: true,
      parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
    },
  },
  {
    path: "inference/vllm/new",
    name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_VLLM,
    component: WorkloadTemplateInferenceHuggingFace,
    meta: {
      pageTitle: "New inference template",
      fullWidth: true,
      requestToLeave: true,
      backPageName: "workload-template-index",
      resourceType: ResourceType.Templates,
      minPermittedActions: [Action.Create, Action.Update],
      requiresAuth: true,
      parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
    },
  },
  {
    path: "inference/nim/new",
    name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NIM,
    component: WorkloadTemplateInferenceNim,
    meta: {
      pageTitle: "New inference template",
      fullWidth: true,
      requestToLeave: true,
      backPageName: "workload-template-index",
      resourceType: ResourceType.Templates,
      minPermittedActions: [Action.Create, Action.Update],
      requiresAuth: true,
      parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
    },
  },
  {
    path: "inference/custom/new",
    name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_NEW,
    component: WorkloadTemplateInferenceCustom,
    meta: {
      pageTitle: "New inference template",
      fullWidth: true,
      requestToLeave: true,
      backPageName: "workload-template-index",
      resourceType: ResourceType.Templates,
      minPermittedActions: [Action.Create, Action.Update],
      requiresAuth: true,
      parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
    },
  },
  {
    path: "inference/custom/edit/:id",
    name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_EDIT,
    component: WorkloadTemplateInferenceCustom,
    meta: {
      pageTitle: "Edit inference template",
      fullWidth: true,
      requestToLeave: true,
      backPageName: "workload-template-index",
      resourceType: ResourceType.Templates,
      minPermittedActions: [Action.Create, Action.Update],
      requiresAuth: true,
      parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
    },
  },
];
