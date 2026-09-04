# SyncInfoResponseItemV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Scope** | Pointer to [**Scope**](Scope.md) |  | [optional] 
**ScopeResourceId** | Pointer to **string** | the id of the scope resource, empty for tenant scope, the id of the department for department scope and the project name for project scope. | [optional] 
**SecondsFromLatestChange** | **int32** |  | 

## Methods

### NewSyncInfoResponseItemV2

`func NewSyncInfoResponseItemV2(secondsFromLatestChange int32, ) *SyncInfoResponseItemV2`

NewSyncInfoResponseItemV2 instantiates a new SyncInfoResponseItemV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncInfoResponseItemV2WithDefaults

`func NewSyncInfoResponseItemV2WithDefaults() *SyncInfoResponseItemV2`

NewSyncInfoResponseItemV2WithDefaults instantiates a new SyncInfoResponseItemV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScope

`func (o *SyncInfoResponseItemV2) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *SyncInfoResponseItemV2) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *SyncInfoResponseItemV2) SetScope(v Scope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *SyncInfoResponseItemV2) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetScopeResourceId

`func (o *SyncInfoResponseItemV2) GetScopeResourceId() string`

GetScopeResourceId returns the ScopeResourceId field if non-nil, zero value otherwise.

### GetScopeResourceIdOk

`func (o *SyncInfoResponseItemV2) GetScopeResourceIdOk() (*string, bool)`

GetScopeResourceIdOk returns a tuple with the ScopeResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeResourceId

`func (o *SyncInfoResponseItemV2) SetScopeResourceId(v string)`

SetScopeResourceId sets ScopeResourceId field to given value.

### HasScopeResourceId

`func (o *SyncInfoResponseItemV2) HasScopeResourceId() bool`

HasScopeResourceId returns a boolean if a field has been set.

### GetSecondsFromLatestChange

`func (o *SyncInfoResponseItemV2) GetSecondsFromLatestChange() int32`

GetSecondsFromLatestChange returns the SecondsFromLatestChange field if non-nil, zero value otherwise.

### GetSecondsFromLatestChangeOk

`func (o *SyncInfoResponseItemV2) GetSecondsFromLatestChangeOk() (*int32, bool)`

GetSecondsFromLatestChangeOk returns a tuple with the SecondsFromLatestChange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecondsFromLatestChange

`func (o *SyncInfoResponseItemV2) SetSecondsFromLatestChange(v int32)`

SetSecondsFromLatestChange sets SecondsFromLatestChange field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


