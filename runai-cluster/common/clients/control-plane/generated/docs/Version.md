# Version

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SdkVersion** | **string** | The version of the CLI SDK | 
**CpVersion** | **string** | The version of the CP | 
**NeedUpgrade** | **bool** | Whether the CLI needs to be upgraded | [default to false]

## Methods

### NewVersion

`func NewVersion(sdkVersion string, cpVersion string, needUpgrade bool, ) *Version`

NewVersion instantiates a new Version object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVersionWithDefaults

`func NewVersionWithDefaults() *Version`

NewVersionWithDefaults instantiates a new Version object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSdkVersion

`func (o *Version) GetSdkVersion() string`

GetSdkVersion returns the SdkVersion field if non-nil, zero value otherwise.

### GetSdkVersionOk

`func (o *Version) GetSdkVersionOk() (*string, bool)`

GetSdkVersionOk returns a tuple with the SdkVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSdkVersion

`func (o *Version) SetSdkVersion(v string)`

SetSdkVersion sets SdkVersion field to given value.


### GetCpVersion

`func (o *Version) GetCpVersion() string`

GetCpVersion returns the CpVersion field if non-nil, zero value otherwise.

### GetCpVersionOk

`func (o *Version) GetCpVersionOk() (*string, bool)`

GetCpVersionOk returns a tuple with the CpVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpVersion

`func (o *Version) SetCpVersion(v string)`

SetCpVersion sets CpVersion field to given value.


### GetNeedUpgrade

`func (o *Version) GetNeedUpgrade() bool`

GetNeedUpgrade returns the NeedUpgrade field if non-nil, zero value otherwise.

### GetNeedUpgradeOk

`func (o *Version) GetNeedUpgradeOk() (*bool, bool)`

GetNeedUpgradeOk returns a tuple with the NeedUpgrade field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNeedUpgrade

`func (o *Version) SetNeedUpgrade(v bool)`

SetNeedUpgrade sets NeedUpgrade field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


