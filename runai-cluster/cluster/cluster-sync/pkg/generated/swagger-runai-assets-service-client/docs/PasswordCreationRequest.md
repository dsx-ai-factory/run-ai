# PasswordCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetCreationRequest**](AssetCreationRequest.md) |  | 
**Spec** | [**PasswordCreationSpec**](PasswordCreationSpec.md) |  | 

## Methods

### NewPasswordCreationRequest

`func NewPasswordCreationRequest(meta AssetCreationRequest, spec PasswordCreationSpec, ) *PasswordCreationRequest`

NewPasswordCreationRequest instantiates a new PasswordCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordCreationRequestWithDefaults

`func NewPasswordCreationRequestWithDefaults() *PasswordCreationRequest`

NewPasswordCreationRequestWithDefaults instantiates a new PasswordCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *PasswordCreationRequest) GetMeta() AssetCreationRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *PasswordCreationRequest) GetMetaOk() (*AssetCreationRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *PasswordCreationRequest) SetMeta(v AssetCreationRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *PasswordCreationRequest) GetSpec() PasswordCreationSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *PasswordCreationRequest) GetSpecOk() (*PasswordCreationSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *PasswordCreationRequest) SetSpec(v PasswordCreationSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


