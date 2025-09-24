# MCP HTTP Streaming Implementation Plan for microM8

## Overview

The MCP Go SDK (v0.2.1+) now includes support for HTTP streaming transport via `StreamableHTTPHandler`, which provides an alternative to the existing SSE (Server-Sent Events) transport. This allows compatibility with AI tools that only support HTTP streaming and not SSE.

## Current Implementation

microM8 currently supports two MCP transport modes:
1. **stdio** (default): Standard input/output communication
2. **sse**: Server-Sent Events over HTTP

The transport is selected via the `-mcp-mode` flag, with SSE using `mcp.NewSSEHandler()`.

## HTTP Streaming Support in SDK

The updated SDK provides:
- `StreamableHTTPHandler`: Main handler for HTTP streaming sessions
- `StreamableServerTransport`: Server-side transport implementation
- Built-in session management with unique session IDs
- Event store for stream resumption capabilities
- Support for GET, POST, and DELETE methods
- Proper handling of `Accept` headers (requires both `application/json` and `text/event-stream`)

## Implementation Changes Required

### 1. Add New Transport Mode Option
- Add `"streaming"` or `"http-streaming"` as a valid value for `-mcp-mode` flag
- Update command-line parsing to recognize the new mode

### 2. Create HTTP Streaming Handler Function
```go
func startMCPServerHTTPStreaming(server *mcp.Server, port int) error {
    // Create StreamableHTTPHandler using the SDK
    streamHandler := mcp.NewStreamableHTTPHandler(
        func(r *http.Request) *mcp.Server {
            return server
        },
        nil, // Use default options
    )
    
    // Set up HTTP mux with proper endpoints
    mux := http.NewServeMux()
    mux.Handle("/mcp/stream", streamHandler)
    mux.HandleFunc("/mcp/health", healthHandler)
    
    // Start HTTP server with CORS support
    // Similar to SSE implementation
}
```

### 3. Update Transport Selection Logic
Modify `StartMCPServerSDK()` to handle three transport modes:
```go
switch *mcpTransport {
case "sse":
    return startMCPServerSSE(server, *mcpPort)
case "streaming", "http-streaming":
    return startMCPServerHTTPStreaming(server, *mcpPort)
default: // stdio
    return startMCPServerStdio(server)
}
```

### 4. Key Differences from SSE Implementation

**SSE Handler:**
- Uses `mcp.NewSSEHandler()`
- Simpler session management
- Custom heartbeat implementation
- Single GET endpoint for event stream

**HTTP Streaming Handler:**
- Uses `mcp.NewStreamableHTTPHandler()`
- Built-in session management with IDs
- Supports GET (stream), POST (messages), DELETE (close session)
- Requires `Mcp-Session-Id` header for session continuity
- Built-in event store for resumption

### 5. Session Management Considerations

The HTTP streaming transport requires:
- Session ID tracking via `Mcp-Session-Id` header
- Support for session resumption with `Last-Event-ID`
- Proper cleanup on DELETE requests
- Event store for replay capabilities

### 6. CORS Configuration

Similar to SSE, but must handle additional headers:
- `Mcp-Protocol-Version`
- `Mcp-Session-Id`
- Standard CORS headers for cross-origin support

### 7. Health Endpoint

Can reuse existing health endpoint, but should indicate transport type:
```json
{
  "status": "running",
  "transport": "http-streaming",
  "active_sessions": 2,
  "uptime": "1h23m"
}
```

## Benefits of HTTP Streaming Transport

1. **Broader Compatibility**: Works with AI tools that don't support SSE
2. **Session Persistence**: Built-in session management and resumption
3. **Standard HTTP Methods**: Uses POST for messages, GET for streaming
4. **Event Replay**: Built-in event store for reliability
5. **Protocol Versioning**: Explicit version negotiation via headers

## Testing Strategy

1. Start server with new transport:
   ```bash
   ./microM8 -mcp -mcp-mode streaming -mcp-port 1977
   ```

2. Test session creation:
   ```bash
   curl -X GET http://localhost:1977/mcp/stream \
     -H "Accept: text/event-stream"
   ```

3. Test message sending (with session ID from response):
   ```bash
   curl -X POST http://localhost:1977/mcp/stream \
     -H "Mcp-Session-Id: <session-id>" \
     -H "Accept: application/json, text/event-stream" \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
   ```

4. Test session cleanup:
   ```bash
   curl -X DELETE http://localhost:1977/mcp/stream \
     -H "Mcp-Session-Id: <session-id>"
   ```

## Implementation Priority

1. **Phase 1**: Basic HTTP streaming support
   - Add transport mode option
   - Implement basic StreamableHTTPHandler
   - Test with simple MCP clients

2. **Phase 2**: Enhanced features
   - Session persistence across reconnects
   - Event replay capabilities
   - Metrics and monitoring

3. **Phase 3**: Documentation
   - Update MCP_DEVELOPMENT_GUIDE.md
   - Add examples for HTTP streaming clients
   - Document session management

## Backward Compatibility

- Existing SSE and stdio transports remain unchanged
- Default behavior (stdio) is preserved
- SSE mode continues to work as before
- New mode is opt-in via command-line flag

## Estimated Effort

- Core implementation: ~200-300 lines of code
- Similar structure to existing SSE implementation
- Can reuse much of the HTTP server setup, CORS, and health endpoint code
- Main work is adapting to StreamableHTTPHandler API

## Next Steps

1. Implement the basic HTTP streaming handler
2. Add command-line flag support
3. Test with various MCP clients
4. Update documentation
5. Consider adding configuration options for event store size and session TTL