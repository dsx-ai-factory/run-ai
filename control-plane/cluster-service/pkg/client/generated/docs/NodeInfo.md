# NodeInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | **string** | The calculated status of the node. | 
**Conditions** | Pointer to [**[]NodeStatusConditionDetails**](NodeStatusConditionDetails.md) |  | [optional] 
**Taints** | Pointer to [**[]NodeTaint**](NodeTaint.md) |  | [optional] 
**NodePool** | **string** | The node&#39;s NodePool. | 
**CreatedAt** | **time.Time** |  | 
**GpuInfo** | Pointer to [**NullableGpuInfo**](GpuInfo.md) |  | [optional] 
**NvLinkDomainUid** | Pointer to **NullableString** | NV Link Domain Uid | [optional] 
**NvLinkCliqueId** | Pointer to **NullableString** | NV Link Clique Id | [optional] 
**NetworkTopologyLabels** | Pointer to [**[]NetworkTopologyLabel**](NetworkTopologyLabel.md) | A list of Kubernetes node label key/value pairs that represent the node’s position in the cluster’s network topology (for example, region, zone, block, rack, or hostname). The topology is attached to the node pool on which the workload runs. The label order represents the hierarchy, from farthest (first) to closest (last). These labels are used by the NVIDIA Run:ai Scheduler when making placement decisions. | [optional] 
**Numa** | Pointer to [**NullableNumaInfo**](NumaInfo.md) |  | [optional] 

## Methods

### NewNodeInfo

`func NewNodeInfo(status string, nodePool string, createdAt time.Time, ) *NodeInfo`

NewNodeInfo instantiates a new NodeInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeInfoWithDefaults

`func NewNodeInfoWithDefaults() *NodeInfo`

NewNodeInfoWithDefaults instantiates a new NodeInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *NodeInfo) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NodeInfo) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NodeInfo) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetConditions

`func (o *NodeInfo) GetConditions() []NodeStatusConditionDetails`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *NodeInfo) GetConditionsOk() (*[]NodeStatusConditionDetails, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *NodeInfo) SetConditions(v []NodeStatusConditionDetails)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *NodeInfo) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetTaints

`func (o *NodeInfo) GetTaints() []NodeTaint`

GetTaints returns the Taints field if non-nil, zero value otherwise.

### GetTaintsOk

`func (o *NodeInfo) GetTaintsOk() (*[]NodeTaint, bool)`

GetTaintsOk returns a tuple with the Taints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTaints

`func (o *NodeInfo) SetTaints(v []NodeTaint)`

SetTaints sets Taints field to given value.

### HasTaints

`func (o *NodeInfo) HasTaints() bool`

HasTaints returns a boolean if a field has been set.

### GetNodePool

`func (o *NodeInfo) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *NodeInfo) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *NodeInfo) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.


### GetCreatedAt

`func (o *NodeInfo) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *NodeInfo) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *NodeInfo) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetGpuInfo

`func (o *NodeInfo) GetGpuInfo() GpuInfo`

GetGpuInfo returns the GpuInfo field if non-nil, zero value otherwise.

### GetGpuInfoOk

`func (o *NodeInfo) GetGpuInfoOk() (*GpuInfo, bool)`

GetGpuInfoOk returns a tuple with the GpuInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuInfo

`func (o *NodeInfo) SetGpuInfo(v GpuInfo)`

SetGpuInfo sets GpuInfo field to given value.

### HasGpuInfo

`func (o *NodeInfo) HasGpuInfo() bool`

HasGpuInfo returns a boolean if a field has been set.

### SetGpuInfoNil

`func (o *NodeInfo) SetGpuInfoNil(b bool)`

 SetGpuInfoNil sets the value for GpuInfo to be an explicit nil

### UnsetGpuInfo
`func (o *NodeInfo) UnsetGpuInfo()`

UnsetGpuInfo ensures that no value is present for GpuInfo, not even an explicit nil
### GetNvLinkDomainUid

`func (o *NodeInfo) GetNvLinkDomainUid() string`

GetNvLinkDomainUid returns the NvLinkDomainUid field if non-nil, zero value otherwise.

### GetNvLinkDomainUidOk

`func (o *NodeInfo) GetNvLinkDomainUidOk() (*string, bool)`

GetNvLinkDomainUidOk returns a tuple with the NvLinkDomainUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNvLinkDomainUid

`func (o *NodeInfo) SetNvLinkDomainUid(v string)`

SetNvLinkDomainUid sets NvLinkDomainUid field to given value.

### HasNvLinkDomainUid

`func (o *NodeInfo) HasNvLinkDomainUid() bool`

HasNvLinkDomainUid returns a boolean if a field has been set.

### SetNvLinkDomainUidNil

`func (o *NodeInfo) SetNvLinkDomainUidNil(b bool)`

 SetNvLinkDomainUidNil sets the value for NvLinkDomainUid to be an explicit nil

### UnsetNvLinkDomainUid
`func (o *NodeInfo) UnsetNvLinkDomainUid()`

UnsetNvLinkDomainUid ensures that no value is present for NvLinkDomainUid, not even an explicit nil
### GetNvLinkCliqueId

`func (o *NodeInfo) GetNvLinkCliqueId() string`

GetNvLinkCliqueId returns the NvLinkCliqueId field if non-nil, zero value otherwise.

### GetNvLinkCliqueIdOk

`func (o *NodeInfo) GetNvLinkCliqueIdOk() (*string, bool)`

GetNvLinkCliqueIdOk returns a tuple with the NvLinkCliqueId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNvLinkCliqueId

`func (o *NodeInfo) SetNvLinkCliqueId(v string)`

SetNvLinkCliqueId sets NvLinkCliqueId field to given value.

### HasNvLinkCliqueId

`func (o *NodeInfo) HasNvLinkCliqueId() bool`

HasNvLinkCliqueId returns a boolean if a field has been set.

### SetNvLinkCliqueIdNil

`func (o *NodeInfo) SetNvLinkCliqueIdNil(b bool)`

 SetNvLinkCliqueIdNil sets the value for NvLinkCliqueId to be an explicit nil

### UnsetNvLinkCliqueId
`func (o *NodeInfo) UnsetNvLinkCliqueId()`

UnsetNvLinkCliqueId ensures that no value is present for NvLinkCliqueId, not even an explicit nil
### GetNetworkTopologyLabels

`func (o *NodeInfo) GetNetworkTopologyLabels() []NetworkTopologyLabel`

GetNetworkTopologyLabels returns the NetworkTopologyLabels field if non-nil, zero value otherwise.

### GetNetworkTopologyLabelsOk

`func (o *NodeInfo) GetNetworkTopologyLabelsOk() (*[]NetworkTopologyLabel, bool)`

GetNetworkTopologyLabelsOk returns a tuple with the NetworkTopologyLabels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyLabels

`func (o *NodeInfo) SetNetworkTopologyLabels(v []NetworkTopologyLabel)`

SetNetworkTopologyLabels sets NetworkTopologyLabels field to given value.

### HasNetworkTopologyLabels

`func (o *NodeInfo) HasNetworkTopologyLabels() bool`

HasNetworkTopologyLabels returns a boolean if a field has been set.

### SetNetworkTopologyLabelsNil

`func (o *NodeInfo) SetNetworkTopologyLabelsNil(b bool)`

 SetNetworkTopologyLabelsNil sets the value for NetworkTopologyLabels to be an explicit nil

### UnsetNetworkTopologyLabels
`func (o *NodeInfo) UnsetNetworkTopologyLabels()`

UnsetNetworkTopologyLabels ensures that no value is present for NetworkTopologyLabels, not even an explicit nil
### GetNuma

`func (o *NodeInfo) GetNuma() NumaInfo`

GetNuma returns the Numa field if non-nil, zero value otherwise.

### GetNumaOk

`func (o *NodeInfo) GetNumaOk() (*NumaInfo, bool)`

GetNumaOk returns a tuple with the Numa field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNuma

`func (o *NodeInfo) SetNuma(v NumaInfo)`

SetNuma sets Numa field to given value.

### HasNuma

`func (o *NodeInfo) HasNuma() bool`

HasNuma returns a boolean if a field has been set.

### SetNumaNil

`func (o *NodeInfo) SetNumaNil(b bool)`

 SetNumaNil sets the value for Numa to be an explicit nil

### UnsetNuma
`func (o *NodeInfo) UnsetNuma()`

UnsetNuma ensures that no value is present for Numa, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


