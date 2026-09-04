# DepartmentUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | 

## Methods

### NewDepartmentUpdateRequest

`func NewDepartmentUpdateRequest(resources []Resources, ) *DepartmentUpdateRequest`

NewDepartmentUpdateRequest instantiates a new DepartmentUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentUpdateRequestWithDefaults

`func NewDepartmentUpdateRequestWithDefaults() *DepartmentUpdateRequest`

NewDepartmentUpdateRequestWithDefaults instantiates a new DepartmentUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *DepartmentUpdateRequest) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentUpdateRequest) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentUpdateRequest) SetResources(v []Resources)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


