# GameGameIdStateListenGet200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | [**GameChipExchangeDTO**](GameChipExchangeDTO.md) |  | [optional] 
**event_type** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**time_sent** | **str** |  | [optional] 

## Example

```python
from openapi_client.models.game_game_id_state_listen_get200_response import GameGameIdStateListenGet200Response

# TODO update the JSON string below
json = "{}"
# create an instance of GameGameIdStateListenGet200Response from a JSON string
game_game_id_state_listen_get200_response_instance = GameGameIdStateListenGet200Response.from_json(json)
# print the JSON string representation of the object
print(GameGameIdStateListenGet200Response.to_json())

# convert the object into a dict
game_game_id_state_listen_get200_response_dict = game_game_id_state_listen_get200_response_instance.to_dict()
# create an instance of GameGameIdStateListenGet200Response from a dict
game_game_id_state_listen_get200_response_from_dict = GameGameIdStateListenGet200Response.from_dict(game_game_id_state_listen_get200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


