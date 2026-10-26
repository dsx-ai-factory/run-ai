# InitialReplicasField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InitialReplicas** | Pointer to **NullableInt32** | The number of replicas to run when initializing the workload for the first time. Defaults to minReplicas, or to 1 if minReplicas is set to 0 | [optional] 

## Methods

### NewInitialReplicasField

`func NewInitialReplicasField() *InitialReplicasField`

NewInitialReplicasField instantiates a new InitialReplicasField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInitialReplicasFieldWithDefaults

`func NewInitialReplicasFieldWithDefaults() *InitialReplicasField`

NewInitialReplicasFieldWithDefaults instantiates a new InitialReplicasField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInitialReplicas

`func (o *InitialReplicasField) GetInitialReplicas() int32`

GetInitialReplicas returns the InitialReplicas field if non-nil, zero value otherwise.

### GetInitialReplicasOk

`func (o *InitialReplicasField) GetInitialReplicasOk() (*int32, bool)`

GetInitialReplicasOk returns a tuple with the InitialReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitialReplicas

`func (o *InitialReplicasField) SetInitialReplicas(v int32)`

SetInitialReplicas sets InitialReplicas field to given value.

### HasInitialReplicas

`func (o *InitialReplicasField) HasInitialReplicas() bool`

HasInitialReplicas returns a boolean if a field has been set.

### SetInitialReplicasNil

`func (o *InitialReplicasField) SetInitialReplicasNil(b bool)`

 SetInitialReplicasNil sets the value for InitialReplicas to be an explicit nil

### UnsetInitialReplicas
`func (o *InitialReplicasField) UnsetInitialReplicas()`

UnsetInitialReplicas ensures that no value is present for InitialReplicas, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


