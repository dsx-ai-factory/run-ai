# NimServiceServingPortRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServiceType** | Pointer to [**NullableServingPortServiceTypeRules**](ServingPortServiceTypeRules.md) |  | [optional] 
**Port** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**GrpcPort** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**MetricsPort** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**ExposeExternally** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**ExposedUrl** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ExposedProtocol** | Pointer to [**NullableServingPortExposedProtocolRules**](ServingPortExposedProtocolRules.md) |  | [optional] 
**AuthorizationType** | Pointer to [**NullableNimServiceServingPortAuthorizationTypeRules**](NimServiceServingPortAuthorizationTypeRules.md) |  | [optional] 
**AuthorizedUsers** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**AuthorizedGroups** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 

## Methods

### NewNimServiceServingPortRules

`func NewNimServiceServingPortRules() *NimServiceServingPortRules`

NewNimServiceServingPortRules instantiates a new NimServiceServingPortRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceServingPortRulesWithDefaults

`func NewNimServiceServingPortRulesWithDefaults() *NimServiceServingPortRules`

NewNimServiceServingPortRulesWithDefaults instantiates a new NimServiceServingPortRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServiceType

`func (o *NimServiceServingPortRules) GetServiceType() ServingPortServiceTypeRules`

GetServiceType returns the ServiceType field if non-nil, zero value otherwise.

### GetServiceTypeOk

`func (o *NimServiceServingPortRules) GetServiceTypeOk() (*ServingPortServiceTypeRules, bool)`

GetServiceTypeOk returns a tuple with the ServiceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceType

`func (o *NimServiceServingPortRules) SetServiceType(v ServingPortServiceTypeRules)`

SetServiceType sets ServiceType field to given value.

### HasServiceType

`func (o *NimServiceServingPortRules) HasServiceType() bool`

HasServiceType returns a boolean if a field has been set.

### SetServiceTypeNil

`func (o *NimServiceServingPortRules) SetServiceTypeNil(b bool)`

 SetServiceTypeNil sets the value for ServiceType to be an explicit nil

### UnsetServiceType
`func (o *NimServiceServingPortRules) UnsetServiceType()`

UnsetServiceType ensures that no value is present for ServiceType, not even an explicit nil
### GetPort

`func (o *NimServiceServingPortRules) GetPort() IntegerRules`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *NimServiceServingPortRules) GetPortOk() (*IntegerRules, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *NimServiceServingPortRules) SetPort(v IntegerRules)`

SetPort sets Port field to given value.

### HasPort

`func (o *NimServiceServingPortRules) HasPort() bool`

HasPort returns a boolean if a field has been set.

### SetPortNil

`func (o *NimServiceServingPortRules) SetPortNil(b bool)`

 SetPortNil sets the value for Port to be an explicit nil

### UnsetPort
`func (o *NimServiceServingPortRules) UnsetPort()`

UnsetPort ensures that no value is present for Port, not even an explicit nil
### GetGrpcPort

`func (o *NimServiceServingPortRules) GetGrpcPort() IntegerRules`

GetGrpcPort returns the GrpcPort field if non-nil, zero value otherwise.

### GetGrpcPortOk

`func (o *NimServiceServingPortRules) GetGrpcPortOk() (*IntegerRules, bool)`

GetGrpcPortOk returns a tuple with the GrpcPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrpcPort

`func (o *NimServiceServingPortRules) SetGrpcPort(v IntegerRules)`

SetGrpcPort sets GrpcPort field to given value.

### HasGrpcPort

`func (o *NimServiceServingPortRules) HasGrpcPort() bool`

HasGrpcPort returns a boolean if a field has been set.

### SetGrpcPortNil

`func (o *NimServiceServingPortRules) SetGrpcPortNil(b bool)`

 SetGrpcPortNil sets the value for GrpcPort to be an explicit nil

### UnsetGrpcPort
`func (o *NimServiceServingPortRules) UnsetGrpcPort()`

UnsetGrpcPort ensures that no value is present for GrpcPort, not even an explicit nil
### GetMetricsPort

`func (o *NimServiceServingPortRules) GetMetricsPort() IntegerRules`

GetMetricsPort returns the MetricsPort field if non-nil, zero value otherwise.

### GetMetricsPortOk

`func (o *NimServiceServingPortRules) GetMetricsPortOk() (*IntegerRules, bool)`

GetMetricsPortOk returns a tuple with the MetricsPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricsPort

`func (o *NimServiceServingPortRules) SetMetricsPort(v IntegerRules)`

SetMetricsPort sets MetricsPort field to given value.

### HasMetricsPort

`func (o *NimServiceServingPortRules) HasMetricsPort() bool`

HasMetricsPort returns a boolean if a field has been set.

### SetMetricsPortNil

`func (o *NimServiceServingPortRules) SetMetricsPortNil(b bool)`

 SetMetricsPortNil sets the value for MetricsPort to be an explicit nil

### UnsetMetricsPort
`func (o *NimServiceServingPortRules) UnsetMetricsPort()`

UnsetMetricsPort ensures that no value is present for MetricsPort, not even an explicit nil
### GetExposeExternally

`func (o *NimServiceServingPortRules) GetExposeExternally() BooleanRules`

GetExposeExternally returns the ExposeExternally field if non-nil, zero value otherwise.

### GetExposeExternallyOk

`func (o *NimServiceServingPortRules) GetExposeExternallyOk() (*BooleanRules, bool)`

GetExposeExternallyOk returns a tuple with the ExposeExternally field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposeExternally

`func (o *NimServiceServingPortRules) SetExposeExternally(v BooleanRules)`

SetExposeExternally sets ExposeExternally field to given value.

### HasExposeExternally

`func (o *NimServiceServingPortRules) HasExposeExternally() bool`

HasExposeExternally returns a boolean if a field has been set.

### SetExposeExternallyNil

`func (o *NimServiceServingPortRules) SetExposeExternallyNil(b bool)`

 SetExposeExternallyNil sets the value for ExposeExternally to be an explicit nil

### UnsetExposeExternally
`func (o *NimServiceServingPortRules) UnsetExposeExternally()`

UnsetExposeExternally ensures that no value is present for ExposeExternally, not even an explicit nil
### GetExposedUrl

`func (o *NimServiceServingPortRules) GetExposedUrl() StringRules`

GetExposedUrl returns the ExposedUrl field if non-nil, zero value otherwise.

### GetExposedUrlOk

`func (o *NimServiceServingPortRules) GetExposedUrlOk() (*StringRules, bool)`

GetExposedUrlOk returns a tuple with the ExposedUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrl

`func (o *NimServiceServingPortRules) SetExposedUrl(v StringRules)`

SetExposedUrl sets ExposedUrl field to given value.

### HasExposedUrl

`func (o *NimServiceServingPortRules) HasExposedUrl() bool`

HasExposedUrl returns a boolean if a field has been set.

### SetExposedUrlNil

`func (o *NimServiceServingPortRules) SetExposedUrlNil(b bool)`

 SetExposedUrlNil sets the value for ExposedUrl to be an explicit nil

### UnsetExposedUrl
`func (o *NimServiceServingPortRules) UnsetExposedUrl()`

UnsetExposedUrl ensures that no value is present for ExposedUrl, not even an explicit nil
### GetExposedProtocol

`func (o *NimServiceServingPortRules) GetExposedProtocol() ServingPortExposedProtocolRules`

GetExposedProtocol returns the ExposedProtocol field if non-nil, zero value otherwise.

### GetExposedProtocolOk

`func (o *NimServiceServingPortRules) GetExposedProtocolOk() (*ServingPortExposedProtocolRules, bool)`

GetExposedProtocolOk returns a tuple with the ExposedProtocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedProtocol

`func (o *NimServiceServingPortRules) SetExposedProtocol(v ServingPortExposedProtocolRules)`

SetExposedProtocol sets ExposedProtocol field to given value.

### HasExposedProtocol

`func (o *NimServiceServingPortRules) HasExposedProtocol() bool`

HasExposedProtocol returns a boolean if a field has been set.

### SetExposedProtocolNil

`func (o *NimServiceServingPortRules) SetExposedProtocolNil(b bool)`

 SetExposedProtocolNil sets the value for ExposedProtocol to be an explicit nil

### UnsetExposedProtocol
`func (o *NimServiceServingPortRules) UnsetExposedProtocol()`

UnsetExposedProtocol ensures that no value is present for ExposedProtocol, not even an explicit nil
### GetAuthorizationType

`func (o *NimServiceServingPortRules) GetAuthorizationType() NimServiceServingPortAuthorizationTypeRules`

GetAuthorizationType returns the AuthorizationType field if non-nil, zero value otherwise.

### GetAuthorizationTypeOk

`func (o *NimServiceServingPortRules) GetAuthorizationTypeOk() (*NimServiceServingPortAuthorizationTypeRules, bool)`

GetAuthorizationTypeOk returns a tuple with the AuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizationType

`func (o *NimServiceServingPortRules) SetAuthorizationType(v NimServiceServingPortAuthorizationTypeRules)`

SetAuthorizationType sets AuthorizationType field to given value.

### HasAuthorizationType

`func (o *NimServiceServingPortRules) HasAuthorizationType() bool`

HasAuthorizationType returns a boolean if a field has been set.

### SetAuthorizationTypeNil

`func (o *NimServiceServingPortRules) SetAuthorizationTypeNil(b bool)`

 SetAuthorizationTypeNil sets the value for AuthorizationType to be an explicit nil

### UnsetAuthorizationType
`func (o *NimServiceServingPortRules) UnsetAuthorizationType()`

UnsetAuthorizationType ensures that no value is present for AuthorizationType, not even an explicit nil
### GetAuthorizedUsers

`func (o *NimServiceServingPortRules) GetAuthorizedUsers() ArrayRules`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *NimServiceServingPortRules) GetAuthorizedUsersOk() (*ArrayRules, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *NimServiceServingPortRules) SetAuthorizedUsers(v ArrayRules)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *NimServiceServingPortRules) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *NimServiceServingPortRules) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *NimServiceServingPortRules) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *NimServiceServingPortRules) GetAuthorizedGroups() ArrayRules`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *NimServiceServingPortRules) GetAuthorizedGroupsOk() (*ArrayRules, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *NimServiceServingPortRules) SetAuthorizedGroups(v ArrayRules)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *NimServiceServingPortRules) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *NimServiceServingPortRules) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *NimServiceServingPortRules) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


