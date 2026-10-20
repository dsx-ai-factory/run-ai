# WorkloadActionRecord

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | A unique ID of the workload action-history record. | 
**WorkloadId** | **string** | A unique ID of the workload. | 
**WorkloadName** | **string** | The name of the workload. | 
**Subject** | **string** | The authenticated subject that performed the workload action. | 
**SubjectType** | **string** | The authenticated subject type, preserved as provided by the identity context. Examples include user, service-account, and app, with casing determined by the identity provider. | 
**Action** | [**WorkloadActionType**](WorkloadActionType.md) |  | 
**AcceptedAt** | **time.Time** | The time when the workload action was accepted. | 

## Methods

### NewWorkloadActionRecord

`func NewWorkloadActionRecord(id string, workloadId string, workloadName string, subject string, subjectType string, action WorkloadActionType, acceptedAt time.Time, ) *WorkloadActionRecord`

NewWorkloadActionRecord instantiates a new WorkloadActionRecord object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadActionRecordWithDefaults

`func NewWorkloadActionRecordWithDefaults() *WorkloadActionRecord`

NewWorkloadActionRecordWithDefaults instantiates a new WorkloadActionRecord object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadActionRecord) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadActionRecord) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadActionRecord) SetId(v string)`

SetId sets Id field to given value.


### GetWorkloadId

`func (o *WorkloadActionRecord) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *WorkloadActionRecord) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *WorkloadActionRecord) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetWorkloadName

`func (o *WorkloadActionRecord) GetWorkloadName() string`

GetWorkloadName returns the WorkloadName field if non-nil, zero value otherwise.

### GetWorkloadNameOk

`func (o *WorkloadActionRecord) GetWorkloadNameOk() (*string, bool)`

GetWorkloadNameOk returns a tuple with the WorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadName

`func (o *WorkloadActionRecord) SetWorkloadName(v string)`

SetWorkloadName sets WorkloadName field to given value.


### GetSubject

`func (o *WorkloadActionRecord) GetSubject() string`

GetSubject returns the Subject field if non-nil, zero value otherwise.

### GetSubjectOk

`func (o *WorkloadActionRecord) GetSubjectOk() (*string, bool)`

GetSubjectOk returns a tuple with the Subject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubject

`func (o *WorkloadActionRecord) SetSubject(v string)`

SetSubject sets Subject field to given value.


### GetSubjectType

`func (o *WorkloadActionRecord) GetSubjectType() string`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *WorkloadActionRecord) GetSubjectTypeOk() (*string, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *WorkloadActionRecord) SetSubjectType(v string)`

SetSubjectType sets SubjectType field to given value.


### GetAction

`func (o *WorkloadActionRecord) GetAction() WorkloadActionType`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *WorkloadActionRecord) GetActionOk() (*WorkloadActionType, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *WorkloadActionRecord) SetAction(v WorkloadActionType)`

SetAction sets Action field to given value.


### GetAcceptedAt

`func (o *WorkloadActionRecord) GetAcceptedAt() time.Time`

GetAcceptedAt returns the AcceptedAt field if non-nil, zero value otherwise.

### GetAcceptedAtOk

`func (o *WorkloadActionRecord) GetAcceptedAtOk() (*time.Time, bool)`

GetAcceptedAtOk returns a tuple with the AcceptedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAcceptedAt

`func (o *WorkloadActionRecord) SetAcceptedAt(v time.Time)`

SetAcceptedAt sets AcceptedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


