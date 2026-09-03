# TrainingPolicyChangeRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableTrainingPolicyDefaultsAndRulesV2**](TrainingPolicyDefaultsAndRulesV2.md) |  | [optional] 
**Reset** | Pointer to **[]string** | set of fields in jsonpath format that is requested to clear their policy (default and rules) | [optional] 

## Methods

### NewTrainingPolicyChangeRequestV2

`func NewTrainingPolicyChangeRequestV2() *TrainingPolicyChangeRequestV2`

NewTrainingPolicyChangeRequestV2 instantiates a new TrainingPolicyChangeRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingPolicyChangeRequestV2WithDefaults

`func NewTrainingPolicyChangeRequestV2WithDefaults() *TrainingPolicyChangeRequestV2`

NewTrainingPolicyChangeRequestV2WithDefaults instantiates a new TrainingPolicyChangeRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingPolicyChangeRequestV2) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingPolicyChangeRequestV2) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingPolicyChangeRequestV2) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *TrainingPolicyChangeRequestV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *TrainingPolicyChangeRequestV2) GetPolicy() TrainingPolicyDefaultsAndRulesV2`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *TrainingPolicyChangeRequestV2) GetPolicyOk() (*TrainingPolicyDefaultsAndRulesV2, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *TrainingPolicyChangeRequestV2) SetPolicy(v TrainingPolicyDefaultsAndRulesV2)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *TrainingPolicyChangeRequestV2) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *TrainingPolicyChangeRequestV2) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *TrainingPolicyChangeRequestV2) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetReset

`func (o *TrainingPolicyChangeRequestV2) GetReset() []string`

GetReset returns the Reset field if non-nil, zero value otherwise.

### GetResetOk

`func (o *TrainingPolicyChangeRequestV2) GetResetOk() (*[]string, bool)`

GetResetOk returns a tuple with the Reset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReset

`func (o *TrainingPolicyChangeRequestV2) SetReset(v []string)`

SetReset sets Reset field to given value.

### HasReset

`func (o *TrainingPolicyChangeRequestV2) HasReset() bool`

HasReset returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


