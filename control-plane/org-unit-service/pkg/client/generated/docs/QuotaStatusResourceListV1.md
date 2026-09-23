# QuotaStatusResourceListV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | Pointer to **float32** | The number of GPUs (allocated/requested/...). | [optional] 
**Cpu** | Pointer to **float32** | CPU in Millicores (allocated/requested/...). | [optional] 
**Memory** | Pointer to **float32** | Memory in Mib (allocated/requested/...). | [optional] 

## Methods

### NewQuotaStatusResourceListV1

`func NewQuotaStatusResourceListV1() *QuotaStatusResourceListV1`

NewQuotaStatusResourceListV1 instantiates a new QuotaStatusResourceListV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQuotaStatusResourceListV1WithDefaults

`func NewQuotaStatusResourceListV1WithDefaults() *QuotaStatusResourceListV1`

NewQuotaStatusResourceListV1WithDefaults instantiates a new QuotaStatusResourceListV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *QuotaStatusResourceListV1) GetGpu() float32`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *QuotaStatusResourceListV1) GetGpuOk() (*float32, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *QuotaStatusResourceListV1) SetGpu(v float32)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *QuotaStatusResourceListV1) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *QuotaStatusResourceListV1) GetCpu() float32`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *QuotaStatusResourceListV1) GetCpuOk() (*float32, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *QuotaStatusResourceListV1) SetCpu(v float32)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *QuotaStatusResourceListV1) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *QuotaStatusResourceListV1) GetMemory() float32`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *QuotaStatusResourceListV1) GetMemoryOk() (*float32, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *QuotaStatusResourceListV1) SetMemory(v float32)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *QuotaStatusResourceListV1) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


