# MetricStoreEndpointAuth

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to [**MetricStoreEndpointAuthType**](MetricStoreEndpointAuthType.md) |  | [optional] 
**Basic** | Pointer to [**NullableMetricStoreEndpointAuthBasic**](MetricStoreEndpointAuthBasic.md) |  | [optional] 

## Methods

### NewMetricStoreEndpointAuth

`func NewMetricStoreEndpointAuth() *MetricStoreEndpointAuth`

NewMetricStoreEndpointAuth instantiates a new MetricStoreEndpointAuth object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricStoreEndpointAuthWithDefaults

`func NewMetricStoreEndpointAuthWithDefaults() *MetricStoreEndpointAuth`

NewMetricStoreEndpointAuthWithDefaults instantiates a new MetricStoreEndpointAuth object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *MetricStoreEndpointAuth) GetType() MetricStoreEndpointAuthType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *MetricStoreEndpointAuth) GetTypeOk() (*MetricStoreEndpointAuthType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *MetricStoreEndpointAuth) SetType(v MetricStoreEndpointAuthType)`

SetType sets Type field to given value.

### HasType

`func (o *MetricStoreEndpointAuth) HasType() bool`

HasType returns a boolean if a field has been set.

### GetBasic

`func (o *MetricStoreEndpointAuth) GetBasic() MetricStoreEndpointAuthBasic`

GetBasic returns the Basic field if non-nil, zero value otherwise.

### GetBasicOk

`func (o *MetricStoreEndpointAuth) GetBasicOk() (*MetricStoreEndpointAuthBasic, bool)`

GetBasicOk returns a tuple with the Basic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBasic

`func (o *MetricStoreEndpointAuth) SetBasic(v MetricStoreEndpointAuthBasic)`

SetBasic sets Basic field to given value.

### HasBasic

`func (o *MetricStoreEndpointAuth) HasBasic() bool`

HasBasic returns a boolean if a field has been set.

### SetBasicNil

`func (o *MetricStoreEndpointAuth) SetBasicNil(b bool)`

 SetBasicNil sets the value for Basic to be an explicit nil

### UnsetBasic
`func (o *MetricStoreEndpointAuth) UnsetBasic()`

UnsetBasic ensures that no value is present for Basic, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


