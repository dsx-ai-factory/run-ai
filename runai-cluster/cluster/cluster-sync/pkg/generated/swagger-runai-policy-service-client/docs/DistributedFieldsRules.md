# DistributedFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumWorkers** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**DistributedFramework** | Pointer to [**NullableDistributedFrameworkRules**](DistributedFrameworkRules.md) |  | [optional] 
**SlotsPerWorker** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**MinReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**MaxReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 

## Methods

### NewDistributedFieldsRules

`func NewDistributedFieldsRules() *DistributedFieldsRules`

NewDistributedFieldsRules instantiates a new DistributedFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedFieldsRulesWithDefaults

`func NewDistributedFieldsRulesWithDefaults() *DistributedFieldsRules`

NewDistributedFieldsRulesWithDefaults instantiates a new DistributedFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumWorkers

`func (o *DistributedFieldsRules) GetNumWorkers() IntegerRules`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *DistributedFieldsRules) GetNumWorkersOk() (*IntegerRules, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *DistributedFieldsRules) SetNumWorkers(v IntegerRules)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *DistributedFieldsRules) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### SetNumWorkersNil

`func (o *DistributedFieldsRules) SetNumWorkersNil(b bool)`

 SetNumWorkersNil sets the value for NumWorkers to be an explicit nil

### UnsetNumWorkers
`func (o *DistributedFieldsRules) UnsetNumWorkers()`

UnsetNumWorkers ensures that no value is present for NumWorkers, not even an explicit nil
### GetDistributedFramework

`func (o *DistributedFieldsRules) GetDistributedFramework() DistributedFrameworkRules`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *DistributedFieldsRules) GetDistributedFrameworkOk() (*DistributedFrameworkRules, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *DistributedFieldsRules) SetDistributedFramework(v DistributedFrameworkRules)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *DistributedFieldsRules) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### SetDistributedFrameworkNil

`func (o *DistributedFieldsRules) SetDistributedFrameworkNil(b bool)`

 SetDistributedFrameworkNil sets the value for DistributedFramework to be an explicit nil

### UnsetDistributedFramework
`func (o *DistributedFieldsRules) UnsetDistributedFramework()`

UnsetDistributedFramework ensures that no value is present for DistributedFramework, not even an explicit nil
### GetSlotsPerWorker

`func (o *DistributedFieldsRules) GetSlotsPerWorker() IntegerRules`

GetSlotsPerWorker returns the SlotsPerWorker field if non-nil, zero value otherwise.

### GetSlotsPerWorkerOk

`func (o *DistributedFieldsRules) GetSlotsPerWorkerOk() (*IntegerRules, bool)`

GetSlotsPerWorkerOk returns a tuple with the SlotsPerWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSlotsPerWorker

`func (o *DistributedFieldsRules) SetSlotsPerWorker(v IntegerRules)`

SetSlotsPerWorker sets SlotsPerWorker field to given value.

### HasSlotsPerWorker

`func (o *DistributedFieldsRules) HasSlotsPerWorker() bool`

HasSlotsPerWorker returns a boolean if a field has been set.

### SetSlotsPerWorkerNil

`func (o *DistributedFieldsRules) SetSlotsPerWorkerNil(b bool)`

 SetSlotsPerWorkerNil sets the value for SlotsPerWorker to be an explicit nil

### UnsetSlotsPerWorker
`func (o *DistributedFieldsRules) UnsetSlotsPerWorker()`

UnsetSlotsPerWorker ensures that no value is present for SlotsPerWorker, not even an explicit nil
### GetMinReplicas

`func (o *DistributedFieldsRules) GetMinReplicas() IntegerRules`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *DistributedFieldsRules) GetMinReplicasOk() (*IntegerRules, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *DistributedFieldsRules) SetMinReplicas(v IntegerRules)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *DistributedFieldsRules) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *DistributedFieldsRules) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *DistributedFieldsRules) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMaxReplicas

`func (o *DistributedFieldsRules) GetMaxReplicas() IntegerRules`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *DistributedFieldsRules) GetMaxReplicasOk() (*IntegerRules, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *DistributedFieldsRules) SetMaxReplicas(v IntegerRules)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *DistributedFieldsRules) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *DistributedFieldsRules) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *DistributedFieldsRules) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


