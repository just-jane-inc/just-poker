using System.Text.Json;
using System.Text.Json.Serialization;

namespace JustPoker.Sdk.Serializers;

[Obsolete("x-nullable set to true in swaggo gen to fix")]
internal sealed class NullTolerantModelConverterFactory : JsonConverterFactory {
    private const string ModelNamespace = "JustPoker.OpenApi.Model";

    public override bool CanConvert(Type t) {
        return t is { IsClass: true, Namespace: ModelNamespace };
    }

    public override JsonConverter CreateConverter(Type t, JsonSerializerOptions options) {
        return (JsonConverter)Activator.CreateInstance(
            typeof(NullTolerantModelConverter<>).MakeGenericType(t))!;
    }
}