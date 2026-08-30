# AccessKeyUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetUpdateRequest**](AssetUpdateRequest.md) |  | 
**Spec** | Pointer to [**NullableAccessKeyUpdateSpec**](AccessKeyUpdateSpec.md) |  | [optional] 

## Methods

### NewAccessKeyUpdateRequest

`func NewAccessKeyUpdateRequest(meta AssetUpdateRequest, ) *AccessKeyUpdateRequest`

NewAccessKeyUpdateRequest instantiates a new AccessKeyUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessKeyUpdateRequestWithDefaults

`func NewAccessKeyUpdateRequestWithDefaults() *AccessKeyUpdateRequest`

NewAccessKeyUpdateRequestWithDefaults instantiates a new AccessKeyUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *AccessKeyUpdateRequest) GetMeta() AssetUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *AccessKeyUpdateRequest) GetMetaOk() (*AssetUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *AccessKeyUpdateRequest) SetMeta(v AssetUpdateRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *AccessKeyUpdateRequest) GetSpec() AccessKeyUpdateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *AccessKeyUpdateRequest) GetSpecOk() (*AccessKeyUpdateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *AccessKeyUpdateRequest) SetSpec(v AccessKeyUpdateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *AccessKeyUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### SetSpecNil

`func (o *AccessKeyUpdateRequest) SetSpecNil(b bool)`

 SetSpecNil sets the value for Spec to be an explicit nil

### UnsetSpec
`func (o *AccessKeyUpdateRequest) UnsetSpec()`

UnsetSpec ensures that no value is present for Spec, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


