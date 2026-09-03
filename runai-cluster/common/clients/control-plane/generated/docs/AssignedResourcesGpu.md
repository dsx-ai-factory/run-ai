# AssignedResourcesGpu

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **float32** | The amount of resources guaranteed to be allocated in case the cluster has those resources. | [optional] 
**MaxAllowed** | Pointer to **float32** | Maximum amount of resources that can be allocated. If equal to deserved, no over-quota will be allowed. Use \&quot;-1\&quot; for unlimited over quota. | [optional] 
**OverQuotaWeight** | Pointer to **float32** | The priority for over quota resources. | [optional] 

## Methods

### NewAssignedResourcesGpu

`func NewAssignedResourcesGpu() *AssignedResourcesGpu`

NewAssignedResourcesGpu instantiates a new AssignedResourcesGpu object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssignedResourcesGpuWithDefaults

`func NewAssignedResourcesGpuWithDefaults() *AssignedResourcesGpu`

NewAssignedResourcesGpuWithDefaults instantiates a new AssignedResourcesGpu object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *AssignedResourcesGpu) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *AssignedResourcesGpu) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *AssignedResourcesGpu) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *AssignedResourcesGpu) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### GetMaxAllowed

`func (o *AssignedResourcesGpu) GetMaxAllowed() float32`

GetMaxAllowed returns the MaxAllowed field if non-nil, zero value otherwise.

### GetMaxAllowedOk

`func (o *AssignedResourcesGpu) GetMaxAllowedOk() (*float32, bool)`

GetMaxAllowedOk returns a tuple with the MaxAllowed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowed

`func (o *AssignedResourcesGpu) SetMaxAllowed(v float32)`

SetMaxAllowed sets MaxAllowed field to given value.

### HasMaxAllowed

`func (o *AssignedResourcesGpu) HasMaxAllowed() bool`

HasMaxAllowed returns a boolean if a field has been set.

### GetOverQuotaWeight

`func (o *AssignedResourcesGpu) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *AssignedResourcesGpu) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *AssignedResourcesGpu) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *AssignedResourcesGpu) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


