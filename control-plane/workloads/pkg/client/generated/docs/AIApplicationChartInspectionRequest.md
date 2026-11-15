# AIApplicationChartInspectionRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChartRef** | [**AIApplicationSpecChartRef**](AIApplicationSpecChartRef.md) |  | 
**Auth** | Pointer to [**AIApplicationAuth**](AIApplicationAuth.md) |  | [optional] 

## Methods

### NewAIApplicationChartInspectionRequest

`func NewAIApplicationChartInspectionRequest(chartRef AIApplicationSpecChartRef, ) *AIApplicationChartInspectionRequest`

NewAIApplicationChartInspectionRequest instantiates a new AIApplicationChartInspectionRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationChartInspectionRequestWithDefaults

`func NewAIApplicationChartInspectionRequestWithDefaults() *AIApplicationChartInspectionRequest`

NewAIApplicationChartInspectionRequestWithDefaults instantiates a new AIApplicationChartInspectionRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChartRef

`func (o *AIApplicationChartInspectionRequest) GetChartRef() AIApplicationSpecChartRef`

GetChartRef returns the ChartRef field if non-nil, zero value otherwise.

### GetChartRefOk

`func (o *AIApplicationChartInspectionRequest) GetChartRefOk() (*AIApplicationSpecChartRef, bool)`

GetChartRefOk returns a tuple with the ChartRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartRef

`func (o *AIApplicationChartInspectionRequest) SetChartRef(v AIApplicationSpecChartRef)`

SetChartRef sets ChartRef field to given value.


### GetAuth

`func (o *AIApplicationChartInspectionRequest) GetAuth() AIApplicationAuth`

GetAuth returns the Auth field if non-nil, zero value otherwise.

### GetAuthOk

`func (o *AIApplicationChartInspectionRequest) GetAuthOk() (*AIApplicationAuth, bool)`

GetAuthOk returns a tuple with the Auth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuth

`func (o *AIApplicationChartInspectionRequest) SetAuth(v AIApplicationAuth)`

SetAuth sets Auth field to given value.

### HasAuth

`func (o *AIApplicationChartInspectionRequest) HasAuth() bool`

HasAuth returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


