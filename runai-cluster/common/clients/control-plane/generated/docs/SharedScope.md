# SharedScope

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScopeType** | [**ScopeType2**](ScopeType2.md) |  | 
**ScopeId** | **string** |  | 
**ScopeName** | **string** | The name of the scope | 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 

## Methods

### NewSharedScope

`func NewSharedScope(scopeType ScopeType2, scopeId string, scopeName string, ) *SharedScope`

NewSharedScope instantiates a new SharedScope object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSharedScopeWithDefaults

`func NewSharedScopeWithDefaults() *SharedScope`

NewSharedScopeWithDefaults instantiates a new SharedScope object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScopeType

`func (o *SharedScope) GetScopeType() ScopeType2`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *SharedScope) GetScopeTypeOk() (*ScopeType2, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *SharedScope) SetScopeType(v ScopeType2)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *SharedScope) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *SharedScope) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *SharedScope) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeName

`func (o *SharedScope) GetScopeName() string`

GetScopeName returns the ScopeName field if non-nil, zero value otherwise.

### GetScopeNameOk

`func (o *SharedScope) GetScopeNameOk() (*string, bool)`

GetScopeNameOk returns a tuple with the ScopeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeName

`func (o *SharedScope) SetScopeName(v string)`

SetScopeName sets ScopeName field to given value.


### GetCreatedAt

`func (o *SharedScope) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SharedScope) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SharedScope) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *SharedScope) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *SharedScope) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *SharedScope) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *SharedScope) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *SharedScope) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


