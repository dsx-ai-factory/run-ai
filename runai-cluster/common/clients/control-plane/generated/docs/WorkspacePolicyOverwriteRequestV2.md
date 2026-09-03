# WorkspacePolicyOverwriteRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableWorkspacePolicyDefaultsAndRulesV2**](WorkspacePolicyDefaultsAndRulesV2.md) |  | [optional] 

## Methods

### NewWorkspacePolicyOverwriteRequestV2

`func NewWorkspacePolicyOverwriteRequestV2() *WorkspacePolicyOverwriteRequestV2`

NewWorkspacePolicyOverwriteRequestV2 instantiates a new WorkspacePolicyOverwriteRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspacePolicyOverwriteRequestV2WithDefaults

`func NewWorkspacePolicyOverwriteRequestV2WithDefaults() *WorkspacePolicyOverwriteRequestV2`

NewWorkspacePolicyOverwriteRequestV2WithDefaults instantiates a new WorkspacePolicyOverwriteRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspacePolicyOverwriteRequestV2) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspacePolicyOverwriteRequestV2) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspacePolicyOverwriteRequestV2) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *WorkspacePolicyOverwriteRequestV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *WorkspacePolicyOverwriteRequestV2) GetPolicy() WorkspacePolicyDefaultsAndRulesV2`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *WorkspacePolicyOverwriteRequestV2) GetPolicyOk() (*WorkspacePolicyDefaultsAndRulesV2, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *WorkspacePolicyOverwriteRequestV2) SetPolicy(v WorkspacePolicyDefaultsAndRulesV2)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *WorkspacePolicyOverwriteRequestV2) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *WorkspacePolicyOverwriteRequestV2) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *WorkspacePolicyOverwriteRequestV2) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


