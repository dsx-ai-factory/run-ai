# AccessRuleStatusRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | **string** | Phase of Access Rule in the cluster | 
**Message** | **string** | Message for status of Access Rule in the cluster | 

## Methods

### NewAccessRuleStatusRequest

`func NewAccessRuleStatusRequest(phase string, message string, ) *AccessRuleStatusRequest`

NewAccessRuleStatusRequest instantiates a new AccessRuleStatusRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleStatusRequestWithDefaults

`func NewAccessRuleStatusRequestWithDefaults() *AccessRuleStatusRequest`

NewAccessRuleStatusRequestWithDefaults instantiates a new AccessRuleStatusRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *AccessRuleStatusRequest) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *AccessRuleStatusRequest) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *AccessRuleStatusRequest) SetPhase(v string)`

SetPhase sets Phase field to given value.


### GetMessage

`func (o *AccessRuleStatusRequest) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AccessRuleStatusRequest) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AccessRuleStatusRequest) SetMessage(v string)`

SetMessage sets Message field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


