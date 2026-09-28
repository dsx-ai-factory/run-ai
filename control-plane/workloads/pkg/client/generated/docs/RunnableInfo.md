# RunnableInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsRunnable** | Pointer to **bool** | indication if the template is runnable, as is | [optional] 
**Issue** | Pointer to **NullableString** | if the template is non runnable, specify a reason | [optional] 

## Methods

### NewRunnableInfo

`func NewRunnableInfo() *RunnableInfo`

NewRunnableInfo instantiates a new RunnableInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunnableInfoWithDefaults

`func NewRunnableInfoWithDefaults() *RunnableInfo`

NewRunnableInfoWithDefaults instantiates a new RunnableInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsRunnable

`func (o *RunnableInfo) GetIsRunnable() bool`

GetIsRunnable returns the IsRunnable field if non-nil, zero value otherwise.

### GetIsRunnableOk

`func (o *RunnableInfo) GetIsRunnableOk() (*bool, bool)`

GetIsRunnableOk returns a tuple with the IsRunnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsRunnable

`func (o *RunnableInfo) SetIsRunnable(v bool)`

SetIsRunnable sets IsRunnable field to given value.

### HasIsRunnable

`func (o *RunnableInfo) HasIsRunnable() bool`

HasIsRunnable returns a boolean if a field has been set.

### GetIssue

`func (o *RunnableInfo) GetIssue() string`

GetIssue returns the Issue field if non-nil, zero value otherwise.

### GetIssueOk

`func (o *RunnableInfo) GetIssueOk() (*string, bool)`

GetIssueOk returns a tuple with the Issue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssue

`func (o *RunnableInfo) SetIssue(v string)`

SetIssue sets Issue field to given value.

### HasIssue

`func (o *RunnableInfo) HasIssue() bool`

HasIssue returns a boolean if a field has been set.

### SetIssueNil

`func (o *RunnableInfo) SetIssueNil(b bool)`

 SetIssueNil sets the value for Issue to be an explicit nil

### UnsetIssue
`func (o *RunnableInfo) UnsetIssue()`

UnsetIssue ensures that no value is present for Issue, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


