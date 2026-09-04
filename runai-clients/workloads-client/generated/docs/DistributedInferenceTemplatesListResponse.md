# DistributedInferenceTemplatesListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | Pointer to [**[]DistributedInferenceTemplatesListResponseItem**](DistributedInferenceTemplatesListResponseItem.md) |  | [optional] 
**Pagination** | Pointer to [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Methods

### NewDistributedInferenceTemplatesListResponse

`func NewDistributedInferenceTemplatesListResponse() *DistributedInferenceTemplatesListResponse`

NewDistributedInferenceTemplatesListResponse instantiates a new DistributedInferenceTemplatesListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceTemplatesListResponseWithDefaults

`func NewDistributedInferenceTemplatesListResponseWithDefaults() *DistributedInferenceTemplatesListResponse`

NewDistributedInferenceTemplatesListResponseWithDefaults instantiates a new DistributedInferenceTemplatesListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *DistributedInferenceTemplatesListResponse) GetEntries() []DistributedInferenceTemplatesListResponseItem`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *DistributedInferenceTemplatesListResponse) GetEntriesOk() (*[]DistributedInferenceTemplatesListResponseItem, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *DistributedInferenceTemplatesListResponse) SetEntries(v []DistributedInferenceTemplatesListResponseItem)`

SetEntries sets Entries field to given value.

### HasEntries

`func (o *DistributedInferenceTemplatesListResponse) HasEntries() bool`

HasEntries returns a boolean if a field has been set.

### GetPagination

`func (o *DistributedInferenceTemplatesListResponse) GetPagination() PaginationInfo`

GetPagination returns the Pagination field if non-nil, zero value otherwise.

### GetPaginationOk

`func (o *DistributedInferenceTemplatesListResponse) GetPaginationOk() (*PaginationInfo, bool)`

GetPaginationOk returns a tuple with the Pagination field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPagination

`func (o *DistributedInferenceTemplatesListResponse) SetPagination(v PaginationInfo)`

SetPagination sets Pagination field to given value.

### HasPagination

`func (o *DistributedInferenceTemplatesListResponse) HasPagination() bool`

HasPagination returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


