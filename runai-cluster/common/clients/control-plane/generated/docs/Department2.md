# Department2

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
**Name** | Pointer to **string** | The name of the department. | [optional] 
**DeservedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | [optional] 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | [optional] 
**Resources** | Pointer to [**DepartmentCommonFieldsResources**](DepartmentCommonFieldsResources.md) |  | [optional] 

## Methods

### NewDepartment2

`func NewDepartment2() *Department2`

NewDepartment2 instantiates a new Department2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartment2WithDefaults

`func NewDepartment2WithDefaults() *Department2`

NewDepartment2WithDefaults instantiates a new Department2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *Department2) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *Department2) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *Department2) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *Department2) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetTenantId

`func (o *Department2) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Department2) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Department2) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Department2) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *Department2) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *Department2) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *Department2) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *Department2) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Department2) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Department2) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Department2) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Department2) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetId

`func (o *Department2) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Department2) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Department2) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *Department2) HasId() bool`

HasId returns a boolean if a field has been set.

### GetProjects

`func (o *Department2) GetProjects() []Project2`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *Department2) GetProjectsOk() (*[]Project2, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *Department2) SetProjects(v []Project2)`

SetProjects sets Projects field to given value.

### HasProjects

`func (o *Department2) HasProjects() bool`

HasProjects returns a boolean if a field has been set.

### GetProjectsDeservedGpus

`func (o *Department2) GetProjectsDeservedGpus() string`

GetProjectsDeservedGpus returns the ProjectsDeservedGpus field if non-nil, zero value otherwise.

### GetProjectsDeservedGpusOk

`func (o *Department2) GetProjectsDeservedGpusOk() (*string, bool)`

GetProjectsDeservedGpusOk returns a tuple with the ProjectsDeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectsDeservedGpus

`func (o *Department2) SetProjectsDeservedGpus(v string)`

SetProjectsDeservedGpus sets ProjectsDeservedGpus field to given value.

### HasProjectsDeservedGpus

`func (o *Department2) HasProjectsDeservedGpus() bool`

HasProjectsDeservedGpus returns a boolean if a field has been set.

### GetDepartmentAdmins

`func (o *Department2) GetDepartmentAdmins() []string`

GetDepartmentAdmins returns the DepartmentAdmins field if non-nil, zero value otherwise.

### GetDepartmentAdminsOk

`func (o *Department2) GetDepartmentAdminsOk() (*[]string, bool)`

GetDepartmentAdminsOk returns a tuple with the DepartmentAdmins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentAdmins

`func (o *Department2) SetDepartmentAdmins(v []string)`

SetDepartmentAdmins sets DepartmentAdmins field to given value.

### HasDepartmentAdmins

`func (o *Department2) HasDepartmentAdmins() bool`

HasDepartmentAdmins returns a boolean if a field has been set.

### GetName

`func (o *Department2) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Department2) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Department2) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Department2) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *Department2) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *Department2) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *Department2) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *Department2) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllowOverQuota

`func (o *Department2) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *Department2) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *Department2) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *Department2) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *Department2) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *Department2) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *Department2) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *Department2) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetResources

`func (o *Department2) GetResources() DepartmentCommonFieldsResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *Department2) GetResourcesOk() (*DepartmentCommonFieldsResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *Department2) SetResources(v DepartmentCommonFieldsResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *Department2) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


