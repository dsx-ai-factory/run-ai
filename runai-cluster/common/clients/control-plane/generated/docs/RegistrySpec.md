# RegistrySpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | **string** | The registry base url. | 
**CredentialKind** | [**CredentialKind**](CredentialKind.md) |  | 
**User** | **string** | The name of the user. | 
**Password** | **string** | The password. | 

## Methods

### NewRegistrySpec

`func NewRegistrySpec(url string, credentialKind CredentialKind, user string, password string, ) *RegistrySpec`

NewRegistrySpec instantiates a new RegistrySpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistrySpecWithDefaults

`func NewRegistrySpecWithDefaults() *RegistrySpec`

NewRegistrySpecWithDefaults instantiates a new RegistrySpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *RegistrySpec) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *RegistrySpec) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *RegistrySpec) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetCredentialKind

`func (o *RegistrySpec) GetCredentialKind() CredentialKind`

GetCredentialKind returns the CredentialKind field if non-nil, zero value otherwise.

### GetCredentialKindOk

`func (o *RegistrySpec) GetCredentialKindOk() (*CredentialKind, bool)`

GetCredentialKindOk returns a tuple with the CredentialKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialKind

`func (o *RegistrySpec) SetCredentialKind(v CredentialKind)`

SetCredentialKind sets CredentialKind field to given value.


### GetUser

`func (o *RegistrySpec) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *RegistrySpec) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *RegistrySpec) SetUser(v string)`

SetUser sets User field to given value.


### GetPassword

`func (o *RegistrySpec) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *RegistrySpec) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *RegistrySpec) SetPassword(v string)`

SetPassword sets Password field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


