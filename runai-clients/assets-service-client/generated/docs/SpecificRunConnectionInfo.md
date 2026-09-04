# SpecificRunConnectionInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Unique identifier of the connection. Will be used to correlate between the information given here and the information provided for the connection in the environment asset. | [optional] 
**NodePort** | Pointer to **NullableInt32** | Port number of the host that will be connected with the container port. Required if and only if isCustomPort is set to true in the environment asset. | [optional] 
**ExternalUrl** | Pointer to **NullableString** | URL to associated with the container port. Required if and only if isCustomExternalUrl is set to true in the environment asset. | [optional] 
**AuthorizedUsers** | Pointer to **[]string** | List of users or applications that are allowed to access the URL. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 
**AuthorizedGroups** | Pointer to **[]string** | Specifies the names of those groups that are allowed to access the connection. Note that authorizedUsers and authorizedGroups are mutually exclusive. | [optional] 

## Methods

### NewSpecificRunConnectionInfo

`func NewSpecificRunConnectionInfo() *SpecificRunConnectionInfo`

NewSpecificRunConnectionInfo instantiates a new SpecificRunConnectionInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunConnectionInfoWithDefaults

`func NewSpecificRunConnectionInfoWithDefaults() *SpecificRunConnectionInfo`

NewSpecificRunConnectionInfoWithDefaults instantiates a new SpecificRunConnectionInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *SpecificRunConnectionInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SpecificRunConnectionInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SpecificRunConnectionInfo) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *SpecificRunConnectionInfo) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNodePort

`func (o *SpecificRunConnectionInfo) GetNodePort() int32`

GetNodePort returns the NodePort field if non-nil, zero value otherwise.

### GetNodePortOk

`func (o *SpecificRunConnectionInfo) GetNodePortOk() (*int32, bool)`

GetNodePortOk returns a tuple with the NodePort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePort

`func (o *SpecificRunConnectionInfo) SetNodePort(v int32)`

SetNodePort sets NodePort field to given value.

### HasNodePort

`func (o *SpecificRunConnectionInfo) HasNodePort() bool`

HasNodePort returns a boolean if a field has been set.

### SetNodePortNil

`func (o *SpecificRunConnectionInfo) SetNodePortNil(b bool)`

 SetNodePortNil sets the value for NodePort to be an explicit nil

### UnsetNodePort
`func (o *SpecificRunConnectionInfo) UnsetNodePort()`

UnsetNodePort ensures that no value is present for NodePort, not even an explicit nil
### GetExternalUrl

`func (o *SpecificRunConnectionInfo) GetExternalUrl() string`

GetExternalUrl returns the ExternalUrl field if non-nil, zero value otherwise.

### GetExternalUrlOk

`func (o *SpecificRunConnectionInfo) GetExternalUrlOk() (*string, bool)`

GetExternalUrlOk returns a tuple with the ExternalUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalUrl

`func (o *SpecificRunConnectionInfo) SetExternalUrl(v string)`

SetExternalUrl sets ExternalUrl field to given value.

### HasExternalUrl

`func (o *SpecificRunConnectionInfo) HasExternalUrl() bool`

HasExternalUrl returns a boolean if a field has been set.

### SetExternalUrlNil

`func (o *SpecificRunConnectionInfo) SetExternalUrlNil(b bool)`

 SetExternalUrlNil sets the value for ExternalUrl to be an explicit nil

### UnsetExternalUrl
`func (o *SpecificRunConnectionInfo) UnsetExternalUrl()`

UnsetExternalUrl ensures that no value is present for ExternalUrl, not even an explicit nil
### GetAuthorizedUsers

`func (o *SpecificRunConnectionInfo) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *SpecificRunConnectionInfo) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *SpecificRunConnectionInfo) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *SpecificRunConnectionInfo) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *SpecificRunConnectionInfo) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *SpecificRunConnectionInfo) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *SpecificRunConnectionInfo) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *SpecificRunConnectionInfo) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *SpecificRunConnectionInfo) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *SpecificRunConnectionInfo) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *SpecificRunConnectionInfo) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *SpecificRunConnectionInfo) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


