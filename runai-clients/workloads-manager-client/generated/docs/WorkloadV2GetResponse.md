# WorkloadV2GetResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Workload** | [**WorkloadV2**](WorkloadV2.md) |  | 
**SyncInfo** | Pointer to [**WorkloadV2SyncInfo**](WorkloadV2SyncInfo.md) |  | [optional] 
**UpdateCount** | Pointer to **int64** | Version of the workload returned by this read. | [optional]

## Methods

### NewWorkloadV2GetResponse

`func NewWorkloadV2GetResponse(workload WorkloadV2, ) *WorkloadV2GetResponse`

NewWorkloadV2GetResponse instantiates a new WorkloadV2GetResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2GetResponseWithDefaults

`func NewWorkloadV2GetResponseWithDefaults() *WorkloadV2GetResponse`

NewWorkloadV2GetResponseWithDefaults instantiates a new WorkloadV2GetResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkload

`func (o *WorkloadV2GetResponse) GetWorkload() WorkloadV2`

GetWorkload returns the Workload field if non-nil, zero value otherwise.

### GetWorkloadOk

`func (o *WorkloadV2GetResponse) GetWorkloadOk() (*WorkloadV2, bool)`

GetWorkloadOk returns a tuple with the Workload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkload

`func (o *WorkloadV2GetResponse) SetWorkload(v WorkloadV2)`

SetWorkload sets Workload field to given value.


### GetSyncInfo

`func (o *WorkloadV2GetResponse) GetSyncInfo() WorkloadV2SyncInfo`

GetSyncInfo returns the SyncInfo field if non-nil, zero value otherwise.

### GetSyncInfoOk

`func (o *WorkloadV2GetResponse) GetSyncInfoOk() (*WorkloadV2SyncInfo, bool)`

GetSyncInfoOk returns a tuple with the SyncInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSyncInfo

`func (o *WorkloadV2GetResponse) SetSyncInfo(v WorkloadV2SyncInfo)`

SetSyncInfo sets SyncInfo field to given value.

### HasSyncInfo

`func (o *WorkloadV2GetResponse) HasSyncInfo() bool`

HasSyncInfo returns a boolean if a field has been set.

### GetUpdateCount

`func (o *WorkloadV2GetResponse) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *WorkloadV2GetResponse) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *WorkloadV2GetResponse) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.

### HasUpdateCount

`func (o *WorkloadV2GetResponse) HasUpdateCount() bool`

HasUpdateCount returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


