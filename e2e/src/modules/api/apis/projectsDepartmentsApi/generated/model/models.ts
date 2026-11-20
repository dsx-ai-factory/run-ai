import localVarRequest from 'request';

export * from './aggregatedResources';
export * from './aggregatedResourcesMemory';
export * from './aggregatedResourcesMemoryAllOf';
export * from './aggregatedResourcesNodePool';
export * from './ancestors';
export * from './areReportsAvailable200Response';
export * from './assignedResourcesV1';
export * from './assignedResourcesV1Cpu';
export * from './assignedResourcesV1Gpu';
export * from './assignedResourcesV1Memory';
export * from './assignedResourcesV1Response';
export * from './assignedResourcesV1ResponseCpu';
export * from './assignedResourcesV1ResponseGpu';
export * from './assignedResourcesV1ResponseMemory';
export * from './batchDeleteReportsRequest';
export * from './batchDeleteReportsResponse';
export * from './batchPatchDepartmentResourcesResponse';
export * from './batchPatchNodePoolResourcesRequest';
export * from './batchPatchProjectResourcesResponse';
export * from './batchResponseFailureItem';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './countResponse';
export * from './dataDepartmentFields';
export * from './dataDepartmentFieldsAllOf';
export * from './dataUpdatableDepartmentFields';
export * from './dataUpdatableFields';
export * from './deletedOrgUnit';
export * from './deletion';
export * from './deletionTypes';
export * from './department';
export * from './departmentAllOf';
export * from './departmentAllOfStatus';
export * from './departmentCreationRequest';
export * from './departmentCreationRequestAllOf';
export * from './departmentFilterSortFields';
export * from './departmentForGetV1';
export * from './departmentForGetV1AllOf';
export * from './departmentForList';
export * from './departmentForListAllOf';
export * from './departmentForListAllOfStatus';
export * from './departmentGetRanking';
export * from './departmentMetricsV1';
export * from './departmentMetricsV1Current';
export * from './departmentMetricsV1CurrentProjectResourcesInner';
export * from './departmentMetricsV1CurrentProjectResourcesInnerAllOf';
export * from './departmentMetricsV1CurrentResourcesInner';
export * from './departmentMetricsV1CurrentResourcesInnerAllOf';
export * from './departmentMetricsV1Metadata';
export * from './departmentMetricsV1Response';
export * from './departmentNodePoolResources';
export * from './departmentNodePoolResourcesFilterSortFields';
export * from './departmentProjectChild';
export * from './departmentProjectParent';
export * from './departmentStatus';
export * from './departmentStatusUpdateRequest';
export * from './departmentSumResponse';
export * from './departmentSync';
export * from './departmentUpdateRequest';
export * from './departmentUpdateRequestAllOf';
export * from './departmentV1AccessControl';
export * from './departmentV1CommonFields';
export * from './departmentV1CommonFieldsResources';
export * from './departmentV1CommonFieldsResponse';
export * from './departmentV1CommonFieldsResponseResources';
export * from './departmentV1CreationRequest';
export * from './departmentsMetricsV1Response';
export * from './departmentsMetricsV1ResponseDataInner';
export * from './departmentsMetricsV1ResponseDataInnerAllOf';
export * from './departmentsResourcesResetRequest';
export * from './descendants';
export * from './deservedResource';
export * from './failedNodePoolResourcesPatchRequest';
export * from './getDepartmentNodePoolResources200Response';
export * from './getDepartments200Response';
export * from './getDepartmentsRanking200Response';
export * from './getDepartmentsV1200ResponseInner';
export * from './getInternalDeletedOrgUnits200Response';
export * from './getInternalOrgTree200Response';
export * from './getNodeTypes200Response';
export * from './getNodeTypes200ResponseNodeTypesInner';
export * from './getProjectNodePoolResources200Response';
export * from './getProjects200Response';
export * from './getProjectsMetadata200Response';
export * from './getProjectsMetadata200ResponseProjectsInner';
export * from './getProjectsRanking200Response';
export * from './getRanks200Response';
export * from './getSyncDepartments200Response';
export * from './getSyncProjects200Response';
export * from './httpResponse';
export * from './listReports200Response';
export * from './measurementResponse';
export * from './measurementResponseValuesInner';
export * from './metricsResourceDataV1';
export * from './metricsResourcesDataV1';
export * from './metricsResponse';
export * from './modelError';
export * from './nodeAffinityV1';
export * from './nodeAffinityV1SelectedTypesInner';
export * from './nodePool';
export * from './nodePoolAssignedResourcesV1';
export * from './nodePoolAssignedResourcesV1Cpu';
export * from './nodePoolAssignedResourcesV1Gpu';
export * from './nodePoolAssignedResourcesV1Memory';
export * from './nodePoolAssignedResourcesV1Response';
export * from './nodePoolAssignedResourcesV1ResponseCpu';
export * from './nodePoolAssignedResourcesV1ResponseGpu';
export * from './nodePoolAssignedResourcesV1ResponseMemory';
export * from './nodePoolAssignedResourcesV1ResponseNodePool';
export * from './nodePoolResourcesPatchRequest';
export * from './nodeType';
export * from './nodeTypeCreateRequest';
export * from './nodeTypeUpdateRequest';
export * from './nodeTypeV1';
export * from './nodeTypesPerWorkload';
export * from './nodepoolResourcesResetAction';
export * from './nonNullResource';
export * from './nonNullResourceV1';
export * from './orgUnit';
export * from './orgUnitMetricType';
export * from './orgUnitSetting';
export * from './orgUnitSettingsGetResponse';
export * from './orgUnitSettingsKeyEnum';
export * from './orgUnitSingleSetting';
export * from './orgUnitTelemetryType';
export * from './orgUnitTelemetryTypeInternal';
export * from './orgUnitsNodepoolResourcesResetRequest';
export * from './orgUnitsNodepoolSummaryInner';
export * from './orgUnitsNodepoolSummaryType';
export * from './orgsNodePoolSummary200Response';
export * from './overtimeData';
export * from './overtimeDataRange24hData';
export * from './overtimeRangeData';
export * from './project';
export * from './projectAggregatedResourcesV1';
export * from './projectAggregatedResourcesV1Cpu';
export * from './projectAggregatedResourcesV1Gpu';
export * from './projectAggregatedResourcesV1Memory';
export * from './projectAggregatedResourcesV1NodePool';
export * from './projectAllOf';
export * from './projectCreationRequest';
export * from './projectCreationRequestAllOf';
export * from './projectEffective';
export * from './projectFilterSortFields';
export * from './projectGetRanking';
export * from './projectInSyncRequest';
export * from './projectMetadata';
export * from './projectMetricsV1';
export * from './projectMetricsV1Current';
export * from './projectMetricsV1Metadata';
export * from './projectMetricsV1Response';
export * from './projectNodePoolResources';
export * from './projectNodePoolResourcesFilterSortFields';
export * from './projectPhase';
export * from './projectQuotasV1Inner';
export * from './projectStatus';
export * from './projectStatusUpdateRequest';
export * from './projectSync';
export * from './projectSyncPermissions';
export * from './projectUpdateRequest';
export * from './projectUpdateRequestAllOf';
export * from './projectV1';
export * from './projectV1AllOf';
export * from './projectV1AllOfStatus';
export * from './projectV1CommonFields';
export * from './projectV1CommonFieldsPermissions';
export * from './projectV1CommonFieldsResources';
export * from './projectV1CreationRequest';
export * from './projectV1NodeAffinity';
export * from './projectV1NodeAffinityInteractive';
export * from './projectV1NodeAffinityResponse';
export * from './projectV1NodeAffinityResponseInteractive';
export * from './projectV1NodeAffinityResponseTrain';
export * from './projectV1NodeAffinityResponseTrainSelectedTypesInner';
export * from './projectV1NodeAffinityTrain';
export * from './projectV1ResponseCommonFields';
export * from './projectV1ResponseCommonFieldsPermissions';
export * from './projectV1ResponseCommonFieldsResources';
export * from './projectV1StatusRequest';
export * from './projectsMetricsV1Response';
export * from './projectsMetricsV1ResponseDataInner';
export * from './projectsResourcesResetRequest';
export * from './projectsResourcesResetRequestAllOf';
export * from './quotaStatus';
export * from './quotaStatusNodePool';
export * from './quotaStatusNodePoolAllOf';
export * from './quotaStatusResource';
export * from './quotaStatusResourceListV1';
export * from './quotaStatusV1Inner';
export * from './range';
export * from './rank';
export * from './rankingMetricType';
export * from './rankingOrderType';
export * from './report';
export * from './reportCalculatedFields';
export * from './reportFilterAndSortFields';
export * from './reportGroupByOptions';
export * from './reportRequestFields';
export * from './reportStatus';
export * from './resource';
export * from './resourceV1';
export * from './resourceV1Response';
export * from './resources';
export * from './resourcesCpu';
export * from './resourcesDataOvertimeV1';
export * from './resourcesDataOvertimeV1AllOf';
export * from './resourcesGpu';
export * from './resourcesMemory';
export * from './resourcesMemoryAllOf';
export * from './resourcesNodePool';
export * from './resourcesNullable';
export * from './resourcesNullableGpu';
export * from './resourcesPerNodePool';
export * from './resourcesPerNodePoolGpu';
export * from './resourcesPerNodePoolNullable';
export * from './resourcesPerNodePoolNullableGpu';
export * from './resourcesResetAction';
export * from './resourcesSync';
export * from './resourcesSyncGpu';
export * from './resourcesSyncMemory';
export * from './schedulingRules';
export * from './scopeType';
export * from './submissionError';
export * from './sumResponse';
export * from './syncPatchRequestInner';
export * from './syncProjectsRequest';
export * from './telemetryResponse';
export * from './telemetryResponseValuesInner';
export * from './telemetryResponseValuesInnerGroupsInner';
export * from './timeRangeV1';
export * from './totalResources';
export * from './units';
export * from './updateOrgUnitSettingByKeyRequest';
export * from './verbosity';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AggregatedResources } from './aggregatedResources';
import { AggregatedResourcesMemory } from './aggregatedResourcesMemory';
import { AggregatedResourcesMemoryAllOf } from './aggregatedResourcesMemoryAllOf';
import { AggregatedResourcesNodePool } from './aggregatedResourcesNodePool';
import { Ancestors } from './ancestors';
import { AreReportsAvailable200Response } from './areReportsAvailable200Response';
import { AssignedResourcesV1 } from './assignedResourcesV1';
import { AssignedResourcesV1Cpu } from './assignedResourcesV1Cpu';
import { AssignedResourcesV1Gpu } from './assignedResourcesV1Gpu';
import { AssignedResourcesV1Memory } from './assignedResourcesV1Memory';
import { AssignedResourcesV1Response } from './assignedResourcesV1Response';
import { AssignedResourcesV1ResponseCpu } from './assignedResourcesV1ResponseCpu';
import { AssignedResourcesV1ResponseGpu } from './assignedResourcesV1ResponseGpu';
import { AssignedResourcesV1ResponseMemory } from './assignedResourcesV1ResponseMemory';
import { BatchDeleteReportsRequest } from './batchDeleteReportsRequest';
import { BatchDeleteReportsResponse } from './batchDeleteReportsResponse';
import { BatchPatchDepartmentResourcesResponse } from './batchPatchDepartmentResourcesResponse';
import { BatchPatchNodePoolResourcesRequest } from './batchPatchNodePoolResourcesRequest';
import { BatchPatchProjectResourcesResponse } from './batchPatchProjectResourcesResponse';
import { BatchResponseFailureItem } from './batchResponseFailureItem';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { CountResponse } from './countResponse';
import { DataDepartmentFields } from './dataDepartmentFields';
import { DataDepartmentFieldsAllOf } from './dataDepartmentFieldsAllOf';
import { DataUpdatableDepartmentFields } from './dataUpdatableDepartmentFields';
import { DataUpdatableFields } from './dataUpdatableFields';
import { DeletedOrgUnit } from './deletedOrgUnit';
import { Deletion } from './deletion';
import { DeletionTypes } from './deletionTypes';
import { Department } from './department';
import { DepartmentAllOf } from './departmentAllOf';
import { DepartmentAllOfStatus } from './departmentAllOfStatus';
import { DepartmentCreationRequest } from './departmentCreationRequest';
import { DepartmentCreationRequestAllOf } from './departmentCreationRequestAllOf';
import { DepartmentFilterSortFields } from './departmentFilterSortFields';
import { DepartmentForGetV1 } from './departmentForGetV1';
import { DepartmentForGetV1AllOf } from './departmentForGetV1AllOf';
import { DepartmentForList } from './departmentForList';
import { DepartmentForListAllOf } from './departmentForListAllOf';
import { DepartmentForListAllOfStatus } from './departmentForListAllOfStatus';
import { DepartmentGetRanking } from './departmentGetRanking';
import { DepartmentMetricsV1 } from './departmentMetricsV1';
import { DepartmentMetricsV1Current } from './departmentMetricsV1Current';
import { DepartmentMetricsV1CurrentProjectResourcesInner } from './departmentMetricsV1CurrentProjectResourcesInner';
import { DepartmentMetricsV1CurrentProjectResourcesInnerAllOf } from './departmentMetricsV1CurrentProjectResourcesInnerAllOf';
import { DepartmentMetricsV1CurrentResourcesInner } from './departmentMetricsV1CurrentResourcesInner';
import { DepartmentMetricsV1CurrentResourcesInnerAllOf } from './departmentMetricsV1CurrentResourcesInnerAllOf';
import { DepartmentMetricsV1Metadata } from './departmentMetricsV1Metadata';
import { DepartmentMetricsV1Response } from './departmentMetricsV1Response';
import { DepartmentNodePoolResources } from './departmentNodePoolResources';
import { DepartmentNodePoolResourcesFilterSortFields } from './departmentNodePoolResourcesFilterSortFields';
import { DepartmentProjectChild } from './departmentProjectChild';
import { DepartmentProjectParent } from './departmentProjectParent';
import { DepartmentStatus } from './departmentStatus';
import { DepartmentStatusUpdateRequest } from './departmentStatusUpdateRequest';
import { DepartmentSumResponse } from './departmentSumResponse';
import { DepartmentSync } from './departmentSync';
import { DepartmentUpdateRequest } from './departmentUpdateRequest';
import { DepartmentUpdateRequestAllOf } from './departmentUpdateRequestAllOf';
import { DepartmentV1AccessControl } from './departmentV1AccessControl';
import { DepartmentV1CommonFields } from './departmentV1CommonFields';
import { DepartmentV1CommonFieldsResources } from './departmentV1CommonFieldsResources';
import { DepartmentV1CommonFieldsResponse } from './departmentV1CommonFieldsResponse';
import { DepartmentV1CommonFieldsResponseResources } from './departmentV1CommonFieldsResponseResources';
import { DepartmentV1CreationRequest } from './departmentV1CreationRequest';
import { DepartmentsMetricsV1Response } from './departmentsMetricsV1Response';
import { DepartmentsMetricsV1ResponseDataInner } from './departmentsMetricsV1ResponseDataInner';
import { DepartmentsMetricsV1ResponseDataInnerAllOf } from './departmentsMetricsV1ResponseDataInnerAllOf';
import { DepartmentsResourcesResetRequest } from './departmentsResourcesResetRequest';
import { Descendants } from './descendants';
import { DeservedResource } from './deservedResource';
import { FailedNodePoolResourcesPatchRequest } from './failedNodePoolResourcesPatchRequest';
import { GetDepartmentNodePoolResources200Response } from './getDepartmentNodePoolResources200Response';
import { GetDepartments200Response } from './getDepartments200Response';
import { GetDepartmentsRanking200Response } from './getDepartmentsRanking200Response';
import { GetDepartmentsV1200ResponseInner } from './getDepartmentsV1200ResponseInner';
import { GetInternalDeletedOrgUnits200Response } from './getInternalDeletedOrgUnits200Response';
import { GetInternalOrgTree200Response } from './getInternalOrgTree200Response';
import { GetNodeTypes200Response } from './getNodeTypes200Response';
import { GetNodeTypes200ResponseNodeTypesInner } from './getNodeTypes200ResponseNodeTypesInner';
import { GetProjectNodePoolResources200Response } from './getProjectNodePoolResources200Response';
import { GetProjects200Response } from './getProjects200Response';
import { GetProjectsMetadata200Response } from './getProjectsMetadata200Response';
import { GetProjectsMetadata200ResponseProjectsInner } from './getProjectsMetadata200ResponseProjectsInner';
import { GetProjectsRanking200Response } from './getProjectsRanking200Response';
import { GetRanks200Response } from './getRanks200Response';
import { GetSyncDepartments200Response } from './getSyncDepartments200Response';
import { GetSyncProjects200Response } from './getSyncProjects200Response';
import { HttpResponse } from './httpResponse';
import { ListReports200Response } from './listReports200Response';
import { MeasurementResponse } from './measurementResponse';
import { MeasurementResponseValuesInner } from './measurementResponseValuesInner';
import { MetricsResourceDataV1 } from './metricsResourceDataV1';
import { MetricsResourcesDataV1 } from './metricsResourcesDataV1';
import { MetricsResponse } from './metricsResponse';
import { ModelError } from './modelError';
import { NodeAffinityV1 } from './nodeAffinityV1';
import { NodeAffinityV1SelectedTypesInner } from './nodeAffinityV1SelectedTypesInner';
import { NodePool } from './nodePool';
import { NodePoolAssignedResourcesV1 } from './nodePoolAssignedResourcesV1';
import { NodePoolAssignedResourcesV1Cpu } from './nodePoolAssignedResourcesV1Cpu';
import { NodePoolAssignedResourcesV1Gpu } from './nodePoolAssignedResourcesV1Gpu';
import { NodePoolAssignedResourcesV1Memory } from './nodePoolAssignedResourcesV1Memory';
import { NodePoolAssignedResourcesV1Response } from './nodePoolAssignedResourcesV1Response';
import { NodePoolAssignedResourcesV1ResponseCpu } from './nodePoolAssignedResourcesV1ResponseCpu';
import { NodePoolAssignedResourcesV1ResponseGpu } from './nodePoolAssignedResourcesV1ResponseGpu';
import { NodePoolAssignedResourcesV1ResponseMemory } from './nodePoolAssignedResourcesV1ResponseMemory';
import { NodePoolAssignedResourcesV1ResponseNodePool } from './nodePoolAssignedResourcesV1ResponseNodePool';
import { NodePoolResourcesPatchRequest } from './nodePoolResourcesPatchRequest';
import { NodeType } from './nodeType';
import { NodeTypeCreateRequest } from './nodeTypeCreateRequest';
import { NodeTypeUpdateRequest } from './nodeTypeUpdateRequest';
import { NodeTypeV1 } from './nodeTypeV1';
import { NodeTypesPerWorkload } from './nodeTypesPerWorkload';
import { NodepoolResourcesResetAction } from './nodepoolResourcesResetAction';
import { NonNullResource } from './nonNullResource';
import { NonNullResourceV1 } from './nonNullResourceV1';
import { OrgUnit } from './orgUnit';
import { OrgUnitMetricType } from './orgUnitMetricType';
import { OrgUnitSetting } from './orgUnitSetting';
import { OrgUnitSettingsGetResponse } from './orgUnitSettingsGetResponse';
import { OrgUnitSettingsKeyEnum } from './orgUnitSettingsKeyEnum';
import { OrgUnitSingleSetting } from './orgUnitSingleSetting';
import { OrgUnitTelemetryType } from './orgUnitTelemetryType';
import { OrgUnitTelemetryTypeInternal } from './orgUnitTelemetryTypeInternal';
import { OrgUnitsNodepoolResourcesResetRequest } from './orgUnitsNodepoolResourcesResetRequest';
import { OrgUnitsNodepoolSummaryInner } from './orgUnitsNodepoolSummaryInner';
import { OrgUnitsNodepoolSummaryType } from './orgUnitsNodepoolSummaryType';
import { OrgsNodePoolSummary200Response } from './orgsNodePoolSummary200Response';
import { OvertimeData } from './overtimeData';
import { OvertimeDataRange24hData } from './overtimeDataRange24hData';
import { OvertimeRangeData } from './overtimeRangeData';
import { Project } from './project';
import { ProjectAggregatedResourcesV1 } from './projectAggregatedResourcesV1';
import { ProjectAggregatedResourcesV1Cpu } from './projectAggregatedResourcesV1Cpu';
import { ProjectAggregatedResourcesV1Gpu } from './projectAggregatedResourcesV1Gpu';
import { ProjectAggregatedResourcesV1Memory } from './projectAggregatedResourcesV1Memory';
import { ProjectAggregatedResourcesV1NodePool } from './projectAggregatedResourcesV1NodePool';
import { ProjectAllOf } from './projectAllOf';
import { ProjectCreationRequest } from './projectCreationRequest';
import { ProjectCreationRequestAllOf } from './projectCreationRequestAllOf';
import { ProjectEffective } from './projectEffective';
import { ProjectFilterSortFields } from './projectFilterSortFields';
import { ProjectGetRanking } from './projectGetRanking';
import { ProjectInSyncRequest } from './projectInSyncRequest';
import { ProjectMetadata } from './projectMetadata';
import { ProjectMetricsV1 } from './projectMetricsV1';
import { ProjectMetricsV1Current } from './projectMetricsV1Current';
import { ProjectMetricsV1Metadata } from './projectMetricsV1Metadata';
import { ProjectMetricsV1Response } from './projectMetricsV1Response';
import { ProjectNodePoolResources } from './projectNodePoolResources';
import { ProjectNodePoolResourcesFilterSortFields } from './projectNodePoolResourcesFilterSortFields';
import { ProjectPhase } from './projectPhase';
import { ProjectQuotasV1Inner } from './projectQuotasV1Inner';
import { ProjectStatus } from './projectStatus';
import { ProjectStatusUpdateRequest } from './projectStatusUpdateRequest';
import { ProjectSync } from './projectSync';
import { ProjectSyncPermissions } from './projectSyncPermissions';
import { ProjectUpdateRequest } from './projectUpdateRequest';
import { ProjectUpdateRequestAllOf } from './projectUpdateRequestAllOf';
import { ProjectV1 } from './projectV1';
import { ProjectV1AllOf } from './projectV1AllOf';
import { ProjectV1AllOfStatus } from './projectV1AllOfStatus';
import { ProjectV1CommonFields } from './projectV1CommonFields';
import { ProjectV1CommonFieldsPermissions } from './projectV1CommonFieldsPermissions';
import { ProjectV1CommonFieldsResources } from './projectV1CommonFieldsResources';
import { ProjectV1CreationRequest } from './projectV1CreationRequest';
import { ProjectV1NodeAffinity } from './projectV1NodeAffinity';
import { ProjectV1NodeAffinityInteractive } from './projectV1NodeAffinityInteractive';
import { ProjectV1NodeAffinityResponse } from './projectV1NodeAffinityResponse';
import { ProjectV1NodeAffinityResponseInteractive } from './projectV1NodeAffinityResponseInteractive';
import { ProjectV1NodeAffinityResponseTrain } from './projectV1NodeAffinityResponseTrain';
import { ProjectV1NodeAffinityResponseTrainSelectedTypesInner } from './projectV1NodeAffinityResponseTrainSelectedTypesInner';
import { ProjectV1NodeAffinityTrain } from './projectV1NodeAffinityTrain';
import { ProjectV1ResponseCommonFields } from './projectV1ResponseCommonFields';
import { ProjectV1ResponseCommonFieldsPermissions } from './projectV1ResponseCommonFieldsPermissions';
import { ProjectV1ResponseCommonFieldsResources } from './projectV1ResponseCommonFieldsResources';
import { ProjectV1StatusRequest } from './projectV1StatusRequest';
import { ProjectsMetricsV1Response } from './projectsMetricsV1Response';
import { ProjectsMetricsV1ResponseDataInner } from './projectsMetricsV1ResponseDataInner';
import { ProjectsResourcesResetRequest } from './projectsResourcesResetRequest';
import { ProjectsResourcesResetRequestAllOf } from './projectsResourcesResetRequestAllOf';
import { QuotaStatus } from './quotaStatus';
import { QuotaStatusNodePool } from './quotaStatusNodePool';
import { QuotaStatusNodePoolAllOf } from './quotaStatusNodePoolAllOf';
import { QuotaStatusResource } from './quotaStatusResource';
import { QuotaStatusResourceListV1 } from './quotaStatusResourceListV1';
import { QuotaStatusV1Inner } from './quotaStatusV1Inner';
import { Range } from './range';
import { Rank } from './rank';
import { RankingMetricType } from './rankingMetricType';
import { RankingOrderType } from './rankingOrderType';
import { Report } from './report';
import { ReportCalculatedFields } from './reportCalculatedFields';
import { ReportFilterAndSortFields } from './reportFilterAndSortFields';
import { ReportGroupByOptions } from './reportGroupByOptions';
import { ReportRequestFields } from './reportRequestFields';
import { ReportStatus } from './reportStatus';
import { Resource } from './resource';
import { ResourceV1 } from './resourceV1';
import { ResourceV1Response } from './resourceV1Response';
import { Resources } from './resources';
import { ResourcesCpu } from './resourcesCpu';
import { ResourcesDataOvertimeV1 } from './resourcesDataOvertimeV1';
import { ResourcesDataOvertimeV1AllOf } from './resourcesDataOvertimeV1AllOf';
import { ResourcesGpu } from './resourcesGpu';
import { ResourcesMemory } from './resourcesMemory';
import { ResourcesMemoryAllOf } from './resourcesMemoryAllOf';
import { ResourcesNodePool } from './resourcesNodePool';
import { ResourcesNullable } from './resourcesNullable';
import { ResourcesNullableGpu } from './resourcesNullableGpu';
import { ResourcesPerNodePool } from './resourcesPerNodePool';
import { ResourcesPerNodePoolGpu } from './resourcesPerNodePoolGpu';
import { ResourcesPerNodePoolNullable } from './resourcesPerNodePoolNullable';
import { ResourcesPerNodePoolNullableGpu } from './resourcesPerNodePoolNullableGpu';
import { ResourcesResetAction } from './resourcesResetAction';
import { ResourcesSync } from './resourcesSync';
import { ResourcesSyncGpu } from './resourcesSyncGpu';
import { ResourcesSyncMemory } from './resourcesSyncMemory';
import { SchedulingRules } from './schedulingRules';
import { ScopeType } from './scopeType';
import { SubmissionError } from './submissionError';
import { SumResponse } from './sumResponse';
import { SyncPatchRequestInner } from './syncPatchRequestInner';
import { SyncProjectsRequest } from './syncProjectsRequest';
import { TelemetryResponse } from './telemetryResponse';
import { TelemetryResponseValuesInner } from './telemetryResponseValuesInner';
import { TelemetryResponseValuesInnerGroupsInner } from './telemetryResponseValuesInnerGroupsInner';
import { TimeRangeV1 } from './timeRangeV1';
import { TotalResources } from './totalResources';
import { Units } from './units';
import { UpdateOrgUnitSettingByKeyRequest } from './updateOrgUnitSettingByKeyRequest';
import { Verbosity } from './verbosity';

/* tslint:disable:no-unused-variable */
let primitives = [
                    "string",
                    "boolean",
                    "double",
                    "integer",
                    "long",
                    "float",
                    "number",
                    "any"
                 ];

let enumsMap: {[index: string]: any} = {
        "DeletedOrgUnit.TypeEnum": DeletedOrgUnit.TypeEnum,
        "DeletionTypes": DeletionTypes,
        "DepartmentFilterSortFields": DepartmentFilterSortFields,
        "DepartmentNodePoolResourcesFilterSortFields": DepartmentNodePoolResourcesFilterSortFields,
        "NodeAffinityV1.AffinityTypeEnum": NodeAffinityV1.AffinityTypeEnum,
        "NodePool.OperationEnum": NodePool.OperationEnum,
        "NodepoolResourcesResetAction": NodepoolResourcesResetAction,
        "OrgUnitMetricType": OrgUnitMetricType,
        "OrgUnitSettingsKeyEnum": OrgUnitSettingsKeyEnum,
        "OrgUnitTelemetryType": OrgUnitTelemetryType,
        "OrgUnitTelemetryTypeInternal": OrgUnitTelemetryTypeInternal,
        "OrgUnitsNodepoolSummaryInner.TypeEnum": OrgUnitsNodepoolSummaryInner.TypeEnum,
        "OrgUnitsNodepoolSummaryType": OrgUnitsNodepoolSummaryType,
        "ProjectFilterSortFields": ProjectFilterSortFields,
        "ProjectNodePoolResourcesFilterSortFields": ProjectNodePoolResourcesFilterSortFields,
        "ProjectPhase": ProjectPhase,
        "ProjectV1NodeAffinityInteractive.AffinityTypeEnum": ProjectV1NodeAffinityInteractive.AffinityTypeEnum,
        "ProjectV1NodeAffinityResponseInteractive.AffinityTypeEnum": ProjectV1NodeAffinityResponseInteractive.AffinityTypeEnum,
        "ProjectV1NodeAffinityResponseTrain.AffinityTypeEnum": ProjectV1NodeAffinityResponseTrain.AffinityTypeEnum,
        "ProjectV1NodeAffinityTrain.AffinityTypeEnum": ProjectV1NodeAffinityTrain.AffinityTypeEnum,
        "RankingMetricType": RankingMetricType,
        "RankingOrderType": RankingOrderType,
        "ReportFilterAndSortFields": ReportFilterAndSortFields,
        "ReportGroupByOptions": ReportGroupByOptions,
        "ReportStatus": ReportStatus,
        "ResourcesResetAction": ResourcesResetAction,
        "ScopeType": ScopeType,
        "Units": Units,
        "Verbosity": Verbosity,
}

let typeMap: {[index: string]: any} = {
    "AggregatedResources": AggregatedResources,
    "AggregatedResourcesMemory": AggregatedResourcesMemory,
    "AggregatedResourcesMemoryAllOf": AggregatedResourcesMemoryAllOf,
    "AggregatedResourcesNodePool": AggregatedResourcesNodePool,
    "Ancestors": Ancestors,
    "AreReportsAvailable200Response": AreReportsAvailable200Response,
    "AssignedResourcesV1": AssignedResourcesV1,
    "AssignedResourcesV1Cpu": AssignedResourcesV1Cpu,
    "AssignedResourcesV1Gpu": AssignedResourcesV1Gpu,
    "AssignedResourcesV1Memory": AssignedResourcesV1Memory,
    "AssignedResourcesV1Response": AssignedResourcesV1Response,
    "AssignedResourcesV1ResponseCpu": AssignedResourcesV1ResponseCpu,
    "AssignedResourcesV1ResponseGpu": AssignedResourcesV1ResponseGpu,
    "AssignedResourcesV1ResponseMemory": AssignedResourcesV1ResponseMemory,
    "BatchDeleteReportsRequest": BatchDeleteReportsRequest,
    "BatchDeleteReportsResponse": BatchDeleteReportsResponse,
    "BatchPatchDepartmentResourcesResponse": BatchPatchDepartmentResourcesResponse,
    "BatchPatchNodePoolResourcesRequest": BatchPatchNodePoolResourcesRequest,
    "BatchPatchProjectResourcesResponse": BatchPatchProjectResourcesResponse,
    "BatchResponseFailureItem": BatchResponseFailureItem,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "CountResponse": CountResponse,
    "DataDepartmentFields": DataDepartmentFields,
    "DataDepartmentFieldsAllOf": DataDepartmentFieldsAllOf,
    "DataUpdatableDepartmentFields": DataUpdatableDepartmentFields,
    "DataUpdatableFields": DataUpdatableFields,
    "DeletedOrgUnit": DeletedOrgUnit,
    "Deletion": Deletion,
    "Department": Department,
    "DepartmentAllOf": DepartmentAllOf,
    "DepartmentAllOfStatus": DepartmentAllOfStatus,
    "DepartmentCreationRequest": DepartmentCreationRequest,
    "DepartmentCreationRequestAllOf": DepartmentCreationRequestAllOf,
    "DepartmentForGetV1": DepartmentForGetV1,
    "DepartmentForGetV1AllOf": DepartmentForGetV1AllOf,
    "DepartmentForList": DepartmentForList,
    "DepartmentForListAllOf": DepartmentForListAllOf,
    "DepartmentForListAllOfStatus": DepartmentForListAllOfStatus,
    "DepartmentGetRanking": DepartmentGetRanking,
    "DepartmentMetricsV1": DepartmentMetricsV1,
    "DepartmentMetricsV1Current": DepartmentMetricsV1Current,
    "DepartmentMetricsV1CurrentProjectResourcesInner": DepartmentMetricsV1CurrentProjectResourcesInner,
    "DepartmentMetricsV1CurrentProjectResourcesInnerAllOf": DepartmentMetricsV1CurrentProjectResourcesInnerAllOf,
    "DepartmentMetricsV1CurrentResourcesInner": DepartmentMetricsV1CurrentResourcesInner,
    "DepartmentMetricsV1CurrentResourcesInnerAllOf": DepartmentMetricsV1CurrentResourcesInnerAllOf,
    "DepartmentMetricsV1Metadata": DepartmentMetricsV1Metadata,
    "DepartmentMetricsV1Response": DepartmentMetricsV1Response,
    "DepartmentNodePoolResources": DepartmentNodePoolResources,
    "DepartmentProjectChild": DepartmentProjectChild,
    "DepartmentProjectParent": DepartmentProjectParent,
    "DepartmentStatus": DepartmentStatus,
    "DepartmentStatusUpdateRequest": DepartmentStatusUpdateRequest,
    "DepartmentSumResponse": DepartmentSumResponse,
    "DepartmentSync": DepartmentSync,
    "DepartmentUpdateRequest": DepartmentUpdateRequest,
    "DepartmentUpdateRequestAllOf": DepartmentUpdateRequestAllOf,
    "DepartmentV1AccessControl": DepartmentV1AccessControl,
    "DepartmentV1CommonFields": DepartmentV1CommonFields,
    "DepartmentV1CommonFieldsResources": DepartmentV1CommonFieldsResources,
    "DepartmentV1CommonFieldsResponse": DepartmentV1CommonFieldsResponse,
    "DepartmentV1CommonFieldsResponseResources": DepartmentV1CommonFieldsResponseResources,
    "DepartmentV1CreationRequest": DepartmentV1CreationRequest,
    "DepartmentsMetricsV1Response": DepartmentsMetricsV1Response,
    "DepartmentsMetricsV1ResponseDataInner": DepartmentsMetricsV1ResponseDataInner,
    "DepartmentsMetricsV1ResponseDataInnerAllOf": DepartmentsMetricsV1ResponseDataInnerAllOf,
    "DepartmentsResourcesResetRequest": DepartmentsResourcesResetRequest,
    "Descendants": Descendants,
    "DeservedResource": DeservedResource,
    "FailedNodePoolResourcesPatchRequest": FailedNodePoolResourcesPatchRequest,
    "GetDepartmentNodePoolResources200Response": GetDepartmentNodePoolResources200Response,
    "GetDepartments200Response": GetDepartments200Response,
    "GetDepartmentsRanking200Response": GetDepartmentsRanking200Response,
    "GetDepartmentsV1200ResponseInner": GetDepartmentsV1200ResponseInner,
    "GetInternalDeletedOrgUnits200Response": GetInternalDeletedOrgUnits200Response,
    "GetInternalOrgTree200Response": GetInternalOrgTree200Response,
    "GetNodeTypes200Response": GetNodeTypes200Response,
    "GetNodeTypes200ResponseNodeTypesInner": GetNodeTypes200ResponseNodeTypesInner,
    "GetProjectNodePoolResources200Response": GetProjectNodePoolResources200Response,
    "GetProjects200Response": GetProjects200Response,
    "GetProjectsMetadata200Response": GetProjectsMetadata200Response,
    "GetProjectsMetadata200ResponseProjectsInner": GetProjectsMetadata200ResponseProjectsInner,
    "GetProjectsRanking200Response": GetProjectsRanking200Response,
    "GetRanks200Response": GetRanks200Response,
    "GetSyncDepartments200Response": GetSyncDepartments200Response,
    "GetSyncProjects200Response": GetSyncProjects200Response,
    "HttpResponse": HttpResponse,
    "ListReports200Response": ListReports200Response,
    "MeasurementResponse": MeasurementResponse,
    "MeasurementResponseValuesInner": MeasurementResponseValuesInner,
    "MetricsResourceDataV1": MetricsResourceDataV1,
    "MetricsResourcesDataV1": MetricsResourcesDataV1,
    "MetricsResponse": MetricsResponse,
    "ModelError": ModelError,
    "NodeAffinityV1": NodeAffinityV1,
    "NodeAffinityV1SelectedTypesInner": NodeAffinityV1SelectedTypesInner,
    "NodePool": NodePool,
    "NodePoolAssignedResourcesV1": NodePoolAssignedResourcesV1,
    "NodePoolAssignedResourcesV1Cpu": NodePoolAssignedResourcesV1Cpu,
    "NodePoolAssignedResourcesV1Gpu": NodePoolAssignedResourcesV1Gpu,
    "NodePoolAssignedResourcesV1Memory": NodePoolAssignedResourcesV1Memory,
    "NodePoolAssignedResourcesV1Response": NodePoolAssignedResourcesV1Response,
    "NodePoolAssignedResourcesV1ResponseCpu": NodePoolAssignedResourcesV1ResponseCpu,
    "NodePoolAssignedResourcesV1ResponseGpu": NodePoolAssignedResourcesV1ResponseGpu,
    "NodePoolAssignedResourcesV1ResponseMemory": NodePoolAssignedResourcesV1ResponseMemory,
    "NodePoolAssignedResourcesV1ResponseNodePool": NodePoolAssignedResourcesV1ResponseNodePool,
    "NodePoolResourcesPatchRequest": NodePoolResourcesPatchRequest,
    "NodeType": NodeType,
    "NodeTypeCreateRequest": NodeTypeCreateRequest,
    "NodeTypeUpdateRequest": NodeTypeUpdateRequest,
    "NodeTypeV1": NodeTypeV1,
    "NodeTypesPerWorkload": NodeTypesPerWorkload,
    "NonNullResource": NonNullResource,
    "NonNullResourceV1": NonNullResourceV1,
    "OrgUnit": OrgUnit,
    "OrgUnitSetting": OrgUnitSetting,
    "OrgUnitSettingsGetResponse": OrgUnitSettingsGetResponse,
    "OrgUnitSingleSetting": OrgUnitSingleSetting,
    "OrgUnitsNodepoolResourcesResetRequest": OrgUnitsNodepoolResourcesResetRequest,
    "OrgUnitsNodepoolSummaryInner": OrgUnitsNodepoolSummaryInner,
    "OrgsNodePoolSummary200Response": OrgsNodePoolSummary200Response,
    "OvertimeData": OvertimeData,
    "OvertimeDataRange24hData": OvertimeDataRange24hData,
    "OvertimeRangeData": OvertimeRangeData,
    "Project": Project,
    "ProjectAggregatedResourcesV1": ProjectAggregatedResourcesV1,
    "ProjectAggregatedResourcesV1Cpu": ProjectAggregatedResourcesV1Cpu,
    "ProjectAggregatedResourcesV1Gpu": ProjectAggregatedResourcesV1Gpu,
    "ProjectAggregatedResourcesV1Memory": ProjectAggregatedResourcesV1Memory,
    "ProjectAggregatedResourcesV1NodePool": ProjectAggregatedResourcesV1NodePool,
    "ProjectAllOf": ProjectAllOf,
    "ProjectCreationRequest": ProjectCreationRequest,
    "ProjectCreationRequestAllOf": ProjectCreationRequestAllOf,
    "ProjectEffective": ProjectEffective,
    "ProjectGetRanking": ProjectGetRanking,
    "ProjectInSyncRequest": ProjectInSyncRequest,
    "ProjectMetadata": ProjectMetadata,
    "ProjectMetricsV1": ProjectMetricsV1,
    "ProjectMetricsV1Current": ProjectMetricsV1Current,
    "ProjectMetricsV1Metadata": ProjectMetricsV1Metadata,
    "ProjectMetricsV1Response": ProjectMetricsV1Response,
    "ProjectNodePoolResources": ProjectNodePoolResources,
    "ProjectQuotasV1Inner": ProjectQuotasV1Inner,
    "ProjectStatus": ProjectStatus,
    "ProjectStatusUpdateRequest": ProjectStatusUpdateRequest,
    "ProjectSync": ProjectSync,
    "ProjectSyncPermissions": ProjectSyncPermissions,
    "ProjectUpdateRequest": ProjectUpdateRequest,
    "ProjectUpdateRequestAllOf": ProjectUpdateRequestAllOf,
    "ProjectV1": ProjectV1,
    "ProjectV1AllOf": ProjectV1AllOf,
    "ProjectV1AllOfStatus": ProjectV1AllOfStatus,
    "ProjectV1CommonFields": ProjectV1CommonFields,
    "ProjectV1CommonFieldsPermissions": ProjectV1CommonFieldsPermissions,
    "ProjectV1CommonFieldsResources": ProjectV1CommonFieldsResources,
    "ProjectV1CreationRequest": ProjectV1CreationRequest,
    "ProjectV1NodeAffinity": ProjectV1NodeAffinity,
    "ProjectV1NodeAffinityInteractive": ProjectV1NodeAffinityInteractive,
    "ProjectV1NodeAffinityResponse": ProjectV1NodeAffinityResponse,
    "ProjectV1NodeAffinityResponseInteractive": ProjectV1NodeAffinityResponseInteractive,
    "ProjectV1NodeAffinityResponseTrain": ProjectV1NodeAffinityResponseTrain,
    "ProjectV1NodeAffinityResponseTrainSelectedTypesInner": ProjectV1NodeAffinityResponseTrainSelectedTypesInner,
    "ProjectV1NodeAffinityTrain": ProjectV1NodeAffinityTrain,
    "ProjectV1ResponseCommonFields": ProjectV1ResponseCommonFields,
    "ProjectV1ResponseCommonFieldsPermissions": ProjectV1ResponseCommonFieldsPermissions,
    "ProjectV1ResponseCommonFieldsResources": ProjectV1ResponseCommonFieldsResources,
    "ProjectV1StatusRequest": ProjectV1StatusRequest,
    "ProjectsMetricsV1Response": ProjectsMetricsV1Response,
    "ProjectsMetricsV1ResponseDataInner": ProjectsMetricsV1ResponseDataInner,
    "ProjectsResourcesResetRequest": ProjectsResourcesResetRequest,
    "ProjectsResourcesResetRequestAllOf": ProjectsResourcesResetRequestAllOf,
    "QuotaStatus": QuotaStatus,
    "QuotaStatusNodePool": QuotaStatusNodePool,
    "QuotaStatusNodePoolAllOf": QuotaStatusNodePoolAllOf,
    "QuotaStatusResource": QuotaStatusResource,
    "QuotaStatusResourceListV1": QuotaStatusResourceListV1,
    "QuotaStatusV1Inner": QuotaStatusV1Inner,
    "Range": Range,
    "Rank": Rank,
    "Report": Report,
    "ReportCalculatedFields": ReportCalculatedFields,
    "ReportRequestFields": ReportRequestFields,
    "Resource": Resource,
    "ResourceV1": ResourceV1,
    "ResourceV1Response": ResourceV1Response,
    "Resources": Resources,
    "ResourcesCpu": ResourcesCpu,
    "ResourcesDataOvertimeV1": ResourcesDataOvertimeV1,
    "ResourcesDataOvertimeV1AllOf": ResourcesDataOvertimeV1AllOf,
    "ResourcesGpu": ResourcesGpu,
    "ResourcesMemory": ResourcesMemory,
    "ResourcesMemoryAllOf": ResourcesMemoryAllOf,
    "ResourcesNodePool": ResourcesNodePool,
    "ResourcesNullable": ResourcesNullable,
    "ResourcesNullableGpu": ResourcesNullableGpu,
    "ResourcesPerNodePool": ResourcesPerNodePool,
    "ResourcesPerNodePoolGpu": ResourcesPerNodePoolGpu,
    "ResourcesPerNodePoolNullable": ResourcesPerNodePoolNullable,
    "ResourcesPerNodePoolNullableGpu": ResourcesPerNodePoolNullableGpu,
    "ResourcesSync": ResourcesSync,
    "ResourcesSyncGpu": ResourcesSyncGpu,
    "ResourcesSyncMemory": ResourcesSyncMemory,
    "SchedulingRules": SchedulingRules,
    "SubmissionError": SubmissionError,
    "SumResponse": SumResponse,
    "SyncPatchRequestInner": SyncPatchRequestInner,
    "SyncProjectsRequest": SyncProjectsRequest,
    "TelemetryResponse": TelemetryResponse,
    "TelemetryResponseValuesInner": TelemetryResponseValuesInner,
    "TelemetryResponseValuesInnerGroupsInner": TelemetryResponseValuesInnerGroupsInner,
    "TimeRangeV1": TimeRangeV1,
    "TotalResources": TotalResources,
    "UpdateOrgUnitSettingByKeyRequest": UpdateOrgUnitSettingByKeyRequest,
}

export class ObjectSerializer {
    public static findCorrectType(data: any, expectedType: string) {
        if (data == undefined) {
            return expectedType;
        } else if (primitives.indexOf(expectedType.toLowerCase()) !== -1) {
            return expectedType;
        } else if (expectedType === "Date") {
            return expectedType;
        } else {
            if (enumsMap[expectedType]) {
                return expectedType;
            }

            if (!typeMap[expectedType]) {
                return expectedType; // w/e we don't know the type
            }

            // Check the discriminator
            let discriminatorProperty = typeMap[expectedType].discriminator;
            if (discriminatorProperty == null) {
                return expectedType; // the type does not have a discriminator. use it.
            } else {
                if (data[discriminatorProperty]) {
                    var discriminatorType = data[discriminatorProperty];
                    if(typeMap[discriminatorType]){
                        return discriminatorType; // use the type given in the discriminator
                    } else {
                        return expectedType; // discriminator did not map to a type
                    }
                } else {
                    return expectedType; // discriminator was not present (or an empty string)
                }
            }
        }
    }

    public static serialize(data: any, type: string) {
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.serialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return data.toISOString();
        } else {
            if (enumsMap[type]) {
                return data;
            }
            if (!typeMap[type]) { // in case we dont know the type
                return data;
            }

            // Get the actual type of this object
            type = this.findCorrectType(data, type);

            // get the map for the correct type.
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            let instance: {[index: string]: any} = {};
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.baseName] = ObjectSerializer.serialize(data[attributeType.name], attributeType.type);
            }
            return instance;
        }
    }

    public static deserialize(data: any, type: string) {
        // polymorphism may change the actual type.
        type = ObjectSerializer.findCorrectType(data, type);
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.deserialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return new Date(data);
        } else {
            if (enumsMap[type]) {// is Enum
                return data;
            }

            if (!typeMap[type]) { // dont know the type
                return data;
            }
            let instance = new typeMap[type]();
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.name] = ObjectSerializer.deserialize(data[attributeType.baseName], attributeType.type);
            }
            return instance;
        }
    }
}

export interface Authentication {
    /**
    * Apply authentication settings to header and query params.
    */
    applyToRequest(requestOptions: localVarRequest.Options): Promise<void> | void;
}

export class HttpBasicAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        requestOptions.auth = {
            username: this.username, password: this.password
        }
    }
}

export class HttpBearerAuth implements Authentication {
    public accessToken: string | (() => string) = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            const accessToken = typeof this.accessToken === 'function'
                            ? this.accessToken()
                            : this.accessToken;
            requestOptions.headers["Authorization"] = "Bearer " + accessToken;
        }
    }
}

export class ApiKeyAuth implements Authentication {
    public apiKey: string = '';

    constructor(private location: string, private paramName: string) {
    }

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (this.location == "query") {
            (<any>requestOptions.qs)[this.paramName] = this.apiKey;
        } else if (this.location == "header" && requestOptions && requestOptions.headers) {
            requestOptions.headers[this.paramName] = this.apiKey;
        } else if (this.location == 'cookie' && requestOptions && requestOptions.headers) {
            if (requestOptions.headers['Cookie']) {
                requestOptions.headers['Cookie'] += '; ' + this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
            else {
                requestOptions.headers['Cookie'] = this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
        }
    }
}

export class OAuth implements Authentication {
    public accessToken: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            requestOptions.headers["Authorization"] = "Bearer " + this.accessToken;
        }
    }
}

export class VoidAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(_: localVarRequest.Options): void {
        // Do nothing
    }
}

export type Interceptor = (requestOptions: localVarRequest.Options) => (Promise<void> | void);
