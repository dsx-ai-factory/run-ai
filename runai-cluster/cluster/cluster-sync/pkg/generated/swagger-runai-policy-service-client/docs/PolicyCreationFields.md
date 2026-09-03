# PolicyCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Scope** | [**ScopeType**](ScopeType.md) |  | 
**ProjectId** | Pointer to **NullableInt32** | The id of the project. | [optional] 
**DepartmentId** | Pointer to **NullableString** | The id of the department. | [optional] 
**ClusterId** | Pointer to **NullableString** | The id of the cluster. | [optional] 
**Name** | **string** | A chosen name for the policy. | 

## Methods

### NewPolicyCreationFields

`func NewPolicyCreationFields(scope ScopeType, name string, ) *PolicyCreationFields`

NewPolicyCreationFields instantiates a new PolicyCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyCreationFieldsWithDefaults

`func NewPolicyCreationFieldsWithDefaults() *PolicyCreationFields`

NewPolicyCreationFieldsWithDefaults instantiates a new PolicyCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScope

`func (o *PolicyCreationFields) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *PolicyCreationFields) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *PolicyCreationFields) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetProjectId

`func (o *PolicyCreationFields) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *PolicyCreationFields) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *PolicyCreationFields) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *PolicyCreationFields) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *PolicyCreationFields) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *PolicyCreationFields) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetDepartmentId

`func (o *PolicyCreationFields) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *PolicyCreationFields) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *PolicyCreationFields) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *PolicyCreationFields) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *PolicyCreationFields) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *PolicyCreationFields) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetClusterId

`func (o *PolicyCreationFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PolicyCreationFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PolicyCreationFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *PolicyCreationFields) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *PolicyCreationFields) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *PolicyCreationFields) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetName

`func (o *PolicyCreationFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PolicyCreationFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PolicyCreationFields) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


