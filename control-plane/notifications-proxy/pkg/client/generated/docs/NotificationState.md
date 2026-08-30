# NotificationState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | Pointer to [**NotificationStateStatus**](NotificationStateStatus.md) |  | [optional] 
**Slack** | Pointer to [**NotificationStateStatus**](NotificationStateStatus.md) |  | [optional] 

## Methods

### NewNotificationState

`func NewNotificationState() *NotificationState`

NewNotificationState instantiates a new NotificationState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNotificationStateWithDefaults

`func NewNotificationStateWithDefaults() *NotificationState`

NewNotificationStateWithDefaults instantiates a new NotificationState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *NotificationState) GetEmail() NotificationStateStatus`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *NotificationState) GetEmailOk() (*NotificationStateStatus, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *NotificationState) SetEmail(v NotificationStateStatus)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *NotificationState) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetSlack

`func (o *NotificationState) GetSlack() NotificationStateStatus`

GetSlack returns the Slack field if non-nil, zero value otherwise.

### GetSlackOk

`func (o *NotificationState) GetSlackOk() (*NotificationStateStatus, bool)`

GetSlackOk returns a tuple with the Slack field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSlack

`func (o *NotificationState) SetSlack(v NotificationStateStatus)`

SetSlack sets Slack field to given value.

### HasSlack

`func (o *NotificationState) HasSlack() bool`

HasSlack returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


