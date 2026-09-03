# InferenceFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServingPort** | Pointer to [**NullableServingPort**](ServingPort.md) |  | [optional] 
**Autoscaling** | Pointer to [**NullableAutoScaling**](AutoScaling.md) |  | [optional] 

## Methods

### NewInferenceFields

`func NewInferenceFields() *InferenceFields`

NewInferenceFields instantiates a new InferenceFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceFieldsWithDefaults

`func NewInferenceFieldsWithDefaults() *InferenceFields`

NewInferenceFieldsWithDefaults instantiates a new InferenceFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServingPort

`func (o *InferenceFields) GetServingPort() ServingPort`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *InferenceFields) GetServingPortOk() (*ServingPort, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *InferenceFields) SetServingPort(v ServingPort)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *InferenceFields) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *InferenceFields) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *InferenceFields) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetAutoscaling

`func (o *InferenceFields) GetAutoscaling() AutoScaling`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *InferenceFields) GetAutoscalingOk() (*AutoScaling, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *InferenceFields) SetAutoscaling(v AutoScaling)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *InferenceFields) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *InferenceFields) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *InferenceFields) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


