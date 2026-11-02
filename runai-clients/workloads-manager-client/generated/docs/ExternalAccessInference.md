# ExternalAccessInference

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Epp** | **map[string]interface{}** | InferenceEndpoint spec.epp, for example the EPP profile. Validated against the installed CRD when applied. | 
**Objectives** | Pointer to **[]map[string]interface{}** | InferenceEndpoint spec.objectives. Omitted when there are none. | [optional] 
**Rewrites** | Pointer to **[]map[string]interface{}** | InferenceEndpoint spec.rewrites. Omitted when there are none. | [optional] 

## Methods

### NewExternalAccessInference

`func NewExternalAccessInference(epp map[string]interface{}, ) *ExternalAccessInference`

NewExternalAccessInference instantiates a new ExternalAccessInference object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExternalAccessInferenceWithDefaults

`func NewExternalAccessInferenceWithDefaults() *ExternalAccessInference`

NewExternalAccessInferenceWithDefaults instantiates a new ExternalAccessInference object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEpp

`func (o *ExternalAccessInference) GetEpp() map[string]interface{}`

GetEpp returns the Epp field if non-nil, zero value otherwise.

### GetEppOk

`func (o *ExternalAccessInference) GetEppOk() (*map[string]interface{}, bool)`

GetEppOk returns a tuple with the Epp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEpp

`func (o *ExternalAccessInference) SetEpp(v map[string]interface{})`

SetEpp sets Epp field to given value.


### GetObjectives

`func (o *ExternalAccessInference) GetObjectives() []map[string]interface{}`

GetObjectives returns the Objectives field if non-nil, zero value otherwise.

### GetObjectivesOk

`func (o *ExternalAccessInference) GetObjectivesOk() (*[]map[string]interface{}, bool)`

GetObjectivesOk returns a tuple with the Objectives field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectives

`func (o *ExternalAccessInference) SetObjectives(v []map[string]interface{})`

SetObjectives sets Objectives field to given value.

### HasObjectives

`func (o *ExternalAccessInference) HasObjectives() bool`

HasObjectives returns a boolean if a field has been set.

### SetObjectivesNil

`func (o *ExternalAccessInference) SetObjectivesNil(b bool)`

 SetObjectivesNil sets the value for Objectives to be an explicit nil

### UnsetObjectives
`func (o *ExternalAccessInference) UnsetObjectives()`

UnsetObjectives ensures that no value is present for Objectives, not even an explicit nil
### GetRewrites

`func (o *ExternalAccessInference) GetRewrites() []map[string]interface{}`

GetRewrites returns the Rewrites field if non-nil, zero value otherwise.

### GetRewritesOk

`func (o *ExternalAccessInference) GetRewritesOk() (*[]map[string]interface{}, bool)`

GetRewritesOk returns a tuple with the Rewrites field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRewrites

`func (o *ExternalAccessInference) SetRewrites(v []map[string]interface{})`

SetRewrites sets Rewrites field to given value.

### HasRewrites

`func (o *ExternalAccessInference) HasRewrites() bool`

HasRewrites returns a boolean if a field has been set.

### SetRewritesNil

`func (o *ExternalAccessInference) SetRewritesNil(b bool)`

 SetRewritesNil sets the value for Rewrites to be an explicit nil

### UnsetRewrites
`func (o *ExternalAccessInference) UnsetRewrites()`

UnsetRewrites ensures that no value is present for Rewrites, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


