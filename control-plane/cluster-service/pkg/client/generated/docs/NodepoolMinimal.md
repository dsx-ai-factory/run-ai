# NodepoolMinimal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The NodePool id | [readonly] 
**Name** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**Phase** | Pointer to [**NodepoolPhase**](NodepoolPhase.md) |  | [optional] 
**GpuNetworkAccelerationDetected** | Pointer to **bool** | Signals to all relevant parties about GPUNetworkAcceleration being detected in some nodepool&#39;s nodes | [optional] 
**SchedulingCapabilities** | Pointer to [**SchedulingCapabilities**](SchedulingCapabilities.md) |  | [optional] 
**GpuTypes** | Pointer to [**[]NodepoolGpuTypeDetails**](NodepoolGpuTypeDetails.md) | The GPU hardware in the node pool, one entry per distinct combination of GPU model and per-device memory found across its nodes. A heterogeneous node pool yields several entries. Omitted when the node pool has no GPU nodes. | [optional] 

## Methods

### NewNodepoolMinimal

`func NewNodepoolMinimal(id string, name string, clusterId string, ) *NodepoolMinimal`

NewNodepoolMinimal instantiates a new NodepoolMinimal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolMinimalWithDefaults

`func NewNodepoolMinimalWithDefaults() *NodepoolMinimal`

NewNodepoolMinimalWithDefaults instantiates a new NodepoolMinimal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NodepoolMinimal) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodepoolMinimal) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodepoolMinimal) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *NodepoolMinimal) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodepoolMinimal) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodepoolMinimal) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *NodepoolMinimal) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NodepoolMinimal) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NodepoolMinimal) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetPhase

`func (o *NodepoolMinimal) GetPhase() NodepoolPhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *NodepoolMinimal) GetPhaseOk() (*NodepoolPhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *NodepoolMinimal) SetPhase(v NodepoolPhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *NodepoolMinimal) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetGpuNetworkAccelerationDetected

`func (o *NodepoolMinimal) GetGpuNetworkAccelerationDetected() bool`

GetGpuNetworkAccelerationDetected returns the GpuNetworkAccelerationDetected field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectedOk

`func (o *NodepoolMinimal) GetGpuNetworkAccelerationDetectedOk() (*bool, bool)`

GetGpuNetworkAccelerationDetectedOk returns a tuple with the GpuNetworkAccelerationDetected field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetected

`func (o *NodepoolMinimal) SetGpuNetworkAccelerationDetected(v bool)`

SetGpuNetworkAccelerationDetected sets GpuNetworkAccelerationDetected field to given value.

### HasGpuNetworkAccelerationDetected

`func (o *NodepoolMinimal) HasGpuNetworkAccelerationDetected() bool`

HasGpuNetworkAccelerationDetected returns a boolean if a field has been set.

### GetSchedulingCapabilities

`func (o *NodepoolMinimal) GetSchedulingCapabilities() SchedulingCapabilities`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *NodepoolMinimal) GetSchedulingCapabilitiesOk() (*SchedulingCapabilities, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *NodepoolMinimal) SetSchedulingCapabilities(v SchedulingCapabilities)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.

### HasSchedulingCapabilities

`func (o *NodepoolMinimal) HasSchedulingCapabilities() bool`

HasSchedulingCapabilities returns a boolean if a field has been set.

### GetGpuTypes

`func (o *NodepoolMinimal) GetGpuTypes() []NodepoolGpuTypeDetails`

GetGpuTypes returns the GpuTypes field if non-nil, zero value otherwise.

### GetGpuTypesOk

`func (o *NodepoolMinimal) GetGpuTypesOk() (*[]NodepoolGpuTypeDetails, bool)`

GetGpuTypesOk returns a tuple with the GpuTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuTypes

`func (o *NodepoolMinimal) SetGpuTypes(v []NodepoolGpuTypeDetails)`

SetGpuTypes sets GpuTypes field to given value.

### HasGpuTypes

`func (o *NodepoolMinimal) HasGpuTypes() bool`

HasGpuTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


