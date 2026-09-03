# AssignedResourcesV1Cpu

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **NullableFloat32** | The amount of resources guaranteed to be allocated in case the cluster has those resources. | [optional] 
**MaxAllowed** | Pointer to **NullableFloat32** | Maximum amount of resources that can be allocated. If equal to deserved, no over-quota will be allowed. Use \&quot;-1\&quot; for unlimited over quota. | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat32** | The priority for over quota resources. | [optional] 

## Methods

### NewAssignedResourcesV1Cpu

`func NewAssignedResourcesV1Cpu() *AssignedResourcesV1Cpu`

NewAssignedResourcesV1Cpu instantiates a new AssignedResourcesV1Cpu object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssignedResourcesV1CpuWithDefaults

`func NewAssignedResourcesV1CpuWithDefaults() *AssignedResourcesV1Cpu`

NewAssignedResourcesV1CpuWithDefaults instantiates a new AssignedResourcesV1Cpu object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *AssignedResourcesV1Cpu) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *AssignedResourcesV1Cpu) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *AssignedResourcesV1Cpu) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *AssignedResourcesV1Cpu) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### SetDeservedNil

`func (o *AssignedResourcesV1Cpu) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *AssignedResourcesV1Cpu) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetMaxAllowed

`func (o *AssignedResourcesV1Cpu) GetMaxAllowed() float32`

GetMaxAllowed returns the MaxAllowed field if non-nil, zero value otherwise.

### GetMaxAllowedOk

`func (o *AssignedResourcesV1Cpu) GetMaxAllowedOk() (*float32, bool)`

GetMaxAllowedOk returns a tuple with the MaxAllowed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowed

`func (o *AssignedResourcesV1Cpu) SetMaxAllowed(v float32)`

SetMaxAllowed sets MaxAllowed field to given value.

### HasMaxAllowed

`func (o *AssignedResourcesV1Cpu) HasMaxAllowed() bool`

HasMaxAllowed returns a boolean if a field has been set.

### SetMaxAllowedNil

`func (o *AssignedResourcesV1Cpu) SetMaxAllowedNil(b bool)`

 SetMaxAllowedNil sets the value for MaxAllowed to be an explicit nil

### UnsetMaxAllowed
`func (o *AssignedResourcesV1Cpu) UnsetMaxAllowed()`

UnsetMaxAllowed ensures that no value is present for MaxAllowed, not even an explicit nil
### GetOverQuotaWeight

`func (o *AssignedResourcesV1Cpu) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *AssignedResourcesV1Cpu) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *AssignedResourcesV1Cpu) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *AssignedResourcesV1Cpu) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *AssignedResourcesV1Cpu) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *AssignedResourcesV1Cpu) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


