# MetricStore

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Read** | Pointer to [**MetricStoreEndpoint**](MetricStoreEndpoint.md) |  | [optional] 
**Write** | Pointer to [**MetricStoreEndpoint**](MetricStoreEndpoint.md) |  | [optional] 

## Methods

### NewMetricStore

`func NewMetricStore() *MetricStore`

NewMetricStore instantiates a new MetricStore object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricStoreWithDefaults

`func NewMetricStoreWithDefaults() *MetricStore`

NewMetricStoreWithDefaults instantiates a new MetricStore object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRead

`func (o *MetricStore) GetRead() MetricStoreEndpoint`

GetRead returns the Read field if non-nil, zero value otherwise.

### GetReadOk

`func (o *MetricStore) GetReadOk() (*MetricStoreEndpoint, bool)`

GetReadOk returns a tuple with the Read field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRead

`func (o *MetricStore) SetRead(v MetricStoreEndpoint)`

SetRead sets Read field to given value.

### HasRead

`func (o *MetricStore) HasRead() bool`

HasRead returns a boolean if a field has been set.

### GetWrite

`func (o *MetricStore) GetWrite() MetricStoreEndpoint`

GetWrite returns the Write field if non-nil, zero value otherwise.

### GetWriteOk

`func (o *MetricStore) GetWriteOk() (*MetricStoreEndpoint, bool)`

GetWriteOk returns a tuple with the Write field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWrite

`func (o *MetricStore) SetWrite(v MetricStoreEndpoint)`

SetWrite sets Write field to given value.

### HasWrite

`func (o *MetricStore) HasWrite() bool`

HasWrite returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


