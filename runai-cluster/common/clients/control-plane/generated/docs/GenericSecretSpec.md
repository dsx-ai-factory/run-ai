# GenericSecretSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | name of existing secret in the cluster from which the general key-value values should be taken. | [optional] 
**KeyValuePairs** | Pointer to [**[]KeyValuePair**](KeyValuePair.md) | Set of key-value pairs to store in the new credential. | [optional] 
**KeyId** | **int32** | id of the encryption key which has been used to encrypt an asset&#39;s data. | 

## Methods

### NewGenericSecretSpec

`func NewGenericSecretSpec(keyId int32, ) *GenericSecretSpec`

NewGenericSecretSpec instantiates a new GenericSecretSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGenericSecretSpecWithDefaults

`func NewGenericSecretSpecWithDefaults() *GenericSecretSpec`

NewGenericSecretSpecWithDefaults instantiates a new GenericSecretSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *GenericSecretSpec) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *GenericSecretSpec) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *GenericSecretSpec) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *GenericSecretSpec) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *GenericSecretSpec) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *GenericSecretSpec) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetKeyValuePairs

`func (o *GenericSecretSpec) GetKeyValuePairs() []KeyValuePair`

GetKeyValuePairs returns the KeyValuePairs field if non-nil, zero value otherwise.

### GetKeyValuePairsOk

`func (o *GenericSecretSpec) GetKeyValuePairsOk() (*[]KeyValuePair, bool)`

GetKeyValuePairsOk returns a tuple with the KeyValuePairs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeyValuePairs

`func (o *GenericSecretSpec) SetKeyValuePairs(v []KeyValuePair)`

SetKeyValuePairs sets KeyValuePairs field to given value.

### HasKeyValuePairs

`func (o *GenericSecretSpec) HasKeyValuePairs() bool`

HasKeyValuePairs returns a boolean if a field has been set.

### SetKeyValuePairsNil

`func (o *GenericSecretSpec) SetKeyValuePairsNil(b bool)`

 SetKeyValuePairsNil sets the value for KeyValuePairs to be an explicit nil

### UnsetKeyValuePairs
`func (o *GenericSecretSpec) UnsetKeyValuePairs()`

UnsetKeyValuePairs ensures that no value is present for KeyValuePairs, not even an explicit nil
### GetKeyId

`func (o *GenericSecretSpec) GetKeyId() int32`

GetKeyId returns the KeyId field if non-nil, zero value otherwise.

### GetKeyIdOk

`func (o *GenericSecretSpec) GetKeyIdOk() (*int32, bool)`

GetKeyIdOk returns a tuple with the KeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeyId

`func (o *GenericSecretSpec) SetKeyId(v int32)`

SetKeyId sets KeyId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


