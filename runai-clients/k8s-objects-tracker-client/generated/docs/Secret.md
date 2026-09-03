# Secret

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | name of the secret. | 
**Type** | [**SecretType**](SecretType.md) |  | 
**Origin** | Pointer to [**NullableOrigin**](Origin.md) |  | [optional] 
**Private** | Pointer to **NullableBool** | Whether the secret is private (user-credential). | [optional] [default to false]
**Issues** | Pointer to [**[]Issue**](Issue.md) | list of issues that was found while syncing the resource to the cluster | [optional] 

## Methods

### NewSecret

`func NewSecret(name string, type_ SecretType, ) *Secret`

NewSecret instantiates a new Secret object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretWithDefaults

`func NewSecretWithDefaults() *Secret`

NewSecretWithDefaults instantiates a new Secret object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Secret) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Secret) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Secret) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *Secret) GetType() SecretType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Secret) GetTypeOk() (*SecretType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Secret) SetType(v SecretType)`

SetType sets Type field to given value.


### GetOrigin

`func (o *Secret) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *Secret) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *Secret) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.

### HasOrigin

`func (o *Secret) HasOrigin() bool`

HasOrigin returns a boolean if a field has been set.

### SetOriginNil

`func (o *Secret) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *Secret) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetPrivate

`func (o *Secret) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *Secret) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *Secret) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *Secret) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### SetPrivateNil

`func (o *Secret) SetPrivateNil(b bool)`

 SetPrivateNil sets the value for Private to be an explicit nil

### UnsetPrivate
`func (o *Secret) UnsetPrivate()`

UnsetPrivate ensures that no value is present for Private, not even an explicit nil
### GetIssues

`func (o *Secret) GetIssues() []Issue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *Secret) GetIssuesOk() (*[]Issue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *Secret) SetIssues(v []Issue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *Secret) HasIssues() bool`

HasIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


