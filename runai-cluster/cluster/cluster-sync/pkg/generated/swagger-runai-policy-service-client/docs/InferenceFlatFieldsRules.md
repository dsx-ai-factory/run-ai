# InferenceFlatFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Image** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicyRules**](ImagePullPolicyRules.md) |  | [optional] 
**WorkingDir** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**CreateHomeDir** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**Probes** | Pointer to [**NullableProbesRules**](ProbesRules.md) |  | [optional] 

## Methods

### NewInferenceFlatFieldsRules

`func NewInferenceFlatFieldsRules() *InferenceFlatFieldsRules`

NewInferenceFlatFieldsRules instantiates a new InferenceFlatFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceFlatFieldsRulesWithDefaults

`func NewInferenceFlatFieldsRulesWithDefaults() *InferenceFlatFieldsRules`

NewInferenceFlatFieldsRulesWithDefaults instantiates a new InferenceFlatFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetImage

`func (o *InferenceFlatFieldsRules) GetImage() StringRules`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *InferenceFlatFieldsRules) GetImageOk() (*StringRules, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *InferenceFlatFieldsRules) SetImage(v StringRules)`

SetImage sets Image field to given value.

### HasImage

`func (o *InferenceFlatFieldsRules) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *InferenceFlatFieldsRules) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *InferenceFlatFieldsRules) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *InferenceFlatFieldsRules) GetImagePullPolicy() ImagePullPolicyRules`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *InferenceFlatFieldsRules) GetImagePullPolicyOk() (*ImagePullPolicyRules, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *InferenceFlatFieldsRules) SetImagePullPolicy(v ImagePullPolicyRules)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *InferenceFlatFieldsRules) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *InferenceFlatFieldsRules) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *InferenceFlatFieldsRules) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *InferenceFlatFieldsRules) GetWorkingDir() StringRules`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *InferenceFlatFieldsRules) GetWorkingDirOk() (*StringRules, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *InferenceFlatFieldsRules) SetWorkingDir(v StringRules)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *InferenceFlatFieldsRules) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *InferenceFlatFieldsRules) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *InferenceFlatFieldsRules) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *InferenceFlatFieldsRules) GetCreateHomeDir() BooleanRules`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *InferenceFlatFieldsRules) GetCreateHomeDirOk() (*BooleanRules, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *InferenceFlatFieldsRules) SetCreateHomeDir(v BooleanRules)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *InferenceFlatFieldsRules) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *InferenceFlatFieldsRules) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *InferenceFlatFieldsRules) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *InferenceFlatFieldsRules) GetProbes() ProbesRules`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *InferenceFlatFieldsRules) GetProbesOk() (*ProbesRules, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *InferenceFlatFieldsRules) SetProbes(v ProbesRules)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *InferenceFlatFieldsRules) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *InferenceFlatFieldsRules) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *InferenceFlatFieldsRules) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


