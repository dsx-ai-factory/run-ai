# DepartmentForGetV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesV1Response**](NodePoolAssignedResourcesV1Response.md) | Resources assigned to the Department per node pool. | [optional] 
**Name** | Pointer to **string** | The name of the department. | [optional] 
**DeservedGpus** | **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | 
**Resources** | Pointer to [**AssignedResourcesV1Response**](AssignedResourcesV1Response.md) | Deprecated. Instead, use &#39;nodePoolsResources&#39;. Total resources assigned to the Department. Can only be used in PUT/POST when there is a single Node Pool in the system. The resources returned in GET are the sum of all Node Pool Resources. | [optional] 
**TenantId** | Pointer to **int32** | The tenant id this cluster belongs to. | [optional] 
**ClusterUuid** | Pointer to **string** | The cluster UUID this department belongs to. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The creation date of the department. | [optional] 
**Id** | Pointer to **int32** | The unique id identifying the department. | [optional] 
**Projects** | [**[]ProjectV1**](ProjectV1.md) | Projects under this department. | 
**ProjectsDeservedGpus** | Pointer to **string** | Deprecated. Instead, use &#39;nodePoolsResources&#39; field. Total deserved GPUs of the projects under this department - as string. | [optional] 
**DepartmentAdmins** | Pointer to **[]string** | Id&#39;s of users with department admin role that are assigned to managed the department | [optional] 
**QuotaStatuses** | Pointer to **[]map[string]interface{}** |  | [optional] 
**ProjectAggregatedNodePoolsResources** | Pointer to **[]map[string]interface{}** |  | [optional] 

## Methods

### NewDepartmentForGetV1

`func NewDepartmentForGetV1(deservedGpus float32, maxAllowedGpus float32, projects []ProjectV1, ) *DepartmentForGetV1`

NewDepartmentForGetV1 instantiates a new DepartmentForGetV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentForGetV1WithDefaults

`func NewDepartmentForGetV1WithDefaults() *DepartmentForGetV1`

NewDepartmentForGetV1WithDefaults instantiates a new DepartmentForGetV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *DepartmentForGetV1) GetNodePoolsResources() []NodePoolAssignedResourcesV1Response`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentForGetV1) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesV1Response, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentForGetV1) SetNodePoolsResources(v []NodePoolAssignedResourcesV1Response)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentForGetV1) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetName

`func (o *DepartmentForGetV1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentForGetV1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentForGetV1) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DepartmentForGetV1) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *DepartmentForGetV1) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentForGetV1) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentForGetV1) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.


### GetAllowOverQuota

`func (o *DepartmentForGetV1) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentForGetV1) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentForGetV1) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentForGetV1) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentForGetV1) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentForGetV1) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentForGetV1) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.


### GetResources

`func (o *DepartmentForGetV1) GetResources() AssignedResourcesV1Response`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentForGetV1) GetResourcesOk() (*AssignedResourcesV1Response, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentForGetV1) SetResources(v AssignedResourcesV1Response)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentForGetV1) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetTenantId

`func (o *DepartmentForGetV1) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *DepartmentForGetV1) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *DepartmentForGetV1) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *DepartmentForGetV1) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *DepartmentForGetV1) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *DepartmentForGetV1) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *DepartmentForGetV1) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *DepartmentForGetV1) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetCreatedAt

`func (o *DepartmentForGetV1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DepartmentForGetV1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DepartmentForGetV1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DepartmentForGetV1) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetId

`func (o *DepartmentForGetV1) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DepartmentForGetV1) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DepartmentForGetV1) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DepartmentForGetV1) HasId() bool`

HasId returns a boolean if a field has been set.

### GetProjects

`func (o *DepartmentForGetV1) GetProjects() []ProjectV1`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *DepartmentForGetV1) GetProjectsOk() (*[]ProjectV1, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *DepartmentForGetV1) SetProjects(v []ProjectV1)`

SetProjects sets Projects field to given value.


### GetProjectsDeservedGpus

`func (o *DepartmentForGetV1) GetProjectsDeservedGpus() string`

GetProjectsDeservedGpus returns the ProjectsDeservedGpus field if non-nil, zero value otherwise.

### GetProjectsDeservedGpusOk

`func (o *DepartmentForGetV1) GetProjectsDeservedGpusOk() (*string, bool)`

GetProjectsDeservedGpusOk returns a tuple with the ProjectsDeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectsDeservedGpus

`func (o *DepartmentForGetV1) SetProjectsDeservedGpus(v string)`

SetProjectsDeservedGpus sets ProjectsDeservedGpus field to given value.

### HasProjectsDeservedGpus

`func (o *DepartmentForGetV1) HasProjectsDeservedGpus() bool`

HasProjectsDeservedGpus returns a boolean if a field has been set.

### GetDepartmentAdmins

`func (o *DepartmentForGetV1) GetDepartmentAdmins() []string`

GetDepartmentAdmins returns the DepartmentAdmins field if non-nil, zero value otherwise.

### GetDepartmentAdminsOk

`func (o *DepartmentForGetV1) GetDepartmentAdminsOk() (*[]string, bool)`

GetDepartmentAdminsOk returns a tuple with the DepartmentAdmins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentAdmins

`func (o *DepartmentForGetV1) SetDepartmentAdmins(v []string)`

SetDepartmentAdmins sets DepartmentAdmins field to given value.

### HasDepartmentAdmins

`func (o *DepartmentForGetV1) HasDepartmentAdmins() bool`

HasDepartmentAdmins returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *DepartmentForGetV1) GetQuotaStatuses() []map[string]interface{}`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *DepartmentForGetV1) GetQuotaStatusesOk() (*[]map[string]interface{}, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *DepartmentForGetV1) SetQuotaStatuses(v []map[string]interface{})`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *DepartmentForGetV1) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.

### SetQuotaStatusesNil

`func (o *DepartmentForGetV1) SetQuotaStatusesNil(b bool)`

 SetQuotaStatusesNil sets the value for QuotaStatuses to be an explicit nil

### UnsetQuotaStatuses
`func (o *DepartmentForGetV1) UnsetQuotaStatuses()`

UnsetQuotaStatuses ensures that no value is present for QuotaStatuses, not even an explicit nil
### GetProjectAggregatedNodePoolsResources

`func (o *DepartmentForGetV1) GetProjectAggregatedNodePoolsResources() []map[string]interface{}`

GetProjectAggregatedNodePoolsResources returns the ProjectAggregatedNodePoolsResources field if non-nil, zero value otherwise.

### GetProjectAggregatedNodePoolsResourcesOk

`func (o *DepartmentForGetV1) GetProjectAggregatedNodePoolsResourcesOk() (*[]map[string]interface{}, bool)`

GetProjectAggregatedNodePoolsResourcesOk returns a tuple with the ProjectAggregatedNodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectAggregatedNodePoolsResources

`func (o *DepartmentForGetV1) SetProjectAggregatedNodePoolsResources(v []map[string]interface{})`

SetProjectAggregatedNodePoolsResources sets ProjectAggregatedNodePoolsResources field to given value.

### HasProjectAggregatedNodePoolsResources

`func (o *DepartmentForGetV1) HasProjectAggregatedNodePoolsResources() bool`

HasProjectAggregatedNodePoolsResources returns a boolean if a field has been set.

### SetProjectAggregatedNodePoolsResourcesNil

`func (o *DepartmentForGetV1) SetProjectAggregatedNodePoolsResourcesNil(b bool)`

 SetProjectAggregatedNodePoolsResourcesNil sets the value for ProjectAggregatedNodePoolsResources to be an explicit nil

### UnsetProjectAggregatedNodePoolsResources
`func (o *DepartmentForGetV1) UnsetProjectAggregatedNodePoolsResources()`

UnsetProjectAggregatedNodePoolsResources ensures that no value is present for ProjectAggregatedNodePoolsResources, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


