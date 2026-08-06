# AdminGameGameIdStatusPostRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**status** | [**AdminGameStatus**](AdminGameStatus.md) |  | [optional] 

## Example

```python
from openapi_client.models.admin_game_game_id_status_post_request import AdminGameGameIdStatusPostRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AdminGameGameIdStatusPostRequest from a JSON string
admin_game_game_id_status_post_request_instance = AdminGameGameIdStatusPostRequest.from_json(json)
# print the JSON string representation of the object
print(AdminGameGameIdStatusPostRequest.to_json())

# convert the object into a dict
admin_game_game_id_status_post_request_dict = admin_game_game_id_status_post_request_instance.to_dict()
# create an instance of AdminGameGameIdStatusPostRequest from a dict
admin_game_game_id_status_post_request_from_dict = AdminGameGameIdStatusPostRequest.from_dict(admin_game_game_id_status_post_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


