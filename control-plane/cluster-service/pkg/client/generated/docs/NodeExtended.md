# NodeExtended

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
**Numa** | Pointer to [**NumaInfoExtended**](NumaInfoExtended.md) |  | [optional] 
**Name** | **string** | The name of the node | 
**Id** | Pointer to **string** | The unique identifier of the node. | [optional] 
**ClusterUuid** | **string** | The id of the cluster. | 
**UpdatedAt** | **time.Time** |  | 

## Methods

### NewNodeExtended

`func NewNodeExtended(status string, nodePool string, createdAt time.Time, name string, clusterUuid string, updatedAt time.Time, ) *NodeExtended`

NewNodeExtended instantiates a new NodeExtended object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeExtendedWithDefaults

`func NewNodeExtendedWithDefaults() *NodeExtended`

NewNodeExtendedWithDefaults instantiates a new NodeExtended object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *NodeExtended) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NodeExtended) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NodeExtended) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetConditions

`func (o *NodeExtended) GetConditions() []NodeStatusConditionDetails`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *NodeExtended) GetConditionsOk() (*[]NodeStatusConditionDetails, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *NodeExtended) SetConditions(v []NodeStatusConditionDetails)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *NodeExtended) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetTaints

`func (o *NodeExtended) GetTaints() []NodeTaint`

GetTaints returns the Taints field if non-nil, zero value otherwise.

### GetTaintsOk

`func (o *NodeExtended) GetTaintsOk() (*[]NodeTaint, bool)`

GetTaintsOk returns a tuple with the Taints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTaints

`func (o *NodeExtended) SetTaints(v []NodeTaint)`

SetTaints sets Taints field to given value.

### HasTaints

`func (o *NodeExtended) HasTaints() bool`

HasTaints returns a boolean if a field has been set.

### GetNodePool

`func (o *NodeExtended) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *NodeExtended) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *NodeExtended) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.


### GetCreatedAt

`func (o *NodeExtended) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *NodeExtended) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *NodeExtended) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetGpuInfo

`func (o *NodeExtended) GetGpuInfo() GpuInfo`

GetGpuInfo returns the GpuInfo field if non-nil, zero value otherwise.

### GetGpuInfoOk

`func (o *NodeExtended) GetGpuInfoOk() (*GpuInfo, bool)`

GetGpuInfoOk returns a tuple with the GpuInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuInfo

`func (o *NodeExtended) SetGpuInfo(v GpuInfo)`

SetGpuInfo sets GpuInfo field to given value.

### HasGpuInfo

`func (o *NodeExtended) HasGpuInfo() bool`

HasGpuInfo returns a boolean if a field has been set.

### SetGpuInfoNil

`func (o *NodeExtended) SetGpuInfoNil(b bool)`

 SetGpuInfoNil sets the value for GpuInfo to be an explicit nil

### UnsetGpuInfo
`func (o *NodeExtended) UnsetGpuInfo()`

UnsetGpuInfo ensures that no value is present for GpuInfo, not even an explicit nil
### GetNvLinkDomainUid

`func (o *NodeExtended) GetNvLinkDomainUid() string`

GetNvLinkDomainUid returns the NvLinkDomainUid field if non-nil, zero value otherwise.

### GetNvLinkDomainUidOk

`func (o *NodeExtended) GetNvLinkDomainUidOk() (*string, bool)`

GetNvLinkDomainUidOk returns a tuple with the NvLinkDomainUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNvLinkDomainUid

`func (o *NodeExtended) SetNvLinkDomainUid(v string)`

SetNvLinkDomainUid sets NvLinkDomainUid field to given value.

### HasNvLinkDomainUid

`func (o *NodeExtended) HasNvLinkDomainUid() bool`

HasNvLinkDomainUid returns a boolean if a field has been set.

### SetNvLinkDomainUidNil

`func (o *NodeExtended) SetNvLinkDomainUidNil(b bool)`

 SetNvLinkDomainUidNil sets the value for NvLinkDomainUid to be an explicit nil

### UnsetNvLinkDomainUid
`func (o *NodeExtended) UnsetNvLinkDomainUid()`

UnsetNvLinkDomainUid ensures that no value is present for NvLinkDomainUid, not even an explicit nil
### GetNvLinkCliqueId

`func (o *NodeExtended) GetNvLinkCliqueId() string`

GetNvLinkCliqueId returns the NvLinkCliqueId field if non-nil, zero value otherwise.

### GetNvLinkCliqueIdOk

`func (o *NodeExtended) GetNvLinkCliqueIdOk() (*string, bool)`

GetNvLinkCliqueIdOk returns a tuple with the NvLinkCliqueId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNvLinkCliqueId

`func (o *NodeExtended) SetNvLinkCliqueId(v string)`

SetNvLinkCliqueId sets NvLinkCliqueId field to given value.

### HasNvLinkCliqueId

`func (o *NodeExtended) HasNvLinkCliqueId() bool`

HasNvLinkCliqueId returns a boolean if a field has been set.

### SetNvLinkCliqueIdNil

`func (o *NodeExtended) SetNvLinkCliqueIdNil(b bool)`

 SetNvLinkCliqueIdNil sets the value for NvLinkCliqueId to be an explicit nil

### UnsetNvLinkCliqueId
`func (o *NodeExtended) UnsetNvLinkCliqueId()`

UnsetNvLinkCliqueId ensures that no value is present for NvLinkCliqueId, not even an explicit nil
### GetNetworkTopologyLabels

`func (o *NodeExtended) GetNetworkTopologyLabels() []NetworkTopologyLabel`

GetNetworkTopologyLabels returns the NetworkTopologyLabels field if non-nil, zero value otherwise.

### GetNetworkTopologyLabelsOk

`func (o *NodeExtended) GetNetworkTopologyLabelsOk() (*[]NetworkTopologyLabel, bool)`

GetNetworkTopologyLabelsOk returns a tuple with the NetworkTopologyLabels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyLabels

`func (o *NodeExtended) SetNetworkTopologyLabels(v []NetworkTopologyLabel)`

SetNetworkTopologyLabels sets NetworkTopologyLabels field to given value.

### HasNetworkTopologyLabels

`func (o *NodeExtended) HasNetworkTopologyLabels() bool`

HasNetworkTopologyLabels returns a boolean if a field has been set.

### SetNetworkTopologyLabelsNil

`func (o *NodeExtended) SetNetworkTopologyLabelsNil(b bool)`

 SetNetworkTopologyLabelsNil sets the value for NetworkTopologyLabels to be an explicit nil

### UnsetNetworkTopologyLabels
`func (o *NodeExtended) UnsetNetworkTopologyLabels()`

UnsetNetworkTopologyLabels ensures that no value is present for NetworkTopologyLabels, not even an explicit nil
### GetNuma

`func (o *NodeExtended) GetNuma() NumaInfoExtended`

GetNuma returns the Numa field if non-nil, zero value otherwise.

### GetNumaOk

`func (o *NodeExtended) GetNumaOk() (*NumaInfoExtended, bool)`

GetNumaOk returns a tuple with the Numa field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNuma

`func (o *NodeExtended) SetNuma(v NumaInfoExtended)`

SetNuma sets Numa field to given value.

### HasNuma

`func (o *NodeExtended) HasNuma() bool`

HasNuma returns a boolean if a field has been set.

### GetName

`func (o *NodeExtended) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodeExtended) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodeExtended) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *NodeExtended) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodeExtended) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodeExtended) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *NodeExtended) HasId() bool`

HasId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *NodeExtended) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *NodeExtended) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *NodeExtended) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.


### GetUpdatedAt

`func (o *NodeExtended) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *NodeExtended) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *NodeExtended) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


