# PVCListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | [**[]PVCAsset**](PVCAsset.md) |  | 

## Methods

### NewPVCListResponse

`func NewPVCListResponse(entries []PVCAsset, ) *PVCListResponse`

NewPVCListResponse instantiates a new PVCListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCListResponseWithDefaults

`func NewPVCListResponseWithDefaults() *PVCListResponse`

NewPVCListResponseWithDefaults instantiates a new PVCListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *PVCListResponse) GetEntries() []PVCAsset`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *PVCListResponse) GetEntriesOk() (*[]PVCAsset, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *PVCListResponse) SetEntries(v []PVCAsset)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


