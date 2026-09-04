# ProjectV1AllOfStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Namespace** | Pointer to **string** | The namespace of the project | [optional] 
**Message** | Pointer to **string** | Detailed phase message | [optional] 
**Phase** | Pointer to **string** | project&#39;s phase | [optional] 
**Conditions** | Pointer to **[]map[string]string** |  | [optional] 
**QuotaStatuses** | [**[]QuotaStatusV1Inner**](QuotaStatusV1Inner.md) | The current quota status for each node pool. Supported only for cluster versions 2.15+ or higher. | 

## Methods

### NewProjectV1AllOfStatus

`func NewProjectV1AllOfStatus(quotaStatuses []QuotaStatusV1Inner, ) *ProjectV1AllOfStatus`

NewProjectV1AllOfStatus instantiates a new ProjectV1AllOfStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1AllOfStatusWithDefaults

`func NewProjectV1AllOfStatusWithDefaults() *ProjectV1AllOfStatus`

NewProjectV1AllOfStatusWithDefaults instantiates a new ProjectV1AllOfStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNamespace

`func (o *ProjectV1AllOfStatus) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectV1AllOfStatus) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectV1AllOfStatus) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectV1AllOfStatus) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetMessage

`func (o *ProjectV1AllOfStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ProjectV1AllOfStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ProjectV1AllOfStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ProjectV1AllOfStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetPhase

`func (o *ProjectV1AllOfStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *ProjectV1AllOfStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *ProjectV1AllOfStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *ProjectV1AllOfStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetConditions

`func (o *ProjectV1AllOfStatus) GetConditions() []map[string]string`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ProjectV1AllOfStatus) GetConditionsOk() (*[]map[string]string, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ProjectV1AllOfStatus) SetConditions(v []map[string]string)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ProjectV1AllOfStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *ProjectV1AllOfStatus) GetQuotaStatuses() []QuotaStatusV1Inner`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *ProjectV1AllOfStatus) GetQuotaStatusesOk() (*[]QuotaStatusV1Inner, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *ProjectV1AllOfStatus) SetQuotaStatuses(v []QuotaStatusV1Inner)`

SetQuotaStatuses sets QuotaStatuses field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


