# MetricStoreEndpoint

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | **string** |  | 
**Auth** | Pointer to [**NullableMetricStoreEndpointAuth**](MetricStoreEndpointAuth.md) |  | [optional] 
**Headers** | Pointer to **map[string]string** |  | [optional] 

## Methods

### NewMetricStoreEndpoint

`func NewMetricStoreEndpoint(url string, ) *MetricStoreEndpoint`

NewMetricStoreEndpoint instantiates a new MetricStoreEndpoint object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricStoreEndpointWithDefaults

`func NewMetricStoreEndpointWithDefaults() *MetricStoreEndpoint`

NewMetricStoreEndpointWithDefaults instantiates a new MetricStoreEndpoint object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *MetricStoreEndpoint) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *MetricStoreEndpoint) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *MetricStoreEndpoint) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetAuth

`func (o *MetricStoreEndpoint) GetAuth() MetricStoreEndpointAuth`

GetAuth returns the Auth field if non-nil, zero value otherwise.

### GetAuthOk

`func (o *MetricStoreEndpoint) GetAuthOk() (*MetricStoreEndpointAuth, bool)`

GetAuthOk returns a tuple with the Auth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuth

`func (o *MetricStoreEndpoint) SetAuth(v MetricStoreEndpointAuth)`

SetAuth sets Auth field to given value.

### HasAuth

`func (o *MetricStoreEndpoint) HasAuth() bool`

HasAuth returns a boolean if a field has been set.

### SetAuthNil

`func (o *MetricStoreEndpoint) SetAuthNil(b bool)`

 SetAuthNil sets the value for Auth to be an explicit nil

### UnsetAuth
`func (o *MetricStoreEndpoint) UnsetAuth()`

UnsetAuth ensures that no value is present for Auth, not even an explicit nil
### GetHeaders

`func (o *MetricStoreEndpoint) GetHeaders() map[string]string`

GetHeaders returns the Headers field if non-nil, zero value otherwise.

### GetHeadersOk

`func (o *MetricStoreEndpoint) GetHeadersOk() (*map[string]string, bool)`

GetHeadersOk returns a tuple with the Headers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeaders

`func (o *MetricStoreEndpoint) SetHeaders(v map[string]string)`

SetHeaders sets Headers field to given value.

### HasHeaders

`func (o *MetricStoreEndpoint) HasHeaders() bool`

HasHeaders returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


