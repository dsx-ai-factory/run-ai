# ReportCalculatedFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**CreatedBy** | **string** |  | 
**TenantId** | **int32** | ID of the tenant where the project is located. | 
**Status** | [**ReportStatus**](ReportStatus.md) |  | 
**StatusUpdatedAt** | **time.Time** |  | 
**StatusMessage** | Pointer to **string** |  | [optional] 

## Methods

### NewReportCalculatedFields

`func NewReportCalculatedFields(id string, createdAt time.Time, createdBy string, tenantId int32, status ReportStatus, statusUpdatedAt time.Time, ) *ReportCalculatedFields`

NewReportCalculatedFields instantiates a new ReportCalculatedFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReportCalculatedFieldsWithDefaults

`func NewReportCalculatedFieldsWithDefaults() *ReportCalculatedFields`

NewReportCalculatedFieldsWithDefaults instantiates a new ReportCalculatedFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ReportCalculatedFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ReportCalculatedFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ReportCalculatedFields) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *ReportCalculatedFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ReportCalculatedFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ReportCalculatedFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCreatedBy

`func (o *ReportCalculatedFields) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *ReportCalculatedFields) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *ReportCalculatedFields) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetTenantId

`func (o *ReportCalculatedFields) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ReportCalculatedFields) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ReportCalculatedFields) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetStatus

`func (o *ReportCalculatedFields) GetStatus() ReportStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ReportCalculatedFields) GetStatusOk() (*ReportStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ReportCalculatedFields) SetStatus(v ReportStatus)`

SetStatus sets Status field to given value.


### GetStatusUpdatedAt

`func (o *ReportCalculatedFields) GetStatusUpdatedAt() time.Time`

GetStatusUpdatedAt returns the StatusUpdatedAt field if non-nil, zero value otherwise.

### GetStatusUpdatedAtOk

`func (o *ReportCalculatedFields) GetStatusUpdatedAtOk() (*time.Time, bool)`

GetStatusUpdatedAtOk returns a tuple with the StatusUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusUpdatedAt

`func (o *ReportCalculatedFields) SetStatusUpdatedAt(v time.Time)`

SetStatusUpdatedAt sets StatusUpdatedAt field to given value.


### GetStatusMessage

`func (o *ReportCalculatedFields) GetStatusMessage() string`

GetStatusMessage returns the StatusMessage field if non-nil, zero value otherwise.

### GetStatusMessageOk

`func (o *ReportCalculatedFields) GetStatusMessageOk() (*string, bool)`

GetStatusMessageOk returns a tuple with the StatusMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusMessage

`func (o *ReportCalculatedFields) SetStatusMessage(v string)`

SetStatusMessage sets StatusMessage field to given value.

### HasStatusMessage

`func (o *ReportCalculatedFields) HasStatusMessage() bool`

HasStatusMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


