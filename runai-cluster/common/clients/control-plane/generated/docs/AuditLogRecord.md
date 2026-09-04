# AuditLogRecord

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **int32** | The unique id of the audit log. | [optional] 
**ClusterUuid** | Pointer to **string** | The id of the cluster. | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**HappenedAt** | Pointer to **time.Time** | The date and time in which the event happened. | [optional] 
**Action** | Pointer to **string** | The action that was performed by the user. | [optional] 
**Version** | Pointer to **string** | The version of the audit log record. | [optional] 
**EntityId** | Pointer to **string** | The id of the action related entity. | [optional] 
**EntityType** | Pointer to **string** | The type of the action related entity. | [optional] 
**EntityName** | Pointer to **string** | The name of the action related entity. | [optional] 
**SourceType** | Pointer to **string** | The type of the source of the action. | [optional] 
**SourceId** | Pointer to **string** | The id of the source of the action. | [optional] 
**SourceName** | Pointer to **string** | The name of the source of the action. | [optional] 
**Error** | Pointer to **string** | In case of a failed action, the corresponding error | [optional] 
**Context** | Pointer to **map[string]interface{}** | The context of the action. | [optional] 
**Body** | Pointer to **map[string]interface{}** | The body of the action http request. | [optional] 

## Methods

### NewAuditLogRecord

`func NewAuditLogRecord() *AuditLogRecord`

NewAuditLogRecord instantiates a new AuditLogRecord object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuditLogRecordWithDefaults

`func NewAuditLogRecordWithDefaults() *AuditLogRecord`

NewAuditLogRecordWithDefaults instantiates a new AuditLogRecord object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AuditLogRecord) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AuditLogRecord) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AuditLogRecord) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *AuditLogRecord) HasId() bool`

HasId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *AuditLogRecord) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *AuditLogRecord) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *AuditLogRecord) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *AuditLogRecord) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetTenantId

`func (o *AuditLogRecord) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AuditLogRecord) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AuditLogRecord) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *AuditLogRecord) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetHappenedAt

`func (o *AuditLogRecord) GetHappenedAt() time.Time`

GetHappenedAt returns the HappenedAt field if non-nil, zero value otherwise.

### GetHappenedAtOk

`func (o *AuditLogRecord) GetHappenedAtOk() (*time.Time, bool)`

GetHappenedAtOk returns a tuple with the HappenedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHappenedAt

`func (o *AuditLogRecord) SetHappenedAt(v time.Time)`

SetHappenedAt sets HappenedAt field to given value.

### HasHappenedAt

`func (o *AuditLogRecord) HasHappenedAt() bool`

HasHappenedAt returns a boolean if a field has been set.

### GetAction

`func (o *AuditLogRecord) GetAction() string`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *AuditLogRecord) GetActionOk() (*string, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *AuditLogRecord) SetAction(v string)`

SetAction sets Action field to given value.

### HasAction

`func (o *AuditLogRecord) HasAction() bool`

HasAction returns a boolean if a field has been set.

### GetVersion

`func (o *AuditLogRecord) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *AuditLogRecord) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *AuditLogRecord) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *AuditLogRecord) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetEntityId

`func (o *AuditLogRecord) GetEntityId() string`

GetEntityId returns the EntityId field if non-nil, zero value otherwise.

### GetEntityIdOk

`func (o *AuditLogRecord) GetEntityIdOk() (*string, bool)`

GetEntityIdOk returns a tuple with the EntityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityId

`func (o *AuditLogRecord) SetEntityId(v string)`

SetEntityId sets EntityId field to given value.

### HasEntityId

`func (o *AuditLogRecord) HasEntityId() bool`

HasEntityId returns a boolean if a field has been set.

### GetEntityType

`func (o *AuditLogRecord) GetEntityType() string`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *AuditLogRecord) GetEntityTypeOk() (*string, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *AuditLogRecord) SetEntityType(v string)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *AuditLogRecord) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetEntityName

`func (o *AuditLogRecord) GetEntityName() string`

GetEntityName returns the EntityName field if non-nil, zero value otherwise.

### GetEntityNameOk

`func (o *AuditLogRecord) GetEntityNameOk() (*string, bool)`

GetEntityNameOk returns a tuple with the EntityName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityName

`func (o *AuditLogRecord) SetEntityName(v string)`

SetEntityName sets EntityName field to given value.

### HasEntityName

`func (o *AuditLogRecord) HasEntityName() bool`

HasEntityName returns a boolean if a field has been set.

### GetSourceType

`func (o *AuditLogRecord) GetSourceType() string`

GetSourceType returns the SourceType field if non-nil, zero value otherwise.

### GetSourceTypeOk

`func (o *AuditLogRecord) GetSourceTypeOk() (*string, bool)`

GetSourceTypeOk returns a tuple with the SourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceType

`func (o *AuditLogRecord) SetSourceType(v string)`

SetSourceType sets SourceType field to given value.

### HasSourceType

`func (o *AuditLogRecord) HasSourceType() bool`

HasSourceType returns a boolean if a field has been set.

### GetSourceId

`func (o *AuditLogRecord) GetSourceId() string`

GetSourceId returns the SourceId field if non-nil, zero value otherwise.

### GetSourceIdOk

`func (o *AuditLogRecord) GetSourceIdOk() (*string, bool)`

GetSourceIdOk returns a tuple with the SourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceId

`func (o *AuditLogRecord) SetSourceId(v string)`

SetSourceId sets SourceId field to given value.

### HasSourceId

`func (o *AuditLogRecord) HasSourceId() bool`

HasSourceId returns a boolean if a field has been set.

### GetSourceName

`func (o *AuditLogRecord) GetSourceName() string`

GetSourceName returns the SourceName field if non-nil, zero value otherwise.

### GetSourceNameOk

`func (o *AuditLogRecord) GetSourceNameOk() (*string, bool)`

GetSourceNameOk returns a tuple with the SourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceName

`func (o *AuditLogRecord) SetSourceName(v string)`

SetSourceName sets SourceName field to given value.

### HasSourceName

`func (o *AuditLogRecord) HasSourceName() bool`

HasSourceName returns a boolean if a field has been set.

### GetError

`func (o *AuditLogRecord) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *AuditLogRecord) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *AuditLogRecord) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *AuditLogRecord) HasError() bool`

HasError returns a boolean if a field has been set.

### GetContext

`func (o *AuditLogRecord) GetContext() map[string]interface{}`

GetContext returns the Context field if non-nil, zero value otherwise.

### GetContextOk

`func (o *AuditLogRecord) GetContextOk() (*map[string]interface{}, bool)`

GetContextOk returns a tuple with the Context field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContext

`func (o *AuditLogRecord) SetContext(v map[string]interface{})`

SetContext sets Context field to given value.

### HasContext

`func (o *AuditLogRecord) HasContext() bool`

HasContext returns a boolean if a field has been set.

### GetBody

`func (o *AuditLogRecord) GetBody() map[string]interface{}`

GetBody returns the Body field if non-nil, zero value otherwise.

### GetBodyOk

`func (o *AuditLogRecord) GetBodyOk() (*map[string]interface{}, bool)`

GetBodyOk returns a tuple with the Body field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBody

`func (o *AuditLogRecord) SetBody(v map[string]interface{})`

SetBody sets Body field to given value.

### HasBody

`func (o *AuditLogRecord) HasBody() bool`

HasBody returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


