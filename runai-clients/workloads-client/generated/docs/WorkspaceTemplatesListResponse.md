# WorkspaceTemplatesListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | Pointer to [**[]WorkspaceTemplatesListResponseItem**](WorkspaceTemplatesListResponseItem.md) |  | [optional] 
**Pagination** | Pointer to [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Methods

### NewWorkspaceTemplatesListResponse

`func NewWorkspaceTemplatesListResponse() *WorkspaceTemplatesListResponse`

NewWorkspaceTemplatesListResponse instantiates a new WorkspaceTemplatesListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceTemplatesListResponseWithDefaults

`func NewWorkspaceTemplatesListResponseWithDefaults() *WorkspaceTemplatesListResponse`

NewWorkspaceTemplatesListResponseWithDefaults instantiates a new WorkspaceTemplatesListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *WorkspaceTemplatesListResponse) GetEntries() []WorkspaceTemplatesListResponseItem`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *WorkspaceTemplatesListResponse) GetEntriesOk() (*[]WorkspaceTemplatesListResponseItem, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *WorkspaceTemplatesListResponse) SetEntries(v []WorkspaceTemplatesListResponseItem)`

SetEntries sets Entries field to given value.

### HasEntries

`func (o *WorkspaceTemplatesListResponse) HasEntries() bool`

HasEntries returns a boolean if a field has been set.

### GetPagination

`func (o *WorkspaceTemplatesListResponse) GetPagination() PaginationInfo`

GetPagination returns the Pagination field if non-nil, zero value otherwise.

### GetPaginationOk

`func (o *WorkspaceTemplatesListResponse) GetPaginationOk() (*PaginationInfo, bool)`

GetPaginationOk returns a tuple with the Pagination field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPagination

`func (o *WorkspaceTemplatesListResponse) SetPagination(v PaginationInfo)`

SetPagination sets Pagination field to given value.

### HasPagination

`func (o *WorkspaceTemplatesListResponse) HasPagination() bool`

HasPagination returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


