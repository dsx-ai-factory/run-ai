# InternalCreateAccessRuleRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SubjectId** | **string** |  | 
**SubjectType** | [**SubjectType**](SubjectType.md) |  | 
**Scope** | Pointer to [**NullableInternalCreateAccessRuleRequestScope**](InternalCreateAccessRuleRequestScope.md) |  | [optional] 

## Methods

### NewInternalCreateAccessRuleRequest

`func NewInternalCreateAccessRuleRequest(subjectId string, subjectType SubjectType, ) *InternalCreateAccessRuleRequest`

NewInternalCreateAccessRuleRequest instantiates a new InternalCreateAccessRuleRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalCreateAccessRuleRequestWithDefaults

`func NewInternalCreateAccessRuleRequestWithDefaults() *InternalCreateAccessRuleRequest`

NewInternalCreateAccessRuleRequestWithDefaults instantiates a new InternalCreateAccessRuleRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSubjectId

`func (o *InternalCreateAccessRuleRequest) GetSubjectId() string`

GetSubjectId returns the SubjectId field if non-nil, zero value otherwise.

### GetSubjectIdOk

`func (o *InternalCreateAccessRuleRequest) GetSubjectIdOk() (*string, bool)`

GetSubjectIdOk returns a tuple with the SubjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectId

`func (o *InternalCreateAccessRuleRequest) SetSubjectId(v string)`

SetSubjectId sets SubjectId field to given value.


### GetSubjectType

`func (o *InternalCreateAccessRuleRequest) GetSubjectType() SubjectType`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *InternalCreateAccessRuleRequest) GetSubjectTypeOk() (*SubjectType, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *InternalCreateAccessRuleRequest) SetSubjectType(v SubjectType)`

SetSubjectType sets SubjectType field to given value.


### GetScope

`func (o *InternalCreateAccessRuleRequest) GetScope() InternalCreateAccessRuleRequestScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *InternalCreateAccessRuleRequest) GetScopeOk() (*InternalCreateAccessRuleRequestScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *InternalCreateAccessRuleRequest) SetScope(v InternalCreateAccessRuleRequestScope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *InternalCreateAccessRuleRequest) HasScope() bool`

HasScope returns a boolean if a field has been set.

### SetScopeNil

`func (o *InternalCreateAccessRuleRequest) SetScopeNil(b bool)`

 SetScopeNil sets the value for Scope to be an explicit nil

### UnsetScope
`func (o *InternalCreateAccessRuleRequest) UnsetScope()`

UnsetScope ensures that no value is present for Scope, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


