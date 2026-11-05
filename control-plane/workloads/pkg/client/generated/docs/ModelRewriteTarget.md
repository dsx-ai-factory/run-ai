# ModelRewriteTarget

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Model** | **string** | The model name sent to the serving pod in place of the requested one. | 
**Weight** | **int32** | The share of requests for this target, relative to the other targets of the same rule. Weights do not need to sum to 100. | 

## Methods

### NewModelRewriteTarget

`func NewModelRewriteTarget(model string, weight int32, ) *ModelRewriteTarget`

NewModelRewriteTarget instantiates a new ModelRewriteTarget object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelRewriteTargetWithDefaults

`func NewModelRewriteTargetWithDefaults() *ModelRewriteTarget`

NewModelRewriteTargetWithDefaults instantiates a new ModelRewriteTarget object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetModel

`func (o *ModelRewriteTarget) GetModel() string`

GetModel returns the Model field if non-nil, zero value otherwise.

### GetModelOk

`func (o *ModelRewriteTarget) GetModelOk() (*string, bool)`

GetModelOk returns a tuple with the Model field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModel

`func (o *ModelRewriteTarget) SetModel(v string)`

SetModel sets Model field to given value.


### GetWeight

`func (o *ModelRewriteTarget) GetWeight() int32`

GetWeight returns the Weight field if non-nil, zero value otherwise.

### GetWeightOk

`func (o *ModelRewriteTarget) GetWeightOk() (*int32, bool)`

GetWeightOk returns a tuple with the Weight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWeight

`func (o *ModelRewriteTarget) SetWeight(v int32)`

SetWeight sets Weight field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


