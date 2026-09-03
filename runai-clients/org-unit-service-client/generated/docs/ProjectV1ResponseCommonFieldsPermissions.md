# ProjectV1ResponseCommonFieldsPermissions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Users** | **[]string** | Names of users that have permissions to the project. | 
**Groups** | **[]string** | Names of groups that have permissions to the project. | 
**Applications** | Pointer to **[]string** | Names of applications that have permissions to the project. | [optional] 

## Methods

### NewProjectV1ResponseCommonFieldsPermissions

`func NewProjectV1ResponseCommonFieldsPermissions(users []string, groups []string, ) *ProjectV1ResponseCommonFieldsPermissions`

NewProjectV1ResponseCommonFieldsPermissions instantiates a new ProjectV1ResponseCommonFieldsPermissions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1ResponseCommonFieldsPermissionsWithDefaults

`func NewProjectV1ResponseCommonFieldsPermissionsWithDefaults() *ProjectV1ResponseCommonFieldsPermissions`

NewProjectV1ResponseCommonFieldsPermissionsWithDefaults instantiates a new ProjectV1ResponseCommonFieldsPermissions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsers

`func (o *ProjectV1ResponseCommonFieldsPermissions) GetUsers() []string`

GetUsers returns the Users field if non-nil, zero value otherwise.

### GetUsersOk

`func (o *ProjectV1ResponseCommonFieldsPermissions) GetUsersOk() (*[]string, bool)`

GetUsersOk returns a tuple with the Users field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsers

`func (o *ProjectV1ResponseCommonFieldsPermissions) SetUsers(v []string)`

SetUsers sets Users field to given value.


### SetUsersNil

`func (o *ProjectV1ResponseCommonFieldsPermissions) SetUsersNil(b bool)`

 SetUsersNil sets the value for Users to be an explicit nil

### UnsetUsers
`func (o *ProjectV1ResponseCommonFieldsPermissions) UnsetUsers()`

UnsetUsers ensures that no value is present for Users, not even an explicit nil
### GetGroups

`func (o *ProjectV1ResponseCommonFieldsPermissions) GetGroups() []string`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *ProjectV1ResponseCommonFieldsPermissions) GetGroupsOk() (*[]string, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *ProjectV1ResponseCommonFieldsPermissions) SetGroups(v []string)`

SetGroups sets Groups field to given value.


### SetGroupsNil

`func (o *ProjectV1ResponseCommonFieldsPermissions) SetGroupsNil(b bool)`

 SetGroupsNil sets the value for Groups to be an explicit nil

### UnsetGroups
`func (o *ProjectV1ResponseCommonFieldsPermissions) UnsetGroups()`

UnsetGroups ensures that no value is present for Groups, not even an explicit nil
### GetApplications

`func (o *ProjectV1ResponseCommonFieldsPermissions) GetApplications() []string`

GetApplications returns the Applications field if non-nil, zero value otherwise.

### GetApplicationsOk

`func (o *ProjectV1ResponseCommonFieldsPermissions) GetApplicationsOk() (*[]string, bool)`

GetApplicationsOk returns a tuple with the Applications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplications

`func (o *ProjectV1ResponseCommonFieldsPermissions) SetApplications(v []string)`

SetApplications sets Applications field to given value.

### HasApplications

`func (o *ProjectV1ResponseCommonFieldsPermissions) HasApplications() bool`

HasApplications returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


