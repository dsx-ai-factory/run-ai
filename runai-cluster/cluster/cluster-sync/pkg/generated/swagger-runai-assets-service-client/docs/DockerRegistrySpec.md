# DockerRegistrySpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | optional name of existing secret in the cluster from which the docker registry details should be taken. If omitted, you will have to provide those credentials and url in relevant fields. The provided credentials are encrypted into the control plane database and cloned to a kubernetes secret in the cluster, along with the url. | [optional] 
**User** | Pointer to **NullableString** | The name of the user, required only when not using existing secret. | [optional] 
**Password** | Pointer to **NullableString** | The password, required only when not using existing secret. | [optional] 
**Url** | Pointer to **NullableString** | The url, required only when not using existing secret. | [optional] 
**KeyId** | **int32** | id of the encryption key which has been used to encrypt an asset&#39;s data. | 

## Methods

### NewDockerRegistrySpec

`func NewDockerRegistrySpec(keyId int32, ) *DockerRegistrySpec`

NewDockerRegistrySpec instantiates a new DockerRegistrySpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDockerRegistrySpecWithDefaults

`func NewDockerRegistrySpecWithDefaults() *DockerRegistrySpec`

NewDockerRegistrySpecWithDefaults instantiates a new DockerRegistrySpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *DockerRegistrySpec) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *DockerRegistrySpec) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *DockerRegistrySpec) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *DockerRegistrySpec) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *DockerRegistrySpec) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *DockerRegistrySpec) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetUser

`func (o *DockerRegistrySpec) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *DockerRegistrySpec) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *DockerRegistrySpec) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *DockerRegistrySpec) HasUser() bool`

HasUser returns a boolean if a field has been set.

### SetUserNil

`func (o *DockerRegistrySpec) SetUserNil(b bool)`

 SetUserNil sets the value for User to be an explicit nil

### UnsetUser
`func (o *DockerRegistrySpec) UnsetUser()`

UnsetUser ensures that no value is present for User, not even an explicit nil
### GetPassword

`func (o *DockerRegistrySpec) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *DockerRegistrySpec) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *DockerRegistrySpec) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *DockerRegistrySpec) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### SetPasswordNil

`func (o *DockerRegistrySpec) SetPasswordNil(b bool)`

 SetPasswordNil sets the value for Password to be an explicit nil

### UnsetPassword
`func (o *DockerRegistrySpec) UnsetPassword()`

UnsetPassword ensures that no value is present for Password, not even an explicit nil
### GetUrl

`func (o *DockerRegistrySpec) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *DockerRegistrySpec) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *DockerRegistrySpec) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *DockerRegistrySpec) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *DockerRegistrySpec) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *DockerRegistrySpec) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil
### GetKeyId

`func (o *DockerRegistrySpec) GetKeyId() int32`

GetKeyId returns the KeyId field if non-nil, zero value otherwise.

### GetKeyIdOk

`func (o *DockerRegistrySpec) GetKeyIdOk() (*int32, bool)`

GetKeyIdOk returns a tuple with the KeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeyId

`func (o *DockerRegistrySpec) SetKeyId(v int32)`

SetKeyId sets KeyId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


