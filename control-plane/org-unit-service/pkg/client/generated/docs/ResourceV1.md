# ResourceV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **NullableFloat32** | The amount of resources guaranteed to be allocated in case the cluster has those resources. | [optional] 
**MaxAllowed** | Pointer to **NullableFloat32** | Maximum amount of resources that can be allocated. If equal to deserved, no over-quota will be allowed. Use \&quot;-1\&quot; for unlimited over quota. | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat32** | The priority for over quota resources. | [optional] 

## Methods

### NewResourceV1

`func NewResourceV1() *ResourceV1`

NewResourceV1 instantiates a new ResourceV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceV1WithDefaults

`func NewResourceV1WithDefaults() *ResourceV1`

NewResourceV1WithDefaults instantiates a new ResourceV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *ResourceV1) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *ResourceV1) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *ResourceV1) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *ResourceV1) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### SetDeservedNil

`func (o *ResourceV1) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *ResourceV1) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetMaxAllowed

`func (o *ResourceV1) GetMaxAllowed() float32`

GetMaxAllowed returns the MaxAllowed field if non-nil, zero value otherwise.

### GetMaxAllowedOk

`func (o *ResourceV1) GetMaxAllowedOk() (*float32, bool)`

GetMaxAllowedOk returns a tuple with the MaxAllowed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowed

`func (o *ResourceV1) SetMaxAllowed(v float32)`

SetMaxAllowed sets MaxAllowed field to given value.

### HasMaxAllowed

`func (o *ResourceV1) HasMaxAllowed() bool`

HasMaxAllowed returns a boolean if a field has been set.

### SetMaxAllowedNil

`func (o *ResourceV1) SetMaxAllowedNil(b bool)`

 SetMaxAllowedNil sets the value for MaxAllowed to be an explicit nil

### UnsetMaxAllowed
`func (o *ResourceV1) UnsetMaxAllowed()`

UnsetMaxAllowed ensures that no value is present for MaxAllowed, not even an explicit nil
### GetOverQuotaWeight

`func (o *ResourceV1) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *ResourceV1) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *ResourceV1) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *ResourceV1) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *ResourceV1) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *ResourceV1) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


