using JustPoker.OpenApi.Model;
using JustPoker.Sdk.Enums;

namespace JustPoker.Sdk.Models;

public sealed class Card(CardRank rank, CardSuit suit) {
    private const int DeckBase = 0x1F0A0;

    private static readonly Dictionary<CardSuit, int> SuitOffsets = new() {
        [CardSuit.Spade] = 0x00,
        [CardSuit.Heart] = 0x10,
        [CardSuit.Diamond] = 0x20,
        [CardSuit.Club] = 0x30
    };

    private static readonly Dictionary<CardRank, int> RankOffsets = new() {
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
        [CardRank.King] = 14
    };

    public CardRank Rank { get; } = rank;
    public CardSuit Suit { get; } = suit;

    public static Card FromDto(GameCardDTO dto) {
        var cardRank = dto.Rank is { } rank && Enum.IsDefined(typeof(CardRank), rank)
            ? (CardRank)rank
            : CardRank.Unknown;
        var cardSuit = dto.Suit is { } suit && Enum.IsDefined(typeof(CardSuit), suit)
            ? (CardSuit)suit
            : CardSuit.Unknown;
        return new Card(cardRank, cardSuit);
    }

    public GameCardDTO ToDto() {
        return new GameCardDTO((int)Rank, (int)Suit);
    }

    public string ToUnicode() {
        if (!SuitOffsets.TryGetValue(Suit, out var suitOffset) || !RankOffsets.TryGetValue(Rank, out var rankOffset))
            return char.ConvertFromUtf32(DeckBase);

        return char.ConvertFromUtf32(DeckBase + suitOffset + rankOffset);
    }

    public override string ToString() {
        return $"{(char)Rank}{(char)Suit}";
    }
}