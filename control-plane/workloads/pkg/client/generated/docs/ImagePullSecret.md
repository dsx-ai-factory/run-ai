# ImagePullSecret

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | The name of the Kubernetes secret containing the image pull credentials. | [optional] 
**UserCredential** | Pointer to **NullableBool** | Indicates whether the secret is a user credential. Set to true if the secret was created by the user and is only accessible by them. | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the secret is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewImagePullSecret

`func NewImagePullSecret() *ImagePullSecret`

NewImagePullSecret instantiates a new ImagePullSecret object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImagePullSecretWithDefaults

`func NewImagePullSecretWithDefaults() *ImagePullSecret`

NewImagePullSecretWithDefaults instantiates a new ImagePullSecret object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ImagePullSecret) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ImagePullSecret) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ImagePullSecret) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ImagePullSecret) HasName() bool`

HasName returns a boolean if a field has been set.

### GetUserCredential

`func (o *ImagePullSecret) GetUserCredential() bool`

GetUserCredential returns the UserCredential field if non-nil, zero value otherwise.

### GetUserCredentialOk

`func (o *ImagePullSecret) GetUserCredentialOk() (*bool, bool)`

GetUserCredentialOk returns a tuple with the UserCredential field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCredential

`func (o *ImagePullSecret) SetUserCredential(v bool)`

SetUserCredential sets UserCredential field to given value.

### HasUserCredential

`func (o *ImagePullSecret) HasUserCredential() bool`

HasUserCredential returns a boolean if a field has been set.

### SetUserCredentialNil

`func (o *ImagePullSecret) SetUserCredentialNil(b bool)`

 SetUserCredentialNil sets the value for UserCredential to be an explicit nil

### UnsetUserCredential
`func (o *ImagePullSecret) UnsetUserCredential()`

UnsetUserCredential ensures that no value is present for UserCredential, not even an explicit nil
### GetExclude

`func (o *ImagePullSecret) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *ImagePullSecret) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *ImagePullSecret) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *ImagePullSecret) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *ImagePullSecret) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *ImagePullSecret) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


