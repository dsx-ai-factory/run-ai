# AccessRulesBatchFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ids** | **[]string** | List of ids to perform the action on. | 
**Action** | [**AccessRulesBatchAction**](AccessRulesBatchAction.md) |  | 

## Methods

### NewAccessRulesBatchFields

`func NewAccessRulesBatchFields(ids []string, action AccessRulesBatchAction, ) *AccessRulesBatchFields`

NewAccessRulesBatchFields instantiates a new AccessRulesBatchFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRulesBatchFieldsWithDefaults

`func NewAccessRulesBatchFieldsWithDefaults() *AccessRulesBatchFields`

NewAccessRulesBatchFieldsWithDefaults instantiates a new AccessRulesBatchFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIds

`func (o *AccessRulesBatchFields) GetIds() []string`

GetIds returns the Ids field if non-nil, zero value otherwise.

### GetIdsOk

`func (o *AccessRulesBatchFields) GetIdsOk() (*[]string, bool)`

GetIdsOk returns a tuple with the Ids field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIds

`func (o *AccessRulesBatchFields) SetIds(v []string)`

SetIds sets Ids field to given value.


### GetAction

`func (o *AccessRulesBatchFields) GetAction() AccessRulesBatchAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *AccessRulesBatchFields) GetActionOk() (*AccessRulesBatchAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *AccessRulesBatchFields) SetAction(v AccessRulesBatchAction)`

SetAction sets Action field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


