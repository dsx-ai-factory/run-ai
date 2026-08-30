# AIApplicationSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChartRef** | [**AIApplicationSpecChartRef**](AIApplicationSpecChartRef.md) |  | 
**Values** | Pointer to **map[string]interface{}** | Inline Helm values used to override the Helm chart’s default configuration. These values are merged with the Helm chart’s &#x60;values.yaml&#x60;. This field is mutually exclusive with &#x60;valuesBase64&#x60;. | [optional] 
**ValuesBase64** | Pointer to **string** | Base64-encoded Helm values used to override the Helm chart’s default configuration. This field is mutually exclusive with &#x60;values&#x60;. | [optional] 
**Auth** | Pointer to [**AIApplicationAuth**](AIApplicationAuth.md) |  | [optional] 

## Methods

### NewAIApplicationSpec

`func NewAIApplicationSpec(chartRef AIApplicationSpecChartRef, ) *AIApplicationSpec`

NewAIApplicationSpec instantiates a new AIApplicationSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationSpecWithDefaults

`func NewAIApplicationSpecWithDefaults() *AIApplicationSpec`

NewAIApplicationSpecWithDefaults instantiates a new AIApplicationSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChartRef

`func (o *AIApplicationSpec) GetChartRef() AIApplicationSpecChartRef`

GetChartRef returns the ChartRef field if non-nil, zero value otherwise.

### GetChartRefOk

`func (o *AIApplicationSpec) GetChartRefOk() (*AIApplicationSpecChartRef, bool)`

GetChartRefOk returns a tuple with the ChartRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartRef

`func (o *AIApplicationSpec) SetChartRef(v AIApplicationSpecChartRef)`

SetChartRef sets ChartRef field to given value.


### GetValues

`func (o *AIApplicationSpec) GetValues() map[string]interface{}`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *AIApplicationSpec) GetValuesOk() (*map[string]interface{}, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *AIApplicationSpec) SetValues(v map[string]interface{})`

SetValues sets Values field to given value.

### HasValues

`func (o *AIApplicationSpec) HasValues() bool`

HasValues returns a boolean if a field has been set.

### GetValuesBase64

`func (o *AIApplicationSpec) GetValuesBase64() string`

GetValuesBase64 returns the ValuesBase64 field if non-nil, zero value otherwise.

### GetValuesBase64Ok

`func (o *AIApplicationSpec) GetValuesBase64Ok() (*string, bool)`

GetValuesBase64Ok returns a tuple with the ValuesBase64 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValuesBase64

`func (o *AIApplicationSpec) SetValuesBase64(v string)`

SetValuesBase64 sets ValuesBase64 field to given value.

### HasValuesBase64

`func (o *AIApplicationSpec) HasValuesBase64() bool`

HasValuesBase64 returns a boolean if a field has been set.

### GetAuth

`func (o *AIApplicationSpec) GetAuth() AIApplicationAuth`

GetAuth returns the Auth field if non-nil, zero value otherwise.

### GetAuthOk

`func (o *AIApplicationSpec) GetAuthOk() (*AIApplicationAuth, bool)`

GetAuthOk returns a tuple with the Auth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuth

`func (o *AIApplicationSpec) SetAuth(v AIApplicationAuth)`

SetAuth sets Auth field to given value.

### HasAuth

`func (o *AIApplicationSpec) HasAuth() bool`

HasAuth returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


