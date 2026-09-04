# User1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UserId** | **string** |  | 
**Email** | Pointer to **string** |  | [optional] 
**Role** | **string** |  | 
**Roles** | Pointer to **[]string** |  | [optional] 
**Groups** | Pointer to **[]string** |  | [optional] 

## Methods

### NewUser1

`func NewUser1(userId string, role string, ) *User1`

NewUser1 instantiates a new User1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUser1WithDefaults

`func NewUser1WithDefaults() *User1`

NewUser1WithDefaults instantiates a new User1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUserId

`func (o *User1) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *User1) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *User1) SetUserId(v string)`

SetUserId sets UserId field to given value.


### GetEmail

`func (o *User1) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *User1) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *User1) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *User1) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetRole

`func (o *User1) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *User1) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *User1) SetRole(v string)`

SetRole sets Role field to given value.


### GetRoles

`func (o *User1) GetRoles() []string`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *User1) GetRolesOk() (*[]string, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *User1) SetRoles(v []string)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *User1) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetGroups

`func (o *User1) GetGroups() []string`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *User1) GetGroupsOk() (*[]string, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *User1) SetGroups(v []string)`

SetGroups sets Groups field to given value.

### HasGroups

`func (o *User1) HasGroups() bool`

HasGroups returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


