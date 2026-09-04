# DockerRegistryUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetUpdateRequest**](AssetUpdateRequest.md) |  | 
**Spec** | Pointer to [**NullableDockerRegistryUpdateSpec**](DockerRegistryUpdateSpec.md) |  | [optional] 

## Methods

### NewDockerRegistryUpdateRequest

`func NewDockerRegistryUpdateRequest(meta AssetUpdateRequest, ) *DockerRegistryUpdateRequest`

NewDockerRegistryUpdateRequest instantiates a new DockerRegistryUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDockerRegistryUpdateRequestWithDefaults

`func NewDockerRegistryUpdateRequestWithDefaults() *DockerRegistryUpdateRequest`

NewDockerRegistryUpdateRequestWithDefaults instantiates a new DockerRegistryUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DockerRegistryUpdateRequest) GetMeta() AssetUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DockerRegistryUpdateRequest) GetMetaOk() (*AssetUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DockerRegistryUpdateRequest) SetMeta(v AssetUpdateRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *DockerRegistryUpdateRequest) GetSpec() DockerRegistryUpdateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DockerRegistryUpdateRequest) GetSpecOk() (*DockerRegistryUpdateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DockerRegistryUpdateRequest) SetSpec(v DockerRegistryUpdateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DockerRegistryUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### SetSpecNil

`func (o *DockerRegistryUpdateRequest) SetSpecNil(b bool)`

 SetSpecNil sets the value for Spec to be an explicit nil

### UnsetSpec
`func (o *DockerRegistryUpdateRequest) UnsetSpec()`

UnsetSpec ensures that no value is present for Spec, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


