# HostPathUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetUpdateRequest**](AssetUpdateRequest.md) |  | 
**Spec** | [**HostPathSpec**](HostPathSpec.md) |  | 

## Methods

### NewHostPathUpdateRequest

`func NewHostPathUpdateRequest(meta AssetUpdateRequest, spec HostPathSpec, ) *HostPathUpdateRequest`

NewHostPathUpdateRequest instantiates a new HostPathUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathUpdateRequestWithDefaults

`func NewHostPathUpdateRequestWithDefaults() *HostPathUpdateRequest`

NewHostPathUpdateRequestWithDefaults instantiates a new HostPathUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *HostPathUpdateRequest) GetMeta() AssetUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *HostPathUpdateRequest) GetMetaOk() (*AssetUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *HostPathUpdateRequest) SetMeta(v AssetUpdateRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *HostPathUpdateRequest) GetSpec() HostPathSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *HostPathUpdateRequest) GetSpecOk() (*HostPathSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *HostPathUpdateRequest) SetSpec(v HostPathSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


