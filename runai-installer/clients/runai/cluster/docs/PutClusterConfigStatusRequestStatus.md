# PutClusterConfigStatusRequestStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActualConfig** | Pointer to [**ClusterFullConfigData**](ClusterFullConfigData.md) |  | [optional] 
**Conditions** | [**[]Condition**](Condition.md) |  | 

## Methods

### NewPutClusterConfigStatusRequestStatus

`func NewPutClusterConfigStatusRequestStatus(conditions []Condition, ) *PutClusterConfigStatusRequestStatus`

NewPutClusterConfigStatusRequestStatus instantiates a new PutClusterConfigStatusRequestStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPutClusterConfigStatusRequestStatusWithDefaults

`func NewPutClusterConfigStatusRequestStatusWithDefaults() *PutClusterConfigStatusRequestStatus`

NewPutClusterConfigStatusRequestStatusWithDefaults instantiates a new PutClusterConfigStatusRequestStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActualConfig

`func (o *PutClusterConfigStatusRequestStatus) GetActualConfig() ClusterFullConfigData`

GetActualConfig returns the ActualConfig field if non-nil, zero value otherwise.

### GetActualConfigOk

`func (o *PutClusterConfigStatusRequestStatus) GetActualConfigOk() (*ClusterFullConfigData, bool)`

GetActualConfigOk returns a tuple with the ActualConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualConfig

`func (o *PutClusterConfigStatusRequestStatus) SetActualConfig(v ClusterFullConfigData)`

SetActualConfig sets ActualConfig field to given value.

### HasActualConfig

`func (o *PutClusterConfigStatusRequestStatus) HasActualConfig() bool`

HasActualConfig returns a boolean if a field has been set.

### GetConditions

`func (o *PutClusterConfigStatusRequestStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *PutClusterConfigStatusRequestStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *PutClusterConfigStatusRequestStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


