# PasswordUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetUpdateRequest**](AssetUpdateRequest.md) |  | 
**Spec** | Pointer to [**NullablePasswordUpdateSpec**](PasswordUpdateSpec.md) |  | [optional] 

## Methods

### NewPasswordUpdateRequest

`func NewPasswordUpdateRequest(meta AssetUpdateRequest, ) *PasswordUpdateRequest`

NewPasswordUpdateRequest instantiates a new PasswordUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordUpdateRequestWithDefaults

`func NewPasswordUpdateRequestWithDefaults() *PasswordUpdateRequest`

NewPasswordUpdateRequestWithDefaults instantiates a new PasswordUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *PasswordUpdateRequest) GetMeta() AssetUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *PasswordUpdateRequest) GetMetaOk() (*AssetUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *PasswordUpdateRequest) SetMeta(v AssetUpdateRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *PasswordUpdateRequest) GetSpec() PasswordUpdateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *PasswordUpdateRequest) GetSpecOk() (*PasswordUpdateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *PasswordUpdateRequest) SetSpec(v PasswordUpdateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *PasswordUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### SetSpecNil

`func (o *PasswordUpdateRequest) SetSpecNil(b bool)`

 SetSpecNil sets the value for Spec to be an explicit nil

### UnsetSpec
`func (o *PasswordUpdateRequest) UnsetSpec()`

UnsetSpec ensures that no value is present for Spec, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


