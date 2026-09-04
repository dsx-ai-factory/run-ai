# \TrialsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTrial**](TrialsApi.md#CreateTrial) | **Post** /api/v1/trials | Create new trial
[**GetTrialState**](TrialsApi.md#GetTrialState) | **Get** /api/v1/trials/state | get trials state
[**ResendEmail**](TrialsApi.md#ResendEmail) | **Post** /api/v1/trials/{trialId}/resend | Resend a trial email



## CreateTrial

> TrialCreationResponse CreateTrial(ctx).TrialCreationRequest(trialCreationRequest).Execute()

Create new trial

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    trialCreationRequest := *openapiclient.NewTrialCreationRequest("FirstName_example", "LastName_example", "Company_example", "Email_example", false) // TrialCreationRequest | The trial to create.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TrialsApi.CreateTrial(context.Background()).TrialCreationRequest(trialCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TrialsApi.CreateTrial``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateTrial`: TrialCreationResponse
    fmt.Fprintf(os.Stdout, "Response from `TrialsApi.CreateTrial`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTrialRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trialCreationRequest** | [**TrialCreationRequest**](TrialCreationRequest.md) | The trial to create. | 

### Return type

[**TrialCreationResponse**](TrialCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTrialState

> []TrialState GetTrialState(ctx).Execute()

get trials state

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TrialsApi.GetTrialState(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TrialsApi.GetTrialState``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTrialState`: []TrialState
    fmt.Fprintf(os.Stdout, "Response from `TrialsApi.GetTrialState`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetTrialStateRequest struct via the builder pattern


### Return type

[**[]TrialState**](TrialState.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ResendEmail

> ResendEmail(ctx, trialId).Execute()

Resend a trial email

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    trialId := "123e4567-e89b-12d3-a456-426614174000" // string | The id of the trial

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TrialsApi.ResendEmail(context.Background(), trialId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TrialsApi.ResendEmail``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**trialId** | **string** | The id of the trial | 

### Other Parameters

Other parameters are passed through a pointer to a apiResendEmailRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

