# Policy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to [**NullablePolicyStatus**](PolicyStatus.md) | Configuration issues detected by the server for this policy. Returned in responses only; not accepted in requests. | [optional] 
**Description** | Pointer to **NullableString** | A description of the policy. | [optional] 
**WorkloadType** | [**GVK**](GVK.md) |  | 
**Scope** | [**ScopeType**](ScopeType.md) |  | 
**ClusterId** | Pointer to **NullableString** | The id of the cluster the policy applies to. Required when the scope is cluster. | [optional] 
**DepartmentId** | Pointer to **NullableString** | The id of the department the policy applies to. Required when the scope is department. | [optional] 
**ProjectId** | Pointer to **NullableString** | The id of the project the policy applies to. Required when the scope is project. | [optional] 
**Rules** | Pointer to [**[]PolicyRule**](PolicyRule.md) | The rules declared on this policy. Each rule targets a field in the workload spec and constrains its allowed values through its enforcements. | [optional] 
**EffectiveRules** | Pointer to [**[]EffectiveRulesEffectiveRulesInner**](EffectiveRulesEffectiveRulesInner.md) | The merged rule set in effect for this policy, including rules inherited from broader scopes (cluster, then department, then project). Each entry adds the sourceOfRule indicating where the rule was defined. Returned in responses only; not accepted in requests. | [optional] 
**Id** | **string** | The unique identifier for the policy. | 
**CreatedBy** | **string** | Identifier of the user who created the policy. | 
**CreatedAt** | **time.Time** | The timestamp for when the policy was created. | 
**UpdatedBy** | **string** | Identifier of the user who last updated the policy. | 
**UpdatedAt** | **time.Time** | The timestamp for when the policy was last updated. | 

## Methods

### NewPolicy

`func NewPolicy(workloadType GVK, scope ScopeType, id string, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time, ) *Policy`

NewPolicy instantiates a new Policy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyWithDefaults

`func NewPolicyWithDefaults() *Policy`

NewPolicyWithDefaults instantiates a new Policy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *Policy) GetStatus() PolicyStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Policy) GetStatusOk() (*PolicyStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Policy) SetStatus(v PolicyStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Policy) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *Policy) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *Policy) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetDescription

`func (o *Policy) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Policy) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Policy) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Policy) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *Policy) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *Policy) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetWorkloadType

`func (o *Policy) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *Policy) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *Policy) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.


### GetScope

`func (o *Policy) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *Policy) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *Policy) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *Policy) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Policy) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Policy) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *Policy) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *Policy) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *Policy) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *Policy) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Policy) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Policy) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Policy) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *Policy) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *Policy) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *Policy) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Policy) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Policy) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *Policy) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *Policy) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *Policy) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetRules

`func (o *Policy) GetRules() []PolicyRule`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *Policy) GetRulesOk() (*[]PolicyRule, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *Policy) SetRules(v []PolicyRule)`

SetRules sets Rules field to given value.

### HasRules

`func (o *Policy) HasRules() bool`

HasRules returns a boolean if a field has been set.

### GetEffectiveRules

`func (o *Policy) GetEffectiveRules() []EffectiveRulesEffectiveRulesInner`

GetEffectiveRules returns the EffectiveRules field if non-nil, zero value otherwise.

### GetEffectiveRulesOk

`func (o *Policy) GetEffectiveRulesOk() (*[]EffectiveRulesEffectiveRulesInner, bool)`

GetEffectiveRulesOk returns a tuple with the EffectiveRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveRules

`func (o *Policy) SetEffectiveRules(v []EffectiveRulesEffectiveRulesInner)`

SetEffectiveRules sets EffectiveRules field to given value.

### HasEffectiveRules

`func (o *Policy) HasEffectiveRules() bool`

HasEffectiveRules returns a boolean if a field has been set.

### GetId

`func (o *Policy) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Policy) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Policy) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedBy

`func (o *Policy) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Policy) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Policy) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *Policy) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Policy) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Policy) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedBy

`func (o *Policy) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *Policy) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *Policy) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetUpdatedAt

`func (o *Policy) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Policy) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Policy) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


