# AccessRuleScopesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 

## Methods

### NewAccessRuleScopesInner

`func NewAccessRuleScopesInner(scopeId string, scopeType ScopeType, ) *AccessRuleScopesInner`

NewAccessRuleScopesInner instantiates a new AccessRuleScopesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleScopesInnerWithDefaults

`func NewAccessRuleScopesInnerWithDefaults() *AccessRuleScopesInner`

NewAccessRuleScopesInnerWithDefaults instantiates a new AccessRuleScopesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScopeId

`func (o *AccessRuleScopesInner) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AccessRuleScopesInner) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AccessRuleScopesInner) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *AccessRuleScopesInner) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AccessRuleScopesInner) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AccessRuleScopesInner) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


