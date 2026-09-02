# WorkloadV2SyncInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | [**WorkloadV2SyncStatus**](WorkloadV2SyncStatus.md) |  | 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewWorkloadV2SyncInfo

`func NewWorkloadV2SyncInfo(status WorkloadV2SyncStatus, ) *WorkloadV2SyncInfo`

NewWorkloadV2SyncInfo instantiates a new WorkloadV2SyncInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2SyncInfoWithDefaults

`func NewWorkloadV2SyncInfoWithDefaults() *WorkloadV2SyncInfo`

NewWorkloadV2SyncInfoWithDefaults instantiates a new WorkloadV2SyncInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *WorkloadV2SyncInfo) GetStatus() WorkloadV2SyncStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkloadV2SyncInfo) GetStatusOk() (*WorkloadV2SyncStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkloadV2SyncInfo) SetStatus(v WorkloadV2SyncStatus)`

SetStatus sets Status field to given value.


### GetMessage

`func (o *WorkloadV2SyncInfo) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *WorkloadV2SyncInfo) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *WorkloadV2SyncInfo) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *WorkloadV2SyncInfo) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


