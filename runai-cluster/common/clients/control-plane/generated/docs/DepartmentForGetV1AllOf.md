# DepartmentForGetV1AllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantId** | Pointer to **int32** | The tenant id this cluster belongs to. | [optional] 
**ClusterUuid** | Pointer to **string** | The cluster UUID this department belongs to. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The creation date of the department. | [optional] 
**Id** | Pointer to **int32** | The unique id identifying the department. | [optional] 
**Projects** | [**[]ProjectV1**](ProjectV1.md) | Projects under this department. | 
**ProjectsDeservedGpus** | Pointer to **string** | Deprecated. Instead, use &#39;nodePoolsResources&#39; field. Total deserved GPUs of the projects under this department - as string. | [optional] 
**DepartmentAdmins** | Pointer to **[]string** | Id&#39;s of users with department admin role that are assigned to managed the department | [optional] 
**QuotaStatuses** | Pointer to **NullableString** |  | [optional] 
**ProjectAggregatedNodePoolsResources** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewDepartmentForGetV1AllOf

`func NewDepartmentForGetV1AllOf(projects []ProjectV1, ) *DepartmentForGetV1AllOf`

NewDepartmentForGetV1AllOf instantiates a new DepartmentForGetV1AllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentForGetV1AllOfWithDefaults

`func NewDepartmentForGetV1AllOfWithDefaults() *DepartmentForGetV1AllOf`

NewDepartmentForGetV1AllOfWithDefaults instantiates a new DepartmentForGetV1AllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantId

`func (o *DepartmentForGetV1AllOf) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *DepartmentForGetV1AllOf) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *DepartmentForGetV1AllOf) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *DepartmentForGetV1AllOf) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *DepartmentForGetV1AllOf) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *DepartmentForGetV1AllOf) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *DepartmentForGetV1AllOf) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *DepartmentForGetV1AllOf) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetCreatedAt

`func (o *DepartmentForGetV1AllOf) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DepartmentForGetV1AllOf) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DepartmentForGetV1AllOf) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DepartmentForGetV1AllOf) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetId

`func (o *DepartmentForGetV1AllOf) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DepartmentForGetV1AllOf) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DepartmentForGetV1AllOf) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DepartmentForGetV1AllOf) HasId() bool`

HasId returns a boolean if a field has been set.

### GetProjects

`func (o *DepartmentForGetV1AllOf) GetProjects() []ProjectV1`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *DepartmentForGetV1AllOf) GetProjectsOk() (*[]ProjectV1, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *DepartmentForGetV1AllOf) SetProjects(v []ProjectV1)`

SetProjects sets Projects field to given value.


### GetProjectsDeservedGpus

`func (o *DepartmentForGetV1AllOf) GetProjectsDeservedGpus() string`

GetProjectsDeservedGpus returns the ProjectsDeservedGpus field if non-nil, zero value otherwise.

### GetProjectsDeservedGpusOk

`func (o *DepartmentForGetV1AllOf) GetProjectsDeservedGpusOk() (*string, bool)`

GetProjectsDeservedGpusOk returns a tuple with the ProjectsDeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectsDeservedGpus

`func (o *DepartmentForGetV1AllOf) SetProjectsDeservedGpus(v string)`

SetProjectsDeservedGpus sets ProjectsDeservedGpus field to given value.

### HasProjectsDeservedGpus

`func (o *DepartmentForGetV1AllOf) HasProjectsDeservedGpus() bool`

HasProjectsDeservedGpus returns a boolean if a field has been set.

### GetDepartmentAdmins

`func (o *DepartmentForGetV1AllOf) GetDepartmentAdmins() []string`

GetDepartmentAdmins returns the DepartmentAdmins field if non-nil, zero value otherwise.

### GetDepartmentAdminsOk

`func (o *DepartmentForGetV1AllOf) GetDepartmentAdminsOk() (*[]string, bool)`

GetDepartmentAdminsOk returns a tuple with the DepartmentAdmins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentAdmins

`func (o *DepartmentForGetV1AllOf) SetDepartmentAdmins(v []string)`

SetDepartmentAdmins sets DepartmentAdmins field to given value.

### HasDepartmentAdmins

`func (o *DepartmentForGetV1AllOf) HasDepartmentAdmins() bool`

HasDepartmentAdmins returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *DepartmentForGetV1AllOf) GetQuotaStatuses() string`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *DepartmentForGetV1AllOf) GetQuotaStatusesOk() (*string, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *DepartmentForGetV1AllOf) SetQuotaStatuses(v string)`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *DepartmentForGetV1AllOf) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.

### SetQuotaStatusesNil

`func (o *DepartmentForGetV1AllOf) SetQuotaStatusesNil(b bool)`

 SetQuotaStatusesNil sets the value for QuotaStatuses to be an explicit nil

### UnsetQuotaStatuses
`func (o *DepartmentForGetV1AllOf) UnsetQuotaStatuses()`

UnsetQuotaStatuses ensures that no value is present for QuotaStatuses, not even an explicit nil
### GetProjectAggregatedNodePoolsResources

`func (o *DepartmentForGetV1AllOf) GetProjectAggregatedNodePoolsResources() string`

GetProjectAggregatedNodePoolsResources returns the ProjectAggregatedNodePoolsResources field if non-nil, zero value otherwise.

### GetProjectAggregatedNodePoolsResourcesOk

`func (o *DepartmentForGetV1AllOf) GetProjectAggregatedNodePoolsResourcesOk() (*string, bool)`

GetProjectAggregatedNodePoolsResourcesOk returns a tuple with the ProjectAggregatedNodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectAggregatedNodePoolsResources

`func (o *DepartmentForGetV1AllOf) SetProjectAggregatedNodePoolsResources(v string)`

SetProjectAggregatedNodePoolsResources sets ProjectAggregatedNodePoolsResources field to given value.

### HasProjectAggregatedNodePoolsResources

`func (o *DepartmentForGetV1AllOf) HasProjectAggregatedNodePoolsResources() bool`

HasProjectAggregatedNodePoolsResources returns a boolean if a field has been set.

### SetProjectAggregatedNodePoolsResourcesNil

`func (o *DepartmentForGetV1AllOf) SetProjectAggregatedNodePoolsResourcesNil(b bool)`

 SetProjectAggregatedNodePoolsResourcesNil sets the value for ProjectAggregatedNodePoolsResources to be an explicit nil

### UnsetProjectAggregatedNodePoolsResources
`func (o *DepartmentForGetV1AllOf) UnsetProjectAggregatedNodePoolsResources()`

UnsetProjectAggregatedNodePoolsResources ensures that no value is present for ProjectAggregatedNodePoolsResources, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


