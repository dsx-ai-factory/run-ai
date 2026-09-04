# HuggingFaceModelInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ModelName** | **string** | the model name | 
**Gated** | Pointer to **bool** | is the model gated | [optional] 
**Authorized** | Pointer to **bool** | is the token authorized for the model | [optional] 
**Language** | Pointer to **bool** | is the model a language model | [optional] 

## Methods

### NewHuggingFaceModelInfo

`func NewHuggingFaceModelInfo(modelName string, ) *HuggingFaceModelInfo`

NewHuggingFaceModelInfo instantiates a new HuggingFaceModelInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHuggingFaceModelInfoWithDefaults

`func NewHuggingFaceModelInfoWithDefaults() *HuggingFaceModelInfo`

NewHuggingFaceModelInfoWithDefaults instantiates a new HuggingFaceModelInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetModelName

`func (o *HuggingFaceModelInfo) GetModelName() string`

GetModelName returns the ModelName field if non-nil, zero value otherwise.

### GetModelNameOk

`func (o *HuggingFaceModelInfo) GetModelNameOk() (*string, bool)`

GetModelNameOk returns a tuple with the ModelName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelName

`func (o *HuggingFaceModelInfo) SetModelName(v string)`

SetModelName sets ModelName field to given value.


### GetGated

`func (o *HuggingFaceModelInfo) GetGated() bool`

GetGated returns the Gated field if non-nil, zero value otherwise.

### GetGatedOk

`func (o *HuggingFaceModelInfo) GetGatedOk() (*bool, bool)`

GetGatedOk returns a tuple with the Gated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGated

`func (o *HuggingFaceModelInfo) SetGated(v bool)`

SetGated sets Gated field to given value.

### HasGated

`func (o *HuggingFaceModelInfo) HasGated() bool`

HasGated returns a boolean if a field has been set.

### GetAuthorized

`func (o *HuggingFaceModelInfo) GetAuthorized() bool`

GetAuthorized returns the Authorized field if non-nil, zero value otherwise.

### GetAuthorizedOk

`func (o *HuggingFaceModelInfo) GetAuthorizedOk() (*bool, bool)`

GetAuthorizedOk returns a tuple with the Authorized field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorized

`func (o *HuggingFaceModelInfo) SetAuthorized(v bool)`

SetAuthorized sets Authorized field to given value.

### HasAuthorized

`func (o *HuggingFaceModelInfo) HasAuthorized() bool`

HasAuthorized returns a boolean if a field has been set.

### GetLanguage

`func (o *HuggingFaceModelInfo) GetLanguage() bool`

GetLanguage returns the Language field if non-nil, zero value otherwise.

### GetLanguageOk

`func (o *HuggingFaceModelInfo) GetLanguageOk() (*bool, bool)`

GetLanguageOk returns a tuple with the Language field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguage

`func (o *HuggingFaceModelInfo) SetLanguage(v bool)`

SetLanguage sets Language field to given value.

### HasLanguage

`func (o *HuggingFaceModelInfo) HasLanguage() bool`

HasLanguage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


