# CredentialsUpdateSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessKey** | Pointer to [**NullableAccessKeyCreationSpec**](AccessKeyCreationSpec.md) |  | [optional] 
**Password** | Pointer to [**NullablePasswordCreationSpec**](PasswordCreationSpec.md) |  | [optional] 
**GenericSecret** | Pointer to [**NullableGenericSecretCreationSpec**](GenericSecretCreationSpec.md) |  | [optional] 
**DockerRegistry** | Pointer to [**NullableDockerRegistryCreationSpec**](DockerRegistryCreationSpec.md) |  | [optional] 

## Methods

### NewCredentialsUpdateSpec

`func NewCredentialsUpdateSpec() *CredentialsUpdateSpec`

NewCredentialsUpdateSpec instantiates a new CredentialsUpdateSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCredentialsUpdateSpecWithDefaults

`func NewCredentialsUpdateSpecWithDefaults() *CredentialsUpdateSpec`

NewCredentialsUpdateSpecWithDefaults instantiates a new CredentialsUpdateSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessKey

`func (o *CredentialsUpdateSpec) GetAccessKey() AccessKeyCreationSpec`

GetAccessKey returns the AccessKey field if non-nil, zero value otherwise.

### GetAccessKeyOk

`func (o *CredentialsUpdateSpec) GetAccessKeyOk() (*AccessKeyCreationSpec, bool)`

GetAccessKeyOk returns a tuple with the AccessKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKey

`func (o *CredentialsUpdateSpec) SetAccessKey(v AccessKeyCreationSpec)`

SetAccessKey sets AccessKey field to given value.

### HasAccessKey

`func (o *CredentialsUpdateSpec) HasAccessKey() bool`

HasAccessKey returns a boolean if a field has been set.

### SetAccessKeyNil

`func (o *CredentialsUpdateSpec) SetAccessKeyNil(b bool)`

 SetAccessKeyNil sets the value for AccessKey to be an explicit nil

### UnsetAccessKey
`func (o *CredentialsUpdateSpec) UnsetAccessKey()`

UnsetAccessKey ensures that no value is present for AccessKey, not even an explicit nil
### GetPassword

`func (o *CredentialsUpdateSpec) GetPassword() PasswordCreationSpec`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *CredentialsUpdateSpec) GetPasswordOk() (*PasswordCreationSpec, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *CredentialsUpdateSpec) SetPassword(v PasswordCreationSpec)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *CredentialsUpdateSpec) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### SetPasswordNil

`func (o *CredentialsUpdateSpec) SetPasswordNil(b bool)`

 SetPasswordNil sets the value for Password to be an explicit nil

### UnsetPassword
`func (o *CredentialsUpdateSpec) UnsetPassword()`

UnsetPassword ensures that no value is present for Password, not even an explicit nil
### GetGenericSecret

`func (o *CredentialsUpdateSpec) GetGenericSecret() GenericSecretCreationSpec`

GetGenericSecret returns the GenericSecret field if non-nil, zero value otherwise.

### GetGenericSecretOk

`func (o *CredentialsUpdateSpec) GetGenericSecretOk() (*GenericSecretCreationSpec, bool)`

GetGenericSecretOk returns a tuple with the GenericSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGenericSecret

`func (o *CredentialsUpdateSpec) SetGenericSecret(v GenericSecretCreationSpec)`

SetGenericSecret sets GenericSecret field to given value.

### HasGenericSecret

`func (o *CredentialsUpdateSpec) HasGenericSecret() bool`

HasGenericSecret returns a boolean if a field has been set.

### SetGenericSecretNil

`func (o *CredentialsUpdateSpec) SetGenericSecretNil(b bool)`

 SetGenericSecretNil sets the value for GenericSecret to be an explicit nil

### UnsetGenericSecret
`func (o *CredentialsUpdateSpec) UnsetGenericSecret()`

UnsetGenericSecret ensures that no value is present for GenericSecret, not even an explicit nil
### GetDockerRegistry

`func (o *CredentialsUpdateSpec) GetDockerRegistry() DockerRegistryCreationSpec`

GetDockerRegistry returns the DockerRegistry field if non-nil, zero value otherwise.

### GetDockerRegistryOk

`func (o *CredentialsUpdateSpec) GetDockerRegistryOk() (*DockerRegistryCreationSpec, bool)`

GetDockerRegistryOk returns a tuple with the DockerRegistry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDockerRegistry

`func (o *CredentialsUpdateSpec) SetDockerRegistry(v DockerRegistryCreationSpec)`

SetDockerRegistry sets DockerRegistry field to given value.

### HasDockerRegistry

`func (o *CredentialsUpdateSpec) HasDockerRegistry() bool`

HasDockerRegistry returns a boolean if a field has been set.

### SetDockerRegistryNil

`func (o *CredentialsUpdateSpec) SetDockerRegistryNil(b bool)`

 SetDockerRegistryNil sets the value for DockerRegistry to be an explicit nil

### UnsetDockerRegistry
`func (o *CredentialsUpdateSpec) UnsetDockerRegistry()`

UnsetDockerRegistry ensures that no value is present for DockerRegistry, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


