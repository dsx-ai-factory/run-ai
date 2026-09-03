# RegistryUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetUpdateRequest**](AssetUpdateRequest.md) |  | 
**Spec** | [**RegistrySpec**](RegistrySpec.md) |  | 

## Methods

### NewRegistryUpdateRequest

`func NewRegistryUpdateRequest(meta AssetUpdateRequest, spec RegistrySpec, ) *RegistryUpdateRequest`

NewRegistryUpdateRequest instantiates a new RegistryUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryUpdateRequestWithDefaults

`func NewRegistryUpdateRequestWithDefaults() *RegistryUpdateRequest`

NewRegistryUpdateRequestWithDefaults instantiates a new RegistryUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *RegistryUpdateRequest) GetMeta() AssetUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *RegistryUpdateRequest) GetMetaOk() (*AssetUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *RegistryUpdateRequest) SetMeta(v AssetUpdateRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *RegistryUpdateRequest) GetSpec() RegistrySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *RegistryUpdateRequest) GetSpecOk() (*RegistrySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *RegistryUpdateRequest) SetSpec(v RegistrySpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


