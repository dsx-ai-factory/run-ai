# DockerRegistryCreationSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | optional name of existing secret in the cluster from which the docker registry details should be taken. If omitted, you will have to provide those credentials and url in relevant fields. The provided credentials are encrypted into the control plane database and cloned to a kubernetes secret in the cluster, along with the url. | [optional] 
**User** | Pointer to **NullableString** | The name of the user, required only when not using existing secret. | [optional] 
**Password** | Pointer to **NullableString** | The password, required only when not using existing secret. | [optional] 
**Url** | Pointer to **NullableString** | The url, required only when not using existing secret. | [optional] 

## Methods

### NewDockerRegistryCreationSpec

`func NewDockerRegistryCreationSpec() *DockerRegistryCreationSpec`

NewDockerRegistryCreationSpec instantiates a new DockerRegistryCreationSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDockerRegistryCreationSpecWithDefaults

`func NewDockerRegistryCreationSpecWithDefaults() *DockerRegistryCreationSpec`

NewDockerRegistryCreationSpecWithDefaults instantiates a new DockerRegistryCreationSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *DockerRegistryCreationSpec) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *DockerRegistryCreationSpec) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *DockerRegistryCreationSpec) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *DockerRegistryCreationSpec) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *DockerRegistryCreationSpec) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *DockerRegistryCreationSpec) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetUser

`func (o *DockerRegistryCreationSpec) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *DockerRegistryCreationSpec) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *DockerRegistryCreationSpec) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *DockerRegistryCreationSpec) HasUser() bool`

HasUser returns a boolean if a field has been set.

### SetUserNil

`func (o *DockerRegistryCreationSpec) SetUserNil(b bool)`

 SetUserNil sets the value for User to be an explicit nil

### UnsetUser
`func (o *DockerRegistryCreationSpec) UnsetUser()`

UnsetUser ensures that no value is present for User, not even an explicit nil
### GetPassword

`func (o *DockerRegistryCreationSpec) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *DockerRegistryCreationSpec) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *DockerRegistryCreationSpec) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *DockerRegistryCreationSpec) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### SetPasswordNil

`func (o *DockerRegistryCreationSpec) SetPasswordNil(b bool)`

 SetPasswordNil sets the value for Password to be an explicit nil

### UnsetPassword
`func (o *DockerRegistryCreationSpec) UnsetPassword()`

UnsetPassword ensures that no value is present for Password, not even an explicit nil
### GetUrl

`func (o *DockerRegistryCreationSpec) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *DockerRegistryCreationSpec) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *DockerRegistryCreationSpec) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *DockerRegistryCreationSpec) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *DockerRegistryCreationSpec) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *DockerRegistryCreationSpec) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


