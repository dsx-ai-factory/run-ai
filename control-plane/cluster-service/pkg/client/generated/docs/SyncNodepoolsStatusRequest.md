# SyncNodepoolsStatusRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Nodepools** | [**[]NodepoolSyncFields**](NodepoolSyncFields.md) |  | 
**DeleteMissing** | Pointer to **bool** |  | [optional] [default to false]

## Methods

### NewSyncNodepoolsStatusRequest

`func NewSyncNodepoolsStatusRequest(nodepools []NodepoolSyncFields, ) *SyncNodepoolsStatusRequest`

NewSyncNodepoolsStatusRequest instantiates a new SyncNodepoolsStatusRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncNodepoolsStatusRequestWithDefaults

`func NewSyncNodepoolsStatusRequestWithDefaults() *SyncNodepoolsStatusRequest`

NewSyncNodepoolsStatusRequestWithDefaults instantiates a new SyncNodepoolsStatusRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodepools

`func (o *SyncNodepoolsStatusRequest) GetNodepools() []NodepoolSyncFields`

GetNodepools returns the Nodepools field if non-nil, zero value otherwise.

### GetNodepoolsOk

`func (o *SyncNodepoolsStatusRequest) GetNodepoolsOk() (*[]NodepoolSyncFields, bool)`

GetNodepoolsOk returns a tuple with the Nodepools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodepools

`func (o *SyncNodepoolsStatusRequest) SetNodepools(v []NodepoolSyncFields)`

SetNodepools sets Nodepools field to given value.


### GetDeleteMissing

`func (o *SyncNodepoolsStatusRequest) GetDeleteMissing() bool`

GetDeleteMissing returns the DeleteMissing field if non-nil, zero value otherwise.

### GetDeleteMissingOk

`func (o *SyncNodepoolsStatusRequest) GetDeleteMissingOk() (*bool, bool)`

GetDeleteMissingOk returns a tuple with the DeleteMissing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleteMissing

`func (o *SyncNodepoolsStatusRequest) SetDeleteMissing(v bool)`

SetDeleteMissing sets DeleteMissing field to given value.

### HasDeleteMissing

`func (o *SyncNodepoolsStatusRequest) HasDeleteMissing() bool`

HasDeleteMissing returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


