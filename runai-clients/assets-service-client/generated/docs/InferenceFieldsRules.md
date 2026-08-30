# InferenceFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Autoscaling** | Pointer to [**NullableAutoScalingRules**](AutoScalingRules.md) |  | [optional] 
**ServingConfiguration** | Pointer to [**NullableServingConfigurationRules**](ServingConfigurationRules.md) |  | [optional] 

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
### GetServingConfiguration

`func (o *InferenceFieldsRules) GetServingConfiguration() ServingConfigurationRules`

GetServingConfiguration returns the ServingConfiguration field if non-nil, zero value otherwise.

### GetServingConfigurationOk

`func (o *InferenceFieldsRules) GetServingConfigurationOk() (*ServingConfigurationRules, bool)`

GetServingConfigurationOk returns a tuple with the ServingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingConfiguration

`func (o *InferenceFieldsRules) SetServingConfiguration(v ServingConfigurationRules)`

SetServingConfiguration sets ServingConfiguration field to given value.

### HasServingConfiguration

`func (o *InferenceFieldsRules) HasServingConfiguration() bool`

HasServingConfiguration returns a boolean if a field has been set.

### SetServingConfigurationNil

`func (o *InferenceFieldsRules) SetServingConfigurationNil(b bool)`

 SetServingConfigurationNil sets the value for ServingConfiguration to be an explicit nil

### UnsetServingConfiguration
`func (o *InferenceFieldsRules) UnsetServingConfiguration()`

UnsetServingConfiguration ensures that no value is present for ServingConfiguration, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


