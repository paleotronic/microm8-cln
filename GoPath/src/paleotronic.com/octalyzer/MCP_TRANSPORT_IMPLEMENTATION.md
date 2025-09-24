# MCP Transport Implementation for microM8

## Overview

This implementation adds multiple transport options to microM8's Model Context Protocol (MCP) server, allowing it to communicate over HTTP (via SSE or HTTP streaming) in addition to the default stdio transport.

## Command-Line Options

- `-mcp`: Enable MCP server mode (existing flag)
- `-mcp-mode`: Transport mode selection (default: `stdio`)
  - `stdio`: Standard input/output (default)
  - `sse`: Server-Sent Events over HTTP
  - `streaming` or `http-streaming`: HTTP streaming transport (for broader AI tool compatibility)
- `-mcp-port`: Port for HTTP server (SSE or streaming) (default: `1983`)

## Usage Examples

### Standard stdio mode (default):
```bash
./microM8 -mcp
```

### SSE mode with default port:
```bash
./microM8 -mcp -mcp-mode sse
```

### SSE mode with custom port:
```bash
./microM8 -mcp -mcp-mode sse -mcp-port 8080
```

### HTTP streaming mode with default port:
```bash
./microM8 -mcp -mcp-mode streaming
```

### HTTP streaming mode with custom port:
```bash
./microM8 -mcp -mcp-mode streaming -mcp-port 8080
```

## Transport-Specific Endpoints

### SSE Mode Endpoints

When running in SSE mode, the following HTTP endpoints are available:

- `/mcp/sse`: Main SSE endpoint for MCP communication
- `/mcp/health`: Health check endpoint providing:
  - Server status
  - Active connection count
  - Server uptime

### HTTP Streaming Mode Endpoints

When running in HTTP streaming mode, the following endpoints are available:

- `/mcp/stream`: Main streaming endpoint supporting:
  - `GET`: Establish event stream connection
  - `POST`: Send JSON-RPC messages
  - `DELETE`: Close session
  - `OPTIONS`: CORS preflight
- `/mcp/health`: Health check endpoint with transport info
- `/mcp/info`: Transport information endpoint

## Features

### Common Features (All HTTP Transports)

#### CORS Support
- Allows connections from any origin
- Supports preflight OPTIONS requests
- Configurable headers based on transport type

#### Connection Tracking
- Tracks active connections with atomic counters
- Logs connection establishment and termination
- Available via health endpoint

### SSE-Specific Features

#### Heartbeat Mechanism
- Sends keepalive messages every 30 seconds
- Uses SSE comments (`:keepalive`) to maintain connection
- Prevents proxy/firewall timeouts

### HTTP Streaming-Specific Features

#### Session Management
- Unique session IDs via `Mcp-Session-Id` header
- Session persistence across reconnections
- Built-in event store for stream resumption

#### Protocol Versioning
- Explicit version negotiation via `Mcp-Protocol-Version` header
- Support for protocol evolution

#### Enhanced Reliability
- Event replay capabilities with `Last-Event-ID`
- Proper handling of connection interruptions
- Multiple HTTP methods for different operations

## Implementation Details

1. **Server Creation**: Refactored into `createMCPServer()` function for reusability
2. **Transport Selection**: Based on `-mcp-mode` flag in `StartMCPServerSDK()`
3. **SSE Handler**: Uses official SDK's `mcp.NewSSEHandler()`
4. **Heartbeat**: Goroutine per connection sending periodic keepalives
5. **CORS Headers**: Set on all requests to allow cross-origin access

## Testing

### Testing SSE Transport

1. Start the server:
   ```bash
   ./microM8 -mcp -mcp-mode sse -mcp-port 1977
   ```

2. Check health endpoint:
   ```bash
   curl http://localhost:1977/mcp/health
   ```

3. Connect with an MCP client that supports SSE transport

### Testing HTTP Streaming Transport

1. Start the server:
   ```bash
   ./microM8 -mcp -mcp-mode streaming -mcp-port 1977
   ```

2. Check health endpoint:
   ```bash
   curl http://localhost:1977/mcp/health
   ```

3. Check info endpoint:
   ```bash
   curl http://localhost:1977/mcp/info
   ```

4. Test streaming connection:
   ```bash
   # Create a new session (GET request)
   curl -N http://localhost:1977/mcp/stream \
     -H "Accept: text/event-stream" \
     -v
   ```

5. Send a message (POST request with session ID):
   ```bash
   curl -X POST http://localhost:1977/mcp/stream \
     -H "Mcp-Session-Id: <session-id-from-GET>" \
     -H "Accept: application/json, text/event-stream" \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
   ```

6. Close session (DELETE request):
   ```bash
   curl -X DELETE http://localhost:1977/mcp/stream \
     -H "Mcp-Session-Id: <session-id>"
   ```

7. Use the provided test script:
   ```bash
   ./test_streaming.sh
   ```

## New Tools

### emulator_state
Provides comprehensive emulator state information for agents without vision capabilities. Returns:
- **Machine Profile**: The computer being emulated (e.g., "apple2e", "apple2c", "zxspectrum48k")
- **CPU State**: Running status (EXEC6502/EXECZ80), program counter, registers (A, X, Y, SP, P), and flags
- **Video Mode**: Current display mode (TEXT, HGR, etc.) with description
- **Text Screen**: Full text screen contents when in TEXT or TXT2 mode (excludes OSD layers)
- **Disk Information**: All mounted disks with drive number, path, write protection, and type
- **JSON Output**: Structured data for programmatic access

This tool is particularly useful for AI agents that cannot use vision/screenshots to understand the emulator state. The machine profile helps agents understand whether they're working with an Apple II, ZX Spectrum, or other supported system.

### get_mounted_disks
Returns information about currently mounted disks for the selected slot, including:
- Drive number (0-2)
- File path
- Write protection status
- Disk type (5.25" floppy or SmartPort)

Example output includes both human-readable format and JSON representation.

### enable_live_rewind
Enables the live rewind functionality by starting video recording. This allows you to rewind the emulator state to previous points in time. The function:
- Checks if recording is already active
- Starts recording if not already recording
- Uses empty filename and false flag for live rewind mode (not saving to file)

### rewind_back
Rewinds the emulator state by a specified number of milliseconds. Features:
- Takes an optional `milliseconds` parameter (default: 5000ms)
- Range: 100ms to 60000ms (60 seconds)
- Requires live rewind to be enabled first
- Executes asynchronously using a goroutine
- Example usage: `{"milliseconds": 3000}` to rewind 3 seconds

### start_file_recording
Starts recording the emulator output to a file. Features:
- Stops any existing recording before starting
- Optional `full_cpu` parameter to enable full CPU recording
- Uses global `FileFullCPURecord` setting if parameter not provided
- The recording will be saved to a file when stopped

### stop_recording
Stops any active recording (both file recording and live rewind). Features:
- Works for both file recording and live rewind recording
- No parameters required
- Safe to call even if no recording is active

## Notes

- The heartbeat interval is set to 30 seconds to balance between connection stability and resource usage
- Logging is enabled for debugging connection issues
- The implementation follows the same pattern as the microspaces project for consistency