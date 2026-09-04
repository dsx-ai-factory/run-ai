# ProjectStatusRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to **string** | A json representing the project&#39;s status | [optional] 
**Name** | Pointer to **string** | Project Name | [optional] 
**DepartmentName** | Pointer to **string** | name of the department owning the project | [optional] 
**QuotaStatuses** | Pointer to [**[]NodePoolQuotaStatus**](NodePoolQuotaStatus.md) | The current quota status for each node pool. Supported only for cluster versions 2.15+ or higher. | [optional] 

## Methods

### NewProjectStatusRequest

`func NewProjectStatusRequest() *ProjectStatusRequest`

NewProjectStatusRequest instantiates a new ProjectStatusRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectStatusRequestWithDefaults

`func NewProjectStatusRequestWithDefaults() *ProjectStatusRequest`

NewProjectStatusRequestWithDefaults instantiates a new ProjectStatusRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *ProjectStatusRequest) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProjectStatusRequest) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProjectStatusRequest) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProjectStatusRequest) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetName

`func (o *ProjectStatusRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectStatusRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectStatusRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProjectStatusRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDepartmentName

`func (o *ProjectStatusRequest) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *ProjectStatusRequest) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *ProjectStatusRequest) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *ProjectStatusRequest) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *ProjectStatusRequest) GetQuotaStatuses() []NodePoolQuotaStatus`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *ProjectStatusRequest) GetQuotaStatusesOk() (*[]NodePoolQuotaStatus, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *ProjectStatusRequest) SetQuotaStatuses(v []NodePoolQuotaStatus)`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *ProjectStatusRequest) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


