# SupersetServingPort

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | Pointer to **NullableInt32** |  | [optional] 
**Container** | Pointer to **NullableInt32** |  | [optional] 
**Protocol** | Pointer to **NullableString** |  | [optional] 
**AuthorizationType** | Pointer to **NullableString** |  | [optional] 
**AuthorizedUsers** | Pointer to **[]string** |  | [optional] 
**AuthorizedGroups** | Pointer to **[]string** |  | [optional] 
**ClusterLocalAccessOnly** | Pointer to **NullableBool** |  | [optional] 
**ExposeExternally** | Pointer to **NullableBool** |  | [optional] 
**ExposedUrl** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewSupersetServingPort

`func NewSupersetServingPort() *SupersetServingPort`

NewSupersetServingPort instantiates a new SupersetServingPort object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetServingPortWithDefaults

`func NewSupersetServingPortWithDefaults() *SupersetServingPort`

NewSupersetServingPortWithDefaults instantiates a new SupersetServingPort object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *SupersetServingPort) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *SupersetServingPort) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *SupersetServingPort) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *SupersetServingPort) HasPort() bool`

HasPort returns a boolean if a field has been set.

### SetPortNil

`func (o *SupersetServingPort) SetPortNil(b bool)`

 SetPortNil sets the value for Port to be an explicit nil

### UnsetPort
`func (o *SupersetServingPort) UnsetPort()`

UnsetPort ensures that no value is present for Port, not even an explicit nil
### GetContainer

`func (o *SupersetServingPort) GetContainer() int32`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *SupersetServingPort) GetContainerOk() (*int32, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *SupersetServingPort) SetContainer(v int32)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *SupersetServingPort) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *SupersetServingPort) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *SupersetServingPort) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetProtocol

`func (o *SupersetServingPort) GetProtocol() string`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *SupersetServingPort) GetProtocolOk() (*string, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *SupersetServingPort) SetProtocol(v string)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *SupersetServingPort) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *SupersetServingPort) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *SupersetServingPort) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil
### GetAuthorizationType

`func (o *SupersetServingPort) GetAuthorizationType() string`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *SupersetServingPort) GetAuthorizationTypeOk() (*string, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *SupersetServingPort) SetAuthorizationType(v string)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *SupersetServingPort) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *SupersetServingPort) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *SupersetServingPort) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *SupersetServingPort) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *SupersetServingPort) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *SupersetServingPort) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *SupersetServingPort) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *SupersetServingPort) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *SupersetServingPort) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *SupersetServingPort) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *SupersetServingPort) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *SupersetServingPort) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *SupersetServingPort) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *SupersetServingPort) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *SupersetServingPort) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetClusterLocalAccessOnly

`func (o *SupersetServingPort) GetClusterLocalAccessOnly() bool`

GetClusterLocalAccessOnly returns the ClusterLocalAccessOnly field if non-nil, zero value otherwise.

### GetClusterLocalAccessOnlyOk

`func (o *SupersetServingPort) GetClusterLocalAccessOnlyOk() (*bool, bool)`

GetClusterLocalAccessOnlyOk returns a tuple with the ClusterLocalAccessOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterLocalAccessOnly

`func (o *SupersetServingPort) SetClusterLocalAccessOnly(v bool)`

SetClusterLocalAccessOnly sets ClusterLocalAccessOnly field to given value.

### HasClusterLocalAccessOnly

`func (o *SupersetServingPort) HasClusterLocalAccessOnly() bool`

HasClusterLocalAccessOnly returns a boolean if a field has been set.

### SetClusterLocalAccessOnlyNil

`func (o *SupersetServingPort) SetClusterLocalAccessOnlyNil(b bool)`

 SetClusterLocalAccessOnlyNil sets the value for ClusterLocalAccessOnly to be an explicit nil

### UnsetClusterLocalAccessOnly
`func (o *SupersetServingPort) UnsetClusterLocalAccessOnly()`

UnsetClusterLocalAccessOnly ensures that no value is present for ClusterLocalAccessOnly, not even an explicit nil
### GetExposeExternally

`func (o *SupersetServingPort) GetExposeExternally() bool`

GetExposeExternally returns the ExposeExternally field if non-nil, zero value otherwise.

### GetExposeExternallyOk

`func (o *SupersetServingPort) GetExposeExternallyOk() (*bool, bool)`

GetExposeExternallyOk returns a tuple with the ExposeExternally field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposeExternally

`func (o *SupersetServingPort) SetExposeExternally(v bool)`

SetExposeExternally sets ExposeExternally field to given value.

### HasExposeExternally

`func (o *SupersetServingPort) HasExposeExternally() bool`

HasExposeExternally returns a boolean if a field has been set.

### SetExposeExternallyNil

`func (o *SupersetServingPort) SetExposeExternallyNil(b bool)`

 SetExposeExternallyNil sets the value for ExposeExternally to be an explicit nil

### UnsetExposeExternally
`func (o *SupersetServingPort) UnsetExposeExternally()`

UnsetExposeExternally ensures that no value is present for ExposeExternally, not even an explicit nil
### GetExposedUrl

`func (o *SupersetServingPort) GetExposedUrl() string`

GetExposedUrl returns the ExposedUrl field if non-nil, zero value otherwise.

### GetExposedUrlOk

`func (o *SupersetServingPort) GetExposedUrlOk() (*string, bool)`

GetExposedUrlOk returns a tuple with the ExposedUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrl

`func (o *SupersetServingPort) SetExposedUrl(v string)`

SetExposedUrl sets ExposedUrl field to given value.

### HasExposedUrl

`func (o *SupersetServingPort) HasExposedUrl() bool`

HasExposedUrl returns a boolean if a field has been set.

### SetExposedUrlNil

`func (o *SupersetServingPort) SetExposedUrlNil(b bool)`

 SetExposedUrlNil sets the value for ExposedUrl to be an explicit nil

### UnsetExposedUrl
`func (o *SupersetServingPort) UnsetExposedUrl()`

UnsetExposedUrl ensures that no value is present for ExposedUrl, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


