# EnvironmentCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetCreationRequest**](AssetCreationRequest.md) |  | 
**Spec** | [**EnvironmentAssetSpec**](EnvironmentAssetSpec.md) |  | 

## Methods

### NewEnvironmentCreationRequest

`func NewEnvironmentCreationRequest(meta AssetCreationRequest, spec EnvironmentAssetSpec, ) *EnvironmentCreationRequest`

NewEnvironmentCreationRequest instantiates a new EnvironmentCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentCreationRequestWithDefaults

`func NewEnvironmentCreationRequestWithDefaults() *EnvironmentCreationRequest`

NewEnvironmentCreationRequestWithDefaults instantiates a new EnvironmentCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *EnvironmentCreationRequest) GetMeta() AssetCreationRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *EnvironmentCreationRequest) GetMetaOk() (*AssetCreationRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *EnvironmentCreationRequest) SetMeta(v AssetCreationRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *EnvironmentCreationRequest) GetSpec() EnvironmentAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *EnvironmentCreationRequest) GetSpecOk() (*EnvironmentAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *EnvironmentCreationRequest) SetSpec(v EnvironmentAssetSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


