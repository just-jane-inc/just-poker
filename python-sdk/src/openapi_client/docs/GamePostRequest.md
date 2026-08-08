# GamePostRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**auto_starts_hands** | **bool** | a flag which indicates true if the game server should wait for a signal to start hands or if it should do so automatically | [optional] 
**big_blind** | **int** | the big blind | [optional] 
**player_count** | **int** | the number of players (max) the game supports | [optional] 
**small_blind** | **int** | the small blind | [optional] 
**starting_chips** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 

## Example

```python
from openapi_client.models.game_post_request import GamePostRequest

# TODO update the JSON string below
json = "{}"
# create an instance of GamePostRequest from a JSON string
game_post_request_instance = GamePostRequest.from_json(json)
# print the JSON string representation of the object
print(GamePostRequest.to_json())

# convert the object into a dict
game_post_request_dict = game_post_request_instance.to_dict()
# create an instance of GamePostRequest from a dict
game_post_request_from_dict = GamePostRequest.from_dict(game_post_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


