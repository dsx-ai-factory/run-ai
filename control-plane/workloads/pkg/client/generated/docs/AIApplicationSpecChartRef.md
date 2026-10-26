# AIApplicationSpecChartRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repo** | Pointer to **string** | The HTTP or HTTPS URL of the Helm repository that hosts the chart. This must match a repository that is accessible. Used together with &#x60;name&#x60; and &#x60;version&#x60;. | [optional] 
**ChartName** | Pointer to **string** | Helm chart name. | [optional] 
**ChartVersion** | Pointer to **string** | Helm chart version. | [optional] 
**ChartURL** | Pointer to **string** | A direct HTTPS URL to a packaged Helm chart archive (.tgz). When provided, the chart is downloaded directly from this location instead of being resolved through a Helm repository reference. | [optional] 
**ChartId** | Pointer to **string** | Identifier of a Helm chart archive that was previously uploaded via the chart upload endpoint. When provided, the control plane serves the uploaded chart to the cluster for installation instead of resolving a Helm repository reference or a direct chart URL.  | [optional] 

## Methods

### NewAIApplicationSpecChartRef

`func NewAIApplicationSpecChartRef() *AIApplicationSpecChartRef`

NewAIApplicationSpecChartRef instantiates a new AIApplicationSpecChartRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationSpecChartRefWithDefaults

`func NewAIApplicationSpecChartRefWithDefaults() *AIApplicationSpecChartRef`

NewAIApplicationSpecChartRefWithDefaults instantiates a new AIApplicationSpecChartRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepo

`func (o *AIApplicationSpecChartRef) GetRepo() string`

GetRepo returns the Repo field if non-nil, zero value otherwise.

### GetRepoOk

`func (o *AIApplicationSpecChartRef) GetRepoOk() (*string, bool)`

GetRepoOk returns a tuple with the Repo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepo

`func (o *AIApplicationSpecChartRef) SetRepo(v string)`

SetRepo sets Repo field to given value.

### HasRepo

`func (o *AIApplicationSpecChartRef) HasRepo() bool`

HasRepo returns a boolean if a field has been set.

### GetChartName

`func (o *AIApplicationSpecChartRef) GetChartName() string`

GetChartName returns the ChartName field if non-nil, zero value otherwise.

### GetChartNameOk

`func (o *AIApplicationSpecChartRef) GetChartNameOk() (*string, bool)`

GetChartNameOk returns a tuple with the ChartName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartName

`func (o *AIApplicationSpecChartRef) SetChartName(v string)`

SetChartName sets ChartName field to given value.

### HasChartName

`func (o *AIApplicationSpecChartRef) HasChartName() bool`

HasChartName returns a boolean if a field has been set.

### GetChartVersion

`func (o *AIApplicationSpecChartRef) GetChartVersion() string`

GetChartVersion returns the ChartVersion field if non-nil, zero value otherwise.

### GetChartVersionOk

`func (o *AIApplicationSpecChartRef) GetChartVersionOk() (*string, bool)`

GetChartVersionOk returns a tuple with the ChartVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartVersion

`func (o *AIApplicationSpecChartRef) SetChartVersion(v string)`

SetChartVersion sets ChartVersion field to given value.

### HasChartVersion

`func (o *AIApplicationSpecChartRef) HasChartVersion() bool`

HasChartVersion returns a boolean if a field has been set.

### GetChartURL

`func (o *AIApplicationSpecChartRef) GetChartURL() string`

GetChartURL returns the ChartURL field if non-nil, zero value otherwise.

### GetChartURLOk

`func (o *AIApplicationSpecChartRef) GetChartURLOk() (*string, bool)`

GetChartURLOk returns a tuple with the ChartURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartURL

`func (o *AIApplicationSpecChartRef) SetChartURL(v string)`

SetChartURL sets ChartURL field to given value.

### HasChartURL

`func (o *AIApplicationSpecChartRef) HasChartURL() bool`

HasChartURL returns a boolean if a field has been set.

### GetChartId

`func (o *AIApplicationSpecChartRef) GetChartId() string`

GetChartId returns the ChartId field if non-nil, zero value otherwise.

### GetChartIdOk

`func (o *AIApplicationSpecChartRef) GetChartIdOk() (*string, bool)`

GetChartIdOk returns a tuple with the ChartId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartId

`func (o *AIApplicationSpecChartRef) SetChartId(v string)`

SetChartId sets ChartId field to given value.

### HasChartId

`func (o *AIApplicationSpecChartRef) HasChartId() bool`

HasChartId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


