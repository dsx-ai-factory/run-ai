# NfssDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableNfsInstance**](NfsInstance.md) |  | [optional] 
**Instances** | Pointer to [**[]NfsInstance**](NfsInstance.md) | Set of nfs volumes to use in the workload. | [optional] 

## Methods

### NewNfssDefaults

`func NewNfssDefaults() *NfssDefaults`

NewNfssDefaults instantiates a new NfssDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNfssDefaultsWithDefaults

`func NewNfssDefaultsWithDefaults() *NfssDefaults`

NewNfssDefaultsWithDefaults instantiates a new NfssDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *NfssDefaults) GetAttributes() NfsInstance`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *NfssDefaults) GetAttributesOk() (*NfsInstance, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *NfssDefaults) SetAttributes(v NfsInstance)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *NfssDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *NfssDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *NfssDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *NfssDefaults) GetInstances() []NfsInstance`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *NfssDefaults) GetInstancesOk() (*[]NfsInstance, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *NfssDefaults) SetInstances(v []NfsInstance)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *NfssDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *NfssDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *NfssDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


