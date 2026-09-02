# GpuInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuType** | **string** |  | 
**GpuCount** | **int32** |  | 

## Methods

### NewGpuInfo

`func NewGpuInfo(gpuType string, gpuCount int32, ) *GpuInfo`

NewGpuInfo instantiates a new GpuInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGpuInfoWithDefaults

`func NewGpuInfoWithDefaults() *GpuInfo`

NewGpuInfoWithDefaults instantiates a new GpuInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuType

`func (o *GpuInfo) GetGpuType() string`

GetGpuType returns the GpuType field if non-nil, zero value otherwise.

### GetGpuTypeOk

`func (o *GpuInfo) GetGpuTypeOk() (*string, bool)`

GetGpuTypeOk returns a tuple with the GpuType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuType

`func (o *GpuInfo) SetGpuType(v string)`

SetGpuType sets GpuType field to given value.


### GetGpuCount

`func (o *GpuInfo) GetGpuCount() int32`

GetGpuCount returns the GpuCount field if non-nil, zero value otherwise.

### GetGpuCountOk

`func (o *GpuInfo) GetGpuCountOk() (*int32, bool)`

GetGpuCountOk returns a tuple with the GpuCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuCount

`func (o *GpuInfo) SetGpuCount(v int32)`

SetGpuCount sets GpuCount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


