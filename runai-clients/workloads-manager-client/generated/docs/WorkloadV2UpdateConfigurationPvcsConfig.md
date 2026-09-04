# WorkloadV2UpdateConfigurationPvcsConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CleanupOnCompletion** | Pointer to **NullableBool** | Delete the workload-owned PVCs when the workload reaches a terminal state. When unset, the cluster default behavior applies.  | [optional] 

## Methods

### NewWorkloadV2UpdateConfigurationPvcsConfig

`func NewWorkloadV2UpdateConfigurationPvcsConfig() *WorkloadV2UpdateConfigurationPvcsConfig`

NewWorkloadV2UpdateConfigurationPvcsConfig instantiates a new WorkloadV2UpdateConfigurationPvcsConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2UpdateConfigurationPvcsConfigWithDefaults

`func NewWorkloadV2UpdateConfigurationPvcsConfigWithDefaults() *WorkloadV2UpdateConfigurationPvcsConfig`

NewWorkloadV2UpdateConfigurationPvcsConfigWithDefaults instantiates a new WorkloadV2UpdateConfigurationPvcsConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCleanupOnCompletion

`func (o *WorkloadV2UpdateConfigurationPvcsConfig) GetCleanupOnCompletion() bool`

GetCleanupOnCompletion returns the CleanupOnCompletion field if non-nil, zero value otherwise.

### GetCleanupOnCompletionOk

`func (o *WorkloadV2UpdateConfigurationPvcsConfig) GetCleanupOnCompletionOk() (*bool, bool)`

GetCleanupOnCompletionOk returns a tuple with the CleanupOnCompletion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCleanupOnCompletion

`func (o *WorkloadV2UpdateConfigurationPvcsConfig) SetCleanupOnCompletion(v bool)`

SetCleanupOnCompletion sets CleanupOnCompletion field to given value.

### HasCleanupOnCompletion

`func (o *WorkloadV2UpdateConfigurationPvcsConfig) HasCleanupOnCompletion() bool`

HasCleanupOnCompletion returns a boolean if a field has been set.

### SetCleanupOnCompletionNil

`func (o *WorkloadV2UpdateConfigurationPvcsConfig) SetCleanupOnCompletionNil(b bool)`

 SetCleanupOnCompletionNil sets the value for CleanupOnCompletion to be an explicit nil

### UnsetCleanupOnCompletion
`func (o *WorkloadV2UpdateConfigurationPvcsConfig) UnsetCleanupOnCompletion()`

UnsetCleanupOnCompletion ensures that no value is present for CleanupOnCompletion, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


