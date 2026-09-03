# EnvironmentVariable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the environment variable. (mandatory) | [optional] 
**Value** | Pointer to **NullableString** | The value of the environment variable. (mutually exclusive with secret) | [optional] 
**Secret** | Pointer to [**NullableEnvironmentVariableSecret**](EnvironmentVariableSecret.md) |  | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the environment variable is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] 

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

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


