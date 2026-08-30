# AccessKeyCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the access key. The name must be unique within the user&#39;s access keys list and can only contain lowercase alphanumeric characters and hyphens. It must start and end with a letter. | 

## Methods

### NewAccessKeyCreationRequest

`func NewAccessKeyCreationRequest(name string, ) *AccessKeyCreationRequest`

NewAccessKeyCreationRequest instantiates a new AccessKeyCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessKeyCreationRequestWithDefaults

`func NewAccessKeyCreationRequestWithDefaults() *AccessKeyCreationRequest`

NewAccessKeyCreationRequestWithDefaults instantiates a new AccessKeyCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AccessKeyCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AccessKeyCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AccessKeyCreationRequest) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


