# ServingPort

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Container** | Pointer to **NullableInt32** | The port that the container running the inference service exposes (mandatory). | [optional] 
**Protocol** | Pointer to [**NullableServingPortProtocol**](ServingPortProtocol.md) |  | [optional] 
**AuthorizationType** | Pointer to [**NullableServingPortAccessAuthorizationTypeEnum**](ServingPortAccessAuthorizationTypeEnum.md) |  | [optional] 
**AuthorizedUsers** | Pointer to **[]string** | List of users or service accounts that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**AuthorizedGroups** | Pointer to **[]string** | List of groups that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**ClusterLocalAccessOnly** | Pointer to **NullableBool** | Configure the serving port URL to be available only on the cluster-local network, and not externally. Defaults to false | [optional] 

## Methods

### NewServingPort

`func NewServingPort() *ServingPort`

NewServingPort instantiates a new ServingPort object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortWithDefaults

`func NewServingPortWithDefaults() *ServingPort`

NewServingPortWithDefaults instantiates a new ServingPort object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainer

`func (o *ServingPort) GetContainer() int32`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *ServingPort) GetContainerOk() (*int32, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *ServingPort) SetContainer(v int32)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *ServingPort) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *ServingPort) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *ServingPort) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetProtocol

`func (o *ServingPort) GetProtocol() ServingPortProtocol`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *ServingPort) GetProtocolOk() (*ServingPortProtocol, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *ServingPort) SetProtocol(v ServingPortProtocol)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *ServingPort) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *ServingPort) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *ServingPort) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil
### GetAuthorizationType

`func (o *ServingPort) GetAuthorizationType() ServingPortAccessAuthorizationTypeEnum`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *ServingPort) GetAuthorizationTypeOk() (*ServingPortAccessAuthorizationTypeEnum, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *ServingPort) SetAuthorizationType(v ServingPortAccessAuthorizationTypeEnum)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *ServingPort) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *ServingPort) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *ServingPort) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *ServingPort) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *ServingPort) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *ServingPort) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *ServingPort) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *ServingPort) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *ServingPort) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *ServingPort) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *ServingPort) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *ServingPort) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *ServingPort) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *ServingPort) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *ServingPort) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetClusterLocalAccessOnly

`func (o *ServingPort) GetClusterLocalAccessOnly() bool`

GetClusterLocalAccessOnly returns the ClusterLocalAccessOnly field if non-nil, zero value otherwise.

### GetClusterLocalAccessOnlyOk

`func (o *ServingPort) GetClusterLocalAccessOnlyOk() (*bool, bool)`

GetClusterLocalAccessOnlyOk returns a tuple with the ClusterLocalAccessOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterLocalAccessOnly

`func (o *ServingPort) SetClusterLocalAccessOnly(v bool)`

SetClusterLocalAccessOnly sets ClusterLocalAccessOnly field to given value.

### HasClusterLocalAccessOnly

`func (o *ServingPort) HasClusterLocalAccessOnly() bool`

HasClusterLocalAccessOnly returns a boolean if a field has been set.

### SetClusterLocalAccessOnlyNil

`func (o *ServingPort) SetClusterLocalAccessOnlyNil(b bool)`

 SetClusterLocalAccessOnlyNil sets the value for ClusterLocalAccessOnly to be an explicit nil

### UnsetClusterLocalAccessOnly
`func (o *ServingPort) UnsetClusterLocalAccessOnly()`

UnsetClusterLocalAccessOnly ensures that no value is present for ClusterLocalAccessOnly, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


