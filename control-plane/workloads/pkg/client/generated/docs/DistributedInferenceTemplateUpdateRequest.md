# DistributedInferenceTemplateUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateUpdateMetaFields**](TemplateUpdateMetaFields.md) |  | 
**Spec** | Pointer to [**DistributedInferenceTemplateSpec**](DistributedInferenceTemplateSpec.md) |  | [optional] 

## Methods

### NewDistributedInferenceTemplateUpdateRequest

`func NewDistributedInferenceTemplateUpdateRequest(meta TemplateUpdateMetaFields, ) *DistributedInferenceTemplateUpdateRequest`

NewDistributedInferenceTemplateUpdateRequest instantiates a new DistributedInferenceTemplateUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceTemplateUpdateRequestWithDefaults

`func NewDistributedInferenceTemplateUpdateRequestWithDefaults() *DistributedInferenceTemplateUpdateRequest`

NewDistributedInferenceTemplateUpdateRequestWithDefaults instantiates a new DistributedInferenceTemplateUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedInferenceTemplateUpdateRequest) GetMeta() TemplateUpdateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedInferenceTemplateUpdateRequest) GetMetaOk() (*TemplateUpdateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedInferenceTemplateUpdateRequest) SetMeta(v TemplateUpdateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *DistributedInferenceTemplateUpdateRequest) GetSpec() DistributedInferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedInferenceTemplateUpdateRequest) GetSpecOk() (*DistributedInferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedInferenceTemplateUpdateRequest) SetSpec(v DistributedInferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedInferenceTemplateUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


