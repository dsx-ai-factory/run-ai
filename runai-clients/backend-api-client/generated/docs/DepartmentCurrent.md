# DepartmentCurrent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | [**[]ClusterCurrentResourcesInner**](ClusterCurrentResourcesInner.md) |  | 
**ProjectResources** | [**[]DepartmentCurrentProjectResourcesInner**](DepartmentCurrentProjectResourcesInner.md) |  | 

## Methods

### NewDepartmentCurrent

`func NewDepartmentCurrent(resources []ClusterCurrentResourcesInner, projectResources []DepartmentCurrentProjectResourcesInner, ) *DepartmentCurrent`

NewDepartmentCurrent instantiates a new DepartmentCurrent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentCurrentWithDefaults

`func NewDepartmentCurrentWithDefaults() *DepartmentCurrent`

NewDepartmentCurrentWithDefaults instantiates a new DepartmentCurrent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *DepartmentCurrent) GetResources() []ClusterCurrentResourcesInner`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentCurrent) GetResourcesOk() (*[]ClusterCurrentResourcesInner, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentCurrent) SetResources(v []ClusterCurrentResourcesInner)`

SetResources sets Resources field to given value.


### GetProjectResources

`func (o *DepartmentCurrent) GetProjectResources() []DepartmentCurrentProjectResourcesInner`

GetProjectResources returns the ProjectResources field if non-nil, zero value otherwise.

### GetProjectResourcesOk

`func (o *DepartmentCurrent) GetProjectResourcesOk() (*[]DepartmentCurrentProjectResourcesInner, bool)`

GetProjectResourcesOk returns a tuple with the ProjectResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectResources

`func (o *DepartmentCurrent) SetProjectResources(v []DepartmentCurrentProjectResourcesInner)`

SetProjectResources sets ProjectResources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


