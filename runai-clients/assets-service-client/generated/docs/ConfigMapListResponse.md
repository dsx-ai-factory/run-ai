# ConfigMapListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | [**[]ConfigMapAsset**](ConfigMapAsset.md) |  | 

## Methods

### NewConfigMapListResponse

`func NewConfigMapListResponse(entries []ConfigMapAsset, ) *ConfigMapListResponse`

NewConfigMapListResponse instantiates a new ConfigMapListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigMapListResponseWithDefaults

`func NewConfigMapListResponseWithDefaults() *ConfigMapListResponse`

NewConfigMapListResponseWithDefaults instantiates a new ConfigMapListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *ConfigMapListResponse) GetEntries() []ConfigMapAsset`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *ConfigMapListResponse) GetEntriesOk() (*[]ConfigMapAsset, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *ConfigMapListResponse) SetEntries(v []ConfigMapAsset)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


