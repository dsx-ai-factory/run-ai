# GetGvkPolicyResp

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
**Breakdown** | Pointer to [**NullableBreakdownBreakdown**](BreakdownBreakdown.md) |  | [optional] 

## Methods

### NewGetGvkPolicyResp

`func NewGetGvkPolicyResp(workloadType GVK, scope ScopeType, id string, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time, ) *GetGvkPolicyResp`

NewGetGvkPolicyResp instantiates a new GetGvkPolicyResp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetGvkPolicyRespWithDefaults

`func NewGetGvkPolicyRespWithDefaults() *GetGvkPolicyResp`

NewGetGvkPolicyRespWithDefaults instantiates a new GetGvkPolicyResp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *GetGvkPolicyResp) GetStatus() PolicyStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *GetGvkPolicyResp) GetStatusOk() (*PolicyStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *GetGvkPolicyResp) SetStatus(v PolicyStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *GetGvkPolicyResp) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *GetGvkPolicyResp) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *GetGvkPolicyResp) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetDescription

`func (o *GetGvkPolicyResp) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GetGvkPolicyResp) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GetGvkPolicyResp) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GetGvkPolicyResp) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *GetGvkPolicyResp) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *GetGvkPolicyResp) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetWorkloadType

`func (o *GetGvkPolicyResp) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *GetGvkPolicyResp) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *GetGvkPolicyResp) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.


### GetScope

`func (o *GetGvkPolicyResp) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *GetGvkPolicyResp) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *GetGvkPolicyResp) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *GetGvkPolicyResp) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *GetGvkPolicyResp) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *GetGvkPolicyResp) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *GetGvkPolicyResp) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *GetGvkPolicyResp) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *GetGvkPolicyResp) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *GetGvkPolicyResp) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *GetGvkPolicyResp) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *GetGvkPolicyResp) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *GetGvkPolicyResp) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *GetGvkPolicyResp) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *GetGvkPolicyResp) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *GetGvkPolicyResp) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *GetGvkPolicyResp) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *GetGvkPolicyResp) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *GetGvkPolicyResp) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *GetGvkPolicyResp) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *GetGvkPolicyResp) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetRules

`func (o *GetGvkPolicyResp) GetRules() []PolicyRule`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *GetGvkPolicyResp) GetRulesOk() (*[]PolicyRule, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *GetGvkPolicyResp) SetRules(v []PolicyRule)`

SetRules sets Rules field to given value.

### HasRules

`func (o *GetGvkPolicyResp) HasRules() bool`

HasRules returns a boolean if a field has been set.

### GetEffectiveRules

`func (o *GetGvkPolicyResp) GetEffectiveRules() []EffectiveRulesEffectiveRulesInner`

GetEffectiveRules returns the EffectiveRules field if non-nil, zero value otherwise.

### GetEffectiveRulesOk

`func (o *GetGvkPolicyResp) GetEffectiveRulesOk() (*[]EffectiveRulesEffectiveRulesInner, bool)`

GetEffectiveRulesOk returns a tuple with the EffectiveRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveRules

`func (o *GetGvkPolicyResp) SetEffectiveRules(v []EffectiveRulesEffectiveRulesInner)`

SetEffectiveRules sets EffectiveRules field to given value.

### HasEffectiveRules

`func (o *GetGvkPolicyResp) HasEffectiveRules() bool`

HasEffectiveRules returns a boolean if a field has been set.

### GetId

`func (o *GetGvkPolicyResp) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetGvkPolicyResp) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetGvkPolicyResp) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedBy

`func (o *GetGvkPolicyResp) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *GetGvkPolicyResp) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *GetGvkPolicyResp) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *GetGvkPolicyResp) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetGvkPolicyResp) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetGvkPolicyResp) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedBy

`func (o *GetGvkPolicyResp) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *GetGvkPolicyResp) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *GetGvkPolicyResp) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetUpdatedAt

`func (o *GetGvkPolicyResp) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GetGvkPolicyResp) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GetGvkPolicyResp) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetBreakdown

`func (o *GetGvkPolicyResp) GetBreakdown() BreakdownBreakdown`

GetBreakdown returns the Breakdown field if non-nil, zero value otherwise.

### GetBreakdownOk

`func (o *GetGvkPolicyResp) GetBreakdownOk() (*BreakdownBreakdown, bool)`

GetBreakdownOk returns a tuple with the Breakdown field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBreakdown

`func (o *GetGvkPolicyResp) SetBreakdown(v BreakdownBreakdown)`

SetBreakdown sets Breakdown field to given value.

### HasBreakdown

`func (o *GetGvkPolicyResp) HasBreakdown() bool`

HasBreakdown returns a boolean if a field has been set.

### SetBreakdownNil

`func (o *GetGvkPolicyResp) SetBreakdownNil(b bool)`

 SetBreakdownNil sets the value for Breakdown to be an explicit nil

### UnsetBreakdown
`func (o *GetGvkPolicyResp) UnsetBreakdown()`

UnsetBreakdown ensures that no value is present for Breakdown, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


