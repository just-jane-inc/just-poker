using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.Json.Serialization;
using System.Runtime.CompilerServices;
using JustPoker.OpenApi.Api;
using JustPoker.OpenApi.Client;
using JustPoker.OpenApi.Extensions;
using JustPoker.OpenApi.Model;
using Microsoft.Extensions.DependencyInjection;
// ReSharper disable NullableWarningSuppressionIsUsed

namespace JustPoker.Sdk;

public sealed class Chips(int denomination, int count)
{
    public int Denomination { get; } = denomination;
    public int Count { get; set; } = count;
    public int Value => Denomination * Count;
    public override string ToString() => $"{Count}x{Denomination}";
}

public enum CardSuit
{
    Spade = 's',
    Heart = 'h',
    Diamond = 'd',
    Club = 'c',
    Unknown = 'x',
}

public enum CardRank
{
    Ace = 'A',
    Two = '2',
    Three = '3',
    Four = '4',
    Five = '5',
    Six = '6',
    Seven = '7',
    Eight = '8',
    Nine = '9',
    Ten = 'T',
    Jack = 'J',
    Queen = 'Q',
    King = 'K',
    Unknown = 'x',
}

public sealed class Card(CardRank rank, CardSuit suit)
{
    private const int DeckBase = 0x1F0A0;

    private static readonly Dictionary<CardSuit, int> SuitOffsets = new()
    {
        [CardSuit.Spade] = 0x00,
        [CardSuit.Heart] = 0x10,
        [CardSuit.Diamond] = 0x20,
        [CardSuit.Club] = 0x30,
    };

    private static readonly Dictionary<CardRank, int> RankOffsets = new()
    {
        [CardRank.Ace] = 1,
        [CardRank.Two] = 2,
        [CardRank.Three] = 3,
        [CardRank.Four] = 4,
        [CardRank.Five] = 5,
        [CardRank.Six] = 6,
        [CardRank.Seven] = 7,
        [CardRank.Eight] = 8,
        [CardRank.Nine] = 9,
        [CardRank.Ten] = 10,
        [CardRank.Jack] = 11,
        [CardRank.Queen] = 13,
        [CardRank.King] = 14,
    };

    public CardRank Rank { get; } = rank;
    public CardSuit Suit { get; } = suit;

    public static Card FromDto(GameCardDTO dto)
    {
        var cardRank = dto.Rank is { } rank && Enum.IsDefined(typeof(CardRank), rank) ? (CardRank)rank : CardRank.Unknown;
        var cardSuit = dto.Suit is { } suit && Enum.IsDefined(typeof(CardSuit), suit) ? (CardSuit)suit : CardSuit.Unknown;
        return new Card(cardRank, cardSuit);
    }

    public GameCardDTO ToDto() => new((int)Rank, (int)Suit);

    public string ToUnicode()
    {
        if (!SuitOffsets.TryGetValue(Suit, out var suitOffset) || !RankOffsets.TryGetValue(Rank, out var rankOffset))
        {
            return char.ConvertFromUtf32(DeckBase);
        }

        return char.ConvertFromUtf32(DeckBase + suitOffset + rankOffset);
    }

    public override string ToString() => $"{(char)Rank}{(char)Suit}";
}

public static class PokerHelpers
{
    public static int ChipSum(IDictionary<string, int> stack) =>
        stack.Sum(entry => int.Parse(entry.Key) * entry.Value);

    public static int ChipSum(IEnumerable<Chips> chips) => chips.Sum(c => c.Value);

    public static List<Chips> ConvertStack(IDictionary<string, int> stack) =>
        stack.Select(entry => new Chips(int.Parse(entry.Key), entry.Value)).ToList();

    public static Dictionary<string, int> ConvertChips(IEnumerable<Chips> chips) =>
        chips.ToDictionary(c => c.Denomination.ToString(), c => c.Count);

    public static ServiceProvider CreateApiProvider(string baseUrl, string token = "")
    {
        if (string.IsNullOrEmpty(baseUrl))
        {
            throw new PokerException("base url not provided");
        }

        var services = new ServiceCollection();
        services.AddLogging();
        services.AddApi(config =>
        {
            config.AddApiHttpClients(client => client.BaseAddress = new Uri(baseUrl));

            if (!string.IsNullOrEmpty(token))
            {
                config.AddTokens(new BearerToken(token));
            }

            // some stuff was null and causing the generated code to blow up (mostly the endDate on gamegamedto JANE!!!)
            config.ConfigureJsonOptions(options =>
                options.Converters.Insert(0, new NullTolerantModelConverterFactory()));
            
            // fix for serializing the action post request to json as it was always ending up empty due to weird swagger shenanigans
            config.ConfigureJsonOptions(options =>
                options.Converters.Insert(0, new ActionPostRequestConverter()));
            // yeah this also no worky for chip exchange, serializer sends empty
            config.ConfigureJsonOptions(options =>
                options.Converters.Insert(0, new ChipExchangeRequestConverter()));
            
            // There may be other post request objects that do not serialize properly
            // TODO look into root cause (swagger gen having union of object and $ref may be why?)
            // Is there a way to make codegen handle this batter or do we need to identify and solve all via converters...
            
        });

        return services.BuildServiceProvider();
    }

    public static JustErrorDTO? GetError<TResponse>(TResponse response)
        where TResponse : IBadRequest<JustResponseMessageJustErrorDTO?>
    {
        if (response.IsSuccessStatusCode || !response.TryBadRequest(out var message))
        {
            return null;
        }

        return message.Type == "error" ? message.Data : null;
    }

    public static void ThrowError<TResponse>(TResponse response, string action)
        where TResponse : IBadRequest<JustResponseMessageJustErrorDTO?>
    {
        if (response.IsSuccessStatusCode)
        {
            return;
        }

        var error = GetError(response);
        throw new PokerException(error is null
            ? $"{action} failed with {(int)response.StatusCode}: {response.RawContent}"
            : $"{action} failed [{error.ErrorCode}]: {error.Error}");
    }

    public static async Task<GameGameDTO?> GetGameStateAsync(IGameApi api, string gameId)
    {
        var response = await api.GameGameIdStateGetAsync(gameId);
        ThrowError(response, "get game state");
        return response.TryOk(out var message) ? message.Data : null;
    }

    public static async Task<string?> CreateGameFromConfigAsync(string baseUrl, string token, GameNewGameConfigDTO config)
    {
        await using var provider = CreateApiProvider(baseUrl, token);
        var api = provider.GetRequiredService<IGameApi>();

        var response = await api.GamePostAsync(new GamePostRequest(config));
        ThrowError(response, "create game");
        return response.TryOk(out var message) ? message.Data : null;
    }

    public static Task<string?> CreateGameAsync(string baseUrl, string token, int bigBlind = 100,
        int smallBlind = 50, IDictionary<string, int>? startingChips = null, int playerCount = 5,
        bool autoStartHands = true, IList<int>? denominations = null)
    {
        startingChips ??= new Dictionary<string, int>
        {
            // Default
            ["10"] = 10,
            ["50"] = 5,
            ["100"] = 2,
            ["500"] = 1,
        };
        
        denominations ??= [10, 50, 100, 500]; // Default

        var config = new GameNewGameConfigDTO(autoStartsHands: autoStartHands, bigBlind: bigBlind,
            chipDenominations: denominations.ToList(), playerCount: playerCount, smallBlind: smallBlind,
            startingChips: new Dictionary<string, int>(startingChips));

        return CreateGameFromConfigAsync(baseUrl, token, config);
    }

    public static async Task StartGameAsync(IGameApi api, string gameId)
    {
        var response = await api.GameGameIdStartedPostAsync(gameId);
        ThrowError(response, "start game");
    }

    public static async Task DeleteUserAsync(string baseUrl, string token)
    {
        await using var provider = CreateApiProvider(baseUrl, token);
        var api = provider.GetRequiredService<IUserApi>();

        var response = await api.UserMeDeleteAsync();
        ThrowError(response, "delete user");
    }
}

internal static class PokerJson
{
    public static JsonSerializerOptions Options { get; } = BuildOptions();

    private static JsonSerializerOptions BuildOptions()
    {
        using var provider = PokerHelpers.CreateApiProvider(ClientUtils.BASE_ADDRESS);
        return provider.GetRequiredService<JsonSerializerOptionsProvider>().Options;
    }
}
internal sealed class NullTolerantModelConverterFactory : JsonConverterFactory
{
    private const string ModelNamespace = "JustPoker.OpenApi.Model";

    public override bool CanConvert(Type t) => t is { IsClass: true, Namespace: ModelNamespace };

    public override JsonConverter CreateConverter(Type t, JsonSerializerOptions options) =>
        (JsonConverter)Activator.CreateInstance(
            typeof(NullTolerantModelConverter<>).MakeGenericType(t))!;
}

internal sealed class NullTolerantModelConverter<T> : JsonConverter<T> where T : class
{
    private static readonly ConditionalWeakTable<JsonSerializerOptions, JsonSerializerOptions> Clean = new();

    private static JsonSerializerOptions CleanOptions(JsonSerializerOptions options) =>
        Clean.GetValue(options, static src =>
        {
            var copy = new JsonSerializerOptions(src);
            for (int i = copy.Converters.Count - 1; i >= 0; i--)
                if (copy.Converters[i] is NullTolerantModelConverterFactory)
                    copy.Converters.RemoveAt(i);
            return copy;
        });

    public override T? Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options)
    {
        var node = JsonNode.Parse(ref reader);
        if (node is null) return null;
        StripNulls(node);
        return node.Deserialize<T>(CleanOptions(options));
    }

    public override void Write(Utf8JsonWriter writer, T value, JsonSerializerOptions options) =>
        JsonSerializer.Serialize(writer, value, CleanOptions(options));

    private static void StripNulls(JsonNode node)
    {
        switch (node)
        {
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

internal sealed class ActionPostRequestConverter : JsonConverter<GameGameIdActionPostRequest>
{
    public override GameGameIdActionPostRequest Read(ref Utf8JsonReader reader, Type objectType, JsonSerializerOptions options) => 
        new(JsonSerializer.Deserialize<GamePlayerActionDTO>(ref reader, options)!);
    
    public override void Write(Utf8JsonWriter writer, GameGameIdActionPostRequest actionReq, JsonSerializerOptions options)
    {
        if (actionReq.GamePlayerActionDTO is not null)
        {
            JsonSerializer.Serialize(writer, actionReq.GamePlayerActionDTO, options);
        }
        else if (actionReq.Object is not null)
        {
            JsonSerializer.Serialize(writer, actionReq.Object, options);
        }
        else
        {
            writer.WriteStartObject();
        }
    }

}

internal sealed class ChipExchangeRequestConverter : JsonConverter<GameGameIdChipExchangePostRequest>
{
    public override GameGameIdChipExchangePostRequest Read(ref Utf8JsonReader reader, Type objectType, JsonSerializerOptions options) => 
        new(JsonSerializer.Deserialize<GameChipExchangeDTO>(ref reader, options)!);
    
    public override void Write(Utf8JsonWriter writer, GameGameIdChipExchangePostRequest actionReq, JsonSerializerOptions options)
    {
        if (actionReq.GameChipExchangeDTO is not null)
        {
            JsonSerializer.Serialize(writer, actionReq.GameChipExchangeDTO, options);
        }
        else if (actionReq.Object is not null)
        {
            JsonSerializer.Serialize(writer, actionReq.Object, options);
        }
        else
        {
            writer.WriteStartObject();
        }
    }

}