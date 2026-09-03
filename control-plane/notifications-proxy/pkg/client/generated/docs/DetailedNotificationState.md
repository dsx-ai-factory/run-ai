# DetailedNotificationState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | Pointer to [**NotificationStateStatus**](NotificationStateStatus.md) |  | [optional] 
**Slack** | Pointer to [**DetailedNotificationStateSlack**](DetailedNotificationStateSlack.md) |  | [optional] 

## Methods

### NewDetailedNotificationState

`func NewDetailedNotificationState() *DetailedNotificationState`

NewDetailedNotificationState instantiates a new DetailedNotificationState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDetailedNotificationStateWithDefaults

`func NewDetailedNotificationStateWithDefaults() *DetailedNotificationState`

NewDetailedNotificationStateWithDefaults instantiates a new DetailedNotificationState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *DetailedNotificationState) GetEmail() NotificationStateStatus`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *DetailedNotificationState) GetEmailOk() (*NotificationStateStatus, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *DetailedNotificationState) SetEmail(v NotificationStateStatus)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *DetailedNotificationState) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetSlack

`func (o *DetailedNotificationState) GetSlack() DetailedNotificationStateSlack`

GetSlack returns the Slack field if non-nil, zero value otherwise.

### GetSlackOk

`func (o *DetailedNotificationState) GetSlackOk() (*DetailedNotificationStateSlack, bool)`

GetSlackOk returns a tuple with the Slack field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSlack

`func (o *DetailedNotificationState) SetSlack(v DetailedNotificationStateSlack)`

SetSlack sets Slack field to given value.

### HasSlack

`func (o *DetailedNotificationState) HasSlack() bool`

HasSlack returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


