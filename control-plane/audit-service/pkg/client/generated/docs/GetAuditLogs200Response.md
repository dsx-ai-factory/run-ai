# GetAuditLogs200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Total** | Pointer to **int32** |  | [optional] 
**Next** | Pointer to **int32** |  | [optional] 
**AuditLogs** | Pointer to [**[]AuditLog**](AuditLog.md) |  | [optional] 

## Methods

### NewGetAuditLogs200Response

`func NewGetAuditLogs200Response() *GetAuditLogs200Response`

NewGetAuditLogs200Response instantiates a new GetAuditLogs200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetAuditLogs200ResponseWithDefaults

`func NewGetAuditLogs200ResponseWithDefaults() *GetAuditLogs200Response`

NewGetAuditLogs200ResponseWithDefaults instantiates a new GetAuditLogs200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTotal

`func (o *GetAuditLogs200Response) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *GetAuditLogs200Response) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *GetAuditLogs200Response) SetTotal(v int32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *GetAuditLogs200Response) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

### GetNext

`func (o *GetAuditLogs200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetAuditLogs200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetAuditLogs200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetAuditLogs200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetAuditLogs

`func (o *GetAuditLogs200Response) GetAuditLogs() []AuditLog`

GetAuditLogs returns the AuditLogs field if non-nil, zero value otherwise.

### GetAuditLogsOk

`func (o *GetAuditLogs200Response) GetAuditLogsOk() (*[]AuditLog, bool)`

GetAuditLogsOk returns a tuple with the AuditLogs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuditLogs

`func (o *GetAuditLogs200Response) SetAuditLogs(v []AuditLog)`

SetAuditLogs sets AuditLogs field to given value.

### HasAuditLogs

`func (o *GetAuditLogs200Response) HasAuditLogs() bool`

HasAuditLogs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


