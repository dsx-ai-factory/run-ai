# HelmChart

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the helm chart. | 
**Versions** | [**[]HelmChartVersion**](HelmChartVersion.md) | The list of available versions for this chart. | 

## Methods

### NewHelmChart

`func NewHelmChart(name string, versions []HelmChartVersion, ) *HelmChart`

NewHelmChart instantiates a new HelmChart object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHelmChartWithDefaults

`func NewHelmChartWithDefaults() *HelmChart`

NewHelmChartWithDefaults instantiates a new HelmChart object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *HelmChart) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *HelmChart) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *HelmChart) SetName(v string)`

SetName sets Name field to given value.


### GetVersions

`func (o *HelmChart) GetVersions() []HelmChartVersion`

GetVersions returns the Versions field if non-nil, zero value otherwise.

### GetVersionsOk

`func (o *HelmChart) GetVersionsOk() (*[]HelmChartVersion, bool)`

GetVersionsOk returns a tuple with the Versions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersions

`func (o *HelmChart) SetVersions(v []HelmChartVersion)`

SetVersions sets Versions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


