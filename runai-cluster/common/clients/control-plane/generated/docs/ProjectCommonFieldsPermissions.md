# ProjectCommonFieldsPermissions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Users** | Pointer to **[]string** | Names of users that have permissions to the project. | [optional] 
**Groups** | Pointer to **[]string** | Names of groups that have permissions to the project. | [optional] 
**Applications** | Pointer to **[]string** | Names of applications that have permissions to the project. | [optional] 

## Methods

### NewProjectCommonFieldsPermissions

`func NewProjectCommonFieldsPermissions() *ProjectCommonFieldsPermissions`

NewProjectCommonFieldsPermissions instantiates a new ProjectCommonFieldsPermissions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectCommonFieldsPermissionsWithDefaults

`func NewProjectCommonFieldsPermissionsWithDefaults() *ProjectCommonFieldsPermissions`

NewProjectCommonFieldsPermissionsWithDefaults instantiates a new ProjectCommonFieldsPermissions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsers

`func (o *ProjectCommonFieldsPermissions) GetUsers() []string`

GetUsers returns the Users field if non-nil, zero value otherwise.

### GetUsersOk

`func (o *ProjectCommonFieldsPermissions) GetUsersOk() (*[]string, bool)`

GetUsersOk returns a tuple with the Users field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsers

`func (o *ProjectCommonFieldsPermissions) SetUsers(v []string)`

SetUsers sets Users field to given value.

### HasUsers

`func (o *ProjectCommonFieldsPermissions) HasUsers() bool`

HasUsers returns a boolean if a field has been set.

### GetGroups

`func (o *ProjectCommonFieldsPermissions) GetGroups() []string`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *ProjectCommonFieldsPermissions) GetGroupsOk() (*[]string, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *ProjectCommonFieldsPermissions) SetGroups(v []string)`

SetGroups sets Groups field to given value.

### HasGroups

`func (o *ProjectCommonFieldsPermissions) HasGroups() bool`

HasGroups returns a boolean if a field has been set.

### GetApplications

`func (o *ProjectCommonFieldsPermissions) GetApplications() []string`

GetApplications returns the Applications field if non-nil, zero value otherwise.

### GetApplicationsOk

`func (o *ProjectCommonFieldsPermissions) GetApplicationsOk() (*[]string, bool)`

GetApplicationsOk returns a tuple with the Applications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplications

`func (o *ProjectCommonFieldsPermissions) SetApplications(v []string)`

SetApplications sets Applications field to given value.

### HasApplications

`func (o *ProjectCommonFieldsPermissions) HasApplications() bool`

HasApplications returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


