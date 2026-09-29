# DepartmentStatusUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Status** | [**DepartmentStatus**](DepartmentStatus.md) |  | 

## Methods

### NewDepartmentStatusUpdateRequest

`func NewDepartmentStatusUpdateRequest(status DepartmentStatus, ) *DepartmentStatusUpdateRequest`

NewDepartmentStatusUpdateRequest instantiates a new DepartmentStatusUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentStatusUpdateRequestWithDefaults

`func NewDepartmentStatusUpdateRequestWithDefaults() *DepartmentStatusUpdateRequest`

NewDepartmentStatusUpdateRequestWithDefaults instantiates a new DepartmentStatusUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *DepartmentStatusUpdateRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DepartmentStatusUpdateRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DepartmentStatusUpdateRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *DepartmentStatusUpdateRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetStatus

`func (o *DepartmentStatusUpdateRequest) GetStatus() DepartmentStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DepartmentStatusUpdateRequest) GetStatusOk() (*DepartmentStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DepartmentStatusUpdateRequest) SetStatus(v DepartmentStatus)`

SetStatus sets Status field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


