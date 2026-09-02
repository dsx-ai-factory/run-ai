# EnvironmentVariableOfAsset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the environment variable. (mandatory) | [optional] 
**Value** | Pointer to **NullableString** | The value of the environment variable. (mutually exclusive with credential, configMap and podFieldRef) | [optional] 
**Credential** | Pointer to [**NullableEnvironmentVariableCredential**](EnvironmentVariableCredential.md) |  | [optional] 
**ConfigMap** | Pointer to [**NullableEnvironmentVariableConfigMap**](EnvironmentVariableConfigMap.md) |  | [optional] 
**PodFieldRef** | Pointer to [**NullableEnvironmentVariablePodFieldReference**](EnvironmentVariablePodFieldReference.md) |  | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the environment variable is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] 
**Description** | Pointer to **NullableString** | Description of the environment variable. | [optional] 

## Methods

### NewEnvironmentVariableOfAsset

`func NewEnvironmentVariableOfAsset() *EnvironmentVariableOfAsset`

NewEnvironmentVariableOfAsset instantiates a new EnvironmentVariableOfAsset object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentVariableOfAssetWithDefaults

`func NewEnvironmentVariableOfAssetWithDefaults() *EnvironmentVariableOfAsset`

NewEnvironmentVariableOfAssetWithDefaults instantiates a new EnvironmentVariableOfAsset object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *EnvironmentVariableOfAsset) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EnvironmentVariableOfAsset) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EnvironmentVariableOfAsset) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *EnvironmentVariableOfAsset) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *EnvironmentVariableOfAsset) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *EnvironmentVariableOfAsset) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetValue

`func (o *EnvironmentVariableOfAsset) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *EnvironmentVariableOfAsset) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *EnvironmentVariableOfAsset) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *EnvironmentVariableOfAsset) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *EnvironmentVariableOfAsset) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *EnvironmentVariableOfAsset) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetCredential

`func (o *EnvironmentVariableOfAsset) GetCredential() EnvironmentVariableCredential`

GetCredential returns the Credential field if non-nil, zero value otherwise.

### GetCredentialOk

`func (o *EnvironmentVariableOfAsset) GetCredentialOk() (*EnvironmentVariableCredential, bool)`

GetCredentialOk returns a tuple with the Credential field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredential

`func (o *EnvironmentVariableOfAsset) SetCredential(v EnvironmentVariableCredential)`

SetCredential sets Credential field to given value.

### HasCredential

`func (o *EnvironmentVariableOfAsset) HasCredential() bool`

HasCredential returns a boolean if a field has been set.

### SetCredentialNil

`func (o *EnvironmentVariableOfAsset) SetCredentialNil(b bool)`

 SetCredentialNil sets the value for Credential to be an explicit nil

### UnsetCredential
`func (o *EnvironmentVariableOfAsset) UnsetCredential()`

UnsetCredential ensures that no value is present for Credential, not even an explicit nil
### GetConfigMap

`func (o *EnvironmentVariableOfAsset) GetConfigMap() EnvironmentVariableConfigMap`

GetConfigMap returns the ConfigMap field if non-nil, zero value otherwise.

### GetConfigMapOk

`func (o *EnvironmentVariableOfAsset) GetConfigMapOk() (*EnvironmentVariableConfigMap, bool)`

GetConfigMapOk returns a tuple with the ConfigMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMap

`func (o *EnvironmentVariableOfAsset) SetConfigMap(v EnvironmentVariableConfigMap)`

SetConfigMap sets ConfigMap field to given value.

### HasConfigMap

`func (o *EnvironmentVariableOfAsset) HasConfigMap() bool`

HasConfigMap returns a boolean if a field has been set.

### SetConfigMapNil

`func (o *EnvironmentVariableOfAsset) SetConfigMapNil(b bool)`

 SetConfigMapNil sets the value for ConfigMap to be an explicit nil

### UnsetConfigMap
`func (o *EnvironmentVariableOfAsset) UnsetConfigMap()`

UnsetConfigMap ensures that no value is present for ConfigMap, not even an explicit nil
### GetPodFieldRef

`func (o *EnvironmentVariableOfAsset) GetPodFieldRef() EnvironmentVariablePodFieldReference`

GetPodFieldRef returns the PodFieldRef field if non-nil, zero value otherwise.

### GetPodFieldRefOk

`func (o *EnvironmentVariableOfAsset) GetPodFieldRefOk() (*EnvironmentVariablePodFieldReference, bool)`

GetPodFieldRefOk returns a tuple with the PodFieldRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodFieldRef

`func (o *EnvironmentVariableOfAsset) SetPodFieldRef(v EnvironmentVariablePodFieldReference)`

SetPodFieldRef sets PodFieldRef field to given value.

### HasPodFieldRef

`func (o *EnvironmentVariableOfAsset) HasPodFieldRef() bool`

HasPodFieldRef returns a boolean if a field has been set.

### SetPodFieldRefNil

`func (o *EnvironmentVariableOfAsset) SetPodFieldRefNil(b bool)`

 SetPodFieldRefNil sets the value for PodFieldRef to be an explicit nil

### UnsetPodFieldRef
`func (o *EnvironmentVariableOfAsset) UnsetPodFieldRef()`

UnsetPodFieldRef ensures that no value is present for PodFieldRef, not even an explicit nil
### GetExclude

`func (o *EnvironmentVariableOfAsset) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *EnvironmentVariableOfAsset) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *EnvironmentVariableOfAsset) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *EnvironmentVariableOfAsset) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *EnvironmentVariableOfAsset) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *EnvironmentVariableOfAsset) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil
### GetDescription

`func (o *EnvironmentVariableOfAsset) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *EnvironmentVariableOfAsset) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *EnvironmentVariableOfAsset) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *EnvironmentVariableOfAsset) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *EnvironmentVariableOfAsset) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *EnvironmentVariableOfAsset) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


