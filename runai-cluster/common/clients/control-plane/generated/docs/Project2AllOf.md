# Project2AllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Project name. | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Project per Node Pool. | [optional] 
**Namespace** | Pointer to **string** | The name of an existing namespace to use for the project in the cluster. Supported only for cluster versions 2.12 or higher. | [optional] 
**Id** | Pointer to **int32** | Project id. | [optional] [readonly] 
**DepartmentId** | Pointer to **int32** | ID of the department that owns the project. | [optional] 
**TenantId** | Pointer to **int32** | ID of the tenant where the project is located. | [optional] 
**ClusterUuid** | Pointer to **string** | ID of the cluster where the project is located. | [optional] 
**DepartmentName** | Pointer to **string** | Name of the department where the project is located. | [optional] 
**InteractiveNodeAffinity** | Pointer to **string** |  | [optional] 
**TrainNodeAffinity** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | Creation date of the project. | [optional] 
**Status** | Pointer to [**ProjectStatus**](ProjectStatus.md) |  | [optional] 
**Phase** | Pointer to **string** | project&#39;s phase | [optional] 

## Methods

### NewProject2AllOf

`func NewProject2AllOf() *Project2AllOf`

NewProject2AllOf instantiates a new Project2AllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProject2AllOfWithDefaults

`func NewProject2AllOfWithDefaults() *Project2AllOf`

NewProject2AllOfWithDefaults instantiates a new Project2AllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Project2AllOf) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Project2AllOf) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Project2AllOf) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Project2AllOf) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *Project2AllOf) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *Project2AllOf) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *Project2AllOf) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *Project2AllOf) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetNamespace

`func (o *Project2AllOf) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Project2AllOf) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Project2AllOf) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *Project2AllOf) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetId

`func (o *Project2AllOf) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Project2AllOf) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Project2AllOf) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *Project2AllOf) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDepartmentId

`func (o *Project2AllOf) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Project2AllOf) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Project2AllOf) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Project2AllOf) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetTenantId

`func (o *Project2AllOf) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Project2AllOf) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Project2AllOf) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Project2AllOf) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *Project2AllOf) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *Project2AllOf) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *Project2AllOf) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *Project2AllOf) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetDepartmentName

`func (o *Project2AllOf) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *Project2AllOf) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *Project2AllOf) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *Project2AllOf) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetInteractiveNodeAffinity

`func (o *Project2AllOf) GetInteractiveNodeAffinity() string`

GetInteractiveNodeAffinity returns the InteractiveNodeAffinity field if non-nil, zero value otherwise.

### GetInteractiveNodeAffinityOk

`func (o *Project2AllOf) GetInteractiveNodeAffinityOk() (*string, bool)`

GetInteractiveNodeAffinityOk returns a tuple with the InteractiveNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveNodeAffinity

`func (o *Project2AllOf) SetInteractiveNodeAffinity(v string)`

SetInteractiveNodeAffinity sets InteractiveNodeAffinity field to given value.

### HasInteractiveNodeAffinity

`func (o *Project2AllOf) HasInteractiveNodeAffinity() bool`

HasInteractiveNodeAffinity returns a boolean if a field has been set.

### GetTrainNodeAffinity

`func (o *Project2AllOf) GetTrainNodeAffinity() string`

GetTrainNodeAffinity returns the TrainNodeAffinity field if non-nil, zero value otherwise.

### GetTrainNodeAffinityOk

`func (o *Project2AllOf) GetTrainNodeAffinityOk() (*string, bool)`

GetTrainNodeAffinityOk returns a tuple with the TrainNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainNodeAffinity

`func (o *Project2AllOf) SetTrainNodeAffinity(v string)`

SetTrainNodeAffinity sets TrainNodeAffinity field to given value.

### HasTrainNodeAffinity

`func (o *Project2AllOf) HasTrainNodeAffinity() bool`

HasTrainNodeAffinity returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Project2AllOf) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Project2AllOf) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Project2AllOf) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Project2AllOf) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetStatus

`func (o *Project2AllOf) GetStatus() ProjectStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Project2AllOf) GetStatusOk() (*ProjectStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Project2AllOf) SetStatus(v ProjectStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Project2AllOf) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetPhase

`func (o *Project2AllOf) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *Project2AllOf) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *Project2AllOf) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *Project2AllOf) HasPhase() bool`

HasPhase returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


