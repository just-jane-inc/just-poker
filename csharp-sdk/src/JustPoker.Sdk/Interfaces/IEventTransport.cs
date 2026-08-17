using JustPoker.Sdk.Models;

namespace JustPoker.Sdk.Interfaces;

public interface IEventTransport : IAsyncDisposable {
    Task ConnectAsync(CancellationToken cancellationToken = default);

    Task CloseAsync();

    IAsyncEnumerable<PokerEvent> EventsAsync(CancellationToken cancellationToken = default);
}