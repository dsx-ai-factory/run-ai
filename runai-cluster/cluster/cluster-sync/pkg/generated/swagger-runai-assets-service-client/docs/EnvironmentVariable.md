# EnvironmentVariable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the environment variable. | 
**Value** | **string** | The value to set the environment variable to. | 
**Deleted** | Pointer to **bool** | Whether to exclude this environment variable from the workload. This is necessary in case the variable definition is inherited from a policy. | [optional] 

## Methods

### NewEnvironmentVariable

`func NewEnvironmentVariable(name string, value string, ) *EnvironmentVariable`

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


### GetDeleted

`func (o *EnvironmentVariable) GetDeleted() bool`

GetDeleted returns the Deleted field if non-nil, zero value otherwise.

### GetDeletedOk

`func (o *EnvironmentVariable) GetDeletedOk() (*bool, bool)`

GetDeletedOk returns a tuple with the Deleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleted

`func (o *EnvironmentVariable) SetDeleted(v bool)`

SetDeleted sets Deleted field to given value.

### HasDeleted

`func (o *EnvironmentVariable) HasDeleted() bool`

HasDeleted returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


