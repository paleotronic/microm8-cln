## New in this version

### MCP (Model Context Protocol) Server Integration
Enables AI assistants and LLMs to interact with microM8: 
  - Compatible with Claude, ChatGPT, and other MCP-enabled clients
  - Two transport modes: stdio (default) and SSE (Server-Sent Events)

#### How to Enable MCP Server
  
Stdio Mode (default):
```shell
./microM8 -mcp
```

Config:
```json
{
    "mcpServers": {
        "microm8": {
            "args": [
                "-mcp"
            ],
            "command": "microM8",
            "description": "MCP server for controlling the microM8 Apple II emulator. Provides tools for disk management, keyboard input, memory access, and emulator control.",
            "name": "microM8 Emulator Control"
        }
    }
}

```

- Uses standard input/output for communication
- Best for direct integration with AI tools
  
SSE Mode (for web-based clients):
```shell
./microM8 -mcp -mcp-mode sse
./microM8 -mcp -mcp-mode sse -mcp-port 8080
```

Config:
```json
{
    "mcpServers": {
        "microm8": {
            "url": "http://localhost:1983/mcp/sse",
            "type": "sse",
            "description": "MCP server for controlling the microM8 Apple II emulator. Provides tools for disk management, keyboard input, memory access, and emulator control."
        }
    }
}
```


- HTTP-based Server-Sent Events transport
- Default port: 1983
- Includes CORS support and heartbeat mechanism
- Health check endpoint at /mcp/health

#### MCP Emulator Control Tools
  - reboot: Restart the emulator
  - pause: Pause/unpause emulation
  - break: Send break signal (Ctrl+C)
  - set_cpu_speed: Adjust CPU speed multiplier (0.25x to 4x)
  - screenshot: Capture emulator screen
  - type_text: Type text with configurable keystroke delay
  - key_event: Send individual keyboard events

#### MCP Disk Management
  - insert_disk: Insert disk images into drives
  - insert_disk_file: Insert disks from virtual file system
  - eject_disk: Remove disks from drives
  - get_mounted_disks: List currently mounted disks with paths and write protection status
  - Automatic detection of 5.25" vs 3.5"/HD disk types

#### MCP Recording and Rewind Tools
  - enable_live_rewind: Enable live rewind functionality
  - rewind_back: Rewind emulator state by milliseconds (default: 5000ms)
  - start_file_recording: Start recording to file with optional full CPU mode
  - stop_recording: Stop any active recording (file or live rewind)

#### MCP File System Access
  - list_files: Browse the virtual file system
  - read_file: Read files with automatic text/binary detection
  - Access local files, disk images, and remote resources
  - Support for all microM8 file providers

#### MCP Memory Operations
  - read_memory: Read bytes from any memory location
  - write_memory: Write individual bytes
  - write_memory_range: Write multiple bytes efficiently
  - get_text_screen: Capture current text display

#### MCP Programming Tools
  - assemble: Compile 6502 assembly code to memory
  - disassemble: Disassemble memory to 6502 instructions
  - applesoft_read: Extract BASIC programs from memory
  - applesoft_write: Tokenize and load BASIC programs

#### MCP Debugger Integration
  - debug_cpu_control: Step, continue, pause execution
  - debug_breakpoint_*: Manage breakpoints
  - debug_register_*: Read/set CPU registers
  - debug_memory_*: Advanced memory operations
  - debug_instruction_trace: View execution history
  - Automatic debugger initialization when needed

### MCP Implementation Details
  - Uses official Go SDK for protocol compliance
  - JSON-RPC 2.0 communication protocol
  - SSE mode includes 30-second heartbeat for connection stability
  - Comprehensive error handling and validation
  - Type-safe parameter structures with JSON schema
