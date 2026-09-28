# AIApplicationCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**ProjectId** | **string** | The id of the project. | 
**Name** | **string** | The name of the application. The name must be unique within the organization and can only contain lowercase alphanumeric characters and hyphens. It must start with a letter. | 
**Spec** | [**AIApplicationSpec**](AIApplicationSpec.md) |  | 

## Methods

### NewAIApplicationCreationRequest

`func NewAIApplicationCreationRequest(clusterId string, projectId string, name string, spec AIApplicationSpec, ) *AIApplicationCreationRequest`

NewAIApplicationCreationRequest instantiates a new AIApplicationCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationCreationRequestWithDefaults

`func NewAIApplicationCreationRequestWithDefaults() *AIApplicationCreationRequest`

NewAIApplicationCreationRequestWithDefaults instantiates a new AIApplicationCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *AIApplicationCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AIApplicationCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AIApplicationCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *AIApplicationCreationRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AIApplicationCreationRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AIApplicationCreationRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetName

`func (o *AIApplicationCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AIApplicationCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AIApplicationCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetSpec

`func (o *AIApplicationCreationRequest) GetSpec() AIApplicationSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *AIApplicationCreationRequest) GetSpecOk() (*AIApplicationSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *AIApplicationCreationRequest) SetSpec(v AIApplicationSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


