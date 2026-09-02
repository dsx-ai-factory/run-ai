# AccessRulesMigrateDeprecatedResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Succeeded** | [**AccessRulesMigrateDeprecatedResponseSucceeded**](AccessRulesMigrateDeprecatedResponseSucceeded.md) |  | 
**Failed** | [**[]AccessRulesMigrateDeprecatedFailureItem**](AccessRulesMigrateDeprecatedFailureItem.md) | Access rules that failed to migrate. Each item identifies the old access rule (still alive) and the failure reason. | 

## Methods

### NewAccessRulesMigrateDeprecatedResponse

`func NewAccessRulesMigrateDeprecatedResponse(succeeded AccessRulesMigrateDeprecatedResponseSucceeded, failed []AccessRulesMigrateDeprecatedFailureItem, ) *AccessRulesMigrateDeprecatedResponse`

NewAccessRulesMigrateDeprecatedResponse instantiates a new AccessRulesMigrateDeprecatedResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRulesMigrateDeprecatedResponseWithDefaults

`func NewAccessRulesMigrateDeprecatedResponseWithDefaults() *AccessRulesMigrateDeprecatedResponse`

NewAccessRulesMigrateDeprecatedResponseWithDefaults instantiates a new AccessRulesMigrateDeprecatedResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSucceeded

`func (o *AccessRulesMigrateDeprecatedResponse) GetSucceeded() AccessRulesMigrateDeprecatedResponseSucceeded`

GetSucceeded returns the Succeeded field if non-nil, zero value otherwise.

### GetSucceededOk

`func (o *AccessRulesMigrateDeprecatedResponse) GetSucceededOk() (*AccessRulesMigrateDeprecatedResponseSucceeded, bool)`

GetSucceededOk returns a tuple with the Succeeded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSucceeded

`func (o *AccessRulesMigrateDeprecatedResponse) SetSucceeded(v AccessRulesMigrateDeprecatedResponseSucceeded)`

SetSucceeded sets Succeeded field to given value.


### GetFailed

`func (o *AccessRulesMigrateDeprecatedResponse) GetFailed() []AccessRulesMigrateDeprecatedFailureItem`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *AccessRulesMigrateDeprecatedResponse) GetFailedOk() (*[]AccessRulesMigrateDeprecatedFailureItem, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *AccessRulesMigrateDeprecatedResponse) SetFailed(v []AccessRulesMigrateDeprecatedFailureItem)`

SetFailed sets Failed field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


