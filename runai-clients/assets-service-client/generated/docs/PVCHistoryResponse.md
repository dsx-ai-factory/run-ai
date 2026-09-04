# PVCHistoryResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Records** | [**[]HistoryRecord**](HistoryRecord.md) |  | 

## Methods

### NewPVCHistoryResponse

`func NewPVCHistoryResponse(records []HistoryRecord, ) *PVCHistoryResponse`

NewPVCHistoryResponse instantiates a new PVCHistoryResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCHistoryResponseWithDefaults

`func NewPVCHistoryResponseWithDefaults() *PVCHistoryResponse`

NewPVCHistoryResponseWithDefaults instantiates a new PVCHistoryResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *PVCHistoryResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *PVCHistoryResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *PVCHistoryResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *PVCHistoryResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetRecords

`func (o *PVCHistoryResponse) GetRecords() []HistoryRecord`

GetRecords returns the Records field if non-nil, zero value otherwise.

### GetRecordsOk

`func (o *PVCHistoryResponse) GetRecordsOk() (*[]HistoryRecord, bool)`

GetRecordsOk returns a tuple with the Records field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecords

`func (o *PVCHistoryResponse) SetRecords(v []HistoryRecord)`

SetRecords sets Records field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


