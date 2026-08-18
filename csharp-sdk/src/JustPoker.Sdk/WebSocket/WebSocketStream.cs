using System.Net.WebSockets;
using System.Runtime.CompilerServices;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using JustPoker.Sdk.Enums;
using JustPoker.Sdk.Interfaces;
using JustPoker.Sdk.Models;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Logging.Abstractions;

namespace JustPoker.Sdk.WebSocket;

public sealed class WebSocketStream : IEventTransport {
    private const int ReceiveBufferSize = 8 * 1024;
    private readonly string _gameId;
    private readonly ILogger _logger;
    private readonly WebSocketStreamOptions _options;
    private readonly string _token;

    private bool _closed;
    private bool _reconnect;

    private ClientWebSocket? _socket;

    public WebSocketStream(string baseUrl, string token, string gameId,
        WebSocketStreamOptions? options = null, ILogger? logger = null) {
        if (string.IsNullOrEmpty(token)) throw new PokerException("api token not provided");

        Url = BuildStateWebSocketUrl(baseUrl, gameId);
        _token = token;
        _gameId = gameId;
        _options = options ?? new WebSocketStreamOptions();
        _logger = logger ?? NullLogger.Instance;
        _reconnect = _options.Reconnect;
    }

    public string Url { get; }

    public async Task ConnectAsync(CancellationToken cancellationToken = default) {
        if (_socket is { State: WebSocketState.Open })
            return;

        _socket?.Dispose();
        _closed = false;

        var socket = new ClientWebSocket();
        socket.Options.SetRequestHeader("Authorization", $"Bearer {_token}");

        try {
            await socket.ConnectAsync(new Uri(Url), cancellationToken);
        }
        catch {
            socket.Dispose();
            throw;
        }

        _socket = socket;
        _logger.LogDebug("connected to game state feed for game {GameId}", _gameId);
    }

    public async Task CloseAsync() {
        var socket = _socket;
        _socket = null;

        if (socket is null) {
            _closed = true;
            return;
        }

        if (!_closed && socket.State == WebSocketState.Open) FireClose(new CloseInfo(-1, "manual close requested"));

        _closed = true;

        try {
            if (socket.State is WebSocketState.Open or WebSocketState.CloseReceived) {
                using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(5));
                await socket.CloseAsync(WebSocketCloseStatus.NormalClosure, string.Empty, timeout.Token);
            }
        }
        catch (Exception e) {
            _logger.LogDebug("error while closing update feed: {Message}", e.Message);
        }
        finally {
            socket.Dispose();
        }
    }

    public async IAsyncEnumerable<PokerEvent> EventsAsync(
        [EnumeratorCancellation] CancellationToken cancellationToken = default) {
        var attempts = 0;
        while (!_closed && !cancellationToken.IsCancellationRequested) {
            var (connected, connectError) = await TryConnectAsync(cancellationToken);
            if (connected) {
                attempts = 0;

                while (true) {
                    var (message, receiveError) = await TryReceiveAsync(cancellationToken);
                    if (receiveError is not null) {
                        await HandleReceiveErrorAsync(receiveError);
                        break;
                    }

                    if (message is null) break;

                    PokerEvent pokerEvent;
                    try {
                        pokerEvent = ParseEvent(message);
                    }
                    catch (PokerException e) {
                        _logger.LogError("dropping bad update message: {Message}", e.Message);
                        continue;
                    }

                    yield return pokerEvent;

                    if (pokerEvent.EventType.Equals(PokerEventType.GameOver)) {
                        _reconnect = false;
                        await CloseAsync();
                        yield break;
                    }
                }
            }
            else if (connectError is not null) {
                if (_closed) yield break;

                if (!_reconnect)
                    throw new PokerException($"game state feed failed: {connectError.Message}", connectError);

                _logger.LogWarning("game state feed dropped, reconnecting: {Message}", connectError.Message);
            }

            if (_closed || !_reconnect || cancellationToken.IsCancellationRequested) yield break;

            attempts++;
            if (_options.MaxRetries > 0 && attempts > _options.MaxRetries)
                throw new PokerException($"game state feed gave up after {_options.MaxRetries} retries");

            await Task.Delay(BackoffFor(attempts), cancellationToken);
        }
    }

    public async ValueTask DisposeAsync() {
        await CloseAsync();
    }

    private static string BuildStateWebSocketUrl(string baseUrl, string gameId) {
        if (string.IsNullOrEmpty(baseUrl)) throw new PokerException("base url not provided");

        if (string.IsNullOrEmpty(gameId)) throw new PokerException("game id not provided");

        var builder = new UriBuilder(baseUrl) {
            Scheme = baseUrl.StartsWith("https", StringComparison.OrdinalIgnoreCase) ? "wss" : "ws"
        };

        builder.Path = $"{builder.Path.TrimEnd('/')}/game/{gameId}/state/ws";
        builder.Query = string.Empty;
        return builder.Uri.ToString();
    }

    private static PokerEvent ParseEvent(string message) {
        JsonNode? payload;
        try {
            payload = JsonNode.Parse(message);
        }
        catch (JsonException e) {
            throw new PokerException($"received non json message on update feed: {e.Message}");
        }

        if (payload is not JsonObject envelope) throw new PokerException("received non object message on update feed");

        if (!envelope.ContainsKey("data")) throw new PokerException($"unexpected data received: {message}");

        var data = envelope["data"];
        if (data is null or JsonValue) throw new PokerException($"unexpected data received: {message}");

        var wireName = envelope["event_type"]?.GetValue<string>() ?? string.Empty;
        var eventType = PokerEventTypeExtensions.FromString(wireName);

        long id = 0;
        var rawId = envelope["id"];
        if (rawId is not null && !TryReadId(rawId, out id))
            throw new PokerException($"non integer event id: {rawId.ToJsonString()}");

        var timeSent = envelope["time_sent"]?.ToString() ?? string.Empty;

        return new PokerEvent(eventType, data.DeepClone(), id, timeSent, envelope);
    }

    private static bool TryReadId(JsonNode node, out long id) {
        if (node is JsonValue value) {
            if (value.TryGetValue(out long parsed)) {
                id = parsed;
                return true;
            }

            if (value.TryGetValue(out string? text) && long.TryParse(text, out parsed)) {
                id = parsed;
                return true;
            }
        }

        id = 0;
        return false;
    }

    private TimeSpan BackoffFor(int attempts) {
        var scaled = _options.RetryBackoff.TotalSeconds * Math.Pow(2, attempts - 1);
        return TimeSpan.FromSeconds(Math.Min(scaled, _options.MaxRetryBackoff.TotalSeconds));
    }

    private async Task<(bool Connected, Exception? Error)> TryConnectAsync(CancellationToken cancellationToken) {
        try {
            await ConnectAsync(cancellationToken);
            return (true, null);
        }
        catch (OperationCanceledException) {
            throw;
        }
        catch (Exception e) {
            return (false, e);
        }
    }

    private async Task<(string? Message, Exception? Error)> TryReceiveAsync(CancellationToken cancellationToken) {
        var socket = _socket;
        if (socket is null) return (null, null);

        var buffer = new byte[ReceiveBufferSize];
        var payload = new MemoryStream();

        try {
            while (true) {
                var result = await socket.ReceiveAsync(new ArraySegment<byte>(buffer), cancellationToken);

                if (result.MessageType == WebSocketMessageType.Close) {
                    FireClose(new CloseInfo((int?)result.CloseStatus, result.CloseStatusDescription ?? string.Empty));
                    _reconnect = false;
                    await CloseAsync();
                    return (null, null);
                }

                payload.Write(buffer, 0, result.Count);

                if (result.EndOfMessage) return (Encoding.UTF8.GetString(payload.ToArray()), null);
            }
        }
        catch (OperationCanceledException) {
            throw;
        }
        catch (Exception e) {
            return (null, e);
        }
    }

    private async Task HandleReceiveErrorAsync(Exception error) {
        if (_closed) return;

        if (error is WebSocketException socketError) {
            FireClose(new CloseInfo(null, socketError.Message));
            _logger.LogWarning("received abnormal close event: {Message}", socketError.Message);
            _reconnect = false;
        }

        await CloseAsync();
    }

    private void FireClose(CloseInfo info) {
        if (_options.OnClose is null) return;

        try {
            _options.OnClose(info);
        }
        catch (Exception e) {
            _logger.LogError("error in on_close handler: {Message}", e.Message);
        }
    }
}