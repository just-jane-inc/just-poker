using System.Runtime.CompilerServices;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.Json.Serialization;

namespace JustPoker.Sdk.Serializers;

[Obsolete("x-nullable set to true in swaggo gen to fix")]
internal sealed class NullTolerantModelConverter<T> : JsonConverter<T> where T : class {
    private static readonly ConditionalWeakTable<JsonSerializerOptions, JsonSerializerOptions> Clean = new();

    private static JsonSerializerOptions CleanOptions(JsonSerializerOptions options) {
        return Clean.GetValue(options, static src => {
            var copy = new JsonSerializerOptions(src);
            for (var i = copy.Converters.Count - 1; i >= 0; i--)
                if (copy.Converters[i] is NullTolerantModelConverterFactory)
                    copy.Converters.RemoveAt(i);
            return copy;
        });
    }

    public override T? Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options) {
        var node = JsonNode.Parse(ref reader);
        if (node is null) return null;
        StripNulls(node);
        return node.Deserialize<T>(CleanOptions(options));
    }

    public override void Write(Utf8JsonWriter writer, T value, JsonSerializerOptions options) {
        JsonSerializer.Serialize(writer, value, CleanOptions(options));
    }

    private static void StripNulls(JsonNode node) {
        switch (node) {
            case JsonObject obj:
                foreach (var key in obj.Where(e => e.Value is null).Select(e => e.Key).ToList())
                    obj.Remove(key);
                foreach (var entry in obj.Where(e => e.Value is not null))
                    StripNulls(entry.Value!);
                break;
            case JsonArray array:
                foreach (var item in array.Where(i => i is not null))
                    StripNulls(item!);
                break;
        }
    }
}