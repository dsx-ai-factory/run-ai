# GenericSecretUpdateSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | optional name of existing secret in the cluster from which the secret details should be taken. If omitted, you will have to provide the credentials details in keyValuePairs field. The provided credentials are encrypted into the control plane database and cloned to a kubernetes secret in the cluster. | [optional] 
**KeyValuePairs** | Pointer to [**[]KeyValuePair**](KeyValuePair.md) | Set of key-value pairs to store in the new credential. | [optional] 

## Methods

### NewGenericSecretUpdateSpec

`func NewGenericSecretUpdateSpec() *GenericSecretUpdateSpec`

NewGenericSecretUpdateSpec instantiates a new GenericSecretUpdateSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGenericSecretUpdateSpecWithDefaults

`func NewGenericSecretUpdateSpecWithDefaults() *GenericSecretUpdateSpec`

NewGenericSecretUpdateSpecWithDefaults instantiates a new GenericSecretUpdateSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *GenericSecretUpdateSpec) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *GenericSecretUpdateSpec) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *GenericSecretUpdateSpec) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *GenericSecretUpdateSpec) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *GenericSecretUpdateSpec) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *GenericSecretUpdateSpec) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetKeyValuePairs

`func (o *GenericSecretUpdateSpec) GetKeyValuePairs() []KeyValuePair`

GetKeyValuePairs returns the KeyValuePairs field if non-nil, zero value otherwise.

### GetKeyValuePairsOk

`func (o *GenericSecretUpdateSpec) GetKeyValuePairsOk() (*[]KeyValuePair, bool)`

GetKeyValuePairsOk returns a tuple with the KeyValuePairs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeyValuePairs

`func (o *GenericSecretUpdateSpec) SetKeyValuePairs(v []KeyValuePair)`

SetKeyValuePairs sets KeyValuePairs field to given value.

### HasKeyValuePairs

`func (o *GenericSecretUpdateSpec) HasKeyValuePairs() bool`

HasKeyValuePairs returns a boolean if a field has been set.

### SetKeyValuePairsNil

`func (o *GenericSecretUpdateSpec) SetKeyValuePairsNil(b bool)`

 SetKeyValuePairsNil sets the value for KeyValuePairs to be an explicit nil

### UnsetKeyValuePairs
`func (o *GenericSecretUpdateSpec) UnsetKeyValuePairs()`

UnsetKeyValuePairs ensures that no value is present for KeyValuePairs, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


