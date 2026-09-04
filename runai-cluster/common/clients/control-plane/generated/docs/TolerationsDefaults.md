# TolerationsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableToleration**](Toleration.md) |  | [optional] 
**Instances** | Pointer to [**[]Toleration**](Toleration.md) | Set of tolerations to apply to the workload. | [optional] 

## Methods

### NewTolerationsDefaults

`func NewTolerationsDefaults() *TolerationsDefaults`

NewTolerationsDefaults instantiates a new TolerationsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTolerationsDefaultsWithDefaults

`func NewTolerationsDefaultsWithDefaults() *TolerationsDefaults`

NewTolerationsDefaultsWithDefaults instantiates a new TolerationsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *TolerationsDefaults) GetAttributes() Toleration`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *TolerationsDefaults) GetAttributesOk() (*Toleration, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *TolerationsDefaults) SetAttributes(v Toleration)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *TolerationsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *TolerationsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *TolerationsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *TolerationsDefaults) GetInstances() []Toleration`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *TolerationsDefaults) GetInstancesOk() (*[]Toleration, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *TolerationsDefaults) SetInstances(v []Toleration)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *TolerationsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *TolerationsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *TolerationsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


