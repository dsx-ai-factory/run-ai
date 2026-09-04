# GpuInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuType** | **string** |  | 
**GpuCount** | **int32** |  | 
**GpuMemoryBytes** | Pointer to **int64** | Memory of a single GPU device on this node, in bytes, taken from the node&#39;s GPU vendor labels. Note this is per device, unlike the TOTAL_GPU_MEMORY_BYTES node telemetry type which reports the node&#39;s total across all of its devices. Omitted when the node&#39;s GPU labels do not report memory. | [optional] 

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


### GetGpuMemoryBytes

`func (o *GpuInfo) GetGpuMemoryBytes() int64`

GetGpuMemoryBytes returns the GpuMemoryBytes field if non-nil, zero value otherwise.

### GetGpuMemoryBytesOk

`func (o *GpuInfo) GetGpuMemoryBytesOk() (*int64, bool)`

GetGpuMemoryBytesOk returns a tuple with the GpuMemoryBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryBytes

`func (o *GpuInfo) SetGpuMemoryBytes(v int64)`

SetGpuMemoryBytes sets GpuMemoryBytes field to given value.

### HasGpuMemoryBytes

`func (o *GpuInfo) HasGpuMemoryBytes() bool`

HasGpuMemoryBytes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


