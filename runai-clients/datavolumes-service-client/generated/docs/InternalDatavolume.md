# InternalDatavolume

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**Scope** | [**ScopeType**](ScopeType.md) |  | 
**DepartmentId** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**OriginPvcName** | **string** | The name of the PVC that the datavolume is based on | 
**SharedScopes** | Pointer to [**[]SharedScope**](SharedScope.md) | Will be returned only if the user has the required permissions to view those scopes | [optional] 
**Status** | [**DatavolumeStatus**](DatavolumeStatus.md) |  | 

## Methods

### NewInternalDatavolume

`func NewInternalDatavolume(id string, name string, scope ScopeType, originPvcName string, status DatavolumeStatus, ) *InternalDatavolume`

NewInternalDatavolume instantiates a new InternalDatavolume object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalDatavolumeWithDefaults

`func NewInternalDatavolumeWithDefaults() *InternalDatavolume`

NewInternalDatavolumeWithDefaults instantiates a new InternalDatavolume object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *InternalDatavolume) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *InternalDatavolume) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *InternalDatavolume) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *InternalDatavolume) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *InternalDatavolume) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *InternalDatavolume) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *InternalDatavolume) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *InternalDatavolume) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *InternalDatavolume) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *InternalDatavolume) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *InternalDatavolume) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *InternalDatavolume) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *InternalDatavolume) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *InternalDatavolume) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *InternalDatavolume) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *InternalDatavolume) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *InternalDatavolume) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *InternalDatavolume) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *InternalDatavolume) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetName

`func (o *InternalDatavolume) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InternalDatavolume) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InternalDatavolume) SetName(v string)`

SetName sets Name field to given value.


### GetScope

`func (o *InternalDatavolume) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *InternalDatavolume) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *InternalDatavolume) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetDepartmentId

`func (o *InternalDatavolume) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *InternalDatavolume) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *InternalDatavolume) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *InternalDatavolume) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *InternalDatavolume) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *InternalDatavolume) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *InternalDatavolume) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *InternalDatavolume) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetDescription

`func (o *InternalDatavolume) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *InternalDatavolume) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *InternalDatavolume) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *InternalDatavolume) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOriginPvcName

`func (o *InternalDatavolume) GetOriginPvcName() string`

GetOriginPvcName returns the OriginPvcName field if non-nil, zero value otherwise.

### GetOriginPvcNameOk

`func (o *InternalDatavolume) GetOriginPvcNameOk() (*string, bool)`

GetOriginPvcNameOk returns a tuple with the OriginPvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPvcName

`func (o *InternalDatavolume) SetOriginPvcName(v string)`

SetOriginPvcName sets OriginPvcName field to given value.


### GetSharedScopes

`func (o *InternalDatavolume) GetSharedScopes() []SharedScope`

GetSharedScopes returns the SharedScopes field if non-nil, zero value otherwise.

### GetSharedScopesOk

`func (o *InternalDatavolume) GetSharedScopesOk() (*[]SharedScope, bool)`

GetSharedScopesOk returns a tuple with the SharedScopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSharedScopes

`func (o *InternalDatavolume) SetSharedScopes(v []SharedScope)`

SetSharedScopes sets SharedScopes field to given value.

### HasSharedScopes

`func (o *InternalDatavolume) HasSharedScopes() bool`

HasSharedScopes returns a boolean if a field has been set.

### GetStatus

`func (o *InternalDatavolume) GetStatus() DatavolumeStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *InternalDatavolume) GetStatusOk() (*DatavolumeStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *InternalDatavolume) SetStatus(v DatavolumeStatus)`

SetStatus sets Status field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


