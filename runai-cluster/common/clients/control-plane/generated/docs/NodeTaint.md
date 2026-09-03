# NodeTaint

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | **string** | The taint key to be applied to a node. | 
**Value** | Pointer to **string** | The taint value corresponding to the taint key. | [optional] 
**Effect** | **string** | The effect of the taint on pods that do not tolerate the taint. | 

## Methods

### NewNodeTaint

`func NewNodeTaint(key string, effect string, ) *NodeTaint`

NewNodeTaint instantiates a new NodeTaint object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeTaintWithDefaults

`func NewNodeTaintWithDefaults() *NodeTaint`

NewNodeTaintWithDefaults instantiates a new NodeTaint object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *NodeTaint) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *NodeTaint) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *NodeTaint) SetKey(v string)`

SetKey sets Key field to given value.


### GetValue

`func (o *NodeTaint) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *NodeTaint) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *NodeTaint) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *NodeTaint) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetEffect

`func (o *NodeTaint) GetEffect() string`

GetEffect returns the Effect field if non-nil, zero value otherwise.

### GetEffectOk

`func (o *NodeTaint) GetEffectOk() (*string, bool)`

GetEffectOk returns a tuple with the Effect field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffect

`func (o *NodeTaint) SetEffect(v string)`

SetEffect sets Effect field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


