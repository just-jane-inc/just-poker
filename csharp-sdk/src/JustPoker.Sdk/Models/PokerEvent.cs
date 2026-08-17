using System.Text.Json;
using System.Text.Json.Nodes;
using JustPoker.Sdk.Enums;
using JustPoker.Sdk.Serializers;

namespace JustPoker.Sdk.Models;

public sealed class PokerEvent(
    PokerEventType eventType,
    JsonNode? data,
    long id = 0,
    string timeSent = "",
    JsonNode? raw = null) {
    public PokerEventType EventType { get; } = eventType;
    public long Id { get; } = id;
    public string TimeSent { get; } = timeSent;
    public JsonNode? RawData { get; } = data;
    public JsonNode? Raw { get; } = raw;

    public object? Data() {
        return Deserialize(EventType.GetDataType());
    }

    public T? DataAs<T>() where T : class {
        if (RawData is null) return null;

        if (typeof(T) == typeof(JsonNode)) return RawData as T;

        if (EventType.GetDataType() == typeof(T)) return Deserialize(typeof(T)) as T;

        return null;
    }

    private object? Deserialize(Type dataType) {
        if (RawData is null) return null;

        if (dataType == typeof(JsonNode)) return RawData;

        try {
            return RawData.Deserialize(dataType, PokerJson.Options);
        }
        catch (JsonException) {
            return RawData;
        }
    }
}