# DistributedFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumWorkers** | Pointer to **NullableInt32** | the number of workers that will be allocated for running the workload. | [optional] 
**DistributedFramework** | Pointer to [**NullableDistributedFramework**](DistributedFramework.md) |  | [optional] 
**SlotsPerWorker** | Pointer to **NullableInt32** | Specifies the number of slots per worker used in hostfile. Defaults to 1. (applicable only for MPI) | [optional] [default to 1]
**MinReplicas** | Pointer to **NullableInt32** | the lower limit for the number of worker pods to which the training job can scale down. (applicable only for PyTorch) | [optional] 
**MaxReplicas** | Pointer to **NullableInt32** | the upper limit for the number of worker pods that can be set by the autoscaler. Cannot be smaller than MinReplicas. (applicable only for PyTorch) | [optional] 

## Methods

### NewDistributedFields

`func NewDistributedFields() *DistributedFields`

NewDistributedFields instantiates a new DistributedFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedFieldsWithDefaults

`func NewDistributedFieldsWithDefaults() *DistributedFields`

NewDistributedFieldsWithDefaults instantiates a new DistributedFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumWorkers

`func (o *DistributedFields) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *DistributedFields) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *DistributedFields) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *DistributedFields) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### SetNumWorkersNil

`func (o *DistributedFields) SetNumWorkersNil(b bool)`

 SetNumWorkersNil sets the value for NumWorkers to be an explicit nil

### UnsetNumWorkers
`func (o *DistributedFields) UnsetNumWorkers()`

UnsetNumWorkers ensures that no value is present for NumWorkers, not even an explicit nil
### GetDistributedFramework

`func (o *DistributedFields) GetDistributedFramework() DistributedFramework`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *DistributedFields) GetDistributedFrameworkOk() (*DistributedFramework, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *DistributedFields) SetDistributedFramework(v DistributedFramework)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *DistributedFields) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### SetDistributedFrameworkNil

`func (o *DistributedFields) SetDistributedFrameworkNil(b bool)`

 SetDistributedFrameworkNil sets the value for DistributedFramework to be an explicit nil

### UnsetDistributedFramework
`func (o *DistributedFields) UnsetDistributedFramework()`

UnsetDistributedFramework ensures that no value is present for DistributedFramework, not even an explicit nil
### GetSlotsPerWorker

`func (o *DistributedFields) GetSlotsPerWorker() int32`

GetSlotsPerWorker returns the SlotsPerWorker field if non-nil, zero value otherwise.

### GetSlotsPerWorkerOk

`func (o *DistributedFields) GetSlotsPerWorkerOk() (*int32, bool)`

GetSlotsPerWorkerOk returns a tuple with the SlotsPerWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSlotsPerWorker

`func (o *DistributedFields) SetSlotsPerWorker(v int32)`

SetSlotsPerWorker sets SlotsPerWorker field to given value.

### HasSlotsPerWorker

`func (o *DistributedFields) HasSlotsPerWorker() bool`

HasSlotsPerWorker returns a boolean if a field has been set.

### SetSlotsPerWorkerNil

`func (o *DistributedFields) SetSlotsPerWorkerNil(b bool)`

 SetSlotsPerWorkerNil sets the value for SlotsPerWorker to be an explicit nil

### UnsetSlotsPerWorker
`func (o *DistributedFields) UnsetSlotsPerWorker()`

UnsetSlotsPerWorker ensures that no value is present for SlotsPerWorker, not even an explicit nil
### GetMinReplicas

`func (o *DistributedFields) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *DistributedFields) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *DistributedFields) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *DistributedFields) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *DistributedFields) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *DistributedFields) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMaxReplicas

`func (o *DistributedFields) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *DistributedFields) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *DistributedFields) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *DistributedFields) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *DistributedFields) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *DistributedFields) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


