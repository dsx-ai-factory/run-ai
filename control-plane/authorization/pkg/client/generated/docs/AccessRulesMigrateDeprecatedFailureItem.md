# AccessRulesMigrateDeprecatedFailureItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OldAccessRuleId** | **int32** | The ID of the access rule on the deprecated role that failed to migrate. This rule is still alive and should be handled separately. | 
**Step** | **string** | Which step of the migration failed. \&quot;create\&quot; means the new access rule could not be created. \&quot;delete\&quot; means the new rule was created (or already existed) but the old rule could not be deleted, so the transaction was rolled back. | 
**Code** | **int32** | A numeric code representing the error type (typically an HTTP status code). | 
**Message** | **string** | A human-readable message describing why the migration failed for this access rule. | 

## Methods

### NewAccessRulesMigrateDeprecatedFailureItem

`func NewAccessRulesMigrateDeprecatedFailureItem(oldAccessRuleId int32, step string, code int32, message string, ) *AccessRulesMigrateDeprecatedFailureItem`

NewAccessRulesMigrateDeprecatedFailureItem instantiates a new AccessRulesMigrateDeprecatedFailureItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRulesMigrateDeprecatedFailureItemWithDefaults

`func NewAccessRulesMigrateDeprecatedFailureItemWithDefaults() *AccessRulesMigrateDeprecatedFailureItem`

NewAccessRulesMigrateDeprecatedFailureItemWithDefaults instantiates a new AccessRulesMigrateDeprecatedFailureItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOldAccessRuleId

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetOldAccessRuleId() int32`

GetOldAccessRuleId returns the OldAccessRuleId field if non-nil, zero value otherwise.

### GetOldAccessRuleIdOk

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetOldAccessRuleIdOk() (*int32, bool)`

GetOldAccessRuleIdOk returns a tuple with the OldAccessRuleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOldAccessRuleId

`func (o *AccessRulesMigrateDeprecatedFailureItem) SetOldAccessRuleId(v int32)`

SetOldAccessRuleId sets OldAccessRuleId field to given value.


### GetStep

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetStep() string`

GetStep returns the Step field if non-nil, zero value otherwise.

### GetStepOk

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetStepOk() (*string, bool)`

GetStepOk returns a tuple with the Step field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStep

`func (o *AccessRulesMigrateDeprecatedFailureItem) SetStep(v string)`

SetStep sets Step field to given value.


### GetCode

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AccessRulesMigrateDeprecatedFailureItem) SetCode(v int32)`

SetCode sets Code field to given value.


### GetMessage

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AccessRulesMigrateDeprecatedFailureItem) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AccessRulesMigrateDeprecatedFailureItem) SetMessage(v string)`

SetMessage sets Message field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


