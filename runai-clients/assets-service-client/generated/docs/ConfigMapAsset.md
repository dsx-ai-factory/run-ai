# ConfigMapAsset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**ConfigMapAssetSpec**](ConfigMapAssetSpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 
**UsageTimes** | Pointer to [**NullableUsageTimesInfo**](UsageTimesInfo.md) |  | [optional] 
**Compliance** | Pointer to [**NullableComplianceInfo**](ComplianceInfo.md) |  | [optional] 
**Status** | Pointer to [**NullableAssetClusterStatusInfo**](AssetClusterStatusInfo.md) |  | [optional] 

## Methods

### NewConfigMapAsset

`func NewConfigMapAsset(meta AssetMeta, spec ConfigMapAssetSpec, ) *ConfigMapAsset`

NewConfigMapAsset instantiates a new ConfigMapAsset object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigMapAssetWithDefaults

`func NewConfigMapAssetWithDefaults() *ConfigMapAsset`

NewConfigMapAssetWithDefaults instantiates a new ConfigMapAsset object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *ConfigMapAsset) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *ConfigMapAsset) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *ConfigMapAsset) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *ConfigMapAsset) GetSpec() ConfigMapAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ConfigMapAsset) GetSpecOk() (*ConfigMapAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ConfigMapAsset) SetSpec(v ConfigMapAssetSpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *ConfigMapAsset) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *ConfigMapAsset) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *ConfigMapAsset) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *ConfigMapAsset) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *ConfigMapAsset) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *ConfigMapAsset) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil
### GetUsageTimes

`func (o *ConfigMapAsset) GetUsageTimes() UsageTimesInfo`

GetUsageTimes returns the UsageTimes field if non-nil, zero value otherwise.

### GetUsageTimesOk

`func (o *ConfigMapAsset) GetUsageTimesOk() (*UsageTimesInfo, bool)`

GetUsageTimesOk returns a tuple with the UsageTimes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsageTimes

`func (o *ConfigMapAsset) SetUsageTimes(v UsageTimesInfo)`

SetUsageTimes sets UsageTimes field to given value.

### HasUsageTimes

`func (o *ConfigMapAsset) HasUsageTimes() bool`

HasUsageTimes returns a boolean if a field has been set.

### SetUsageTimesNil

`func (o *ConfigMapAsset) SetUsageTimesNil(b bool)`

 SetUsageTimesNil sets the value for UsageTimes to be an explicit nil

### UnsetUsageTimes
`func (o *ConfigMapAsset) UnsetUsageTimes()`

UnsetUsageTimes ensures that no value is present for UsageTimes, not even an explicit nil
### GetCompliance

`func (o *ConfigMapAsset) GetCompliance() ComplianceInfo`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *ConfigMapAsset) GetComplianceOk() (*ComplianceInfo, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *ConfigMapAsset) SetCompliance(v ComplianceInfo)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *ConfigMapAsset) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### SetComplianceNil

`func (o *ConfigMapAsset) SetComplianceNil(b bool)`

 SetComplianceNil sets the value for Compliance to be an explicit nil

### UnsetCompliance
`func (o *ConfigMapAsset) UnsetCompliance()`

UnsetCompliance ensures that no value is present for Compliance, not even an explicit nil
### GetStatus

`func (o *ConfigMapAsset) GetStatus() AssetClusterStatusInfo`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ConfigMapAsset) GetStatusOk() (*AssetClusterStatusInfo, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ConfigMapAsset) SetStatus(v AssetClusterStatusInfo)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ConfigMapAsset) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *ConfigMapAsset) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *ConfigMapAsset) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


