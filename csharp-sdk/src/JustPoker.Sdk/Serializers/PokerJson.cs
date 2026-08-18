using System.Text.Json;
using JustPoker.OpenApi.Client;
using Microsoft.Extensions.DependencyInjection;

namespace JustPoker.Sdk.Serializers;

internal static class PokerJson {
    public static JsonSerializerOptions Options { get; } = BuildOptions();

    private static JsonSerializerOptions BuildOptions() {
        using var provider = PokerHelpers.CreateApiProvider(ClientUtils.BASE_ADDRESS);
        return provider.GetRequiredService<JsonSerializerOptionsProvider>().Options;
    }
}