# ExternalAccessInference

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Epp** | **map[string]interface{}** | InferenceEndpoint spec.epp, for example the EPP profile. Validated against the installed CRD when applied. | 

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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


