# PolicySyncResponseResponseInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LowestUpdate** | Pointer to **int64** | lowest update count included in this bulk. zero if the bulk is empty. | [optional] 
**HighestUpdate** | Pointer to **int64** | highest update count included in this bulk. zero if the bulk is empty. | [optional] 

## Methods

### NewPolicySyncResponseResponseInfo

`func NewPolicySyncResponseResponseInfo() *PolicySyncResponseResponseInfo`

NewPolicySyncResponseResponseInfo instantiates a new PolicySyncResponseResponseInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicySyncResponseResponseInfoWithDefaults

`func NewPolicySyncResponseResponseInfoWithDefaults() *PolicySyncResponseResponseInfo`

NewPolicySyncResponseResponseInfoWithDefaults instantiates a new PolicySyncResponseResponseInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLowestUpdate

`func (o *PolicySyncResponseResponseInfo) GetLowestUpdate() int64`

GetLowestUpdate returns the LowestUpdate field if non-nil, zero value otherwise.

### GetLowestUpdateOk

`func (o *PolicySyncResponseResponseInfo) GetLowestUpdateOk() (*int64, bool)`

GetLowestUpdateOk returns a tuple with the LowestUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLowestUpdate

`func (o *PolicySyncResponseResponseInfo) SetLowestUpdate(v int64)`

SetLowestUpdate sets LowestUpdate field to given value.

### HasLowestUpdate

`func (o *PolicySyncResponseResponseInfo) HasLowestUpdate() bool`

HasLowestUpdate returns a boolean if a field has been set.

### GetHighestUpdate

`func (o *PolicySyncResponseResponseInfo) GetHighestUpdate() int64`

GetHighestUpdate returns the HighestUpdate field if non-nil, zero value otherwise.

### GetHighestUpdateOk

`func (o *PolicySyncResponseResponseInfo) GetHighestUpdateOk() (*int64, bool)`

GetHighestUpdateOk returns a tuple with the HighestUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHighestUpdate

`func (o *PolicySyncResponseResponseInfo) SetHighestUpdate(v int64)`

SetHighestUpdate sets HighestUpdate field to given value.

### HasHighestUpdate

`func (o *PolicySyncResponseResponseInfo) HasHighestUpdate() bool`

HasHighestUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


