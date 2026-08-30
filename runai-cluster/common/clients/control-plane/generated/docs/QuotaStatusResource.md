# QuotaStatusResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Gpu** | Pointer to **float32** | The number of GPUs (allocated/requested/...). | [optional] 
**Cpu** | Pointer to **float32** | CPU in Millicores (allocated/requested/...). | [optional] 
**Memory** | Pointer to **float32** | Memory in Mib (allocated/requested/...). | [optional] 

## Methods

### NewQuotaStatusResource

`func NewQuotaStatusResource() *QuotaStatusResource`

NewQuotaStatusResource instantiates a new QuotaStatusResource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQuotaStatusResourceWithDefaults

`func NewQuotaStatusResourceWithDefaults() *QuotaStatusResource`

NewQuotaStatusResourceWithDefaults instantiates a new QuotaStatusResource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpu

`func (o *QuotaStatusResource) GetGpu() float32`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *QuotaStatusResource) GetGpuOk() (*float32, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *QuotaStatusResource) SetGpu(v float32)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *QuotaStatusResource) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetCpu

`func (o *QuotaStatusResource) GetCpu() float32`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *QuotaStatusResource) GetCpuOk() (*float32, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *QuotaStatusResource) SetCpu(v float32)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *QuotaStatusResource) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *QuotaStatusResource) GetMemory() float32`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *QuotaStatusResource) GetMemoryOk() (*float32, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *QuotaStatusResource) SetMemory(v float32)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *QuotaStatusResource) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


