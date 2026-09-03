/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { getBootConfig } from "@/services/infra/boot-config.service/boot-config.service";

export const ROOT_URL = "https://run-ai-docs.nvidia.com";
const LATEST_URL = "https://docs.run.ai/latest";
// when you change this link make sure you update the login-success.html as well
const bootConfig = getBootConfig();

export const DOCS = {
  ROOT: bootConfig.docs.baseUrl,

  INSTALLATION: {
    CLUSTER: getUrlByEnvType(`${ROOT_URL}/self-hosted/getting-started/overview#nvidia-run-ai-cluster`),
    PREREQUISITES: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/getting-started/installation/install-using-helm/system-requirements`,
    ),
    NETWORK_REQUIREMENTS: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/getting-started/installation/install-using-helm/network-requirements`,
    ),
    HELM_INSTALL: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/getting-started/installation/install-using-helm/helm-install`,
    ),
    CUSTOM_INSTALL: getUrlByEnvType(`${ROOT_URL}/self-hosted/getting-started/installation/custom-install`),
    UNINSTALL: getUrlByEnvType(`${ROOT_URL}/self-hosted/getting-started/installation/install-using-helm/uninstall`),
    FQDN_REQUIREMENTS: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/getting-started/installation/install-using-helm/system-requirements#fully-qualified-domain-name-fqdn`,
    ),
    DISTRIBUTED_TRAINING: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/getting-started/installation/install-using-helm/system-requirements#distributed-training`,
    ),
    EMAIL_NOTIFICATIONS: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/settings/general-settings/notifications#email-notifications`,
    ),
    USER_INVITATIONS: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/infrastructure-setup/authentication/users#creating-a-local-user`,
    ),
  },

  AUTHENTICATION: {
    OVERVIEW: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/overview`),
    ACCESS_RULES: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/accessrules`),
    SSO: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/sso/`),
    SAML: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/sso/saml`),
    OIDC: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/sso/openidconnect`),
    OPENSHIFT: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/sso/openshift`),
    ROLES: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/roles#roles-in-nvidia-run-ai`),
    CUSTOM_ROLES_API: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/infrastructure-setup/authentication/roles#custom-roles-api-only`,
    ),
    RESEARCHER: {
      OVERVIEW: getUrlByEnvType(`${LATEST_URL}/admin/runai-setup/authentication/researcher-authentication`),
      CLI_ACCESS: getUrlByEnvType(
        `${LATEST_URL}/admin/runai-setup/authentication/researcher-authentication/#command-line-interface-access`,
      ),
      K8S_CONFIG: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/authentication/cluster-authentication`),
    },
  },

  CLI: {
    INSTALL: getUrlByEnvType(`${ROOT_URL}/self-hosted/reference/cli/install-cli`),
    SUBMIT: getUrlByEnvType(`${ROOT_URL}/self-hosted/reference/cli/runai`),
  },

  PLATFORM_MANAGEMENT: {
    PROJECTS: getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/aiinitiatives/organization/projects`),
    DEPARTMENTS: getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/aiinitiatives/organization/departments`),
    POLICIES: getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/policies/policies-and-rules`),
    SCHEDULING_RULES: getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/policies/scheduling-rules`),
    SCHEDULING_CONCEPTS: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/platform-management/runai-scheduler/scheduling/concepts-and-principles`,
    ),
    DYNAMIC_FRACTION: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/platform-management/runai-scheduler/resource-optimization/dynamic-fractions`,
    ),
    ORGANIZATION: {
      PROJECTS: getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/aiinitiatives/organization/projects`),
      DEPARTMENTS: getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/aiinitiatives/organization/departments`),
    },
    NODE_GROUPS: getUrlByEnvType(`${LATEST_URL}/admin/researcher-setup/limit-to-node-group`),
    REPORTS: getUrlByEnvType(`${ROOT_URL}/saas/platform-management/monitor-performance/reports`),
  },

  WORKLOADS: {
    DATA_VOLUMES: getUrlByEnvType(`${ROOT_URL}/self-hosted/workloads-in-nvidia-run-ai/assets/data-volumes`),
    EXTERNAL_ACCESS_TO_CONTAINERS: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/infrastructure-setup/advanced-setup/container-access/external-access-to-containers`,
    ),
    CREATING_A_HUGGING_FACE_INFERENCE_WORKLOAD: getUrlByEnvType(
      `${ROOT_URL}/saas/workloads-in-nvidia-run-ai/using-inference/hugging-face-inference#creating-a-hugging-face-inference-workload`,
    ),
  },

  SLACK_NOTIFICATIONS: {
    CONNECT_WORKSPACE: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/settings/general-settings/notifications#slack-notifications`,
    ),
  },

  TROUBLESHOOTING: {
    CLUSTERS: getUrlByEnvType(`${ROOT_URL}/self-hosted/infrastructure-setup/procedures/clusters`),
    TERMINAL: getUrlByEnvType(`${ROOT_URL}/self-hosted/workloads-in-nvidia-run-ai/workloads#troubleshooting`),
  },

  CONFIGURATION: {
    NON_ROOT_CONTAINERS: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/latest/infrastructure-setup/advanced-setup/container-access/user-identity-in-containers`,
    ),
  },

  TOPOLOGIES: {
    SCHEDULING: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/platform-management/aiinitiatives/resources/topology-aware-scheduling`,
    ),
    GB200: getUrlByEnvType(`${ROOT_URL}/self-hosted/platform-management/aiinitiatives/resources/using-gb200`),
  },

  NODE_POOLS: {
    TIME_BASED_FAIR_SHARE: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/platform-management/runai-scheduler/scheduling/how-the-scheduler-works#time-based-fairshare`,
    ),
    PERSISTENT_DATA: getUrlByEnvType(
      `${ROOT_URL}/saas/infrastructure-setup/advanced-setup/cluster-config#all-configurations-helm-and-runaiconfig`,
    ),
    NUMA_AWARE_SCHEDULING: getUrlByEnvType(
      `${ROOT_URL}/self-hosted/platform-management/aiinitiatives/resources/numa-aware-scheduling`,
    ),
  },

  WHATS_NEW: {
    URL: bootConfig.docs.whatsNew.url,
    LABEL: bootConfig.docs.whatsNew.label,
    DEPRECATION_NOTIFICATIONS: getUrlByEnvType(`${ROOT_URL}/self-hosted/getting-started/whats-new/whats-new-2-24`),
  },

  EXTERNAL: {
    VLLM: {
      SUPPORTED_MODELS: "https://docs.vllm.ai/en/latest/models/supported_models.html",
    },
    HUGGING_FACE: {
      BASE_URL: "https://huggingface.co",
      TGI_SUPPORTED_MODELS: "https://huggingface.co/docs/text-generation-inference/en/supported_models",
    },
    NIM: {
      CATALOG_BASE_URL: "https://catalog.ngc.nvidia.com/orgs/nim/teams",
      NGC_CATALOG: "https://docs.nvidia.com/ngc/gpu-cloud/ngc-catalog-user-guide/index.html#what-is-ngc-catalog",
      DEPLOY_INFERENCE: `${ROOT_URL}/self-hosted/workloads-in-nvidia-run-ai/using-inference/nim-inference`,
    },
  },

  // when you change these links make sure you update the login-success.html as well
  SUPPORT: {
    URL: bootConfig.support.url ?? "",
    LABEL: bootConfig.support.label,
  },

  KUBERNETES: {
    TASKS: {
      EXTENDED_RESOURCES: "https://kubernetes.io/docs/tasks/configure-pod-container/extended-resource/",
    },
    REFERENCE: {
      API: {
        QUANTITY: "https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/",
      },
    },
  },
} as const;

// Rebase legacy Run:ai documentation URLs onto the deployment-provided docs base.
export function getUrlByEnvType(url: string): string {
  const docsBaseUrl = getBootConfig().docs.baseUrl.replace(/\/$/, "");
  const legacyRoot = [`${ROOT_URL}/self-hosted`, `${ROOT_URL}/saas`].find(
    (candidate) => url === candidate || url.startsWith(`${candidate}/`),
  );
  if (!legacyRoot) return url;

  const suffixParts = url.slice(legacyRoot.length).split("/");
  if (isDocsVersionSegment(suffixParts[1])) suffixParts.splice(1, 1);
  return docsBaseUrl + suffixParts.join("/");
}

export function resolveDocsLinksInHtml(html: string): string {
  return html.replace(/(\bhref\s*=\s*)(["'])([^"']+)\2/gi, (match, prefix, quote, href: string) => {
    if (!href.startsWith(ROOT_URL)) return match;
    return `${prefix}${quote}${getUrlByEnvType(href)}${quote}`;
  });
}

function isDocsVersionSegment(value: string | undefined): boolean {
  if (!value) return false;
  const normalized = value.startsWith("v") ? value.slice(1) : value;
  const parts = normalized.split(".");
  return (
    parts.length === 2 && parts.every((part) => part.length > 0 && [...part].every((char) => char >= "0" && char <= "9"))
  );
}

export type DocsType = typeof DOCS;

export default DOCS;
