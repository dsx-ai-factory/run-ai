# TemplatesListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | Pointer to [**[]SupersetTemplate**](SupersetTemplate.md) |  | [optional] 
**Pagination** | Pointer to [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Methods

### NewTemplatesListResponse

`func NewTemplatesListResponse() *TemplatesListResponse`

NewTemplatesListResponse instantiates a new TemplatesListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTemplatesListResponseWithDefaults

`func NewTemplatesListResponseWithDefaults() *TemplatesListResponse`

NewTemplatesListResponseWithDefaults instantiates a new TemplatesListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *TemplatesListResponse) GetEntries() []SupersetTemplate`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *TemplatesListResponse) GetEntriesOk() (*[]SupersetTemplate, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *TemplatesListResponse) SetEntries(v []SupersetTemplate)`

SetEntries sets Entries field to given value.

### HasEntries

`func (o *TemplatesListResponse) HasEntries() bool`

HasEntries returns a boolean if a field has been set.

### GetPagination

`func (o *TemplatesListResponse) GetPagination() PaginationInfo`

GetPagination returns the Pagination field if non-nil, zero value otherwise.

### GetPaginationOk

`func (o *TemplatesListResponse) GetPaginationOk() (*PaginationInfo, bool)`

GetPaginationOk returns a tuple with the Pagination field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPagination

`func (o *TemplatesListResponse) SetPagination(v PaginationInfo)`

SetPagination sets Pagination field to given value.

### HasPagination

`func (o *TemplatesListResponse) HasPagination() bool`

HasPagination returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


