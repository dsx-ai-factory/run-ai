# AccessRuleDeletionBatchResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique ID of the batch operation. | 
**Succeeded** | Pointer to **[]string** | List of access rule IDs that were successfully deleted. | [optional] 
**Failed** | Pointer to [**[]AccessRuleDeletionBatchResponseFailureItem**](AccessRuleDeletionBatchResponseFailureItem.md) | Details of the access rule IDs that failed to delete. | [optional] 

## Methods

### NewAccessRuleDeletionBatchResponse

`func NewAccessRuleDeletionBatchResponse(id string, ) *AccessRuleDeletionBatchResponse`

NewAccessRuleDeletionBatchResponse instantiates a new AccessRuleDeletionBatchResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleDeletionBatchResponseWithDefaults

`func NewAccessRuleDeletionBatchResponseWithDefaults() *AccessRuleDeletionBatchResponse`

NewAccessRuleDeletionBatchResponseWithDefaults instantiates a new AccessRuleDeletionBatchResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AccessRuleDeletionBatchResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AccessRuleDeletionBatchResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AccessRuleDeletionBatchResponse) SetId(v string)`

SetId sets Id field to given value.


### GetSucceeded

`func (o *AccessRuleDeletionBatchResponse) GetSucceeded() []string`

GetSucceeded returns the Succeeded field if non-nil, zero value otherwise.

### GetSucceededOk

`func (o *AccessRuleDeletionBatchResponse) GetSucceededOk() (*[]string, bool)`

GetSucceededOk returns a tuple with the Succeeded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSucceeded

`func (o *AccessRuleDeletionBatchResponse) SetSucceeded(v []string)`

SetSucceeded sets Succeeded field to given value.

### HasSucceeded

`func (o *AccessRuleDeletionBatchResponse) HasSucceeded() bool`

HasSucceeded returns a boolean if a field has been set.

### GetFailed

`func (o *AccessRuleDeletionBatchResponse) GetFailed() []AccessRuleDeletionBatchResponseFailureItem`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *AccessRuleDeletionBatchResponse) GetFailedOk() (*[]AccessRuleDeletionBatchResponseFailureItem, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *AccessRuleDeletionBatchResponse) SetFailed(v []AccessRuleDeletionBatchResponseFailureItem)`

SetFailed sets Failed field to given value.

### HasFailed

`func (o *AccessRuleDeletionBatchResponse) HasFailed() bool`

HasFailed returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


