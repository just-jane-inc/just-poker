using System.Reflection;
using System.Text.Json.Nodes;
using JustPoker.OpenApi.Model;

namespace JustPoker.Sdk.Enums;

public enum PokerEventType {
    Unknown,
    [PokerEventMeta(Text = "*")] All,

    [PokerEventMeta(Text = "welcome", DataType = typeof(GameGameDTO))]
    Welcome,

    [PokerEventMeta(Text = "game_state_update", DataType = typeof(GameGameDTO))]
    GameStateUpdate,

    [PokerEventMeta(Text = "starting_game", DataType = typeof(GameGameDTO))]
    StartingGame,

    [PokerEventMeta(Text = "player_action", DataType = typeof(GamePlayerActionDTO))]
    PlayerAction,

    [PokerEventMeta(Text = "payout", DataType = typeof(JsonNode))]
    Payout,

    [PokerEventMeta(Text = "round_start", DataType = typeof(GameRoundDTO))]
    RoundStart,

    [PokerEventMeta(Text = "hand_started", DataType = typeof(JsonNode))]
    HandStarted,

    [PokerEventMeta(Text = "game_over", DataType = typeof(List<GamePlayerDTO>))]
    GameOver
}

public static class PokerEventTypeExtensions {
    private static readonly Dictionary<PokerEventType, PokerEventMetaAttribute> Meta = BuildMeta();

    private static readonly Dictionary<string, PokerEventType> ByText = Meta
        .Where(entry => entry.Key is not (PokerEventType.All or PokerEventType.Unknown))
        .ToDictionary(entry => entry.Value.Text ?? entry.Key.ToString(), entry => entry.Key,
            StringComparer.OrdinalIgnoreCase);

    public static string GetText(this PokerEventType value) {
        return Meta.TryGetValue(value, out var meta) && meta.Text is { Length: > 0 } text ? text : value.ToString();
    }

    public static Type GetDataType(this PokerEventType value) {
        return Meta.TryGetValue(value, out var meta) ? meta.DataType : typeof(JsonNode);
    }

    public static PokerEventType FromString(string type) {
        return ByText.GetValueOrDefault(type, PokerEventType.Unknown);
    }

    private static Dictionary<PokerEventType, PokerEventMetaAttribute> BuildMeta() {
        var meta = new Dictionary<PokerEventType, PokerEventMetaAttribute>();

        foreach (var field in typeof(PokerEventType).GetFields(BindingFlags.Public | BindingFlags.Static))
            if (field.GetCustomAttribute<PokerEventMetaAttribute>() is { } attribute)
                meta[(PokerEventType)field.GetValue(null)!] = attribute;

        return meta;
    }
}

[AttributeUsage(AttributeTargets.Field)]
public sealed class PokerEventMetaAttribute : Attribute {
    public string? Text { get; set; }
    public Type DataType { get; set; } = typeof(JsonNode);
}