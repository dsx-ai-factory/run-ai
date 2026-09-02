# ProjectStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Namespace** | Pointer to **string** | The namespace of the project | [optional] 
**Message** | Pointer to **string** | Detailed phase message | [optional] 
**QuotaStatuses** | Pointer to [**[]NodePoolQuotaStatus**](NodePoolQuotaStatus.md) | The current quota status for each nodepool. Supported only for cluster versions 2.15+. | [optional] 

## Methods

### NewProjectStatus

`func NewProjectStatus() *ProjectStatus`

NewProjectStatus instantiates a new ProjectStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectStatusWithDefaults

`func NewProjectStatusWithDefaults() *ProjectStatus`

NewProjectStatusWithDefaults instantiates a new ProjectStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNamespace

`func (o *ProjectStatus) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectStatus) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectStatus) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectStatus) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetMessage

`func (o *ProjectStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ProjectStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ProjectStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ProjectStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *ProjectStatus) GetQuotaStatuses() []NodePoolQuotaStatus`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *ProjectStatus) GetQuotaStatusesOk() (*[]NodePoolQuotaStatus, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *ProjectStatus) SetQuotaStatuses(v []NodePoolQuotaStatus)`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *ProjectStatus) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


