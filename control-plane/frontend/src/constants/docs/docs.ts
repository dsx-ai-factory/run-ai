/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { getBootConfig } from "@/services/infra/boot-config.service/boot-config.service";

import { DOCS_CATALOG_V1, DOCS_CATALOG_V2, type DocsKey } from "./catalogs";

const bootConfig = getBootConfig();

export function getDocsUrl(path: string): string {
  const docsBaseUrl = getBootConfig().docs.baseUrl.replace(/\/+$/, "");
  const normalizedPath = path.replace(/^\/+/, "");
  return normalizedPath ? `${docsBaseUrl}/${normalizedPath}` : docsBaseUrl;
}

export const DOCS = {
  ROOT: bootConfig.docs.baseUrl,

  INSTALLATION: {
    CLUSTER: getDocsLink("clusterOverview"),
    PREREQUISITES: getDocsLink("systemRequirements"),
    CLUSTER_PREREQUISITES: getDocsLink("clusterPrerequisites"),
    NETWORK_REQUIREMENTS: getDocsLink("networkRequirements"),
    HELM_INSTALL: getDocsLink("clusterInstallation"),
    CUSTOM_INSTALL: getDocsLink("customInstallation"),
    UNINSTALL: getDocsLink("uninstallation"),
    FQDN_REQUIREMENTS: getDocsLink("fqdnRequirements"),
    DISTRIBUTED_TRAINING: getDocsLink("distributedTraining"),
    EMAIL_NOTIFICATIONS: getDocsLink("emailNotifications"),
    USER_INVITATIONS: getDocsLink("localUserCreation"),
  },

  AUTHENTICATION: {
    OVERVIEW: getDocsLink("authenticationOverview"),
    ACCESS_RULES: getDocsLink("accessRules"),
    SSO: getDocsLink("sso"),
    SAML: getDocsLink("saml"),
    OIDC: getDocsLink("oidc"),
    OPENSHIFT: getDocsLink("openshift"),
    ROLES: getDocsLink("roles"),
    CUSTOM_ROLES_API: getDocsLink("customRoles"),
    RESEARCHER: {
      OVERVIEW: getDocsLink("authenticationOverview"),
      CLI_ACCESS: getDocsLink("cliInstallation"),
      K8S_CONFIG: getDocsLink("clusterAuthentication"),
    },
  },

  CLI: {
    INSTALL: getDocsLink("cliInstallation"),
    SUBMIT: getDocsLink("cliReference"),
  },

  PLATFORM_MANAGEMENT: {
    PROJECTS: getDocsLink("projects"),
    DEPARTMENTS: getDocsLink("departments"),
    POLICIES: getDocsLink("policies"),
    SCHEDULING_RULES: getDocsLink("schedulingRules"),
    SCHEDULING_CONCEPTS: getDocsLink("schedulingConcepts"),
    DYNAMIC_FRACTION: getDocsLink("dynamicFractions"),
    ORGANIZATION: {
      PROJECTS: getDocsLink("projects"),
      DEPARTMENTS: getDocsLink("departments"),
    },
    NODE_GROUPS: getDocsLink("nodePools"),
    REPORTS: getDocsLink("reports"),
  },

  WORKLOADS: {
    DATA_VOLUMES: getDocsLink("dataVolumes"),
    EXTERNAL_ACCESS_TO_CONTAINERS: getDocsLink("externalContainerAccess"),
    CREATING_A_HUGGING_FACE_INFERENCE_WORKLOAD: getDocsLink("huggingFaceWorkloadCreation"),
  },

  SLACK_NOTIFICATIONS: {
    CONNECT_WORKSPACE: getDocsLink("slackNotifications"),
  },

  TROUBLESHOOTING: {
    CLUSTERS: getDocsLink("clusters"),
    TERMINAL: getDocsLink("workloadTroubleshooting"),
  },

  CONFIGURATION: {
    NON_ROOT_CONTAINERS: getDocsLink("containerIdentity"),
  },

  TOPOLOGIES: {
    SCHEDULING: getDocsLink("topologyScheduling"),
    GB200: getDocsLink("gb200"),
  },

  NODE_POOLS: {
    TIME_BASED_FAIR_SHARE: getDocsLink("timeBasedFairShare"),
    PERSISTENT_DATA: getDocsLink("clusterConfiguration"),
    NUMA_AWARE_SCHEDULING: getDocsLink("numaScheduling"),
  },

  WHATS_NEW: {
    URL: bootConfig.docs.whatsNew.url,
    LABEL: bootConfig.docs.whatsNew.label,
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
      DEPLOY_INFERENCE: getDocsLink("nimInference"),
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

export function resolveDocsPath(key: DocsKey): string | null {
  const catalog = getBootConfig().docs.catalogVersion === "v2" ? DOCS_CATALOG_V2 : DOCS_CATALOG_V1;
  // eslint-disable-next-line security/detect-object-injection -- Keys are restricted to the DocsKey union.
  return catalog[key];
}

type ConfirmedDocsKey = {
  [Key in DocsKey]: (typeof DOCS_CATALOG_V1 | typeof DOCS_CATALOG_V2)[Key] extends string ? Key : never;
}[DocsKey];

export function getDocsLink(key: ConfirmedDocsKey): string;
export function getDocsLink(key: DocsKey): string | undefined;
export function getDocsLink(key: DocsKey): string | undefined {
  const path = resolveDocsPath(key);
  return path === null ? undefined : getDocsUrl(path);
}

export function resolveDocsLinksInHtml(html: string): string {
  const template = document.createElement("template");
  template.innerHTML = html;
  template.content.querySelectorAll<HTMLAnchorElement>("a[data-doc-key]").forEach((anchor) => {
    const key = anchor.getAttribute("data-doc-key") ?? "";
    const href = Object.prototype.hasOwnProperty.call(DOCS_CATALOG_V1, key) ? getDocsLink(key as DocsKey) : undefined;
    if (href) {
      anchor.setAttribute("href", href);
    } else {
      anchor.removeAttribute("href");
    }
    anchor.removeAttribute("data-doc-key");
  });
  return template.innerHTML;
}

export type DocsType = typeof DOCS;

export default DOCS;
