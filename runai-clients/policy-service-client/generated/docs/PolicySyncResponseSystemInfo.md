# PolicySyncResponseSystemInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**HighestUpdate** | Pointer to **int64** | the highest update count in the system. | [optional] 
**UpdatesPerBulk** | Pointer to **int32** | number of policy updates the system is configured to include in every bulk. | [optional] 

## Methods

### NewPolicySyncResponseSystemInfo

`func NewPolicySyncResponseSystemInfo() *PolicySyncResponseSystemInfo`

NewPolicySyncResponseSystemInfo instantiates a new PolicySyncResponseSystemInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicySyncResponseSystemInfoWithDefaults

`func NewPolicySyncResponseSystemInfoWithDefaults() *PolicySyncResponseSystemInfo`

NewPolicySyncResponseSystemInfoWithDefaults instantiates a new PolicySyncResponseSystemInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHighestUpdate

`func (o *PolicySyncResponseSystemInfo) GetHighestUpdate() int64`

GetHighestUpdate returns the HighestUpdate field if non-nil, zero value otherwise.

### GetHighestUpdateOk

`func (o *PolicySyncResponseSystemInfo) GetHighestUpdateOk() (*int64, bool)`

GetHighestUpdateOk returns a tuple with the HighestUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHighestUpdate

`func (o *PolicySyncResponseSystemInfo) SetHighestUpdate(v int64)`

SetHighestUpdate sets HighestUpdate field to given value.

### HasHighestUpdate

`func (o *PolicySyncResponseSystemInfo) HasHighestUpdate() bool`

HasHighestUpdate returns a boolean if a field has been set.

### GetUpdatesPerBulk

`func (o *PolicySyncResponseSystemInfo) GetUpdatesPerBulk() int32`

GetUpdatesPerBulk returns the UpdatesPerBulk field if non-nil, zero value otherwise.

### GetUpdatesPerBulkOk

`func (o *PolicySyncResponseSystemInfo) GetUpdatesPerBulkOk() (*int32, bool)`

GetUpdatesPerBulkOk returns a tuple with the UpdatesPerBulk field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatesPerBulk

`func (o *PolicySyncResponseSystemInfo) SetUpdatesPerBulk(v int32)`

SetUpdatesPerBulk sets UpdatesPerBulk field to given value.

### HasUpdatesPerBulk

`func (o *PolicySyncResponseSystemInfo) HasUpdatesPerBulk() bool`

HasUpdatesPerBulk returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


