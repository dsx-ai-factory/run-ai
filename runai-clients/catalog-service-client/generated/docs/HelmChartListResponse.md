# HelmChartListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Charts** | [**[]HelmChart**](HelmChart.md) | The list of charts in the helm repo. | 

## Methods

### NewHelmChartListResponse

`func NewHelmChartListResponse(charts []HelmChart, ) *HelmChartListResponse`

NewHelmChartListResponse instantiates a new HelmChartListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHelmChartListResponseWithDefaults

`func NewHelmChartListResponseWithDefaults() *HelmChartListResponse`

NewHelmChartListResponseWithDefaults instantiates a new HelmChartListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCharts

`func (o *HelmChartListResponse) GetCharts() []HelmChart`

GetCharts returns the Charts field if non-nil, zero value otherwise.

### GetChartsOk

`func (o *HelmChartListResponse) GetChartsOk() (*[]HelmChart, bool)`

GetChartsOk returns a tuple with the Charts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCharts

`func (o *HelmChartListResponse) SetCharts(v []HelmChart)`

SetCharts sets Charts field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


