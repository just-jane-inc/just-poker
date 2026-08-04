# GameHandDTO

the current hand

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**big_blind** | **int** | the amount of chips for the big blind in this hand | [optional] 
**id** | **int** | the (non decreasing) hand counter | [optional] 
**small_blind** | **int** | the amount of chips for the small blind in this hand | [optional] 
**started_at** | **str** | the time that this hand started | [optional] 

## Example

```python
from openapi_client.models.game_hand_dto import GameHandDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameHandDTO from a JSON string
game_hand_dto_instance = GameHandDTO.from_json(json)
# print the JSON string representation of the object
print(GameHandDTO.to_json())

# convert the object into a dict
game_hand_dto_dict = game_hand_dto_instance.to_dict()
# create an instance of GameHandDTO from a dict
game_hand_dto_from_dict = GameHandDTO.from_dict(game_hand_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


