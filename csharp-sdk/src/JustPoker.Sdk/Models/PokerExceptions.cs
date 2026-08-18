namespace JustPoker.Sdk.Models;

public class PokerException : Exception {
    public PokerException(string message)
        : base(message) {
    }

    public PokerException(string message, Exception inner)
        : base(message, inner) {
    }
}

public sealed class InvalidBetException(string message) : PokerException(message);