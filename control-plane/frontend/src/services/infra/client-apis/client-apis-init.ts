/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsServiceApi } from "./assets-service-api/assets-service-api";
import { authorizationApi } from "./authorization-api/authorization-api";
import { backendApi } from "./backend-api/backend-api";
import { bffServiceApi } from "./bff-service-api/bff-service-api";
import { clusterApi } from "./cluster-api/cluster-api";
import { clusterServiceApi } from "./cluster-service-api/cluster-service-api";
import { identityManagerApi } from "./identity-manager-api/identity-manager-api";
import { metricsServiceApi } from "./metrics-service-api/metrics-service-api";
import { policyServiceApi } from "./policy-service-api/policy-service-api";
import { tenantManagerApi } from "./tenants-manager-api/tenants-manager-api";
import { workloadServiceApi } from "./workloads-service-api/workloads-service-api";
import { k8sObjectTrackerApi } from "./k8s-object-tracker-api/k8s-object-tracker-api";
import { cliApi } from "./cli-exposer-api/cli-exposer-api";
import { orgUnitServiceApi } from "@/services/infra/client-apis/org-unit-service-api/org-unit-service-api";
import { notificationServiceApi } from "@/services/infra/client-apis/notification-service-api/notification-service-api";
import { auditServiceApi } from "@/services/infra/client-apis/audit-service-api/audit-service-api";
import { dataVolumesServiceApi } from "./data-volume-service-api/data-volume-service-api";
import { catalogServiceApi } from "./catalog-service-api/catalog-service-api";
import { workloadsManagerApi } from "./workloads-manager-api/workloads-manager-api";

export function initClientApis(accessToken: string): void {
  assetsServiceApi.setAccessToken = accessToken;
  authorizationApi.setAccessToken = accessToken;
  backendApi.setAccessToken = accessToken;
  bffServiceApi.setAccessToken = accessToken;
  clusterApi.setAccessToken = accessToken;
  clusterServiceApi.setAccessToken = accessToken;
  identityManagerApi.setAccessToken = accessToken;
  metricsServiceApi.setAccessToken = accessToken;
  policyServiceApi.setAccessToken = accessToken;
  tenantManagerApi.setAccessToken = accessToken;
  workloadServiceApi.setAccessToken = accessToken;
  k8sObjectTrackerApi.setAccessToken = accessToken;
  cliApi.setAccessToken = accessToken;
  orgUnitServiceApi.setAccessToken = accessToken;
  notificationServiceApi.setAccessToken = accessToken;
  auditServiceApi.setAccessToken = accessToken;
  dataVolumesServiceApi.setAccessToken = accessToken;
  catalogServiceApi.setAccessToken = accessToken;
  workloadsManagerApi.setAccessToken = accessToken;
}
