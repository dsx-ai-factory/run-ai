# DockerRegistry

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**DockerRegistrySpec**](DockerRegistrySpec.md) |  | 
**UsedBy** | Pointer to [**NullableAssetUsageInfo**](AssetUsageInfo.md) |  | [optional] 
**Compliance** | Pointer to [**NullableComplianceInfo**](ComplianceInfo.md) |  | [optional] 
**Status** | Pointer to [**NullableAssetClusterStatusInfo**](AssetClusterStatusInfo.md) |  | [optional] 
**ClusterInfo** | Pointer to [**NullableClusterInfo**](ClusterInfo.md) |  | [optional] 

## Methods

### NewDockerRegistry

`func NewDockerRegistry(meta AssetMeta, spec DockerRegistrySpec, ) *DockerRegistry`

NewDockerRegistry instantiates a new DockerRegistry object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDockerRegistryWithDefaults

`func NewDockerRegistryWithDefaults() *DockerRegistry`

NewDockerRegistryWithDefaults instantiates a new DockerRegistry object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DockerRegistry) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DockerRegistry) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DockerRegistry) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *DockerRegistry) GetSpec() DockerRegistrySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DockerRegistry) GetSpecOk() (*DockerRegistrySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DockerRegistry) SetSpec(v DockerRegistrySpec)`

SetSpec sets Spec field to given value.


### GetUsedBy

`func (o *DockerRegistry) GetUsedBy() AssetUsageInfo`

GetUsedBy returns the UsedBy field if non-nil, zero value otherwise.

### GetUsedByOk

`func (o *DockerRegistry) GetUsedByOk() (*AssetUsageInfo, bool)`

GetUsedByOk returns a tuple with the UsedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsedBy

`func (o *DockerRegistry) SetUsedBy(v AssetUsageInfo)`

SetUsedBy sets UsedBy field to given value.

### HasUsedBy

`func (o *DockerRegistry) HasUsedBy() bool`

HasUsedBy returns a boolean if a field has been set.

### SetUsedByNil

`func (o *DockerRegistry) SetUsedByNil(b bool)`

 SetUsedByNil sets the value for UsedBy to be an explicit nil

### UnsetUsedBy
`func (o *DockerRegistry) UnsetUsedBy()`

UnsetUsedBy ensures that no value is present for UsedBy, not even an explicit nil
### GetCompliance

`func (o *DockerRegistry) GetCompliance() ComplianceInfo`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *DockerRegistry) GetComplianceOk() (*ComplianceInfo, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *DockerRegistry) SetCompliance(v ComplianceInfo)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *DockerRegistry) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### SetComplianceNil

`func (o *DockerRegistry) SetComplianceNil(b bool)`

 SetComplianceNil sets the value for Compliance to be an explicit nil

### UnsetCompliance
`func (o *DockerRegistry) UnsetCompliance()`

UnsetCompliance ensures that no value is present for Compliance, not even an explicit nil
### GetStatus

`func (o *DockerRegistry) GetStatus() AssetClusterStatusInfo`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DockerRegistry) GetStatusOk() (*AssetClusterStatusInfo, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DockerRegistry) SetStatus(v AssetClusterStatusInfo)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DockerRegistry) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *DockerRegistry) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *DockerRegistry) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetClusterInfo

`func (o *DockerRegistry) GetClusterInfo() ClusterInfo`

GetClusterInfo returns the ClusterInfo field if non-nil, zero value otherwise.

### GetClusterInfoOk

`func (o *DockerRegistry) GetClusterInfoOk() (*ClusterInfo, bool)`

GetClusterInfoOk returns a tuple with the ClusterInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterInfo

`func (o *DockerRegistry) SetClusterInfo(v ClusterInfo)`

SetClusterInfo sets ClusterInfo field to given value.

### HasClusterInfo

`func (o *DockerRegistry) HasClusterInfo() bool`

HasClusterInfo returns a boolean if a field has been set.

### SetClusterInfoNil

`func (o *DockerRegistry) SetClusterInfoNil(b bool)`

 SetClusterInfoNil sets the value for ClusterInfo to be an explicit nil

### UnsetClusterInfo
`func (o *DockerRegistry) UnsetClusterInfo()`

UnsetClusterInfo ensures that no value is present for ClusterInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


