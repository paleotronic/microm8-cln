# microM8 - Apple II Emulator

microM8 is a comprehensive Apple II emulator with advanced features for both casual users and developers. It provides accurate emulation of Apple II hardware along with modern integration capabilities including MCP (Model Context Protocol) support for AI assistant interaction.

## Features

- Full Apple II/II+ emulation
- 6502 CPU emulation with cycle-accurate timing
- Support for disk operations and various disk image formats
- Built-in assembler (asm65xx)
- MCP server integration for AI-powered control and automation
- Cross-platform support (Linux, macOS, Windows)
- Remote interface capabilities
- Debugging and development tools

## Project Structure

```
microm8-clean/
├── GoPath/src/paleotronic.com/
│   ├── octalyzer/        # Main emulator client
│   ├── server/           # Server components
│   ├── core/             # Core emulation logic
│   ├── debugger/         # Debugging tools
│   └── ...
└── ...
```

## Building from Source

### Prerequisites

- Go programming language (1.16 or later)
- Git
- Standard build tools (gcc, make, etc.)

### Build Instructions

1. Clone the repository:
```bash
git clone <repository-url>
cd microm8-clean
```

2. Navigate to the octalyzer directory:
```bash
cd GoPath/src/paleotronic.com/octalyzer
```

3. Build the project using the lmake.sh script:
```bash
./lmake.sh build
```

This will create the `microM8` executable in the current directory.

### Additional Build Options

The lmake.sh script supports several build targets:

- `./lmake.sh build` - Standard build (default)
- `./lmake.sh run` - Build and run the emulator
- `./lmake.sh asm` - Build the assembler component
- `./lmake.sh remint` - Build the remote interface version
- `./lmake.sh nox` - Build without X11 dependencies (noxarchaist)
- `./lmake.sh macos` - Build for macOS x86_64 (requires xgo)
- `./lmake.sh profile` - Build and run with profiling

**Note:** Build messages about xgo not being found can be safely ignored unless you're cross-compiling for macOS.

## Running microM8

After building, you can run the emulator:

```bash
./microM8
```

### MCP Server Mode

To enable MCP server for AI integration:

**Stdio mode (default):**
```bash
./microM8 -mcp
```

**SSE mode (for web-based clients):**
```bash
./microM8 -mcp -mcp-mode sse
./microM8 -mcp -mcp-mode sse -mcp-port 8080
```

## Development

### Building Components Separately

**Server:**
```bash
cd GoPath/src/paleotronic.com/server
go build
```

**Client (Octalyzer):**
```bash
cd GoPath/src/paleotronic.com/octalyzer
go build
```

**Remote Interface:**
```bash
cd GoPath/src/paleotronic.com/octalyzer
go build -tags remint
```

## License

MIT License

## Contributing

[Contribution guidelines to be added]

## Support

[Support information to be added]
