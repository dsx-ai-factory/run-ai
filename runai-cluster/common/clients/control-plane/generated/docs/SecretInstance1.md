# SecretInstance1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Secret** | Pointer to **NullableString** | The name of the Secret resource. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workspace to which the Secret will be mapped to. (mandatory) | [optional] 

## Methods

### NewSecretInstance1

`func NewSecretInstance1() *SecretInstance1`

NewSecretInstance1 instantiates a new SecretInstance1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretInstance1WithDefaults

`func NewSecretInstance1WithDefaults() *SecretInstance1`

NewSecretInstance1WithDefaults instantiates a new SecretInstance1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *SecretInstance1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SecretInstance1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SecretInstance1) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *SecretInstance1) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *SecretInstance1) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *SecretInstance1) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetSecret

`func (o *SecretInstance1) GetSecret() string`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *SecretInstance1) GetSecretOk() (*string, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *SecretInstance1) SetSecret(v string)`

SetSecret sets Secret field to given value.

### HasSecret

`func (o *SecretInstance1) HasSecret() bool`

HasSecret returns a boolean if a field has been set.

### SetSecretNil

`func (o *SecretInstance1) SetSecretNil(b bool)`

 SetSecretNil sets the value for Secret to be an explicit nil

### UnsetSecret
`func (o *SecretInstance1) UnsetSecret()`

UnsetSecret ensures that no value is present for Secret, not even an explicit nil
### GetMountPath

`func (o *SecretInstance1) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *SecretInstance1) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *SecretInstance1) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *SecretInstance1) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *SecretInstance1) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *SecretInstance1) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


