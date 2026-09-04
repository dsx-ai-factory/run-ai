# HostPathCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetCreationRequest**](AssetCreationRequest.md) |  | 
**Spec** | [**HostPathSpec**](HostPathSpec.md) |  | 

## Methods

### NewHostPathCreationRequest

`func NewHostPathCreationRequest(meta AssetCreationRequest, spec HostPathSpec, ) *HostPathCreationRequest`

NewHostPathCreationRequest instantiates a new HostPathCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathCreationRequestWithDefaults

`func NewHostPathCreationRequestWithDefaults() *HostPathCreationRequest`

NewHostPathCreationRequestWithDefaults instantiates a new HostPathCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *HostPathCreationRequest) GetMeta() AssetCreationRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *HostPathCreationRequest) GetMetaOk() (*AssetCreationRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *HostPathCreationRequest) SetMeta(v AssetCreationRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *HostPathCreationRequest) GetSpec() HostPathSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *HostPathCreationRequest) GetSpecOk() (*HostPathSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *HostPathCreationRequest) SetSpec(v HostPathSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


