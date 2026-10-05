# AgentInferenceEndpoint

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SchemaVersion** | **string** | Version of the transport envelope, currently v1. Consumers reject unsupported versions. | 
**Spec** | **map[string]interface{}** | Resolved InferenceEndpoint spec, validated against the installed CRD when applied. Does not include resource identity, ownership, or status. | 

## Methods

### NewAgentInferenceEndpoint

`func NewAgentInferenceEndpoint(schemaVersion string, spec map[string]interface{}, ) *AgentInferenceEndpoint`

NewAgentInferenceEndpoint instantiates a new AgentInferenceEndpoint object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAgentInferenceEndpointWithDefaults

`func NewAgentInferenceEndpointWithDefaults() *AgentInferenceEndpoint`

NewAgentInferenceEndpointWithDefaults instantiates a new AgentInferenceEndpoint object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSchemaVersion

`func (o *AgentInferenceEndpoint) GetSchemaVersion() string`

GetSchemaVersion returns the SchemaVersion field if non-nil, zero value otherwise.

### GetSchemaVersionOk

`func (o *AgentInferenceEndpoint) GetSchemaVersionOk() (*string, bool)`

GetSchemaVersionOk returns a tuple with the SchemaVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchemaVersion

`func (o *AgentInferenceEndpoint) SetSchemaVersion(v string)`

SetSchemaVersion sets SchemaVersion field to given value.


### GetSpec

`func (o *AgentInferenceEndpoint) GetSpec() map[string]interface{}`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *AgentInferenceEndpoint) GetSpecOk() (*map[string]interface{}, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *AgentInferenceEndpoint) SetSpec(v map[string]interface{})`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


