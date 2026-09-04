# ReportsBatchGetResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID of the batch operation. | 
**Succeeded** | Pointer to [**[]Report**](Report.md) | Reports that were found for the tenant. | [optional] 
**Failed** | Pointer to [**[]BatchResponseFailureItem**](BatchResponseFailureItem.md) | Report ids that could not be returned, with per-id error details. | [optional] 

## Methods

### NewReportsBatchGetResponse

`func NewReportsBatchGetResponse(id string, ) *ReportsBatchGetResponse`

NewReportsBatchGetResponse instantiates a new ReportsBatchGetResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReportsBatchGetResponseWithDefaults

`func NewReportsBatchGetResponseWithDefaults() *ReportsBatchGetResponse`

NewReportsBatchGetResponseWithDefaults instantiates a new ReportsBatchGetResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ReportsBatchGetResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ReportsBatchGetResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ReportsBatchGetResponse) SetId(v string)`

SetId sets Id field to given value.


### GetSucceeded

`func (o *ReportsBatchGetResponse) GetSucceeded() []Report`

GetSucceeded returns the Succeeded field if non-nil, zero value otherwise.

### GetSucceededOk

`func (o *ReportsBatchGetResponse) GetSucceededOk() (*[]Report, bool)`

GetSucceededOk returns a tuple with the Succeeded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSucceeded

`func (o *ReportsBatchGetResponse) SetSucceeded(v []Report)`

SetSucceeded sets Succeeded field to given value.

### HasSucceeded

`func (o *ReportsBatchGetResponse) HasSucceeded() bool`

HasSucceeded returns a boolean if a field has been set.

### GetFailed

`func (o *ReportsBatchGetResponse) GetFailed() []BatchResponseFailureItem`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *ReportsBatchGetResponse) GetFailedOk() (*[]BatchResponseFailureItem, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *ReportsBatchGetResponse) SetFailed(v []BatchResponseFailureItem)`

SetFailed sets Failed field to given value.

### HasFailed

`func (o *ReportsBatchGetResponse) HasFailed() bool`

HasFailed returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


