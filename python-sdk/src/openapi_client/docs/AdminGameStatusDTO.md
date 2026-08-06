# AdminGameStatusDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**status** | [**AdminGameStatus**](AdminGameStatus.md) |  | [optional] 

## Example

```python
from openapi_client.models.admin_game_status_dto import AdminGameStatusDTO

# TODO update the JSON string below
json = "{}"
# create an instance of AdminGameStatusDTO from a JSON string
admin_game_status_dto_instance = AdminGameStatusDTO.from_json(json)
# print the JSON string representation of the object
print(AdminGameStatusDTO.to_json())

# convert the object into a dict
admin_game_status_dto_dict = admin_game_status_dto_instance.to_dict()
# create an instance of AdminGameStatusDTO from a dict
admin_game_status_dto_from_dict = AdminGameStatusDTO.from_dict(admin_game_status_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


