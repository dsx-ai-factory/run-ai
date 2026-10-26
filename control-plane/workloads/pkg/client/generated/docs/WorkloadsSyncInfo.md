# WorkloadsSyncInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UpdateCount** | Pointer to **int64** | the value of an update counter of a data in the system. | [optional] 
**Workload** | Pointer to [**SupersetWorkload**](SupersetWorkload.md) |  | [optional] 

## Methods

### NewWorkloadsSyncInfo

`func NewWorkloadsSyncInfo() *WorkloadsSyncInfo`

NewWorkloadsSyncInfo instantiates a new WorkloadsSyncInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadsSyncInfoWithDefaults

`func NewWorkloadsSyncInfoWithDefaults() *WorkloadsSyncInfo`

NewWorkloadsSyncInfoWithDefaults instantiates a new WorkloadsSyncInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUpdateCount

`func (o *WorkloadsSyncInfo) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *WorkloadsSyncInfo) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *WorkloadsSyncInfo) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.

### HasUpdateCount

`func (o *WorkloadsSyncInfo) HasUpdateCount() bool`

HasUpdateCount returns a boolean if a field has been set.

### GetWorkload

`func (o *WorkloadsSyncInfo) GetWorkload() SupersetWorkload`

GetWorkload returns the Workload field if non-nil, zero value otherwise.

### GetWorkloadOk

`func (o *WorkloadsSyncInfo) GetWorkloadOk() (*SupersetWorkload, bool)`

GetWorkloadOk returns a tuple with the Workload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkload

`func (o *WorkloadsSyncInfo) SetWorkload(v SupersetWorkload)`

SetWorkload sets Workload field to given value.

### HasWorkload

`func (o *WorkloadsSyncInfo) HasWorkload() bool`

HasWorkload returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


