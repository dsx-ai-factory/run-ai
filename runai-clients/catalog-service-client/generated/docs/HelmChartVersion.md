# HelmChartVersion

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | **string** | The version of the chart (SemVer). | 
**AppVersion** | Pointer to **string** | The application version the chart represents. | [optional] 
**Description** | Pointer to **string** | A one-line description of the chart version. | [optional] 
**Created** | Pointer to **time.Time** | The date and time this chart version was created. | [optional] 

## Methods

### NewHelmChartVersion

`func NewHelmChartVersion(version string, ) *HelmChartVersion`

NewHelmChartVersion instantiates a new HelmChartVersion object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHelmChartVersionWithDefaults

`func NewHelmChartVersionWithDefaults() *HelmChartVersion`

NewHelmChartVersionWithDefaults instantiates a new HelmChartVersion object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVersion

`func (o *HelmChartVersion) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *HelmChartVersion) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *HelmChartVersion) SetVersion(v string)`

SetVersion sets Version field to given value.


### GetAppVersion

`func (o *HelmChartVersion) GetAppVersion() string`

GetAppVersion returns the AppVersion field if non-nil, zero value otherwise.

### GetAppVersionOk

`func (o *HelmChartVersion) GetAppVersionOk() (*string, bool)`

GetAppVersionOk returns a tuple with the AppVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppVersion

`func (o *HelmChartVersion) SetAppVersion(v string)`

SetAppVersion sets AppVersion field to given value.

### HasAppVersion

`func (o *HelmChartVersion) HasAppVersion() bool`

HasAppVersion returns a boolean if a field has been set.

### GetDescription

`func (o *HelmChartVersion) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *HelmChartVersion) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *HelmChartVersion) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *HelmChartVersion) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetCreated

`func (o *HelmChartVersion) GetCreated() time.Time`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *HelmChartVersion) GetCreatedOk() (*time.Time, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *HelmChartVersion) SetCreated(v time.Time)`

SetCreated sets Created field to given value.

### HasCreated

`func (o *HelmChartVersion) HasCreated() bool`

HasCreated returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


