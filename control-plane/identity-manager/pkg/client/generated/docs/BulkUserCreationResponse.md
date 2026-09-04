# BulkUserCreationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Results** | [**[]BulkUserCreationResponseResultsInner**](BulkUserCreationResponseResultsInner.md) |  | 
**TotalRequested** | Pointer to **int32** | Total number of users requested for creation | [optional] 
**TotalSuccessful** | Pointer to **int32** | Number of users successfully created | [optional] 
**TotalFailed** | Pointer to **int32** | Number of users that failed to be created | [optional] 

## Methods

### NewBulkUserCreationResponse

`func NewBulkUserCreationResponse(results []BulkUserCreationResponseResultsInner, ) *BulkUserCreationResponse`

NewBulkUserCreationResponse instantiates a new BulkUserCreationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkUserCreationResponseWithDefaults

`func NewBulkUserCreationResponseWithDefaults() *BulkUserCreationResponse`

NewBulkUserCreationResponseWithDefaults instantiates a new BulkUserCreationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResults

`func (o *BulkUserCreationResponse) GetResults() []BulkUserCreationResponseResultsInner`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *BulkUserCreationResponse) GetResultsOk() (*[]BulkUserCreationResponseResultsInner, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *BulkUserCreationResponse) SetResults(v []BulkUserCreationResponseResultsInner)`

SetResults sets Results field to given value.


### GetTotalRequested

`func (o *BulkUserCreationResponse) GetTotalRequested() int32`

GetTotalRequested returns the TotalRequested field if non-nil, zero value otherwise.

### GetTotalRequestedOk

`func (o *BulkUserCreationResponse) GetTotalRequestedOk() (*int32, bool)`

GetTotalRequestedOk returns a tuple with the TotalRequested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequested

`func (o *BulkUserCreationResponse) SetTotalRequested(v int32)`

SetTotalRequested sets TotalRequested field to given value.

### HasTotalRequested

`func (o *BulkUserCreationResponse) HasTotalRequested() bool`

HasTotalRequested returns a boolean if a field has been set.

### GetTotalSuccessful

`func (o *BulkUserCreationResponse) GetTotalSuccessful() int32`

GetTotalSuccessful returns the TotalSuccessful field if non-nil, zero value otherwise.

### GetTotalSuccessfulOk

`func (o *BulkUserCreationResponse) GetTotalSuccessfulOk() (*int32, bool)`

GetTotalSuccessfulOk returns a tuple with the TotalSuccessful field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalSuccessful

`func (o *BulkUserCreationResponse) SetTotalSuccessful(v int32)`

SetTotalSuccessful sets TotalSuccessful field to given value.

### HasTotalSuccessful

`func (o *BulkUserCreationResponse) HasTotalSuccessful() bool`

HasTotalSuccessful returns a boolean if a field has been set.

### GetTotalFailed

`func (o *BulkUserCreationResponse) GetTotalFailed() int32`

GetTotalFailed returns the TotalFailed field if non-nil, zero value otherwise.

### GetTotalFailedOk

`func (o *BulkUserCreationResponse) GetTotalFailedOk() (*int32, bool)`

GetTotalFailedOk returns a tuple with the TotalFailed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalFailed

`func (o *BulkUserCreationResponse) SetTotalFailed(v int32)`

SetTotalFailed sets TotalFailed field to given value.

### HasTotalFailed

`func (o *BulkUserCreationResponse) HasTotalFailed() bool`

HasTotalFailed returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


