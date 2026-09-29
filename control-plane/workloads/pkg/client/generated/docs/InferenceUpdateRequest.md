# InferenceUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | Pointer to [**InferenceUpdateSpecSpec**](InferenceUpdateSpecSpec.md) |  | [optional] 

## Methods

### NewInferenceUpdateRequest

`func NewInferenceUpdateRequest() *InferenceUpdateRequest`

NewInferenceUpdateRequest instantiates a new InferenceUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceUpdateRequestWithDefaults

`func NewInferenceUpdateRequestWithDefaults() *InferenceUpdateRequest`

NewInferenceUpdateRequestWithDefaults instantiates a new InferenceUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpec

`func (o *InferenceUpdateRequest) GetSpec() InferenceUpdateSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InferenceUpdateRequest) GetSpecOk() (*InferenceUpdateSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InferenceUpdateRequest) SetSpec(v InferenceUpdateSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InferenceUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


