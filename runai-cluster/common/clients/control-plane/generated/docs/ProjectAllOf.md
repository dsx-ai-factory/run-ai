# ProjectAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**ParentId** | Pointer to **NullableString** | department parent uuid | [optional] 
**RequestedNamespace** | Pointer to **NullableString** | the project&#39;s namespace only if requested | [optional] 
**Status** | Pointer to [**ProjectAllOfStatus**](ProjectAllOfStatus.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Parent** | Pointer to [**NullableDepartmentProjectParent**](DepartmentProjectParent.md) |  | [optional] 

## Methods

### NewProjectAllOf

`func NewProjectAllOf() *ProjectAllOf`

NewProjectAllOf instantiates a new ProjectAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectAllOfWithDefaults

`func NewProjectAllOfWithDefaults() *ProjectAllOf`

NewProjectAllOfWithDefaults instantiates a new ProjectAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectAllOf) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectAllOf) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectAllOf) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ProjectAllOf) HasId() bool`

HasId returns a boolean if a field has been set.

### GetParentId

`func (o *ProjectAllOf) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *ProjectAllOf) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *ProjectAllOf) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *ProjectAllOf) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### SetParentIdNil

`func (o *ProjectAllOf) SetParentIdNil(b bool)`

 SetParentIdNil sets the value for ParentId to be an explicit nil

### UnsetParentId
`func (o *ProjectAllOf) UnsetParentId()`

UnsetParentId ensures that no value is present for ParentId, not even an explicit nil
### GetRequestedNamespace

`func (o *ProjectAllOf) GetRequestedNamespace() string`

GetRequestedNamespace returns the RequestedNamespace field if non-nil, zero value otherwise.

### GetRequestedNamespaceOk

`func (o *ProjectAllOf) GetRequestedNamespaceOk() (*string, bool)`

GetRequestedNamespaceOk returns a tuple with the RequestedNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNamespace

`func (o *ProjectAllOf) SetRequestedNamespace(v string)`

SetRequestedNamespace sets RequestedNamespace field to given value.

### HasRequestedNamespace

`func (o *ProjectAllOf) HasRequestedNamespace() bool`

HasRequestedNamespace returns a boolean if a field has been set.

### SetRequestedNamespaceNil

`func (o *ProjectAllOf) SetRequestedNamespaceNil(b bool)`

 SetRequestedNamespaceNil sets the value for RequestedNamespace to be an explicit nil

### UnsetRequestedNamespace
`func (o *ProjectAllOf) UnsetRequestedNamespace()`

UnsetRequestedNamespace ensures that no value is present for RequestedNamespace, not even an explicit nil
### GetStatus

`func (o *ProjectAllOf) GetStatus() ProjectAllOfStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProjectAllOf) GetStatusOk() (*ProjectAllOfStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProjectAllOf) SetStatus(v ProjectAllOfStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProjectAllOf) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ProjectAllOf) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ProjectAllOf) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ProjectAllOf) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ProjectAllOf) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *ProjectAllOf) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ProjectAllOf) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ProjectAllOf) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *ProjectAllOf) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *ProjectAllOf) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *ProjectAllOf) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *ProjectAllOf) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *ProjectAllOf) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *ProjectAllOf) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *ProjectAllOf) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *ProjectAllOf) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *ProjectAllOf) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetParent

`func (o *ProjectAllOf) GetParent() DepartmentProjectParent`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *ProjectAllOf) GetParentOk() (*DepartmentProjectParent, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *ProjectAllOf) SetParent(v DepartmentProjectParent)`

SetParent sets Parent field to given value.

### HasParent

`func (o *ProjectAllOf) HasParent() bool`

HasParent returns a boolean if a field has been set.

### SetParentNil

`func (o *ProjectAllOf) SetParentNil(b bool)`

 SetParentNil sets the value for Parent to be an explicit nil

### UnsetParent
`func (o *ProjectAllOf) UnsetParent()`

UnsetParent ensures that no value is present for Parent, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


