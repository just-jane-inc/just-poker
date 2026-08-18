# JustPoker.OpenApi.Model.GameTableDTO
the table

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BigBlindPosition** | **int** | the position of the big blind | [optional] 
**ButtonPosition** | **int** | the position of the button | [optional] 
**CurrentHand** | [**GameHandDTO**](GameHandDTO.md) |  | [optional] 
**CurrentRound** | [**GameRoundDTO**](GameRoundDTO.md) |  | [optional] 
**Players** | [**List&lt;GamePlayerDTO&gt;**](GamePlayerDTO.md) | An array of players at the table | [optional] 
**Pot** | **Dictionary&lt;string, int&gt;** | an optional mapping of chips that is required by some action types. | [optional] 
**SmallBlindPosition** | **int** | the position of the small blind | [optional] 
**Street** | [**List&lt;GameCardDTO&gt;**](GameCardDTO.md) | the cards that are on the street (community cards) | [optional] 

[[Back to Model list]](../../README.md#documentation-for-models) [[Back to API list]](../../README.md#documentation-for-api-endpoints) [[Back to README]](../../README.md)

