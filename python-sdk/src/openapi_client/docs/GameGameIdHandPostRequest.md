# GameGameIdHandPostRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**deck** | [**List[GameCardDTO]**](GameCardDTO.md) | the deck, optionally provided to determine trhe order of cards if not provided the cards will be ordered randomly by the server | [optional] 

## Example

```python
from openapi_client.models.game_game_id_hand_post_request import GameGameIdHandPostRequest

# TODO update the JSON string below
json = "{}"
# create an instance of GameGameIdHandPostRequest from a JSON string
game_game_id_hand_post_request_instance = GameGameIdHandPostRequest.from_json(json)
# print the JSON string representation of the object
print(GameGameIdHandPostRequest.to_json())

# convert the object into a dict
game_game_id_hand_post_request_dict = game_game_id_hand_post_request_instance.to_dict()
# create an instance of GameGameIdHandPostRequest from a dict
game_game_id_hand_post_request_from_dict = GameGameIdHandPostRequest.from_dict(game_game_id_hand_post_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


