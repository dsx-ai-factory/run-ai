# SharedScopeAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ScopeName** | **string** | The name of the scope | 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 

## Methods

### NewSharedScopeAllOf

`func NewSharedScopeAllOf(scopeName string, ) *SharedScopeAllOf`

NewSharedScopeAllOf instantiates a new SharedScopeAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSharedScopeAllOfWithDefaults

`func NewSharedScopeAllOfWithDefaults() *SharedScopeAllOf`

NewSharedScopeAllOfWithDefaults instantiates a new SharedScopeAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScopeName

`func (o *SharedScopeAllOf) GetScopeName() string`

GetScopeName returns the ScopeName field if non-nil, zero value otherwise.

### GetScopeNameOk

`func (o *SharedScopeAllOf) GetScopeNameOk() (*string, bool)`

GetScopeNameOk returns a tuple with the ScopeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeName

`func (o *SharedScopeAllOf) SetScopeName(v string)`

SetScopeName sets ScopeName field to given value.


### GetCreatedAt

`func (o *SharedScopeAllOf) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SharedScopeAllOf) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SharedScopeAllOf) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *SharedScopeAllOf) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *SharedScopeAllOf) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *SharedScopeAllOf) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *SharedScopeAllOf) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *SharedScopeAllOf) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


