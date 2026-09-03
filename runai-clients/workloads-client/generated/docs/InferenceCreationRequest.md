# InferenceCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**UseGivenNameAsPrefix** | Pointer to **bool** | When true, the requested name will be treated as a prefix. The final name of the workload will be composed of the name followed by a random set of characters. | [optional] [default to false]
**ProjectId** | **string** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 
**TemplateId** | Pointer to **NullableString** | The unique identifier of the template to use for submitting this workload. The combined values provided in the spec, template and assets will be used to create the workload. | [optional] 
**Assets** | Pointer to [**NullableAssetSubmissionIds**](AssetSubmissionIds.md) |  | [optional] 
**Spec** | Pointer to [**InferenceSpecSpec**](InferenceSpecSpec.md) |  | [optional] 

## Methods

### NewInferenceCreationRequest

`func NewInferenceCreationRequest(name string, projectId string, clusterId string, ) *InferenceCreationRequest`

NewInferenceCreationRequest instantiates a new InferenceCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceCreationRequestWithDefaults

`func NewInferenceCreationRequestWithDefaults() *InferenceCreationRequest`

NewInferenceCreationRequestWithDefaults instantiates a new InferenceCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *InferenceCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InferenceCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InferenceCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUseGivenNameAsPrefix

`func (o *InferenceCreationRequest) GetUseGivenNameAsPrefix() bool`

GetUseGivenNameAsPrefix returns the UseGivenNameAsPrefix field if non-nil, zero value otherwise.

### GetUseGivenNameAsPrefixOk

`func (o *InferenceCreationRequest) GetUseGivenNameAsPrefixOk() (*bool, bool)`

GetUseGivenNameAsPrefixOk returns a tuple with the UseGivenNameAsPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseGivenNameAsPrefix

`func (o *InferenceCreationRequest) SetUseGivenNameAsPrefix(v bool)`

SetUseGivenNameAsPrefix sets UseGivenNameAsPrefix field to given value.

### HasUseGivenNameAsPrefix

`func (o *InferenceCreationRequest) HasUseGivenNameAsPrefix() bool`

HasUseGivenNameAsPrefix returns a boolean if a field has been set.

### GetProjectId

`func (o *InferenceCreationRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *InferenceCreationRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *InferenceCreationRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *InferenceCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *InferenceCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *InferenceCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetTemplateId

`func (o *InferenceCreationRequest) GetTemplateId() string`

GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.

### GetTemplateIdOk

`func (o *InferenceCreationRequest) GetTemplateIdOk() (*string, bool)`

GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplateId

`func (o *InferenceCreationRequest) SetTemplateId(v string)`

SetTemplateId sets TemplateId field to given value.

### HasTemplateId

`func (o *InferenceCreationRequest) HasTemplateId() bool`

HasTemplateId returns a boolean if a field has been set.

### SetTemplateIdNil

`func (o *InferenceCreationRequest) SetTemplateIdNil(b bool)`

 SetTemplateIdNil sets the value for TemplateId to be an explicit nil

### UnsetTemplateId
`func (o *InferenceCreationRequest) UnsetTemplateId()`

UnsetTemplateId ensures that no value is present for TemplateId, not even an explicit nil
### GetAssets

`func (o *InferenceCreationRequest) GetAssets() AssetSubmissionIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceCreationRequest) GetAssetsOk() (*AssetSubmissionIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceCreationRequest) SetAssets(v AssetSubmissionIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *InferenceCreationRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### SetAssetsNil

`func (o *InferenceCreationRequest) SetAssetsNil(b bool)`

 SetAssetsNil sets the value for Assets to be an explicit nil

### UnsetAssets
`func (o *InferenceCreationRequest) UnsetAssets()`

UnsetAssets ensures that no value is present for Assets, not even an explicit nil
### GetSpec

`func (o *InferenceCreationRequest) GetSpec() InferenceSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InferenceCreationRequest) GetSpecOk() (*InferenceSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InferenceCreationRequest) SetSpec(v InferenceSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InferenceCreationRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


