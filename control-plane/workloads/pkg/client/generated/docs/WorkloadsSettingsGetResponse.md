# WorkloadsSettingsGetResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EnablePolicies** | Pointer to [**GetEnablePolicies**](GetEnablePolicies.md) |  | [optional] 
**WorkloadSubmissionV2Enabled** | Pointer to [**GetWorkloadSubmissionV2Enabled**](GetWorkloadSubmissionV2Enabled.md) |  | [optional] 
**EnableTemplatesV2** | Pointer to [**GetEnableTemplatesV2**](GetEnableTemplatesV2.md) |  | [optional] 
**EnableHybridSubmission** | Pointer to [**GetEnableHybridSubmission**](GetEnableHybridSubmission.md) |  | [optional] 
**GpuOverProvisioning** | Pointer to [**GetGpuOverProvisioning**](GetGpuOverProvisioning.md) |  | [optional] 
**ApplicationSubmission** | Pointer to [**GetApplicationSubmission**](GetApplicationSubmission.md) |  | [optional] 

## Methods

### NewWorkloadsSettingsGetResponse

`func NewWorkloadsSettingsGetResponse() *WorkloadsSettingsGetResponse`

NewWorkloadsSettingsGetResponse instantiates a new WorkloadsSettingsGetResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadsSettingsGetResponseWithDefaults

`func NewWorkloadsSettingsGetResponseWithDefaults() *WorkloadsSettingsGetResponse`

NewWorkloadsSettingsGetResponseWithDefaults instantiates a new WorkloadsSettingsGetResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnablePolicies

`func (o *WorkloadsSettingsGetResponse) GetEnablePolicies() GetEnablePolicies`

GetEnablePolicies returns the EnablePolicies field if non-nil, zero value otherwise.

### GetEnablePoliciesOk

`func (o *WorkloadsSettingsGetResponse) GetEnablePoliciesOk() (*GetEnablePolicies, bool)`

GetEnablePoliciesOk returns a tuple with the EnablePolicies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnablePolicies

`func (o *WorkloadsSettingsGetResponse) SetEnablePolicies(v GetEnablePolicies)`

SetEnablePolicies sets EnablePolicies field to given value.

### HasEnablePolicies

`func (o *WorkloadsSettingsGetResponse) HasEnablePolicies() bool`

HasEnablePolicies returns a boolean if a field has been set.

### GetWorkloadSubmissionV2Enabled

`func (o *WorkloadsSettingsGetResponse) GetWorkloadSubmissionV2Enabled() GetWorkloadSubmissionV2Enabled`

GetWorkloadSubmissionV2Enabled returns the WorkloadSubmissionV2Enabled field if non-nil, zero value otherwise.

### GetWorkloadSubmissionV2EnabledOk

`func (o *WorkloadsSettingsGetResponse) GetWorkloadSubmissionV2EnabledOk() (*GetWorkloadSubmissionV2Enabled, bool)`

GetWorkloadSubmissionV2EnabledOk returns a tuple with the WorkloadSubmissionV2Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadSubmissionV2Enabled

`func (o *WorkloadsSettingsGetResponse) SetWorkloadSubmissionV2Enabled(v GetWorkloadSubmissionV2Enabled)`

SetWorkloadSubmissionV2Enabled sets WorkloadSubmissionV2Enabled field to given value.

### HasWorkloadSubmissionV2Enabled

`func (o *WorkloadsSettingsGetResponse) HasWorkloadSubmissionV2Enabled() bool`

HasWorkloadSubmissionV2Enabled returns a boolean if a field has been set.

### GetEnableTemplatesV2

`func (o *WorkloadsSettingsGetResponse) GetEnableTemplatesV2() GetEnableTemplatesV2`

GetEnableTemplatesV2 returns the EnableTemplatesV2 field if non-nil, zero value otherwise.

### GetEnableTemplatesV2Ok

`func (o *WorkloadsSettingsGetResponse) GetEnableTemplatesV2Ok() (*GetEnableTemplatesV2, bool)`

GetEnableTemplatesV2Ok returns a tuple with the EnableTemplatesV2 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnableTemplatesV2

`func (o *WorkloadsSettingsGetResponse) SetEnableTemplatesV2(v GetEnableTemplatesV2)`

SetEnableTemplatesV2 sets EnableTemplatesV2 field to given value.

### HasEnableTemplatesV2

`func (o *WorkloadsSettingsGetResponse) HasEnableTemplatesV2() bool`

HasEnableTemplatesV2 returns a boolean if a field has been set.

### GetEnableHybridSubmission

`func (o *WorkloadsSettingsGetResponse) GetEnableHybridSubmission() GetEnableHybridSubmission`

GetEnableHybridSubmission returns the EnableHybridSubmission field if non-nil, zero value otherwise.

### GetEnableHybridSubmissionOk

`func (o *WorkloadsSettingsGetResponse) GetEnableHybridSubmissionOk() (*GetEnableHybridSubmission, bool)`

GetEnableHybridSubmissionOk returns a tuple with the EnableHybridSubmission field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnableHybridSubmission

`func (o *WorkloadsSettingsGetResponse) SetEnableHybridSubmission(v GetEnableHybridSubmission)`

SetEnableHybridSubmission sets EnableHybridSubmission field to given value.

### HasEnableHybridSubmission

`func (o *WorkloadsSettingsGetResponse) HasEnableHybridSubmission() bool`

HasEnableHybridSubmission returns a boolean if a field has been set.

### GetGpuOverProvisioning

`func (o *WorkloadsSettingsGetResponse) GetGpuOverProvisioning() GetGpuOverProvisioning`

GetGpuOverProvisioning returns the GpuOverProvisioning field if non-nil, zero value otherwise.

### GetGpuOverProvisioningOk

`func (o *WorkloadsSettingsGetResponse) GetGpuOverProvisioningOk() (*GetGpuOverProvisioning, bool)`

GetGpuOverProvisioningOk returns a tuple with the GpuOverProvisioning field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuOverProvisioning

`func (o *WorkloadsSettingsGetResponse) SetGpuOverProvisioning(v GetGpuOverProvisioning)`

SetGpuOverProvisioning sets GpuOverProvisioning field to given value.

### HasGpuOverProvisioning

`func (o *WorkloadsSettingsGetResponse) HasGpuOverProvisioning() bool`

HasGpuOverProvisioning returns a boolean if a field has been set.

### GetApplicationSubmission

`func (o *WorkloadsSettingsGetResponse) GetApplicationSubmission() GetApplicationSubmission`

GetApplicationSubmission returns the ApplicationSubmission field if non-nil, zero value otherwise.

### GetApplicationSubmissionOk

`func (o *WorkloadsSettingsGetResponse) GetApplicationSubmissionOk() (*GetApplicationSubmission, bool)`

GetApplicationSubmissionOk returns a tuple with the ApplicationSubmission field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplicationSubmission

`func (o *WorkloadsSettingsGetResponse) SetApplicationSubmission(v GetApplicationSubmission)`

SetApplicationSubmission sets ApplicationSubmission field to given value.

### HasApplicationSubmission

`func (o *WorkloadsSettingsGetResponse) HasApplicationSubmission() bool`

HasApplicationSubmission returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


