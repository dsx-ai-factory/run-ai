# OrgUnit

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Type** | [**ScopeType**](ScopeType.md) |  | 
**Name** | **string** |  | 
**ParentId** | Pointer to **string** |  | [optional] 
**ParentType** | Pointer to [**ScopeType**](ScopeType.md) |  | [optional] 

## Methods

### NewOrgUnit

`func NewOrgUnit(id string, type_ ScopeType, name string, ) *OrgUnit`

NewOrgUnit instantiates a new OrgUnit object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrgUnitWithDefaults

`func NewOrgUnitWithDefaults() *OrgUnit`

NewOrgUnitWithDefaults instantiates a new OrgUnit object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *OrgUnit) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *OrgUnit) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *OrgUnit) SetId(v string)`

SetId sets Id field to given value.


### GetType

`func (o *OrgUnit) GetType() ScopeType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *OrgUnit) GetTypeOk() (*ScopeType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *OrgUnit) SetType(v ScopeType)`

SetType sets Type field to given value.


### GetName

`func (o *OrgUnit) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *OrgUnit) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *OrgUnit) SetName(v string)`

SetName sets Name field to given value.


### GetParentId

`func (o *OrgUnit) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *OrgUnit) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *OrgUnit) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *OrgUnit) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### GetParentType

`func (o *OrgUnit) GetParentType() ScopeType`

GetParentType returns the ParentType field if non-nil, zero value otherwise.

### GetParentTypeOk

`func (o *OrgUnit) GetParentTypeOk() (*ScopeType, bool)`

GetParentTypeOk returns a tuple with the ParentType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentType

`func (o *OrgUnit) SetParentType(v ScopeType)`

SetParentType sets ParentType field to given value.

### HasParentType

`func (o *OrgUnit) HasParentType() bool`

HasParentType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


