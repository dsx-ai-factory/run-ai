# AccessRulesMigrateDeprecatedFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OldRoleId** | **int32** | The numeric ID of the deprecated role whose access rules should be migrated. Must be a deprecated role in the current tenant. | 
**NewRoleId** | **int32** | The numeric ID of the non-deprecated role to migrate access rules to. Must not be a deprecated role, and must be enabled. | 

## Methods

### NewAccessRulesMigrateDeprecatedFields

`func NewAccessRulesMigrateDeprecatedFields(oldRoleId int32, newRoleId int32, ) *AccessRulesMigrateDeprecatedFields`

NewAccessRulesMigrateDeprecatedFields instantiates a new AccessRulesMigrateDeprecatedFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRulesMigrateDeprecatedFieldsWithDefaults

`func NewAccessRulesMigrateDeprecatedFieldsWithDefaults() *AccessRulesMigrateDeprecatedFields`

NewAccessRulesMigrateDeprecatedFieldsWithDefaults instantiates a new AccessRulesMigrateDeprecatedFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOldRoleId

`func (o *AccessRulesMigrateDeprecatedFields) GetOldRoleId() int32`

GetOldRoleId returns the OldRoleId field if non-nil, zero value otherwise.

### GetOldRoleIdOk

`func (o *AccessRulesMigrateDeprecatedFields) GetOldRoleIdOk() (*int32, bool)`

GetOldRoleIdOk returns a tuple with the OldRoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOldRoleId

`func (o *AccessRulesMigrateDeprecatedFields) SetOldRoleId(v int32)`

SetOldRoleId sets OldRoleId field to given value.


### GetNewRoleId

`func (o *AccessRulesMigrateDeprecatedFields) GetNewRoleId() int32`

GetNewRoleId returns the NewRoleId field if non-nil, zero value otherwise.

### GetNewRoleIdOk

`func (o *AccessRulesMigrateDeprecatedFields) GetNewRoleIdOk() (*int32, bool)`

GetNewRoleIdOk returns a tuple with the NewRoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewRoleId

`func (o *AccessRulesMigrateDeprecatedFields) SetNewRoleId(v int32)`

SetNewRoleId sets NewRoleId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


