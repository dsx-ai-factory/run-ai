# DataDepartmentFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | Pointer to [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 

## Methods

### NewDataDepartmentFields

`func NewDataDepartmentFields() *DataDepartmentFields`

NewDataDepartmentFields instantiates a new DataDepartmentFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataDepartmentFieldsWithDefaults

`func NewDataDepartmentFieldsWithDefaults() *DataDepartmentFields`

NewDataDepartmentFieldsWithDefaults instantiates a new DataDepartmentFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *DataDepartmentFields) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DataDepartmentFields) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DataDepartmentFields) SetResources(v []Resources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DataDepartmentFields) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetName

`func (o *DataDepartmentFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DataDepartmentFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DataDepartmentFields) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DataDepartmentFields) HasName() bool`

HasName returns a boolean if a field has been set.

### GetClusterId

`func (o *DataDepartmentFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DataDepartmentFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DataDepartmentFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *DataDepartmentFields) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


