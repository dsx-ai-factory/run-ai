# ExposedUrlRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Container** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**CustomUrl** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**Url** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AuthorizedUsers** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**AuthorizedGroups** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**ToolType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ToolName** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewExposedUrlRules

`func NewExposedUrlRules() *ExposedUrlRules`

NewExposedUrlRules instantiates a new ExposedUrlRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExposedUrlRulesWithDefaults

`func NewExposedUrlRulesWithDefaults() *ExposedUrlRules`

NewExposedUrlRulesWithDefaults instantiates a new ExposedUrlRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainer

`func (o *ExposedUrlRules) GetContainer() IntegerRules`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *ExposedUrlRules) GetContainerOk() (*IntegerRules, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *ExposedUrlRules) SetContainer(v IntegerRules)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *ExposedUrlRules) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### SetContainerNil

`func (o *ExposedUrlRules) SetContainerNil(b bool)`

 SetContainerNil sets the value for Container to be an explicit nil

### UnsetContainer
`func (o *ExposedUrlRules) UnsetContainer()`

UnsetContainer ensures that no value is present for Container, not even an explicit nil
### GetCustomUrl

`func (o *ExposedUrlRules) GetCustomUrl() BooleanRules`

GetCustomUrl returns the CustomUrl field if non-nil, zero value otherwise.

### GetCustomUrlOk

`func (o *ExposedUrlRules) GetCustomUrlOk() (*BooleanRules, bool)`

GetCustomUrlOk returns a tuple with the CustomUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomUrl

`func (o *ExposedUrlRules) SetCustomUrl(v BooleanRules)`

SetCustomUrl sets CustomUrl field to given value.

### HasCustomUrl

`func (o *ExposedUrlRules) HasCustomUrl() bool`

HasCustomUrl returns a boolean if a field has been set.

### SetCustomUrlNil

`func (o *ExposedUrlRules) SetCustomUrlNil(b bool)`

 SetCustomUrlNil sets the value for CustomUrl to be an explicit nil

### UnsetCustomUrl
`func (o *ExposedUrlRules) UnsetCustomUrl()`

UnsetCustomUrl ensures that no value is present for CustomUrl, not even an explicit nil
### GetUrl

`func (o *ExposedUrlRules) GetUrl() StringRules`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *ExposedUrlRules) GetUrlOk() (*StringRules, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *ExposedUrlRules) SetUrl(v StringRules)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *ExposedUrlRules) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *ExposedUrlRules) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *ExposedUrlRules) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil
### GetAuthorizedUsers

`func (o *ExposedUrlRules) GetAuthorizedUsers() ArrayRules`

GetAuthorizedUsers returns the AuthorizedUsers field if non-nil, zero value otherwise.

### GetAuthorizedUsersOk

`func (o *ExposedUrlRules) GetAuthorizedUsersOk() (*ArrayRules, bool)`

GetAuthorizedUsersOk returns a tuple with the AuthorizedUsers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedUsers

`func (o *ExposedUrlRules) SetAuthorizedUsers(v ArrayRules)`

SetAuthorizedUsers sets AuthorizedUsers field to given value.

### HasAuthorizedUsers

`func (o *ExposedUrlRules) HasAuthorizedUsers() bool`

HasAuthorizedUsers returns a boolean if a field has been set.

### SetAuthorizedUsersNil

`func (o *ExposedUrlRules) SetAuthorizedUsersNil(b bool)`

 SetAuthorizedUsersNil sets the value for AuthorizedUsers to be an explicit nil

### UnsetAuthorizedUsers
`func (o *ExposedUrlRules) UnsetAuthorizedUsers()`

UnsetAuthorizedUsers ensures that no value is present for AuthorizedUsers, not even an explicit nil
### GetAuthorizedGroups

`func (o *ExposedUrlRules) GetAuthorizedGroups() ArrayRules`

GetAuthorizedGroups returns the AuthorizedGroups field if non-nil, zero value otherwise.

### GetAuthorizedGroupsOk

`func (o *ExposedUrlRules) GetAuthorizedGroupsOk() (*ArrayRules, bool)`

GetAuthorizedGroupsOk returns a tuple with the AuthorizedGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedGroups

`func (o *ExposedUrlRules) SetAuthorizedGroups(v ArrayRules)`

SetAuthorizedGroups sets AuthorizedGroups field to given value.

### HasAuthorizedGroups

`func (o *ExposedUrlRules) HasAuthorizedGroups() bool`

HasAuthorizedGroups returns a boolean if a field has been set.

### SetAuthorizedGroupsNil

`func (o *ExposedUrlRules) SetAuthorizedGroupsNil(b bool)`

 SetAuthorizedGroupsNil sets the value for AuthorizedGroups to be an explicit nil

### UnsetAuthorizedGroups
`func (o *ExposedUrlRules) UnsetAuthorizedGroups()`

UnsetAuthorizedGroups ensures that no value is present for AuthorizedGroups, not even an explicit nil
### GetToolType

`func (o *ExposedUrlRules) GetToolType() StringRules`

GetToolType returns the ToolType field if non-nil, zero value otherwise.

### GetToolTypeOk

`func (o *ExposedUrlRules) GetToolTypeOk() (*StringRules, bool)`

GetToolTypeOk returns a tuple with the ToolType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolType

`func (o *ExposedUrlRules) SetToolType(v StringRules)`

SetToolType sets ToolType field to given value.

### HasToolType

`func (o *ExposedUrlRules) HasToolType() bool`

HasToolType returns a boolean if a field has been set.

### SetToolTypeNil

`func (o *ExposedUrlRules) SetToolTypeNil(b bool)`

 SetToolTypeNil sets the value for ToolType to be an explicit nil

### UnsetToolType
`func (o *ExposedUrlRules) UnsetToolType()`

UnsetToolType ensures that no value is present for ToolType, not even an explicit nil
### GetToolName

`func (o *ExposedUrlRules) GetToolName() StringRules`

GetToolName returns the ToolName field if non-nil, zero value otherwise.

### GetToolNameOk

`func (o *ExposedUrlRules) GetToolNameOk() (*StringRules, bool)`

GetToolNameOk returns a tuple with the ToolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolName

`func (o *ExposedUrlRules) SetToolName(v StringRules)`

SetToolName sets ToolName field to given value.

### HasToolName

`func (o *ExposedUrlRules) HasToolName() bool`

HasToolName returns a boolean if a field has been set.

### SetToolNameNil

`func (o *ExposedUrlRules) SetToolNameNil(b bool)`

 SetToolNameNil sets the value for ToolName to be an explicit nil

### UnsetToolName
`func (o *ExposedUrlRules) UnsetToolName()`

UnsetToolName ensures that no value is present for ToolName, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


