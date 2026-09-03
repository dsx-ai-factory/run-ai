# SyncStatusUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**ScopeType** | Pointer to [**ScopeType**](ScopeType.md) |  | [optional] 
**ScopeId** | Pointer to **NullableString** | cluster, department or project id, depending on the scope type | [optional] 
**UpdateCounter** | Pointer to **int64** | number which uniquely identifies the update. For deletion, the updateCounter holds the update count of the recent applied policy, rather than of the deletion. | [optional] 
**Action** | Pointer to **string** | The action that the agent performed | [optional] 
**ErrorMessage** | Pointer to **NullableString** | optional error indicating the update has failed. | [optional] 

## Methods

### NewSyncStatusUpdate

`func NewSyncStatusUpdate() *SyncStatusUpdate`

NewSyncStatusUpdate instantiates a new SyncStatusUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncStatusUpdateWithDefaults

`func NewSyncStatusUpdateWithDefaults() *SyncStatusUpdate`

NewSyncStatusUpdateWithDefaults instantiates a new SyncStatusUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *SyncStatusUpdate) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *SyncStatusUpdate) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *SyncStatusUpdate) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *SyncStatusUpdate) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetScopeType

`func (o *SyncStatusUpdate) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *SyncStatusUpdate) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *SyncStatusUpdate) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.

### HasScopeType

`func (o *SyncStatusUpdate) HasScopeType() bool`

HasScopeType returns a boolean if a field has been set.

### GetScopeId

`func (o *SyncStatusUpdate) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *SyncStatusUpdate) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *SyncStatusUpdate) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *SyncStatusUpdate) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.

### SetScopeIdNil

`func (o *SyncStatusUpdate) SetScopeIdNil(b bool)`

 SetScopeIdNil sets the value for ScopeId to be an explicit nil

### UnsetScopeId
`func (o *SyncStatusUpdate) UnsetScopeId()`

UnsetScopeId ensures that no value is present for ScopeId, not even an explicit nil
### GetUpdateCounter

`func (o *SyncStatusUpdate) GetUpdateCounter() int64`

GetUpdateCounter returns the UpdateCounter field if non-nil, zero value otherwise.

### GetUpdateCounterOk

`func (o *SyncStatusUpdate) GetUpdateCounterOk() (*int64, bool)`

GetUpdateCounterOk returns a tuple with the UpdateCounter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCounter

`func (o *SyncStatusUpdate) SetUpdateCounter(v int64)`

SetUpdateCounter sets UpdateCounter field to given value.

### HasUpdateCounter

`func (o *SyncStatusUpdate) HasUpdateCounter() bool`

HasUpdateCounter returns a boolean if a field has been set.

### GetAction

`func (o *SyncStatusUpdate) GetAction() string`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *SyncStatusUpdate) GetActionOk() (*string, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *SyncStatusUpdate) SetAction(v string)`

SetAction sets Action field to given value.

### HasAction

`func (o *SyncStatusUpdate) HasAction() bool`

HasAction returns a boolean if a field has been set.

### GetErrorMessage

`func (o *SyncStatusUpdate) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *SyncStatusUpdate) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *SyncStatusUpdate) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.

### HasErrorMessage

`func (o *SyncStatusUpdate) HasErrorMessage() bool`

HasErrorMessage returns a boolean if a field has been set.

### SetErrorMessageNil

`func (o *SyncStatusUpdate) SetErrorMessageNil(b bool)`

 SetErrorMessageNil sets the value for ErrorMessage to be an explicit nil

### UnsetErrorMessage
`func (o *SyncStatusUpdate) UnsetErrorMessage()`

UnsetErrorMessage ensures that no value is present for ErrorMessage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


