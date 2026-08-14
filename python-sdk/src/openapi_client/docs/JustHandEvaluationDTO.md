# JustHandEvaluationDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**error** | **str** |  | [optional] 
**evaluation** | **int** |  | [optional] 

## Example

```python
from openapi_client.models.just_hand_evaluation_dto import JustHandEvaluationDTO

# TODO update the JSON string below
json = "{}"
# create an instance of JustHandEvaluationDTO from a JSON string
just_hand_evaluation_dto_instance = JustHandEvaluationDTO.from_json(json)
# print the JSON string representation of the object
print(JustHandEvaluationDTO.to_json())

# convert the object into a dict
just_hand_evaluation_dto_dict = just_hand_evaluation_dto_instance.to_dict()
# create an instance of JustHandEvaluationDTO from a dict
just_hand_evaluation_dto_from_dict = JustHandEvaluationDTO.from_dict(just_hand_evaluation_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


