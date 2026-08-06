# GameGameDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ended_at** | **str** | the time that the game ended | [optional] 
**game_config** | [**GameNewGameConfigDTO**](GameNewGameConfigDTO.md) |  | [optional] 
**id** | **str** | the id of the game | [optional] 
**started_at** | **str** | the time that the game started originally | [optional] 
**table** | [**GameTableDTO**](GameTableDTO.md) |  | [optional] 

## Example

```python
from openapi_client.models.game_game_dto import GameGameDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameGameDTO from a JSON string
game_game_dto_instance = GameGameDTO.from_json(json)
# print the JSON string representation of the object
print(GameGameDTO.to_json())

# convert the object into a dict
game_game_dto_dict = game_game_dto_instance.to_dict()
# create an instance of GameGameDTO from a dict
game_game_dto_from_dict = GameGameDTO.from_dict(game_game_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


