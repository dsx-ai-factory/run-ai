# DistributedInferenceServingPortRulesProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Protocol** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AuthorizationType** | Pointer to [**NullableDistributedInferenceServingPortAuthorizationTypeRules**](DistributedInferenceServingPortAuthorizationTypeRules.md) |  | [optional] 
**AuthorizedUsers** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**AuthorizedGroups** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**ExposeExternally** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ExposedUrl** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewDistributedInferenceServingPortRulesProperties

`func NewDistributedInferenceServingPortRulesProperties() *DistributedInferenceServingPortRulesProperties`

NewDistributedInferenceServingPortRulesProperties instantiates a new DistributedInferenceServingPortRulesProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceServingPortRulesPropertiesWithDefaults

`func NewDistributedInferenceServingPortRulesPropertiesWithDefaults() *DistributedInferenceServingPortRulesProperties`

NewDistributedInferenceServingPortRulesPropertiesWithDefaults instantiates a new DistributedInferenceServingPortRulesProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *DistributedInferenceServingPortRulesProperties) GetPort() IntegerRules`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *DistributedInferenceServingPortRulesProperties) GetPortOk() (*IntegerRules, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *DistributedInferenceServingPortRulesProperties) SetPort(v IntegerRules)`

SetPort sets Port field to given value.

### HasPort

`func (o *DistributedInferenceServingPortRulesProperties) HasPort() bool`

HasPort returns a boolean if a field has been set.

### SetPortNil

`func (o *DistributedInferenceServingPortRulesProperties) SetPortNil(b bool)`

 SetPortNil sets the value for Port to be an explicit nil

### UnsetPort
`func (o *DistributedInferenceServingPortRulesProperties) UnsetPort()`

UnsetPort ensures that no value is present for Port, not even an explicit nil
### GetProtocol

`func (o *DistributedInferenceServingPortRulesProperties) GetProtocol() StringRules`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *DistributedInferenceServingPortRulesProperties) GetProtocolOk() (*StringRules, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *DistributedInferenceServingPortRulesProperties) SetProtocol(v StringRules)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *DistributedInferenceServingPortRulesProperties) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### SetProtocolNil

`func (o *DistributedInferenceServingPortRulesProperties) SetProtocolNil(b bool)`

 SetProtocolNil sets the value for Protocol to be an explicit nil

### UnsetProtocol
`func (o *DistributedInferenceServingPortRulesProperties) UnsetProtocol()`

UnsetProtocol ensures that no value is present for Protocol, not even an explicit nil
### GetAuthorizationType

`func (o *DistributedInferenceServingPortRulesProperties) GetAuthorizationType() DistributedInferenceServingPortAuthorizationTypeRules`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *DistributedInferenceServingPortRulesProperties) GetAuthorizationTypeOk() (*DistributedInferenceServingPortAuthorizationTypeRules, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *DistributedInferenceServingPortRulesProperties) SetAuthorizationType(v DistributedInferenceServingPortAuthorizationTypeRules)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *DistributedInferenceServingPortRulesProperties) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *DistributedInferenceServingPortRulesProperties) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *DistributedInferenceServingPortRulesProperties) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *DistributedInferenceServingPortRulesProperties) GetAuthorizedUsers() ArrayRules`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *DistributedInferenceServingPortRulesProperties) GetAuthorizedUsersOk() (*ArrayRules, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *DistributedInferenceServingPortRulesProperties) SetAuthorizedUsers(v ArrayRules)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *DistributedInferenceServingPortRulesProperties) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *DistributedInferenceServingPortRulesProperties) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *DistributedInferenceServingPortRulesProperties) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *DistributedInferenceServingPortRulesProperties) GetAuthorizedGroups() ArrayRules`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *DistributedInferenceServingPortRulesProperties) GetAuthorizedGroupsOk() (*ArrayRules, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *DistributedInferenceServingPortRulesProperties) SetAuthorizedGroups(v ArrayRules)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *DistributedInferenceServingPortRulesProperties) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *DistributedInferenceServingPortRulesProperties) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *DistributedInferenceServingPortRulesProperties) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetExposeExternally

`func (o *DistributedInferenceServingPortRulesProperties) GetExposeExternally() BooleanRules`

GetExposeExternally returns the ExposeExternally field if non-nil, zero value otherwise.

### GetExposeExternallyOk

`func (o *DistributedInferenceServingPortRulesProperties) GetExposeExternallyOk() (*BooleanRules, bool)`

GetExposeExternallyOk returns a tuple with the ExposeExternally field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposeExternally

`func (o *DistributedInferenceServingPortRulesProperties) SetExposeExternally(v BooleanRules)`

SetExposeExternally sets ExposeExternally field to given value.

### HasExposeExternally

`func (o *DistributedInferenceServingPortRulesProperties) HasExposeExternally() bool`

HasExposeExternally returns a boolean if a field has been set.

### SetExposeExternallyNil

`func (o *DistributedInferenceServingPortRulesProperties) SetExposeExternallyNil(b bool)`

 SetExposeExternallyNil sets the value for ExposeExternally to be an explicit nil

### UnsetExposeExternally
`func (o *DistributedInferenceServingPortRulesProperties) UnsetExposeExternally()`

UnsetExposeExternally ensures that no value is present for ExposeExternally, not even an explicit nil
### GetExposedUrl

`func (o *DistributedInferenceServingPortRulesProperties) GetExposedUrl() StringRules`

GetExposedUrl returns the ExposedUrl field if non-nil, zero value otherwise.

### GetExposedUrlOk

`func (o *DistributedInferenceServingPortRulesProperties) GetExposedUrlOk() (*StringRules, bool)`

GetExposedUrlOk returns a tuple with the ExposedUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrl

`func (o *DistributedInferenceServingPortRulesProperties) SetExposedUrl(v StringRules)`

SetExposedUrl sets ExposedUrl field to given value.

### HasExposedUrl

`func (o *DistributedInferenceServingPortRulesProperties) HasExposedUrl() bool`

HasExposedUrl returns a boolean if a field has been set.

### SetExposedUrlNil

`func (o *DistributedInferenceServingPortRulesProperties) SetExposedUrlNil(b bool)`

 SetExposedUrlNil sets the value for ExposedUrl to be an explicit nil

### UnsetExposedUrl
`func (o *DistributedInferenceServingPortRulesProperties) UnsetExposedUrl()`

UnsetExposedUrl ensures that no value is present for ExposedUrl, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


