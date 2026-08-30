# PVCStatusRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Pvcs** | Pointer to [**[]PVCStatusRequestItem**](PVCStatusRequestItem.md) | List of resources which exist in the cluster and their status | [optional] 

## Methods

### NewPVCStatusRequest

`func NewPVCStatusRequest() *PVCStatusRequest`

NewPVCStatusRequest instantiates a new PVCStatusRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCStatusRequestWithDefaults

`func NewPVCStatusRequestWithDefaults() *PVCStatusRequest`

NewPVCStatusRequestWithDefaults instantiates a new PVCStatusRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPvcs

`func (o *PVCStatusRequest) GetPvcs() []PVCStatusRequestItem`

GetPvcs returns the Pvcs field if non-nil, zero value otherwise.

### GetPvcsOk

`func (o *PVCStatusRequest) GetPvcsOk() (*[]PVCStatusRequestItem, bool)`

GetPvcsOk returns a tuple with the Pvcs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvcs

`func (o *PVCStatusRequest) SetPvcs(v []PVCStatusRequestItem)`

SetPvcs sets Pvcs field to given value.

### HasPvcs

`func (o *PVCStatusRequest) HasPvcs() bool`

HasPvcs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


