# GameRoundStartEventDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**round_type** | [**GameRoundType**](GameRoundType.md) |  | [optional] 

## Example

```python
from openapi_client.models.game_round_start_event_dto import GameRoundStartEventDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameRoundStartEventDTO from a JSON string
game_round_start_event_dto_instance = GameRoundStartEventDTO.from_json(json)
# print the JSON string representation of the object
print(GameRoundStartEventDTO.to_json())

# convert the object into a dict
game_round_start_event_dto_dict = game_round_start_event_dto_instance.to_dict()
# create an instance of GameRoundStartEventDTO from a dict
game_round_start_event_dto_from_dict = GameRoundStartEventDTO.from_dict(game_round_start_event_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


