# PolicyStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to [**PolicyPhase**](PolicyPhase.md) |  | [optional] 
**Issues** | Pointer to **[]string** | The list of configuration issues detected for the policy. | [optional] 

## Methods

### NewPolicyStatus

`func NewPolicyStatus() *PolicyStatus`

NewPolicyStatus instantiates a new PolicyStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyStatusWithDefaults

`func NewPolicyStatusWithDefaults() *PolicyStatus`

NewPolicyStatusWithDefaults instantiates a new PolicyStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *PolicyStatus) GetPhase() PolicyPhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PolicyStatus) GetPhaseOk() (*PolicyPhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PolicyStatus) SetPhase(v PolicyPhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PolicyStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetIssues

`func (o *PolicyStatus) GetIssues() []string`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *PolicyStatus) GetIssuesOk() (*[]string, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *PolicyStatus) SetIssues(v []string)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *PolicyStatus) HasIssues() bool`

HasIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


