# AIApplicationForSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Source** | Pointer to **string** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**DepartmentId** | Pointer to **string** |  | [optional] 
**DepartmentName** | Pointer to **string** |  | [optional] 
**ProjectId** | Pointer to **string** |  | [optional] 
**ProjectName** | Pointer to **string** |  | [optional] 
**K8sPhase** | Pointer to **string** |  | [optional] 
**Resources** | Pointer to [**[]AIApplicationSyncResource**](AIApplicationSyncResource.md) |  | [optional] 
**Action** | Pointer to [**AIApplicationSyncAction**](AIApplicationSyncAction.md) |  | [optional] 
**ActionInfo** | Pointer to [**NullableAIApplicationSyncActionInfo**](AIApplicationSyncActionInfo.md) |  | [optional] 

## Methods

### NewAIApplicationForSync

`func NewAIApplicationForSync() *AIApplicationForSync`

NewAIApplicationForSync instantiates a new AIApplicationForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationForSyncWithDefaults

`func NewAIApplicationForSyncWithDefaults() *AIApplicationForSync`

NewAIApplicationForSyncWithDefaults instantiates a new AIApplicationForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AIApplicationForSync) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AIApplicationForSync) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AIApplicationForSync) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *AIApplicationForSync) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *AIApplicationForSync) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AIApplicationForSync) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AIApplicationForSync) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *AIApplicationForSync) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSource

`func (o *AIApplicationForSync) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *AIApplicationForSync) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *AIApplicationForSync) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *AIApplicationForSync) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetType

`func (o *AIApplicationForSync) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AIApplicationForSync) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AIApplicationForSync) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *AIApplicationForSync) HasType() bool`

HasType returns a boolean if a field has been set.

### GetDepartmentId

`func (o *AIApplicationForSync) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *AIApplicationForSync) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *AIApplicationForSync) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *AIApplicationForSync) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetDepartmentName

`func (o *AIApplicationForSync) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *AIApplicationForSync) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *AIApplicationForSync) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *AIApplicationForSync) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetProjectId

`func (o *AIApplicationForSync) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AIApplicationForSync) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AIApplicationForSync) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *AIApplicationForSync) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetProjectName

`func (o *AIApplicationForSync) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *AIApplicationForSync) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *AIApplicationForSync) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *AIApplicationForSync) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetK8sPhase

`func (o *AIApplicationForSync) GetK8sPhase() string`

GetK8sPhase returns the K8sPhase field if non-nil, zero value otherwise.

### GetK8sPhaseOk

`func (o *AIApplicationForSync) GetK8sPhaseOk() (*string, bool)`

GetK8sPhaseOk returns a tuple with the K8sPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhase

`func (o *AIApplicationForSync) SetK8sPhase(v string)`

SetK8sPhase sets K8sPhase field to given value.

### HasK8sPhase

`func (o *AIApplicationForSync) HasK8sPhase() bool`

HasK8sPhase returns a boolean if a field has been set.

### GetResources

`func (o *AIApplicationForSync) GetResources() []AIApplicationSyncResource`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *AIApplicationForSync) GetResourcesOk() (*[]AIApplicationSyncResource, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *AIApplicationForSync) SetResources(v []AIApplicationSyncResource)`

SetResources sets Resources field to given value.

### HasResources

`func (o *AIApplicationForSync) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetAction

`func (o *AIApplicationForSync) GetAction() AIApplicationSyncAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *AIApplicationForSync) GetActionOk() (*AIApplicationSyncAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *AIApplicationForSync) SetAction(v AIApplicationSyncAction)`

SetAction sets Action field to given value.

### HasAction

`func (o *AIApplicationForSync) HasAction() bool`

HasAction returns a boolean if a field has been set.

### GetActionInfo

`func (o *AIApplicationForSync) GetActionInfo() AIApplicationSyncActionInfo`

GetActionInfo returns the ActionInfo field if non-nil, zero value otherwise.

### GetActionInfoOk

`func (o *AIApplicationForSync) GetActionInfoOk() (*AIApplicationSyncActionInfo, bool)`

GetActionInfoOk returns a tuple with the ActionInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionInfo

`func (o *AIApplicationForSync) SetActionInfo(v AIApplicationSyncActionInfo)`

SetActionInfo sets ActionInfo field to given value.

### HasActionInfo

`func (o *AIApplicationForSync) HasActionInfo() bool`

HasActionInfo returns a boolean if a field has been set.

### SetActionInfoNil

`func (o *AIApplicationForSync) SetActionInfoNil(b bool)`

 SetActionInfoNil sets the value for ActionInfo to be an explicit nil

### UnsetActionInfo
`func (o *AIApplicationForSync) UnsetActionInfo()`

UnsetActionInfo ensures that no value is present for ActionInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


