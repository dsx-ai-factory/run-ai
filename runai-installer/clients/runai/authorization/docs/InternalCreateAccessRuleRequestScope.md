# InternalCreateAccessRuleRequestScope

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Type** | [**ScopeType**](ScopeType.md) |  | 
**Name** | Pointer to **string** |  | [optional] 

## Methods

### NewInternalCreateAccessRuleRequestScope

`func NewInternalCreateAccessRuleRequestScope(id string, type_ ScopeType, ) *InternalCreateAccessRuleRequestScope`

NewInternalCreateAccessRuleRequestScope instantiates a new InternalCreateAccessRuleRequestScope object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalCreateAccessRuleRequestScopeWithDefaults

`func NewInternalCreateAccessRuleRequestScopeWithDefaults() *InternalCreateAccessRuleRequestScope`

NewInternalCreateAccessRuleRequestScopeWithDefaults instantiates a new InternalCreateAccessRuleRequestScope object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *InternalCreateAccessRuleRequestScope) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *InternalCreateAccessRuleRequestScope) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *InternalCreateAccessRuleRequestScope) SetId(v string)`

SetId sets Id field to given value.


### GetType

`func (o *InternalCreateAccessRuleRequestScope) GetType() ScopeType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *InternalCreateAccessRuleRequestScope) GetTypeOk() (*ScopeType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *InternalCreateAccessRuleRequestScope) SetType(v ScopeType)`

SetType sets Type field to given value.


### GetName

`func (o *InternalCreateAccessRuleRequestScope) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InternalCreateAccessRuleRequestScope) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InternalCreateAccessRuleRequestScope) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *InternalCreateAccessRuleRequestScope) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


