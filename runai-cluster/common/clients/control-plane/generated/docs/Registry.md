# Registry

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**RegistrySpec**](RegistrySpec.md) |  | 

## Methods

### NewRegistry

`func NewRegistry(meta AssetMeta, spec RegistrySpec, ) *Registry`

NewRegistry instantiates a new Registry object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryWithDefaults

`func NewRegistryWithDefaults() *Registry`

NewRegistryWithDefaults instantiates a new Registry object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *Registry) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *Registry) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *Registry) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *Registry) GetSpec() RegistrySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Registry) GetSpecOk() (*RegistrySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Registry) SetSpec(v RegistrySpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


