# SecretsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableSecretInstance**](SecretInstance.md) |  | [optional] 
**Instances** | Pointer to [**[]SecretInstance1**](SecretInstance1.md) | Set of secret volumes to use in the workload | [optional] 

## Methods

### NewSecretsDefaults

`func NewSecretsDefaults() *SecretsDefaults`

NewSecretsDefaults instantiates a new SecretsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretsDefaultsWithDefaults

`func NewSecretsDefaultsWithDefaults() *SecretsDefaults`

NewSecretsDefaultsWithDefaults instantiates a new SecretsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *SecretsDefaults) GetAttributes() SecretInstance`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *SecretsDefaults) GetAttributesOk() (*SecretInstance, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *SecretsDefaults) SetAttributes(v SecretInstance)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *SecretsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *SecretsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *SecretsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *SecretsDefaults) GetInstances() []SecretInstance1`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *SecretsDefaults) GetInstancesOk() (*[]SecretInstance1, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *SecretsDefaults) SetInstances(v []SecretInstance1)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *SecretsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *SecretsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *SecretsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


