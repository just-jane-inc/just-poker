# GameHandEvaluationDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**error** | **str** |  | [optional] 
**evaluation** | **int** |  | [optional] 

## Example

```python
from openapi_client.models.game_hand_evaluation_dto import GameHandEvaluationDTO

# TODO update the JSON string below
json = "{}"
# create an instance of GameHandEvaluationDTO from a JSON string
game_hand_evaluation_dto_instance = GameHandEvaluationDTO.from_json(json)
# print the JSON string representation of the object
print(GameHandEvaluationDTO.to_json())

# convert the object into a dict
game_hand_evaluation_dto_dict = game_hand_evaluation_dto_instance.to_dict()
# create an instance of GameHandEvaluationDTO from a dict
game_hand_evaluation_dto_from_dict = GameHandEvaluationDTO.from_dict(game_hand_evaluation_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


