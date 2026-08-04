# GameActiveGameDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **str** |  | [optional] 
**player_ids** | **List[str]** |  | [optional] 

## Example

```python
from openapi_client.models.game_active_game_dto import GameActiveGameDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameActiveGameDTO from a JSON string
game_active_game_dto_instance = GameActiveGameDTO.from_json(json)
# print the JSON string representation of the object
print(GameActiveGameDTO.to_json())

# convert the object into a dict
game_active_game_dto_dict = game_active_game_dto_instance.to_dict()
# create an instance of GameActiveGameDTO from a dict
game_active_game_dto_from_dict = GameActiveGameDTO.from_dict(game_active_game_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


