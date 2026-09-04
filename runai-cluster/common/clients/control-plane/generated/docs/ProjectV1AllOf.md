# ProjectV1AllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Project name. | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesV1Response**](NodePoolAssignedResourcesV1Response.md) | Resources assigned to this Project per Node Pool. | [optional] 
**Namespace** | Pointer to **string** | The name of an existing namespace to use for the project in the cluster. Supported only for cluster versions 2.12 or higher. | [optional] 
**Id** | Pointer to **int32** | Project id. | [optional] [readonly] 
**DepartmentId** | Pointer to **int32** | ID of the department that owns the project. | [optional] 
**TenantId** | Pointer to **int32** | ID of the tenant where the project is located. | [optional] 
**ClusterUuid** | Pointer to **string** | ID of the cluster where the project is located. | [optional] 
**DepartmentName** | Pointer to **string** | Name of the department where the project is located. | [optional] 
**InteractiveNodeAffinity** | Pointer to **string** |  | [optional] 
**TrainNodeAffinity** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | Creation date of the project. | [optional] 
**Status** | Pointer to [**ProjectV1AllOfStatus**](ProjectV1AllOfStatus.md) |  | [optional] 
**Phase** | Pointer to **string** | project&#39;s phase | [optional] 
**QuotaStatuses** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewProjectV1AllOf

`func NewProjectV1AllOf() *ProjectV1AllOf`

NewProjectV1AllOf instantiates a new ProjectV1AllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1AllOfWithDefaults

`func NewProjectV1AllOfWithDefaults() *ProjectV1AllOf`

NewProjectV1AllOfWithDefaults instantiates a new ProjectV1AllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProjectV1AllOf) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectV1AllOf) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectV1AllOf) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProjectV1AllOf) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *ProjectV1AllOf) GetNodePoolsResources() []NodePoolAssignedResourcesV1Response`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *ProjectV1AllOf) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesV1Response, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *ProjectV1AllOf) SetNodePoolsResources(v []NodePoolAssignedResourcesV1Response)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *ProjectV1AllOf) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetNamespace

`func (o *ProjectV1AllOf) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectV1AllOf) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectV1AllOf) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectV1AllOf) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetId

`func (o *ProjectV1AllOf) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectV1AllOf) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectV1AllOf) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *ProjectV1AllOf) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDepartmentId

`func (o *ProjectV1AllOf) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *ProjectV1AllOf) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *ProjectV1AllOf) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *ProjectV1AllOf) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetTenantId

`func (o *ProjectV1AllOf) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ProjectV1AllOf) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ProjectV1AllOf) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *ProjectV1AllOf) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *ProjectV1AllOf) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *ProjectV1AllOf) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *ProjectV1AllOf) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *ProjectV1AllOf) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetDepartmentName

`func (o *ProjectV1AllOf) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *ProjectV1AllOf) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *ProjectV1AllOf) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *ProjectV1AllOf) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetInteractiveNodeAffinity

`func (o *ProjectV1AllOf) GetInteractiveNodeAffinity() string`

GetInteractiveNodeAffinity returns the InteractiveNodeAffinity field if non-nil, zero value otherwise.

### GetInteractiveNodeAffinityOk

`func (o *ProjectV1AllOf) GetInteractiveNodeAffinityOk() (*string, bool)`

GetInteractiveNodeAffinityOk returns a tuple with the InteractiveNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveNodeAffinity

`func (o *ProjectV1AllOf) SetInteractiveNodeAffinity(v string)`

SetInteractiveNodeAffinity sets InteractiveNodeAffinity field to given value.

### HasInteractiveNodeAffinity

`func (o *ProjectV1AllOf) HasInteractiveNodeAffinity() bool`

HasInteractiveNodeAffinity returns a boolean if a field has been set.

### GetTrainNodeAffinity

`func (o *ProjectV1AllOf) GetTrainNodeAffinity() string`

GetTrainNodeAffinity returns the TrainNodeAffinity field if non-nil, zero value otherwise.

### GetTrainNodeAffinityOk

`func (o *ProjectV1AllOf) GetTrainNodeAffinityOk() (*string, bool)`

GetTrainNodeAffinityOk returns a tuple with the TrainNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainNodeAffinity

`func (o *ProjectV1AllOf) SetTrainNodeAffinity(v string)`

SetTrainNodeAffinity sets TrainNodeAffinity field to given value.

### HasTrainNodeAffinity

`func (o *ProjectV1AllOf) HasTrainNodeAffinity() bool`

HasTrainNodeAffinity returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ProjectV1AllOf) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ProjectV1AllOf) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ProjectV1AllOf) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ProjectV1AllOf) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetStatus

`func (o *ProjectV1AllOf) GetStatus() ProjectV1AllOfStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProjectV1AllOf) GetStatusOk() (*ProjectV1AllOfStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProjectV1AllOf) SetStatus(v ProjectV1AllOfStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProjectV1AllOf) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetPhase

`func (o *ProjectV1AllOf) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *ProjectV1AllOf) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *ProjectV1AllOf) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *ProjectV1AllOf) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *ProjectV1AllOf) GetQuotaStatuses() string`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *ProjectV1AllOf) GetQuotaStatusesOk() (*string, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *ProjectV1AllOf) SetQuotaStatuses(v string)`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *ProjectV1AllOf) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.

### SetQuotaStatusesNil

`func (o *ProjectV1AllOf) SetQuotaStatusesNil(b bool)`

 SetQuotaStatusesNil sets the value for QuotaStatuses to be an explicit nil

### UnsetQuotaStatuses
`func (o *ProjectV1AllOf) UnsetQuotaStatuses()`

UnsetQuotaStatuses ensures that no value is present for QuotaStatuses, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


