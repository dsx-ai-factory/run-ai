# GpuInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PciDeviceId** | Pointer to **string** | NVIDIA PCI device and vendor ID from reported physical hardware; omitted when unknown. | [optional] 
**PciDeviceIdObservedAt** | Pointer to **NullableTime** | Time the hardware collector observed pciDeviceId. Identity is unknown without a paired observation within five minutes and at most one minute in the future. This timestamp is never replaced by the control-plane ingestion time. | [optional] 
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

### GetPciDeviceId

`func (o *GpuInfo) GetPciDeviceId() string`

GetPciDeviceId returns the PciDeviceId field if non-nil, zero value otherwise.

### GetPciDeviceIdOk

`func (o *GpuInfo) GetPciDeviceIdOk() (*string, bool)`

GetPciDeviceIdOk returns a tuple with the PciDeviceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPciDeviceId

`func (o *GpuInfo) SetPciDeviceId(v string)`

SetPciDeviceId sets PciDeviceId field to given value.

### HasPciDeviceId

`func (o *GpuInfo) HasPciDeviceId() bool`

HasPciDeviceId returns a boolean if a field has been set.

### GetPciDeviceIdObservedAt

`func (o *GpuInfo) GetPciDeviceIdObservedAt() time.Time`

GetPciDeviceIdObservedAt returns the PciDeviceIdObservedAt field if non-nil, zero value otherwise.

### GetPciDeviceIdObservedAtOk

`func (o *GpuInfo) GetPciDeviceIdObservedAtOk() (*time.Time, bool)`

GetPciDeviceIdObservedAtOk returns a tuple with the PciDeviceIdObservedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPciDeviceIdObservedAt

`func (o *GpuInfo) SetPciDeviceIdObservedAt(v time.Time)`

SetPciDeviceIdObservedAt sets PciDeviceIdObservedAt field to given value.

### HasPciDeviceIdObservedAt

`func (o *GpuInfo) HasPciDeviceIdObservedAt() bool`

HasPciDeviceIdObservedAt returns a boolean if a field has been set.

### SetPciDeviceIdObservedAtNil

`func (o *GpuInfo) SetPciDeviceIdObservedAtNil(b bool)`

 SetPciDeviceIdObservedAtNil sets the value for PciDeviceIdObservedAt to be an explicit nil

### UnsetPciDeviceIdObservedAt
`func (o *GpuInfo) UnsetPciDeviceIdObservedAt()`

UnsetPciDeviceIdObservedAt ensures that no value is present for PciDeviceIdObservedAt, not even an explicit nil
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


