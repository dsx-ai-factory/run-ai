# ServingPortRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Container** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Protocol** | Pointer to [**NullableServingPortProtocolRules**](ServingPortProtocolRules.md) |  | [optional] 
**AuthorizationType** | Pointer to [**NullableServingPortAuthorizationTypeRules**](ServingPortAuthorizationTypeRules.md) |  | [optional] 
**AuthorizedUsers** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**AuthorizedGroups** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**ClusterLocalAccessOnly** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 

## Methods

### NewServingPortRules

`func NewServingPortRules() *ServingPortRules`

NewServingPortRules instantiates a new ServingPortRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortRulesWithDefaults

`func NewServingPortRulesWithDefaults() *ServingPortRules`

NewServingPortRulesWithDefaults instantiates a new ServingPortRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainer

`func (o *ServingPortRules) GetContainer() IntegerRules`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *ServingPortRules) GetContainerOk() (*IntegerRules, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *ServingPortRules) SetContainer(v IntegerRules)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *ServingPortRules) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *ServingPortRules) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *ServingPortRules) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetProtocol

`func (o *ServingPortRules) GetProtocol() ServingPortProtocolRules`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *ServingPortRules) GetProtocolOk() (*ServingPortProtocolRules, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *ServingPortRules) SetProtocol(v ServingPortProtocolRules)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *ServingPortRules) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *ServingPortRules) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *ServingPortRules) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil
### GetAuthorizationType

`func (o *ServingPortRules) GetAuthorizationType() ServingPortAuthorizationTypeRules`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *ServingPortRules) GetAuthorizationTypeOk() (*ServingPortAuthorizationTypeRules, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *ServingPortRules) SetAuthorizationType(v ServingPortAuthorizationTypeRules)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *ServingPortRules) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *ServingPortRules) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *ServingPortRules) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *ServingPortRules) GetAuthorizedUsers() ArrayRules`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *ServingPortRules) GetAuthorizedUsersOk() (*ArrayRules, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *ServingPortRules) SetAuthorizedUsers(v ArrayRules)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *ServingPortRules) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *ServingPortRules) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *ServingPortRules) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *ServingPortRules) GetAuthorizedGroups() ArrayRules`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *ServingPortRules) GetAuthorizedGroupsOk() (*ArrayRules, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *ServingPortRules) SetAuthorizedGroups(v ArrayRules)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *ServingPortRules) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *ServingPortRules) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *ServingPortRules) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetClusterLocalAccessOnly

`func (o *ServingPortRules) GetClusterLocalAccessOnly() BooleanRules`

GetClusterLocalAccessOnly returns the ClusterLocalAccessOnly field if non-nil, zero value otherwise.

### GetClusterLocalAccessOnlyOk

`func (o *ServingPortRules) GetClusterLocalAccessOnlyOk() (*BooleanRules, bool)`

GetClusterLocalAccessOnlyOk returns a tuple with the ClusterLocalAccessOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterLocalAccessOnly

`func (o *ServingPortRules) SetClusterLocalAccessOnly(v BooleanRules)`

SetClusterLocalAccessOnly sets ClusterLocalAccessOnly field to given value.

### HasClusterLocalAccessOnly

`func (o *ServingPortRules) HasClusterLocalAccessOnly() bool`

HasClusterLocalAccessOnly returns a boolean if a field has been set.

### SetClusterLocalAccessOnlyNil

`func (o *ServingPortRules) SetClusterLocalAccessOnlyNil(b bool)`

 SetClusterLocalAccessOnlyNil sets the value for ClusterLocalAccessOnly to be an explicit nil

### UnsetClusterLocalAccessOnly
`func (o *ServingPortRules) UnsetClusterLocalAccessOnly()`

UnsetClusterLocalAccessOnly ensures that no value is present for ClusterLocalAccessOnly, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


