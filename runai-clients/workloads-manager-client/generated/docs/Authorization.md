# Authorization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AuthorizationType** | Pointer to [**NullableAuthorizationType**](AuthorizationType.md) |  | [optional] 
**AuthorizedUsers** | Pointer to **[]string** | List of users or service accounts that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**AuthorizedGroups** | Pointer to **[]string** | List of groups that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 

## Methods

### NewAuthorization

`func NewAuthorization() *Authorization`

NewAuthorization instantiates a new Authorization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthorizationWithDefaults

`func NewAuthorizationWithDefaults() *Authorization`

NewAuthorizationWithDefaults instantiates a new Authorization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthorizationType

`func (o *Authorization) GetAuthorizationType() AuthorizationType`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *Authorization) GetAuthorizationTypeOk() (*AuthorizationType, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *Authorization) SetAuthorizationType(v AuthorizationType)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *Authorization) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *Authorization) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *Authorization) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *Authorization) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *Authorization) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *Authorization) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *Authorization) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *Authorization) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *Authorization) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *Authorization) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *Authorization) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *Authorization) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *Authorization) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *Authorization) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *Authorization) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


