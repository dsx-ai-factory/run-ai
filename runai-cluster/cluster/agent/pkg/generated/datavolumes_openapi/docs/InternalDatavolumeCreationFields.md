# InternalDatavolumeCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Scope** | [**ScopeType**](ScopeType.md) |  | 
**DepartmentId** | Pointer to **string** | The ID of the department that in its namespace the origin pvc is located (mandatory if scope is department) | [optional] 
**ClusterId** | Pointer to **string** | The ID of the cluster that in its namespace the origin pvc is located (mandatory if scope is cluster) | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**OriginPvcName** | **string** | The name of the PVC that the datavolume is based on | 

## Methods

### NewInternalDatavolumeCreationFields

`func NewInternalDatavolumeCreationFields(name string, scope ScopeType, originPvcName string, ) *InternalDatavolumeCreationFields`

NewInternalDatavolumeCreationFields instantiates a new InternalDatavolumeCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalDatavolumeCreationFieldsWithDefaults

`func NewInternalDatavolumeCreationFieldsWithDefaults() *InternalDatavolumeCreationFields`

NewInternalDatavolumeCreationFieldsWithDefaults instantiates a new InternalDatavolumeCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *InternalDatavolumeCreationFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InternalDatavolumeCreationFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InternalDatavolumeCreationFields) SetName(v string)`

SetName sets Name field to given value.


### GetScope

`func (o *InternalDatavolumeCreationFields) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *InternalDatavolumeCreationFields) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *InternalDatavolumeCreationFields) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetDepartmentId

`func (o *InternalDatavolumeCreationFields) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *InternalDatavolumeCreationFields) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *InternalDatavolumeCreationFields) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *InternalDatavolumeCreationFields) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *InternalDatavolumeCreationFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *InternalDatavolumeCreationFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *InternalDatavolumeCreationFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *InternalDatavolumeCreationFields) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetDescription

`func (o *InternalDatavolumeCreationFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *InternalDatavolumeCreationFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *InternalDatavolumeCreationFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *InternalDatavolumeCreationFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOriginPvcName

`func (o *InternalDatavolumeCreationFields) GetOriginPvcName() string`

GetOriginPvcName returns the OriginPvcName field if non-nil, zero value otherwise.

### GetOriginPvcNameOk

`func (o *InternalDatavolumeCreationFields) GetOriginPvcNameOk() (*string, bool)`

GetOriginPvcNameOk returns a tuple with the OriginPvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPvcName

`func (o *InternalDatavolumeCreationFields) SetOriginPvcName(v string)`

SetOriginPvcName sets OriginPvcName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


