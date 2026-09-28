# AIApplicationChartUploadResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChartId** | **string** | Identifier of a Helm chart archive that was previously uploaded via the chart upload endpoint. When provided, the control plane serves the uploaded chart to the cluster for installation instead of resolving a Helm repository reference or a direct chart URL.  | 

## Methods

### NewAIApplicationChartUploadResponse

`func NewAIApplicationChartUploadResponse(chartId string, ) *AIApplicationChartUploadResponse`

NewAIApplicationChartUploadResponse instantiates a new AIApplicationChartUploadResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationChartUploadResponseWithDefaults

`func NewAIApplicationChartUploadResponseWithDefaults() *AIApplicationChartUploadResponse`

NewAIApplicationChartUploadResponseWithDefaults instantiates a new AIApplicationChartUploadResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChartId

`func (o *AIApplicationChartUploadResponse) GetChartId() string`

GetChartId returns the ChartId field if non-nil, zero value otherwise.

### GetChartIdOk

`func (o *AIApplicationChartUploadResponse) GetChartIdOk() (*string, bool)`

GetChartIdOk returns a tuple with the ChartId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartId

`func (o *AIApplicationChartUploadResponse) SetChartId(v string)`

SetChartId sets ChartId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


