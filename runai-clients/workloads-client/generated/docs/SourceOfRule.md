# SourceOfRule

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Scope** | [**ScopeType**](ScopeType.md) |  | 
**ProjectId** | Pointer to **NullableInt32** | The id of the project. | [optional] 
**DepartmentId** | Pointer to **NullableString** | The id of the department. | [optional] 
**ClusterId** | Pointer to **NullableString** | The id of the cluster. | [optional] 
**IsFallback** | Pointer to **bool** | source of this rule is the fallback policy | [optional] 

## Methods

### NewSourceOfRule

`func NewSourceOfRule(scope ScopeType, ) *SourceOfRule`

NewSourceOfRule instantiates a new SourceOfRule object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSourceOfRuleWithDefaults

`func NewSourceOfRuleWithDefaults() *SourceOfRule`

NewSourceOfRuleWithDefaults instantiates a new SourceOfRule object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScope

`func (o *SourceOfRule) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *SourceOfRule) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *SourceOfRule) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetProjectId

`func (o *SourceOfRule) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *SourceOfRule) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *SourceOfRule) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *SourceOfRule) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *SourceOfRule) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *SourceOfRule) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetDepartmentId

`func (o *SourceOfRule) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *SourceOfRule) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *SourceOfRule) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *SourceOfRule) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *SourceOfRule) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *SourceOfRule) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetClusterId

`func (o *SourceOfRule) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *SourceOfRule) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *SourceOfRule) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *SourceOfRule) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *SourceOfRule) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *SourceOfRule) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetIsFallback

`func (o *SourceOfRule) GetIsFallback() bool`

GetIsFallback returns the IsFallback field if non-nil, zero value otherwise.

### GetIsFallbackOk

`func (o *SourceOfRule) GetIsFallbackOk() (*bool, bool)`

GetIsFallbackOk returns a tuple with the IsFallback field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsFallback

`func (o *SourceOfRule) SetIsFallback(v bool)`

SetIsFallback sets IsFallback field to given value.

### HasIsFallback

`func (o *SourceOfRule) HasIsFallback() bool`

HasIsFallback returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


