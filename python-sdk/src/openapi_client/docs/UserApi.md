# openapi_client.UserApi

All URIs are relative to *https://game.bahms.org/api/poker*

Method | HTTP request | Description
------------- | ------------- | -------------
[**user_me_delete**](UserApi.md#user_me_delete) | **DELETE** /user/me | delete requesting user


# **user_me_delete**
> JustResponseMessageAny user_me_delete(body=body)

delete requesting user

deletes the user associated to the token used to authenticate this request

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.just_response_message_any import JustResponseMessageAny
from openapi_client.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://game.bahms.org/api/poker
# See configuration.py for a list of all supported configuration parameters.
configuration = openapi_client.Configuration(
    host = "https://game.bahms.org/api/poker"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure Bearer authorization: BearerAuth
configuration = openapi_client.Configuration(
    access_token = os.environ["BEARER_TOKEN"]
)

# Enter a context with an instance of the API client
async with openapi_client.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = openapi_client.UserApi(api_client)
    body = None # object |  (optional)

    try:
        # delete requesting user
        api_response = await api_instance.user_me_delete(body=body)
        print("The response of UserApi->user_me_delete:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling UserApi->user_me_delete: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | **object**|  | [optional] 

### Return type

[**JustResponseMessageAny**](JustResponseMessageAny.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | OK |  -  |
**400** | Bad Request |  -  |
**401** | Unauthorized |  -  |
**403** | Forbidden |  -  |
**404** | Not Found |  -  |
**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

