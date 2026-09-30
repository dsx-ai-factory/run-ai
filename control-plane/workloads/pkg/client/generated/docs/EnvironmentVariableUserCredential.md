# EnvironmentVariableUserCredential

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the user credential.  (mandatory) | [optional] 
**Key** | Pointer to **NullableString** | The key in the user credential resource. (mandatory) | [optional] 

## Methods

### NewEnvironmentVariableUserCredential

`func NewEnvironmentVariableUserCredential() *EnvironmentVariableUserCredential`

NewEnvironmentVariableUserCredential instantiates a new EnvironmentVariableUserCredential object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentVariableUserCredentialWithDefaults

`func NewEnvironmentVariableUserCredentialWithDefaults() *EnvironmentVariableUserCredential`

NewEnvironmentVariableUserCredentialWithDefaults instantiates a new EnvironmentVariableUserCredential object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *EnvironmentVariableUserCredential) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EnvironmentVariableUserCredential) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EnvironmentVariableUserCredential) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *EnvironmentVariableUserCredential) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *EnvironmentVariableUserCredential) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *EnvironmentVariableUserCredential) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetKey

`func (o *EnvironmentVariableUserCredential) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *EnvironmentVariableUserCredential) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *EnvironmentVariableUserCredential) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *EnvironmentVariableUserCredential) HasKey() bool`

HasKey returns a boolean if a field has been set.

### SetKeyNil

`func (o *EnvironmentVariableUserCredential) SetKeyNil(b bool)`

 SetKeyNil sets the value for Key to be an explicit nil

### UnsetKey
`func (o *EnvironmentVariableUserCredential) UnsetKey()`

UnsetKey ensures that no value is present for Key, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


