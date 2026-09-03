# RegistryListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | [**[]Registry**](Registry.md) |  | 

## Methods

### NewRegistryListResponse

`func NewRegistryListResponse(entries []Registry, ) *RegistryListResponse`

NewRegistryListResponse instantiates a new RegistryListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryListResponseWithDefaults

`func NewRegistryListResponseWithDefaults() *RegistryListResponse`

NewRegistryListResponseWithDefaults instantiates a new RegistryListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *RegistryListResponse) GetEntries() []Registry`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *RegistryListResponse) GetEntriesOk() (*[]Registry, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *RegistryListResponse) SetEntries(v []Registry)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


