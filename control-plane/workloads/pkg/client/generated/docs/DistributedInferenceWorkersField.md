# DistributedInferenceWorkersField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Workers** | Pointer to **NullableInt32** | Specifies the number of worker nodes to run. If set to 0, only the leader node will run, and no worker pods will be created. In this case, worker spec is not required. | [optional] [default to 0]

## Methods

### NewDistributedInferenceWorkersField

`func NewDistributedInferenceWorkersField() *DistributedInferenceWorkersField`

NewDistributedInferenceWorkersField instantiates a new DistributedInferenceWorkersField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceWorkersFieldWithDefaults

`func NewDistributedInferenceWorkersFieldWithDefaults() *DistributedInferenceWorkersField`

NewDistributedInferenceWorkersFieldWithDefaults instantiates a new DistributedInferenceWorkersField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkers

`func (o *DistributedInferenceWorkersField) GetWorkers() int32`

GetWorkers returns the Workers field if non-nil, zero value otherwise.

### GetWorkersOk

`func (o *DistributedInferenceWorkersField) GetWorkersOk() (*int32, bool)`

GetWorkersOk returns a tuple with the Workers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkers

`func (o *DistributedInferenceWorkersField) SetWorkers(v int32)`

SetWorkers sets Workers field to given value.

### HasWorkers

`func (o *DistributedInferenceWorkersField) HasWorkers() bool`

HasWorkers returns a boolean if a field has been set.

### SetWorkersNil

`func (o *DistributedInferenceWorkersField) SetWorkersNil(b bool)`

 SetWorkersNil sets the value for Workers to be an explicit nil

### UnsetWorkers
`func (o *DistributedInferenceWorkersField) UnsetWorkers()`

UnsetWorkers ensures that no value is present for Workers, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


