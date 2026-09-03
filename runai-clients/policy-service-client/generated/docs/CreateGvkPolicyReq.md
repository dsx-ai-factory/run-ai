# CreateGvkPolicyReq

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **NullableString** | A description of the policy. | [optional] 
**WorkloadType** | [**GVK**](GVK.md) |  | 
**Scope** | [**ScopeType**](ScopeType.md) |  | 
**ClusterId** | Pointer to **NullableString** | The id of the cluster the policy applies to. Required when the scope is cluster. | [optional] 
**DepartmentId** | Pointer to **NullableString** | The id of the department the policy applies to. Required when the scope is department. | [optional] 
**ProjectId** | Pointer to **NullableString** | The id of the project the policy applies to. Required when the scope is project. | [optional] 
**Rules** | Pointer to [**[]PolicyRule**](PolicyRule.md) | The rules declared on this policy. Each rule targets a field in the workload spec and constrains its allowed values through its enforcements. | [optional] 

## Methods

### NewCreateGvkPolicyReq

`func NewCreateGvkPolicyReq(workloadType GVK, scope ScopeType, ) *CreateGvkPolicyReq`

NewCreateGvkPolicyReq instantiates a new CreateGvkPolicyReq object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateGvkPolicyReqWithDefaults

`func NewCreateGvkPolicyReqWithDefaults() *CreateGvkPolicyReq`

NewCreateGvkPolicyReqWithDefaults instantiates a new CreateGvkPolicyReq object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *CreateGvkPolicyReq) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *CreateGvkPolicyReq) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *CreateGvkPolicyReq) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *CreateGvkPolicyReq) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *CreateGvkPolicyReq) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *CreateGvkPolicyReq) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetWorkloadType

`func (o *CreateGvkPolicyReq) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *CreateGvkPolicyReq) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *CreateGvkPolicyReq) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.


### GetScope

`func (o *CreateGvkPolicyReq) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *CreateGvkPolicyReq) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *CreateGvkPolicyReq) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *CreateGvkPolicyReq) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *CreateGvkPolicyReq) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *CreateGvkPolicyReq) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *CreateGvkPolicyReq) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *CreateGvkPolicyReq) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *CreateGvkPolicyReq) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *CreateGvkPolicyReq) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *CreateGvkPolicyReq) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *CreateGvkPolicyReq) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *CreateGvkPolicyReq) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *CreateGvkPolicyReq) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *CreateGvkPolicyReq) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *CreateGvkPolicyReq) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *CreateGvkPolicyReq) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *CreateGvkPolicyReq) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *CreateGvkPolicyReq) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *CreateGvkPolicyReq) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *CreateGvkPolicyReq) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetRules

`func (o *CreateGvkPolicyReq) GetRules() []PolicyRule`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *CreateGvkPolicyReq) GetRulesOk() (*[]PolicyRule, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *CreateGvkPolicyReq) SetRules(v []PolicyRule)`

SetRules sets Rules field to given value.

### HasRules

`func (o *CreateGvkPolicyReq) HasRules() bool`

HasRules returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


