# GameHandStartEventDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**big_blind_cost** | **int** |  | [optional] 
**big_blind_position** | **int** |  | [optional] 
**button_position** | **int** |  | [optional] 
**id** | **int** |  | [optional] 
**small_blind_cost** | **int** |  | [optional] 
**small_blind_position** | **int** |  | [optional] 

## Example

```python
from openapi_client.models.game_hand_start_event_dto import GameHandStartEventDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameHandStartEventDTO from a JSON string
game_hand_start_event_dto_instance = GameHandStartEventDTO.from_json(json)
# print the JSON string representation of the object
print(GameHandStartEventDTO.to_json())

# convert the object into a dict
game_hand_start_event_dto_dict = game_hand_start_event_dto_instance.to_dict()
# create an instance of GameHandStartEventDTO from a dict
game_hand_start_event_dto_from_dict = GameHandStartEventDTO.from_dict(game_hand_start_event_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


