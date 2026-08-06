# GameGameStatusDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**status** | [**GameGameStatus**](GameGameStatus.md) |  | [optional] 

## Example

```python
from openapi_client.models.game_game_status_dto import GameGameStatusDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameGameStatusDTO from a JSON string
game_game_status_dto_instance = GameGameStatusDTO.from_json(json)
# print the JSON string representation of the object
print(GameGameStatusDTO.to_json())

# convert the object into a dict
game_game_status_dto_dict = game_game_status_dto_instance.to_dict()
# create an instance of GameGameStatusDTO from a dict
game_game_status_dto_from_dict = GameGameStatusDTO.from_dict(game_game_status_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


