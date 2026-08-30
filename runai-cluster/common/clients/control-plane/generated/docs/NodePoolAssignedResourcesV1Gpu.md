# NodePoolAssignedResourcesV1Gpu

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | **float32** | The amount of resources guaranteed to be allocated in case the cluster has those resources. | 
**MaxAllowed** | Pointer to **NullableFloat32** | Maximum amount of resources that can be allocated. If equal to deserved, no over-quota will be allowed. Use \&quot;-1\&quot; for unlimited over quota. | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat32** | The priority for over quota resources. | [optional] 

## Methods

### NewNodePoolAssignedResourcesV1Gpu

`func NewNodePoolAssignedResourcesV1Gpu(deserved float32, ) *NodePoolAssignedResourcesV1Gpu`

NewNodePoolAssignedResourcesV1Gpu instantiates a new NodePoolAssignedResourcesV1Gpu object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolAssignedResourcesV1GpuWithDefaults

`func NewNodePoolAssignedResourcesV1GpuWithDefaults() *NodePoolAssignedResourcesV1Gpu`

NewNodePoolAssignedResourcesV1GpuWithDefaults instantiates a new NodePoolAssignedResourcesV1Gpu object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *NodePoolAssignedResourcesV1Gpu) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *NodePoolAssignedResourcesV1Gpu) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *NodePoolAssignedResourcesV1Gpu) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.


### GetMaxAllowed

`func (o *NodePoolAssignedResourcesV1Gpu) GetMaxAllowed() float32`

GetMaxAllowed returns the MaxAllowed field if non-nil, zero value otherwise.

### GetMaxAllowedOk

`func (o *NodePoolAssignedResourcesV1Gpu) GetMaxAllowedOk() (*float32, bool)`

GetMaxAllowedOk returns a tuple with the MaxAllowed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowed

`func (o *NodePoolAssignedResourcesV1Gpu) SetMaxAllowed(v float32)`

SetMaxAllowed sets MaxAllowed field to given value.

### HasMaxAllowed

`func (o *NodePoolAssignedResourcesV1Gpu) HasMaxAllowed() bool`

HasMaxAllowed returns a boolean if a field has been set.

### SetMaxAllowedNil

`func (o *NodePoolAssignedResourcesV1Gpu) SetMaxAllowedNil(b bool)`

 SetMaxAllowedNil sets the value for MaxAllowed to be an explicit nil

### UnsetMaxAllowed
`func (o *NodePoolAssignedResourcesV1Gpu) UnsetMaxAllowed()`

UnsetMaxAllowed ensures that no value is present for MaxAllowed, not even an explicit nil
### GetOverQuotaWeight

`func (o *NodePoolAssignedResourcesV1Gpu) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *NodePoolAssignedResourcesV1Gpu) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *NodePoolAssignedResourcesV1Gpu) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *NodePoolAssignedResourcesV1Gpu) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *NodePoolAssignedResourcesV1Gpu) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *NodePoolAssignedResourcesV1Gpu) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


