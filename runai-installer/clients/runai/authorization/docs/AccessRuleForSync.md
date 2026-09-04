# AccessRuleForSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | [**AccessRuleClusterPhase**](AccessRuleClusterPhase.md) |  | 
**Message** | **string** | Message for status of Access Rule in the cluster | 
**AccessRuleId** | **int32** | The id of the access rule to retrieve | 

## Methods

### NewAccessRuleForSync

`func NewAccessRuleForSync(phase AccessRuleClusterPhase, message string, accessRuleId int32, ) *AccessRuleForSync`

NewAccessRuleForSync instantiates a new AccessRuleForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleForSyncWithDefaults

`func NewAccessRuleForSyncWithDefaults() *AccessRuleForSync`

NewAccessRuleForSyncWithDefaults instantiates a new AccessRuleForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *AccessRuleForSync) GetPhase() AccessRuleClusterPhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *AccessRuleForSync) GetPhaseOk() (*AccessRuleClusterPhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *AccessRuleForSync) SetPhase(v AccessRuleClusterPhase)`

SetPhase sets Phase field to given value.


### GetMessage

`func (o *AccessRuleForSync) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AccessRuleForSync) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AccessRuleForSync) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetAccessRuleId

`func (o *AccessRuleForSync) GetAccessRuleId() int32`

GetAccessRuleId returns the AccessRuleId field if non-nil, zero value otherwise.

### GetAccessRuleIdOk

`func (o *AccessRuleForSync) GetAccessRuleIdOk() (*int32, bool)`

GetAccessRuleIdOk returns a tuple with the AccessRuleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessRuleId

`func (o *AccessRuleForSync) SetAccessRuleId(v int32)`

SetAccessRuleId sets AccessRuleId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


