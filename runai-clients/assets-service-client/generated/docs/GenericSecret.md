# GenericSecret

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**GenericSecretSpec**](GenericSecretSpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 
**Compliance** | Pointer to [**NullableComplianceInfo**](ComplianceInfo.md) |  | [optional] 
**Status** | Pointer to [**NullableAssetClusterStatusInfo**](AssetClusterStatusInfo.md) |  | [optional] 
**ClusterInfo** | Pointer to [**NullableClusterInfo**](ClusterInfo.md) |  | [optional] 

## Methods

### NewGenericSecret

`func NewGenericSecret(meta AssetMeta, spec GenericSecretSpec, ) *GenericSecret`

NewGenericSecret instantiates a new GenericSecret object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGenericSecretWithDefaults

`func NewGenericSecretWithDefaults() *GenericSecret`

NewGenericSecretWithDefaults instantiates a new GenericSecret object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *GenericSecret) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *GenericSecret) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *GenericSecret) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *GenericSecret) GetSpec() GenericSecretSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *GenericSecret) GetSpecOk() (*GenericSecretSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *GenericSecret) SetSpec(v GenericSecretSpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *GenericSecret) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *GenericSecret) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *GenericSecret) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *GenericSecret) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *GenericSecret) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *GenericSecret) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil
### GetCompliance

`func (o *GenericSecret) GetCompliance() ComplianceInfo`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *GenericSecret) GetComplianceOk() (*ComplianceInfo, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *GenericSecret) SetCompliance(v ComplianceInfo)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *GenericSecret) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### SetComplianceNil

`func (o *GenericSecret) SetComplianceNil(b bool)`

 SetComplianceNil sets the value for Compliance to be an explicit nil

### UnsetCompliance
`func (o *GenericSecret) UnsetCompliance()`

UnsetCompliance ensures that no value is present for Compliance, not even an explicit nil
### GetStatus

`func (o *GenericSecret) GetStatus() AssetClusterStatusInfo`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *GenericSecret) GetStatusOk() (*AssetClusterStatusInfo, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *GenericSecret) SetStatus(v AssetClusterStatusInfo)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *GenericSecret) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *GenericSecret) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *GenericSecret) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetClusterInfo

`func (o *GenericSecret) GetClusterInfo() ClusterInfo`

GetClusterInfo returns the ClusterInfo field if non-nil, zero value otherwise.

### GetClusterInfoOk

`func (o *GenericSecret) GetClusterInfoOk() (*ClusterInfo, bool)`

GetClusterInfoOk returns a tuple with the ClusterInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterInfo

`func (o *GenericSecret) SetClusterInfo(v ClusterInfo)`

SetClusterInfo sets ClusterInfo field to given value.

### HasClusterInfo

`func (o *GenericSecret) HasClusterInfo() bool`

HasClusterInfo returns a boolean if a field has been set.

### SetClusterInfoNil

`func (o *GenericSecret) SetClusterInfoNil(b bool)`

 SetClusterInfoNil sets the value for ClusterInfo to be an explicit nil

### UnsetClusterInfo
`func (o *GenericSecret) UnsetClusterInfo()`

UnsetClusterInfo ensures that no value is present for ClusterInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


