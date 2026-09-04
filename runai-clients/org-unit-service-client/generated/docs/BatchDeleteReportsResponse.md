# BatchDeleteReportsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique ID of the batch operation. | 
**Accepted** | **[]string** | IDs of reports accepted for deletion. | 
**Failed** | [**[]BatchResponseFailureItem**](BatchResponseFailureItem.md) | Reports that could not be deleted, with per-report error details. | 
**TotalRequested** | **int32** | Total number of items requested in the batch. | 
**TotalAccepted** | **int32** | Number of items accepted for deletion. | 
**TotalFailed** | **int32** | Number of items that failed. | 

## Methods

### NewBatchDeleteReportsResponse

`func NewBatchDeleteReportsResponse(id string, accepted []string, failed []BatchResponseFailureItem, totalRequested int32, totalAccepted int32, totalFailed int32, ) *BatchDeleteReportsResponse`

NewBatchDeleteReportsResponse instantiates a new BatchDeleteReportsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBatchDeleteReportsResponseWithDefaults

`func NewBatchDeleteReportsResponseWithDefaults() *BatchDeleteReportsResponse`

NewBatchDeleteReportsResponseWithDefaults instantiates a new BatchDeleteReportsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *BatchDeleteReportsResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *BatchDeleteReportsResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *BatchDeleteReportsResponse) SetId(v string)`

SetId sets Id field to given value.


### GetAccepted

`func (o *BatchDeleteReportsResponse) GetAccepted() []string`

GetAccepted returns the Accepted field if non-nil, zero value otherwise.

### GetAcceptedOk

`func (o *BatchDeleteReportsResponse) GetAcceptedOk() (*[]string, bool)`

GetAcceptedOk returns a tuple with the Accepted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccepted

`func (o *BatchDeleteReportsResponse) SetAccepted(v []string)`

SetAccepted sets Accepted field to given value.


### GetFailed

`func (o *BatchDeleteReportsResponse) GetFailed() []BatchResponseFailureItem`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *BatchDeleteReportsResponse) GetFailedOk() (*[]BatchResponseFailureItem, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *BatchDeleteReportsResponse) SetFailed(v []BatchResponseFailureItem)`

SetFailed sets Failed field to given value.


### GetTotalRequested

`func (o *BatchDeleteReportsResponse) GetTotalRequested() int32`

GetTotalRequested returns the TotalRequested field if non-nil, zero value otherwise.

### GetTotalRequestedOk

`func (o *BatchDeleteReportsResponse) GetTotalRequestedOk() (*int32, bool)`

GetTotalRequestedOk returns a tuple with the TotalRequested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequested

`func (o *BatchDeleteReportsResponse) SetTotalRequested(v int32)`

SetTotalRequested sets TotalRequested field to given value.


### GetTotalAccepted

`func (o *BatchDeleteReportsResponse) GetTotalAccepted() int32`

GetTotalAccepted returns the TotalAccepted field if non-nil, zero value otherwise.

### GetTotalAcceptedOk

`func (o *BatchDeleteReportsResponse) GetTotalAcceptedOk() (*int32, bool)`

GetTotalAcceptedOk returns a tuple with the TotalAccepted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalAccepted

`func (o *BatchDeleteReportsResponse) SetTotalAccepted(v int32)`

SetTotalAccepted sets TotalAccepted field to given value.


### GetTotalFailed

`func (o *BatchDeleteReportsResponse) GetTotalFailed() int32`

GetTotalFailed returns the TotalFailed field if non-nil, zero value otherwise.

### GetTotalFailedOk

`func (o *BatchDeleteReportsResponse) GetTotalFailedOk() (*int32, bool)`

GetTotalFailedOk returns a tuple with the TotalFailed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalFailed

`func (o *BatchDeleteReportsResponse) SetTotalFailed(v int32)`

SetTotalFailed sets TotalFailed field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


