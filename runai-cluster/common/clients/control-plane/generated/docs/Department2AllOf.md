# Department2AllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to the Department per node pool. | [optional] 
**TenantId** | Pointer to **int32** | The tenant id this cluster belongs to. | [optional] 
**ClusterUuid** | Pointer to **string** | The cluster UUID this department belongs to. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The creation date of the department. | [optional] 
**Id** | Pointer to **int32** | The unique id identifying the department. | [optional] 
**Projects** | Pointer to [**[]Project2**](Project2.md) | Projects under this department. | [optional] 
**ProjectsDeservedGpus** | Pointer to **string** | Deprecated. Instead, use &#39;nodePoolsResources&#39; field. Total deserved GPUs of the projects under this department - as string. | [optional] 
**DepartmentAdmins** | Pointer to **[]string** | Id&#39;s of users with department admin role that are assigned to managed the department | [optional] 

## Methods

### NewDepartment2AllOf

`func NewDepartment2AllOf() *Department2AllOf`

NewDepartment2AllOf instantiates a new Department2AllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartment2AllOfWithDefaults

`func NewDepartment2AllOfWithDefaults() *Department2AllOf`

NewDepartment2AllOfWithDefaults instantiates a new Department2AllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *Department2AllOf) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *Department2AllOf) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *Department2AllOf) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *Department2AllOf) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetTenantId

`func (o *Department2AllOf) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Department2AllOf) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Department2AllOf) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Department2AllOf) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *Department2AllOf) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *Department2AllOf) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *Department2AllOf) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *Department2AllOf) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Department2AllOf) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Department2AllOf) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Department2AllOf) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Department2AllOf) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetId

`func (o *Department2AllOf) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Department2AllOf) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Department2AllOf) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *Department2AllOf) HasId() bool`

HasId returns a boolean if a field has been set.

### GetProjects

`func (o *Department2AllOf) GetProjects() []Project2`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *Department2AllOf) GetProjectsOk() (*[]Project2, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *Department2AllOf) SetProjects(v []Project2)`

SetProjects sets Projects field to given value.

### HasProjects

`func (o *Department2AllOf) HasProjects() bool`

HasProjects returns a boolean if a field has been set.

### GetProjectsDeservedGpus

`func (o *Department2AllOf) GetProjectsDeservedGpus() string`

GetProjectsDeservedGpus returns the ProjectsDeservedGpus field if non-nil, zero value otherwise.

### GetProjectsDeservedGpusOk

`func (o *Department2AllOf) GetProjectsDeservedGpusOk() (*string, bool)`

GetProjectsDeservedGpusOk returns a tuple with the ProjectsDeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectsDeservedGpus

`func (o *Department2AllOf) SetProjectsDeservedGpus(v string)`

SetProjectsDeservedGpus sets ProjectsDeservedGpus field to given value.

### HasProjectsDeservedGpus

`func (o *Department2AllOf) HasProjectsDeservedGpus() bool`

HasProjectsDeservedGpus returns a boolean if a field has been set.

### GetDepartmentAdmins

`func (o *Department2AllOf) GetDepartmentAdmins() []string`

GetDepartmentAdmins returns the DepartmentAdmins field if non-nil, zero value otherwise.

### GetDepartmentAdminsOk

`func (o *Department2AllOf) GetDepartmentAdminsOk() (*[]string, bool)`

GetDepartmentAdminsOk returns a tuple with the DepartmentAdmins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentAdmins

`func (o *Department2AllOf) SetDepartmentAdmins(v []string)`

SetDepartmentAdmins sets DepartmentAdmins field to given value.

### HasDepartmentAdmins

`func (o *Department2AllOf) HasDepartmentAdmins() bool`

HasDepartmentAdmins returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


