# DistributedInferenceStartupPolicyOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StartupPolicy** | Pointer to [**NullableDistributedInferenceStartupPolicy**](DistributedInferenceStartupPolicy.md) |  | [optional] [default to DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED]

## Methods

### NewDistributedInferenceStartupPolicyOptions

`func NewDistributedInferenceStartupPolicyOptions() *DistributedInferenceStartupPolicyOptions`

NewDistributedInferenceStartupPolicyOptions instantiates a new DistributedInferenceStartupPolicyOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceStartupPolicyOptionsWithDefaults

`func NewDistributedInferenceStartupPolicyOptionsWithDefaults() *DistributedInferenceStartupPolicyOptions`

NewDistributedInferenceStartupPolicyOptionsWithDefaults instantiates a new DistributedInferenceStartupPolicyOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStartupPolicy

`func (o *DistributedInferenceStartupPolicyOptions) GetStartupPolicy() DistributedInferenceStartupPolicy`

GetStartupPolicy returns the StartupPolicy field if non-nil, zero value otherwise.

### GetStartupPolicyOk

`func (o *DistributedInferenceStartupPolicyOptions) GetStartupPolicyOk() (*DistributedInferenceStartupPolicy, bool)`

GetStartupPolicyOk returns a tuple with the StartupPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartupPolicy

`func (o *DistributedInferenceStartupPolicyOptions) SetStartupPolicy(v DistributedInferenceStartupPolicy)`

SetStartupPolicy sets StartupPolicy field to given value.

### HasStartupPolicy

`func (o *DistributedInferenceStartupPolicyOptions) HasStartupPolicy() bool`

HasStartupPolicy returns a boolean if a field has been set.

### SetStartupPolicyNil

`func (o *DistributedInferenceStartupPolicyOptions) SetStartupPolicyNil(b bool)`

 SetStartupPolicyNil sets the value for StartupPolicy to be an explicit nil

### UnsetStartupPolicy
`func (o *DistributedInferenceStartupPolicyOptions) UnsetStartupPolicy()`

UnsetStartupPolicy ensures that no value is present for StartupPolicy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


