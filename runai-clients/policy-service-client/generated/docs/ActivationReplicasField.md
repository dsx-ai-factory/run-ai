# ActivationReplicasField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActivationReplicas** | Pointer to **NullableInt32** | The number of replicas to run when scaling-up from zero. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 

## Methods

### NewActivationReplicasField

`func NewActivationReplicasField() *ActivationReplicasField`

NewActivationReplicasField instantiates a new ActivationReplicasField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewActivationReplicasFieldWithDefaults

`func NewActivationReplicasFieldWithDefaults() *ActivationReplicasField`

NewActivationReplicasFieldWithDefaults instantiates a new ActivationReplicasField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActivationReplicas

`func (o *ActivationReplicasField) GetActivationReplicas() int32`

GetActivationReplicas returns the ActivationReplicas field if non-nil, zero value otherwise.

### GetActivationReplicasOk

`func (o *ActivationReplicasField) GetActivationReplicasOk() (*int32, bool)`

GetActivationReplicasOk returns a tuple with the ActivationReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActivationReplicas

`func (o *ActivationReplicasField) SetActivationReplicas(v int32)`

SetActivationReplicas sets ActivationReplicas field to given value.

### HasActivationReplicas

`func (o *ActivationReplicasField) HasActivationReplicas() bool`

HasActivationReplicas returns a boolean if a field has been set.

### SetActivationReplicasNil

`func (o *ActivationReplicasField) SetActivationReplicasNil(b bool)`

 SetActivationReplicasNil sets the value for ActivationReplicas to be an explicit nil

### UnsetActivationReplicas
`func (o *ActivationReplicasField) UnsetActivationReplicas()`

UnsetActivationReplicas ensures that no value is present for ActivationReplicas, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


