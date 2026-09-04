# DeploymentsSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectName** | **string** | The project name we are syncing objects in | 
**Deployments** | [**[]Deployment**](Deployment.md) |  | 

## Methods

### NewDeploymentsSync

`func NewDeploymentsSync(projectName string, deployments []Deployment, ) *DeploymentsSync`

NewDeploymentsSync instantiates a new DeploymentsSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentsSyncWithDefaults

`func NewDeploymentsSyncWithDefaults() *DeploymentsSync`

NewDeploymentsSyncWithDefaults instantiates a new DeploymentsSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectName

`func (o *DeploymentsSync) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *DeploymentsSync) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *DeploymentsSync) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetDeployments

`func (o *DeploymentsSync) GetDeployments() []Deployment`

GetDeployments returns the Deployments field if non-nil, zero value otherwise.

### GetDeploymentsOk

`func (o *DeploymentsSync) GetDeploymentsOk() (*[]Deployment, bool)`

GetDeploymentsOk returns a tuple with the Deployments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeployments

`func (o *DeploymentsSync) SetDeployments(v []Deployment)`

SetDeployments sets Deployments field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


