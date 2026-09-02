# NodePoolSettingsGetResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectManualOverQuota** | Pointer to [**NodePoolSetting**](NodePoolSetting.md) |  | [optional] 
**CpuResourcesQuota** | Pointer to [**NodePoolSetting**](NodePoolSetting.md) |  | [optional] 

## Methods

### NewNodePoolSettingsGetResponse

`func NewNodePoolSettingsGetResponse() *NodePoolSettingsGetResponse`

NewNodePoolSettingsGetResponse instantiates a new NodePoolSettingsGetResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolSettingsGetResponseWithDefaults

`func NewNodePoolSettingsGetResponseWithDefaults() *NodePoolSettingsGetResponse`

NewNodePoolSettingsGetResponseWithDefaults instantiates a new NodePoolSettingsGetResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectManualOverQuota

`func (o *NodePoolSettingsGetResponse) GetProjectManualOverQuota() NodePoolSetting`

GetProjectManualOverQuota returns the ProjectManualOverQuota field if non-nil, zero value otherwise.

### GetProjectManualOverQuotaOk

`func (o *NodePoolSettingsGetResponse) GetProjectManualOverQuotaOk() (*NodePoolSetting, bool)`

GetProjectManualOverQuotaOk returns a tuple with the ProjectManualOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectManualOverQuota

`func (o *NodePoolSettingsGetResponse) SetProjectManualOverQuota(v NodePoolSetting)`

SetProjectManualOverQuota sets ProjectManualOverQuota field to given value.

### HasProjectManualOverQuota

`func (o *NodePoolSettingsGetResponse) HasProjectManualOverQuota() bool`

HasProjectManualOverQuota returns a boolean if a field has been set.

### GetCpuResourcesQuota

`func (o *NodePoolSettingsGetResponse) GetCpuResourcesQuota() NodePoolSetting`

GetCpuResourcesQuota returns the CpuResourcesQuota field if non-nil, zero value otherwise.

### GetCpuResourcesQuotaOk

`func (o *NodePoolSettingsGetResponse) GetCpuResourcesQuotaOk() (*NodePoolSetting, bool)`

GetCpuResourcesQuotaOk returns a tuple with the CpuResourcesQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuResourcesQuota

`func (o *NodePoolSettingsGetResponse) SetCpuResourcesQuota(v NodePoolSetting)`

SetCpuResourcesQuota sets CpuResourcesQuota field to given value.

### HasCpuResourcesQuota

`func (o *NodePoolSettingsGetResponse) HasCpuResourcesQuota() bool`

HasCpuResourcesQuota returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


