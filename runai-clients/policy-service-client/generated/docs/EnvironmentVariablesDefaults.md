# EnvironmentVariablesDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Instances** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workload. | [optional] 

## Methods

### NewEnvironmentVariablesDefaults

`func NewEnvironmentVariablesDefaults() *EnvironmentVariablesDefaults`

NewEnvironmentVariablesDefaults instantiates a new EnvironmentVariablesDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentVariablesDefaultsWithDefaults

`func NewEnvironmentVariablesDefaultsWithDefaults() *EnvironmentVariablesDefaults`

NewEnvironmentVariablesDefaultsWithDefaults instantiates a new EnvironmentVariablesDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInstances

`func (o *EnvironmentVariablesDefaults) GetInstances() []EnvironmentVariable`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *EnvironmentVariablesDefaults) GetInstancesOk() (*[]EnvironmentVariable, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *EnvironmentVariablesDefaults) SetInstances(v []EnvironmentVariable)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *EnvironmentVariablesDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *EnvironmentVariablesDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *EnvironmentVariablesDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


