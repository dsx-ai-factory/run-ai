# PVCAsset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**PVCAssetSpec**](PVCAssetSpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 
**UsageTimes** | Pointer to [**NullableUsageTimesInfo**](UsageTimesInfo.md) |  | [optional] 
**Compliance** | Pointer to [**NullableComplianceInfo**](ComplianceInfo.md) |  | [optional] 
**Status** | Pointer to [**NullableAssetClusterStatusInfo**](AssetClusterStatusInfo.md) |  | [optional] 
**ClusterInfo** | Pointer to [**NullableClusterInfo**](ClusterInfo.md) |  | [optional] 

## Methods

### NewPVCAsset

`func NewPVCAsset(meta AssetMeta, spec PVCAssetSpec, ) *PVCAsset`

NewPVCAsset instantiates a new PVCAsset object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCAssetWithDefaults

`func NewPVCAssetWithDefaults() *PVCAsset`

NewPVCAssetWithDefaults instantiates a new PVCAsset object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *PVCAsset) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *PVCAsset) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *PVCAsset) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *PVCAsset) GetSpec() PVCAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *PVCAsset) GetSpecOk() (*PVCAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *PVCAsset) SetSpec(v PVCAssetSpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *PVCAsset) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *PVCAsset) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *PVCAsset) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *PVCAsset) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *PVCAsset) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *PVCAsset) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil
### GetUsageTimes

`func (o *PVCAsset) GetUsageTimes() UsageTimesInfo`

GetUsageTimes returns the UsageTimes field if non-nil, zero value otherwise.

### GetUsageTimesOk

`func (o *PVCAsset) GetUsageTimesOk() (*UsageTimesInfo, bool)`

GetUsageTimesOk returns a tuple with the UsageTimes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsageTimes

`func (o *PVCAsset) SetUsageTimes(v UsageTimesInfo)`

SetUsageTimes sets UsageTimes field to given value.

### HasUsageTimes

`func (o *PVCAsset) HasUsageTimes() bool`

HasUsageTimes returns a boolean if a field has been set.

### SetUsageTimesNil

`func (o *PVCAsset) SetUsageTimesNil(b bool)`

 SetUsageTimesNil sets the value for UsageTimes to be an explicit nil

### UnsetUsageTimes
`func (o *PVCAsset) UnsetUsageTimes()`

UnsetUsageTimes ensures that no value is present for UsageTimes, not even an explicit nil
### GetCompliance

`func (o *PVCAsset) GetCompliance() ComplianceInfo`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *PVCAsset) GetComplianceOk() (*ComplianceInfo, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *PVCAsset) SetCompliance(v ComplianceInfo)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *PVCAsset) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### SetComplianceNil

`func (o *PVCAsset) SetComplianceNil(b bool)`

 SetComplianceNil sets the value for Compliance to be an explicit nil

### UnsetCompliance
`func (o *PVCAsset) UnsetCompliance()`

UnsetCompliance ensures that no value is present for Compliance, not even an explicit nil
### GetStatus

`func (o *PVCAsset) GetStatus() AssetClusterStatusInfo`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PVCAsset) GetStatusOk() (*AssetClusterStatusInfo, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PVCAsset) SetStatus(v AssetClusterStatusInfo)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *PVCAsset) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *PVCAsset) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *PVCAsset) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetClusterInfo

`func (o *PVCAsset) GetClusterInfo() ClusterInfo`

GetClusterInfo returns the ClusterInfo field if non-nil, zero value otherwise.

### GetClusterInfoOk

`func (o *PVCAsset) GetClusterInfoOk() (*ClusterInfo, bool)`

GetClusterInfoOk returns a tuple with the ClusterInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterInfo

`func (o *PVCAsset) SetClusterInfo(v ClusterInfo)`

SetClusterInfo sets ClusterInfo field to given value.

### HasClusterInfo

`func (o *PVCAsset) HasClusterInfo() bool`

HasClusterInfo returns a boolean if a field has been set.

### SetClusterInfoNil

`func (o *PVCAsset) SetClusterInfoNil(b bool)`

 SetClusterInfoNil sets the value for ClusterInfo to be an explicit nil

### UnsetClusterInfo
`func (o *PVCAsset) UnsetClusterInfo()`

UnsetClusterInfo ensures that no value is present for ClusterInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


