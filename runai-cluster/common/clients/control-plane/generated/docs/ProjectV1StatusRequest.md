# ProjectV1StatusRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to **string** | A json representing the project&#39;s status | [optional] 
**Name** | Pointer to **string** | Project Name | [optional] 
**DepartmentName** | Pointer to **string** | name of the department owning the project | [optional] 
**QuotaStatuses** | Pointer to [**[]QuotaStatusV1Inner**](QuotaStatusV1Inner.md) | The current quota status for each node pool. Supported only for cluster versions 2.15+ or higher. | [optional] 

## Methods

### NewProjectV1StatusRequest

`func NewProjectV1StatusRequest() *ProjectV1StatusRequest`

NewProjectV1StatusRequest instantiates a new ProjectV1StatusRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1StatusRequestWithDefaults

`func NewProjectV1StatusRequestWithDefaults() *ProjectV1StatusRequest`

NewProjectV1StatusRequestWithDefaults instantiates a new ProjectV1StatusRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *ProjectV1StatusRequest) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProjectV1StatusRequest) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProjectV1StatusRequest) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProjectV1StatusRequest) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetName

`func (o *ProjectV1StatusRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectV1StatusRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectV1StatusRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProjectV1StatusRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDepartmentName

`func (o *ProjectV1StatusRequest) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *ProjectV1StatusRequest) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *ProjectV1StatusRequest) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *ProjectV1StatusRequest) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *ProjectV1StatusRequest) GetQuotaStatuses() []QuotaStatusV1Inner`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *ProjectV1StatusRequest) GetQuotaStatusesOk() (*[]QuotaStatusV1Inner, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *ProjectV1StatusRequest) SetQuotaStatuses(v []QuotaStatusV1Inner)`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *ProjectV1StatusRequest) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


