# MetricStoreDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | [**MetricStoreType**](MetricStoreType.md) |  | 
**Read** | [**MetricStoreEndpoint**](MetricStoreEndpoint.md) |  | 
**Write** | [**MetricStoreEndpoint**](MetricStoreEndpoint.md) |  | 

## Methods

### NewMetricStoreDetails

`func NewMetricStoreDetails(type_ MetricStoreType, read MetricStoreEndpoint, write MetricStoreEndpoint, ) *MetricStoreDetails`

NewMetricStoreDetails instantiates a new MetricStoreDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricStoreDetailsWithDefaults

`func NewMetricStoreDetailsWithDefaults() *MetricStoreDetails`

NewMetricStoreDetailsWithDefaults instantiates a new MetricStoreDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *MetricStoreDetails) GetType() MetricStoreType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *MetricStoreDetails) GetTypeOk() (*MetricStoreType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *MetricStoreDetails) SetType(v MetricStoreType)`

SetType sets Type field to given value.


### GetRead

`func (o *MetricStoreDetails) GetRead() MetricStoreEndpoint`

GetRead returns the Read field if non-nil, zero value otherwise.

### GetReadOk

`func (o *MetricStoreDetails) GetReadOk() (*MetricStoreEndpoint, bool)`

GetReadOk returns a tuple with the Read field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRead

`func (o *MetricStoreDetails) SetRead(v MetricStoreEndpoint)`

SetRead sets Read field to given value.


### GetWrite

`func (o *MetricStoreDetails) GetWrite() MetricStoreEndpoint`

GetWrite returns the Write field if non-nil, zero value otherwise.

### GetWriteOk

`func (o *MetricStoreDetails) GetWriteOk() (*MetricStoreEndpoint, bool)`

GetWriteOk returns a tuple with the Write field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWrite

`func (o *MetricStoreDetails) SetWrite(v MetricStoreEndpoint)`

SetWrite sets Write field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


