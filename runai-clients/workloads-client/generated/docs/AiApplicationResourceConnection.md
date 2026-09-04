# AiApplicationResourceConnection

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **string** | The Kubernetes UID of the referenced resource. | [optional] 
**Name** | **string** | The name of the referenced resource. | 
**Kind** | [**ResourceRefKind**](ResourceRefKind.md) |  | 
**WorkloadId** | Pointer to **string** | A unique ID of the workload. | [optional] 

## Methods

### NewAiApplicationResourceConnection

`func NewAiApplicationResourceConnection(name string, kind ResourceRefKind, ) *AiApplicationResourceConnection`

NewAiApplicationResourceConnection instantiates a new AiApplicationResourceConnection object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAiApplicationResourceConnectionWithDefaults

`func NewAiApplicationResourceConnectionWithDefaults() *AiApplicationResourceConnection`

NewAiApplicationResourceConnectionWithDefaults instantiates a new AiApplicationResourceConnection object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *AiApplicationResourceConnection) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *AiApplicationResourceConnection) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *AiApplicationResourceConnection) SetUid(v string)`

SetUid sets Uid field to given value.

### HasUid

`func (o *AiApplicationResourceConnection) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetName

`func (o *AiApplicationResourceConnection) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AiApplicationResourceConnection) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AiApplicationResourceConnection) SetName(v string)`

SetName sets Name field to given value.


### GetKind

`func (o *AiApplicationResourceConnection) GetKind() ResourceRefKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *AiApplicationResourceConnection) GetKindOk() (*ResourceRefKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *AiApplicationResourceConnection) SetKind(v ResourceRefKind)`

SetKind sets Kind field to given value.


### GetWorkloadId

`func (o *AiApplicationResourceConnection) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *AiApplicationResourceConnection) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *AiApplicationResourceConnection) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.

### HasWorkloadId

`func (o *AiApplicationResourceConnection) HasWorkloadId() bool`

HasWorkloadId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


