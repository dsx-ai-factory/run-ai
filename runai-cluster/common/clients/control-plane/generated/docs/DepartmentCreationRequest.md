# DepartmentCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | 
**Name** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 

## Methods

### NewDepartmentCreationRequest

`func NewDepartmentCreationRequest(resources []Resources, name string, clusterId string, ) *DepartmentCreationRequest`

NewDepartmentCreationRequest instantiates a new DepartmentCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentCreationRequestWithDefaults

`func NewDepartmentCreationRequestWithDefaults() *DepartmentCreationRequest`

NewDepartmentCreationRequestWithDefaults instantiates a new DepartmentCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *DepartmentCreationRequest) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentCreationRequest) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentCreationRequest) SetResources(v []Resources)`

SetResources sets Resources field to given value.


### GetName

`func (o *DepartmentCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *DepartmentCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DepartmentCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DepartmentCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


