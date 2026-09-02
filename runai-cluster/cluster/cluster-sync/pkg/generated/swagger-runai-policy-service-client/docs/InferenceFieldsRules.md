# InferenceFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServingPort** | Pointer to [**NullableServingPortRules**](ServingPortRules.md) |  | [optional] 
**Autoscaling** | Pointer to [**NullableAutoScalingRules**](AutoScalingRules.md) |  | [optional] 

## Methods

### NewInferenceFieldsRules

`func NewInferenceFieldsRules() *InferenceFieldsRules`

NewInferenceFieldsRules instantiates a new InferenceFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceFieldsRulesWithDefaults

`func NewInferenceFieldsRulesWithDefaults() *InferenceFieldsRules`

NewInferenceFieldsRulesWithDefaults instantiates a new InferenceFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServingPort

`func (o *InferenceFieldsRules) GetServingPort() ServingPortRules`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *InferenceFieldsRules) GetServingPortOk() (*ServingPortRules, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *InferenceFieldsRules) SetServingPort(v ServingPortRules)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *InferenceFieldsRules) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *InferenceFieldsRules) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *InferenceFieldsRules) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetAutoscaling

`func (o *InferenceFieldsRules) GetAutoscaling() AutoScalingRules`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *InferenceFieldsRules) GetAutoscalingOk() (*AutoScalingRules, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *InferenceFieldsRules) SetAutoscaling(v AutoScalingRules)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *InferenceFieldsRules) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *InferenceFieldsRules) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *InferenceFieldsRules) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


