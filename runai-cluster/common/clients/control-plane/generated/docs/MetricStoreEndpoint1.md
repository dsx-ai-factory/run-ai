# MetricStoreEndpoint1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | **string** |  | 
**Auth** | Pointer to [**NullableMetricStoreEndpoint1Auth**](MetricStoreEndpoint1Auth.md) |  | [optional] 
**Headers** | Pointer to **map[string]string** |  | [optional] 

## Methods

### NewMetricStoreEndpoint1

`func NewMetricStoreEndpoint1(url string, ) *MetricStoreEndpoint1`

NewMetricStoreEndpoint1 instantiates a new MetricStoreEndpoint1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricStoreEndpoint1WithDefaults

`func NewMetricStoreEndpoint1WithDefaults() *MetricStoreEndpoint1`

NewMetricStoreEndpoint1WithDefaults instantiates a new MetricStoreEndpoint1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *MetricStoreEndpoint1) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *MetricStoreEndpoint1) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *MetricStoreEndpoint1) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetAuth

`func (o *MetricStoreEndpoint1) GetAuth() MetricStoreEndpoint1Auth`

GetAuth returns the Auth field if non-nil, zero value otherwise.

### GetAuthOk

`func (o *MetricStoreEndpoint1) GetAuthOk() (*MetricStoreEndpoint1Auth, bool)`

GetAuthOk returns a tuple with the Auth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuth

`func (o *MetricStoreEndpoint1) SetAuth(v MetricStoreEndpoint1Auth)`

SetAuth sets Auth field to given value.

### HasAuth

`func (o *MetricStoreEndpoint1) HasAuth() bool`

HasAuth returns a boolean if a field has been set.

### SetAuthNil

`func (o *MetricStoreEndpoint1) SetAuthNil(b bool)`

 SetAuthNil sets the value for Auth to be an explicit nil

### UnsetAuth
`func (o *MetricStoreEndpoint1) UnsetAuth()`

UnsetAuth ensures that no value is present for Auth, not even an explicit nil
### GetHeaders

`func (o *MetricStoreEndpoint1) GetHeaders() map[string]string`

GetHeaders returns the Headers field if non-nil, zero value otherwise.

### GetHeadersOk

`func (o *MetricStoreEndpoint1) GetHeadersOk() (*map[string]string, bool)`

GetHeadersOk returns a tuple with the Headers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeaders

`func (o *MetricStoreEndpoint1) SetHeaders(v map[string]string)`

SetHeaders sets Headers field to given value.

### HasHeaders

`func (o *MetricStoreEndpoint1) HasHeaders() bool`

HasHeaders returns a boolean if a field has been set.

### SetHeadersNil

`func (o *MetricStoreEndpoint1) SetHeadersNil(b bool)`

 SetHeadersNil sets the value for Headers to be an explicit nil

### UnsetHeaders
`func (o *MetricStoreEndpoint1) UnsetHeaders()`

UnsetHeaders ensures that no value is present for Headers, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


