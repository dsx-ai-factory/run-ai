# UnifiedWorkload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**WorkloadMeta**](WorkloadMeta.md) |  | [optional] 
**Spec** | Pointer to [**NullableUnifiedSpec**](UnifiedSpec.md) |  | [optional] 
**PolicyInfo** | Pointer to [**NullableUnifiedPolicyInfo**](UnifiedPolicyInfo.md) |  | [optional] 
**ProjectInfo** | Pointer to [**NullableUnifiedProjectInfo**](UnifiedProjectInfo.md) |  | [optional] 

## Methods

### NewUnifiedWorkload

`func NewUnifiedWorkload() *UnifiedWorkload`

NewUnifiedWorkload instantiates a new UnifiedWorkload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUnifiedWorkloadWithDefaults

`func NewUnifiedWorkloadWithDefaults() *UnifiedWorkload`

NewUnifiedWorkloadWithDefaults instantiates a new UnifiedWorkload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *UnifiedWorkload) GetMeta() WorkloadMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *UnifiedWorkload) GetMetaOk() (*WorkloadMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *UnifiedWorkload) SetMeta(v WorkloadMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *UnifiedWorkload) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetSpec

`func (o *UnifiedWorkload) GetSpec() UnifiedSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *UnifiedWorkload) GetSpecOk() (*UnifiedSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *UnifiedWorkload) SetSpec(v UnifiedSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *UnifiedWorkload) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### SetSpecNil

`func (o *UnifiedWorkload) SetSpecNil(b bool)`

 SetSpecNil sets the value for Spec to be an explicit nil

### UnsetSpec
`func (o *UnifiedWorkload) UnsetSpec()`

UnsetSpec ensures that no value is present for Spec, not even an explicit nil
### GetPolicyInfo

`func (o *UnifiedWorkload) GetPolicyInfo() UnifiedPolicyInfo`

GetPolicyInfo returns the PolicyInfo field if non-nil, zero value otherwise.

### GetPolicyInfoOk

`func (o *UnifiedWorkload) GetPolicyInfoOk() (*UnifiedPolicyInfo, bool)`

GetPolicyInfoOk returns a tuple with the PolicyInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicyInfo

`func (o *UnifiedWorkload) SetPolicyInfo(v UnifiedPolicyInfo)`

SetPolicyInfo sets PolicyInfo field to given value.

### HasPolicyInfo

`func (o *UnifiedWorkload) HasPolicyInfo() bool`

HasPolicyInfo returns a boolean if a field has been set.

### SetPolicyInfoNil

`func (o *UnifiedWorkload) SetPolicyInfoNil(b bool)`

 SetPolicyInfoNil sets the value for PolicyInfo to be an explicit nil

### UnsetPolicyInfo
`func (o *UnifiedWorkload) UnsetPolicyInfo()`

UnsetPolicyInfo ensures that no value is present for PolicyInfo, not even an explicit nil
### GetProjectInfo

`func (o *UnifiedWorkload) GetProjectInfo() UnifiedProjectInfo`

GetProjectInfo returns the ProjectInfo field if non-nil, zero value otherwise.

### GetProjectInfoOk

`func (o *UnifiedWorkload) GetProjectInfoOk() (*UnifiedProjectInfo, bool)`

GetProjectInfoOk returns a tuple with the ProjectInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectInfo

`func (o *UnifiedWorkload) SetProjectInfo(v UnifiedProjectInfo)`

SetProjectInfo sets ProjectInfo field to given value.

### HasProjectInfo

`func (o *UnifiedWorkload) HasProjectInfo() bool`

HasProjectInfo returns a boolean if a field has been set.

### SetProjectInfoNil

`func (o *UnifiedWorkload) SetProjectInfoNil(b bool)`

 SetProjectInfoNil sets the value for ProjectInfo to be an explicit nil

### UnsetProjectInfo
`func (o *UnifiedWorkload) UnsetProjectInfo()`

UnsetProjectInfo ensures that no value is present for ProjectInfo, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


