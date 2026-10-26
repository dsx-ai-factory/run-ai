# RevisionSyncFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Name** | **string** |  | 
**Id** | **string** |  | 
**WorkloadId** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**DepartmentId** | **string** |  | 
**ProjectId** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**RequestedResources** | Pointer to [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | [optional] 
**RequestedPods** | Pointer to [**RequestedPods**](RequestedPods.md) |  | [optional] 
**RequestedNodePools** | Pointer to **[]string** |  | [optional] 
**Images** | Pointer to **[]string** |  | [optional] 
**EnvironmentVariables** | Pointer to **map[string]string** |  | [optional] 
**Command** | Pointer to **string** |  | [optional] 
**Arguments** | Pointer to **string** |  | [optional] 
**ChildrenIds** | Pointer to [**[]WorkloadChildrenIdsInner**](WorkloadChildrenIdsInner.md) |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Phase** | [**Phase**](Phase.md) |  | 
**PhaseMessage** | Pointer to **string** |  | [optional] 
**PhaseUpdatedAt** | Pointer to **time.Time** |  | [optional] 
**AdditionalFields** | Pointer to **string** |  | [optional] 

## Methods

### NewRevisionSyncFields

`func NewRevisionSyncFields(type_ string, name string, id string, workloadId string, clusterId string, departmentId string, projectId string, createdAt time.Time, phase Phase, ) *RevisionSyncFields`

NewRevisionSyncFields instantiates a new RevisionSyncFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRevisionSyncFieldsWithDefaults

`func NewRevisionSyncFieldsWithDefaults() *RevisionSyncFields`

NewRevisionSyncFieldsWithDefaults instantiates a new RevisionSyncFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *RevisionSyncFields) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *RevisionSyncFields) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *RevisionSyncFields) SetType(v string)`

SetType sets Type field to given value.


### GetName

`func (o *RevisionSyncFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RevisionSyncFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RevisionSyncFields) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *RevisionSyncFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RevisionSyncFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RevisionSyncFields) SetId(v string)`

SetId sets Id field to given value.


### GetWorkloadId

`func (o *RevisionSyncFields) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *RevisionSyncFields) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *RevisionSyncFields) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetClusterId

`func (o *RevisionSyncFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *RevisionSyncFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *RevisionSyncFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetDepartmentId

`func (o *RevisionSyncFields) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *RevisionSyncFields) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *RevisionSyncFields) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.


### GetProjectId

`func (o *RevisionSyncFields) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *RevisionSyncFields) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *RevisionSyncFields) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetCreatedAt

`func (o *RevisionSyncFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *RevisionSyncFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *RevisionSyncFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetRequestedResources

`func (o *RevisionSyncFields) GetRequestedResources() WorkloadRequestResources`

GetRequestedResources returns the RequestedResources field if non-nil, zero value otherwise.

### GetRequestedResourcesOk

`func (o *RevisionSyncFields) GetRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetRequestedResourcesOk returns a tuple with the RequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedResources

`func (o *RevisionSyncFields) SetRequestedResources(v WorkloadRequestResources)`

SetRequestedResources sets RequestedResources field to given value.

### HasRequestedResources

`func (o *RevisionSyncFields) HasRequestedResources() bool`

HasRequestedResources returns a boolean if a field has been set.

### SetRequestedResourcesNil

`func (o *RevisionSyncFields) SetRequestedResourcesNil(b bool)`

 SetRequestedResourcesNil sets the value for RequestedResources to be an explicit nil

### UnsetRequestedResources
`func (o *RevisionSyncFields) UnsetRequestedResources()`

UnsetRequestedResources ensures that no value is present for RequestedResources, not even an explicit nil
### GetRequestedPods

`func (o *RevisionSyncFields) GetRequestedPods() RequestedPods`

GetRequestedPods returns the RequestedPods field if non-nil, zero value otherwise.

### GetRequestedPodsOk

`func (o *RevisionSyncFields) GetRequestedPodsOk() (*RequestedPods, bool)`

GetRequestedPodsOk returns a tuple with the RequestedPods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedPods

`func (o *RevisionSyncFields) SetRequestedPods(v RequestedPods)`

SetRequestedPods sets RequestedPods field to given value.

### HasRequestedPods

`func (o *RevisionSyncFields) HasRequestedPods() bool`

HasRequestedPods returns a boolean if a field has been set.

### GetRequestedNodePools

`func (o *RevisionSyncFields) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *RevisionSyncFields) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *RevisionSyncFields) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *RevisionSyncFields) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetImages

`func (o *RevisionSyncFields) GetImages() []string`

GetImages returns the Images field if non-nil, zero value otherwise.

### GetImagesOk

`func (o *RevisionSyncFields) GetImagesOk() (*[]string, bool)`

GetImagesOk returns a tuple with the Images field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImages

`func (o *RevisionSyncFields) SetImages(v []string)`

SetImages sets Images field to given value.

### HasImages

`func (o *RevisionSyncFields) HasImages() bool`

HasImages returns a boolean if a field has been set.

### GetEnvironmentVariables

`func (o *RevisionSyncFields) GetEnvironmentVariables() map[string]string`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *RevisionSyncFields) GetEnvironmentVariablesOk() (*map[string]string, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *RevisionSyncFields) SetEnvironmentVariables(v map[string]string)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *RevisionSyncFields) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### GetCommand

`func (o *RevisionSyncFields) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *RevisionSyncFields) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *RevisionSyncFields) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *RevisionSyncFields) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetArguments

`func (o *RevisionSyncFields) GetArguments() string`

GetArguments returns the Arguments field if non-nil, zero value otherwise.

### GetArgumentsOk

`func (o *RevisionSyncFields) GetArgumentsOk() (*string, bool)`

GetArgumentsOk returns a tuple with the Arguments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArguments

`func (o *RevisionSyncFields) SetArguments(v string)`

SetArguments sets Arguments field to given value.

### HasArguments

`func (o *RevisionSyncFields) HasArguments() bool`

HasArguments returns a boolean if a field has been set.

### GetChildrenIds

`func (o *RevisionSyncFields) GetChildrenIds() []WorkloadChildrenIdsInner`

GetChildrenIds returns the ChildrenIds field if non-nil, zero value otherwise.

### GetChildrenIdsOk

`func (o *RevisionSyncFields) GetChildrenIdsOk() (*[]WorkloadChildrenIdsInner, bool)`

GetChildrenIdsOk returns a tuple with the ChildrenIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildrenIds

`func (o *RevisionSyncFields) SetChildrenIds(v []WorkloadChildrenIdsInner)`

SetChildrenIds sets ChildrenIds field to given value.

### HasChildrenIds

`func (o *RevisionSyncFields) HasChildrenIds() bool`

HasChildrenIds returns a boolean if a field has been set.

### GetConditions

`func (o *RevisionSyncFields) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *RevisionSyncFields) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *RevisionSyncFields) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *RevisionSyncFields) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetPhase

`func (o *RevisionSyncFields) GetPhase() Phase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *RevisionSyncFields) GetPhaseOk() (*Phase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *RevisionSyncFields) SetPhase(v Phase)`

SetPhase sets Phase field to given value.


### GetPhaseMessage

`func (o *RevisionSyncFields) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *RevisionSyncFields) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *RevisionSyncFields) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *RevisionSyncFields) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetPhaseUpdatedAt

`func (o *RevisionSyncFields) GetPhaseUpdatedAt() time.Time`

GetPhaseUpdatedAt returns the PhaseUpdatedAt field if non-nil, zero value otherwise.

### GetPhaseUpdatedAtOk

`func (o *RevisionSyncFields) GetPhaseUpdatedAtOk() (*time.Time, bool)`

GetPhaseUpdatedAtOk returns a tuple with the PhaseUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseUpdatedAt

`func (o *RevisionSyncFields) SetPhaseUpdatedAt(v time.Time)`

SetPhaseUpdatedAt sets PhaseUpdatedAt field to given value.

### HasPhaseUpdatedAt

`func (o *RevisionSyncFields) HasPhaseUpdatedAt() bool`

HasPhaseUpdatedAt returns a boolean if a field has been set.

### GetAdditionalFields

`func (o *RevisionSyncFields) GetAdditionalFields() string`

GetAdditionalFields returns the AdditionalFields field if non-nil, zero value otherwise.

### GetAdditionalFieldsOk

`func (o *RevisionSyncFields) GetAdditionalFieldsOk() (*string, bool)`

GetAdditionalFieldsOk returns a tuple with the AdditionalFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalFields

`func (o *RevisionSyncFields) SetAdditionalFields(v string)`

SetAdditionalFields sets AdditionalFields field to given value.

### HasAdditionalFields

`func (o *RevisionSyncFields) HasAdditionalFields() bool`

HasAdditionalFields returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


