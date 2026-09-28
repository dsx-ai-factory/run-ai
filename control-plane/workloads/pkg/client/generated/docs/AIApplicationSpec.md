# AIApplicationSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChartRef** | [**AIApplicationSpecChartRef**](AIApplicationSpecChartRef.md) |  | 
**Values** | Pointer to **map[string]interface{}** | Inline Helm values used to override the Helm chart’s default configuration. These values are merged with the Helm chart’s &#x60;values.yaml&#x60;. This field is mutually exclusive with &#x60;valuesBase64&#x60;.  A complete string value can reference a personal credential using &#x60;USERCREDENTIAL:&lt;name&gt;&#x60;. Generic secret and Docker registry credentials use that form. NGC credentials require an explicit representation: &#x60;USERCREDENTIAL:&lt;name&gt;,generic-secret&#x60; or &#x60;USERCREDENTIAL:&lt;name&gt;,docker-registry&#x60;. Embedded or partial placeholders are not supported. The resolved value is a Kubernetes Secret name, never credential plaintext.  | [optional] 
**ValuesBase64** | Pointer to **string** | Base64-encoded Helm values used to override the Helm chart’s default configuration. This field is mutually exclusive with &#x60;values&#x60;. When the decoded content contains personal credential references, it must be a YAML object.  String values support the same exact personal credential placeholder grammar as &#x60;values&#x60;: &#x60;USERCREDENTIAL:&lt;name&gt;&#x60; for generic secret and Docker registry credentials, or &#x60;USERCREDENTIAL:&lt;name&gt;,generic-secret&#x60; and &#x60;USERCREDENTIAL:&lt;name&gt;,docker-registry&#x60; for NGC credentials. Embedded or partial placeholders are not supported.  | [optional] 
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


