# WorkloadHistoryMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreationTimestamp** | **float32** | the creation timestamp of the workload history item | 
**Type** | **string** | the type of the workload history item | 

## Methods

### NewWorkloadHistoryMeta

`func NewWorkloadHistoryMeta(creationTimestamp float32, type_ string, ) *WorkloadHistoryMeta`

NewWorkloadHistoryMeta instantiates a new WorkloadHistoryMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadHistoryMetaWithDefaults

`func NewWorkloadHistoryMetaWithDefaults() *WorkloadHistoryMeta`

NewWorkloadHistoryMetaWithDefaults instantiates a new WorkloadHistoryMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreationTimestamp

`func (o *WorkloadHistoryMeta) GetCreationTimestamp() float32`

GetCreationTimestamp returns the CreationTimestamp field if non-nil, zero value otherwise.

### GetCreationTimestampOk

`func (o *WorkloadHistoryMeta) GetCreationTimestampOk() (*float32, bool)`

GetCreationTimestampOk returns a tuple with the CreationTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreationTimestamp

`func (o *WorkloadHistoryMeta) SetCreationTimestamp(v float32)`

SetCreationTimestamp sets CreationTimestamp field to given value.


### GetType

`func (o *WorkloadHistoryMeta) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *WorkloadHistoryMeta) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *WorkloadHistoryMeta) SetType(v string)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


