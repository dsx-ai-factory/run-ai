# GitsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableGitInstance**](GitInstance.md) |  | [optional] 
**Instances** | Pointer to [**[]GitInstance**](GitInstance.md) | Set of git repositories to use in the workload. | [optional] 

## Methods

### NewGitsDefaults

`func NewGitsDefaults() *GitsDefaults`

NewGitsDefaults instantiates a new GitsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitsDefaultsWithDefaults

`func NewGitsDefaultsWithDefaults() *GitsDefaults`

NewGitsDefaultsWithDefaults instantiates a new GitsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *GitsDefaults) GetAttributes() GitInstance`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *GitsDefaults) GetAttributesOk() (*GitInstance, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *GitsDefaults) SetAttributes(v GitInstance)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *GitsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *GitsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *GitsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *GitsDefaults) GetInstances() []GitInstance`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *GitsDefaults) GetInstancesOk() (*[]GitInstance, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *GitsDefaults) SetInstances(v []GitInstance)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *GitsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *GitsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *GitsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


