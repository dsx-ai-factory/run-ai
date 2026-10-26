# InferencesMaxReplicasField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MaxReplicas** | Pointer to **NullableInt32** | The maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 

## Methods

### NewInferencesMaxReplicasField

`func NewInferencesMaxReplicasField() *InferencesMaxReplicasField`

NewInferencesMaxReplicasField instantiates a new InferencesMaxReplicasField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencesMaxReplicasFieldWithDefaults

`func NewInferencesMaxReplicasFieldWithDefaults() *InferencesMaxReplicasField`

NewInferencesMaxReplicasFieldWithDefaults instantiates a new InferencesMaxReplicasField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMaxReplicas

`func (o *InferencesMaxReplicasField) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *InferencesMaxReplicasField) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *InferencesMaxReplicasField) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *InferencesMaxReplicasField) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *InferencesMaxReplicasField) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *InferencesMaxReplicasField) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


