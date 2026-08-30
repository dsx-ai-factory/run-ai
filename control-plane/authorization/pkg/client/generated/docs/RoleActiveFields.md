# RoleActiveFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScopeType** | Pointer to **string** |  | [optional] [default to "tenant"]
**Enabled** | **bool** |  | [default to true]

## Methods

### NewRoleActiveFields

`func NewRoleActiveFields(enabled bool, ) *RoleActiveFields`

NewRoleActiveFields instantiates a new RoleActiveFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleActiveFieldsWithDefaults

`func NewRoleActiveFieldsWithDefaults() *RoleActiveFields`

NewRoleActiveFieldsWithDefaults instantiates a new RoleActiveFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScopeType

`func (o *RoleActiveFields) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *RoleActiveFields) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *RoleActiveFields) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.

### HasScopeType

`func (o *RoleActiveFields) HasScopeType() bool`

HasScopeType returns a boolean if a field has been set.

### GetEnabled

`func (o *RoleActiveFields) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *RoleActiveFields) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *RoleActiveFields) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


