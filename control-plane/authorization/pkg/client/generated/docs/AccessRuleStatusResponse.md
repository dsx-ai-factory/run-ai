# AccessRuleStatusResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to [**AccessRulePhase**](AccessRulePhase.md) |  | [optional] 
**Message** | Pointer to **string** | Message for status of Access Rule in the cluster | [optional] 
**Details** | Pointer to **string** | Details for status of Access Rule in the clusters | [optional] 

## Methods

### NewAccessRuleStatusResponse

`func NewAccessRuleStatusResponse() *AccessRuleStatusResponse`

NewAccessRuleStatusResponse instantiates a new AccessRuleStatusResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleStatusResponseWithDefaults

`func NewAccessRuleStatusResponseWithDefaults() *AccessRuleStatusResponse`

NewAccessRuleStatusResponseWithDefaults instantiates a new AccessRuleStatusResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *AccessRuleStatusResponse) GetPhase() AccessRulePhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *AccessRuleStatusResponse) GetPhaseOk() (*AccessRulePhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *AccessRuleStatusResponse) SetPhase(v AccessRulePhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *AccessRuleStatusResponse) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetMessage

`func (o *AccessRuleStatusResponse) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AccessRuleStatusResponse) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AccessRuleStatusResponse) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *AccessRuleStatusResponse) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetDetails

`func (o *AccessRuleStatusResponse) GetDetails() string`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *AccessRuleStatusResponse) GetDetailsOk() (*string, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *AccessRuleStatusResponse) SetDetails(v string)`

SetDetails sets Details field to given value.

### HasDetails

`func (o *AccessRuleStatusResponse) HasDetails() bool`

HasDetails returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


