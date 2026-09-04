# SecretAssetUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetUpdateRequest**](AssetUpdateRequest.md) |  | 
**Spec** | [**SecretAssetSpec**](SecretAssetSpec.md) |  | 

## Methods

### NewSecretAssetUpdateRequest

`func NewSecretAssetUpdateRequest(meta AssetUpdateRequest, spec SecretAssetSpec, ) *SecretAssetUpdateRequest`

NewSecretAssetUpdateRequest instantiates a new SecretAssetUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretAssetUpdateRequestWithDefaults

`func NewSecretAssetUpdateRequestWithDefaults() *SecretAssetUpdateRequest`

NewSecretAssetUpdateRequestWithDefaults instantiates a new SecretAssetUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *SecretAssetUpdateRequest) GetMeta() AssetUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *SecretAssetUpdateRequest) GetMetaOk() (*AssetUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *SecretAssetUpdateRequest) SetMeta(v AssetUpdateRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *SecretAssetUpdateRequest) GetSpec() SecretAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *SecretAssetUpdateRequest) GetSpecOk() (*SecretAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *SecretAssetUpdateRequest) SetSpec(v SecretAssetSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


