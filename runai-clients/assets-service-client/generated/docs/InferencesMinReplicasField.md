# InferencesMinReplicasField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MinReplicas** | Pointer to **NullableInt32** | The minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero | [optional] 

## Methods

### NewInferencesMinReplicasField

`func NewInferencesMinReplicasField() *InferencesMinReplicasField`

NewInferencesMinReplicasField instantiates a new InferencesMinReplicasField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencesMinReplicasFieldWithDefaults

`func NewInferencesMinReplicasFieldWithDefaults() *InferencesMinReplicasField`

NewInferencesMinReplicasFieldWithDefaults instantiates a new InferencesMinReplicasField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMinReplicas

`func (o *InferencesMinReplicasField) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *InferencesMinReplicasField) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *InferencesMinReplicasField) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *InferencesMinReplicasField) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *InferencesMinReplicasField) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *InferencesMinReplicasField) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


