# ModelRewrite

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Model** | **string** | The exact model name that clients put in the request body. | 
**Targets** | [**[]ModelRewriteTarget**](ModelRewriteTarget.md) | The served model names and their traffic shares. | 

## Methods

### NewModelRewrite

`func NewModelRewrite(model string, targets []ModelRewriteTarget, ) *ModelRewrite`

NewModelRewrite instantiates a new ModelRewrite object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelRewriteWithDefaults

`func NewModelRewriteWithDefaults() *ModelRewrite`

NewModelRewriteWithDefaults instantiates a new ModelRewrite object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetModel

`func (o *ModelRewrite) GetModel() string`

GetModel returns the Model field if non-nil, zero value otherwise.

### GetModelOk

`func (o *ModelRewrite) GetModelOk() (*string, bool)`

GetModelOk returns a tuple with the Model field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModel

`func (o *ModelRewrite) SetModel(v string)`

SetModel sets Model field to given value.


### GetTargets

`func (o *ModelRewrite) GetTargets() []ModelRewriteTarget`

GetTargets returns the Targets field if non-nil, zero value otherwise.

### GetTargetsOk

`func (o *ModelRewrite) GetTargetsOk() (*[]ModelRewriteTarget, bool)`

GetTargetsOk returns a tuple with the Targets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargets

`func (o *ModelRewrite) SetTargets(v []ModelRewriteTarget)`

SetTargets sets Targets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


