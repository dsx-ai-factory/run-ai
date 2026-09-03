# WorkloadEvent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterUuid** | **string** | clusterUuid of the workload event | 
**TenantId** | **float32** | tenantId of the workload event | 
**Uid** | **string** | uid of the workload event | 
**Type** | **string** | type of the workload event | 
**Message** | **string** | message of the workload event | 
**Reason** | **string** | reason of the workload event | 
**SourceComponent** | **string** | the event source component | 
**InvolvedObjUID** | **string** | involvedObjUID of the workload event | 
**InvolvedObjKind** | **string** | involvedObjKind of the workload event | 
**InvolvedObjName** | **string** | involvedObjName of the workload event | 
**InvolvedObjNamespace** | **string** | involvedObjNamespace of the workload event | 

## Methods

### NewWorkloadEvent

`func NewWorkloadEvent(clusterUuid string, tenantId float32, uid string, type_ string, message string, reason string, sourceComponent string, involvedObjUID string, involvedObjKind string, involvedObjName string, involvedObjNamespace string, ) *WorkloadEvent`

NewWorkloadEvent instantiates a new WorkloadEvent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadEventWithDefaults

`func NewWorkloadEventWithDefaults() *WorkloadEvent`

NewWorkloadEventWithDefaults instantiates a new WorkloadEvent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterUuid

`func (o *WorkloadEvent) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *WorkloadEvent) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *WorkloadEvent) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.


### GetTenantId

`func (o *WorkloadEvent) GetTenantId() float32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *WorkloadEvent) GetTenantIdOk() (*float32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *WorkloadEvent) SetTenantId(v float32)`

SetTenantId sets TenantId field to given value.


### GetUid

`func (o *WorkloadEvent) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *WorkloadEvent) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *WorkloadEvent) SetUid(v string)`

SetUid sets Uid field to given value.


### GetType

`func (o *WorkloadEvent) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *WorkloadEvent) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *WorkloadEvent) SetType(v string)`

SetType sets Type field to given value.


### GetMessage

`func (o *WorkloadEvent) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *WorkloadEvent) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *WorkloadEvent) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetReason

`func (o *WorkloadEvent) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *WorkloadEvent) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *WorkloadEvent) SetReason(v string)`

SetReason sets Reason field to given value.


### GetSourceComponent

`func (o *WorkloadEvent) GetSourceComponent() string`

GetSourceComponent returns the SourceComponent field if non-nil, zero value otherwise.

### GetSourceComponentOk

`func (o *WorkloadEvent) GetSourceComponentOk() (*string, bool)`

GetSourceComponentOk returns a tuple with the SourceComponent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceComponent

`func (o *WorkloadEvent) SetSourceComponent(v string)`

SetSourceComponent sets SourceComponent field to given value.


### GetInvolvedObjUID

`func (o *WorkloadEvent) GetInvolvedObjUID() string`

GetInvolvedObjUID returns the InvolvedObjUID field if non-nil, zero value otherwise.

### GetInvolvedObjUIDOk

`func (o *WorkloadEvent) GetInvolvedObjUIDOk() (*string, bool)`

GetInvolvedObjUIDOk returns a tuple with the InvolvedObjUID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjUID

`func (o *WorkloadEvent) SetInvolvedObjUID(v string)`

SetInvolvedObjUID sets InvolvedObjUID field to given value.


### GetInvolvedObjKind

`func (o *WorkloadEvent) GetInvolvedObjKind() string`

GetInvolvedObjKind returns the InvolvedObjKind field if non-nil, zero value otherwise.

### GetInvolvedObjKindOk

`func (o *WorkloadEvent) GetInvolvedObjKindOk() (*string, bool)`

GetInvolvedObjKindOk returns a tuple with the InvolvedObjKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjKind

`func (o *WorkloadEvent) SetInvolvedObjKind(v string)`

SetInvolvedObjKind sets InvolvedObjKind field to given value.


### GetInvolvedObjName

`func (o *WorkloadEvent) GetInvolvedObjName() string`

GetInvolvedObjName returns the InvolvedObjName field if non-nil, zero value otherwise.

### GetInvolvedObjNameOk

`func (o *WorkloadEvent) GetInvolvedObjNameOk() (*string, bool)`

GetInvolvedObjNameOk returns a tuple with the InvolvedObjName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjName

`func (o *WorkloadEvent) SetInvolvedObjName(v string)`

SetInvolvedObjName sets InvolvedObjName field to given value.


### GetInvolvedObjNamespace

`func (o *WorkloadEvent) GetInvolvedObjNamespace() string`

GetInvolvedObjNamespace returns the InvolvedObjNamespace field if non-nil, zero value otherwise.

### GetInvolvedObjNamespaceOk

`func (o *WorkloadEvent) GetInvolvedObjNamespaceOk() (*string, bool)`

GetInvolvedObjNamespaceOk returns a tuple with the InvolvedObjNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObjNamespace

`func (o *WorkloadEvent) SetInvolvedObjNamespace(v string)`

SetInvolvedObjNamespace sets InvolvedObjNamespace field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


