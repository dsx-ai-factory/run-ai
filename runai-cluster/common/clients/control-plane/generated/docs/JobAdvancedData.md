# JobAdvancedData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IdleSeconds** | Pointer to **float32** |  | [optional] 
**GrEngineActive** | Pointer to **float32** | The fraction of time any portion of the graphics or compute engines were active | [optional] 
**DramActive** | Pointer to **float32** | The fraction of cycles where data was sent to or received from device memory | [optional] 
**SmActive** | Pointer to **float32** | The fraction of time at least one warp was active on a multiprocessor, averaged over all multiprocessors | [optional] 
**SmOccupancy** | Pointer to **float32** | The fraction of resident warps on a multiprocessor, relative to the maximum number of concurrent warps supported on a multiprocessor | [optional] 
**PipeTensorActive** | Pointer to **float32** | The fraction of cycles the tensor (HMMA / IMMA) pipe was active | [optional] 
**PipeFp64Active** | Pointer to **float32** | The fraction of cycles the FP64 (double precision) pipe was active | [optional] 
**PipeFp32Active** | Pointer to **float32** | The fraction of cycles the FMA (FP32 (single precision), and integer) pipe was active | [optional] 
**PipeFp16Active** | Pointer to **float32** | The fraction of cycles the FP16 (half precision) pipe was active | [optional] 
**NvlinkTxBytes** | Pointer to **float32** | The rate of data transmitted over NVLink, not including protocol headers, in bytes per second | [optional] 
**NvlinkRxBytes** | Pointer to **float32** | The rate of data received over NVLink, not including protocol headers, in bytes per second | [optional] 
**PcieTxBytes** | Pointer to **float32** | The rate of data transmitted over the PCIe bus, including both protocol headers and data payloads, in bytes per second | [optional] 
**PcieRxBytes** | Pointer to **float32** | The rate of data received over the PCIe bus, including both protocol headers and data payloads, in bytes per second | [optional] 

## Methods

### NewJobAdvancedData

`func NewJobAdvancedData() *JobAdvancedData`

NewJobAdvancedData instantiates a new JobAdvancedData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobAdvancedDataWithDefaults

`func NewJobAdvancedDataWithDefaults() *JobAdvancedData`

NewJobAdvancedDataWithDefaults instantiates a new JobAdvancedData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIdleSeconds

`func (o *JobAdvancedData) GetIdleSeconds() float32`

GetIdleSeconds returns the IdleSeconds field if non-nil, zero value otherwise.

### GetIdleSecondsOk

`func (o *JobAdvancedData) GetIdleSecondsOk() (*float32, bool)`

GetIdleSecondsOk returns a tuple with the IdleSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdleSeconds

`func (o *JobAdvancedData) SetIdleSeconds(v float32)`

SetIdleSeconds sets IdleSeconds field to given value.

### HasIdleSeconds

`func (o *JobAdvancedData) HasIdleSeconds() bool`

HasIdleSeconds returns a boolean if a field has been set.

### GetGrEngineActive

`func (o *JobAdvancedData) GetGrEngineActive() float32`

GetGrEngineActive returns the GrEngineActive field if non-nil, zero value otherwise.

### GetGrEngineActiveOk

`func (o *JobAdvancedData) GetGrEngineActiveOk() (*float32, bool)`

GetGrEngineActiveOk returns a tuple with the GrEngineActive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrEngineActive

`func (o *JobAdvancedData) SetGrEngineActive(v float32)`

SetGrEngineActive sets GrEngineActive field to given value.

### HasGrEngineActive

`func (o *JobAdvancedData) HasGrEngineActive() bool`

HasGrEngineActive returns a boolean if a field has been set.

### GetDramActive

`func (o *JobAdvancedData) GetDramActive() float32`

GetDramActive returns the DramActive field if non-nil, zero value otherwise.

### GetDramActiveOk

`func (o *JobAdvancedData) GetDramActiveOk() (*float32, bool)`

GetDramActiveOk returns a tuple with the DramActive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDramActive

`func (o *JobAdvancedData) SetDramActive(v float32)`

SetDramActive sets DramActive field to given value.

### HasDramActive

`func (o *JobAdvancedData) HasDramActive() bool`

HasDramActive returns a boolean if a field has been set.

### GetSmActive

`func (o *JobAdvancedData) GetSmActive() float32`

GetSmActive returns the SmActive field if non-nil, zero value otherwise.

### GetSmActiveOk

`func (o *JobAdvancedData) GetSmActiveOk() (*float32, bool)`

GetSmActiveOk returns a tuple with the SmActive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmActive

`func (o *JobAdvancedData) SetSmActive(v float32)`

SetSmActive sets SmActive field to given value.

### HasSmActive

`func (o *JobAdvancedData) HasSmActive() bool`

HasSmActive returns a boolean if a field has been set.

### GetSmOccupancy

`func (o *JobAdvancedData) GetSmOccupancy() float32`

GetSmOccupancy returns the SmOccupancy field if non-nil, zero value otherwise.

### GetSmOccupancyOk

`func (o *JobAdvancedData) GetSmOccupancyOk() (*float32, bool)`

GetSmOccupancyOk returns a tuple with the SmOccupancy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmOccupancy

`func (o *JobAdvancedData) SetSmOccupancy(v float32)`

SetSmOccupancy sets SmOccupancy field to given value.

### HasSmOccupancy

`func (o *JobAdvancedData) HasSmOccupancy() bool`

HasSmOccupancy returns a boolean if a field has been set.

### GetPipeTensorActive

`func (o *JobAdvancedData) GetPipeTensorActive() float32`

GetPipeTensorActive returns the PipeTensorActive field if non-nil, zero value otherwise.

### GetPipeTensorActiveOk

`func (o *JobAdvancedData) GetPipeTensorActiveOk() (*float32, bool)`

GetPipeTensorActiveOk returns a tuple with the PipeTensorActive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPipeTensorActive

`func (o *JobAdvancedData) SetPipeTensorActive(v float32)`

SetPipeTensorActive sets PipeTensorActive field to given value.

### HasPipeTensorActive

`func (o *JobAdvancedData) HasPipeTensorActive() bool`

HasPipeTensorActive returns a boolean if a field has been set.

### GetPipeFp64Active

`func (o *JobAdvancedData) GetPipeFp64Active() float32`

GetPipeFp64Active returns the PipeFp64Active field if non-nil, zero value otherwise.

### GetPipeFp64ActiveOk

`func (o *JobAdvancedData) GetPipeFp64ActiveOk() (*float32, bool)`

GetPipeFp64ActiveOk returns a tuple with the PipeFp64Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPipeFp64Active

`func (o *JobAdvancedData) SetPipeFp64Active(v float32)`

SetPipeFp64Active sets PipeFp64Active field to given value.

### HasPipeFp64Active

`func (o *JobAdvancedData) HasPipeFp64Active() bool`

HasPipeFp64Active returns a boolean if a field has been set.

### GetPipeFp32Active

`func (o *JobAdvancedData) GetPipeFp32Active() float32`

GetPipeFp32Active returns the PipeFp32Active field if non-nil, zero value otherwise.

### GetPipeFp32ActiveOk

`func (o *JobAdvancedData) GetPipeFp32ActiveOk() (*float32, bool)`

GetPipeFp32ActiveOk returns a tuple with the PipeFp32Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPipeFp32Active

`func (o *JobAdvancedData) SetPipeFp32Active(v float32)`

SetPipeFp32Active sets PipeFp32Active field to given value.

### HasPipeFp32Active

`func (o *JobAdvancedData) HasPipeFp32Active() bool`

HasPipeFp32Active returns a boolean if a field has been set.

### GetPipeFp16Active

`func (o *JobAdvancedData) GetPipeFp16Active() float32`

GetPipeFp16Active returns the PipeFp16Active field if non-nil, zero value otherwise.

### GetPipeFp16ActiveOk

`func (o *JobAdvancedData) GetPipeFp16ActiveOk() (*float32, bool)`

GetPipeFp16ActiveOk returns a tuple with the PipeFp16Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPipeFp16Active

`func (o *JobAdvancedData) SetPipeFp16Active(v float32)`

SetPipeFp16Active sets PipeFp16Active field to given value.

### HasPipeFp16Active

`func (o *JobAdvancedData) HasPipeFp16Active() bool`

HasPipeFp16Active returns a boolean if a field has been set.

### GetNvlinkTxBytes

`func (o *JobAdvancedData) GetNvlinkTxBytes() float32`

GetNvlinkTxBytes returns the NvlinkTxBytes field if non-nil, zero value otherwise.

### GetNvlinkTxBytesOk

`func (o *JobAdvancedData) GetNvlinkTxBytesOk() (*float32, bool)`

GetNvlinkTxBytesOk returns a tuple with the NvlinkTxBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNvlinkTxBytes

`func (o *JobAdvancedData) SetNvlinkTxBytes(v float32)`

SetNvlinkTxBytes sets NvlinkTxBytes field to given value.

### HasNvlinkTxBytes

`func (o *JobAdvancedData) HasNvlinkTxBytes() bool`

HasNvlinkTxBytes returns a boolean if a field has been set.

### GetNvlinkRxBytes

`func (o *JobAdvancedData) GetNvlinkRxBytes() float32`

GetNvlinkRxBytes returns the NvlinkRxBytes field if non-nil, zero value otherwise.

### GetNvlinkRxBytesOk

`func (o *JobAdvancedData) GetNvlinkRxBytesOk() (*float32, bool)`

GetNvlinkRxBytesOk returns a tuple with the NvlinkRxBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNvlinkRxBytes

`func (o *JobAdvancedData) SetNvlinkRxBytes(v float32)`

SetNvlinkRxBytes sets NvlinkRxBytes field to given value.

### HasNvlinkRxBytes

`func (o *JobAdvancedData) HasNvlinkRxBytes() bool`

HasNvlinkRxBytes returns a boolean if a field has been set.

### GetPcieTxBytes

`func (o *JobAdvancedData) GetPcieTxBytes() float32`

GetPcieTxBytes returns the PcieTxBytes field if non-nil, zero value otherwise.

### GetPcieTxBytesOk

`func (o *JobAdvancedData) GetPcieTxBytesOk() (*float32, bool)`

GetPcieTxBytesOk returns a tuple with the PcieTxBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPcieTxBytes

`func (o *JobAdvancedData) SetPcieTxBytes(v float32)`

SetPcieTxBytes sets PcieTxBytes field to given value.

### HasPcieTxBytes

`func (o *JobAdvancedData) HasPcieTxBytes() bool`

HasPcieTxBytes returns a boolean if a field has been set.

### GetPcieRxBytes

`func (o *JobAdvancedData) GetPcieRxBytes() float32`

GetPcieRxBytes returns the PcieRxBytes field if non-nil, zero value otherwise.

### GetPcieRxBytesOk

`func (o *JobAdvancedData) GetPcieRxBytesOk() (*float32, bool)`

GetPcieRxBytesOk returns a tuple with the PcieRxBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPcieRxBytes

`func (o *JobAdvancedData) SetPcieRxBytes(v float32)`

SetPcieRxBytes sets PcieRxBytes field to given value.

### HasPcieRxBytes

`func (o *JobAdvancedData) HasPcieRxBytes() bool`

HasPcieRxBytes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


