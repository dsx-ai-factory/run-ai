# AssignedResourcesV1ResponseMemory

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | **NullableFloat32** | The amount of resources guaranteed to be allocated in case the cluster has those resources. | 
**MaxAllowed** | **NullableFloat32** | Maximum amount of resources that can be allocated. If equal to deserved, no over-quota will be allowed. Use \&quot;-1\&quot; for unlimited over quota. | 
**OverQuotaWeight** | **NullableFloat32** | The priority for over quota resources. | 

## Methods

### NewAssignedResourcesV1ResponseMemory

`func NewAssignedResourcesV1ResponseMemory(deserved NullableFloat32, maxAllowed NullableFloat32, overQuotaWeight NullableFloat32, ) *AssignedResourcesV1ResponseMemory`

NewAssignedResourcesV1ResponseMemory instantiates a new AssignedResourcesV1ResponseMemory object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssignedResourcesV1ResponseMemoryWithDefaults

`func NewAssignedResourcesV1ResponseMemoryWithDefaults() *AssignedResourcesV1ResponseMemory`

NewAssignedResourcesV1ResponseMemoryWithDefaults instantiates a new AssignedResourcesV1ResponseMemory object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *AssignedResourcesV1ResponseMemory) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *AssignedResourcesV1ResponseMemory) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *AssignedResourcesV1ResponseMemory) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.


### SetDeservedNil

`func (o *AssignedResourcesV1ResponseMemory) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *AssignedResourcesV1ResponseMemory) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetMaxAllowed

`func (o *AssignedResourcesV1ResponseMemory) GetMaxAllowed() float32`

GetMaxAllowed returns the MaxAllowed field if non-nil, zero value otherwise.

### GetMaxAllowedOk

`func (o *AssignedResourcesV1ResponseMemory) GetMaxAllowedOk() (*float32, bool)`

GetMaxAllowedOk returns a tuple with the MaxAllowed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowed

`func (o *AssignedResourcesV1ResponseMemory) SetMaxAllowed(v float32)`

SetMaxAllowed sets MaxAllowed field to given value.


### SetMaxAllowedNil

`func (o *AssignedResourcesV1ResponseMemory) SetMaxAllowedNil(b bool)`

 SetMaxAllowedNil sets the value for MaxAllowed to be an explicit nil

### UnsetMaxAllowed
`func (o *AssignedResourcesV1ResponseMemory) UnsetMaxAllowed()`

UnsetMaxAllowed ensures that no value is present for MaxAllowed, not even an explicit nil
### GetOverQuotaWeight

`func (o *AssignedResourcesV1ResponseMemory) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *AssignedResourcesV1ResponseMemory) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *AssignedResourcesV1ResponseMemory) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.


### SetOverQuotaWeightNil

`func (o *AssignedResourcesV1ResponseMemory) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *AssignedResourcesV1ResponseMemory) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


