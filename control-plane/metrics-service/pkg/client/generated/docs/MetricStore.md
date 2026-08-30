# MetricStore

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantId** | **int32** | The id of the tenant. | 
**DatasourceId** | Pointer to **string** | Id of the datasource. Optional - when empty, the metrics-service either (a) generates a new datasource via the configured provider when metricStoreDetails is null, or (b) saves an external metric store with datasourceId&#x3D;\&quot;external\&quot; when metricStoreDetails is provided.  | [optional] 
**MetricStoreDetails** | Pointer to [**NullableMetricStoreDetails**](MetricStoreDetails.md) |  | [optional] 

## Methods

### NewMetricStore

`func NewMetricStore(tenantId int32, ) *MetricStore`

NewMetricStore instantiates a new MetricStore object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricStoreWithDefaults

`func NewMetricStoreWithDefaults() *MetricStore`

NewMetricStoreWithDefaults instantiates a new MetricStore object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantId

`func (o *MetricStore) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *MetricStore) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *MetricStore) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetDatasourceId

`func (o *MetricStore) GetDatasourceId() string`

GetDatasourceId returns the DatasourceId field if non-nil, zero value otherwise.

### GetDatasourceIdOk

`func (o *MetricStore) GetDatasourceIdOk() (*string, bool)`

GetDatasourceIdOk returns a tuple with the DatasourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasourceId

`func (o *MetricStore) SetDatasourceId(v string)`

SetDatasourceId sets DatasourceId field to given value.

### HasDatasourceId

`func (o *MetricStore) HasDatasourceId() bool`

HasDatasourceId returns a boolean if a field has been set.

### GetMetricStoreDetails

`func (o *MetricStore) GetMetricStoreDetails() MetricStoreDetails`

GetMetricStoreDetails returns the MetricStoreDetails field if non-nil, zero value otherwise.

### GetMetricStoreDetailsOk

`func (o *MetricStore) GetMetricStoreDetailsOk() (*MetricStoreDetails, bool)`

GetMetricStoreDetailsOk returns a tuple with the MetricStoreDetails field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricStoreDetails

`func (o *MetricStore) SetMetricStoreDetails(v MetricStoreDetails)`

SetMetricStoreDetails sets MetricStoreDetails field to given value.

### HasMetricStoreDetails

`func (o *MetricStore) HasMetricStoreDetails() bool`

HasMetricStoreDetails returns a boolean if a field has been set.

### SetMetricStoreDetailsNil

`func (o *MetricStore) SetMetricStoreDetailsNil(b bool)`

 SetMetricStoreDetailsNil sets the value for MetricStoreDetails to be an explicit nil

### UnsetMetricStoreDetails
`func (o *MetricStore) UnsetMetricStoreDetails()`

UnsetMetricStoreDetails ensures that no value is present for MetricStoreDetails, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


