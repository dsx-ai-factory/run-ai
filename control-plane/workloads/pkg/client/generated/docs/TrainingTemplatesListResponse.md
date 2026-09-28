# TrainingTemplatesListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | Pointer to [**[]TrainingTemplatesListResponseItem**](TrainingTemplatesListResponseItem.md) |  | [optional] 
**Pagination** | Pointer to [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Methods

### NewTrainingTemplatesListResponse

`func NewTrainingTemplatesListResponse() *TrainingTemplatesListResponse`

NewTrainingTemplatesListResponse instantiates a new TrainingTemplatesListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingTemplatesListResponseWithDefaults

`func NewTrainingTemplatesListResponseWithDefaults() *TrainingTemplatesListResponse`

NewTrainingTemplatesListResponseWithDefaults instantiates a new TrainingTemplatesListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *TrainingTemplatesListResponse) GetEntries() []TrainingTemplatesListResponseItem`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *TrainingTemplatesListResponse) GetEntriesOk() (*[]TrainingTemplatesListResponseItem, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *TrainingTemplatesListResponse) SetEntries(v []TrainingTemplatesListResponseItem)`

SetEntries sets Entries field to given value.

### HasEntries

`func (o *TrainingTemplatesListResponse) HasEntries() bool`

HasEntries returns a boolean if a field has been set.

### GetPagination

`func (o *TrainingTemplatesListResponse) GetPagination() PaginationInfo`

GetPagination returns the Pagination field if non-nil, zero value otherwise.

### GetPaginationOk

`func (o *TrainingTemplatesListResponse) GetPaginationOk() (*PaginationInfo, bool)`

GetPaginationOk returns a tuple with the Pagination field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPagination

`func (o *TrainingTemplatesListResponse) SetPagination(v PaginationInfo)`

SetPagination sets Pagination field to given value.

### HasPagination

`func (o *TrainingTemplatesListResponse) HasPagination() bool`

HasPagination returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


