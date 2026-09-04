# Connection1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**ToolType** | **string** |  | 
**ConnectionType** | **string** |  | 
**Url** | Pointer to **string** |  | [optional] 
**AuthorizedUsers** | Pointer to **[]string** |  | [optional] 
**AuthorizedGroups** | Pointer to **[]string** |  | [optional] 

## Methods

### NewConnection1

`func NewConnection1(name string, toolType string, connectionType string, ) *Connection1`

NewConnection1 instantiates a new Connection1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnection1WithDefaults

`func NewConnection1WithDefaults() *Connection1`

NewConnection1WithDefaults instantiates a new Connection1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Connection1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Connection1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Connection1) SetName(v string)`

SetName sets Name field to given value.


### GetToolType

`func (o *Connection1) GetToolType() string`

GetToolType returns the ToolType field if non-nil, zero value otherwise.

### GetToolTypeOk

`func (o *Connection1) GetToolTypeOk() (*string, bool)`

GetToolTypeOk returns a tuple with the ToolType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolType

`func (o *Connection1) SetToolType(v string)`

SetToolType sets ToolType field to given value.


### GetConnectionType

`func (o *Connection1) GetConnectionType() string`

GetConnectionType returns the ConnectionType field if non-nil, zero value otherwise.

### GetConnectionTypeOk

`func (o *Connection1) GetConnectionTypeOk() (*string, bool)`

GetConnectionTypeOk returns a tuple with the ConnectionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionType

`func (o *Connection1) SetConnectionType(v string)`

SetConnectionType sets ConnectionType field to given value.


### GetUrl

`func (o *Connection1) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *Connection1) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *Connection1) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *Connection1) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetAuthorizedUsers

`func (o *Connection1) GetAuthorizedUsers() []string`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *Connection1) GetAuthorizedUsersOk() (*[]string, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *Connection1) SetAuthorizedUsers(v []string)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *Connection1) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### GetAuthorizedGroups

`func (o *Connection1) GetAuthorizedGroups() []string`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *Connection1) GetAuthorizedGroupsOk() (*[]string, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *Connection1) SetAuthorizedGroups(v []string)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *Connection1) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


