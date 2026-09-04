# GetWorkloadHistory200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Records** | [**[]HistoryRecord**](HistoryRecord.md) |  | 

## Methods

### NewGetWorkloadHistory200Response

`func NewGetWorkloadHistory200Response(records []HistoryRecord, ) *GetWorkloadHistory200Response`

NewGetWorkloadHistory200Response instantiates a new GetWorkloadHistory200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetWorkloadHistory200ResponseWithDefaults

`func NewGetWorkloadHistory200ResponseWithDefaults() *GetWorkloadHistory200Response`

NewGetWorkloadHistory200ResponseWithDefaults instantiates a new GetWorkloadHistory200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetWorkloadHistory200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetWorkloadHistory200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetWorkloadHistory200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetWorkloadHistory200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetRecords

`func (o *GetWorkloadHistory200Response) GetRecords() []HistoryRecord`

GetRecords returns the Records field if non-nil, zero value otherwise.

### GetRecordsOk

`func (o *GetWorkloadHistory200Response) GetRecordsOk() (*[]HistoryRecord, bool)`

GetRecordsOk returns a tuple with the Records field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecords

`func (o *GetWorkloadHistory200Response) SetRecords(v []HistoryRecord)`

SetRecords sets Records field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


