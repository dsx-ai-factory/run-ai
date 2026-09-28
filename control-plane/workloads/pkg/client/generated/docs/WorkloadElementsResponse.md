# WorkloadElementsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** | Offset for the next page of results. Present only when there are more elements beyond the current page (i.e., the number of returned elements equals the limit).  | [optional] 
**Elements** | [**[]WorkloadElement**](WorkloadElement.md) | A flat list of structure elements for this workload, ordered by hierarchy level. This page may contain fewer elements than are stored in the database — use the \&quot;next\&quot; field to retrieve additional pages. Use the /elements/count endpoint to retrieve total and stored counts.  | 

## Methods

### NewWorkloadElementsResponse

`func NewWorkloadElementsResponse(elements []WorkloadElement, ) *WorkloadElementsResponse`

NewWorkloadElementsResponse instantiates a new WorkloadElementsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadElementsResponseWithDefaults

`func NewWorkloadElementsResponseWithDefaults() *WorkloadElementsResponse`

NewWorkloadElementsResponseWithDefaults instantiates a new WorkloadElementsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *WorkloadElementsResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *WorkloadElementsResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *WorkloadElementsResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *WorkloadElementsResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetElements

`func (o *WorkloadElementsResponse) GetElements() []WorkloadElement`

GetElements returns the Elements field if non-nil, zero value otherwise.

### GetElementsOk

`func (o *WorkloadElementsResponse) GetElementsOk() (*[]WorkloadElement, bool)`

GetElementsOk returns a tuple with the Elements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetElements

`func (o *WorkloadElementsResponse) SetElements(v []WorkloadElement)`

SetElements sets Elements field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


