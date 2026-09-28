# SupersetServingPortRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Container** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Protocol** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AuthorizationType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AuthorizedUsers** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**AuthorizedGroups** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**ClusterLocalAccessOnly** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ExposeExternally** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ExposedUrl** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewSupersetServingPortRules

`func NewSupersetServingPortRules() *SupersetServingPortRules`

NewSupersetServingPortRules instantiates a new SupersetServingPortRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetServingPortRulesWithDefaults

`func NewSupersetServingPortRulesWithDefaults() *SupersetServingPortRules`

NewSupersetServingPortRulesWithDefaults instantiates a new SupersetServingPortRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *SupersetServingPortRules) GetPort() IntegerRules`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *SupersetServingPortRules) GetPortOk() (*IntegerRules, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *SupersetServingPortRules) SetPort(v IntegerRules)`

SetPort sets Port field to given value.

### HasPort

`func (o *SupersetServingPortRules) HasPort() bool`

HasPort returns a boolean if a field has been set.

### SetPortNil

`func (o *SupersetServingPortRules) SetPortNil(b bool)`

 SetPortNil sets the value for Port to be an explicit nil

### UnsetPort
`func (o *SupersetServingPortRules) UnsetPort()`

UnsetPort ensures that no value is present for Port, not even an explicit nil
### GetContainer

`func (o *SupersetServingPortRules) GetContainer() IntegerRules`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *SupersetServingPortRules) GetContainerOk() (*IntegerRules, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *SupersetServingPortRules) SetContainer(v IntegerRules)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *SupersetServingPortRules) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *SupersetServingPortRules) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *SupersetServingPortRules) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetProtocol

`func (o *SupersetServingPortRules) GetProtocol() StringRules`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *SupersetServingPortRules) GetProtocolOk() (*StringRules, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *SupersetServingPortRules) SetProtocol(v StringRules)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *SupersetServingPortRules) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *SupersetServingPortRules) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *SupersetServingPortRules) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil
### GetAuthorizationType

`func (o *SupersetServingPortRules) GetAuthorizationType() StringRules`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *SupersetServingPortRules) GetAuthorizationTypeOk() (*StringRules, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *SupersetServingPortRules) SetAuthorizationType(v StringRules)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *SupersetServingPortRules) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *SupersetServingPortRules) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *SupersetServingPortRules) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *SupersetServingPortRules) GetAuthorizedUsers() ArrayRules`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *SupersetServingPortRules) GetAuthorizedUsersOk() (*ArrayRules, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *SupersetServingPortRules) SetAuthorizedUsers(v ArrayRules)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *SupersetServingPortRules) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *SupersetServingPortRules) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *SupersetServingPortRules) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *SupersetServingPortRules) GetAuthorizedGroups() ArrayRules`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *SupersetServingPortRules) GetAuthorizedGroupsOk() (*ArrayRules, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *SupersetServingPortRules) SetAuthorizedGroups(v ArrayRules)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *SupersetServingPortRules) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *SupersetServingPortRules) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *SupersetServingPortRules) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetClusterLocalAccessOnly

`func (o *SupersetServingPortRules) GetClusterLocalAccessOnly() BooleanRules`

GetClusterLocalAccessOnly returns the ClusterLocalAccessOnly field if non-nil, zero value otherwise.

### GetClusterLocalAccessOnlyOk

`func (o *SupersetServingPortRules) GetClusterLocalAccessOnlyOk() (*BooleanRules, bool)`

GetClusterLocalAccessOnlyOk returns a tuple with the ClusterLocalAccessOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterLocalAccessOnly

`func (o *SupersetServingPortRules) SetClusterLocalAccessOnly(v BooleanRules)`

SetClusterLocalAccessOnly sets ClusterLocalAccessOnly field to given value.

### HasClusterLocalAccessOnly

`func (o *SupersetServingPortRules) HasClusterLocalAccessOnly() bool`

HasClusterLocalAccessOnly returns a boolean if a field has been set.

### SetClusterLocalAccessOnlyNil

`func (o *SupersetServingPortRules) SetClusterLocalAccessOnlyNil(b bool)`

 SetClusterLocalAccessOnlyNil sets the value for ClusterLocalAccessOnly to be an explicit nil

### UnsetClusterLocalAccessOnly
`func (o *SupersetServingPortRules) UnsetClusterLocalAccessOnly()`

UnsetClusterLocalAccessOnly ensures that no value is present for ClusterLocalAccessOnly, not even an explicit nil
### GetExposeExternally

`func (o *SupersetServingPortRules) GetExposeExternally() BooleanRules`

GetExposeExternally returns the ExposeExternally field if non-nil, zero value otherwise.

### GetExposeExternallyOk

`func (o *SupersetServingPortRules) GetExposeExternallyOk() (*BooleanRules, bool)`

GetExposeExternallyOk returns a tuple with the ExposeExternally field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposeExternally

`func (o *SupersetServingPortRules) SetExposeExternally(v BooleanRules)`

SetExposeExternally sets ExposeExternally field to given value.

### HasExposeExternally

`func (o *SupersetServingPortRules) HasExposeExternally() bool`

HasExposeExternally returns a boolean if a field has been set.

### SetExposeExternallyNil

`func (o *SupersetServingPortRules) SetExposeExternallyNil(b bool)`

 SetExposeExternallyNil sets the value for ExposeExternally to be an explicit nil

### UnsetExposeExternally
`func (o *SupersetServingPortRules) UnsetExposeExternally()`

UnsetExposeExternally ensures that no value is present for ExposeExternally, not even an explicit nil
### GetExposedUrl

`func (o *SupersetServingPortRules) GetExposedUrl() StringRules`

GetExposedUrl returns the ExposedUrl field if non-nil, zero value otherwise.

### GetExposedUrlOk

`func (o *SupersetServingPortRules) GetExposedUrlOk() (*StringRules, bool)`

GetExposedUrlOk returns a tuple with the ExposedUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrl

`func (o *SupersetServingPortRules) SetExposedUrl(v StringRules)`

SetExposedUrl sets ExposedUrl field to given value.

### HasExposedUrl

`func (o *SupersetServingPortRules) HasExposedUrl() bool`

HasExposedUrl returns a boolean if a field has been set.

### SetExposedUrlNil

`func (o *SupersetServingPortRules) SetExposedUrlNil(b bool)`

 SetExposedUrlNil sets the value for ExposedUrl to be an explicit nil

### UnsetExposedUrl
`func (o *SupersetServingPortRules) UnsetExposedUrl()`

UnsetExposedUrl ensures that no value is present for ExposedUrl, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


