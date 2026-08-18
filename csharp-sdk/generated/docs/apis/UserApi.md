# JustPoker.OpenApi.Api.UserApi

All URIs are relative to *https://game.bahms.org/api/poker*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**UserMeDelete**](UserApi.md#usermedelete) | **DELETE** /user/me | delete requesting user |

<a id="usermedelete"></a>
# **UserMeDelete**
> JustResponseMessageAny UserMeDelete (Object body = null)

delete requesting user

deletes the user associated to the token used to authenticate this request


### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **body** | **Object** |  | [optional]  |

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
| **200** | OK |  -  |
| **400** | Bad Request |  -  |
| **401** | Unauthorized |  -  |
| **403** | Forbidden |  -  |
| **404** | Not Found |  -  |
| **500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to Model list]](../../README.md#documentation-for-models) [[Back to README]](../../README.md)

