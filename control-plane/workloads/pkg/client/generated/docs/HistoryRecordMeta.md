# HistoryRecordMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreationTimestamp** | **time.Time** |  | 
**Type** | [**HistoryRecordType**](HistoryRecordType.md) |  | 

## Methods

### NewHistoryRecordMeta

`func NewHistoryRecordMeta(creationTimestamp time.Time, type_ HistoryRecordType, ) *HistoryRecordMeta`

NewHistoryRecordMeta instantiates a new HistoryRecordMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHistoryRecordMetaWithDefaults

`func NewHistoryRecordMetaWithDefaults() *HistoryRecordMeta`

NewHistoryRecordMetaWithDefaults instantiates a new HistoryRecordMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreationTimestamp

`func (o *HistoryRecordMeta) GetCreationTimestamp() time.Time`

GetCreationTimestamp returns the CreationTimestamp field if non-nil, zero value otherwise.

### GetCreationTimestampOk

`func (o *HistoryRecordMeta) GetCreationTimestampOk() (*time.Time, bool)`

GetCreationTimestampOk returns a tuple with the CreationTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreationTimestamp

`func (o *HistoryRecordMeta) SetCreationTimestamp(v time.Time)`

SetCreationTimestamp sets CreationTimestamp field to given value.


### GetType

`func (o *HistoryRecordMeta) GetType() HistoryRecordType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *HistoryRecordMeta) GetTypeOk() (*HistoryRecordType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *HistoryRecordMeta) SetType(v HistoryRecordType)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


