# Revision

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Name** | **string** |  | 
**Id** | **string** |  | 
**WorkloadId** | **string** |  | 
**TenantId** | **int32** | The id of the tenant. | 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectId** | **string** | The id of the project. | 
**DepartmentId** | **string** | The id of the department. | 
**CreatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**RevisionRequestedResources** | Pointer to [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | [optional] 
**PodsRequestedResources** | Pointer to [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | [optional] 
**RequestedPods** | Pointer to [**RequestedPods**](RequestedPods.md) |  | [optional] 
**RequestedNodePools** | Pointer to **[]string** |  | [optional] 
**AllocatedResources** | Pointer to [**NullableWorkloadAllocatedResources**](WorkloadAllocatedResources.md) |  | [optional] 
**CurrentNodePools** | Pointer to **[]string** |  | [optional] 
**RunningPods** | Pointer to **int32** |  | [optional] 
**Images** | Pointer to **[]string** |  | [optional] 
**EnvironmentVariables** | Pointer to **map[string]string** |  | [optional] 
**Command** | Pointer to **string** |  | [optional] 
**Arguments** | Pointer to **string** |  | [optional] 
**ChildrenIds** | Pointer to [**[]WorkloadChildrenIdsInner**](WorkloadChildrenIdsInner.md) |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Phase** | [**Phase**](Phase.md) |  | 
**PhaseMessage** | Pointer to **string** |  | [optional] 
**PhaseUpdatedAt** | Pointer to **time.Time** |  | [optional] 
**AdditionalFields** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewRevision

`func NewRevision(type_ string, name string, id string, workloadId string, tenantId int32, clusterId string, projectId string, departmentId string, createdAt time.Time, phase Phase, ) *Revision`

NewRevision instantiates a new Revision object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRevisionWithDefaults

`func NewRevisionWithDefaults() *Revision`

NewRevisionWithDefaults instantiates a new Revision object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *Revision) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Revision) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Revision) SetType(v string)`

SetType sets Type field to given value.


### GetName

`func (o *Revision) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Revision) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Revision) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *Revision) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Revision) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Revision) SetId(v string)`

SetId sets Id field to given value.


### GetWorkloadId

`func (o *Revision) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *Revision) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *Revision) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetTenantId

`func (o *Revision) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Revision) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Revision) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetClusterId

`func (o *Revision) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Revision) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Revision) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *Revision) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Revision) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Revision) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetDepartmentId

`func (o *Revision) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Revision) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Revision) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.


### GetCreatedAt

`func (o *Revision) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Revision) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Revision) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetDeletedAt

`func (o *Revision) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Revision) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Revision) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *Revision) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *Revision) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Revision) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetRevisionRequestedResources

`func (o *Revision) GetRevisionRequestedResources() WorkloadRequestResources`

GetRevisionRequestedResources returns the RevisionRequestedResources field if non-nil, zero value otherwise.

### GetRevisionRequestedResourcesOk

`func (o *Revision) GetRevisionRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetRevisionRequestedResourcesOk returns a tuple with the RevisionRequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevisionRequestedResources

`func (o *Revision) SetRevisionRequestedResources(v WorkloadRequestResources)`

SetRevisionRequestedResources sets RevisionRequestedResources field to given value.

### HasRevisionRequestedResources

`func (o *Revision) HasRevisionRequestedResources() bool`

HasRevisionRequestedResources returns a boolean if a field has been set.

### SetRevisionRequestedResourcesNil

`func (o *Revision) SetRevisionRequestedResourcesNil(b bool)`

 SetRevisionRequestedResourcesNil sets the value for RevisionRequestedResources to be an explicit nil

### UnsetRevisionRequestedResources
`func (o *Revision) UnsetRevisionRequestedResources()`

UnsetRevisionRequestedResources ensures that no value is present for RevisionRequestedResources, not even an explicit nil
### GetPodsRequestedResources

`func (o *Revision) GetPodsRequestedResources() WorkloadRequestResources`

GetPodsRequestedResources returns the PodsRequestedResources field if non-nil, zero value otherwise.

### GetPodsRequestedResourcesOk

`func (o *Revision) GetPodsRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetPodsRequestedResourcesOk returns a tuple with the PodsRequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodsRequestedResources

`func (o *Revision) SetPodsRequestedResources(v WorkloadRequestResources)`

SetPodsRequestedResources sets PodsRequestedResources field to given value.

### HasPodsRequestedResources

`func (o *Revision) HasPodsRequestedResources() bool`

HasPodsRequestedResources returns a boolean if a field has been set.

### SetPodsRequestedResourcesNil

`func (o *Revision) SetPodsRequestedResourcesNil(b bool)`

 SetPodsRequestedResourcesNil sets the value for PodsRequestedResources to be an explicit nil

### UnsetPodsRequestedResources
`func (o *Revision) UnsetPodsRequestedResources()`

UnsetPodsRequestedResources ensures that no value is present for PodsRequestedResources, not even an explicit nil
### GetRequestedPods

`func (o *Revision) GetRequestedPods() RequestedPods`

GetRequestedPods returns the RequestedPods field if non-nil, zero value otherwise.

### GetRequestedPodsOk

`func (o *Revision) GetRequestedPodsOk() (*RequestedPods, bool)`

GetRequestedPodsOk returns a tuple with the RequestedPods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedPods

`func (o *Revision) SetRequestedPods(v RequestedPods)`

SetRequestedPods sets RequestedPods field to given value.

### HasRequestedPods

`func (o *Revision) HasRequestedPods() bool`

HasRequestedPods returns a boolean if a field has been set.

### GetRequestedNodePools

`func (o *Revision) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *Revision) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *Revision) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *Revision) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetAllocatedResources

`func (o *Revision) GetAllocatedResources() WorkloadAllocatedResources`

GetAllocatedResources returns the AllocatedResources field if non-nil, zero value otherwise.

### GetAllocatedResourcesOk

`func (o *Revision) GetAllocatedResourcesOk() (*WorkloadAllocatedResources, bool)`

GetAllocatedResourcesOk returns a tuple with the AllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResources

`func (o *Revision) SetAllocatedResources(v WorkloadAllocatedResources)`

SetAllocatedResources sets AllocatedResources field to given value.

### HasAllocatedResources

`func (o *Revision) HasAllocatedResources() bool`

HasAllocatedResources returns a boolean if a field has been set.

### SetAllocatedResourcesNil

`func (o *Revision) SetAllocatedResourcesNil(b bool)`

 SetAllocatedResourcesNil sets the value for AllocatedResources to be an explicit nil

### UnsetAllocatedResources
`func (o *Revision) UnsetAllocatedResources()`

UnsetAllocatedResources ensures that no value is present for AllocatedResources, not even an explicit nil
### GetCurrentNodePools

`func (o *Revision) GetCurrentNodePools() []string`

GetCurrentNodePools returns the CurrentNodePools field if non-nil, zero value otherwise.

### GetCurrentNodePoolsOk

`func (o *Revision) GetCurrentNodePoolsOk() (*[]string, bool)`

GetCurrentNodePoolsOk returns a tuple with the CurrentNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentNodePools

`func (o *Revision) SetCurrentNodePools(v []string)`

SetCurrentNodePools sets CurrentNodePools field to given value.

### HasCurrentNodePools

`func (o *Revision) HasCurrentNodePools() bool`

HasCurrentNodePools returns a boolean if a field has been set.

### GetRunningPods

`func (o *Revision) GetRunningPods() int32`

GetRunningPods returns the RunningPods field if non-nil, zero value otherwise.

### GetRunningPodsOk

`func (o *Revision) GetRunningPodsOk() (*int32, bool)`

GetRunningPodsOk returns a tuple with the RunningPods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunningPods

`func (o *Revision) SetRunningPods(v int32)`

SetRunningPods sets RunningPods field to given value.

### HasRunningPods

`func (o *Revision) HasRunningPods() bool`

HasRunningPods returns a boolean if a field has been set.

### GetImages

`func (o *Revision) GetImages() []string`

GetImages returns the Images field if non-nil, zero value otherwise.

### GetImagesOk

`func (o *Revision) GetImagesOk() (*[]string, bool)`

GetImagesOk returns a tuple with the Images field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImages

`func (o *Revision) SetImages(v []string)`

SetImages sets Images field to given value.

### HasImages

`func (o *Revision) HasImages() bool`

HasImages returns a boolean if a field has been set.

### GetEnvironmentVariables

`func (o *Revision) GetEnvironmentVariables() map[string]string`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *Revision) GetEnvironmentVariablesOk() (*map[string]string, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *Revision) SetEnvironmentVariables(v map[string]string)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *Revision) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### GetCommand

`func (o *Revision) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *Revision) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *Revision) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *Revision) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetArguments

`func (o *Revision) GetArguments() string`

GetArguments returns the Arguments field if non-nil, zero value otherwise.

### GetArgumentsOk

`func (o *Revision) GetArgumentsOk() (*string, bool)`

GetArgumentsOk returns a tuple with the Arguments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArguments

`func (o *Revision) SetArguments(v string)`

SetArguments sets Arguments field to given value.

### HasArguments

`func (o *Revision) HasArguments() bool`

HasArguments returns a boolean if a field has been set.

### GetChildrenIds

`func (o *Revision) GetChildrenIds() []WorkloadChildrenIdsInner`

GetChildrenIds returns the ChildrenIds field if non-nil, zero value otherwise.

### GetChildrenIdsOk

`func (o *Revision) GetChildrenIdsOk() (*[]WorkloadChildrenIdsInner, bool)`

GetChildrenIdsOk returns a tuple with the ChildrenIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildrenIds

`func (o *Revision) SetChildrenIds(v []WorkloadChildrenIdsInner)`

SetChildrenIds sets ChildrenIds field to given value.

### HasChildrenIds

`func (o *Revision) HasChildrenIds() bool`

HasChildrenIds returns a boolean if a field has been set.

### GetConditions

`func (o *Revision) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *Revision) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *Revision) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *Revision) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetPhase

`func (o *Revision) GetPhase() Phase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *Revision) GetPhaseOk() (*Phase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *Revision) SetPhase(v Phase)`

SetPhase sets Phase field to given value.


### GetPhaseMessage

`func (o *Revision) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *Revision) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *Revision) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *Revision) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetPhaseUpdatedAt

`func (o *Revision) GetPhaseUpdatedAt() time.Time`

GetPhaseUpdatedAt returns the PhaseUpdatedAt field if non-nil, zero value otherwise.

### GetPhaseUpdatedAtOk

`func (o *Revision) GetPhaseUpdatedAtOk() (*time.Time, bool)`

GetPhaseUpdatedAtOk returns a tuple with the PhaseUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseUpdatedAt

`func (o *Revision) SetPhaseUpdatedAt(v time.Time)`

SetPhaseUpdatedAt sets PhaseUpdatedAt field to given value.

### HasPhaseUpdatedAt

`func (o *Revision) HasPhaseUpdatedAt() bool`

HasPhaseUpdatedAt returns a boolean if a field has been set.

### GetAdditionalFields

`func (o *Revision) GetAdditionalFields() map[string]interface{}`

GetAdditionalFields returns the AdditionalFields field if non-nil, zero value otherwise.

### GetAdditionalFieldsOk

`func (o *Revision) GetAdditionalFieldsOk() (*map[string]interface{}, bool)`

GetAdditionalFieldsOk returns a tuple with the AdditionalFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalFields

`func (o *Revision) SetAdditionalFields(v map[string]interface{})`

SetAdditionalFields sets AdditionalFields field to given value.

### HasAdditionalFields

`func (o *Revision) HasAdditionalFields() bool`

HasAdditionalFields returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


