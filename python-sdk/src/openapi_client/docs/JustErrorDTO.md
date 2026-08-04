# JustErrorDTO


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**error** | **str** |  | [optional] 
**error_code** | **int** |  | [optional] 

## Example

```python
from openapi_client.models.just_error_dto import JustErrorDTO

# TODO update the JSON string below
json = "{}"
# create an instance of JustErrorDTO from a JSON string
just_error_dto_instance = JustErrorDTO.from_json(json)
# print the JSON string representation of the object
print(JustErrorDTO.to_json())

# convert the object into a dict
just_error_dto_dict = just_error_dto_instance.to_dict()
# create an instance of JustErrorDTO from a dict
just_error_dto_from_dict = JustErrorDTO.from_dict(just_error_dto_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


