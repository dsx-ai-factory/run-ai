# HuggingFaceModelInfoRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ModelName** | **string** | the model name | 
**HfToken** | Pointer to **NullableString** | the Hugging Face token, required only for gated models | [optional] 

## Methods

### NewHuggingFaceModelInfoRequest

`func NewHuggingFaceModelInfoRequest(modelName string, ) *HuggingFaceModelInfoRequest`

NewHuggingFaceModelInfoRequest instantiates a new HuggingFaceModelInfoRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHuggingFaceModelInfoRequestWithDefaults

`func NewHuggingFaceModelInfoRequestWithDefaults() *HuggingFaceModelInfoRequest`

NewHuggingFaceModelInfoRequestWithDefaults instantiates a new HuggingFaceModelInfoRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetModelName

`func (o *HuggingFaceModelInfoRequest) GetModelName() string`

GetModelName returns the ModelName field if non-nil, zero value otherwise.

### GetModelNameOk

`func (o *HuggingFaceModelInfoRequest) GetModelNameOk() (*string, bool)`

GetModelNameOk returns a tuple with the ModelName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelName

`func (o *HuggingFaceModelInfoRequest) SetModelName(v string)`

SetModelName sets ModelName field to given value.


### GetHfToken

`func (o *HuggingFaceModelInfoRequest) GetHfToken() string`

GetHfToken returns the HfToken field if non-nil, zero value otherwise.

### GetHfTokenOk

`func (o *HuggingFaceModelInfoRequest) GetHfTokenOk() (*string, bool)`

GetHfTokenOk returns a tuple with the HfToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHfToken

`func (o *HuggingFaceModelInfoRequest) SetHfToken(v string)`

SetHfToken sets HfToken field to given value.

### HasHfToken

`func (o *HuggingFaceModelInfoRequest) HasHfToken() bool`

HasHfToken returns a boolean if a field has been set.

### SetHfTokenNil

`func (o *HuggingFaceModelInfoRequest) SetHfTokenNil(b bool)`

 SetHfTokenNil sets the value for HfToken to be an explicit nil

### UnsetHfToken
`func (o *HuggingFaceModelInfoRequest) UnsetHfToken()`

UnsetHfToken ensures that no value is present for HfToken, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


