using JustPoker.Sdk.Models;

namespace JustPoker.Sdk.WebSocket;

public sealed class WebSocketStreamOptions {
    public bool Reconnect { get; set; } = true;

    public int MaxRetries { get; set; }

    public TimeSpan RetryBackoff { get; set; } = TimeSpan.FromSeconds(1);

    public TimeSpan MaxRetryBackoff { get; set; } = TimeSpan.FromSeconds(30);

    public Action<CloseInfo>? OnClose { get; set; }
}