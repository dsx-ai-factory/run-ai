# NodepoolGpuTypeDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | The GPU model, as reported by the GPU vendor labels on the node pool&#39;s nodes. | 
**GpuMemoryBytes** | Pointer to **int64** | Memory of a single GPU device of this type, in bytes. Omitted when the node pool&#39;s nodes do not report GPU memory. | [optional] 
**TotalDevices** | **int32** | Total number of GPU devices of this type across the node pool&#39;s nodes. | 

## Methods

### NewNodepoolGpuTypeDetails

`func NewNodepoolGpuTypeDetails(type_ string, totalDevices int32, ) *NodepoolGpuTypeDetails`

NewNodepoolGpuTypeDetails instantiates a new NodepoolGpuTypeDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolGpuTypeDetailsWithDefaults

`func NewNodepoolGpuTypeDetailsWithDefaults() *NodepoolGpuTypeDetails`

NewNodepoolGpuTypeDetailsWithDefaults instantiates a new NodepoolGpuTypeDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *NodepoolGpuTypeDetails) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *NodepoolGpuTypeDetails) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *NodepoolGpuTypeDetails) SetType(v string)`

SetType sets Type field to given value.


### GetGpuMemoryBytes

`func (o *NodepoolGpuTypeDetails) GetGpuMemoryBytes() int64`

GetGpuMemoryBytes returns the GpuMemoryBytes field if non-nil, zero value otherwise.

### GetGpuMemoryBytesOk

`func (o *NodepoolGpuTypeDetails) GetGpuMemoryBytesOk() (*int64, bool)`

GetGpuMemoryBytesOk returns a tuple with the GpuMemoryBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuMemoryBytes

`func (o *NodepoolGpuTypeDetails) SetGpuMemoryBytes(v int64)`

SetGpuMemoryBytes sets GpuMemoryBytes field to given value.

### HasGpuMemoryBytes

`func (o *NodepoolGpuTypeDetails) HasGpuMemoryBytes() bool`

HasGpuMemoryBytes returns a boolean if a field has been set.

### GetTotalDevices

`func (o *NodepoolGpuTypeDetails) GetTotalDevices() int32`

GetTotalDevices returns the TotalDevices field if non-nil, zero value otherwise.

### GetTotalDevicesOk

`func (o *NodepoolGpuTypeDetails) GetTotalDevicesOk() (*int32, bool)`

GetTotalDevicesOk returns a tuple with the TotalDevices field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalDevices

`func (o *NodepoolGpuTypeDetails) SetTotalDevices(v int32)`

SetTotalDevices sets TotalDevices field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


