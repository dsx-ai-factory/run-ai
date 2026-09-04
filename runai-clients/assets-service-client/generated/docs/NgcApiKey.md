# NgcApiKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**NgcApiKeySpec**](NgcApiKeySpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 
**Compliance** | Pointer to [**NullableComplianceInfo**](ComplianceInfo.md) |  | [optional] 
**Status** | Pointer to [**NullableAssetClusterStatusInfo**](AssetClusterStatusInfo.md) |  | [optional] 
**ClusterInfo** | Pointer to [**NullableClusterInfo**](ClusterInfo.md) |  | [optional] 

## Methods

### NewNgcApiKey

`func NewNgcApiKey(meta AssetMeta, spec NgcApiKeySpec, ) *NgcApiKey`

NewNgcApiKey instantiates a new NgcApiKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNgcApiKeyWithDefaults

`func NewNgcApiKeyWithDefaults() *NgcApiKey`

NewNgcApiKeyWithDefaults instantiates a new NgcApiKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *NgcApiKey) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *NgcApiKey) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *NgcApiKey) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *NgcApiKey) GetSpec() NgcApiKeySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *NgcApiKey) GetSpecOk() (*NgcApiKeySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *NgcApiKey) SetSpec(v NgcApiKeySpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *NgcApiKey) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *NgcApiKey) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *NgcApiKey) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *NgcApiKey) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *NgcApiKey) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *NgcApiKey) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil
### GetCompliance

`func (o *NgcApiKey) GetCompliance() ComplianceInfo`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *NgcApiKey) GetComplianceOk() (*ComplianceInfo, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *NgcApiKey) SetCompliance(v ComplianceInfo)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *NgcApiKey) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### SetComplianceNil

`func (o *NgcApiKey) SetComplianceNil(b bool)`

 SetComplianceNil sets the value for Compliance to be an explicit nil

### UnsetCompliance
`func (o *NgcApiKey) UnsetCompliance()`

UnsetCompliance ensures that no value is present for Compliance, not even an explicit nil
### GetStatus

`func (o *NgcApiKey) GetStatus() AssetClusterStatusInfo`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NgcApiKey) GetStatusOk() (*AssetClusterStatusInfo, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NgcApiKey) SetStatus(v AssetClusterStatusInfo)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NgcApiKey) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *NgcApiKey) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *NgcApiKey) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetClusterInfo

`func (o *NgcApiKey) GetClusterInfo() ClusterInfo`

GetClusterInfo returns the ClusterInfo field if non-nil, zero value otherwise.

### GetClusterInfoOk

`func (o *NgcApiKey) GetClusterInfoOk() (*ClusterInfo, bool)`

GetClusterInfoOk returns a tuple with the ClusterInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterInfo

`func (o *NgcApiKey) SetClusterInfo(v ClusterInfo)`

SetClusterInfo sets ClusterInfo field to given value.

### HasClusterInfo

`func (o *NgcApiKey) HasClusterInfo() bool`

HasClusterInfo returns a boolean if a field has been set.

### SetClusterInfoNil

`func (o *NgcApiKey) SetClusterInfoNil(b bool)`

 SetClusterInfoNil sets the value for ClusterInfo to be an explicit nil

### UnsetClusterInfo
`func (o *NgcApiKey) UnsetClusterInfo()`

UnsetClusterInfo ensures that no value is present for ClusterInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


