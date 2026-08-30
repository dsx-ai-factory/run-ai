# CredentialsListResponseAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessKey** | Pointer to [**NullableAccessKeySpec**](AccessKeySpec.md) |  | [optional] 
**Password** | Pointer to [**NullablePasswordSpec**](PasswordSpec.md) |  | [optional] 
**DockerRegistry** | Pointer to [**NullableDockerRegistrySpec**](DockerRegistrySpec.md) |  | [optional] 
**GenericSecret** | Pointer to [**NullableGenericSecretSpec**](GenericSecretSpec.md) |  | [optional] 
**NgcApiKey** | Pointer to [**NullableNgcApiKeySpec**](NgcApiKeySpec.md) |  | [optional] 

## Methods

### NewCredentialsListResponseAssetSpec

`func NewCredentialsListResponseAssetSpec() *CredentialsListResponseAssetSpec`

NewCredentialsListResponseAssetSpec instantiates a new CredentialsListResponseAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCredentialsListResponseAssetSpecWithDefaults

`func NewCredentialsListResponseAssetSpecWithDefaults() *CredentialsListResponseAssetSpec`

NewCredentialsListResponseAssetSpecWithDefaults instantiates a new CredentialsListResponseAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessKey

`func (o *CredentialsListResponseAssetSpec) GetAccessKey() AccessKeySpec`

GetAccessKey returns the AccessKey field if non-nil, zero value otherwise.

### GetAccessKeyOk

`func (o *CredentialsListResponseAssetSpec) GetAccessKeyOk() (*AccessKeySpec, bool)`

GetAccessKeyOk returns a tuple with the AccessKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKey

`func (o *CredentialsListResponseAssetSpec) SetAccessKey(v AccessKeySpec)`

SetAccessKey sets AccessKey field to given value.

### HasAccessKey

`func (o *CredentialsListResponseAssetSpec) HasAccessKey() bool`

HasAccessKey returns a boolean if a field has been set.

### SetAccessKeyNil

`func (o *CredentialsListResponseAssetSpec) SetAccessKeyNil(b bool)`

 SetAccessKeyNil sets the value for AccessKey to be an explicit nil

### UnsetAccessKey
`func (o *CredentialsListResponseAssetSpec) UnsetAccessKey()`

UnsetAccessKey ensures that no value is present for AccessKey, not even an explicit nil
### GetPassword

`func (o *CredentialsListResponseAssetSpec) GetPassword() PasswordSpec`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *CredentialsListResponseAssetSpec) GetPasswordOk() (*PasswordSpec, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *CredentialsListResponseAssetSpec) SetPassword(v PasswordSpec)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *CredentialsListResponseAssetSpec) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### SetPasswordNil

`func (o *CredentialsListResponseAssetSpec) SetPasswordNil(b bool)`

 SetPasswordNil sets the value for Password to be an explicit nil

### UnsetPassword
`func (o *CredentialsListResponseAssetSpec) UnsetPassword()`

UnsetPassword ensures that no value is present for Password, not even an explicit nil
### GetDockerRegistry

`func (o *CredentialsListResponseAssetSpec) GetDockerRegistry() DockerRegistrySpec`

GetDockerRegistry returns the DockerRegistry field if non-nil, zero value otherwise.

### GetDockerRegistryOk

`func (o *CredentialsListResponseAssetSpec) GetDockerRegistryOk() (*DockerRegistrySpec, bool)`

GetDockerRegistryOk returns a tuple with the DockerRegistry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDockerRegistry

`func (o *CredentialsListResponseAssetSpec) SetDockerRegistry(v DockerRegistrySpec)`

SetDockerRegistry sets DockerRegistry field to given value.

### HasDockerRegistry

`func (o *CredentialsListResponseAssetSpec) HasDockerRegistry() bool`

HasDockerRegistry returns a boolean if a field has been set.

### SetDockerRegistryNil

`func (o *CredentialsListResponseAssetSpec) SetDockerRegistryNil(b bool)`

 SetDockerRegistryNil sets the value for DockerRegistry to be an explicit nil

### UnsetDockerRegistry
`func (o *CredentialsListResponseAssetSpec) UnsetDockerRegistry()`

UnsetDockerRegistry ensures that no value is present for DockerRegistry, not even an explicit nil
### GetGenericSecret

`func (o *CredentialsListResponseAssetSpec) GetGenericSecret() GenericSecretSpec`

GetGenericSecret returns the GenericSecret field if non-nil, zero value otherwise.

### GetGenericSecretOk

`func (o *CredentialsListResponseAssetSpec) GetGenericSecretOk() (*GenericSecretSpec, bool)`

GetGenericSecretOk returns a tuple with the GenericSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGenericSecret

`func (o *CredentialsListResponseAssetSpec) SetGenericSecret(v GenericSecretSpec)`

SetGenericSecret sets GenericSecret field to given value.

### HasGenericSecret

`func (o *CredentialsListResponseAssetSpec) HasGenericSecret() bool`

HasGenericSecret returns a boolean if a field has been set.

### SetGenericSecretNil

`func (o *CredentialsListResponseAssetSpec) SetGenericSecretNil(b bool)`

 SetGenericSecretNil sets the value for GenericSecret to be an explicit nil

### UnsetGenericSecret
`func (o *CredentialsListResponseAssetSpec) UnsetGenericSecret()`

UnsetGenericSecret ensures that no value is present for GenericSecret, not even an explicit nil
### GetNgcApiKey

`func (o *CredentialsListResponseAssetSpec) GetNgcApiKey() NgcApiKeySpec`

GetNgcApiKey returns the NgcApiKey field if non-nil, zero value otherwise.

### GetNgcApiKeyOk

`func (o *CredentialsListResponseAssetSpec) GetNgcApiKeyOk() (*NgcApiKeySpec, bool)`

GetNgcApiKeyOk returns a tuple with the NgcApiKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNgcApiKey

`func (o *CredentialsListResponseAssetSpec) SetNgcApiKey(v NgcApiKeySpec)`

SetNgcApiKey sets NgcApiKey field to given value.

### HasNgcApiKey

`func (o *CredentialsListResponseAssetSpec) HasNgcApiKey() bool`

HasNgcApiKey returns a boolean if a field has been set.

### SetNgcApiKeyNil

`func (o *CredentialsListResponseAssetSpec) SetNgcApiKeyNil(b bool)`

 SetNgcApiKeyNil sets the value for NgcApiKey to be an explicit nil

### UnsetNgcApiKey
`func (o *CredentialsListResponseAssetSpec) UnsetNgcApiKey()`

UnsetNgcApiKey ensures that no value is present for NgcApiKey, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


