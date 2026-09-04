# JobsNodeAffinityTrain

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AffinityType** | Pointer to **string** | The type of affinity of the jobs on the nodes. | [optional] 
**SelectedTypes** | Pointer to [**[]NodeTypeForProject**](NodeTypeForProject.md) |  | [optional] 

## Methods

### NewJobsNodeAffinityTrain

`func NewJobsNodeAffinityTrain() *JobsNodeAffinityTrain`

NewJobsNodeAffinityTrain instantiates a new JobsNodeAffinityTrain object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobsNodeAffinityTrainWithDefaults

`func NewJobsNodeAffinityTrainWithDefaults() *JobsNodeAffinityTrain`

NewJobsNodeAffinityTrainWithDefaults instantiates a new JobsNodeAffinityTrain object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAffinityType

`func (o *JobsNodeAffinityTrain) GetAffinityType() string`

GetAffinityType returns the AffinityType field if non-nil, zero value otherwise.

### GetAffinityTypeOk

`func (o *JobsNodeAffinityTrain) GetAffinityTypeOk() (*string, bool)`

GetAffinityTypeOk returns a tuple with the AffinityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffinityType

`func (o *JobsNodeAffinityTrain) SetAffinityType(v string)`

SetAffinityType sets AffinityType field to given value.

### HasAffinityType

`func (o *JobsNodeAffinityTrain) HasAffinityType() bool`

HasAffinityType returns a boolean if a field has been set.

### GetSelectedTypes

`func (o *JobsNodeAffinityTrain) GetSelectedTypes() []NodeTypeForProject`

GetSelectedTypes returns the SelectedTypes field if non-nil, zero value otherwise.

### GetSelectedTypesOk

`func (o *JobsNodeAffinityTrain) GetSelectedTypesOk() (*[]NodeTypeForProject, bool)`

GetSelectedTypesOk returns a tuple with the SelectedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectedTypes

`func (o *JobsNodeAffinityTrain) SetSelectedTypes(v []NodeTypeForProject)`

SetSelectedTypes sets SelectedTypes field to given value.

### HasSelectedTypes

`func (o *JobsNodeAffinityTrain) HasSelectedTypes() bool`

HasSelectedTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


