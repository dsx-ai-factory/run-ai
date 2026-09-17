# DepartmentProjectParent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**Parent** | Pointer to [**NullableDepartmentProjectParent**](DepartmentProjectParent.md) |  | [optional] 

## Methods

### NewDepartmentProjectParent

`func NewDepartmentProjectParent(id string, name string, ) *DepartmentProjectParent`

NewDepartmentProjectParent instantiates a new DepartmentProjectParent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentProjectParentWithDefaults

`func NewDepartmentProjectParentWithDefaults() *DepartmentProjectParent`

NewDepartmentProjectParentWithDefaults instantiates a new DepartmentProjectParent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DepartmentProjectParent) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DepartmentProjectParent) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DepartmentProjectParent) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *DepartmentProjectParent) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentProjectParent) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentProjectParent) SetName(v string)`

SetName sets Name field to given value.


### GetParent

`func (o *DepartmentProjectParent) GetParent() DepartmentProjectParent`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *DepartmentProjectParent) GetParentOk() (*DepartmentProjectParent, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *DepartmentProjectParent) SetParent(v DepartmentProjectParent)`

SetParent sets Parent field to given value.

### HasParent

`func (o *DepartmentProjectParent) HasParent() bool`

HasParent returns a boolean if a field has been set.

### SetParentNil

`func (o *DepartmentProjectParent) SetParentNil(b bool)`

 SetParentNil sets the value for Parent to be an explicit nil

### UnsetParent
`func (o *DepartmentProjectParent) UnsetParent()`

UnsetParent ensures that no value is present for Parent, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


