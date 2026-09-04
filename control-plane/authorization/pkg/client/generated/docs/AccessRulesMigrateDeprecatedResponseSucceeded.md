# AccessRulesMigrateDeprecatedResponseSucceeded

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Created** | [**[]AccessRule**](AccessRule.md) | New-role access rules that were freshly created as part of this migration. | 
**PreExisting** | [**[]AccessRule**](AccessRule.md) | New-role access rules that already existed at the start of the call (same subject + scope + new role). No new rule was created for these, but the corresponding old (deprecated-role) rule was still deleted since the subject already has the new role on that scope. | 
**Deleted** | **[]string** | List of access rule IDs that were successfully deleted. | 

## Methods

### NewAccessRulesMigrateDeprecatedResponseSucceeded

`func NewAccessRulesMigrateDeprecatedResponseSucceeded(created []AccessRule, preExisting []AccessRule, deleted []string, ) *AccessRulesMigrateDeprecatedResponseSucceeded`

NewAccessRulesMigrateDeprecatedResponseSucceeded instantiates a new AccessRulesMigrateDeprecatedResponseSucceeded object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRulesMigrateDeprecatedResponseSucceededWithDefaults

`func NewAccessRulesMigrateDeprecatedResponseSucceededWithDefaults() *AccessRulesMigrateDeprecatedResponseSucceeded`

NewAccessRulesMigrateDeprecatedResponseSucceededWithDefaults instantiates a new AccessRulesMigrateDeprecatedResponseSucceeded object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreated

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) GetCreated() []AccessRule`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) GetCreatedOk() (*[]AccessRule, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) SetCreated(v []AccessRule)`

SetCreated sets Created field to given value.


### GetPreExisting

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) GetPreExisting() []AccessRule`

GetPreExisting returns the PreExisting field if non-nil, zero value otherwise.

### GetPreExistingOk

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) GetPreExistingOk() (*[]AccessRule, bool)`

GetPreExistingOk returns a tuple with the PreExisting field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreExisting

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) SetPreExisting(v []AccessRule)`

SetPreExisting sets PreExisting field to given value.


### GetDeleted

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) GetDeleted() []string`

GetDeleted returns the Deleted field if non-nil, zero value otherwise.

### GetDeletedOk

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) GetDeletedOk() (*[]string, bool)`

GetDeletedOk returns a tuple with the Deleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleted

`func (o *AccessRulesMigrateDeprecatedResponseSucceeded) SetDeleted(v []string)`

SetDeleted sets Deleted field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


