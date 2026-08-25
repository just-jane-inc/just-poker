# GamePayoutEventDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**chips** | **Dict[str, int]** | an optional mapping of chips that is required by some action types. | [optional] 
**player_id** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.game_payout_event_dto import GamePayoutEventDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GamePayoutEventDTO from a JSON string
game_payout_event_dto_instance = GamePayoutEventDTO.from_json(json)
# print the JSON string representation of the object
print(GamePayoutEventDTO.to_json())

# convert the object into a dict
game_payout_event_dto_dict = game_payout_event_dto_instance.to_dict()
# create an instance of GamePayoutEventDTO from a dict
game_payout_event_dto_from_dict = GamePayoutEventDTO.from_dict(game_payout_event_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


