# openapi_client.AdminApi

All URIs are relative to *https://game.bahms.org/api/poker*

Method | HTTP request | Description
------------- | ------------- | -------------
[**admin_game_game_id_status_post**](AdminApi.md#admin_game_game_id_status_post) | **POST** /admin/game/{game_id}/status | Update Game Status
[**admin_game_game_id_table_post**](AdminApi.md#admin_game_game_id_table_post) | **POST** /admin/game/{game_id}/table | Change Game Table


# **admin_game_game_id_status_post**
> object admin_game_game_id_status_post(game_id, admin_game_game_id_status_post_request)

Update Game Status

Changes the status of an active game, for use by admins

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.admin_game_game_id_status_post_request import AdminGameGameIdStatusPostRequest
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
    api_instance = openapi_client.AdminApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game to update status of
    admin_game_game_id_status_post_request = openapi_client.AdminGameGameIdStatusPostRequest() # AdminGameGameIdStatusPostRequest | the status of the game to set

    try:
        # Update Game Status
        api_response = await api_instance.admin_game_game_id_status_post(game_id, admin_game_game_id_status_post_request)
        print("The response of AdminApi->admin_game_game_id_status_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->admin_game_game_id_status_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game to update status of | 
 **admin_game_game_id_status_post_request** | [**AdminGameGameIdStatusPostRequest**](AdminGameGameIdStatusPostRequest.md)| the status of the game to set | 

### Return type

**object**

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **admin_game_game_id_table_post**
> object admin_game_game_id_table_post(game_id, admin_game_game_id_table_post_request)

Change Game Table

Changes the state of an active game

### Example

* Bearer Authentication (BearerAuth):

```python
import openapi_client
from openapi_client.models.admin_game_game_id_table_post_request import AdminGameGameIdTablePostRequest
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
    api_instance = openapi_client.AdminApi(api_client)
    game_id = 'game_id_example' # str | ID of the Game to update status of
    admin_game_game_id_table_post_request = openapi_client.AdminGameGameIdTablePostRequest() # AdminGameGameIdTablePostRequest | the table struct to apply

    try:
        # Change Game Table
        api_response = await api_instance.admin_game_game_id_table_post(game_id, admin_game_game_id_table_post_request)
        print("The response of AdminApi->admin_game_game_id_table_post:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->admin_game_game_id_table_post: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **game_id** | **str**| ID of the Game to update status of | 
 **admin_game_game_id_table_post_request** | [**AdminGameGameIdTablePostRequest**](AdminGameGameIdTablePostRequest.md)| the table struct to apply | 

### Return type

**object**

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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

