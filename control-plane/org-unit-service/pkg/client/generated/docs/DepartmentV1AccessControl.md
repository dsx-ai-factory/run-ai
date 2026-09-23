# DepartmentV1AccessControl

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DepartmentId** | Pointer to **int32** | The unique id identifying the department. | [optional] 
**DepartmentAdmins** | Pointer to **[]string** | Id&#39;s of users with department admin role that are assigned to managed the department | [optional] 

## Methods

### NewDepartmentV1AccessControl

`func NewDepartmentV1AccessControl() *DepartmentV1AccessControl`

NewDepartmentV1AccessControl instantiates a new DepartmentV1AccessControl object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentV1AccessControlWithDefaults

`func NewDepartmentV1AccessControlWithDefaults() *DepartmentV1AccessControl`

NewDepartmentV1AccessControlWithDefaults instantiates a new DepartmentV1AccessControl object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDepartmentId

`func (o *DepartmentV1AccessControl) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *DepartmentV1AccessControl) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *DepartmentV1AccessControl) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *DepartmentV1AccessControl) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetDepartmentAdmins

`func (o *DepartmentV1AccessControl) GetDepartmentAdmins() []string`

GetDepartmentAdmins returns the DepartmentAdmins field if non-nil, zero value otherwise.

### GetDepartmentAdminsOk

`func (o *DepartmentV1AccessControl) GetDepartmentAdminsOk() (*[]string, bool)`

GetDepartmentAdminsOk returns a tuple with the DepartmentAdmins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentAdmins

`func (o *DepartmentV1AccessControl) SetDepartmentAdmins(v []string)`

SetDepartmentAdmins sets DepartmentAdmins field to given value.

### HasDepartmentAdmins

`func (o *DepartmentV1AccessControl) HasDepartmentAdmins() bool`

HasDepartmentAdmins returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


