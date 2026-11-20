/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config } from '../helpers';
import { BackendApi } from './apis/backendApi';
import { BackendClusterApi } from './apis/backendClusterApi';
import { BackendUndocumentedApi } from './apis/backendUndocumentedApi';
import { ProjectsDepartmentsApi } from './apis/projectsDepartmentsApi';
import { AssetsApi } from './apis/assetsApi';
import { AuthorizationApi } from './apis/authorizationApi';
import { IdentityManagerApi } from './apis/identityManagerApi';
import { TenantApi as TenantApiV1 } from './apis/backendApi/generated/api';
import { ClusterApi } from './apis/clusterApi';
import { k8sObjectsTrackerApi } from './apis/k8sObjectsTrackerApi';
import { PolicyApiHandler } from './apis/policyApi';
import { WorkloadsService, WorkloadsCrudService } from './apis/workloadsApi';
import { TenantsApi, LogoApi, TenantApi } from './apis/tenantsManagerApi/generated/api';
import { CliExposer } from './apis/cliExposer';
import { IdpsApi } from './apis/identityManagerApi/generated/api';
import { DatavolumesService } from './apis/datavolumesApi';
import { OrgUnitService } from './apis/orgUnitApi';
import { AuditApi } from './apis/auditApi';
import { NotificationsProxyApi } from './apis/notificationsProxyApi';
import { NetworkTopologiesApi } from './apis/clusterApi/generated/api/networkTopologiesApi';
import { WorkloadsManagerService } from './apis/workloadsManagerApi';
import { CatalogApi } from './apis/catalogApi';
import { TempBackendSettingsApi } from './apis/tempBackendSettingsApi';
import { BffServiceApi } from './apis/bffServiceApi';

const backendClusterApi = new BackendClusterApi(config.tenantUrl);
const backendApi = new BackendApi(config.tenantUrl);
const backendUndocumentedApi = new BackendUndocumentedApi(config.tenantUrl);
const projectsDepartmentsApi = new ProjectsDepartmentsApi(config.tenantUrl);
const assetsApi = new AssetsApi(config.tenantUrl);
const authorizationApi = new AuthorizationApi(config.tenantUrl);
const identityManagerApi = new IdentityManagerApi(config.tenantUrl);
const tenantApiV1 = new TenantApiV1(config.tenantUrl);
const clusterApi = new ClusterApi(config.tenantUrl);
const k8sObjectTrackerApi = new k8sObjectsTrackerApi(config.tenantUrl);
const cliExposerApi = new CliExposer(config.tenantUrl);
const datavolumesService = new DatavolumesService(config.tenantUrl);
const orgUnitService = new OrgUnitService(config.tenantUrl);
const auditApi = new AuditApi(config.tenantUrl);
const notificationsProxyApi = new NotificationsProxyApi(config.tenantUrl);

const policyApi = new PolicyApiHandler(config.tenantUrl);
const workloadsService = new WorkloadsService(config.tenantUrl);
const workloadsManagerService = new WorkloadsManagerService(config.tenantUrl);

// in workloads phase1 we direct workloads crud APIs to researcher service
const workloadsCrudService = new WorkloadsCrudService(config.tenantUrl);
const tenantsApi = new TenantsApi(config.tenantUrl);
const logoApi = new LogoApi(config.tenantUrl);
const tenantApi = new TenantApi(config.tenantUrl);
const idp = new IdpsApi(config.tenantUrl);
const networkTopologiesApi = new NetworkTopologiesApi(config.tenantUrl);
const catalogApi = new CatalogApi(config.tenantUrl);
const tempBackendSettingsApi = new TempBackendSettingsApi(config.tenantUrl);
const bffServiceApi = new BffServiceApi(config.tenantUrl);
export {
    backendUndocumentedApi,
    backendClusterApi,
    backendApi,
    projectsDepartmentsApi,
    assetsApi,
    authorizationApi,
    identityManagerApi,
    tenantApiV1,
    clusterApi,
    k8sObjectTrackerApi,
    policyApi,
    workloadsService,
    workloadsCrudService,
    workloadsManagerService,
    datavolumesService,
    orgUnitService,
    tenantsApi,
    logoApi,
    tenantApi,
    cliExposerApi,
    idp,
    auditApi,
    notificationsProxyApi,
    networkTopologiesApi,
    catalogApi,
    tempBackendSettingsApi,
    bffServiceApi
};
