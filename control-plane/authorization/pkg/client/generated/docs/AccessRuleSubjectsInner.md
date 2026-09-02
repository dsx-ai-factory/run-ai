# AccessRuleSubjectsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SubjectId** | **string** | The unique identifier of the subject e.g., user, group, or app). | 
**SubjectType** | [**SubjectType**](SubjectType.md) |  | 

## Methods

### NewAccessRuleSubjectsInner

`func NewAccessRuleSubjectsInner(subjectId string, subjectType SubjectType, ) *AccessRuleSubjectsInner`

NewAccessRuleSubjectsInner instantiates a new AccessRuleSubjectsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleSubjectsInnerWithDefaults

`func NewAccessRuleSubjectsInnerWithDefaults() *AccessRuleSubjectsInner`

NewAccessRuleSubjectsInnerWithDefaults instantiates a new AccessRuleSubjectsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSubjectId

`func (o *AccessRuleSubjectsInner) GetSubjectId() string`

GetSubjectId returns the SubjectId field if non-nil, zero value otherwise.

### GetSubjectIdOk

`func (o *AccessRuleSubjectsInner) GetSubjectIdOk() (*string, bool)`

GetSubjectIdOk returns a tuple with the SubjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectId

`func (o *AccessRuleSubjectsInner) SetSubjectId(v string)`

SetSubjectId sets SubjectId field to given value.


### GetSubjectType

`func (o *AccessRuleSubjectsInner) GetSubjectType() SubjectType`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *AccessRuleSubjectsInner) GetSubjectTypeOk() (*SubjectType, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *AccessRuleSubjectsInner) SetSubjectType(v SubjectType)`

SetSubjectType sets SubjectType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


