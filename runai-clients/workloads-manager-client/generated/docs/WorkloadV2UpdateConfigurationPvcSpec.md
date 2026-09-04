# WorkloadV2UpdateConfigurationPvcSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessModes** | [**[]WorkloadV2UpdateConfigurationPvcAccessMode**](WorkloadV2UpdateConfigurationPvcAccessMode.md) | Access modes the volume should support. At least one is required. | 
**Resources** | [**WorkloadV2UpdateConfigurationPvcResources**](WorkloadV2UpdateConfigurationPvcResources.md) |  | 
**StorageClassName** | Pointer to **NullableString** | Name of the StorageClass to use. When omitted, the cluster&#39;s default StorageClass is used (if one is configured).  | [optional] 
**VolumeMode** | Pointer to [**NullableWorkloadV2UpdateConfigurationPvcVolumeMode**](WorkloadV2UpdateConfigurationPvcVolumeMode.md) |  | [optional] 

## Methods

### NewWorkloadV2UpdateConfigurationPvcSpec

`func NewWorkloadV2UpdateConfigurationPvcSpec(accessModes []WorkloadV2UpdateConfigurationPvcAccessMode, resources WorkloadV2UpdateConfigurationPvcResources, ) *WorkloadV2UpdateConfigurationPvcSpec`

NewWorkloadV2UpdateConfigurationPvcSpec instantiates a new WorkloadV2UpdateConfigurationPvcSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2UpdateConfigurationPvcSpecWithDefaults

`func NewWorkloadV2UpdateConfigurationPvcSpecWithDefaults() *WorkloadV2UpdateConfigurationPvcSpec`

NewWorkloadV2UpdateConfigurationPvcSpecWithDefaults instantiates a new WorkloadV2UpdateConfigurationPvcSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessModes

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetAccessModes() []WorkloadV2UpdateConfigurationPvcAccessMode`

GetAccessModes returns the AccessModes field if non-nil, zero value otherwise.

### GetAccessModesOk

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetAccessModesOk() (*[]WorkloadV2UpdateConfigurationPvcAccessMode, bool)`

GetAccessModesOk returns a tuple with the AccessModes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessModes

`func (o *WorkloadV2UpdateConfigurationPvcSpec) SetAccessModes(v []WorkloadV2UpdateConfigurationPvcAccessMode)`

SetAccessModes sets AccessModes field to given value.


### GetResources

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetResources() WorkloadV2UpdateConfigurationPvcResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetResourcesOk() (*WorkloadV2UpdateConfigurationPvcResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *WorkloadV2UpdateConfigurationPvcSpec) SetResources(v WorkloadV2UpdateConfigurationPvcResources)`

SetResources sets Resources field to given value.


### GetStorageClassName

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetStorageClassName() string`

GetStorageClassName returns the StorageClassName field if non-nil, zero value otherwise.

### GetStorageClassNameOk

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetStorageClassNameOk() (*string, bool)`

GetStorageClassNameOk returns a tuple with the StorageClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassName

`func (o *WorkloadV2UpdateConfigurationPvcSpec) SetStorageClassName(v string)`

SetStorageClassName sets StorageClassName field to given value.

### HasStorageClassName

`func (o *WorkloadV2UpdateConfigurationPvcSpec) HasStorageClassName() bool`

HasStorageClassName returns a boolean if a field has been set.

### SetStorageClassNameNil

`func (o *WorkloadV2UpdateConfigurationPvcSpec) SetStorageClassNameNil(b bool)`

 SetStorageClassNameNil sets the value for StorageClassName to be an explicit nil

### UnsetStorageClassName
`func (o *WorkloadV2UpdateConfigurationPvcSpec) UnsetStorageClassName()`

UnsetStorageClassName ensures that no value is present for StorageClassName, not even an explicit nil
### GetVolumeMode

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetVolumeMode() WorkloadV2UpdateConfigurationPvcVolumeMode`

GetVolumeMode returns the VolumeMode field if non-nil, zero value otherwise.

### GetVolumeModeOk

`func (o *WorkloadV2UpdateConfigurationPvcSpec) GetVolumeModeOk() (*WorkloadV2UpdateConfigurationPvcVolumeMode, bool)`

GetVolumeModeOk returns a tuple with the VolumeMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeMode

`func (o *WorkloadV2UpdateConfigurationPvcSpec) SetVolumeMode(v WorkloadV2UpdateConfigurationPvcVolumeMode)`

SetVolumeMode sets VolumeMode field to given value.

### HasVolumeMode

`func (o *WorkloadV2UpdateConfigurationPvcSpec) HasVolumeMode() bool`

HasVolumeMode returns a boolean if a field has been set.

### SetVolumeModeNil

`func (o *WorkloadV2UpdateConfigurationPvcSpec) SetVolumeModeNil(b bool)`

 SetVolumeModeNil sets the value for VolumeMode to be an explicit nil

### UnsetVolumeMode
`func (o *WorkloadV2UpdateConfigurationPvcSpec) UnsetVolumeMode()`

UnsetVolumeMode ensures that no value is present for VolumeMode, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


