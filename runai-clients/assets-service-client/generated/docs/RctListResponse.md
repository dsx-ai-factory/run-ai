# RctListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** | Offset for the next page of results. Present only when there are more entries beyond the current page (i.e., the number of returned entries equals the limit).  | [optional] 
**Entries** | [**[]Rct**](Rct.md) |  | 

## Methods

### NewRctListResponse

`func NewRctListResponse(entries []Rct, ) *RctListResponse`

NewRctListResponse instantiates a new RctListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRctListResponseWithDefaults

`func NewRctListResponseWithDefaults() *RctListResponse`

NewRctListResponseWithDefaults instantiates a new RctListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *RctListResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *RctListResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *RctListResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *RctListResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetEntries

`func (o *RctListResponse) GetEntries() []Rct`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *RctListResponse) GetEntriesOk() (*[]Rct, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *RctListResponse) SetEntries(v []Rct)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


