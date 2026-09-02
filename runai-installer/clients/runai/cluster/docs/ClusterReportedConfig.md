# ClusterReportedConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadOwnershipProtection** | Pointer to **NullableBool** | indicates if the cluster workload ownership protection is enabled | [optional] 
**SubdomainEnabled** | Pointer to **NullableBool** | indicates if the cluster subdomain is enabled | [optional] 
**IngressClass** | Pointer to **NullableString** | the ingress class that is configured for the cluster | [optional] 

## Methods

### NewClusterReportedConfig

`func NewClusterReportedConfig() *ClusterReportedConfig`

NewClusterReportedConfig instantiates a new ClusterReportedConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedConfigWithDefaults

`func NewClusterReportedConfigWithDefaults() *ClusterReportedConfig`

NewClusterReportedConfigWithDefaults instantiates a new ClusterReportedConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadOwnershipProtection

`func (o *ClusterReportedConfig) GetWorkloadOwnershipProtection() bool`

GetWorkloadOwnershipProtection returns the WorkloadOwnershipProtection field if non-nil, zero value otherwise.

### GetWorkloadOwnershipProtectionOk

`func (o *ClusterReportedConfig) GetWorkloadOwnershipProtectionOk() (*bool, bool)`

GetWorkloadOwnershipProtectionOk returns a tuple with the WorkloadOwnershipProtection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadOwnershipProtection

`func (o *ClusterReportedConfig) SetWorkloadOwnershipProtection(v bool)`

SetWorkloadOwnershipProtection sets WorkloadOwnershipProtection field to given value.

### HasWorkloadOwnershipProtection

`func (o *ClusterReportedConfig) HasWorkloadOwnershipProtection() bool`

HasWorkloadOwnershipProtection returns a boolean if a field has been set.

### SetWorkloadOwnershipProtectionNil

`func (o *ClusterReportedConfig) SetWorkloadOwnershipProtectionNil(b bool)`

 SetWorkloadOwnershipProtectionNil sets the value for WorkloadOwnershipProtection to be an explicit nil

### UnsetWorkloadOwnershipProtection
`func (o *ClusterReportedConfig) UnsetWorkloadOwnershipProtection()`

UnsetWorkloadOwnershipProtection ensures that no value is present for WorkloadOwnershipProtection, not even an explicit nil
### GetSubdomainEnabled

`func (o *ClusterReportedConfig) GetSubdomainEnabled() bool`

GetSubdomainEnabled returns the SubdomainEnabled field if non-nil, zero value otherwise.

### GetSubdomainEnabledOk

`func (o *ClusterReportedConfig) GetSubdomainEnabledOk() (*bool, bool)`

GetSubdomainEnabledOk returns a tuple with the SubdomainEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubdomainEnabled

`func (o *ClusterReportedConfig) SetSubdomainEnabled(v bool)`

SetSubdomainEnabled sets SubdomainEnabled field to given value.

### HasSubdomainEnabled

`func (o *ClusterReportedConfig) HasSubdomainEnabled() bool`

HasSubdomainEnabled returns a boolean if a field has been set.

### SetSubdomainEnabledNil

`func (o *ClusterReportedConfig) SetSubdomainEnabledNil(b bool)`

 SetSubdomainEnabledNil sets the value for SubdomainEnabled to be an explicit nil

### UnsetSubdomainEnabled
`func (o *ClusterReportedConfig) UnsetSubdomainEnabled()`

UnsetSubdomainEnabled ensures that no value is present for SubdomainEnabled, not even an explicit nil
### GetIngressClass

`func (o *ClusterReportedConfig) GetIngressClass() string`

GetIngressClass returns the IngressClass field if non-nil, zero value otherwise.

### GetIngressClassOk

`func (o *ClusterReportedConfig) GetIngressClassOk() (*string, bool)`

GetIngressClassOk returns a tuple with the IngressClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIngressClass

`func (o *ClusterReportedConfig) SetIngressClass(v string)`

SetIngressClass sets IngressClass field to given value.

### HasIngressClass

`func (o *ClusterReportedConfig) HasIngressClass() bool`

HasIngressClass returns a boolean if a field has been set.

### SetIngressClassNil

`func (o *ClusterReportedConfig) SetIngressClassNil(b bool)`

 SetIngressClassNil sets the value for IngressClass to be an explicit nil

### UnsetIngressClass
`func (o *ClusterReportedConfig) UnsetIngressClass()`

UnsetIngressClass ensures that no value is present for IngressClass, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


