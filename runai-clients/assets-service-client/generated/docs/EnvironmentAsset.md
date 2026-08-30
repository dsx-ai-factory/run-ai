# EnvironmentAsset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**EnvironmentAssetSpec**](EnvironmentAssetSpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 
**UsageTimes** | Pointer to [**NullableUsageTimesInfo**](UsageTimesInfo.md) |  | [optional] 
**Compliance** | Pointer to [**NullableComplianceInfo**](ComplianceInfo.md) |  | [optional] 
**Status** | Pointer to [**NullableAssetClusterStatusInfo**](AssetClusterStatusInfo.md) |  | [optional] 

## Methods

### NewEnvironmentAsset

`func NewEnvironmentAsset(meta AssetMeta, spec EnvironmentAssetSpec, ) *EnvironmentAsset`

NewEnvironmentAsset instantiates a new EnvironmentAsset object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentAssetWithDefaults

`func NewEnvironmentAssetWithDefaults() *EnvironmentAsset`

NewEnvironmentAssetWithDefaults instantiates a new EnvironmentAsset object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *EnvironmentAsset) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *EnvironmentAsset) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *EnvironmentAsset) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *EnvironmentAsset) GetSpec() EnvironmentAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *EnvironmentAsset) GetSpecOk() (*EnvironmentAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *EnvironmentAsset) SetSpec(v EnvironmentAssetSpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *EnvironmentAsset) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *EnvironmentAsset) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *EnvironmentAsset) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *EnvironmentAsset) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *EnvironmentAsset) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *EnvironmentAsset) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil
### GetUsageTimes

`func (o *EnvironmentAsset) GetUsageTimes() UsageTimesInfo`

GetUsageTimes returns the UsageTimes field if non-nil, zero value otherwise.

### GetUsageTimesOk

`func (o *EnvironmentAsset) GetUsageTimesOk() (*UsageTimesInfo, bool)`

GetUsageTimesOk returns a tuple with the UsageTimes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsageTimes

`func (o *EnvironmentAsset) SetUsageTimes(v UsageTimesInfo)`

SetUsageTimes sets UsageTimes field to given value.

### HasUsageTimes

`func (o *EnvironmentAsset) HasUsageTimes() bool`

HasUsageTimes returns a boolean if a field has been set.

### SetUsageTimesNil

`func (o *EnvironmentAsset) SetUsageTimesNil(b bool)`

 SetUsageTimesNil sets the value for UsageTimes to be an explicit nil

### UnsetUsageTimes
`func (o *EnvironmentAsset) UnsetUsageTimes()`

UnsetUsageTimes ensures that no value is present for UsageTimes, not even an explicit nil
### GetCompliance

`func (o *EnvironmentAsset) GetCompliance() ComplianceInfo`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *EnvironmentAsset) GetComplianceOk() (*ComplianceInfo, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *EnvironmentAsset) SetCompliance(v ComplianceInfo)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *EnvironmentAsset) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### SetComplianceNil

`func (o *EnvironmentAsset) SetComplianceNil(b bool)`

 SetComplianceNil sets the value for Compliance to be an explicit nil

### UnsetCompliance
`func (o *EnvironmentAsset) UnsetCompliance()`

UnsetCompliance ensures that no value is present for Compliance, not even an explicit nil
### GetStatus

`func (o *EnvironmentAsset) GetStatus() AssetClusterStatusInfo`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *EnvironmentAsset) GetStatusOk() (*AssetClusterStatusInfo, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *EnvironmentAsset) SetStatus(v AssetClusterStatusInfo)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *EnvironmentAsset) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *EnvironmentAsset) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *EnvironmentAsset) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


