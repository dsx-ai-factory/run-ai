# AuditLog

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Timestamp** | Pointer to **time.Time** |  | [optional] 
**TenantId** | **string** | the ID of the tenant | 
**Subject** | **string** | the user/app which triggered this API | 
**SubjectType** | **string** |  | [default to "User"]
**SourceIp** | Pointer to **string** |  | [optional] 
**Action** | **string** |  | 
**HttpMethod** | Pointer to **string** |  | [optional] 
**Url** | Pointer to **string** |  | [optional] 
**EntityType** | Pointer to **string** |  | [optional] 
**EntityName** | Pointer to **string** |  | [optional] 
**EntityId** | Pointer to **string** |  | [optional] 
**Result** | **string** |  | 
**HttpStatusCode** | Pointer to **int32** |  | [optional] 
**ClusterName** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** |  | [optional] 
**RequestId** | Pointer to **string** |  | [optional] 
**Metadata** | Pointer to **map[string]string** |  | [optional] 

## Methods

### NewAuditLog

`func NewAuditLog(tenantId string, subject string, subjectType string, action string, result string, ) *AuditLog`

NewAuditLog instantiates a new AuditLog object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuditLogWithDefaults

`func NewAuditLogWithDefaults() *AuditLog`

NewAuditLogWithDefaults instantiates a new AuditLog object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AuditLog) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AuditLog) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AuditLog) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *AuditLog) HasId() bool`

HasId returns a boolean if a field has been set.

### GetTimestamp

`func (o *AuditLog) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *AuditLog) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *AuditLog) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *AuditLog) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.

### GetTenantId

`func (o *AuditLog) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AuditLog) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AuditLog) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.


### GetSubject

`func (o *AuditLog) GetSubject() string`

GetSubject returns the Subject field if non-nil, zero value otherwise.

### GetSubjectOk

`func (o *AuditLog) GetSubjectOk() (*string, bool)`

GetSubjectOk returns a tuple with the Subject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubject

`func (o *AuditLog) SetSubject(v string)`

SetSubject sets Subject field to given value.


### GetSubjectType

`func (o *AuditLog) GetSubjectType() string`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *AuditLog) GetSubjectTypeOk() (*string, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *AuditLog) SetSubjectType(v string)`

SetSubjectType sets SubjectType field to given value.


### GetSourceIp

`func (o *AuditLog) GetSourceIp() string`

GetSourceIp returns the SourceIp field if non-nil, zero value otherwise.

### GetSourceIpOk

`func (o *AuditLog) GetSourceIpOk() (*string, bool)`

GetSourceIpOk returns a tuple with the SourceIp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceIp

`func (o *AuditLog) SetSourceIp(v string)`

SetSourceIp sets SourceIp field to given value.

### HasSourceIp

`func (o *AuditLog) HasSourceIp() bool`

HasSourceIp returns a boolean if a field has been set.

### GetAction

`func (o *AuditLog) GetAction() string`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *AuditLog) GetActionOk() (*string, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *AuditLog) SetAction(v string)`

SetAction sets Action field to given value.


### GetHttpMethod

`func (o *AuditLog) GetHttpMethod() string`

GetHttpMethod returns the HttpMethod field if non-nil, zero value otherwise.

### GetHttpMethodOk

`func (o *AuditLog) GetHttpMethodOk() (*string, bool)`

GetHttpMethodOk returns a tuple with the HttpMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHttpMethod

`func (o *AuditLog) SetHttpMethod(v string)`

SetHttpMethod sets HttpMethod field to given value.

### HasHttpMethod

`func (o *AuditLog) HasHttpMethod() bool`

HasHttpMethod returns a boolean if a field has been set.

### GetUrl

`func (o *AuditLog) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *AuditLog) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *AuditLog) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *AuditLog) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetEntityType

`func (o *AuditLog) GetEntityType() string`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *AuditLog) GetEntityTypeOk() (*string, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *AuditLog) SetEntityType(v string)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *AuditLog) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetEntityName

`func (o *AuditLog) GetEntityName() string`

GetEntityName returns the EntityName field if non-nil, zero value otherwise.

### GetEntityNameOk

`func (o *AuditLog) GetEntityNameOk() (*string, bool)`

GetEntityNameOk returns a tuple with the EntityName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityName

`func (o *AuditLog) SetEntityName(v string)`

SetEntityName sets EntityName field to given value.

### HasEntityName

`func (o *AuditLog) HasEntityName() bool`

HasEntityName returns a boolean if a field has been set.

### GetEntityId

`func (o *AuditLog) GetEntityId() string`

GetEntityId returns the EntityId field if non-nil, zero value otherwise.

### GetEntityIdOk

`func (o *AuditLog) GetEntityIdOk() (*string, bool)`

GetEntityIdOk returns a tuple with the EntityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityId

`func (o *AuditLog) SetEntityId(v string)`

SetEntityId sets EntityId field to given value.

### HasEntityId

`func (o *AuditLog) HasEntityId() bool`

HasEntityId returns a boolean if a field has been set.

### GetResult

`func (o *AuditLog) GetResult() string`

GetResult returns the Result field if non-nil, zero value otherwise.

### GetResultOk

`func (o *AuditLog) GetResultOk() (*string, bool)`

GetResultOk returns a tuple with the Result field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResult

`func (o *AuditLog) SetResult(v string)`

SetResult sets Result field to given value.


### GetHttpStatusCode

`func (o *AuditLog) GetHttpStatusCode() int32`

GetHttpStatusCode returns the HttpStatusCode field if non-nil, zero value otherwise.

### GetHttpStatusCodeOk

`func (o *AuditLog) GetHttpStatusCodeOk() (*int32, bool)`

GetHttpStatusCodeOk returns a tuple with the HttpStatusCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHttpStatusCode

`func (o *AuditLog) SetHttpStatusCode(v int32)`

SetHttpStatusCode sets HttpStatusCode field to given value.

### HasHttpStatusCode

`func (o *AuditLog) HasHttpStatusCode() bool`

HasHttpStatusCode returns a boolean if a field has been set.

### GetClusterName

`func (o *AuditLog) GetClusterName() string`

GetClusterName returns the ClusterName field if non-nil, zero value otherwise.

### GetClusterNameOk

`func (o *AuditLog) GetClusterNameOk() (*string, bool)`

GetClusterNameOk returns a tuple with the ClusterName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterName

`func (o *AuditLog) SetClusterName(v string)`

SetClusterName sets ClusterName field to given value.

### HasClusterName

`func (o *AuditLog) HasClusterName() bool`

HasClusterName returns a boolean if a field has been set.

### GetClusterId

`func (o *AuditLog) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AuditLog) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AuditLog) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AuditLog) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetRequestId

`func (o *AuditLog) GetRequestId() string`

GetRequestId returns the RequestId field if non-nil, zero value otherwise.

### GetRequestIdOk

`func (o *AuditLog) GetRequestIdOk() (*string, bool)`

GetRequestIdOk returns a tuple with the RequestId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestId

`func (o *AuditLog) SetRequestId(v string)`

SetRequestId sets RequestId field to given value.

### HasRequestId

`func (o *AuditLog) HasRequestId() bool`

HasRequestId returns a boolean if a field has been set.

### GetMetadata

`func (o *AuditLog) GetMetadata() map[string]string`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *AuditLog) GetMetadataOk() (*map[string]string, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *AuditLog) SetMetadata(v map[string]string)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *AuditLog) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


