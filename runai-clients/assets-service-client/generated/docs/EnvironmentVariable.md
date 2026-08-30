# EnvironmentVariable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the environment variable. (mandatory) | [optional] 
**Value** | Pointer to **NullableString** | The value of the environment variable. (mutually exclusive with secret, userCredential, configMap and podFieldRef) | [optional] 
**Secret** | Pointer to [**NullableEnvironmentVariableSecret**](EnvironmentVariableSecret.md) |  | [optional] 
**ConfigMap** | Pointer to [**NullableEnvironmentVariableConfigMap**](EnvironmentVariableConfigMap.md) |  | [optional] 
**PodFieldRef** | Pointer to [**NullableEnvironmentVariablePodFieldReference**](EnvironmentVariablePodFieldReference.md) |  | [optional] 
**UserCredential** | Pointer to [**NullableEnvironmentVariableUserCredential**](EnvironmentVariableUserCredential.md) |  | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the environment variable is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] 
**Description** | Pointer to **NullableString** | Description of the environment variable. | [optional] 

## Methods

### NewEnvironmentVariable

`func NewEnvironmentVariable() *EnvironmentVariable`

NewEnvironmentVariable instantiates a new EnvironmentVariable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentVariableWithDefaults

`func NewEnvironmentVariableWithDefaults() *EnvironmentVariable`

NewEnvironmentVariableWithDefaults instantiates a new EnvironmentVariable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *EnvironmentVariable) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EnvironmentVariable) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EnvironmentVariable) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *EnvironmentVariable) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *EnvironmentVariable) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *EnvironmentVariable) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetValue

`func (o *EnvironmentVariable) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *EnvironmentVariable) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *EnvironmentVariable) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *EnvironmentVariable) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *EnvironmentVariable) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *EnvironmentVariable) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetSecret

`func (o *EnvironmentVariable) GetSecret() EnvironmentVariableSecret`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *EnvironmentVariable) GetSecretOk() (*EnvironmentVariableSecret, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *EnvironmentVariable) SetSecret(v EnvironmentVariableSecret)`

SetSecret sets Secret field to given value.

### HasSecret

`func (o *EnvironmentVariable) HasSecret() bool`

HasSecret returns a boolean if a field has been set.

### SetSecretNil

`func (o *EnvironmentVariable) SetSecretNil(b bool)`

 SetSecretNil sets the value for Secret to be an explicit nil

### UnsetSecret
`func (o *EnvironmentVariable) UnsetSecret()`

UnsetSecret ensures that no value is present for Secret, not even an explicit nil
### GetConfigMap

`func (o *EnvironmentVariable) GetConfigMap() EnvironmentVariableConfigMap`

GetConfigMap returns the ConfigMap field if non-nil, zero value otherwise.

### GetConfigMapOk

`func (o *EnvironmentVariable) GetConfigMapOk() (*EnvironmentVariableConfigMap, bool)`

GetConfigMapOk returns a tuple with the ConfigMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMap

`func (o *EnvironmentVariable) SetConfigMap(v EnvironmentVariableConfigMap)`

SetConfigMap sets ConfigMap field to given value.

### HasConfigMap

`func (o *EnvironmentVariable) HasConfigMap() bool`

HasConfigMap returns a boolean if a field has been set.

### SetConfigMapNil

`func (o *EnvironmentVariable) SetConfigMapNil(b bool)`

 SetConfigMapNil sets the value for ConfigMap to be an explicit nil

### UnsetConfigMap
`func (o *EnvironmentVariable) UnsetConfigMap()`

UnsetConfigMap ensures that no value is present for ConfigMap, not even an explicit nil
### GetPodFieldRef

`func (o *EnvironmentVariable) GetPodFieldRef() EnvironmentVariablePodFieldReference`

GetPodFieldRef returns the PodFieldRef field if non-nil, zero value otherwise.

### GetPodFieldRefOk

`func (o *EnvironmentVariable) GetPodFieldRefOk() (*EnvironmentVariablePodFieldReference, bool)`

GetPodFieldRefOk returns a tuple with the PodFieldRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodFieldRef

`func (o *EnvironmentVariable) SetPodFieldRef(v EnvironmentVariablePodFieldReference)`

SetPodFieldRef sets PodFieldRef field to given value.

### HasPodFieldRef

`func (o *EnvironmentVariable) HasPodFieldRef() bool`

HasPodFieldRef returns a boolean if a field has been set.

### SetPodFieldRefNil

`func (o *EnvironmentVariable) SetPodFieldRefNil(b bool)`

 SetPodFieldRefNil sets the value for PodFieldRef to be an explicit nil

### UnsetPodFieldRef
`func (o *EnvironmentVariable) UnsetPodFieldRef()`

UnsetPodFieldRef ensures that no value is present for PodFieldRef, not even an explicit nil
### GetUserCredential

`func (o *EnvironmentVariable) GetUserCredential() EnvironmentVariableUserCredential`

GetUserCredential returns the UserCredential field if non-nil, zero value otherwise.

### GetUserCredentialOk

`func (o *EnvironmentVariable) GetUserCredentialOk() (*EnvironmentVariableUserCredential, bool)`

GetUserCredentialOk returns a tuple with the UserCredential field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCredential

`func (o *EnvironmentVariable) SetUserCredential(v EnvironmentVariableUserCredential)`

SetUserCredential sets UserCredential field to given value.

### HasUserCredential

`func (o *EnvironmentVariable) HasUserCredential() bool`

HasUserCredential returns a boolean if a field has been set.

### SetUserCredentialNil

`func (o *EnvironmentVariable) SetUserCredentialNil(b bool)`

 SetUserCredentialNil sets the value for UserCredential to be an explicit nil

### UnsetUserCredential
`func (o *EnvironmentVariable) UnsetUserCredential()`

UnsetUserCredential ensures that no value is present for UserCredential, not even an explicit nil
### GetExclude

`func (o *EnvironmentVariable) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *EnvironmentVariable) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *EnvironmentVariable) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *EnvironmentVariable) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *EnvironmentVariable) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *EnvironmentVariable) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil
### GetDescription

`func (o *EnvironmentVariable) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *EnvironmentVariable) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *EnvironmentVariable) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *EnvironmentVariable) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *EnvironmentVariable) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *EnvironmentVariable) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


