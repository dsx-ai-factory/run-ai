# PVCListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Pvcs** | [**[]PVCListInfo**](PVCListInfo.md) |  | 

## Methods

### NewPVCListResponse

`func NewPVCListResponse(pvcs []PVCListInfo, ) *PVCListResponse`

NewPVCListResponse instantiates a new PVCListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCListResponseWithDefaults

`func NewPVCListResponseWithDefaults() *PVCListResponse`

NewPVCListResponseWithDefaults instantiates a new PVCListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPvcs

`func (o *PVCListResponse) GetPvcs() []PVCListInfo`

GetPvcs returns the Pvcs field if non-nil, zero value otherwise.

### GetPvcsOk

`func (o *PVCListResponse) GetPvcsOk() (*[]PVCListInfo, bool)`

GetPvcsOk returns a tuple with the Pvcs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvcs

`func (o *PVCListResponse) SetPvcs(v []PVCListInfo)`

SetPvcs sets Pvcs field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


