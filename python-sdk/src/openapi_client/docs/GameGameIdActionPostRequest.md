# GameGameIdActionPostRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**accepted_at** | **str** | a timestamp capturing when a succesful action was accepted by the game | [optional] 
**chips** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**intent** | [**GamePlayerIntent**](GamePlayerIntent.md) |  | [optional] 
**player_id** | **str** | the id of the player preforming the action | [optional] 

## Example

```python
from openapi_client.models.game_game_id_action_post_request import GameGameIdActionPostRequest

# TODO update the JSON string below
json = "{}"
# create an instance of GameGameIdActionPostRequest from a JSON string
game_game_id_action_post_request_instance = GameGameIdActionPostRequest.from_json(json)
# print the JSON string representation of the object
print(GameGameIdActionPostRequest.to_json())

# convert the object into a dict
game_game_id_action_post_request_dict = game_game_id_action_post_request_instance.to_dict()
# create an instance of GameGameIdActionPostRequest from a dict
game_game_id_action_post_request_from_dict = GameGameIdActionPostRequest.from_dict(game_game_id_action_post_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


