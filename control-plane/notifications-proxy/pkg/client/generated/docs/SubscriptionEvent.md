# SubscriptionEvent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Category** | **string** |  | 
**Types** | **[]string** |  | 

## Methods

### NewSubscriptionEvent

`func NewSubscriptionEvent(category string, types []string, ) *SubscriptionEvent`

NewSubscriptionEvent instantiates a new SubscriptionEvent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubscriptionEventWithDefaults

`func NewSubscriptionEventWithDefaults() *SubscriptionEvent`

NewSubscriptionEventWithDefaults instantiates a new SubscriptionEvent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategory

`func (o *SubscriptionEvent) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *SubscriptionEvent) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *SubscriptionEvent) SetCategory(v string)`

SetCategory sets Category field to given value.


### GetTypes

`func (o *SubscriptionEvent) GetTypes() []string`

GetTypes returns the Types field if non-nil, zero value otherwise.

### GetTypesOk

`func (o *SubscriptionEvent) GetTypesOk() (*[]string, bool)`

GetTypesOk returns a tuple with the Types field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypes

`func (o *SubscriptionEvent) SetTypes(v []string)`

SetTypes sets Types field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


