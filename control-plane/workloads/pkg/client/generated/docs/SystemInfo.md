# SystemInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**HighestUpdate** | Pointer to **int64** | the value of an update counter of a data in the system. | [optional] 

## Methods

### NewSystemInfo

`func NewSystemInfo() *SystemInfo`

NewSystemInfo instantiates a new SystemInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSystemInfoWithDefaults

`func NewSystemInfoWithDefaults() *SystemInfo`

NewSystemInfoWithDefaults instantiates a new SystemInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHighestUpdate

`func (o *SystemInfo) GetHighestUpdate() int64`

GetHighestUpdate returns the HighestUpdate field if non-nil, zero value otherwise.

### GetHighestUpdateOk

`func (o *SystemInfo) GetHighestUpdateOk() (*int64, bool)`

GetHighestUpdateOk returns a tuple with the HighestUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHighestUpdate

`func (o *SystemInfo) SetHighestUpdate(v int64)`

SetHighestUpdate sets HighestUpdate field to given value.

### HasHighestUpdate

`func (o *SystemInfo) HasHighestUpdate() bool`

HasHighestUpdate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


