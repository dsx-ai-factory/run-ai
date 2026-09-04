# GenericSecretCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetCreationRequest**](AssetCreationRequest.md) |  | 
**Spec** | [**GenericSecretCreationSpec**](GenericSecretCreationSpec.md) |  | 

## Methods

### NewGenericSecretCreationRequest

`func NewGenericSecretCreationRequest(meta AssetCreationRequest, spec GenericSecretCreationSpec, ) *GenericSecretCreationRequest`

NewGenericSecretCreationRequest instantiates a new GenericSecretCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGenericSecretCreationRequestWithDefaults

`func NewGenericSecretCreationRequestWithDefaults() *GenericSecretCreationRequest`

NewGenericSecretCreationRequestWithDefaults instantiates a new GenericSecretCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *GenericSecretCreationRequest) GetMeta() AssetCreationRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *GenericSecretCreationRequest) GetMetaOk() (*AssetCreationRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *GenericSecretCreationRequest) SetMeta(v AssetCreationRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *GenericSecretCreationRequest) GetSpec() GenericSecretCreationSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *GenericSecretCreationRequest) GetSpecOk() (*GenericSecretCreationSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *GenericSecretCreationRequest) SetSpec(v GenericSecretCreationSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


