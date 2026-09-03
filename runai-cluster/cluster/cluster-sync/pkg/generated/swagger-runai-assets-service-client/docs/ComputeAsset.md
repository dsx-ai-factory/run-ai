# ComputeAsset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**ComputeAssetSpec**](ComputeAssetSpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 
**UsageTimes** | Pointer to [**NullableUsageTimesInfo**](UsageTimesInfo.md) |  | [optional] 
**Compliance** | Pointer to [**NullableComplianceInfo**](ComplianceInfo.md) |  | [optional] 

## Methods

### NewComputeAsset

`func NewComputeAsset(meta AssetMeta, spec ComputeAssetSpec, ) *ComputeAsset`

NewComputeAsset instantiates a new ComputeAsset object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComputeAssetWithDefaults

`func NewComputeAssetWithDefaults() *ComputeAsset`

NewComputeAssetWithDefaults instantiates a new ComputeAsset object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *ComputeAsset) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *ComputeAsset) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *ComputeAsset) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *ComputeAsset) GetSpec() ComputeAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ComputeAsset) GetSpecOk() (*ComputeAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ComputeAsset) SetSpec(v ComputeAssetSpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *ComputeAsset) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *ComputeAsset) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *ComputeAsset) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *ComputeAsset) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *ComputeAsset) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *ComputeAsset) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil
### GetUsageTimes

`func (o *ComputeAsset) GetUsageTimes() UsageTimesInfo`

GetUsageTimes returns the UsageTimes field if non-nil, zero value otherwise.

### GetUsageTimesOk

`func (o *ComputeAsset) GetUsageTimesOk() (*UsageTimesInfo, bool)`

GetUsageTimesOk returns a tuple with the UsageTimes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsageTimes

`func (o *ComputeAsset) SetUsageTimes(v UsageTimesInfo)`

SetUsageTimes sets UsageTimes field to given value.

### HasUsageTimes

`func (o *ComputeAsset) HasUsageTimes() bool`

HasUsageTimes returns a boolean if a field has been set.

### SetUsageTimesNil

`func (o *ComputeAsset) SetUsageTimesNil(b bool)`

 SetUsageTimesNil sets the value for UsageTimes to be an explicit nil

### UnsetUsageTimes
`func (o *ComputeAsset) UnsetUsageTimes()`

UnsetUsageTimes ensures that no value is present for UsageTimes, not even an explicit nil
### GetCompliance

`func (o *ComputeAsset) GetCompliance() ComplianceInfo`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *ComputeAsset) GetComplianceOk() (*ComplianceInfo, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *ComputeAsset) SetCompliance(v ComplianceInfo)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *ComputeAsset) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### SetComplianceNil

`func (o *ComputeAsset) SetComplianceNil(b bool)`

 SetComplianceNil sets the value for Compliance to be an explicit nil

### UnsetCompliance
`func (o *ComputeAsset) UnsetCompliance()`

UnsetCompliance ensures that no value is present for Compliance, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


