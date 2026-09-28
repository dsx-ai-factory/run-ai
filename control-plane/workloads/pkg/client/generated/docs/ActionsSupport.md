# ActionsSupport

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Delete** | Pointer to **bool** |  | [optional] 
**Suspend** | Pointer to **bool** | This permission includes resume as well as suspend. | [optional] 

## Methods

### NewActionsSupport

`func NewActionsSupport() *ActionsSupport`

NewActionsSupport instantiates a new ActionsSupport object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewActionsSupportWithDefaults

`func NewActionsSupportWithDefaults() *ActionsSupport`

NewActionsSupportWithDefaults instantiates a new ActionsSupport object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDelete

`func (o *ActionsSupport) GetDelete() bool`

GetDelete returns the Delete field if non-nil, zero value otherwise.

### GetDeleteOk

`func (o *ActionsSupport) GetDeleteOk() (*bool, bool)`

GetDeleteOk returns a tuple with the Delete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDelete

`func (o *ActionsSupport) SetDelete(v bool)`

SetDelete sets Delete field to given value.

### HasDelete

`func (o *ActionsSupport) HasDelete() bool`

HasDelete returns a boolean if a field has been set.

### GetSuspend

`func (o *ActionsSupport) GetSuspend() bool`

GetSuspend returns the Suspend field if non-nil, zero value otherwise.

### GetSuspendOk

`func (o *ActionsSupport) GetSuspendOk() (*bool, bool)`

GetSuspendOk returns a tuple with the Suspend field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuspend

`func (o *ActionsSupport) SetSuspend(v bool)`

SetSuspend sets Suspend field to given value.

### HasSuspend

`func (o *ActionsSupport) HasSuspend() bool`

HasSuspend returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


