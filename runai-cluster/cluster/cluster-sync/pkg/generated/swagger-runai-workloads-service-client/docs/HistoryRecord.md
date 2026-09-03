# HistoryRecord

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**HistoryRecordMeta**](HistoryRecordMeta.md) |  | 
**Spec** | [**HistoryRecordSpec**](HistoryRecordSpec.md) |  | 

## Methods

### NewHistoryRecord

`func NewHistoryRecord(meta HistoryRecordMeta, spec HistoryRecordSpec, ) *HistoryRecord`

NewHistoryRecord instantiates a new HistoryRecord object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHistoryRecordWithDefaults

`func NewHistoryRecordWithDefaults() *HistoryRecord`

NewHistoryRecordWithDefaults instantiates a new HistoryRecord object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *HistoryRecord) GetMeta() HistoryRecordMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *HistoryRecord) GetMetaOk() (*HistoryRecordMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *HistoryRecord) SetMeta(v HistoryRecordMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *HistoryRecord) GetSpec() HistoryRecordSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *HistoryRecord) GetSpecOk() (*HistoryRecordSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *HistoryRecord) SetSpec(v HistoryRecordSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


