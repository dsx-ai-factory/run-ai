# ProjectStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Namespace** | Pointer to **NullableString** | the project&#39;s namespace | [optional] 
**Phase** | Pointer to [**ProjectPhase**](ProjectPhase.md) |  | [optional] 
**PhaseMessage** | Pointer to **NullableString** | the project&#39;s phase message | [optional] 
**LastUpdatedTime** | Pointer to **NullableTime** |  | [optional] 
**NodePoolQuotaStatuses** | Pointer to [**[]QuotaStatusNodePool**](QuotaStatusNodePool.md) |  | [optional] 
**QuotaStatus** | Pointer to [**QuotaStatus**](QuotaStatus.md) |  | [optional] 
**AdditionalStatusData** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewProjectStatus

`func NewProjectStatus() *ProjectStatus`

NewProjectStatus instantiates a new ProjectStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectStatusWithDefaults

`func NewProjectStatusWithDefaults() *ProjectStatus`

NewProjectStatusWithDefaults instantiates a new ProjectStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNamespace

`func (o *ProjectStatus) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectStatus) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectStatus) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectStatus) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### SetNamespaceNil

`func (o *ProjectStatus) SetNamespaceNil(b bool)`

 SetNamespaceNil sets the value for Namespace to be an explicit nil

### UnsetNamespace
`func (o *ProjectStatus) UnsetNamespace()`

UnsetNamespace ensures that no value is present for Namespace, not even an explicit nil
### GetPhase

`func (o *ProjectStatus) GetPhase() ProjectPhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *ProjectStatus) GetPhaseOk() (*ProjectPhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *ProjectStatus) SetPhase(v ProjectPhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *ProjectStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetPhaseMessage

`func (o *ProjectStatus) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *ProjectStatus) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *ProjectStatus) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *ProjectStatus) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### SetPhaseMessageNil

`func (o *ProjectStatus) SetPhaseMessageNil(b bool)`

 SetPhaseMessageNil sets the value for PhaseMessage to be an explicit nil

### UnsetPhaseMessage
`func (o *ProjectStatus) UnsetPhaseMessage()`

UnsetPhaseMessage ensures that no value is present for PhaseMessage, not even an explicit nil
### GetLastUpdatedTime

`func (o *ProjectStatus) GetLastUpdatedTime() time.Time`

GetLastUpdatedTime returns the LastUpdatedTime field if non-nil, zero value otherwise.

### GetLastUpdatedTimeOk

`func (o *ProjectStatus) GetLastUpdatedTimeOk() (*time.Time, bool)`

GetLastUpdatedTimeOk returns a tuple with the LastUpdatedTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUpdatedTime

`func (o *ProjectStatus) SetLastUpdatedTime(v time.Time)`

SetLastUpdatedTime sets LastUpdatedTime field to given value.

### HasLastUpdatedTime

`func (o *ProjectStatus) HasLastUpdatedTime() bool`

HasLastUpdatedTime returns a boolean if a field has been set.

### SetLastUpdatedTimeNil

`func (o *ProjectStatus) SetLastUpdatedTimeNil(b bool)`

 SetLastUpdatedTimeNil sets the value for LastUpdatedTime to be an explicit nil

### UnsetLastUpdatedTime
`func (o *ProjectStatus) UnsetLastUpdatedTime()`

UnsetLastUpdatedTime ensures that no value is present for LastUpdatedTime, not even an explicit nil
### GetNodePoolQuotaStatuses

`func (o *ProjectStatus) GetNodePoolQuotaStatuses() []QuotaStatusNodePool`

GetNodePoolQuotaStatuses returns the NodePoolQuotaStatuses field if non-nil, zero value otherwise.

### GetNodePoolQuotaStatusesOk

`func (o *ProjectStatus) GetNodePoolQuotaStatusesOk() (*[]QuotaStatusNodePool, bool)`

GetNodePoolQuotaStatusesOk returns a tuple with the NodePoolQuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolQuotaStatuses

`func (o *ProjectStatus) SetNodePoolQuotaStatuses(v []QuotaStatusNodePool)`

SetNodePoolQuotaStatuses sets NodePoolQuotaStatuses field to given value.

### HasNodePoolQuotaStatuses

`func (o *ProjectStatus) HasNodePoolQuotaStatuses() bool`

HasNodePoolQuotaStatuses returns a boolean if a field has been set.

### GetQuotaStatus

`func (o *ProjectStatus) GetQuotaStatus() QuotaStatus`

GetQuotaStatus returns the QuotaStatus field if non-nil, zero value otherwise.

### GetQuotaStatusOk

`func (o *ProjectStatus) GetQuotaStatusOk() (*QuotaStatus, bool)`

GetQuotaStatusOk returns a tuple with the QuotaStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatus

`func (o *ProjectStatus) SetQuotaStatus(v QuotaStatus)`

SetQuotaStatus sets QuotaStatus field to given value.

### HasQuotaStatus

`func (o *ProjectStatus) HasQuotaStatus() bool`

HasQuotaStatus returns a boolean if a field has been set.

### GetAdditionalStatusData

`func (o *ProjectStatus) GetAdditionalStatusData() map[string]interface{}`

GetAdditionalStatusData returns the AdditionalStatusData field if non-nil, zero value otherwise.

### GetAdditionalStatusDataOk

`func (o *ProjectStatus) GetAdditionalStatusDataOk() (*map[string]interface{}, bool)`

GetAdditionalStatusDataOk returns a tuple with the AdditionalStatusData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalStatusData

`func (o *ProjectStatus) SetAdditionalStatusData(v map[string]interface{})`

SetAdditionalStatusData sets AdditionalStatusData field to given value.

### HasAdditionalStatusData

`func (o *ProjectStatus) HasAdditionalStatusData() bool`

HasAdditionalStatusData returns a boolean if a field has been set.

### SetAdditionalStatusDataNil

`func (o *ProjectStatus) SetAdditionalStatusDataNil(b bool)`

 SetAdditionalStatusDataNil sets the value for AdditionalStatusData to be an explicit nil

### UnsetAdditionalStatusData
`func (o *ProjectStatus) UnsetAdditionalStatusData()`

UnsetAdditionalStatusData ensures that no value is present for AdditionalStatusData, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


