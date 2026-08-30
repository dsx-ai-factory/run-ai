# DepartmentAccessControl

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DepartmentId** | Pointer to **int32** | The unique id identifying the department. | [optional] 
**DepartmentAdmins** | Pointer to **[]string** | Id&#39;s of users with department admin role that are assigned to managed the department | [optional] 

## Methods

### NewDepartmentAccessControl

`func NewDepartmentAccessControl() *DepartmentAccessControl`

NewDepartmentAccessControl instantiates a new DepartmentAccessControl object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentAccessControlWithDefaults

`func NewDepartmentAccessControlWithDefaults() *DepartmentAccessControl`

NewDepartmentAccessControlWithDefaults instantiates a new DepartmentAccessControl object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDepartmentId

`func (o *DepartmentAccessControl) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *DepartmentAccessControl) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *DepartmentAccessControl) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *DepartmentAccessControl) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetDepartmentAdmins

`func (o *DepartmentAccessControl) GetDepartmentAdmins() []string`

GetDepartmentAdmins returns the DepartmentAdmins field if non-nil, zero value otherwise.

### GetDepartmentAdminsOk

`func (o *DepartmentAccessControl) GetDepartmentAdminsOk() (*[]string, bool)`

GetDepartmentAdminsOk returns a tuple with the DepartmentAdmins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentAdmins

`func (o *DepartmentAccessControl) SetDepartmentAdmins(v []string)`

SetDepartmentAdmins sets DepartmentAdmins field to given value.

### HasDepartmentAdmins

`func (o *DepartmentAccessControl) HasDepartmentAdmins() bool`

HasDepartmentAdmins returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


