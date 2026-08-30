# ServingPortAccess

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AuthorizationType** | Pointer to [**NullableServingPortAccessAuthorizationTypeEnum**](ServingPortAccessAuthorizationTypeEnum.md) |  | [optional] 
**AuthorizedUsers** | Pointer to **[]string** | List of users or service accounts that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**AuthorizedGroups** | Pointer to **[]string** | List of groups that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**ClusterLocalAccessOnly** | Pointer to **NullableBool** | Configure the serving port URL to be available only on the cluster-local network, and not externally. Defaults to false | [optional] 

## Methods

### NewServingPortAccess

`func NewServingPortAccess() *ServingPortAccess`

NewServingPortAccess instantiates a new ServingPortAccess object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortAccessWithDefaults

`func NewServingPortAccessWithDefaults() *ServingPortAccess`

NewServingPortAccessWithDefaults instantiates a new ServingPortAccess object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthorizationType

`func (o *ServingPortAccess) GetAuthorizationType() ServingPortAccessAuthorizationTypeEnum`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *ServingPortAccess) GetAuthorizationTypeOk() (*ServingPortAccessAuthorizationTypeEnum, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *ServingPortAccess) SetAuthorizationType(v ServingPortAccessAuthorizationTypeEnum)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *ServingPortAccess) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *ServingPortAccess) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *ServingPortAccess) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *ServingPortAccess) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *ServingPortAccess) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *ServingPortAccess) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *ServingPortAccess) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *ServingPortAccess) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *ServingPortAccess) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *ServingPortAccess) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *ServingPortAccess) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *ServingPortAccess) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *ServingPortAccess) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *ServingPortAccess) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *ServingPortAccess) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetClusterLocalAccessOnly

`func (o *ServingPortAccess) GetClusterLocalAccessOnly() bool`

GetClusterLocalAccessOnly returns the ClusterLocalAccessOnly field if non-nil, zero value otherwise.

### GetClusterLocalAccessOnlyOk

`func (o *ServingPortAccess) GetClusterLocalAccessOnlyOk() (*bool, bool)`

GetClusterLocalAccessOnlyOk returns a tuple with the ClusterLocalAccessOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterLocalAccessOnly

`func (o *ServingPortAccess) SetClusterLocalAccessOnly(v bool)`

SetClusterLocalAccessOnly sets ClusterLocalAccessOnly field to given value.

### HasClusterLocalAccessOnly

`func (o *ServingPortAccess) HasClusterLocalAccessOnly() bool`

HasClusterLocalAccessOnly returns a boolean if a field has been set.

### SetClusterLocalAccessOnlyNil

`func (o *ServingPortAccess) SetClusterLocalAccessOnlyNil(b bool)`

 SetClusterLocalAccessOnlyNil sets the value for ClusterLocalAccessOnly to be an explicit nil

### UnsetClusterLocalAccessOnly
`func (o *ServingPortAccess) UnsetClusterLocalAccessOnly()`

UnsetClusterLocalAccessOnly ensures that no value is present for ClusterLocalAccessOnly, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


